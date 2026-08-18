// Package websocket provides WebSocket handler implementations.
package websocket

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
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
	log             *slog.Logger
	upgrader        websocket.Upgrader
	hmacVerifier    cryptohmac.Verifier
	config          config.Config
	allowDevAuth    bool
}

// NewStreamUpgrader creates a new StreamUpgrader.
func NewStreamUpgrader(
	log *slog.Logger,
	cfg config.Config,
	hmacVerifier cryptohmac.Verifier,
) *StreamUpgrader {
	originValidator := infraauth.NewOriginValidator(cfg.AllowedOrigins)
	originValidator.SetLogger(log)

	allowDevAuth := false
	if cfg.Env != "production" && !cfg.EnforceHMAC {
		log.Warn("WebSocket HMAC authentication is disabled - this is insecure and should not be used in production",
			"env", cfg.Env,
			"enforceHMAC", cfg.EnforceHMAC,
		)
		allowDevAuth = true
	}

	return &StreamUpgrader{
		originValidator: originValidator,
		hmacVerifier:    hmacVerifier,
		config:          cfg,
		log:             log,
		allowDevAuth:    allowDevAuth,
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

	// In development, only skip if allowDevAuth is true (meaning EnforceHMAC=false was set).
	if u.config.Env == "production" || u.config.EnforceHMAC {
		// Use query param based verification for WebSocket since headers can't be set during upgrade.
		if err := u.verifyWebSocketHMAC(c.Request, deviceID); err != nil {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "WebSocket HMAC verification failed"))
			return nil, nil, err
		}
	} else if !u.allowDevAuth {
		_ = c.Error(
			// This case handles production with EnforceHMAC=false (shouldn't happen but is defensive).
			apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "WebSocket authentication is required"))
		return nil, nil, http.ErrNotSupported
	}

	// Perform WebSocket upgrade.
	conn, err := u.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, nil, err
	}

	// Create client for hub.
	client := &hub.Client{
		DeviceID: deviceID,
		Conn:     conn,
		Send:     make(chan command.CommandFrame, 32),
		Hub:      nil, // Will be set by caller.

		Done: make(chan struct{}),
	}

	return conn, client, nil
}

// verifyWebSocketHMAC verifies the HMAC signature from query parameters.
// This is needed because WebSocket upgrade requests cannot set custom headers.
func (u *StreamUpgrader) verifyWebSocketHMAC(r *http.Request, deviceID string) error {
	return u.hmacVerifier.VerifyWebSocketConnect(r, deviceID)
}
