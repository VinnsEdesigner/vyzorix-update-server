package auth

import (
	"context"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.RefreshTokenRequest
	_ openapi.RefreshTokenResult
	_ openapi.ErrorResponse
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

// Handle handles POST /v1/auth/refresh.
// @Summary      Refresh access token
// @Description  Exchanges a refresh token for a new access/refresh token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.RefreshTokenRequest  true  "refresh token"
// @Success      200  {object}  openapi.RefreshTokenResult  "new token pair"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid request"
// @Failure      401  {object}  openapi.ErrorResponse  "invalid or expired refresh token"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/refresh [post]
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

	// Add request timeout.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Rotate the refresh token and get new tokens.
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
