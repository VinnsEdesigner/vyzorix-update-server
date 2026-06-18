package auth

import (
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"

	"github.com/gin-gonic/gin"
)

// LoginHandler handles POST /v1/auth/login.
type LoginHandler struct {
	authService *auth.AuthService
}

// NewLoginHandler creates a new LoginHandler.
func NewLoginHandler(authService *auth.AuthService) *LoginHandler {
	return &LoginHandler{authService: authService}
}

// Handle processes the login request.
func (h *LoginHandler) Handle(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "email and password required"})
		return
	}

	result, session, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials", "message": "invalid email or password"})
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
