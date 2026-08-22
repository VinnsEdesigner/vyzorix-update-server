// Package channel implements org-scoped live channels: typed channel
// addresses, permission-gated subscriptions, and server-managed streams.
// Alert notifications publish to managed channels instead of raw
// broadcast-to-all.
package channel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// ErrForbidden is returned when a subscribe attempt lacks the required scope.
var ErrForbidden = errors.New("channel subscribe forbidden")

// ErrInvalidChannel is returned for malformed channel addresses.
var ErrInvalidChannel = errors.New("invalid channel address")

// Channel identifies one logical event stream: "stream/<org>/<scope>".
// Scope carries what the channel emits (alerts, commands, members, telemetry);
// the org prefix is the security boundary.
type Channel struct {
	OrgID string
	Scope string
}

// Parse validates and splits a channel address. Valid: "stream/<org>/<scope>"
// where scope is one of the named constants below (or a sub-scope like
// "telemetry/device/<id>").
func Parse(addr string) (*Channel, error) {
	parts := strings.Split(addr, "/")
	if len(parts) < 3 || parts[0] != "stream" || parts[1] == "" || parts[2] == "" {
		return nil, fmt.Errorf("%w: %s", ErrInvalidChannel, addr)
	}
	return &Channel{OrgID: parts[1], Scope: strings.Join(parts[2:], "/")}, nil
}

// String renders the canonical channel address.
func (c *Channel) String() string {
	return "stream/" + c.OrgID + "/" + c.Scope
}

// Alert is the channel scope for alert transitions.
const Alert = "alerts"

// Commands is the channel scope for command dispatch/status events.
const Commands = "commands"

// Members is the channel scope for org membership changes.
const Members = "members"

// Telemetry is the channel scope root for device telemetry frames.
const Telemetry = "telemetry"

// TelemetrySubScope prefixes a device-scoped telemetry channel.
const TelemetrySubScope = "telemetry/device/"

// Message is one event delivered to subscribers.
type Message struct {
	Data    map[string]any `json:"data"`
	At      time.Time      `json:"at"`
	Channel string         `json:"channel"`
	Type    string         `json:"type"`
}

// Authorizer gates subscriptions: scope-level permission check per channel.
type Authorizer interface {
	// Allowed reports whether the subject (operator or service account) may
	// subscribe to the channel's scope. ErrForbidden on denial.
	Allowed(ctx context.Context, orgID, subjectID string, ch *Channel) error
}

// MemberLookup resolves org membership.
type MemberLookup interface {
	FindByOperatorAndOrg(ctx context.Context, operatorID, orgID string) (*organization.OrganizationMember, error)
}

// MembershipAuthorizer gates on org membership alone.
type MembershipAuthorizer struct {
	members MemberLookup
}

// NewMembershipAuthorizer builds the org-membership gate.
func NewMembershipAuthorizer(members MemberLookup) Authorizer {
	return &MembershipAuthorizer{members: members}
}

// Allowed checks org membership; no scope-level differentiation.
func (a *MembershipAuthorizer) Allowed(ctx context.Context, orgID, subjectID string, ch *Channel) error {
	if a.members == nil || subjectID == "" {
		return ErrForbidden
	}
	if _, err := a.members.FindByOperatorAndOrg(ctx, subjectID, orgID); err != nil {
		return ErrForbidden
	}
	return nil
}

// PermissionAuthorizer extends membership with per-scope RBAC action check.
type PermissionAuthorizer struct {
	members    MemberLookup
	permission func(ctx context.Context, orgID, subjectID, scope string) bool
}

// NewPermissionAuthorizer builds the RBAC-gated subscribe authorizer.
func NewPermissionAuthorizer(members MemberLookup, permission func(ctx context.Context, orgID, subjectID, scope string) bool) Authorizer {
	return &PermissionAuthorizer{members: members, permission: permission}
}

// Allowed checks org membership and the scope-mapped RBAC action.
func (a *PermissionAuthorizer) Allowed(ctx context.Context, orgID, subjectID string, ch *Channel) error {
	if a.members == nil || subjectID == "" {
		return ErrForbidden
	}
	if _, err := a.members.FindByOperatorAndOrg(ctx, subjectID, orgID); err != nil {
		return ErrForbidden
	}
	if a.permission != nil && !a.permission(ctx, orgID, subjectID, ch.Scope) {
		return ErrForbidden
	}
	return nil
}

