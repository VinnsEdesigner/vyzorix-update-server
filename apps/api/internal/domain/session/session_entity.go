package session

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a session is not found.
var ErrNotFound = errors.New("session not found")

// Session represents an authentication session.
type Session struct {
	ID         string
	OperatorID string
	ExpiresAt  time.Time
	CreatedAt  time.Time

	// Optional metadata.
	IPAddress string
	UserAgent string
}

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid returns true if the session is valid (not expired).
func (s *Session) IsValid() bool {
	return !s.IsExpired()
}

// TimeUntilExpiry returns the duration until the session expires.
func (s *Session) TimeUntilExpiry() time.Duration {
	return time.Until(s.ExpiresAt)
}

// RemainingLifetime returns the remaining lifetime as a percentage (0-100).
func (s *Session) RemainingLifetime() int {
	total := s.ExpiresAt.Sub(s.CreatedAt).Seconds()
	remaining := s.TimeUntilExpiry().Seconds()

	// Avoid division by zero
	if total == 0 {
		return 0
	}

	return int((remaining / total) * 100)
}

// SessionRevocation represents a revoked session entry.
type SessionRevocation struct {
	TokenHash string
	RevokedAt time.Time
	Reason    string
}
