package auth

import (
	"crypto/subtle"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	// DummyPasswordHashDuration is the target duration for fake password hash computation.
	DummyPasswordHashDuration = 100 * time.Millisecond
)

var (
	// dummySalt is a static salt for fake password hash computation.
	dummySalt = []byte("dummy_salt_for_timing_uniformity_16b")
)

// ConstantTimeCompare performs a constant-time comparison of two strings.
// Returns true if they are equal, false otherwise.
// This prevents timing attacks on string comparisons.
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ConstantTimeValidate validates a value against an expected value.
// This is used for timing-safe validation of secrets.
func ConstantTimeValidate(expected, actual string) bool {
	if len(expected) != len(actual) {
		subtle.ConstantTimeCompare([]byte(expected), []byte(actual))
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}

// ComputeFakePasswordHash performs dummy computation to maintain constant timing.
// for user enumeration prevention. It computes an Argon2id hash to match the.
// time cost of a real password verification.
func ComputeFakePasswordHash() {
	// Compute Argon2id with dummy inputs.
	argon2.IDKey(
		[]byte("dummy_password_for_timing_uniformity"),
		dummySalt,
		3,              // iterations
		64*1024,        // memory (64 MB)
		4,              // parallelism
		32,             // key length
	)
}

// FakePasswordHash performs a fake password hash computation for timing uniformity.
// This should be called when a fake hash is needed (e.g., wrong username).
// to ensure consistent timing regardless of whether the user exists.
func FakePasswordHash() string {
	ComputeFakePasswordHash()
	return "$argon2id$fake$hash$for$timing$uniformity"
}
