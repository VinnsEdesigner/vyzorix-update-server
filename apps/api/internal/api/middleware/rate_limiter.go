// Package middleware provides HTTP middleware.
package middleware

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter provides token bucket rate limiting with automatic cleanup.
// Uses a background goroutine to prevent memory leaks from stale entries.
type RateLimiter struct {
	buckets     map[string]*bucket
	Capacity    int
	Refill      time.Duration
	CleanupInterval time.Duration
	TTL          time.Duration
	mu           sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *slog.Logger
}

type bucket struct {
	last   time.Time
	tokens int
}

// DefaultCleanupInterval is the default interval between cleanup runs.
const DefaultCleanupInterval = 5 * time.Minute

// DefaultBucketTTL is the default TTL for idle buckets.
const DefaultBucketTTL = 30 * time.Minute

// NewRateLimiter creates a new RateLimiter with token bucket algorithm.
// The cleanup goroutine is started automatically.
func NewRateLimiter(capacity int, refill time.Duration) *RateLimiter {
	return NewRateLimiterWithOptions(
		capacity,
		refill,
		WithCleanupInterval(DefaultCleanupInterval),
		WithBucketTTL(DefaultBucketTTL),
	)
}

// RateLimiterOption configures the RateLimiter.
type RateLimiterOption func(*RateLimiter)

// WithCleanupInterval sets the interval between cleanup runs.
func WithCleanupInterval(interval time.Duration) RateLimiterOption {
	return func(rl *RateLimiter) {
		rl.CleanupInterval = interval
	}
}

// WithBucketTTL sets the TTL for idle buckets.
func WithBucketTTL(ttl time.Duration) RateLimiterOption {
	return func(rl *RateLimiter) {
		rl.TTL = ttl
	}
}

// WithLogger sets the logger for the RateLimiter.
func WithLogger(logger *slog.Logger) RateLimiterOption {
	return func(rl *RateLimiter) {
		rl.logger = logger
	}
}

// NewRateLimiterWithOptions creates a RateLimiter with custom options.
func NewRateLimiterWithOptions(capacity int, refill time.Duration, opts ...RateLimiterOption) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	rl := &RateLimiter{
		buckets:         make(map[string]*bucket),
		Capacity:        capacity,
		Refill:          refill,
		CleanupInterval: DefaultCleanupInterval,
		TTL:              DefaultBucketTTL,
		ctx:             ctx,
		cancel:          cancel,
		logger:          slog.Default(),
	}

	for _, opt := range opts {
		opt(rl)
	}

	// Start cleanup goroutine.
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop periodically removes stale buckets to prevent memory leaks.
func (l *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			l.cleanup()
		}
	}
}

// cleanup removes buckets that have been idle longer than TTL.
func (l *RateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	var removed int

	for key, b := range l.buckets {
		// Check if bucket is idle beyond TTL.
		if now.Sub(b.last) > l.TTL {
			delete(l.buckets, key)
			removed++
		}
	}

	if removed > 0 {
		l.logger.Debug("rate limiter cleanup",
			slog.Int("removed", removed),
			slog.Int("remaining", len(l.buckets)),
		)
	}
}

// Stop stops the cleanup goroutine and releases resources.
func (l *RateLimiter) Stop() {
	l.cancel()
}

// Middleware returns a Gin middleware that rate limits requests.
func (l *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.Allow(c.ClientIP()) {
			c.JSON(429, map[string]string{"error": "rate_limited", "message": "too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Allow checks if a request from the given key should be allowed.
func (l *RateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b := l.buckets[key]
	if b == nil {
		b = &bucket{tokens: l.Capacity, last: now}
		l.buckets[key] = b
	}

	// Only refill tokens if refill duration is > 0 (avoid divide by zero).
	if l.Refill > 0 {
		if elapsed := int(now.Sub(b.last) / l.Refill); elapsed > 0 {
			b.tokens += elapsed
			if b.tokens > l.Capacity {
				b.tokens = l.Capacity
			}
			b.last = now
		}
	}

	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}

// Stats returns current rate limiter statistics.
func (l *RateLimiter) Stats() RateLimiterStats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return RateLimiterStats{
		Buckets:    len(l.buckets),
		Capacity:   l.Capacity,
		Refill:     l.Refill,
		CleanupInt: l.CleanupInterval,
		TTL:        l.TTL,
	}
}

// RateLimiterStats contains rate limiter statistics.
type RateLimiterStats struct {
	Buckets    int           `json:"buckets"`
	Capacity   int           `json:"capacity"`
	Refill     time.Duration `json:"refill"`
	CleanupInt time.Duration `json:"cleanup_interval"`
	TTL        time.Duration `json:"ttl"`
}
