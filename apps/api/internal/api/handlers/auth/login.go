package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// LoginHandler handles POST /v1/auth/login.
type LoginHandler struct {
	authService *auth.AuthService
	presenter   *response.Presenter
}

// NewLoginHandler creates a new LoginHandler.
func NewLoginHandler(authService *auth.AuthService, presenter *response.Presenter) *LoginHandler {
	return &LoginHandler{
		authService: authService,
		presenter:   presenter,
	}
}

// Handle processes the login request.
func (h *LoginHandler) Handle(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request")
		return
	}

	// Normalize email (trim whitespace and lowercase) - matches old behavior
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		h.presenter.BadRequest(c, "email and password required")
		return
	}

	// Validate email format using enterprise-grade validator (prevents SQL injection via email)
	if _, err := infraauth.ValidateEmail(req.Email); err != nil {
		// Return generic error to prevent email enumeration
		h.presenter.LoginFailure(c, req.Email, "invalid email or password")
		h.presenter.Unauthorized(c, "invalid email or password")

		return
	}

	// Add request timeout - prevents hanging on slow DB queries (matches old behavior)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, session, err := h.authService.Login(ctx, &req)
	if err != nil {
		// Record failed attempt for IP intelligence
		h.presenter.LoginFailure(c, req.Email, err.Error())

		switch {
		case errors.Is(err, application.ErrInvalidCredentials):
			h.presenter.Unauthorized(c, "invalid email or password")
		case errors.Is(err, application.ErrMFARequired):
			h.presenter.OK(c, gin.H{
				"mfa_required": true,
				"operator_id":  result.OperatorID,
			})
		default:
			h.presenter.InternalError(c, "an error occurred")
		}

		return
	}

	// Clear failed attempts on successful login
	h.presenter.LoginSuccess(c, result.OperatorID)

	// Create session cookie using session manager from authService
	if session != nil && h.authService.GetSessionManager() != nil {
		cookie, err := h.authService.GetSessionManager().CreateCookie(result.OperatorID)
		if err == nil {
			h.presenter.SetSessionCookie(c, cookie)
		}
	}

	h.presenter.OK(c, result)
}
