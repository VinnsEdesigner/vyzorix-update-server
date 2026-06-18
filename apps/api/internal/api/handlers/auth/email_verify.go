package auth

import (
"net/http"

"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"

"github.com/gin-gonic/gin"
)

// EmailVerifyHandler handles email verification endpoints.
type EmailVerifyHandler struct {
authService *auth.AuthService
emailSvc    *emailService.Service
}

// NewEmailVerifyHandler creates a new EmailVerifyHandler.
func NewEmailVerifyHandler(authService *auth.AuthService, emailSvc *emailService.Service) *EmailVerifyHandler {
return &EmailVerifyHandler{
authService: authService,
emailSvc:    emailSvc,
}
}

// VerifyEmail handles POST /v1/auth/verify-email.
func (h *EmailVerifyHandler) VerifyEmail(c *gin.Context) {
var req struct {
Token string `json:"token"`
}
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid JSON body"})
return
}

if req.Token == "" {
c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "token is required"})
return
}

result, err := h.authService.VerifyEmail(c.Request.Context(), req.Token)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_token", "message": "invalid or expired verification token"})
return
}

c.JSON(http.StatusOK, gin.H{"verified": true, "email": result.Email})
}

// ResendVerification handles POST /v1/auth/resend-verification.
func (h *EmailVerifyHandler) ResendVerification(c *gin.Context) {
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

// Delete old verification tokens and create/send new one
err := h.authService.ResendVerification(c.Request.Context(), req.Email)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to resend verification email"})
return
}

c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a verification email has been sent."})
}

// PollVerification handles GET /v1/auth/poll-verification.
func (h *EmailVerifyHandler) PollVerification(c *gin.Context) {
token := c.Query("token")
if token == "" {
c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "token is required"})
return
}

status, email, err := h.authService.PollVerification(c.Request.Context(), token)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "verification check failed"})
return
}

c.JSON(http.StatusOK, gin.H{"status": status, "email": email})
}

// CancelVerification handles POST /v1/auth/cancel-verification.
func (h *EmailVerifyHandler) CancelVerification(c *gin.Context) {
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

// Cancel verification - return success for security (don't reveal if email exists)
err := h.authService.CancelVerification(c.Request.Context(), req.Email)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to cancel verification"})
return
}

c.JSON(http.StatusOK, gin.H{"success": true})
}
