package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/password_reset"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
)

// PasswordResetToken represents a password reset token.
type PasswordResetToken struct {
	Token     string
	ExpiresAt time.Time
	Email     string
}

// ResendRateLimitResult holds the result of a resend rate limit check.
type ResendRateLimitResult struct {
	LockedUntil *time.Time
	RetryAfter  int
	TotalSent   int
	Allowed     bool
}

// MaxResendCount is the maximum number of password reset requests allowed.
const MaxResendCount = 5

// ResendLockoutDuration is how long the lockout lasts after exceeding MaxResendCount.
const ResendLockoutDuration = 15 * time.Minute

// GeneratePasswordResetToken generates a password reset token.
func (s *AuthService) GeneratePasswordResetToken(ctx context.Context, email string) (string, error) {
	// Normalize email.
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if operator exists.
	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, operator.ErrNotFound) {
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
		if errors.Is(err, password_reset.ErrNotFound) {
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

	// Check if password was breached (using k-anonymity to avoid sending password to external API).
	if breached, _ := infraauth.CheckPasswordBreached(newPassword); breached {
		return application.ErrPasswordBreached
	}

	// Hash the token.
	tokenHash := hashTokenSha256(token)

	// Look up token.
	prt, err := s.passwordResetRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, password_reset.ErrNotFound) {
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
		if errors.Is(err, password_reset.ErrNotFound) {
			return &ResendRateLimitResult{Allowed: true, TotalSent: 0}, nil
		}

		return nil, err
	}

	// Check if currently locked out.
	if tracker.IsLockedOut() {
		return &ResendRateLimitResult{
			Allowed:     false,
			LockedUntil: tracker.LockoutUntil,
			RetryAfter:  int(time.Until(*tracker.LockoutUntil).Seconds()),
			TotalSent:   tracker.ResendCount,
		}, nil
	}

	return &ResendRateLimitResult{
		Allowed:    true,
		RetryAfter: 0,
		TotalSent:  tracker.ResendCount,
	}, nil
}

// UpdateResendTracker updates the resend tracker for an email.
func (s *AuthService) UpdateResendTracker(ctx context.Context, email string) error {
	emailHash := hashEmailForTracker(email)
	now := time.Now().UTC()

	var newCount int
	var lockoutUntil *time.Time

	// Get existing tracker.
	tracker, err := s.passwordResetRepo.GetResendTracker(ctx, emailHash)
	if err != nil {
		if errors.Is(err, password_reset.ErrNotFound) {
			// Create new tracker.
			newCount = 1
		} else {
			return err
		}
	} else {
		newCount, lockoutUntil = s.calculateResendCount(tracker)
	}

	// Generate ID for tracker.
	id := shared.GenerateID()

	newTracker := &password_reset.ResendTracker{
		ID:           id,
		EmailHash:    emailHash,
		ResendCount:  newCount,
		LastResendAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if lockoutUntil != nil {
		newTracker.LockoutUntil = lockoutUntil
	}

	// If tracker exists, preserve created_at.
	if tracker != nil && !tracker.IsLockedOut() {
		newTracker.ID = tracker.ID
		newTracker.CreatedAt = tracker.CreatedAt
	}

	return s.passwordResetRepo.UpsertResendTracker(ctx, newTracker)
}

// calculateResendCount calculates the resend count and wait time.
func (s *AuthService) calculateResendCount(tracker *password_reset.ResendTracker) (int, *time.Time) {
	if tracker != nil && tracker.IsLockedOut() {
		return tracker.ResendCount, tracker.LockoutUntil
	}

	var newCount int
	if tracker == nil {
		newCount = 1
	} else {
		newCount = tracker.ResendCount + 1
	}

	if newCount > MaxResendCount {
		lockout := time.Now().UTC().Add(ResendLockoutDuration)
		if tracker != nil {
			return tracker.ResendCount, &lockout
		}
		return 0, &lockout
	}

	return newCount, nil
}

// DeleteResendTracker deletes the resend tracker for an email.
func (s *AuthService) DeleteResendTracker(ctx context.Context, email string) error {
	emailHash := hashEmailForTracker(email)

	return s.passwordResetRepo.DeleteResendTracker(ctx, emailHash)
}
