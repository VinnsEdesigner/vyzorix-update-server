package email_verification

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when an email verification is not found.
var ErrNotFound = errors.New("email verification not found")

// ErrExpired is returned when the verification token has expired.
var ErrExpired = errors.New("verification token expired")

// EmailVerification represents a pending email verification token.
type EmailVerification struct {
	ExpiresAt  time.Time
	CreatedAt  time.Time
	ID         string
	OperatorID string
	TokenHash  string
}

// IsExpired returns true if the verification has expired.
func (e *EmailVerification) IsExpired() bool {
	return time.Now().UTC().After(e.ExpiresAt)
}

// Repository defines the interface for email verification data access.
type Repository interface {
	// Create creates a new email verification.
	Create(ctx context.Context, ev *EmailVerification) error

	// FindByTokenHash retrieves an email verification by token hash.
	FindByTokenHash(ctx context.Context, tokenHash string) (*EmailVerification, error)

	// Delete removes an email verification by ID.
	Delete(ctx context.Context, id string) error

	// DeleteByOperator removes all verifications for an operator.
	DeleteByOperator(ctx context.Context, operatorID string) error
}
