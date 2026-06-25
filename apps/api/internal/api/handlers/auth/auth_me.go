package auth

import (
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"

	"github.com/gin-gonic/gin"
)

// MeHandler handles GET /v1/auth/me.
type MeHandler struct {
	authService *auth.AuthService
	presenter   *response.Presenter
}

// NewMeHandler creates a new MeHandler.
func NewMeHandler(authService *auth.AuthService, presenter *response.Presenter) *MeHandler {
	return &MeHandler{authService: authService, presenter: presenter}
}

// Handle processes the me request.
func (h *MeHandler) Handle(c *gin.Context) {
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "session invalid or expired")
			return
		}

		h.presenter.InternalError(c, "an error occurred")

		return
	}

	h.presenter.OK(c, gin.H{
		"id":             op.ID,
		"email":          op.Email,
		"name":           op.Name,
		"role":           op.Role,
		"mfa_enabled":    op.MFAEnabled,
		"email_verified": op.EmailVerified,
	})
}
