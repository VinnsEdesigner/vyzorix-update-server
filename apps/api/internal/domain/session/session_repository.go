package session

import (
	"context"
	"time"
)

// Repository defines the interface for session data access.
type Repository interface {
	// FindByID retrieves a session by ID.
	FindByID(ctx context.Context, id string) (*Session, error)

	// FindByOperatorID retrieves all sessions for an operator.
	FindByOperatorID(ctx context.Context, operatorID string) ([]*Session, error)

	// ListActiveByOperator retrieves active (non-expired) sessions for an operator, ordered by creation time.
	ListActiveByOperator(ctx context.Context, operatorID string) ([]*Session, error)

	// Create creates a new session.
	Create(ctx context.Context, s *Session) error

	// Delete deletes a session.
	Delete(ctx context.Context, id string) error

	// DeleteByOperatorID deletes all sessions for an operator.
	DeleteByOperatorID(ctx context.Context, operatorID string) error

	// DeleteExpired deletes all expired sessions.
	DeleteExpired(ctx context.Context) (int, error)

	// Extend extends a session's expiration time.
	Extend(ctx context.Context, id string, newExpiry time.Time) error

	// UpdateOrganizationID updates the selected organization ID for a session.
	UpdateOrganizationID(ctx context.Context, sessionID, organizationID string) error

	// SetMFAVerifiedAt sets the MFAVerifiedAt timestamp for a session.
	// This is called after successful MFA verification during login.
	SetMFAVerifiedAt(ctx context.Context, sessionID string, verifiedAt time.Time) error

	// AddSessionRevocation adds a session token hash to the revocation list.
	AddSessionRevocation(ctx context.Context, tokenHash, reason string) error

	// IsSessionRevoked checks if a session token hash is in the revocation list.
	IsSessionRevoked(ctx context.Context, tokenHash string) (bool, error)

	// RemoveSessionRevocation removes a session from the revocation list.
	RemoveSessionRevocation(ctx context.Context, tokenHash string) error

	// ListSessionRevocations retrieves all revoked sessions, optionally filtered by reason.
	ListSessionRevocations(ctx context.Context, reason string, limit int) ([]*SessionRevocation, error)

	// CleanupSessionRevocations removes revocation entries older than the specified duration.
	CleanupSessionRevocations(ctx context.Context, olderThan time.Duration) (int, error)

	// RevokeAllOperatorSessions revokes all sessions for a specific operator.
	RevokeAllOperatorSessions(ctx context.Context, operatorID string) error
}
