package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix-update-server/config"
	"github.com/VinnsEdesigner/vyzorix-update-server/hub"
	"github.com/VinnsEdesigner/vyzorix-update-server/models"
	"github.com/VinnsEdesigner/vyzorix-update-server/security"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// UpgraderFactory creates and configures WebSocket upgraders with consistent settings.
type UpgraderFactory struct {
	originValidator *security.OriginValidator
	handshakeTimeout time.Duration
}

// NewUpgraderFactory creates a new UpgraderFactory with the given origin validator.
func NewUpgraderFactory(originValidator *security.OriginValidator) *UpgraderFactory {
	return &UpgraderFactory{
		originValidator: originValidator,
		handshakeTimeout: 10 * time.Second,
	}
}

// SetHandshakeTimeout sets the WebSocket handshake timeout.
func (f *UpgraderFactory) SetHandshakeTimeout(timeout time.Duration) *UpgraderFactory {
	f.handshakeTimeout = timeout
	return f
}

// Create returns a configured websocket.Upgrader.
func (f *UpgraderFactory) Create() websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin:     f.originValidator.CheckOrigin(),
		HandshakeTimeout: f.handshakeTimeout,
	}
}

// WebSocketHandler manages WebSocket upgrade and client lifecycle.
type WebSocketHandler struct {
	log             *slog.Logger
	config          config.Config
	hub             *hub.Hub
	hmac            security.Verifier
	upgrader        websocket.Upgrader
	originValidator *security.OriginValidator
}

func NewWebSocketHandler(
	log *slog.Logger,
	cfg config.Config,
	h *hub.Hub,
	hmac security.Verifier,
) *WebSocketHandler {
	// Initialize origin validator
	originValidator := security.NewOriginValidator(cfg.AllowedOrigins)
	originValidator.SetLogger(log)

	return &WebSocketHandler{
		log:             log,
		config:          cfg,
		hub:             h,
		hmac:            hmac,
		originValidator: originValidator,
		upgrader: websocket.Upgrader{
			CheckOrigin:     originValidator.CheckOrigin(),
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// NewWebSocketHandlerWithValidator creates a WebSocketHandler with a pre-configured OriginValidator.
// Use this when you need to share the same validator instance across handlers.
func NewWebSocketHandlerWithValidator(
	log *slog.Logger,
	cfg config.Config,
	h *hub.Hub,
	hmac security.Verifier,
	originValidator *security.OriginValidator,
) *WebSocketHandler {
	// Use the UpgraderFactory for consistent configuration
	factory := NewUpgraderFactory(originValidator)
	return &WebSocketHandler{
		log:             log,
		config:          cfg,
		hub:             h,
		hmac:            hmac,
		originValidator: originValidator,
		upgrader:        factory.Create(),
	}
}

// NewWebSocketHandlerWithFactory creates a WebSocketHandler using a shared UpgraderFactory.
// This ensures all WebSocket handlers use the same configuration.
func NewWebSocketHandlerWithFactory(
	log *slog.Logger,
	cfg config.Config,
	h *hub.Hub,
	hmac security.Verifier,
	factory *UpgraderFactory,
) *WebSocketHandler {
	return &WebSocketHandler{
		log:             log,
		config:          cfg,
		hub:             h,
		hmac:            hmac,
		originValidator: factory.originValidator,
		upgrader:        factory.Create(),
	}
}

// OriginValidator returns the origin validator for external use.
func (s *WebSocketHandler) OriginValidator() *security.OriginValidator {
	return s.originValidator
}

// HandleStream upgrades HTTP to WebSocket and registers the client.
// GET /v1/device/:id/stream
func (s *WebSocketHandler) HandleStream(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "device id required"})
		return
	}

	// HMAC verification for WebSocket upgrade if enforced
	if s.config.EnforceHMAC {
		body, err := s.hmac.ReadAndVerifyHTTP(c.Request)
		if err != nil {
			c.JSON(401, map[string]string{"error": "bad_hmac", "message": err.Error()})
			return
		}
		_ = body // Body consumed for verification
	}

	// Perform WebSocket upgrade
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.log.Warn("websocket upgrade failed", "deviceId", id, "err", err)
		return
	}

	// Register client with hub
	client := &hub.Client{
		DeviceID: id,
		Conn:     conn,
		Send:     make(chan models.CommandFrame, 32),
		Hub:      s.hub,
	}
	s.hub.Register(client)

	s.log.Info("device connected via websocket", "deviceId", id)

	// Start pumps - ReadPump blocks, so run WritePump in goroutine
	go client.WritePump()
	client.ReadPump()
}

