package auth

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"

	"github.com/gin-gonic/gin"
)

// PasswordResetHandler handles password reset endpoints.
type PasswordResetHandler struct {
	authService *auth.AuthService
	emailSvc   *emailService.Service
}

// NewPasswordResetHandler creates a new PasswordResetHandler.
func NewPasswordResetHandler(authService *auth.AuthService, emailSvc *emailService.Service) *PasswordResetHandler {
	return &PasswordResetHandler{
		authService: authService,
		emailSvc:   emailSvc,
	}
}

// ForgotPassword handles POST /v1/auth/forgot-password.
func (h *PasswordResetHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid JSON body"})
		return
	}

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "email is required"})
		return
	}

	// Generate token (returns empty string if email not found - security)
	token, err := h.authService.GeneratePasswordResetToken(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "request failed"})
		return
	}

	// Send password reset email if configured
	if token != "" && h.emailSvc != nil && h.emailSvc.IsConfigured() {
		// Get operator name for email
		op, err := h.authService.GetOperatorByEmail(c.Request.Context(), req.Email)
		name := ""
		if err == nil && op != nil {
			name = op.Name
		}
		if err := h.emailSvc.SendPasswordResetEmail(c.Request.Context(), req.Email, name, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to send email"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a password reset link has been sent."})
}

// ResendPasswordReset handles POST /v1/auth/resend-password-reset.
func (h *PasswordResetHandler) ResendPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid request body"})
		return
	}

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "email is required"})
		return
	}

	// Check rate limit before processing.
	rateLimit, err := h.authService.CheckResendRateLimit(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to check rate limit"})
		return
	}

	if !rateLimit.Allowed {
		// Return rate limit response.
		if rateLimit.LockedUntil != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limited",
				"message":     "Too many requests. Please try again later.",
				"lockedUntil": rateLimit.LockedUntil.UnixMilli(),
			})
			return
		}
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "rate_limited",
			"message":     "Please wait before requesting another reset link.",
			"retryAfter":  rateLimit.RetryAfter,
		})
		return
	}

	token, err := h.authService.GeneratePasswordResetToken(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "request failed"})
		return
	}

	// Send password reset email if configured.
	if token != "" && h.emailSvc != nil && h.emailSvc.IsConfigured() {
		op, err := h.authService.GetOperatorByEmail(c.Request.Context(), req.Email)
		name := ""
		if err == nil && op != nil {
			name = op.Name
		}
		if err := h.emailSvc.SendPasswordResetEmail(c.Request.Context(), req.Email, name, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to send email"})
			return
		}
	}

	// Update rate limit tracker after successful email send.
	// Use a background context since the response is already sent.
	if err := h.authService.UpdateResendTracker(c.Request.Context(), req.Email); err != nil {
		// Log but don't fail the request - email was already sent.
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password reset link sent."})
}

// ResetPassword handles POST /v1/auth/reset-password.
func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid JSON body"})
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "token and newPassword are required"})
		return
	}

	// Get operator email from token for validation
	// First, validate the token exists and is not expired
	if err := h.authService.ValidatePasswordResetToken(c.Request.Context(), req.Token, ""); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unauthorized", "message": "invalid or expired reset token"})
		return
	}

	// Reset password
	err := h.authService.ResetPassword(c.Request.Context(), req.Token, "", req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "internal_error", "message": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password has been reset successfully."})
}
