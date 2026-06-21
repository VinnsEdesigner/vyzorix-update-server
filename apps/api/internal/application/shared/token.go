package shared

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/uuid"
)

const (
	// TokenSize is the default token size in bytes.
	TokenSize = 32
	// ResetTokenSize is the size for password reset tokens.
	ResetTokenSize = 32
	// VerificationTokenSize is the size for email verification tokens.
	VerificationTokenSize = 32
)

// GenerateToken generates a cryptographically secure random token.
// Returns a hex-encoded string of TokenSize bytes.
func GenerateToken() (string, error) {
	b := make([]byte, TokenSize)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// GenerateTokenWithSize generates a cryptographically secure random token of specified size.
// Returns a hex-encoded string.
func GenerateTokenWithSize(size int) (string, error) {
	if size <= 0 {
		return "", fmt.Errorf("invalid token size: %d", size)
	}

	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// GenerateResetToken generates a password reset token.
func GenerateResetToken() (string, error) {
	return GenerateTokenWithSize(ResetTokenSize)
}

// GenerateVerificationToken generates an email verification token.
func GenerateVerificationToken() (string, error) {
	return GenerateTokenWithSize(VerificationTokenSize)
}

// GenerateID generates a UUIDv7 for time-ordered unique identifiers.
// Use this for entity IDs that benefit from timestamp ordering.
func GenerateID() string {
	return uuid.New()
}
