package hub

import (
	"testing"
	"time"
)

// TestSubscriptionKeys_ReleasedOnUnsubscribe proves the subscription manager drops
// the map key once its last subscriber leaves. The manager is a process-lifetime
// singleton, so otherwise it accumulates a key per device/org ever subscribed to.
func TestSubscriptionKeys_ReleasedOnUnsubscribe(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	unsub := h.SubscribeDeviceUpdates("org-1", "dev-1", func(interface{}) error { return nil })
	time.Sleep(20 * time.Millisecond)

	subscriptionMgr.mu.RLock()
	keys := len(subscriptionMgr.deviceUpdates)
	subscriptionMgr.mu.RUnlock()
	if keys != 1 {
		t.Fatalf("expected 1 subscription key, got %d", keys)
	}

	unsub()

	subscriptionMgr.mu.RLock()
	keys = len(subscriptionMgr.deviceUpdates)
	subscriptionMgr.mu.RUnlock()
	if keys != 0 {
		t.Fatalf("key should be released after the last unsubscribe, got %d keys", keys)
	}
}

// TestSubscriptionKeys_KeepWhileSubscribersRemain proves removing one of several
// subscribers keeps the key (so other subscribers still receive events).
func TestSubscriptionKeys_KeepWhileSubscribersRemain(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	a := h.SubscribeDeviceUpdates("org-1", "dev-1", func(interface{}) error { return nil })
	b := h.SubscribeDeviceUpdates("org-1", "dev-1", func(interface{}) error { return nil })
	time.Sleep(20 * time.Millisecond)

	a()

	subscriptionMgr.mu.RLock()
	remaining := len(subscriptionMgr.deviceUpdates["org-1:dev-1"])
	keys := len(subscriptionMgr.deviceUpdates)
	subscriptionMgr.mu.RUnlock()
	if remaining != 1 || keys != 1 {
		t.Fatalf("expected the key to remain with 1 subscriber, got subscribers=%d keys=%d", remaining, keys)
	}

	b()

	subscriptionMgr.mu.RLock()
	keys = len(subscriptionMgr.deviceUpdates)
	subscriptionMgr.mu.RUnlock()
	if keys != 0 {
		t.Fatalf("expected the key to be released, got %d keys", keys)
	}
}

// TestSubscriptionKeys_AllMapsReleased covers the other four subscription maps.
func TestSubscriptionKeys_AllMapsReleased(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	unsubs := []func(){
		h.SubscribeTelemetry("org-1", "dev-1", func(interface{}) error { return nil }),
		h.SubscribeCommandStatus("org-1", "dispatch-1", func(interface{}) error { return nil }),
		h.SubscribeOrganizationEvents("op-1", "org-1", func(interface{}) error { return nil }),
		h.SubscribeMemberEvents("op-1", "org-1", func(interface{}) error { return nil }),
	}
	time.Sleep(20 * time.Millisecond)

	for _, u := range unsubs {
		u()
	}

	subscriptionMgr.mu.RLock()
	telemetry := len(subscriptionMgr.telemetry)
	command := len(subscriptionMgr.commandStatus)
	org := len(subscriptionMgr.orgEvents)
	member := len(subscriptionMgr.memberEvents)
	subscriptionMgr.mu.RUnlock()

	if telemetry != 0 || command != 0 || org != 0 || member != 0 {
		t.Fatalf("subscription keys leaked: telemetry=%d command=%d org=%d member=%d",
			telemetry, command, org, member)
	}
}
