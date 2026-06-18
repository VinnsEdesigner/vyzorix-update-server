package password_reset

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a password reset token is not found.
var ErrNotFound = errors.New("password reset token not found")

// ErrExpired is returned when the token has expired.
var ErrExpired = errors.New("password reset token expired")

// ErrUsed is returned when the token has already been used.
var ErrUsed = errors.New("password reset token already used")

// PasswordResetToken represents a pending password reset token.
type PasswordResetToken struct {
	ID         string
	OperatorID string
	TokenHash  string
	ExpiresAt  time.Time
	UsedAt     *time.Time
	CreatedAt  time.Time
}

// IsExpired returns true if the token has expired.
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

// IsUsed returns true if the token has been used.
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// ResendTracker tracks password reset resend attempts for rate limiting.
type ResendTracker struct {
	ID           string
	EmailHash    string
	ResendCount  int
	LastResendAt time.Time
	LockoutUntil *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsLockedOut returns true if the tracker is currently in lockout.
func (t *ResendTracker) IsLockedOut() bool {
	if t.LockoutUntil == nil {
		return false
	}
	return time.Now().UTC().Before(*t.LockoutUntil)
}

// Repository defines the interface for password reset token data access.
type Repository interface {
	// Create creates a new password reset token.
	Create(ctx context.Context, token *PasswordResetToken) error
	
	// FindByTokenHash retrieves a token by its hash.
	FindByTokenHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	
	// MarkUsed marks a token as used.
	MarkUsed(ctx context.Context, id string) error
	
	// DeleteByOperator removes all tokens for an operator.
	DeleteByOperator(ctx context.Context, operatorID string) error
	
	// GetResendTracker retrieves the resend tracker for an email hash.
	GetResendTracker(ctx context.Context, emailHash string) (*ResendTracker, error)
	
	// UpsertResendTracker creates or updates a resend tracker.
	UpsertResendTracker(ctx context.Context, tracker *ResendTracker) error
	
	// DeleteResendTracker removes a resend tracker by email hash.
	DeleteResendTracker(ctx context.Context, emailHash string) error
	
	// CleanupResendTrackers removes old resend trackers.
	CleanupResendTrackers(ctx context.Context, maxAgeHours int) (int64, error)
}