// SubscribeEvent carries one subscribe attempt.
type SubscribeEvent struct {
	Channel   *Channel
	SubjectID string
}

// Stream manages one logical channel's lifecycle: first subscribe opens,
// last unsubscribe closes.
type Stream struct {
	channel     *Channel
	subscribers map[string]func(*Message)
	mu          sync.RWMutex
	active      bool
}

// NewStream creates a stream for a channel.
func NewStream(ch *Channel) *Stream {
	return &Stream{channel: ch, subscribers: map[string]func(*Message){}}
}

// Channel returns the channel this stream serves.
func (s *Stream) Channel() *Channel { return s.channel }

// Subscribe adds a callback; opens the stream on first subscriber.
func (s *Stream) Subscribe(id string, cb func(*Message)) (opened bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscribers[id] = cb
	if !s.active {
		s.active = true
		opened = true
	}
	return opened
}

// Unsubscribe removes a callback; closes the stream when empty.
func (s *Stream) Unsubscribe(id string) (closed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subscribers, id)
	if s.active && len(s.subscribers) == 0 {
		s.active = false
		closed = true
	}
	return closed
}

// Publish fans out a message to all current subscribers.
func (s *Stream) Publish(msg *Message) {
	s.mu.RLock()
	subs := make([]func(*Message), 0, len(s.subscribers))
	for _, cb := range s.subscribers {
		subs = append(subs, cb)
	}
	s.mu.RUnlock()
	for _, cb := range subs {
		cb(msg)
	}
}

// Active reports whether the stream has subscribers.
func (s *Stream) Active() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// SubscriberCount reports how many subscribers are attached.
func (s *Stream) SubscriberCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subscribers)
}

// Manager owns all streams, keyed by channel address, and routes events.
type Manager struct {
	authorizer Authorizer
	streams    map[string]*Stream
	mu         sync.RWMutex
}

// NewManager creates a stream manager with the subscribe gate.
func NewManager(authz Authorizer) *Manager {
	return &Manager{authorizer: authz, streams: map[string]*Stream{}}
}

// Subscribe checks the gate then opens/attaches to the channel's stream.
func (m *Manager) Subscribe(ctx context.Context, evt SubscribeEvent) (*Stream, error) {
	if evt.Channel == nil || evt.SubjectID == "" {
		return nil, ErrInvalidChannel
	}
	if m.authorizer == nil {
		return nil, ErrForbidden
	}
	if err := m.authorizer.Allowed(ctx, evt.Channel.OrgID, evt.SubjectID, evt.Channel); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	stream, ok := m.streams[evt.Channel.String()]
	if !ok {
		stream = NewStream(evt.Channel)
		m.streams[evt.Channel.String()] = stream
	}
	return stream, nil
}

// Unsubscribe detaches and closes streams with no remaining subscribers.
func (m *Manager) Unsubscribe(channelAddr, subscriberID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stream, ok := m.streams[channelAddr]
	if !ok {
		return
	}
	if stream.Unsubscribe(subscriberID) {
		delete(m.streams, channelAddr)
	}
}

// Publish routes a message to its channel's stream, if any.
func (m *Manager) Publish(channelAddr string, msg *Message) {
	m.mu.RLock()
	stream, ok := m.streams[channelAddr]
	m.mu.RUnlock()
	if !ok {
		return
	}
	stream.Publish(msg)
}

// StreamCount reports active streams (test/support visibility).
func (m *Manager) StreamCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.streams)
}

// MessageMapAdapter adapts the manager to the evaluator's map payload.
type MessageMapAdapter struct {
	mgr *Manager
}

// NewMessageMapAdapter creates the adapter.
func NewMessageMapAdapter(mgr *Manager) *MessageMapAdapter {
	return &MessageMapAdapter{mgr: mgr}
}

// Publish adapts map payloads to the manager's Message type on a stream address.
func (a *MessageMapAdapter) Publish(channel string, payload map[string]interface{}) {
	a.mgr.Publish(channel, &Message{Channel: channel, Data: payload})
}
