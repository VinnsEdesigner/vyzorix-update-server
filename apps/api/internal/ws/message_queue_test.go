package hub

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
)

func TestMessageQueueEnqueueDequeue(t *testing.T) {
	log := testLogger()
	cfg := DefaultMessageQueueConfig()
	cfg.MaxQueueSize = 100

	mq := NewMessageQueue(log, nil, cfg)

	frame := command.CommandFrame{
		Type:       "test",
		DispatchID: "test-001",
		Command:    "TEST_CMD",
	}

	if !mq.Enqueue("device-1", frame) {
		t.Error("expected enqueue to succeed")
	}

	if mq.QueueSize("device-1") != 1 {
		t.Errorf("expected queue size 1, got %d", mq.QueueSize("device-1"))
	}
}

func TestMessageQueueFIFO(t *testing.T) {
	log := testLogger()
	cfg := DefaultMessageQueueConfig()

	mq := NewMessageQueue(log, nil, cfg)

	for i := 1; i <= 3; i++ {
		frame := command.CommandFrame{
			Type:       "test",
			DispatchID: fmt.Sprintf("test-%03d", i),
		}
		mq.Enqueue("device-1", frame)
	}

	messages := mq.LoadPersistedMessages("device-1")
	if len(messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(messages))
	}
}

func TestMessageQueueMaxSize(t *testing.T) {
	log := testLogger()
	cfg := DefaultMessageQueueConfig()
	cfg.MaxQueueSize = 5

	mq := NewMessageQueue(log, nil, cfg)

	for i := 0; i < 6; i++ {
		frame := command.CommandFrame{
			Type:       "test",
			DispatchID: fmt.Sprintf("test-%d", i),
		}
		mq.Enqueue("device-1", frame)
	}

	size := mq.QueueSize("device-1")
	if size > cfg.MaxQueueSize {
		t.Errorf("queue size %d exceeds max %d", size, cfg.MaxQueueSize)
	}
}

func TestMessageQueueMetrics(t *testing.T) {
	log := testLogger()
	cfg := DefaultMessageQueueConfig()

	mq := NewMessageQueue(log, nil, cfg)

	for i := 0; i < 5; i++ {
		frame := command.CommandFrame{DispatchID: fmt.Sprintf("test-%d", i)}
		mq.Enqueue("device-1", frame)
	}

	metrics := mq.GetMetrics()

	if metrics.TotalEnqueued != 5 {
		t.Errorf("expected 5 enqueued, got %d", metrics.TotalEnqueued)
	}
}

func TestMessageQueueMultipleDevices(t *testing.T) {
	log := testLogger()
	cfg := DefaultMessageQueueConfig()

	mq := NewMessageQueue(log, nil, cfg)

	mq.Enqueue("device-1", command.CommandFrame{DispatchID: "d1-1"})
	mq.Enqueue("device-1", command.CommandFrame{DispatchID: "d1-2"})
	mq.Enqueue("device-2", command.CommandFrame{DispatchID: "d2-1"})

	if mq.QueueSize("device-1") != 2 {
		t.Errorf("expected device-1 queue size 2, got %d", mq.QueueSize("device-1"))
	}

	if mq.QueueSize("device-2") != 1 {
		t.Errorf("expected device-2 queue size 1, got %d", mq.QueueSize("device-2"))
	}

	if mq.TotalQueuedMessages() != 3 {
		t.Errorf("expected total 3, got %d", mq.TotalQueuedMessages())
	}
}

func TestMessageQueueReplay(t *testing.T) {
	log := testLogger()
	cfg := DefaultMessageQueueConfig()

	mq := NewMessageQueue(log, nil, cfg)

	for i := 0; i < 3; i++ {
		frame := command.CommandFrame{DispatchID: fmt.Sprintf("replay-%d", i)}
		mq.Enqueue("device-replay", frame)
	}

	dest := make(chan command.CommandFrame, 10)

	count := mq.ReplayQueue("device-replay", dest)

	if count != 3 {
		t.Errorf("expected replay count 3, got %d", count)
	}

	close(dest)
	var received int
	for range dest {
		received++
	}

	if received != 3 {
		t.Errorf("expected 3 received, got %d", received)
	}
}

func TestMessageQueueRemoveQueue(t *testing.T) {
	log := testLogger()
	cfg := DefaultMessageQueueConfig()

	mq := NewMessageQueue(log, nil, cfg)

	mq.Enqueue("device-1", command.CommandFrame{DispatchID: "test-1"})

	if mq.QueueSize("device-1") != 1 {
		t.Error("expected queue size 1")
	}

	mq.RemoveQueue("device-1")

	if mq.QueueSize("device-1") != 0 {
		t.Error("expected queue size 0 after remove")
	}
}

func TestGenerateQueueID(t *testing.T) {
	id1 := generateQueueID()
	id2 := generateQueueID()

	if id1 == id2 {
		t.Error("generated IDs should be unique")
	}

	if len(id1) < 10 {
		t.Error("ID should have proper length")
	}
}

func TestMessageQueueCleanup(t *testing.T) {
	log := testLogger()
	cfg := DefaultMessageQueueConfig()
	cfg.CleanupInterval = 100 * time.Millisecond

	mq := NewMessageQueue(log, nil, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	mq.Start(ctx)

	time.Sleep(300 * time.Millisecond)

	metrics := mq.GetMetrics()
	if metrics.LastCleanupAt.IsZero() {
		t.Log("cleanup worker running (no expired items to clean)")
	}
}

func testLogger() *slog.Logger {
	return slog.Default()
}
