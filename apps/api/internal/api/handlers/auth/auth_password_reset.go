package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.ForgotPasswordRequest
	_ openapi.ResetPasswordRequest
	_ openapi.MessageResult
	_ openapi.SuccessResult
	_ openapi.ErrorResponse
)

// PasswordResetHandler handles password reset endpoints.
type PasswordResetHandler struct {
	authService *auth.AuthService
	emailSvc    *emailService.Service
	presenter   *response.Presenter
}

// NewPasswordResetHandler creates a new PasswordResetHandler.
func NewPasswordResetHandler(authService *auth.AuthService, emailSvc *emailService.Service, presenter *response.Presenter) *PasswordResetHandler {
	return &PasswordResetHandler{
		authService: authService,
		emailSvc:    emailSvc,
		presenter:   presenter,
	}
}

// ForgotPassword handles POST /v1/auth/forgot-password.
// @Summary      Forgot password
// @Description  Sends a password reset link to the email if it exists
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.ForgotPasswordRequest  true  "email"
// @Success      200  {object}  openapi.MessageResult  "reset link sent (if email exists)"
// @Failure      400  {object}  openapi.ErrorResponse  "email required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/forgot-password [post]
func (h *PasswordResetHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid JSON body")
		return
	}

	if req.Email == "" {
		h.presenter.BadRequest(c, "email is required")
		return
	}

	// Generate token (returns empty string if email not found - security).
	token, err := h.authService.GeneratePasswordResetToken(c.Request.Context(), req.Email)
	if err != nil {
		h.presenter.InternalError(c, "request failed")
		return
	}

	// Send password reset email if configured.
	if token != "" && h.emailSvc != nil && h.emailSvc.IsConfigured() {
		// Get operator name for email.
		op, err := h.authService.GetOperatorByEmail(c.Request.Context(), req.Email)
		name := ""

		if err == nil && op != nil {
			name = op.Name
		}

		if err := h.emailSvc.SendPasswordResetEmail(c.Request.Context(), req.Email, name, token); err != nil {
			h.presenter.InternalError(c, "failed to send email")
			return
		}
	}

	h.presenter.OK(c, gin.H{"message": "If that email exists, a password reset link has been sent."})
}

// ResendPasswordReset handles POST /v1/auth/resend-password-reset.
// @Summary      Resend password reset
// @Description  Resends a password reset link with rate limiting
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.ForgotPasswordRequest  true  "email"
// @Success      200  {object}  openapi.SuccessResult  "reset link sent"
// @Failure      400  {object}  openapi.ErrorResponse  "email required"
// @Failure      429  {object}  openapi.ErrorResponse  "rate limited"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/resend-password-reset [post]
func (h *PasswordResetHandler) ResendPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid request body")
		return
	}

	if req.Email == "" {
		h.presenter.BadRequest(c, "email is required")
		return
	}

	// Check rate limit before processing.
	rateLimit, err := h.authService.CheckResendRateLimit(c.Request.Context(), req.Email)
	if err != nil {
		h.presenter.InternalError(c, "failed to check rate limit")
		return
	}

	if !rateLimit.Allowed {
		// Return rate limit response.
		if rateLimit.LockedUntil != nil {
			h.presenter.OK(c, gin.H{
				"error":       "rate_limited",
				"message":     "Too many requests. Please try again later.",
				"lockedUntil": rateLimit.LockedUntil.UnixMilli(),
			})

			return
		}

		h.presenter.OK(c, gin.H{
			"error":      "rate_limited",
			"message":    "Please wait before requesting another reset link.",
			"retryAfter": rateLimit.RetryAfter,
		})

		return
	}

	token, err := h.authService.GeneratePasswordResetToken(c.Request.Context(), req.Email)
	if err != nil {
		h.presenter.InternalError(c, "request failed")
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
			h.presenter.InternalError(c, "failed to send email")
			return
		}
	}

	// Update rate limit tracker after successful email send.
	// Use a background context since the response is already sent.
	if err := h.authService.UpdateResendTracker(c.Request.Context(), req.Email); err != nil {
		// Log but don't fail the request - email was already sent.
		_ = c.Error(err)
	}

	h.presenter.OK(c, gin.H{"success": true, "message": "Password reset link sent."})
}

// ResetPassword handles POST /v1/auth/reset-password.
// @Summary      Reset password
// @Description  Resets a password using a valid reset token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.ResetPasswordRequest  true  "reset token + new password"
// @Success      200  {object}  openapi.SuccessResult  "password reset"
// @Failure      400  {object}  openapi.ErrorResponse  "token and newPassword required"
// @Failure      401  {object}  openapi.ErrorResponse  "invalid or expired reset token"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/reset-password [post]
func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid JSON body")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		h.presenter.BadRequest(c, "token and newPassword are required")
		return
	}

	// Get operator email from token for validation.
	// First, validate the token exists and is not expired.
	if err := h.authService.ValidatePasswordResetToken(c.Request.Context(), req.Token, ""); err != nil {
		h.presenter.Unauthorized(c, "invalid or expired reset token")
		return
	}

	// Reset password.
	err := h.authService.ResetPassword(c.Request.Context(), req.Token, "", req.NewPassword)
	if err != nil {
		h.presenter.BadRequest(c, "Invalid request")
		return
	}

	h.presenter.OK(c, gin.H{"success": true, "message": "Password has been reset successfully."})
}
