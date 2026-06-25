// Package websocket provides WebSocket handler implementations.
package websocket

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	config "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// StreamUpgrader handles WebSocket upgrade logic including origin validation and HMAC verification.
type StreamUpgrader struct {
	originValidator *infraauth.OriginValidator
	upgrader        websocket.Upgrader
	hmacVerifier    cryptohmac.Verifier
	config          config.Config
}

// NewStreamUpgrader creates a new StreamUpgrader.
func NewStreamUpgrader(
	log *slog.Logger,
	cfg config.Config,
	hmacVerifier cryptohmac.Verifier,
) *StreamUpgrader {
	originValidator := infraauth.NewOriginValidator(cfg.AllowedOrigins)
	originValidator.SetLogger(log)

	return &StreamUpgrader{
		originValidator: originValidator,
		hmacVerifier:    hmacVerifier,
		config:          cfg,
		upgrader: websocket.Upgrader{
			CheckOrigin:      originValidator.CheckOrigin(),
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// Upgrade attempts to upgrade an HTTP connection to WebSocket.
// Returns the WebSocket connection and registered client on success.
// Returns an error and sends appropriate HTTP response on failure.
func (u *StreamUpgrader) Upgrade(c *gin.Context, deviceID string) (*websocket.Conn, *hub.Client, error) {
	// Verify HMAC signature if enforced
	if u.config.EnforceHMAC {
		if err := u.verifyHMAC(c.Request); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid request"})
			return nil, nil, err
		}
	}

	// Perform WebSocket upgrade
	conn, err := u.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, nil, err
	}

	// Create client for hub
	client := &hub.Client{
		DeviceID: deviceID,
		Conn:     conn,
		Send:     make(chan command.CommandFrame, 32),
		Hub:      nil, // Will be set by caller
	}

	return conn, client, nil
}

// verifyHMAC verifies the HMAC signature of the request.
func (u *StreamUpgrader) verifyHMAC(r *http.Request) error {
	_, err := u.hmacVerifier.ReadAndVerifyHTTP(r)
	return err
}
