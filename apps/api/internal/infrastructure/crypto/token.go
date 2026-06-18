package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const (
	// TokenSize is the default token size in bytes.
	TokenSize = 32
	// TokenStringLength is the hex-encoded string length for TokenSize bytes.
	TokenStringLength = TokenSize * 2
)

var (
	// ErrTokenGenerationFailed is returned when token generation fails.
	ErrTokenGenerationFailed = fmt.Errorf("failed to generate token: crypto/rand unavailable")
)

// GenerateToken generates a cryptographically secure random token.
// Returns a hex-encoded string of TokenSize bytes.
func GenerateToken() (string, error) {
	b := make([]byte, TokenSize)
	if _, err := rand.Read(b); err != nil {
		return "", ErrTokenGenerationFailed
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
		return "", ErrTokenGenerationFailed
	}
	return hex.EncodeToString(b), nil
}

// GenerateRandomBytes generates cryptographically secure random bytes.
func GenerateRandomBytes(size int) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("invalid size: %d", size)
	}
	
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return b, nil
}
