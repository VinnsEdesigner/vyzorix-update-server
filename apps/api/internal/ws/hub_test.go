package hub

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
)

func TestNew(t *testing.T) {
	logger := slog.Default()
	h := New(logger, nil, nil)

	if h == nil {
		t.Fatal("New() returned nil")
	}
	if h.clients == nil {
		t.Error("clients map not initialized")
	}
	if h.register == nil {
		t.Error("register channel not initialized")
	}
	if h.unreg == nil {
		t.Error("unreg channel not initialized")
	}
	if h.broadcast == nil {
		t.Error("broadcast channel not initialized")
	}
}

func TestHub_Send_noClient(t *testing.T) {
	h := New(nil, nil, nil)

	frame := command.CommandFrame{
		Type:       "update",
		DispatchID: "dispatch-001",
	}

	// Should return false when no client connected.
	if h.Send("nonexistent-device", frame) {
		t.Error("Send() should return false for nonexistent device")
	}
}

func TestHub_Online_noClient(t *testing.T) {
	h := New(nil, nil, nil)

	if h.Online("nonexistent") {
		t.Error("Online() should return false for nonexistent device")
	}
}

func TestHub_ClientCount_empty(t *testing.T) {
	h := New(nil, nil, nil)

	if h.ClientCount() != 0 {
		t.Errorf("ClientCount() = %d, want 0", h.ClientCount())
	}
}

func TestHub_Clients_empty(t *testing.T) {
	h := New(nil, nil, nil)

	clients := h.Clients()
	if len(clients) != 0 {
		t.Errorf("Clients() length = %d, want 0", len(clients))
	}
}

func TestHub_GetClient_noClient(t *testing.T) {
	h := New(nil, nil, nil)

	c := h.GetClient("nonexistent")
	if c != nil {
		t.Error("GetClient() should return nil for nonexistent device")
	}
}

func TestHub_Run_contextCancel(t *testing.T) {
	logger := slog.Default()
	h := New(logger, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Run should return immediately when context is cancelled.
	done := make(chan struct{})
	go func() {
		h.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Success - Run returned when context was cancelled.
	case <-time.After(200 * time.Millisecond):
		t.Error("Run() did not exit when context was cancelled")
	}
}

func TestCommandFrame_Fields(t *testing.T) {
	frame := command.CommandFrame{
		Type:       "update",
		DispatchID: "dispatch-001",
		Args:       []byte(`{"version":"2.0.0"}`),
	}

	if frame.Type != "update" {
		t.Errorf("Type = %s, want update", frame.Type)
	}
	if frame.DispatchID != "dispatch-001" {
		t.Errorf("DispatchID = %s, want dispatch-001", frame.DispatchID)
	}
}

func TestTelemetryFrame_Fields(t *testing.T) {
	frame := telemetry.TelemetryFrame{
		DeviceID:    "device-001",
		RiskScore:   25,
		BufferLevel: 60,
		ThermalTemp: 35.5,
	}

	if frame.DeviceID != "device-001" {
		t.Errorf("DeviceID = %s, want device-001", frame.DeviceID)
	}
	if frame.RiskScore != 25 {
		t.Errorf("RiskScore = %d, want 25", frame.RiskScore)
	}
	if frame.BufferLevel != 60 {
		t.Errorf("BufferLevel = %d, want 60", frame.BufferLevel)
	}
	if frame.ThermalTemp != 35.5 {
		t.Errorf("ThermalTemp = %f, want 35.5", frame.ThermalTemp)
	}
}

func TestClient_SendChannel(t *testing.T) {
	c := &Client{
		DeviceID: "device-001",
		Send:     make(chan command.CommandFrame, 5),
	}

	// Should be able to send to channel.
	select {
	case c.Send <- command.CommandFrame{Type: "test"}:
		// Success.
	default:
		t.Error("failed to send to client channel")
	}
}

func TestClient_SendChannel_full(t *testing.T) {
	c := &Client{
		DeviceID: "device-001",
		Send:     make(chan command.CommandFrame, 1), // Buffer of 1
	}

	// Fill the buffer.
	c.Send <- command.CommandFrame{Type: "first"}

	// This should block since buffer is full.
	select {
	case c.Send <- command.CommandFrame{Type: "second"}:
		// If we get here immediately, channel buffer isn't working.
		t.Error("send should block on full channel")
	case <-time.After(10 * time.Millisecond):
		// Expected - send blocked.
	}
}
