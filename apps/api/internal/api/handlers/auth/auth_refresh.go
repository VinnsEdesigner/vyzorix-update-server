package auth

import (
	"context"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"

	"github.com/gin-gonic/gin"
)

// RefreshHandler handles POST /v1/auth/refresh.
type RefreshHandler struct {
	authService *auth.AuthService
	presenter   *response.Presenter
}

// NewRefreshHandler creates a new RefreshHandler.
func NewRefreshHandler(authService *auth.AuthService, presenter *response.Presenter) *RefreshHandler {
	return &RefreshHandler{
		authService: authService,
		presenter:   presenter,
	}
}

// Handle processes the refresh token request.
// POST /v1/auth/refresh
// Request: { "refresh_token": "..." }
// Response: { "access_token": "...", "refresh_token": "...", "expires_at": 1234567890, "session_id": "..." }.
func (h *RefreshHandler) Handle(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request")
		return
	}

	if req.RefreshToken == "" {
		h.presenter.BadRequest(c, "refresh_token required")
		return
	}

	// Add request timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Rotate the refresh token and get new tokens
	result, err := h.authService.RotateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrUnauthorized):
			h.presenter.Unauthorized(c, "invalid or expired refresh token")
		case errors.Is(err, application.ErrTokenExpired):
			h.presenter.Unauthorized(c, "Refresh token has expired")
		default:
			h.presenter.InternalError(c, "Failed to refresh token")
		}
		return
	}

	h.presenter.OK(c, dto.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.Unix(),
		SessionID:    result.SessionID,
	})
}
