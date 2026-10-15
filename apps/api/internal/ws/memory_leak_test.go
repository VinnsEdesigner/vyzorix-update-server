package hub

import (
	"context"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
)

func frame(dispatchID string) command.CommandFrame {
	return command.CommandFrame{Type: "command", DispatchID: dispatchID}
}

// TestRemoveQueue_DropsMapEntry proves a device's queue is released so the
// per-device queue map does not grow for the lifetime of the process.
func TestRemoveQueue_DropsMapEntry(t *testing.T) {
	q := NewMessageQueue(testLogger(), nil, DefaultMessageQueueConfig())

	q.GetOrCreateQueue("dev-1")
	q.GetOrCreateQueue("dev-2")
	if got := q.QueueCount(); got != 2 {
		t.Fatalf("expected 2 queues, got %d", got)
	}

	q.RemoveQueue("dev-1")
	if got := q.QueueCount(); got != 1 {
		t.Fatalf("expected 1 queue after removal, got %d", got)
	}

	// Removing a device with no queue must be a no-op, not a panic.
	q.RemoveQueue("dev-never-seen")
	if got := q.QueueCount(); got != 1 {
		t.Fatalf("expected 1 queue, got %d", got)
	}
}

// TestRemoveQueue_ThenEnqueueRecreates proves RemoveQueue leaves the queue usable:
// enqueueing again recreates it rather than panicking on a closed channel.
func TestRemoveQueue_ThenEnqueueRecreates(t *testing.T) {
	q := NewMessageQueue(testLogger(), nil, DefaultMessageQueueConfig())

	q.GetOrCreateQueue("dev-1")
	q.RemoveQueue("dev-1")

	if ok := q.Enqueue("dev-1", frame("dispatch-1")); !ok {
		t.Fatal("enqueue after RemoveQueue should succeed by recreating the queue")
	}
	if got := q.QueueCount(); got != 1 {
		t.Fatalf("expected the queue to be recreated, count=%d", got)
	}
}

// TestRemoveQueue_ConcurrentEnqueue is the regression guard for why the channel
// must not be closed: Enqueue sends without holding the queue lock, so closing it
// would panic with "send on closed channel". Meaningful under -race.
func TestRemoveQueue_ConcurrentEnqueue(t *testing.T) {
	q := NewMessageQueue(testLogger(), nil, DefaultMessageQueueConfig())

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 3000; i++ {
			q.Enqueue("dev-1", frame("d"))
		}
	}()

	for i := 0; i < 3000; i++ {
		q.RemoveQueue("dev-1")
	}

	<-done
}

// TestUnregisterReleasesPerDeviceState proves the disconnect path frees the
// per-device maps (latency + queue) rather than retaining an entry forever.
func TestUnregisterReleasesPerDeviceState(t *testing.T) {
	h := newTestHub()
	h.latencyConfig = &LatencyConfig{Enabled: true, SampleRate: 1.0}
	h.messageQueue = NewMessageQueue(testLogger(), nil, DefaultMessageQueueConfig())

	// Record latency and create a queue for the device.
	h.trackLatency("dev-1", "dispatch-1", time.Now().Add(-5*time.Millisecond), true)
	h.messageQueue.GetOrCreateQueue("dev-1")

	h.metricsMu.RLock()
	if len(h.deviceLatency) == 0 {
		h.metricsMu.RUnlock()
		t.Fatal("expected a deviceLatency entry")
	}
	h.metricsMu.RUnlock()
	if h.messageQueue.QueueCount() != 1 {
		t.Fatal("expected one queue")
	}

	// Run the real unregister cleanup.
	h.handleClientUnregistration(context.Background(), &Client{DeviceID: "dev-1", Send: make(chan command.CommandFrame, 4)})

	h.metricsMu.RLock()
	latency := len(h.deviceLatency)
	h.metricsMu.RUnlock()
	if latency != 0 {
		t.Fatalf("deviceLatency should be empty after unregister, got %d", latency)
	}
	if got := h.messageQueue.QueueCount(); got != 0 {
		t.Fatalf("queue should be released after unregister, got %d", got)
	}
}
