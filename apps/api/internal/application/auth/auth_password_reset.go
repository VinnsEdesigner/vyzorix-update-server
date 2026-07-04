package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/password_reset"
)

// PasswordResetToken represents a password reset token.
type PasswordResetToken struct {
	Token     string
	ExpiresAt time.Time
	Email     string
}

// ResendRateLimitResult holds the result of a resend rate limit check.
type ResendRateLimitResult struct {
	Allowed     bool
	LockedUntil *time.Time
	RetryAfter  int
	TotalSent   int
}

// GeneratePasswordResetToken generates a password reset token.
func (s *AuthService) GeneratePasswordResetToken(ctx context.Context, email string) (string, error) {
	// Normalize email.
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if operator exists.
	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == operator.ErrNotFound {
			// Return success anyway to prevent email enumeration.
			return "", nil
		}

		return "", err
	}

	// Delete old tokens for this operator.
	_ = s.passwordResetRepo.DeleteByOperator(ctx, op.ID)

	// Generate token.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)
	tokenHash := hashTokenSha256(token)

	// Generate ID.
	id := shared.GenerateID()

	// Create reset token record.
	prt := &password_reset.PasswordResetToken{
		ID:         id,
		OperatorID: op.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().UTC().Add(15 * time.Minute),
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.passwordResetRepo.Create(ctx, prt); err != nil {
		return "", err
	}

	return token, nil
}

// ValidatePasswordResetToken validates a password reset token.
func (s *AuthService) ValidatePasswordResetToken(ctx context.Context, token, email string) error {
	// Normalize email.
	email = strings.ToLower(strings.TrimSpace(email))

	// Hash the token.
	tokenHash := hashTokenSha256(token)

	// Look up token.
	prt, err := s.passwordResetRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == password_reset.ErrNotFound {
			return application.ErrInvalidInput
		}

		return err
	}

	// Check if expired.
	if prt.IsExpired() {
		return application.ErrTokenExpired
	}

	// Check if already used.
	if prt.IsUsed() {
		return password_reset.ErrUsed
	}

	// Verify operator email matches.
	op, err := s.operatorRepo.FindByID(ctx, prt.OperatorID)
	if err != nil {
		return err
	}

	if strings.ToLower(op.Email) != email {
		return application.ErrInvalidInput
	}

	return nil
}

// ResetPassword resets a password using a token.
func (s *AuthService) ResetPassword(ctx context.Context, token, _email string, newPassword string) error {
	// Note: email parameter kept for API compatibility but operator is derived from token.
	// Validate password.
	if err := ValidatePassword(newPassword, DefaultPasswordPolicy); err != nil {
		return application.ErrInvalidInput
	}

	// Hash the token.
	tokenHash := hashTokenSha256(token)

	// Look up token.
	prt, err := s.passwordResetRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if err == password_reset.ErrNotFound {
			return application.ErrInvalidInput
		}
		return err
	}

	// Check if expired.
	if prt.IsExpired() {
		return application.ErrTokenExpired
	}

	// Check if already used.
	if prt.IsUsed() {
		return password_reset.ErrUsed
	}

	// Get operator.
	op, err := s.operatorRepo.FindByID(ctx, prt.OperatorID)
	if err != nil {
		return err
	}

	// Hash new password.
	hash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return err
	}

	// Update operator password.
	op.PasswordHash = hash
	op.UpdatedAt = time.Now()

	if err := s.operatorRepo.Update(ctx, op); err != nil {
		return err
	}

	// Mark token as used.
	_ = s.passwordResetRepo.MarkUsed(ctx, prt.ID)

	// Invalidate all sessions and refresh tokens.
	_ = s.LogoutAll(ctx, op.ID)
	_ = s.RevokeAllRefreshTokens(ctx, op.ID)

	return nil
}

// GetResendTracker gets the resend tracker for an email.
func (s *AuthService) GetResendTracker(ctx context.Context, email string) (*password_reset.ResendTracker, error) {
	emailHash := hashEmailForTracker(email)

	return s.passwordResetRepo.GetResendTracker(ctx, emailHash)
}

// CheckResendRateLimit checks the resend rate limit.
func (s *AuthService) CheckResendRateLimit(ctx context.Context, email string) (*ResendRateLimitResult, error) {
	tracker, err := s.GetResendTracker(ctx, email)
	if err != nil {
		if err == password_reset.ErrNotFound {
			return &ResendRateLimitResult{Allowed: true, TotalSent: 0}, nil
		}

		return nil, err
	}

	canResend, waitSeconds, totalSent := s.calculateResendCount(tracker)

	if !canResend {
		lockedUntil := time.Now().Add(time.Duration(waitSeconds) * time.Second)
		return &ResendRateLimitResult{
			Allowed:     false,
			LockedUntil: &lockedUntil,
			RetryAfter:  waitSeconds,
			TotalSent:   totalSent,
		}, nil
	}

	return &ResendRateLimitResult{
		Allowed:    true,
		RetryAfter: 0,
		TotalSent:  totalSent,
	}, nil
}

// UpdateResendTracker updates the resend tracker for an email.
func (s *AuthService) UpdateResendTracker(ctx context.Context, email string) error {
	emailHash := hashEmailForTracker(email)

	// Get existing tracker.
	tracker, err := s.passwordResetRepo.GetResendTracker(ctx, emailHash)
	if err != nil {
		if err == password_reset.ErrNotFound {
			// Create new tracker.
			tracker = &password_reset.ResendTracker{
				ID:           shared.GenerateID(),
				EmailHash:    emailHash,
				ResendCount:  1,
				LastResendAt: time.Now(),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			return s.passwordResetRepo.UpsertResendTracker(ctx, tracker)
		}

		return err
	}

	// Update existing tracker.
	tracker.ResendCount++
	tracker.LastResendAt = time.Now()
	tracker.UpdatedAt = time.Now()

	return s.passwordResetRepo.UpsertResendTracker(ctx, tracker)
}

// calculateResendCount calculates the resend count and wait time.
func (s *AuthService) calculateResendCount(tracker *password_reset.ResendTracker) (bool, int, int) {
	if tracker == nil {
		return true, 0, 0
	}

	count := tracker.ResendCount
	elapsed := time.Since(tracker.LastResendAt)

	switch count {
	case 0:
		return true, 0, 0
	case 1:
		wait := 30*time.Second - elapsed
		if wait <= 0 {
			return true, 0, 1
		}
		return false, int(wait.Seconds()), 1
	case 2:
		wait := 60*time.Second - elapsed
		if wait <= 0 {
			return true, 0, 2
		}
		return false, int(wait.Seconds()), 2
	case 3:
		wait := 2*time.Minute - elapsed
		if wait <= 0 {
			return true, 0, 3
		}
		return false, int(wait.Seconds()), 3
	default:
		wait := 5*time.Minute - elapsed
		if wait <= 0 {
			return true, 0, count
		}
		return false, int(wait.Seconds()), count
	}
}

// DeleteResendTracker deletes the resend tracker for an email.
func (s *AuthService) DeleteResendTracker(ctx context.Context, email string) error {
	emailHash := hashEmailForTracker(email)

	return s.passwordResetRepo.DeleteResendTracker(ctx, emailHash)
}
