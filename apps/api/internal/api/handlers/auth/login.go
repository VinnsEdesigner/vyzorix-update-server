package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// LoginHandler handles POST /v1/auth/login.
type LoginHandler struct {
	authService    *auth.AuthService
	auditLogger   *audit.Logger
	ipIntelligence *middleware.IPIntelligence
}

// NewLoginHandler creates a new LoginHandler.
func NewLoginHandler(authService *auth.AuthService, auditLogger *audit.Logger, ipIntelligence *middleware.IPIntelligence) *LoginHandler {
	return &LoginHandler{
		authService:    authService,
		auditLogger:    auditLogger,
		ipIntelligence: ipIntelligence,
	}
}

// Handle processes the login request.
func (h *LoginHandler) Handle(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request"})
		return
	}

	// Normalize email (trim whitespace and lowercase) - matches old behavior
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "email and password required"})
		return
	}

	// Validate email format using enterprise-grade validator (prevents SQL injection via email)
	if _, err := infraauth.ValidateEmail(req.Email); err != nil {
		// Return generic error to prevent email enumeration
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "invalid email or password"})
		return
	}

	// Add request timeout - prevents hanging on slow DB queries (matches old behavior)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, session, err := h.authService.Login(ctx, &req)
	if err != nil {
		// Record failed attempt for IP intelligence
		if h.ipIntelligence != nil {
			h.ipIntelligence.RecordAuthFailure(c)
		}
		
		switch {
		case errors.Is(err, application.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "invalid email or password"})
		case errors.Is(err, application.ErrMFARequired):
			c.JSON(http.StatusOK, gin.H{
				"mfa_required": true,
				"operator_id":  result.OperatorID,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "an error occurred"})
		}
		return
	}

	// Clear failed attempts on successful login
	if h.ipIntelligence != nil {
		h.ipIntelligence.RecordAuthSuccess(c)
	}

	if session != nil {
		c.SetCookie(
			"session_id",
			session.ID,
			int(session.ExpiresAt.Sub(session.CreatedAt).Seconds()),
			"/",
			"",
			false,
			true,
		)
	}

	c.JSON(http.StatusOK, result)
}
