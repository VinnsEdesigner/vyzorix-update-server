package websocket

import (
	"log/slog"
	"net/http"
	"time"

	security "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/crypto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// StreamHandler handles WebSocket connections for device streaming.
type StreamHandler struct {
	log             *slog.Logger
	hub             *hub.Hub
	originValidator *security.OriginValidator
	upgrader        websocket.Upgrader
	hmacVerifier    cryptohmac.Verifier
	config          config.Config
}

// NewStreamHandler creates a new StreamHandler.
func NewStreamHandler(log *slog.Logger, cfg config.Config, h *hub.Hub, hmacVerifier cryptohmac.Verifier) *StreamHandler {
	originValidator := security.NewOriginValidator(cfg.AllowedOrigins)
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "bad_hmac", "message": err.Error()})
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
		Send:     make(chan models.CommandFrame, 32),
		Hub:      h.hub,
	}
	h.hub.Register(client)

	h.log.Info("device connected via websocket", "deviceId", id)

	// Start pumps - ReadPump blocks, so run WritePump in goroutine
	go client.WritePump()
	client.ReadPump()
}
