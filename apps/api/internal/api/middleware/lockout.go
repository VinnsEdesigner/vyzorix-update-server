// Package middleware provides HTTP middleware.
package middleware

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrAccountLocked is returned when an account is locked.
var ErrAccountLocked = errors.New("account locked")

// LockoutConfig holds lockout configuration.
type LockoutConfig struct {
	Enabled           bool
	MaxAttempts       int
	LockoutDuration   time.Duration
	MaxLockoutDuration time.Duration
}

// DefaultLockoutConfig returns default lockout configuration.
func DefaultLockoutConfig() LockoutConfig {
	return LockoutConfig{
		Enabled:           false,
		MaxAttempts:       5,
		LockoutDuration:   time.Hour,
		MaxLockoutDuration: 24 * time.Hour,
	}
}

type attemptInfo struct {
	count        int
	firstAt     time.Time
	lockedUntil *time.Time
}

// Lockout tracks failed login attempts.
type Lockout struct {
	mu       sync.RWMutex
	attempts map[string]*attemptInfo
	config   LockoutConfig
}

// IsEnabled returns whether lockout is enabled.
func (l *Lockout) IsEnabled() bool {
	return l.config.Enabled
}

// NewLockout creates a new lockout tracker.
func NewLockout(config LockoutConfig) *Lockout {
	return &Lockout{
		attempts: make(map[string]*attemptInfo),
		config:  config,
	}
}

// RecordFailedAttempt records a failed login attempt.
func (l *Lockout) RecordFailedAttempt(email string) error {
	if !l.config.Enabled {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	info, exists := l.attempts[email]
	now := time.Now()

	if !exists {
		l.attempts[email] = &attemptInfo{
			count:    1,
			firstAt: now,
		}
		return nil
	}

	if info.lockedUntil != nil && now.Before(*info.lockedUntil) {
		return ErrAccountLocked
	}

	info.count++

	if info.count >= l.config.MaxAttempts {
		multiplier := info.count - l.config.MaxAttempts
		if multiplier > 4 {
			multiplier = 4
		}
		lockDuration := time.Duration(1<<uint(multiplier)) * time.Hour
		if lockDuration > l.config.MaxLockoutDuration {
			lockDuration = l.config.MaxLockoutDuration
		}
		lockedUntil := now.Add(lockDuration)
		info.lockedUntil = &lockedUntil
	}

	return nil
}

// RecordSuccessfulAttempt clears failed attempts for an email.
func (l *Lockout) RecordSuccessfulAttempt(email string) {
	if !l.config.Enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, email)
}

// IsLocked checks if an account is currently locked.
func (l *Lockout) IsLocked(email string) bool {
	if !l.config.Enabled {
		return false
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	info, exists := l.attempts[email]
	if !exists || info.lockedUntil == nil {
		return false
	}

	return time.Now().Before(*info.lockedUntil)
}

// GetLockoutInfo returns lockout status.
func (l *Lockout) GetLockoutInfo(email string) (locked bool, attemptsRemaining int, retryAfter time.Duration) {
	if !l.config.Enabled {
		return false, l.config.MaxAttempts, 0
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	info, exists := l.attempts[email]
	if !exists {
		return false, l.config.MaxAttempts, 0
	}

	if info.lockedUntil != nil && time.Now().Before(*info.lockedUntil) {
		return true, 0, time.Until(*info.lockedUntil)
	}

	remaining := l.config.MaxAttempts - info.count
	if remaining < 0 {
		remaining = 0
	}

	return false, remaining, 0
}

// LockoutMiddleware returns a middleware that checks lockout status.
func LockoutMiddleware(lockout *Lockout) func(c *gin.Context) {
	return func(c *gin.Context) {
		if !lockout.config.Enabled {
			c.Next()
			return
		}

		email := c.PostForm("email")
		if email == "" {
			email = c.Query("email")
		}

		if email != "" && lockout.IsLocked(email) {
			_, _, retryAfter := lockout.GetLockoutInfo(email)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "account_locked",
				"message":     "Too many failed attempts, please try again later",
				"retry_after": retryAfter.Seconds(),
			})
			return
		}

		c.Next()
	}
}

// LoadLockoutConfig loads lockout configuration from environment variables.
func LoadLockoutConfig() LockoutConfig {
	enabled := os.Getenv("ACCOUNT_LOCKOUT_ENABLED") == "true"
	maxAttempts := 5
	if v := os.Getenv("ACCOUNT_LOCKOUT_ATTEMPTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxAttempts = n
		}
	}

	duration := time.Hour
	if v := os.Getenv("ACCOUNT_LOCKOUT_DURATION"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			duration = time.Duration(n) * time.Second
		}
	}

	return LockoutConfig{
		Enabled:           enabled,
		MaxAttempts:       maxAttempts,
		LockoutDuration:   duration,
		MaxLockoutDuration: 24 * time.Hour,
	}
}