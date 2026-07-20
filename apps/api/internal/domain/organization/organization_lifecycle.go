package organization

import (
	"errors"
	"fmt"
	"time"
)

// OrganizationLifecycle represents the lifecycle state of an organization.
type OrganizationLifecycle string

const (
	// OrganizationLifecycleActive represents an active organization.
	OrganizationLifecycleActive OrganizationLifecycle = "active"
	// OrganizationLifecycleInactive represents an inactive organization (suspended).
	OrganizationLifecycleInactive OrganizationLifecycle = "inactive"
	// OrganizationLifecycleArchived represents a soft-deleted organization.
	OrganizationLifecycleArchived OrganizationLifecycle = "archived"
)

// ErrInvalidOrganizationTransition is returned when an organization lifecycle transition is not allowed.
var ErrInvalidOrganizationTransition = errors.New("invalid organization lifecycle transition")

// OrganizationLifecycleTransitions defines valid state transitions for organization lifecycle.
var OrganizationLifecycleTransitions = map[OrganizationLifecycle]map[OrganizationLifecycle]bool{
	OrganizationLifecycleActive: {
		OrganizationLifecycleInactive: true,
		OrganizationLifecycleArchived: true,
	},
	OrganizationLifecycleInactive: {
		OrganizationLifecycleActive:  true,
		OrganizationLifecycleArchived: true,
	},
	OrganizationLifecycleArchived: {}, // No transitions allowed from archived
}

// CanTransitionTo returns true if the lifecycle can transition to the target state.
func (l OrganizationLifecycle) CanTransitionTo(target OrganizationLifecycle) bool {
	allowed, exists := OrganizationLifecycleTransitions[l]
	if !exists {
		return false
	}
	return allowed[target]
}

// TransitionTo transitions the lifecycle to a new state.
// Returns ErrInvalidOrganizationTransition if the transition is not allowed.
func (l *OrganizationLifecycle) TransitionTo(target OrganizationLifecycle) error {
	if !l.CanTransitionTo(target) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidOrganizationTransition, *l, target)
	}
	*l = target
	return nil
}

// IsActive returns true if the organization is active.
func (l OrganizationLifecycle) IsActive() bool {
	return l == OrganizationLifecycleActive
}

// IsInactive returns true if the organization is inactive.
func (l OrganizationLifecycle) IsInactive() bool {
	return l == OrganizationLifecycleInactive
}

// IsArchived returns true if the organization is archived.
func (l OrganizationLifecycle) IsArchived() bool {
	return l == OrganizationLifecycleArchived
}

// CanAcceptMembers returns true if the organization can accept new members.
func (l OrganizationLifecycle) CanAcceptMembers() bool {
	return l == OrganizationLifecycleActive
}

// String returns the string representation of the lifecycle.
func (l OrganizationLifecycle) String() string {
	return string(l)
}

// Format implements fmt.Formatter for pretty printing.
func (l OrganizationLifecycle) Format(s fmt.State, verb rune) {
	format := "%s"
	if verb == 'v' && s.Flag('#') {
		format = "%q"
	}
	_, _ = fmt.Fprintf(s, format, string(l))
}

// MemberLifecycle represents the lifecycle state of an organization member.
type MemberLifecycle string

const (
	// MemberLifecycleInvited represents a member who has been invited but hasn't joined.
	MemberLifecycleInvited MemberLifecycle = "invited"
	// MemberLifecycleActive represents an active member.
	MemberLifecycleActive MemberLifecycle = "active"
	// MemberLifecycleSuspended represents a suspended member.
	MemberLifecycleSuspended MemberLifecycle = "suspended"
	// MemberLifecycleRemoved represents a removed member (soft delete).
	MemberLifecycleRemoved MemberLifecycle = "removed"
)

// ErrInvalidMemberTransition is returned when a member lifecycle transition is not allowed.
var ErrInvalidMemberTransition = errors.New("invalid member lifecycle transition")

// MemberLifecycleTransitions defines valid state transitions for member lifecycle.
var MemberLifecycleTransitions = map[MemberLifecycle]map[MemberLifecycle]bool{
	MemberLifecycleInvited: {
		MemberLifecycleActive:    true,
		MemberLifecycleRemoved:   true,
	},
	MemberLifecycleActive: {
		MemberLifecycleSuspended: true,
		MemberLifecycleRemoved:   true,
	},
	MemberLifecycleSuspended: {
		MemberLifecycleActive:  true,
		MemberLifecycleRemoved: true,
	},
	MemberLifecycleRemoved: {}, // No transitions allowed from removed
}

// CanTransitionTo returns true if the lifecycle can transition to the target state.
func (l MemberLifecycle) CanTransitionTo(target MemberLifecycle) bool {
	allowed, exists := MemberLifecycleTransitions[l]
	if !exists {
		return false
	}
	return allowed[target]
}

// TransitionTo transitions the lifecycle to a new state.
// Returns ErrInvalidMemberTransition if the transition is not allowed.
func (l *MemberLifecycle) TransitionTo(target MemberLifecycle) error {
	if !l.CanTransitionTo(target) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidMemberTransition, *l, target)
	}
	*l = target
	return nil
}

// IsInvited returns true if the member is invited.
func (l MemberLifecycle) IsInvited() bool {
	return l == MemberLifecycleInvited
}

// IsActive returns true if the member is active.
func (l MemberLifecycle) IsActive() bool {
	return l == MemberLifecycleActive
}

// IsSuspended returns true if the member is suspended.
func (l MemberLifecycle) IsSuspended() bool {
	return l == MemberLifecycleSuspended
}

// IsRemoved returns true if the member is removed.
func (l MemberLifecycle) IsRemoved() bool {
	return l == MemberLifecycleRemoved
}

// CanAccessResources returns true if the member can access organization resources.
func (l MemberLifecycle) CanAccessResources() bool {
	return l == MemberLifecycleActive
}

// String returns the string representation of the lifecycle.
func (l MemberLifecycle) String() string {
	return string(l)
}

// Format implements fmt.Formatter for pretty printing.
func (l MemberLifecycle) Format(s fmt.State, verb rune) {
	format := "%s"
	if verb == 'v' && s.Flag('#') {
		format = "%q"
	}
	_, _ = fmt.Fprintf(s, format, string(l))
}

// MemberTransitionHelper provides methods for member lifecycle state changes.
type MemberTransitionHelper struct {
	Lifecycle       MemberLifecycle
	RemovedAt       *time.Time
	SuspendedAt     *time.Time
}

// Invite creates an invited member from scratch.
func (m *MemberTransitionHelper) Invite() {
	m.Lifecycle = MemberLifecycleInvited
	m.RemovedAt = nil
	m.SuspendedAt = nil
}

// Join transitions from invited to active.
func (m *MemberTransitionHelper) Join() error {
	return m.Lifecycle.TransitionTo(MemberLifecycleActive)
}

// Suspend transitions from active to suspended.
func (m *MemberTransitionHelper) Suspend() error {
	if err := m.Lifecycle.TransitionTo(MemberLifecycleSuspended); err != nil {
		return err
	}
	now := time.Now()
	m.SuspendedAt = &now
	return nil
}

// Reinstate transitions from suspended back to active.
func (m *MemberTransitionHelper) Reinstate() error {
	return m.Lifecycle.TransitionTo(MemberLifecycleActive)
}

// Remove transitions to removed state.
func (m *MemberTransitionHelper) Remove() error {
	if err := m.Lifecycle.TransitionTo(MemberLifecycleRemoved); err != nil {
		return err
	}
	now := time.Now()
	m.RemovedAt = &now
	return nil
}
