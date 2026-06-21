package refresh_token

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a refresh token is not found.
var ErrNotFound = errors.New("refresh token not found")

// ErrExpired is returned when a refresh token has expired.
var ErrExpired = errors.New("refresh token expired")

// ErrRevoked is returned when a refresh token has been revoked.
var ErrRevoked = errors.New("refresh token revoked")

// RefreshToken represents a refresh token for session rotation.
type RefreshToken struct {
	ID            string    // Unique identifier (UUIDv7)
	TokenHash     string    // SHA-256 hash of the actual token
	OperatorID    string    // The operator this token belongs to
	SessionID     string    // The session this token is associated with
	ExpiresAt     time.Time // When this token expires
	CreatedAt     time.Time // When this token was created
	RevokedAt     time.Time // When this token was revoked (if applicable)
	ReplacedByID  string    // ID of the token that replaced this one (for rotation)
	IsRevoked     bool      // Whether this token has been revoked
}

// IsExpired returns true if the refresh token has expired.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid returns true if the refresh token is valid (not expired and not revoked).
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked
}

// TokenFamily represents a family of rotated refresh tokens.
// When a token is rotated, the old one is marked as revoked and replaced by the new one.
// This allows for detection of token theft (if an old token is used after rotation).
type TokenFamily struct {
	FamilyID   string    // Unique identifier for this token family
	TokenID    string    // Current active token ID in this family
	OperatorID string    // The operator this family belongs to
	CreatedAt  time.Time // When this family was created
	ExpiresAt  time.Time // When this family expires (last token in family)
}
