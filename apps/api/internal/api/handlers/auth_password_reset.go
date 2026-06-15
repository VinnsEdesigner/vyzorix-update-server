package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	security "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
)

// ForgotPassword handles password reset requests.
// POST /v1/auth/forgot-password.
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
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

	// Find operator by email
	op, err := ac.store.GetOperatorByEmail(ctx, email)
	if err != nil {
		ac.log.Warn("forgotPassword: db error", "err", err)
		// Still return success for security
		c.JSON(200, models.MessageResponse{Message: "If that email exists, a password reset link has been sent."})
		return
	}

	// Always return success for security (don't reveal if email exists)
	if op == nil {
		ac.log.Info("forgotPassword: email not found (silently)", "email", email)
		c.JSON(200, models.MessageResponse{Message: "If that email exists, a password reset link has been sent."})
		return
	}

	// Check if password-based account (Google-only accounts can't reset via email)
	if op.PasswordHash == "" {
		c.JSON(400, models.ErrorResponse{Error: "google_account", Message: "this account uses Google sign-in and cannot reset password via email"})
		return
	}

	// Delete old reset tokens
	if err := ac.store.DeletePasswordResetTokensByOperator(ctx, op.ID); err != nil {
		ac.log.Warn("forgotPassword: failed to delete old tokens", "err", err)
	}

	// Generate reset token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		ac.log.Warn("forgotPassword: failed to generate token", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "request failed"})
		return
	}
	token := hex.EncodeToString(tokenBytes)
	tokenHash := security.HashToken(token)

	prt := &storage.PasswordResetToken{
		ID:         GenerateID(),
		OperatorID: op.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().UTC().Add(ac.config.PasswordResetTokenExpiry),
		CreatedAt:  time.Now().UTC(),
	}

	if err := ac.store.CreatePasswordResetToken(ctx, prt); err != nil {
		ac.log.Warn("forgotPassword: failed to store token", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "request failed"})
		return
	}

	// Send email (async)
	go func() {
		if err := ac.emailSvc.SendPasswordResetEmail(context.Background(), op.Email, op.Name, token); err != nil {
			ac.log.Error("forgotPassword: failed to send email", "email", op.Email, "err", err)
		} else {
			ac.log.Info("forgotPassword: sent", "email", op.Email)
		}
	}()

	c.JSON(200, models.MessageResponse{Message: "If that email exists, a password reset link has been sent."})
}

