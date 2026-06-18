package auth

import (
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"

	"github.com/gin-gonic/gin"
)

// MeHandler handles GET /v1/auth/me.
type MeHandler struct {
	authService *auth.AuthService
}

// NewMeHandler creates a new MeHandler.
func NewMeHandler(authService *auth.AuthService) *MeHandler {
	return &MeHandler{authService: authService}
}

// Handle processes the me request.
func (h *MeHandler) Handle(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
		return
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "session invalid or expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "an error occurred"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              op.ID,
		"email":           op.Email,
		"name":            op.Name,
		"role":            op.Role,
		"mfa_enabled":     op.MFAEnabled,
		"email_verified": op.EmailVerified,
	})
}
