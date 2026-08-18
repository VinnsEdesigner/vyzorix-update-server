<<<<<<< HEAD

=======
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
// Package confirmation models short-lived, single-use authorization tokens
// that gate risky device commands. A confirmation is issued for a specific
// (operator, command, device) triple and must be presented back when the
// command is actually executed. It expires after the command's risk profile
// TTL and can be consumed at most once.
package confirmation

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a confirmation does not exist or has expired.
var ErrNotFound = errors.New("confirmation not found")

// ErrExpired is returned when a confirmation exists but its TTL has elapsed.
var ErrExpired = errors.New("confirmation expired")

// ErrAlreadyConsumed is returned when a confirmation token has already been
// used to authorize a command.
var ErrAlreadyConsumed = errors.New("confirmation already consumed")

// ErrMismatch is returned when a confirmation token does not match the
// operator/command/device it was issued for.
var ErrMismatch = errors.New("confirmation does not match the request")

// PendingConfirmation is a single-use, TTL-bounded authorization for a risky
// command. It is created by the confirmation endpoint and consumed by the
// command execution handler.
type PendingConfirmation struct {
	Token      string
	OperatorID string
	OrgID      string
	Command    string
	DeviceID   string
	RiskTier   string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	// ConsumedAt is nil until the confirmation is used to authorize a command.
	ConsumedAt *time.Time
}

// IsExpired reports whether the confirmation is past its TTL.
func (p *PendingConfirmation) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// IsConsumed reports whether the confirmation has already been used.
func (p *PendingConfirmation) IsConsumed() bool {
	return p.ConsumedAt != nil
}

// Matches reports whether the confirmation was issued for the given operator,
// command, and device. An empty deviceID matches any device (the confirmation
// is scoped to operator+command when issued without a specific device).
func (p *PendingConfirmation) Matches(operatorID, command, deviceID string) bool {
	if p.OperatorID != operatorID || p.Command != command {
		return false
	}
	if p.DeviceID != "" && p.DeviceID != deviceID {
		return false
	}
	return true
}

// Repository persists pending confirmations.
type Repository interface {
	// Create stores a new pending confirmation. Tokens are unique.
	Create(ctx context.Context, c *PendingConfirmation) error
	// Get retrieves a pending confirmation by its token.
	Get(ctx context.Context, token string) (*PendingConfirmation, error)
	// Consume atomically marks a confirmation as consumed at the given time and
	// returns the updated record. It must reject already-consumed or expired
	// confirmations.
	Consume(ctx context.Context, token string, at time.Time) (*PendingConfirmation, error)
	// DeleteExpired removes expired confirmations (cleanup).
	DeleteExpired(ctx context.Context) (int64, error)
<<<<<<< HEAD
}
=======
}
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
