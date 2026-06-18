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
}
