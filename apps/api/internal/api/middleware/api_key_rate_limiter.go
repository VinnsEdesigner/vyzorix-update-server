package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitResult represents the result of a rate limit check.
type RateLimitResult struct {
	ResetAt    time.Time
	Remaining  int
	RetryAfter int
	Allowed    bool
}

// InMemoryRateLimiter implements an in-memory rate limiter for API keys.
// This is a fallback for deployments without Redis.
type InMemoryRateLimiter struct {
	data   map[string]*windowedCounter
	limit  int
	window time.Duration
	mu     sync.RWMutex
}

type windowedCounter struct {
	windowEnd time.Time
	count     int64
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter.
func NewInMemoryRateLimiter(limit int, window time.Duration) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		data:   make(map[string]*windowedCounter),
		limit:  limit,
		window: window,
	}
}

// Allow checks if a request is allowed under the rate limit.
func (r *InMemoryRateLimiter) Allow(keyID string) *RateLimitResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	windowEnd := now.Add(r.window).Truncate(r.window).Add(r.window)

	counter, exists := r.data[keyID]
	if !exists || now.After(counter.windowEnd) {
		counter = &windowedCounter{windowEnd: windowEnd}
		r.data[keyID] = counter
	}

	counter.count++

	if counter.count > int64(r.limit) {
		retryAfter := int(counter.windowEnd.Sub(now).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetAt:    counter.windowEnd,
			RetryAfter: retryAfter,
		}
	}

	return &RateLimitResult{
		Allowed:    true,
		Remaining:  r.limit - int(counter.count),
		ResetAt:    counter.windowEnd,
		RetryAfter: 0,
	}
}

// Cleanup removes expired entries from the rate limiter.
func (r *InMemoryRateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for key, counter := range r.data {
		if now.After(counter.windowEnd) {
			delete(r.data, key)
		}
	}
}

// APIKeyRateLimitMiddleware creates middleware that rate limits requests by API key.
func APIKeyRateLimitMiddleware(limiter *InMemoryRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key ID from context (set by tenant_api_key middleware).
		keyIDVal, exists := c.Get("api_key_id")
		if !exists {
			// No API key, skip rate limiting.
			c.Next()
			return
		}

		keyID, ok := keyIDVal.(string)
		if !ok || keyID == "" {
			c.Next()
			return
		}

		result := limiter.Allow(keyID)

		// Set rate limit headers.
		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			c.Header("Retry-After", strconv.Itoa(result.RetryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "too many requests",
				"retry_after": result.RetryAfter,
			})
			return
		}

		c.Next()
	}
}
