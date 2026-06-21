package websocket

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	config "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// StreamHandler handles WebSocket connections for device streaming.
type StreamHandler struct {
	log             *slog.Logger
	hub             *hub.Hub
	originValidator *auth.OriginValidator
	upgrader        websocket.Upgrader
	hmacVerifier    cryptohmac.Verifier
	config          config.Config
}

// NewStreamHandler creates a new StreamHandler.
func NewStreamHandler(log *slog.Logger, cfg config.Config, h *hub.Hub, hmacVerifier cryptohmac.Verifier) *StreamHandler {
	originValidator := auth.NewOriginValidator(cfg.AllowedOrigins)
	originValidator.SetLogger(log)

	return &StreamHandler{
		log:             log,
		hub:             h,
		originValidator: originValidator,
		config:          cfg,
		hmacVerifier:    hmacVerifier,
		upgrader: websocket.Upgrader{
			CheckOrigin:      originValidator.CheckOrigin(),
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// Handle handles GET /v1/device/:id/stream.
func (h *StreamHandler) Handle(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "device id required"})
		return
	}

	// HMAC verification for WebSocket upgrade if enforced
	if h.config.EnforceHMAC {
		body, err := h.hmacVerifier.ReadAndVerifyHTTP(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid request"})
			return
		}
		_ = body // Body consumed for verification
	}

	// Perform WebSocket upgrade
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Warn("websocket upgrade failed", "deviceId", id, "err", err)
		return
	}

	// Register client with hub
	client := &hub.Client{
		DeviceID: id,
		Conn:     conn,
		Send:     make(chan command.CommandFrame, 32),
		Hub:      h.hub,
	}
	h.hub.Register(client)

	h.log.Info("device connected via websocket", "deviceId", id)

	// Start pumps - ReadPump blocks, so run WritePump in goroutine
	go client.WritePump()
	client.ReadPump()
}

// BroadcastTelemetry sends telemetry data to all connected dashboard clients.
func (h *StreamHandler) BroadcastTelemetry(raw []byte) {
	h.hub.BroadcastTelemetry(raw)
}

// ClientCount returns the number of connected WebSocket clients.
func (h *StreamHandler) ClientCount() int {
	return h.hub.ClientCount()
}

// GetClient retrieves a specific client by device ID.
func (h *StreamHandler) GetClient(deviceID string) *hub.Client {
	return h.hub.GetClient(deviceID)
}

// DisconnectClient forcefully disconnects a client by device ID.
func (h *StreamHandler) DisconnectClient(deviceID string) {
	client := h.hub.GetClient(deviceID)
	if client != nil {
		h.hub.Unregister(client)
		if err := client.Conn.Close(); err != nil {
			h.log.Warn("client close failed", "deviceId", deviceID, "err", err)
		}
	}
}

// HandleIncomingMessage processes incoming WebSocket messages from devices.
func (h *StreamHandler) HandleIncomingMessage(client *hub.Client, raw []byte) error {
	var env struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		h.log.Warn("bad ws frame", "deviceId", client.DeviceID, "err", err)
		return err
	}

	switch env.Type {
	case "telemetry":
		return h.handleTelemetry(client, raw)
	case "pong":
		return h.handlePong(client)
	case "status":
		return h.handleStatus(client, raw)
	default:
		h.log.Warn("unknown ws message type", "deviceId", client.DeviceID, "type", env.Type)
	}

	return nil
}

// handleTelemetry processes telemetry frames from devices.
func (h *StreamHandler) handleTelemetry(client *hub.Client, raw []byte) error {
	var t telemetry.TelemetryFrame
	if err := json.Unmarshal(raw, &t); err != nil {
		return err
	}
	t.Raw = raw
	if t.DeviceID == "" {
		t.DeviceID = client.DeviceID
	}

	// Broadcast to dashboard
	h.hub.BroadcastTelemetry(raw)

	h.log.Debug("telemetry received", "deviceId", client.DeviceID, "riskScore", t.RiskScore)
	return nil
}

// handlePong handles ping/pong heartbeat responses.
func (h *StreamHandler) handlePong(client *hub.Client) error {
	return nil
}

// handleStatus processes status updates from devices.
func (h *StreamHandler) handleStatus(client *hub.Client, raw []byte) error {
	h.log.Info("status update", "deviceId", client.DeviceID)
	return nil
}

// SendToClient sends a command frame to a specific device.
func (h *StreamHandler) SendToClient(deviceID string, frame command.CommandFrame) bool {
	return h.hub.Send(deviceID, frame)
}

// IsOnline checks if a device is currently connected via WebSocket.
func (h *StreamHandler) IsOnline(deviceID string) bool {
	return h.hub.Online(deviceID)
}
