package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	security "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
)

// VerifyEmail handles email verification requests.
// POST /v1/auth/verify-email.
func (ac *AuthController) VerifyEmail(c *gin.Context) {
	var req models.VerifyEmailRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "token is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Find verification token.
	tokenHash := security.HashToken(token)
	ev, err := ac.store.GetEmailVerificationByTokenHash(ctx, tokenHash)
	if err != nil {
		ac.log.Warn("verifyEmail: db error", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "verification failed"})
		return
	}

	if ev == nil {
		c.JSON(400, models.ErrorResponse{Error: "invalid_token", Message: "invalid or expired verification token"})
		return
	}

	// Check if expired.
	if time.Now().UTC().After(ev.ExpiresAt) {
		c.JSON(400, models.ErrorResponse{Error: "token_expired", Message: "verification token has expired"})
		return
	}

	// Mark email as verified.
	if err := ac.store.SetOperatorEmailVerified(ctx, ev.OperatorID, true); err != nil {
		ac.log.Warn("verifyEmail: failed to set verified", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "verification failed"})
		return
	}

	// Delete the verification token (single use).
	if err := ac.store.DeleteEmailVerification(ctx, ev.ID); err != nil {
		ac.log.Warn("verifyEmail: failed to delete verification", "err", err)
	}

	ac.log.Info("verifyEmail: success", "operatorID", ev.OperatorID)
	c.JSON(200, models.EmailVerifiedResponse{Verified: true})
}

// ResendVerification resends the verification email.
// POST /v1/auth/resend-verification.
func (ac *AuthController) ResendVerification(c *gin.Context) {
	var req models.ResendVerificationRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "email is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Find operator by email.
	op, err := ac.store.GetOperatorByEmail(ctx, email)
	if err != nil {
		ac.log.Warn("resendVerification: db error", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "request failed"})
		return
	}

	// Always return success for security (don't reveal if email exists).
	if op == nil {
		ac.log.Info("resendVerification: email not found (silently)", "email", email)
		c.JSON(200, models.MessageResponse{Message: "If that email exists, a verification email has been sent."})
		return
	}

	// Check if already verified.
	if op.EmailVerified {
		c.JSON(400, models.ErrorResponse{Error: "already_verified", Message: "this email is already verified"})
		return
	}

	// Delete old verification tokens.
	if err := ac.store.DeleteEmailVerificationsByOperator(ctx, op.ID); err != nil {
		ac.log.Warn("resendVerification: failed to delete old verifications", "err", err)
	}

	// Send new verification email.
	ac.sendVerificationEmail(ctx, op)

	ac.log.Info("resendVerification: sent", "email", email)
	c.JSON(200, models.MessageResponse{Message: "If that email exists, a verification email has been sent."})
}

// PollVerification checks the status of an email verification token.
func (ac *AuthController) PollVerification(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "token is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Hash the token for lookup.
	tokenHash := security.HashToken(token)
	ev, err := ac.store.GetEmailVerificationByTokenHash(ctx, tokenHash)
	if err != nil {
		ac.log.Warn("pollVerification: db error", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "verification check failed"})
		return
	}

	if ev == nil {
		c.JSON(200, models.VerificationPollResponse{Status: "invalid", Email: ""})
		return
	}

	// Check if expired.
	if time.Now().UTC().After(ev.ExpiresAt) {
		// Delete expired token (ignore error - best effort cleanup).
		_ = ac.store.DeleteEmailVerification(ctx, ev.ID) //nolint:errcheck
		c.JSON(200, models.VerificationPollResponse{Status: "expired", Email: ""})
		return
	}

	// Get operator to check verification status and get email.
	op, err := ac.store.GetOperatorByID(ctx, ev.OperatorID)
	if err != nil || op == nil {
		// Operator not yet fully created, still waiting.
		c.JSON(200, models.VerificationPollResponse{Status: "waiting", Email: ""})
		return
	}

	if op.EmailVerified {
		// Already verified!
		c.JSON(200, models.VerificationPollResponse{Status: "success", Email: op.Email})
		return
	}

	// Still waiting for user to click verification link.
	c.JSON(200, models.VerificationPollResponse{Status: "waiting", Email: op.Email})
}

// CancelVerification removes pending verification tokens for an email.
func (ac *AuthController) CancelVerification(c *gin.Context) {
	var req models.CancelVerificationRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid request body"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "email is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Find operator by email.
	op, err := ac.store.GetOperatorByEmail(ctx, email)
	if err != nil {
		ac.log.Warn("cancelVerification: db error", "err", err)
		// Always return success for security.
		c.JSON(200, map[string]bool{"success": true})
		return
	}

	// Delete any pending verification tokens.
	if op != nil {
		if err := ac.store.DeleteEmailVerificationsByOperator(ctx, op.ID); err != nil {
			ac.log.Warn("cancelVerification: failed to delete verifications", "err", err)
		}
	}

	// Always return success for security (prevents email enumeration).
	c.JSON(200, map[string]bool{"success": true})
}

// sendVerificationEmail creates a verification token and sends the verification email.
func (ac *AuthController) sendVerificationEmail(ctx context.Context, op *models.Operator) {
	if !ac.emailSvc.IsConfigured() {
		ac.log.Warn("sendVerificationEmail: email service not configured, skipping", "email", op.Email)
		return
	}

	// Generate verification token.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		ac.log.Warn("sendVerificationEmail: failed to generate token", "err", err)
		return
	}
	token := hex.EncodeToString(tokenBytes)
	tokenHash := security.HashToken(token) // Store hash in DB

	// Create verification record.
	ev := &storage.EmailVerification{
		ID:         GenerateID(),
		OperatorID: op.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().UTC().Add(ac.config.EmailVerifyTokenExpiry),
		CreatedAt:  time.Now().UTC(),
	}

	if err := ac.store.CreateEmailVerification(ctx, ev); err != nil {
		ac.log.Warn("sendVerificationEmail: failed to store token", "err", err)
		return
	}

	// Send email (async, don't block registration).
	go func() {
		if err := ac.emailSvc.SendVerificationEmail(context.Background(), op.Email, op.Name, token); err != nil {
			ac.log.Error("sendVerificationEmail: failed to send email", "email", op.Email, "err", err)
		} else {
			ac.log.Info("sendVerificationEmail: sent", "email", op.Email)
		}
	}()
}