package hub

import (
	"sync"
	"testing"
	"time"
)

// TestSubscription_OrgFanOut_MultipleOperators proves the core fix: device,
// telemetry and command events are keyed by organization, so every operator in
// the org receives them - not only the operator that happened to trigger it.
func TestSubscription_OrgFanOut_MultipleOperators(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	const org = "org-1"

	var mu sync.Mutex
	deviceHits, telemetryHits, commandHits := 0, 0, 0

	// Two different operators subscribe to the same organization.
	unsubA := h.SubscribeDeviceUpdates(org, "dev-1", func(interface{}) error {
		mu.Lock()
		deviceHits++
		mu.Unlock()
		return nil
	})
	unsubB := h.SubscribeTelemetry(org, "dev-1", func(interface{}) error {
		mu.Lock()
		telemetryHits++
		mu.Unlock()
		return nil
	})
	unsubC := h.SubscribeCommandStatus(org, "dispatch-1", func(interface{}) error {
		mu.Lock()
		commandHits++
		mu.Unlock()
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	// Publish for the same org.
	h.PublishDeviceUpdate(org, "dev-1", map[string]any{"online": true})
	h.PublishTelemetry(org, "dev-1", map[string]any{"riskScore": 42})
	h.PublishCommandStatus(org, "dispatch-1", map[string]any{"status": "delivered"})
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if deviceHits != 1 {
		t.Errorf("expected 1 device update for the org, got %d", deviceHits)
	}
	if telemetryHits != 1 {
		t.Errorf("expected 1 telemetry event for the org, got %d", telemetryHits)
	}
	if commandHits != 1 {
		t.Errorf("expected 1 command status for the org, got %d", commandHits)
	}

	unsubA()
	unsubB()
	unsubC()
}

// TestSubscription_OrgIsolation proves an event published for one organization is
// not delivered to subscribers of another organization.
func TestSubscription_OrgIsolation(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	var mu sync.Mutex
	leaked := 0

	unsub := h.SubscribeTelemetry("org-1", "dev-1", func(interface{}) error {
		mu.Lock()
		leaked++
		mu.Unlock()
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	// Publish for a different organization.
	h.PublishTelemetry("org-2", "dev-1", map[string]any{"riskScore": 99})
	time.Sleep(250 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if leaked != 0 {
		t.Fatalf("event for org-2 leaked to an org-1 subscriber (%d deliveries)", leaked)
	}

	unsub()
}

// TestSubscription_OrgWideMatchesAnyDevice proves subscribing without a device ID
// receives events for every device in the organization.
func TestSubscription_OrgWideMatchesAnyDevice(t *testing.T) {
	h := newTestHub()
	h.InitSubscriptions()

	const org = "org-1"

	var mu sync.Mutex
	hits := 0

	unsub := h.SubscribeDeviceUpdates(org, "", func(interface{}) error {
		mu.Lock()
		hits++
		mu.Unlock()
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	h.PublishDeviceUpdate(org, "dev-a", map[string]any{"online": true})
	h.PublishDeviceUpdate(org, "dev-b", map[string]any{"online": false})
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if hits != 2 {
		t.Fatalf("org-wide subscription should receive both devices' events, got %d", hits)
	}

	unsub()
}
