package auth

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"

	"github.com/gin-gonic/gin"
)

// LogoutHandler handles POST /v1/auth/logout.
type LogoutHandler struct {
	authService *auth.AuthService
	auditLogger *audit.Logger
}

// NewLogoutHandler creates a new LogoutHandler.
func NewLogoutHandler(authService *auth.AuthService, auditLogger *audit.Logger) *LogoutHandler {
	return &LogoutHandler{authService: authService, auditLogger: auditLogger}
}

// Handle processes the logout request.
func (h *LogoutHandler) Handle(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), sessionID); err != nil {
		// Log error but don't fail - logout should always succeed
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