// ResetPassword handles password reset with a valid token.
// POST /v1/auth/reset-password.
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}

	token := strings.TrimSpace(req.Token)
	newPassword := req.NewPassword

	if token == "" || newPassword == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "token and newPassword are required"})
		return
	}

	// Validate password complexity
	if err := security.ValidatePassword(newPassword, security.DefaultPasswordPolicy); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_password", Message: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Find reset token
	tokenHash := security.HashToken(token)
	prt, err := ac.store.GetPasswordResetTokenByHash(ctx, tokenHash)
	if err != nil {
		ac.log.Warn("resetPassword: db error", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "reset failed"})
		return
	}

	if prt == nil {
		c.JSON(400, models.ErrorResponse{Error: "invalid_token", Message: "invalid or expired reset token"})
		return
	}

	// Check if expired
	if time.Now().UTC().After(prt.ExpiresAt) {
		c.JSON(400, models.ErrorResponse{Error: "token_expired", Message: "reset token has expired"})
		return
	}

	// Check if already used
	if prt.UsedAt != nil {
		c.JSON(400, models.ErrorResponse{Error: "token_used", Message: "this reset token has already been used"})
		return
	}

	// Get operator and update password
	op, err := ac.store.GetOperatorByID(ctx, prt.OperatorID)
	if err != nil || op == nil {
		ac.log.Warn("resetPassword: operator not found", "operatorID", prt.OperatorID)
		c.JSON(400, models.ErrorResponse{Error: "invalid_token", Message: "invalid reset token"})
		return
	}

	// Hash new password
	newHash, err := storage.HashPassword(newPassword)
	if err != nil {
		ac.log.Warn("resetPassword: password hash failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "reset failed"})
		return
	}

	// Update operator password
	if err := ac.store.UpdateOperatorPassword(ctx, op.ID, newHash); err != nil {
		ac.log.Warn("resetPassword: update failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "reset failed"})
		return
	}

	// Mark token as used
	if err := ac.store.MarkPasswordResetTokenUsed(ctx, prt.ID); err != nil {
		ac.log.Warn("resetPassword: failed to mark token used", "err", err)
	}

	// Delete all sessions (force logout)
	if err := ac.store.DeleteAllSessionsForOperator(ctx, op.ID); err != nil {
		ac.log.Warn("resetPassword: failed to delete sessions", "err", err)
	}

	// Send confirmation email
	go func() {
		if err := ac.emailSvc.SendPasswordChangedEmail(context.Background(), op.Email, op.Name); err != nil {
			ac.log.Error("resetPassword: failed to send confirmation email", "email", op.Email, "err", err)
		}
	}()

	ac.log.Info("resetPassword: success", "operatorID", op.ID)
	c.JSON(200, models.MessageResponse{Message: "Password reset successful. Please log in with your new password."})
}

// ResendPasswordReset handles resending the password reset email with rate limiting.
// POST /v1/auth/resend-password-reset
//
// Rate limiting rules:
//   - 1st resend: immediate
//   - 2nd resend: 30 second delay
//   - 3rd resend: 60 second delay
//   - 4th+ resend: (count - 1) * 30 seconds delay
//   - After 6 attempts: 5 hour lockout
func (ac *AuthController) ResendPasswordReset(c *gin.Context) {
	var req models.ResendPasswordResetRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid request body"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "email is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Hash email for tracker lookup (privacy)
	emailHash := security.HashToken(email)
	now := time.Now().UTC()

	// Check rate limit
	tracker, err := ac.checkResendRateLimitAndGetTracker(ctx, emailHash)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		ac.log.WarnContext(ctx, "resendPasswordReset: failed to get tracker", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "failed to process request"})
		return
	}
	// Treat ErrNotFound as nil tracker (first resend)
	if errors.Is(err, storage.ErrNotFound) {
		tracker = nil
	}

	allowed, retryAfter, lockedUntil := CheckResendRateLimit(tracker, now)
	if !allowed {
		HandleRateLimitResponse(c, retryAfter, lockedUntil)
		return
	}

	// Send reset email if operator exists
	ac.sendResetEmailIfOperatorExists(ctx, email, now)

	// Update tracker and cleanup
	newResendCount := UpdateResendTracker(ctx, ac.store, ac.log, tracker, emailHash, now)
	TriggerTrackerCleanup(ac.store, ac.log)

	c.JSON(200, models.ResendPasswordResetResponse{
		Success: true,
		Message: "Password reset link sent.",
	})
	_ = newResendCount // suppress unused warning
}

// checkResendRateLimitAndGetTracker retrieves the tracker, treating ErrNotFound as nil.
func (ac *AuthController) checkResendRateLimitAndGetTracker(
	ctx context.Context,
	emailHash string,
) (*models.PasswordResetResendTracker, error) {
	tracker, err := ac.store.GetPasswordResetResendTracker(ctx, emailHash)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}
	if errors.Is(err, storage.ErrNotFound) {
		return nil, storage.ErrNotFound
	}
	return tracker, nil
}

// sendResetEmailIfOperatorExists sends the password reset email if the operator exists.
func (ac *AuthController) sendResetEmailIfOperatorExists(ctx context.Context, email string, now time.Time) {
	op, err := ac.store.GetOperatorByEmail(ctx, email)
	if err != nil {
		// Log but don't fail - operator might not exist (email enumeration protection)
		return
	}
	if op != nil {
		_ = ac.sendPasswordResetEmail(ctx, op, now) //nolint:errcheck
	}
}

// sendPasswordResetEmail handles the email sending part of resend.
func (ac *AuthController) sendPasswordResetEmail(
	ctx context.Context,
	op *models.Operator,
	now time.Time,
) error {
	// Delete any existing password reset tokens
	if err := ac.store.DeletePasswordResetTokensByOperator(ctx, op.ID); err != nil {
		ac.log.WarnContext(ctx, "resendPasswordReset: failed to delete old tokens", "err", err)
	}

	// Create new password reset token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	tokenHash := security.HashToken(token)

	prt := &storage.PasswordResetToken{
		ID:         GenerateID(),
		OperatorID: op.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  now.Add(15 * time.Minute),
		CreatedAt:  now,
	}

	if err := ac.store.CreatePasswordResetToken(ctx, prt); err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	return ac.emailSvc.SendPasswordResetEmail(ctx, op.Email, op.Name, token)
}