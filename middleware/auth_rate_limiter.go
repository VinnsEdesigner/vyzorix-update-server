package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthRateLimiter provides stricter rate limiting specifically for auth endpoints.
// This prevents brute force attacks on login, register, and password reset flows.
type AuthRateLimiter struct {
	LoginLimiter        *RateLimiter // 5 attempts/minute per IP
	RegisterLimiter     *RateLimiter // 3 attempts/minute per IP
	PasswordResetLimiter *RateLimiter // 2 attempts/minute per IP
}

// AuthRateLimitConfig holds the configuration for auth rate limits
type AuthRateLimitConfig struct {
	LoginLimit        int           // Max login attempts per minute
	RegisterLimit     int           // Max registration attempts per minute
	PasswordResetLimit int          // Max password reset requests per minute
	RefillDuration    time.Duration // Token refill interval
}

// DefaultAuthRateLimits provides sensible defaults for auth rate limiting
var DefaultAuthRateLimits = AuthRateLimitConfig{
	LoginLimit:         10, // 10 login attempts per minute
	RegisterLimit:      5,  // 5 registration attempts per minute
	PasswordResetLimit: 3,  // 3 password reset requests per minute
	RefillDuration:     time.Minute,
}

// StrictAuthRateLimits for high-security environments
var StrictAuthRateLimits = AuthRateLimitConfig{
	LoginLimit:         5,  // 5 login attempts per minute
	RegisterLimit:      3,  // 3 registration attempts per minute
	PasswordResetLimit: 2,  // 2 password reset requests per minute
	RefillDuration:     time.Minute,
}

// NewAuthRateLimiter creates a new auth-specific rate limiter
func NewAuthRateLimiter(config AuthRateLimitConfig) *AuthRateLimiter {
	return &AuthRateLimiter{
		LoginLimiter:        NewRateLimiter(config.LoginLimit, config.RefillDuration),
		RegisterLimiter:     NewRateLimiter(config.RegisterLimit, config.RefillDuration),
		PasswordResetLimiter: NewRateLimiter(config.PasswordResetLimit, config.RefillDuration),
	}
}

// LoginMiddleware returns middleware that rate limits login requests
func (a *AuthRateLimiter) LoginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.LoginLimiter.Allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "Too many login attempts. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RegisterMiddleware returns middleware that rate limits registration requests
func (a *AuthRateLimiter) RegisterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.RegisterLimiter.Allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "Too many registration attempts. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// PasswordResetMiddleware returns middleware that rate limits password reset requests
func (a *AuthRateLimiter) PasswordResetMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.PasswordResetLimiter.Allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "Too many password reset attempts. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthRateLimiterStatus represents the current state of auth rate limiters
type AuthRateLimiterStatus struct {
	LoginLimit        int `json:"loginLimit"`
	LoginRemaining    int `json:"loginRemaining"`
	RegisterLimit     int `json:"registerLimit"`
	RegisterRemaining int `json:"registerRemaining"`
	PasswordResetLimit int `json:"passwordResetLimit"`
	PasswordResetRemaining int `json:"passwordResetRemaining"`
}

// GetStatus returns the current status of all rate limiters for a client IP
func (a *AuthRateLimiter) GetStatus(clientIP string) AuthRateLimiterStatus {
	return AuthRateLimiterStatus{
		LoginLimit:              a.LoginLimiter.Capacity,
		LoginRemaining:         getRemaining(a.LoginLimiter, clientIP),
		RegisterLimit:          a.RegisterLimiter.Capacity,
		RegisterRemaining:      getRemaining(a.RegisterLimiter, clientIP),
		PasswordResetLimit:     a.PasswordResetLimiter.Capacity,
		PasswordResetRemaining: getRemaining(a.PasswordResetLimiter, clientIP),
	}
}

// getRemaining calculates remaining tokens for a key (approximate, not thread-safe)
func getRemaining(limiter *RateLimiter, key string) int {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	
	b := limiter.buckets[key]
	if b == nil {
		return limiter.Capacity
	}
	return b.tokens
}