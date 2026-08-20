package hub

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/google/uuid"
)

// TestHubClients_EmptyAfterInit verifies that a freshly created hub reports zero
// connected clients before any registration happens.
func TestHubClients_EmptyAfterInit(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	if count := h.ClientCount(); count != 0 {
		t.Errorf("ClientCount = %d, want 0", count)
	}
	if clients := h.Clients(); len(clients) != 0 {
		t.Errorf("Clients() = %d entries, want 0", len(clients))
	}
}

// TestHubGetClient_NotFound verifies that retrieving a non-existent device
// returns nil rather than panicking.
func TestHubGetClient_NotFound(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	c := h.GetClient("nonexistent-device")
	if c != nil {
		t.Errorf("GetClient for nonexistent device should return nil")
	}
}

// TestClientMetrics_FullLifecycle exercises the Client metric recording
// methods: attempt → success → messages → rate limit → disconnect.
func TestClientMetrics_FullLifecycle(t *testing.T) {
	c := &Client{
		DeviceID:    "test-device",
		Send:        make(chan command.CommandFrame, 1),
		Hub:         nil,
		ClientID:    uuid.New().String(),
		Done:        make(chan struct{}),
		isConnected: atomic.Bool{},
	}

	c.RecordConnectAttempt()
	c.RecordConnectSuccess()
	if !c.IsConnected() {
		t.Error("should be connected after success")
	}
	c.RecordMessageSent()
	c.RecordMessageReceived()
	c.RecordPongMissed()
	c.RecordRateLimited()

	m := c.GetMetrics()
	if m.ConnectAttempts != 1 {
		t.Errorf("ConnectAttempts = %d, want 1", m.ConnectAttempts)
	}
	if m.ConnectSuccesses != 1 {
		t.Errorf("ConnectSuccesses = %d, want 1", m.ConnectSuccesses)
	}
	if m.MessagesSent != 1 {
		t.Errorf("MessagesSent = %d, want 1", m.MessagesSent)
	}
	if m.MessagesReceived != 1 {
		t.Errorf("MessagesReceived = %d, want 1", m.MessagesReceived)
	}
	if m.PongMissedCount != 1 {
		t.Errorf("PongMissedCount = %d, want 1", m.PongMissedCount)
	}
	if m.RateLimitedCount != 1 {
		t.Errorf("RateLimitedCount = %d, want 1", m.RateLimitedCount)
	}
	if m.LastRateLimitedAt == 0 {
		t.Error("LastRateLimitedAt should be set")
	}

	c.RecordConnectFailure()
	c.RecordDisconnect()
	if c.IsConnected() {
		t.Error("should be disconnected after disconnect")
	}

	m2 := c.GetMetrics()
	if m2.ConnectFailures != 1 {
		t.Errorf("ConnectFailures = %d, want 1", m2.ConnectFailures)
	}
	if m2.LastDisconnectedAt == 0 {
		t.Error("LastDisconnectedAt should be set")
	}
}

// TestClientUptime_AfterConnect verifies that Uptime returns a non-zero value
// after a successful connection, and zero before connecting.
func TestClientUptime_AfterConnect(t *testing.T) {
	c := &Client{
		DeviceID:    "uptime-test",
		Send:        make(chan command.CommandFrame, 1),
		ClientID:    uuid.New().String(),
		Done:        make(chan struct{}),
		isConnected: atomic.Bool{},
	}

	if uptime := c.Uptime(); uptime != 0 {
		t.Errorf("Uptime before connect = %d, want 0", uptime)
	}

	c.RecordConnectSuccess()
	if !c.IsConnected() {
		t.Fatal("should be connected after RecordConnectSuccess")
	}
	// Uptime uses Unix seconds; sleep to ensure > 0.
	time.Sleep(1100 * time.Millisecond)
	uptime := c.Uptime()
	if uptime <= 0 {
		t.Errorf("Uptime after connect = %d, want > 0", uptime)
	}
}

// TestHubBroadcastTelemetryToFiltered_NoFilter verifies that when no telemetry
// filter is configured, BroadcastTelemetryToFiltered falls back to a plain
// broadcast. This test registers a client and verifies it receives telemetry.
func TestHubBroadcastTelemetryToFiltered_NoFilter(t *testing.T) {
	h := newTestHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go h.Run(ctx)
	time.Sleep(50 * time.Millisecond)

	c := &Client{
		DeviceID: "dev-listen",
		Send:     make(chan command.CommandFrame, 32),
		Hub:      h,
		Done:     make(chan struct{}),
	}
	h.Register(c)
	time.Sleep(100 * time.Millisecond)

	data := []byte(`{"type":"telemetry","battery":80}`)

	// With no filter, this broadcasts to all clients via the hub's broadcast
	// handler, which forwards to each client's Send channel.
	h.BroadcastTelemetryToFiltered("device-1", data)

	// The broadcast handler in Run() forwards to clients; verify the data
	// arrived in the client's Send channel.
	select {
	case frame := <-c.Send:
		// When no filter is configured, BroadcastTelemetryToFiltered calls
		// BroadcastTelemetry which goes through the broadcast handler as type
		// "broadcast". If a filter IS configured, it sends directly as "telemetry".
		if frame.Type != "broadcast" && frame.Type != "telemetry" {
			t.Errorf("frame type = %s, want broadcast or telemetry", frame.Type)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("telemetry was not forwarded to client within timeout")
	}
}
