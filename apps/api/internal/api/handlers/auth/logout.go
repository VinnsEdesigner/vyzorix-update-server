package auth

import (
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"

	"github.com/gin-gonic/gin"
)

// LogoutHandler handles POST /v1/auth/logout.
type LogoutHandler struct {
	authService *auth.AuthService
	logger      *slog.Logger
}

// NewLogoutHandler creates a new LogoutHandler.
func NewLogoutHandler(authService *auth.AuthService) *LogoutHandler {
	return &LogoutHandler{authService: authService}
}

// WithLogger sets the logger for the handler.
func (h *LogoutHandler) WithLogger(logger *slog.Logger) *LogoutHandler {
	h.logger = logger
	return h
}

// Handle processes the logout request.
func (h *LogoutHandler) Handle(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), sessionID); err != nil {
		if h.logger != nil {
			h.logger.Warn("Logout failed", "err", err)
		}
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