// BroadcastTelemetry sends telemetry data to all connected dashboard clients.
func (s *WebSocketHandler) BroadcastTelemetry(raw []byte) {
	s.hub.BroadcastTelemetry(raw)
}

// ClientCount returns the number of connected WebSocket clients.
func (s *WebSocketHandler) ClientCount() int {
	return len(s.hub.Clients())
}

// GetClient retrieves a specific client by device ID.
func (s *WebSocketHandler) GetClient(deviceID string) *hub.Client {
	return s.hub.GetClient(deviceID)
}

// DisconnectClient forcefully disconnects a client.
func (s *WebSocketHandler) DisconnectClient(deviceID string) {
	client := s.hub.GetClient(deviceID)
	if client != nil {
		client.Hub.Unregister(client)
		_ = client.Conn.Close()
	}
}

// HandleIncomingMessage processes incoming WebSocket messages from devices.
func (s *WebSocketHandler) HandleIncomingMessage(client *hub.Client, raw []byte) error {
	var env struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		s.log.Warn("bad ws frame", "deviceId", client.DeviceID, "err", err)
		return err
	}

	switch env.Type {
	case "telemetry":
		return s.handleTelemetry(client, raw)
	case "pong":
		return s.handlePong(client)
	case "status":
		return s.handleStatus(client, raw)
	default:
		s.log.Warn("unknown ws message type", "deviceId", client.DeviceID, "type", env.Type)
	}

	return nil
}

// handleTelemetry processes telemetry frames from devices.
func (s *WebSocketHandler) handleTelemetry(client *hub.Client, raw []byte) error {
	var t models.TelemetryFrame
	if err := json.Unmarshal(raw, &t); err != nil {
		return err
	}
	t.Raw = raw
	if t.DeviceID == "" {
		t.DeviceID = client.DeviceID
	}

	// Save telemetry to store
	if err := client.Hub.Store().SaveTelemetry(context.Background(), client.DeviceID, raw, t); err != nil {
		s.log.Warn("telemetry save failed", "deviceId", client.DeviceID, "err", err)
	}

	// Broadcast to dashboard
	client.Hub.BroadcastTelemetry(raw)

	s.log.Debug("telemetry received", "deviceId", client.DeviceID, "riskScore", t.RiskScore)
	return nil
}

// handlePong handles ping/pong heartbeat responses.
func (s *WebSocketHandler) handlePong(client *hub.Client) error {
	return client.Hub.Store().Touch(context.Background(), client.DeviceID)
}

// handleStatus processes status updates from devices.
func (s *WebSocketHandler) handleStatus(client *hub.Client, raw []byte) error {
	s.log.Info("status update", "deviceId", client.DeviceID)
	return client.Hub.Store().Touch(context.Background(), client.DeviceID)
}

// SendToClient sends a command frame to a specific device.
func (s *WebSocketHandler) SendToClient(deviceID string, frame models.CommandFrame) bool {
	return s.hub.Send(deviceID, frame)
}

// IsOnline checks if a device is currently connected via WebSocket.
func (s *WebSocketHandler) IsOnline(deviceID string) bool {
	return s.hub.Online(deviceID)
}