// Package lockout provides account lockout functionality for brute force protection.
package lockout

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/crypto/argon2"
)

// Config holds account lockout configuration.
type Config struct {
	Enabled         bool
	MaxAttempts     int
	LockoutDuration int
	ResetAfter      int
}

// DefaultConfig returns the default lockout configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:         true,
		MaxAttempts:     5,
		LockoutDuration: 3600,
		ResetAfter:      1800,
	}
}

// Reason holds information about why an account was locked.
type Reason struct {
	Reason   string `json:"reason"`
	Until    int64  `json:"until,omitempty"`
	Attempts int    `json:"attempts"`
}

// Storage interface for account lockout operations.
type Storage interface {
	GetAccountLockout(ctx context.Context, operatorID string) (int64, error)
	SetAccountLockout(ctx context.Context, operatorID string, until int64) error
	ClearAccountLockout(ctx context.Context, operatorID string) error
	GetFailedAttempts(ctx context.Context, operatorID string) (int, error)
	SetFailedAttempts(ctx context.Context, operatorID string, attempts int) error
	ClearFailedAttempts(ctx context.Context, operatorID string) error
	// IncrementFailedAttempts atomically increments and returns the new count.
	// This prevents race conditions in concurrent login attempts.
	IncrementFailedAttempts(ctx context.Context, operatorID string) (int, error)
}

// Handler provides account lockout functionality.
type Handler struct {
	storage Storage
	logger  *slog.Logger
	config  Config
}

// New creates a new account lockout handler.
func New(storage Storage, config Config) *Handler {
	return &Handler{
		storage: storage,
		config:  config,
		logger:  slog.Default(),
	}
}

// IsLocked checks if an account is currently locked.
func (h *Handler) IsLocked(ctx context.Context, operatorID string) (bool, *Reason, error) {
	if !h.config.Enabled {
		return false, nil, nil
	}

	lockedUntil, err := h.storage.GetAccountLockout(ctx, operatorID)
	if err != nil {
		return false, nil, err
	}

	if lockedUntil == 0 {
		return false, nil, nil
	}

	now := time.Now().UnixMilli()
	if lockedUntil > now {
		return true, &Reason{
			Reason: "Too many failed login attempts",
			Until:  lockedUntil,
		}, nil
	}

	// Expired lockout - clear it and log any errors.
	if err := h.storage.ClearAccountLockout(ctx, operatorID); err != nil {
		h.logger.Warn("failed to clear expired lockout",
			"operator_id", operatorID,
			"error", err)
	}
	if err := h.storage.ClearFailedAttempts(ctx, operatorID); err != nil {
		h.logger.Warn("failed to clear failed attempts",
			"operator_id", operatorID,
			"error", err)
	}

	return false, nil, nil
}

// RecordFailedAttempt records a failed login attempt using atomic increment.
func (h *Handler) RecordFailedAttempt(ctx context.Context, operatorID string) error {
	if !h.config.Enabled {
		return nil
	}

	// Use atomic increment to prevent race conditions.
	attempts, err := h.storage.IncrementFailedAttempts(ctx, operatorID)
	if err != nil {
		return err
	}

	if attempts >= h.config.MaxAttempts {
		lockoutUntil := time.Now().Add(time.Duration(h.config.LockoutDuration) * time.Second).UnixMilli()
		return h.storage.SetAccountLockout(ctx, operatorID, lockoutUntil)
	}

	return nil
}

// ClearLockout clears the lockout for an account.
func (h *Handler) ClearLockout(ctx context.Context, operatorID string) error {
	if err := h.storage.ClearAccountLockout(ctx, operatorID); err != nil {
		return err
	}

	return h.storage.ClearFailedAttempts(ctx, operatorID)
}

var (
	ErrInvalidPassword       = errors.New("invalid password")
	ErrTokenGenerationFailed = errors.New("failed to generate token: crypto/rand unavailable")
)

// Pre-computed dummy hash for timing uniformity - never matches any real password.
var dummyPasswordHash = argon2.IDKey(
	[]byte("dummy_password_for_timing_uniformity"),
	[]byte(""),
	3,
	64*1024,
	4,
	32,
)

// FakeHash performs constant-time comparison to prevent timing attacks.
func FakeHash(a, b string) bool {
	// Always perform constant-time comparison to avoid timing leaks.
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// IsValidPassword always returns false to prevent user enumeration.
// This is a timing-attack mitigation - even if an attacker tries different.
// passwords, the function always takes the same time to return false.
func IsValidPassword(password string) bool {
	if len(password) == 0 {
		return false
	}

	// Compute a hash using the same parameters but with the provided password.
	// This ensures consistent timing regardless of input.
	computedHash := argon2.IDKey(
		[]byte(password),
		[]byte(""),
		3,
		64*1024,
		4,
		32,
	)

	// Constant-time comparison with the dummy hash.
	_ = subtle.ConstantTimeCompare(computedHash, dummyPasswordHash)

	// Always return false - we never validate passwords here.
	return false
}

// GenerateFakeToken generates a cryptographically secure fake token.
// Returns an error instead of panicking on crypto failure.
func GenerateFakeToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", ErrTokenGenerationFailed
	}

	return "fake_" + hex.EncodeToString(b), nil
}
