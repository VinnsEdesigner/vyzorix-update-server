package security

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig holds configuration for a rate limiter.
type RateLimitConfig struct {
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

// RateLimiter manages request rate limiting.
type RateLimiter struct {
	bucket map[string]*tokenBucket
	ttl    time.Duration
	max    int
	mu     sync.RWMutex
}

// tokenBucket represents a rate limit bucket for a specific key.
type tokenBucket struct {
	lastReset time.Time
	tokens    int
}

// NewRateLimiter creates a new rate limiter with the given window and max requests.
func NewRateLimiter(window time.Duration, max int) *RateLimiter {
	rl := &RateLimiter{
		bucket: make(map[string]*tokenBucket),
		ttl:    window,
		max:    max,
	}
	// Start cleanup goroutine
	go rl.cleanup()
	return rl
}

// Allow checks if a request should be allowed for the given key.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.bucket[key]

	if !exists || now.Sub(b.lastReset) >= rl.ttl {
		// New or expired bucket
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
func (rl *RateLimiter) GetRemaining(key string) int {
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
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.bucket, key)
}

// cleanup periodically removes expired buckets.
func (rl *RateLimiter) cleanup() {
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
func (rl *RateLimiter) Middleware(cfg RateLimitConfig) gin.HandlerFunc {
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

		// Set rate limit headers for successful requests
		remaining := rl.GetRemaining(key)
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.max))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(remaining), 10))

		c.Next()
	}
}

// MultiWindowLimiter supports multiple rate limit windows (e.g., per-minute and per-hour).
type MultiWindowLimiter struct {
	limiters map[string]*RateLimiter
	_        struct{} // silence unused field lint
	config   map[string]struct {
		Window time.Duration
		Max    int
	}
}

// NewMultiWindowLimiter creates a limiter with multiple windows.
func NewMultiWindowLimiter(limits map[string]struct {
	Window time.Duration
	Max    int
}) *MultiWindowLimiter {
	ml := &MultiWindowLimiter{
		limiters: make(map[string]*RateLimiter),
		config:   limits,
	}

	for name, cfg := range limits {
		ml.limiters[name] = NewRateLimiter(cfg.Window, cfg.Max)
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

// AuthRateLimiter is a pre-configured rate limiter for auth endpoints.
// Limits: 5 requests per minute, 20 requests per hour per IP.
var AuthRateLimiter = NewMultiWindowLimiter(map[string]struct {
	Window time.Duration
	Max    int
}{
	"minute": {Window: time.Minute, Max: 5},
	"hour":   {Window: time.Hour, Max: 20},
})
