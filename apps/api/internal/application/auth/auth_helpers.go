package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// hashTokenSha256 creates a SHA-256 hash of a token.
func hashTokenSha256(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// hashEmailForTracker normalizes email for tracking in resend trackers.
func hashEmailForTracker(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	return hashTokenSha256(email)[:16]
}
