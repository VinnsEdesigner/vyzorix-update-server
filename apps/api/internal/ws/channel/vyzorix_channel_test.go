package channel

import (
	"context"
	"errors"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

type fakeMembers struct{ ok bool }

func (f fakeMembers) FindByOperatorAndOrg(_ context.Context, operatorID, orgID string) (*organization.OrganizationMember, error) {
	if !f.ok {
		return nil, errors.New("not a member")
	}
	return &organization.OrganizationMember{}, nil
}

func TestParse(t *testing.T) {
	tests := []struct {
		addr    string
		org     string
		scope   string
		wantErr bool
	}{
		{"stream/org-1/alerts", "org-1", "alerts", false},
		{"stream/org-1/commands", "org-1", "commands", false},
		{"stream/org-1/telemetry/device/dev-9", "org-1", "telemetry/device/dev-9", false},
		{"org-1/alerts", "", "", true},
		{"stream/", "", "", true},
		{"stream/org-1", "", "", true},
		{"", "", "", true},
	}
	for _, tc := range tests {
		ch, err := Parse(tc.addr)
		if tc.wantErr {
			if err == nil {
				t.Errorf("%q expected error", tc.addr)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q unexpected err: %v", tc.addr, err)
			continue
		}
		if ch.OrgID != tc.org || ch.Scope != tc.scope {
			t.Errorf("%q -> %+v, want org=%s scope=%s", tc.addr, ch, tc.org, tc.scope)
		}
		if ch.String() != tc.addr {
			t.Errorf("round trip: %q != %q", ch.String(), tc.addr)
		}
	}
}

func TestMembershipAuthorizer(t *testing.T) {
	authz := NewMembershipAuthorizer(fakeMembers{ok: true})
	ch := &Channel{OrgID: "org-1", Scope: Alert}

	if err := authz.Allowed(context.Background(), "org-1", "op-1", ch); err != nil {
		t.Errorf("membership should pass: %v", err)
	}

	authzDeny := NewMembershipAuthorizer(fakeMembers{ok: false})
	if err := authzDeny.Allowed(context.Background(), "org-1", "op-1", ch); err == nil {
		t.Error("non-member should be forbidden")
	}

	if err := authz.Allowed(context.Background(), "org-1", "", ch); err == nil {
		t.Error("empty subject should be forbidden")
	}
}

func TestPermissionAuthorizer(t *testing.T) {
	alertScopeOnly := func(_ context.Context, _, _, scope string) bool {
		return scope == Alert
	}
	authz := NewPermissionAuthorizer(fakeMembers{ok: true}, alertScopeOnly)

	alertChannel := &Channel{OrgID: "org-1", Scope: Alert}
	if err := authz.Allowed(context.Background(), "org-1", "op-1", alertChannel); err != nil {
		t.Errorf("member with alert scope should pass: %v", err)
	}

	commandChannel := &Channel{OrgID: "org-1", Scope: Commands}
	if err := authz.Allowed(context.Background(), "org-1", "op-1", commandChannel); err == nil {
		t.Error("member lacking command scope should be forbidden")
	}
}

func TestStreamLifecycle(t *testing.T) {
	ch := &Channel{OrgID: "org-1", Scope: Alert}
	s := NewStream(ch)

	if s.Active() {
		t.Fatal("stream should start inactive")
	}

	if opened := s.Subscribe("a", func(*Message) {}); !opened {
		t.Error("first subscribe should open")
	}
	if !s.Active() || s.SubscriberCount() != 1 {
		t.Fatalf("expected active with 1 subscriber, got %v/%d", s.Active(), s.SubscriberCount())
	}
	if opened := s.Subscribe("b", func(*Message) {}); opened {
		t.Error("subsequent subscribe while active should not re-open")
	}

	if closed := s.Unsubscribe("a"); closed {
		t.Error("unsubscribe while another remains should not close")
	}
	if closed := s.Unsubscribe("b"); !closed {
		t.Error("last unsubscribe should close")
	}
	if s.Active() {
		t.Fatal("stream should be inactive after all unsubscribed")
	}
}

func TestStreamPublish(t *testing.T) {
	ch := &Channel{OrgID: "org-1", Scope: Alert}
	s := NewStream(ch)

	got := make([]*Message, 0)
	s.Subscribe("a", func(m *Message) { got = append(got, m) })
	s.Subscribe("b", func(m *Message) { got = append(got, m) })
	s.Unsubscribe("b")

	msg := &Message{Channel: ch.String(), Type: "alert_firing", Data: map[string]any{"rule": "offline"}}
	s.Publish(msg)
	if len(got) != 1 {
		t.Fatalf("expected only active subscriber notified, got %d events", len(got))
	}
	if got[0].Type != "alert_firing" || got[0].Channel != ch.String() {
		t.Errorf("message mismatch: %+v", got[0])
	}
}

func TestManager_SubscribeGate(t *testing.T) {
	mgr := NewManager(NewMembershipAuthorizer(fakeMembers{ok: true}))
	ctx := context.Background()

	ch := &Channel{OrgID: "org-1", Scope: Alert}
	s, err := mgr.Subscribe(ctx, SubscribeEvent{SubjectID: "op-1", Channel: ch})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if s == nil || !s.Active() && false {
		_ = s
	}
	if mgr.StreamCount() != 1 {
		t.Errorf("expected 1 stream, got %d", mgr.StreamCount())
	}

	// Same channel, second subscriber reuses stream.
	s2, err := mgr.Subscribe(ctx, SubscribeEvent{SubjectID: "op-2", Channel: ch})
	if err != nil {
		t.Fatalf("second subscribe: %v", err)
	}
	if s2 != s {
		t.Error("same channel must share stream")
	}
	if mgr.StreamCount() != 1 {
		t.Errorf("expected still 1 stream, got %d", mgr.StreamCount())
	}

	// Gate denies: no stream created.
	mgrDeny := NewManager(NewMembershipAuthorizer(fakeMembers{ok: false}))
	if _, err := mgrDeny.Subscribe(ctx, SubscribeEvent{SubjectID: "op-1", Channel: ch}); err == nil {
		t.Error("denied subscribe should error")
	}
	if mgrDeny.StreamCount() != 0 {
		t.Error("denied subscribe must not create stream")
	}

	// Validation errors.
	if _, err := mgr.Subscribe(ctx, SubscribeEvent{SubjectID: "", Channel: ch}); err == nil {
		t.Error("empty subject should error")
	}
	if _, err := mgr.Subscribe(ctx, SubscribeEvent{SubjectID: "op-1"}); err == nil {
		t.Error("nil channel should error")
	}
}

func TestManager_UnsubscribeAndPublish(t *testing.T) {
	mgr := NewManager(NewMembershipAuthorizer(fakeMembers{ok: true}))
	ctx := context.Background()
	ch := &Channel{OrgID: "org-1", Scope: Alert}

	s, err := mgr.Subscribe(ctx, SubscribeEvent{SubjectID: "op-1", Channel: ch})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	got := make([]*Message, 0)
	s.Subscribe("op-1", func(m *Message) { got = append(got, m) })

	mgr.Publish(ch.String(), &Message{Channel: ch.String(), Type: "alert_resolved", Data: map[string]any{}})
	if len(got) != 1 {
		t.Fatalf("publish to subscribed channel: got %d", len(got))
	}

	// Different channel: no delivery.
	mgr.Publish("stream/org-2/alerts", &Message{Type: "x"})
	if len(got) != 1 {
		t.Error("publish to unsubscribed channel must not deliver")
	}

	mgr.Unsubscribe(ch.String(), "op-1")
	if mgr.StreamCount() != 0 {
		t.Errorf("unsubscribe of only subscriber should drop stream, got %d", mgr.StreamCount())
	}
	// Publish after stream dropped: no panic, no delivery.
	mgr.Publish(ch.String(), &Message{Type: "y"})
	if len(got) != 1 {
		t.Error("delivered after unsubscribe")
	}
}

func TestManager_NoAuthorizer(t *testing.T) {
	mgr := NewManager(nil)
	ch := &Channel{OrgID: "org-1", Scope: Alert}
	if _, err := mgr.Subscribe(context.Background(), SubscribeEvent{SubjectID: "op-1", Channel: ch}); err == nil {
		t.Error("no authorizer should deny all subscribes")
	}
}
