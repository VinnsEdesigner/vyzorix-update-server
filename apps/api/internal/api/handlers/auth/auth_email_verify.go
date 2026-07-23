package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"

	"github.com/gin-gonic/gin"
)

// EmailVerifyHandler handles email verification endpoints.
type EmailVerifyHandler struct {
	authService *auth.AuthService
	emailSvc    *emailService.Service
	presenter   *response.Presenter
}

// NewEmailVerifyHandler creates a new EmailVerifyHandler.
func NewEmailVerifyHandler(authService *auth.AuthService, emailSvc *emailService.Service, presenter *response.Presenter) *EmailVerifyHandler {
	return &EmailVerifyHandler{
		authService: authService,
		emailSvc:    emailSvc,
		presenter:   presenter,
	}
}

// VerifyEmail handles POST /v1/auth/verify-email.
func (h *EmailVerifyHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid JSON body")
		return
	}

	if req.Token == "" {
		h.presenter.BadRequest(c, "token is required")
		return
	}

	result, err := h.authService.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		h.presenter.Unauthorized(c, "invalid or expired verification token")
		return
	}

	h.presenter.OK(c, gin.H{"verified": true, "email": result.Email})
}

// ResendVerification handles POST /v1/auth/resend-verification.
func (h *EmailVerifyHandler) ResendVerification(c *gin.Context) {
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

	// Delete old verification tokens and create/send new one
	err := h.authService.ResendVerification(c.Request.Context(), req.Email)
	if err != nil {
		h.presenter.InternalError(c, "failed to resend verification email")
		return
	}

	h.presenter.OK(c, gin.H{"message": "If that email exists, a verification email has been sent."})
}

// PollVerification handles GET /v1/auth/poll-verification.
func (h *EmailVerifyHandler) PollVerification(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		h.presenter.BadRequest(c, "token is required")
		return
	}

	result, err := h.authService.PollVerification(c.Request.Context(), token)
	if err != nil {
		h.presenter.InternalError(c, "verification check failed")
		return
	}

	h.presenter.OK(c, result)
}

// VerifyEmailGet handles GET /v1/auth/verify-email (alternative to POST).
func (h *EmailVerifyHandler) VerifyEmailGet(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		h.presenter.BadRequest(c, "token query parameter is required")
		return
	}

	result, err := h.authService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		h.presenter.Unauthorized(c, "invalid or expired verification token")
		return
	}

	h.presenter.OK(c, gin.H{"verified": true, "email": result.Email})
}

// ResendVerificationGet handles GET /v1/auth/resend-verification (alternative to POST).
func (h *EmailVerifyHandler) ResendVerificationGet(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		h.presenter.BadRequest(c, "email query parameter is required")
		return
	}

	// Delete old verification tokens and create/send new one
	err := h.authService.ResendVerification(c.Request.Context(), email)
	if err != nil {
		h.presenter.InternalError(c, "failed to resend verification email")
		return
	}

	h.presenter.OK(c, gin.H{"message": "If that email exists, a verification email has been sent."})
}

// CancelVerification handles POST /v1/auth/cancel-verification.
func (h *EmailVerifyHandler) CancelVerification(c *gin.Context) {
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

	// Cancel verification - return success for security (don't reveal if email exists)
	err := h.authService.CancelVerification(c.Request.Context(), req.Email)
	if err != nil {
		h.presenter.InternalError(c, "failed to cancel verification")
		return
	}

	h.presenter.OK(c, gin.H{"success": true})
}
