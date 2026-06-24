// Package ratelimit provides HTTP request rate limiting.
package ratelimit

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Config holds configuration for a rate limiter.
type Config struct {
	KeyFunc   func(*gin.Context) string
	OnLimit   func(*gin.Context)
	Window    time.Duration
	MaxReq    int
	SkipOnErr bool
}

// DefaultKeyFunc returns the client IP address as the rate limit key.
func DefaultKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// Limiter manages request rate limiting.
type Limiter struct {
	bucket map[string]*tokenBucket
	ttl    time.Duration
	max    int
	mu     sync.RWMutex
}

type tokenBucket struct {
	lastReset time.Time
	tokens    int
}

// New creates a new rate limiter with the given window and max requests.
func New(window time.Duration, maxRequests int) *Limiter {
	rl := &Limiter{
		bucket: make(map[string]*tokenBucket),
		ttl:    window,
		max:    maxRequests,
	}
	go rl.cleanup()

	return rl
}

// Allow checks if a request should be allowed for the given key.
func (rl *Limiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.bucket[key]

	if !exists || now.Sub(b.lastReset) >= rl.ttl {
		rl.bucket[key] = &tokenBucket{
			tokens:    1,
			lastReset: now,
		}

		return true
	}

	if b.tokens >= rl.max {
		return false
	}

	b.tokens++

	return true
}

// GetRemaining returns the number of remaining requests for a key.
func (rl *Limiter) GetRemaining(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	b, exists := rl.bucket[key]
	if !exists {
		return rl.max
	}

	if time.Since(b.lastReset) >= rl.ttl {
		return rl.max
	}

	remaining := rl.max - b.tokens
	if remaining < 0 {
		return 0
	}

	return remaining
}

// Reset clears the rate limit for a specific key.
func (rl *Limiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.bucket, key)
}

func (rl *Limiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()

		now := time.Now()
		for key, b := range rl.bucket {
			if now.Sub(b.lastReset) >= rl.ttl {
				delete(rl.bucket, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns a Gin middleware that applies rate limiting.
func (rl *Limiter) Middleware(cfg Config) gin.HandlerFunc {
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = DefaultKeyFunc
	}

	return func(c *gin.Context) {
		key := cfg.KeyFunc(c)

		if !rl.Allow(key) {
			remaining := rl.GetRemaining(key)
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.max))
			c.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(remaining), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(rl.ttl.Seconds()), 10))

			if cfg.OnLimit != nil {
				cfg.OnLimit(c)
				return
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})

			return
		}

		remaining := rl.GetRemaining(key)
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.max))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(remaining), 10))

		c.Next()
	}
}

// MultiWindowLimiter supports multiple rate limit windows.
type MultiWindowLimiter struct {
	limiters map[string]*Limiter
	config   map[string]struct {
		Window time.Duration
		Max    int
	}
}

// NewMultiWindow creates a limiter with multiple windows.
func NewMultiWindow(limits map[string]struct {
	Window time.Duration
	Max    int
}) *MultiWindowLimiter {
	ml := &MultiWindowLimiter{
		limiters: make(map[string]*Limiter),
		config:   limits,
	}

	for name, cfg := range limits {
		ml.limiters[name] = New(cfg.Window, cfg.Max)
	}

	return ml
}

// Middleware creates a middleware that applies all configured limits.
func (ml *MultiWindowLimiter) Middleware(keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)

		for name, limiter := range ml.limiters {
			if !limiter.Allow(key) {
				cfg := ml.config[name]
				c.Header("Retry-After", strconv.FormatInt(int64(cfg.Window.Seconds()), 10))
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error":   "rate_limit_exceeded",
					"message": "Too many requests. Please try again later.",
				})

				return
			}
		}

		c.Next()
	}
}

// AuthLimiter is a pre-configured rate limiter for auth endpoints.
var AuthLimiter = NewMultiWindow(map[string]struct {
	Window time.Duration
	Max    int
}{
	"minute": {Window: time.Minute, Max: 5},
	"hour":   {Window: time.Hour, Max: 20},
})
