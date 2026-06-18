// Package auth provides authentication utilities.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/argon2"
)

// LockoutConfig holds account lockout configuration.
type LockoutConfig struct {
	// Enabled enables account lockout.
	Enabled bool

	// MaxAttempts is the number of failed attempts before lockout.
	MaxAttempts int

	// LockoutDuration is how long the account is locked (seconds).
	LockoutDuration int

	// ResetAfter is when failed attempts are reset (seconds).
	ResetAfter int
}

// DefaultLockoutConfig returns the default lockout configuration.
func DefaultLockoutConfig() LockoutConfig {
	return LockoutConfig{
		Enabled:         true,
		MaxAttempts:    5,
		LockoutDuration: 3600, // 1 hour
		ResetAfter:     1800, // 30 minutes
	}
}

// LockoutReason holds information about why an account was locked.
type LockoutReason struct {
	Reason    string `json:"reason"`
	Until    int64  `json:"until,omitempty"`
	Attempts  int    `json:"attempts"`
}

// AccountLockout provides account lockout functionality.
type AccountLockout struct {
	storage Storage
	config  LockoutConfig
}

// NewAccountLockout creates a new account lockout handler.
func NewAccountLockout(storage Storage, config LockoutConfig) *AccountLockout {
	return &AccountLockout{
		storage: storage,
		config:  config,
	}
}

// IsLocked checks if an account is currently locked.
func (al *AccountLockout) IsLocked(ctx context.Context, operatorID string) (bool, *LockoutReason, error) {
	if !al.config.Enabled {
		return false, nil, nil
	}

	lockedUntil, err := al.storage.GetAccountLockout(ctx, operatorID)
	if err != nil {
		return false, nil, err
	}

	if lockedUntil == 0 {
		// Not locked.
		return false, nil, nil
	}

	now := time.Now().UnixMilli()
	if lockedUntil > now {
		// Still locked.
		return true, &LockoutReason{
			Reason: "Too many failed login attempts",
			Until: lockedUntil,
		}, nil
	}

	// Lock expired, clear it.
	_ = al.storage.ClearAccountLockout(ctx, operatorID)
	_ = al.storage.ClearFailedAttempts(ctx, operatorID)

	return false, nil, nil
}

// RecordFailedAttempt records a failed login attempt.
func (al *AccountLockout) RecordFailedAttempt(ctx context.Context, operatorID string) error {
	if !al.config.Enabled {
		return nil
	}

	// Check current attempts.
	attempts, err := al.storage.GetFailedAttempts(ctx, operatorID)
	if err != nil {
		return err
	}

	attempts++
	
	if err := al.storage.SetFailedAttempts(ctx, operatorID, attempts); err != nil {
		return err
	}

	// Check if we should lock.
	if attempts >= al.config.MaxAttempts {
		lockoutUntil := time.Now().Add(time.Duration(al.config.LockoutDuration) * time.Second).UnixMilli()
		return al.storage.SetAccountLockout(ctx, operatorID, lockoutUntil)
	}

	return nil
}

// ClearLockout clears the lockout for an account.
func (al *AccountLockout) ClearLockout(ctx context.Context, operatorID string) error {
	if err := al.storage.ClearAccountLockout(ctx, operatorID); err != nil {
		return err
	}
	return al.storage.ClearFailedAttempts(ctx, operatorID)
}

// Storage interface for account lockout operations.
type Storage interface {
	GetAccountLockout(ctx context.Context, operatorID string) (int64, error)
	SetAccountLockout(ctx context.Context, operatorID string, until int64) error
	ClearAccountLockout(ctx context.Context, operatorID string) error
	GetFailedAttempts(ctx context.Context, operatorID string) (int, error)
	SetFailedAttempts(ctx context.Context, operatorID string, attempts int) error
	ClearFailedAttempts(ctx context.Context, operatorID string) error
}

// ErrInvalidPassword is returned when password validation fails.
var ErrInvalidPassword = errors.New("invalid password")

// Pre-computed dummy hash for timing uniformity.
// This hash is used in IsValidPassword to ensure constant-time validation.
// regardless of whether the password matches or not.
// Generated using Argon2id with a fixed dummy password.
var dummyPasswordHash = argon2.IDKey(
	[]byte("dummy_password_for_timing_uniformity"),
	[]byte("固定的盐值用于虚拟密码哈希"), // Static salt - 16 bytes
	3,               // iterations
	64*1024,        // memory (64 MB)
	4,              // parallelism
	32,             // key length
)

// FakeHash performs constant-time comparison to prevent timing attacks.
func FakeHash(a, b string) bool {
	if len(a) != len(b) {
		// Still do comparison to maintain constant time.
		subtle.ConstantTimeCompare([]byte(a), []byte(b))
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// IsValidPassword is a dummy function for user enumeration prevention.
// It performs a real Argon2id hash computation to maintain constant-time behavior.
// regardless of input, preventing timing attacks that could reveal valid usernames.
// Uses a pre-computed dummy hash to avoid expensive computation on every call.
func IsValidPassword(password string) bool {
	// Always take same time regardless of password validity.
	// Uses real Argon2id computation for proper timing uniformity.
	if len(password) == 0 {
		return false
	}
	
	// Compute Argon2id of input against dummy salt.
	// This ensures consistent timing whether password is "valid" or not.
	computedHash := argon2.IDKey(
		[]byte(password),
		[]byte("固定的盐值用于虚拟密码哈希"), // Same salt as dummy hash
		3,
		64*1024,
		4,
		32,
	)
	
	// Constant-time comparison with dummy hash.
	return subtle.ConstantTimeCompare(computedHash, dummyPasswordHash) == 1
}

// ErrTokenGenerationFailed is returned when token generation fails.
var ErrTokenGenerationFailed = errors.New("failed to generate token: crypto/rand unavailable")

// GenerateFakeToken generates a cryptographically secure fake token for timing uniformity.
// Uses crypto/rand directly - panics if unavailable since predictable tokens are a security risk.
func GenerateFakeToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand MUST be available in production. If it's not, the system.
		// is in a critical state and generating predictable tokens would be worse.
		panic(ErrTokenGenerationFailed)
	}
	return "fake_" + hex.EncodeToString(b)
}
