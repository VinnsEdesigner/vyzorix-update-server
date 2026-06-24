// Package websocket provides WebSocket handler implementations.
package websocket

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
)

// MessageRouter handles routing of incoming WebSocket messages.
type MessageRouter struct {
	log *slog.Logger
	hub *hub.Hub
}

// NewMessageRouter creates a new MessageRouter.
func NewMessageRouter(h *hub.Hub, log *slog.Logger) *MessageRouter {
	return &MessageRouter{
		log: log,
		hub: h,
	}
}

// HandleIncomingMessage processes incoming WebSocket messages from devices.
func (r *MessageRouter) HandleIncomingMessage(client *hub.Client, raw []byte) error {
	var env struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(raw, &env); err != nil {
		r.logBadFrame(client.DeviceID, err)
		return err
	}

	switch env.Type {
	case "telemetry":
		return r.handleTelemetry(client, raw)
	case "pong":
		return r.handlePong(client)
	case "status":
		return r.handleStatus(client, raw)
	default:
		r.logUnknownMessageType(client.DeviceID, env.Type)
	}

	return nil
}

// handleTelemetry processes telemetry frames from devices.
func (r *MessageRouter) handleTelemetry(client *hub.Client, raw []byte) error {
	var t telemetry.TelemetryFrame
	if err := json.Unmarshal(raw, &t); err != nil {
		return err
	}

	t.Raw = raw
	if t.DeviceID == "" {
		t.DeviceID = client.DeviceID
	}

	// Broadcast to dashboard
	r.hub.BroadcastTelemetry(raw)

	r.logTelemetryReceived(client.DeviceID, t.RiskScore)

	return nil
}

// handlePong handles ping/pong heartbeat responses.
func (r *MessageRouter) handlePong(client *hub.Client) error {
	return nil
}

// handleStatus processes status updates from devices.
func (r *MessageRouter) handleStatus(client *hub.Client, _ []byte) error {
	r.log.Info("status update", "deviceId", client.DeviceID)
	return nil
}

// DisconnectClient forcefully disconnects a client by device ID.
func (r *MessageRouter) DisconnectClient(deviceID string) {
	client := r.hub.GetClient(deviceID)
	if client != nil {
		r.hub.Unregister(client)

		if err := client.Conn.Close(); err != nil {
			r.logClientCloseFailed(deviceID, err)
		}
	}
}

// Logging helpers

func (r *MessageRouter) logBadFrame(deviceID string, err error) {
	if r.log != nil {
		r.log.Warn("bad websocket frame", "deviceId", deviceID, "err", err)
	}
}

func (r *MessageRouter) logUnknownMessageType(deviceID, msgType string) {
	if r.log != nil {
		r.log.Info("unknown ws message type", "deviceId", deviceID, "type", msgType)
	}
}

func (r *MessageRouter) logTelemetryReceived(deviceID string, riskScore int) {
	if r.log != nil {
		r.log.Debug("telemetry received", "deviceId", deviceID, "riskScore", riskScore)
	}
}

func (r *MessageRouter) logClientCloseFailed(deviceID string, err error) {
	if r.log != nil {
		r.log.Warn("client close failed", "deviceId", deviceID, "err", err)
	}
}

// LogTelemetrySaveFailed logs a telemetry save failure.
func (r *MessageRouter) LogTelemetrySaveFailed(ctx context.Context, deviceID string, err error) {
	if r.log != nil {
		r.log.Warn("telemetry save failed", "deviceId", deviceID, "err", err)
	}
}

// LogRateLimited logs a rate limiting event.
func (r *MessageRouter) LogRateLimited(ctx context.Context, deviceID, dispatchID string) {
	if r.log != nil {
		r.log.Debug("client rate limited, dropping message", "deviceId", deviceID, "dispatchId", dispatchID)
	}
}

// LogReadError logs a WebSocket read error.
func (r *MessageRouter) LogReadError(ctx context.Context, deviceID string, err error) {
	if r.log != nil {
		r.log.Debug("ws read error", "deviceId", deviceID, "err", err)
	}
}

// LogWriteError logs a WebSocket write error.
func (r *MessageRouter) LogWriteError(ctx context.Context, deviceID string, err error) {
	if r.log != nil {
		r.log.Warn("ws write error", "deviceId", deviceID, "err", err)
	}
}
