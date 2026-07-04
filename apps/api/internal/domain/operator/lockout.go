package operator

import (
	"context"
	"time"
)

// LockoutReason defines the reason for account lockout.
type LockoutReason string

const (
	LockoutReasonAdmin       LockoutReason = "admin_action"
	LockoutReasonBruteForce  LockoutReason = "brute_force"
	LockoutReasonSuspicious  LockoutReason = "suspicious_activity"
)

// OperatorLockout represents a temporary account lockout.
type OperatorLockout struct {
	ID          string
	OperatorID  string
	Reason      LockoutReason
	LockedBy    string // Operator ID that initiated lockout (empty for system)
	Message     string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

// IsActive returns true if the lockout is currently active.
func (l *OperatorLockout) IsActive() bool {
	return time.Now().Before(l.ExpiresAt)
}

// OperatorLockoutRepository defines the interface for lockout data access.
type OperatorLockoutRepository interface {
	// Create creates a new lockout record.
	Create(ctx context.Context, l *OperatorLockout) error

	// FindActiveByOperator retrieves active lockout for an operator.
	FindActiveByOperator(ctx context.Context, operatorID string) (*OperatorLockout, error)

	// Delete removes a lockout record.
	Delete(ctx context.Context, id string) error

	// DeleteByOperator removes all lockout records for an operator.
	DeleteByOperator(ctx context.Context, operatorID string) error

	// DeleteExpired removes expired lockout records.
	DeleteExpired(ctx context.Context) (int, error)
}
