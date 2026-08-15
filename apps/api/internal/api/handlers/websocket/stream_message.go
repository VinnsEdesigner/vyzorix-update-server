// Package websocket provides WebSocket handler implementations.
package websocket

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
)

// WSSubscribePayload represents the payload for SUBSCRIBE messages.
type WSSubscribePayload struct {
	DeviceID string `json:"deviceId"`
}

// WSUnsubscribePayload represents the payload for UNSUBSCRIBE messages.
type WSUnsubscribePayload struct {
	DeviceID string `json:"deviceId"`
}

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
	case "SUBSCRIBE":
		return r.handleSubscribe(client, raw)
	case "UNSUBSCRIBE":
		return r.handleUnsubscribe(client, raw)
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

	// Broadcast to dashboard with filtering based on subscriptions.
	// BroadcastTelemetryToFiltered sends only to clients subscribed to this device.
	r.hub.BroadcastTelemetryToFiltered(t.DeviceID, raw)

	r.logTelemetryReceived(client.DeviceID, t.RiskScore)

	return nil
}

// handleSubscribe processes SUBSCRIBE messages from dashboard clients.
// Subscribes the client to receive telemetry from specific devices.
func (r *MessageRouter) handleSubscribe(client *hub.Client, raw []byte) error {
	var msg struct {
		Type    string             `json:"type"`
		Payload WSSubscribePayload `json:"payload"`
	}

	if err := json.Unmarshal(raw, &msg); err != nil {
		r.logBadFrame(client.ClientID, err)
		return err
	}

	deviceID := msg.Payload.DeviceID
	if deviceID == "" {
		r.sendSubscribeAck(client, "", false, "missing deviceId")
		return nil
	}

	var success bool
	var errorMsg string

	// Wildcard subscription - client wants all telemetry (dashboard mode).
	// No subscriptions = receives all (TelemetryFilter.ShouldForward returns true when no subscriptions).
	if deviceID == "*" {
		if tf := r.hub.TelemetryFilter(); tf != nil {
			tf.UnsubscribeAll(client.ClientID)
		}
		r.log.Info("client subscribed to all devices (dashboard mode)", "clientId", client.ClientID)
		success = true
	} else {
		// Subscribe client to specific device.
		success = r.hub.Subscribe(client.ClientID, deviceID)
		if success {
			r.log.Info("client subscribed to device",
				"clientId", client.ClientID,
				"deviceId", deviceID,
			)
		} else {
			errorMsg = "failed to subscribe - check max subscriptions limit"
			r.log.Warn("failed to subscribe client to device",
				"clientId", client.ClientID,
				"deviceId", deviceID,
			)
		}
	}

	// Send acknowledgment.
	r.sendSubscribeAck(client, deviceID, success, errorMsg)

	return nil
}

// handleUnsubscribe processes UNSUBSCRIBE messages from dashboard clients.
// Unsubscribes the client from receiving telemetry from specific devices.
func (r *MessageRouter) handleUnsubscribe(client *hub.Client, raw []byte) error {
	var msg struct {
		Type    string               `json:"type"`
		Payload WSUnsubscribePayload `json:"payload"`
	}

	if err := json.Unmarshal(raw, &msg); err != nil {
		r.logBadFrame(client.ClientID, err)
		return err
	}

	deviceID := msg.Payload.DeviceID
	if deviceID == "" {
		r.sendUnsubscribeAck(client, "", false, "missing deviceId")
		return nil
	}

	var success bool
	var errorMsg string

	// Wildcard unsubscribe - unsubscribe from all devices.
	if deviceID == "*" {
		if tf := r.hub.TelemetryFilter(); tf != nil {
			tf.UnsubscribeAll(client.ClientID)
		}
		r.log.Info("client unsubscribed from all devices", "clientId", client.ClientID)
		success = true
	} else {
		// Unsubscribe client from specific device.
		success = r.hub.Unsubscribe(client.ClientID, deviceID)
		if success {
			r.log.Info("client unsubscribed from device",
				"clientId", client.ClientID,
				"deviceId", deviceID,
			)
		} else {
			errorMsg = "client was not subscribed to this device"
			r.log.Warn("client was not subscribed to device",
				"clientId", client.ClientID,
				"deviceId", deviceID,
			)
		}
	}

	// Send acknowledgment.
	r.sendUnsubscribeAck(client, deviceID, success, errorMsg)

	return nil
}

// WSAckPayload represents the acknowledgment payload for subscription responses.
type WSAckPayload struct {
	DeviceID string `json:"deviceId"`
	Error    string `json:"error,omitempty"`
	Success  bool   `json:"success"`
}

// sendSubscribeAck sends a SUBSCRIBE_ACK message to the client.
func (r *MessageRouter) sendSubscribeAck(client *hub.Client, deviceID string, success bool, errorMsg string) {
	ack := struct {
		Type    string       `json:"type"`
		Payload WSAckPayload `json:"payload"`
	}{
		Type:    "SUBSCRIBE_ACK",
		Payload: WSAckPayload{Success: success, DeviceID: deviceID, Error: errorMsg},
	}

	if err := client.Conn.WriteJSON(ack); err != nil {
		r.log.Warn("failed to send SUBSCRIBE_ACK", "clientId", client.ClientID, "err", err)
	}
}

// sendUnsubscribeAck sends a UNSUBSCRIBE_ACK message to the client.
func (r *MessageRouter) sendUnsubscribeAck(client *hub.Client, deviceID string, success bool, errorMsg string) {
	ack := struct {
		Type    string       `json:"type"`
		Payload WSAckPayload `json:"payload"`
	}{
		Type:    "UNSUBSCRIBE_ACK",
		Payload: WSAckPayload{Success: success, DeviceID: deviceID, Error: errorMsg},
	}

	if err := client.Conn.WriteJSON(ack); err != nil {
		r.log.Warn("failed to send UNSUBSCRIBE_ACK", "clientId", client.ClientID, "err", err)
	}
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

// Logging helpers.

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
