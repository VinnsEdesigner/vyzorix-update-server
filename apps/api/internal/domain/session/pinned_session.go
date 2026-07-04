package session

import (
	"context"
	"time"
)

// PinnedSession represents a trusted device session that doesn't require MFA on subsequent logins.
type PinnedSession struct {
	ID          string
	SessionID   string
	OperatorID  string
	Fingerprint string // Device fingerprint hash
	DeviceName  string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	IsActive    bool
}

// Repository defines the interface for pinned session data access.
type PinnedSessionRepository interface {
	// Create creates a new pinned session.
	Create(ctx context.Context, p *PinnedSession) error

	// FindBySessionID retrieves a pinned session by session ID.
	FindBySessionID(ctx context.Context, sessionID string) (*PinnedSession, error)

	// FindActiveByOperator retrieves all active pinned sessions for an operator.
	FindActiveByOperator(ctx context.Context, operatorID string) ([]*PinnedSession, error)

	// Delete removes a pinned session.
	Delete(ctx context.Context, id string) error

	// DeleteBySessionID removes a pinned session by session ID.
	DeleteBySessionID(ctx context.Context, sessionID string) error

	// DeleteExpired removes expired pinned sessions.
	DeleteExpired(ctx context.Context) (int, error)
}
