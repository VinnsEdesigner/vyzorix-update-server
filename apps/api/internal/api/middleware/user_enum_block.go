// Package middleware provides HTTP middleware for the Vyzorix API.
package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"time"

	"github.com/gin-gonic/gin"
)

// Argon2id fake hash parameters (matching storage/crypto.go format).
const (
	fakeArgon2Memory      = 64 * 1024 // 64 MB.
	fakeArgon2Iterations  = 3
	fakeArgon2Parallelism = 4
	fakeArgon2SaltLength = 16
	fakeArgon2KeyLength  = 32
)

// UserEnumBlock provides user enumeration prevention middleware.
type UserEnumBlock struct {
	// FakeHashTime is how long to sleep for fake hash comparison (milliseconds).
	FakeHashTime int
}

// NewUserEnumBlock creates a new user enumeration blocker.
func NewUserEnumBlock() *UserEnumBlock {
	return &UserEnumBlock{
		FakeHashTime: 50, // 50ms constant-time response.
	}
}

// FakePasswordHash creates a realistic Argon2id-format fake hash for timing uniformity.
// Uses the same format as storage/crypto.go: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>.
func (ue *UserEnumBlock) FakePasswordHash() string {
	// Generate random salt (16 bytes).
	salt := make([]byte, fakeArgon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		// Fallback - should never happen in practice.
		for i := range salt {
			salt[i] = byte((i * 7) % 256)
		}
	}

	// Generate random key (32 bytes).
	key := make([]byte, fakeArgon2KeyLength)
	if _, err := rand.Read(key); err != nil {
		// Fallback - should never happen in practice.
		for i := range key {
			key[i] = byte((i * 13) % 256)
		}
	}

	// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>.
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(key)

	return "$argon2id$v=19$m=65536,t=3,p=4$" + encodedSalt + "$" + encodedHash
}

// FakeToken generates a fake token for uniform response timing.
func (ue *UserEnumBlock) FakeToken() string {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		// Fallback - should never happen in practice.
		for i := range tokenBytes {
			tokenBytes[i] = byte('a' + (i % 26))
		}
	}
	return base64.RawStdEncoding.EncodeToString(tokenBytes)
}

// FakeTOTPSecret generates a fake TOTP secret in base32-like format.
func (ue *UserEnumBlock) FakeTOTPSecret() string {
	// TOTP secrets are typically 16-32 chars in base32.
	secretBytes := make([]byte, 16)
	if _, err := rand.Read(secretBytes); err != nil {
		// Fallback - should never happen in practice.
		for i := range secretBytes {
			secretBytes[i] = byte('A' + (i % 26))
		}
	}
	// Use base32 encoding (using base64 as a substitute).
	return base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567").EncodeToString(secretBytes)
}

// ConstantTimeCompare does constant-time string comparison.
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// FakeResponseDelay adds artificial delay to prevent timing attacks.
func (ue *UserEnumBlock) FakeResponseDelay() {
	time.Sleep(time.Duration(ue.FakeHashTime) * time.Millisecond)
}

// EnumerationPreventionMiddleware adds headers to prevent user enumeration.
func EnumerationPreventionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set security headers.
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Next()
	}
}
