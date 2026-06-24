// Package websocket provides WebSocket handler implementations.
package websocket

import (
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	config "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
)

// StreamHandler handles WebSocket connections for device streaming.
// This is the main entry point for the device streaming API.
type StreamHandler struct {
	log           *slog.Logger
	hub           *hub.Hub
	upgrader      *StreamUpgrader
	messageRouter *MessageRouter
	presenter     *Presenter
	config        config.Config
}

// NewStreamHandler creates a new StreamHandler with all dependencies.
func NewStreamHandler(
	log *slog.Logger,
	cfg config.Config,
	h *hub.Hub,
	hmacVerifier cryptohmac.Verifier,
	auditLogger *audit.Logger,
) *StreamHandler {
	upgrader := NewStreamUpgrader(log, cfg, hmacVerifier)
	messageRouter := NewMessageRouter(h, log)
	presenter := NewPresenter(auditLogger, log)

	return &StreamHandler{
		log:           log,
		config:        cfg,
		hub:           h,
		upgrader:      upgrader,
		messageRouter: messageRouter,
		presenter:     presenter,
	}
}

// Handle handles GET /v1/device/:id/stream.
// This is the HTTP entry point that upgrades to WebSocket.
func (h *StreamHandler) Handle(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "device id required"})
		return
	}

	// Attempt WebSocket upgrade
	_, client, err := h.upgrader.Upgrade(c, deviceID)
	if err != nil {
		// Log failure via presenter
		if h.config.EnforceHMAC {
			h.presenter.LogHMACFailed(c.Request.Context(), deviceID)
		} else {
			h.presenter.LogUpgradeFailed(c.Request.Context(), deviceID, err.Error())
		}

		return
	}

	// Set hub on client
	client.Hub = h.hub

	// Register client with hub
	h.hub.Register(client)

	// Log connection and audit
	h.presenter.LogConnect(c.Request.Context(), deviceID)
	h.presenter.AuditDeviceConnect(c.Request.Context(), deviceID)

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
	h.messageRouter.DisconnectClient(deviceID)
}

// SendToClient sends a command frame to a specific device.
func (h *StreamHandler) SendToClient(deviceID string, frame command.CommandFrame) bool {
	return h.hub.Send(deviceID, frame)
}

// IsOnline checks if a device is currently connected via WebSocket.
func (h *StreamHandler) IsOnline(deviceID string) bool {
	return h.hub.Online(deviceID)
}
