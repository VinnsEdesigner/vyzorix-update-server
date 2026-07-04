package hub

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// RateLimiterConfig holds configuration for the rate limiter.
type RateLimiterConfig struct {
	// Rate is the number of tokens added per second (default 100)
	Rate float64
	// Burst is the maximum bucket size (default 200)
	Burst int
	// CleanupInterval is how often to clean up idle entries
	CleanupInterval time.Duration
}

// DefaultRateLimiterConfig returns the default rate limiter configuration.
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		Rate:            100,
		Burst:           200,
		CleanupInterval: 5 * time.Minute,
	}
}

// RateLimiterMetrics holds rate limiter metrics.
type RateLimiterMetrics struct {
	TotalRequests int64 `json:"totalRequests"`
	TotalAllowed  int64 `json:"totalAllowed"`
	TotalLimited  int64 `json:"totalLimited"`
	ActiveClients int   `json:"activeClients"`
}

// tokenBucket implements the token bucket algorithm.
type tokenBucket struct {
	lastUpdate time.Time
	tokens     float64
	rate       float64
	burst      int
}

// RateLimiter implements per-client WebSocket rate limiting using token bucket algorithm.
type RateLimiter struct {
	log       *slog.Logger
	config    *RateLimiterConfig
	buckets   map[string]*tokenBucket
	mu        sync.RWMutex
	metrics   RateLimiterMetrics
	metricsMu sync.RWMutex
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(log *slog.Logger, cfg *RateLimiterConfig) *RateLimiter {
	if cfg == nil {
		cfg = DefaultRateLimiterConfig()
	}

	rl := &RateLimiter{
		log:     log,
		config:  cfg,
		buckets: make(map[string]*tokenBucket),
	}

	return rl
}

// Allow checks if a request from the given client ID is allowed.
// Returns true if allowed, false if rate limited.
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[clientID]
	if !exists {
		bucket = &tokenBucket{
			tokens:     float64(rl.config.Burst),
			lastUpdate: time.Now(),
			rate:       rl.config.Rate,
			burst:      rl.config.Burst,
		}
		rl.buckets[clientID] = bucket
	}

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	bucket.lastUpdate = now

	// Add tokens based on rate
	bucket.tokens += elapsed * bucket.rate
	if bucket.tokens > float64(bucket.burst) {
		bucket.tokens = float64(bucket.burst)
	}

	// Check if we have tokens
	if bucket.tokens >= 1 {
		bucket.tokens--

		rl.incrementAllowed()

		return true
	}

	rl.incrementLimited()
	rl.log.Debug("rate limit exceeded", "clientId", clientID, "tokens", bucket.tokens)

	return false
}

// AllowN checks if N requests from the given client ID are allowed.
// Returns true if all N requests are allowed, false if rate limited.
func (rl *RateLimiter) AllowN(clientID string, n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[clientID]
	if !exists {
		bucket = &tokenBucket{
			tokens:     float64(rl.config.Burst),
			lastUpdate: time.Now(),
			rate:       rl.config.Rate,
			burst:      rl.config.Burst,
		}
		rl.buckets[clientID] = bucket
	}

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	bucket.lastUpdate = now

	// Add tokens based on rate
	bucket.tokens += elapsed * bucket.rate
	if bucket.tokens > float64(bucket.burst) {
		bucket.tokens = float64(bucket.burst)
	}

	// Check if we have enough tokens
	if bucket.tokens >= float64(n) {
		bucket.tokens -= float64(n)

		rl.incrementAllowed()

		return true
	}

	rl.incrementLimited()

	return false
}

func (rl *RateLimiter) incrementAllowed() {
	rl.metricsMu.Lock()
	rl.metrics.TotalRequests++
	rl.metrics.TotalAllowed++
	rl.metricsMu.Unlock()
}

func (rl *RateLimiter) incrementLimited() {
	rl.metricsMu.Lock()
	rl.metrics.TotalRequests++
	rl.metrics.TotalLimited++
	rl.metricsMu.Unlock()
}

// Reset resets the rate limit for a client.
func (rl *RateLimiter) Reset(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.buckets, clientID)
}

// GetMetrics returns current rate limiter metrics.
func (rl *RateLimiter) GetMetrics() RateLimiterMetrics {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	rl.metricsMu.Lock()
	defer rl.metricsMu.Unlock()

	return RateLimiterMetrics{
		TotalRequests: rl.metrics.TotalRequests,
		TotalAllowed:  rl.metrics.TotalAllowed,
		TotalLimited:  rl.metrics.TotalLimited,
		ActiveClients: len(rl.buckets),
	}
}

// StartCleanup starts the background cleanup of idle entries.
func (rl *RateLimiter) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(rl.config.CleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rl.cleanupIdle()
			}
		}
	}()
}

// cleanupIdle removes idle entries that have exhausted their tokens.
func (rl *RateLimiter) cleanupIdle() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	idleThreshold := 10 * time.Minute
	removed := 0

	for clientID, bucket := range rl.buckets {
		// Remove entries that haven't been used in a while and have no tokens
		if bucket.tokens < 1 && now.Sub(bucket.lastUpdate) > idleThreshold {
			delete(rl.buckets, clientID)

			removed++
		}
	}

	if removed > 0 {
		rl.log.Info("cleaned up idle rate limiter entries", "removed", removed)
	}
}

// ClientCount returns the number of tracked clients.
func (rl *RateLimiter) ClientCount() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return len(rl.buckets)
}

// TokenLevel returns the current token level for a client (for debugging/monitoring).
func (rl *RateLimiter) TokenLevel(clientID string) float64 {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	bucket, exists := rl.buckets[clientID]
	if !exists {
		return float64(rl.config.Burst)
	}

	// Calculate current tokens including refill
	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate).Seconds()

	currentTokens := bucket.tokens + elapsed*bucket.rate
	if currentTokens > float64(bucket.burst) {
		currentTokens = float64(bucket.burst)
	}

	return currentTokens
}
