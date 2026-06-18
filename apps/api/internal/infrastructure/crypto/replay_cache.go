package crypto

import (
	"sync"
	"time"
)

const (
	// DefaultMaxSize is the default maximum number of entries.
	DefaultMaxSize = 100000
	// DefaultTTL is the default time-to-live for entries.
	DefaultTTL = 10 * time.Minute
	// DefaultCleanupInterval is the default cleanup interval.
	DefaultCleanupInterval = 5 * time.Minute
)

// ReplayCache provides thread-safe replay attack protection using a token bucket approach.
// It stores signatures with timestamps and automatically cleans up expired entries.
type ReplayCache struct {
	mu              sync.RWMutex
	entries         map[string]time.Time
	maxSize         int
	ttl             time.Duration
	cleanupInterval time.Duration
	stopCh          chan struct{}
}

// ReplayCacheOption configures the ReplayCache.
type ReplayCacheOption func(*ReplayCache)

// WithMaxSize sets the maximum number of entries.
func WithMaxSize(maxSize int) ReplayCacheOption {
	return func(rc *ReplayCache) {
		rc.maxSize = maxSize
	}
}

// WithTTL sets the time-to-live for entries.
func WithTTL(ttl time.Duration) ReplayCacheOption {
	return func(rc *ReplayCache) {
		rc.ttl = ttl
	}
}

// WithCleanupInterval sets the interval between cleanup runs.
func WithCleanupInterval(interval time.Duration) ReplayCacheOption {
	return func(rc *ReplayCache) {
		rc.cleanupInterval = interval
	}
}

// NewReplayCache creates a new ReplayCache with default settings.
func NewReplayCache(opts ...ReplayCacheOption) *ReplayCache {
	rc := &ReplayCache{
		entries:         make(map[string]time.Time),
		maxSize:         DefaultMaxSize,
		ttl:             DefaultTTL,
		cleanupInterval: DefaultCleanupInterval,
		stopCh:          make(chan struct{}),
	}

	for _, opt := range opts {
		opt(rc)
	}

	// Start background cleanup goroutine.
	go rc.cleanupLoop()

	return rc
}

// cleanupLoop periodically removes expired entries.
func (rc *ReplayCache) cleanupLoop() {
	ticker := time.NewTicker(rc.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rc.stopCh:
			return
		case <-ticker.C:
			rc.cleanup()
		}
	}
}

// cleanup removes entries that have expired.
func (rc *ReplayCache) cleanup() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	cutoff := time.Now().Add(-rc.ttl)
	removed := 0

	for sig, timestamp := range rc.entries {
		if timestamp.Before(cutoff) {
			delete(rc.entries, sig)
			removed++
		}
	}

	if removed > 0 {
		// In production, this could log the cleanup stats.
		_ = removed // suppress unused variable warning
	}
}

// Use checks if a signature has been seen before and marks it as seen if not.
// Returns true if the signature is new (not a replay), false if it's a replay.
func (rc *ReplayCache) Use(signature string) bool {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Check if signature already exists.
	if _, exists := rc.entries[signature]; exists {
		return false // Replay detected
	}

	// Evict oldest entries if at capacity.
	if len(rc.entries) >= rc.maxSize {
		rc.evictOldest(rc.maxSize / 10) // Evict 10%
	}

	// Mark signature as seen.
	rc.entries[signature] = time.Now()
	return true
}

// evictOldest removes the oldest n entries from the cache.
func (rc *ReplayCache) evictOldest(n int) {
	if n <= 0 {
		return
	}

	// Find n oldest entries.
	type entry struct {
		sig      string
		timestamp time.Time
	}

	oldest := make([]entry, 0, n)
	for sig, timestamp := range rc.entries {
		oldest = append(oldest, entry{sig: sig, timestamp: timestamp})
	}

	// Sort by timestamp (oldest first).
	for i := 0; i < len(oldest)-1; i++ {
		for j := i + 1; j < len(oldest); j++ {
			if oldest[i].timestamp.After(oldest[j].timestamp) {
				oldest[i], oldest[j] = oldest[j], oldest[i]
			}
	}
	}

	// Remove oldest n.
	for i := 0; i < n && i < len(oldest); i++ {
		delete(rc.entries, oldest[i].sig)
	}
}

// Stop stops the background cleanup goroutine.
func (rc *ReplayCache) Stop() {
	close(rc.stopCh)
}

// Len returns the current number of entries in the cache.
func (rc *ReplayCache) Len() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return len(rc.entries)
}

// Stats returns current cache statistics.
func (rc *ReplayCache) Stats() ReplayCacheStats {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return ReplayCacheStats{
		Entries:        len(rc.entries),
		MaxSize:        rc.maxSize,
		TTL:            rc.ttl,
		CleanupInt:     rc.cleanupInterval,
	}
}

// ReplayCacheStats contains cache statistics.
type ReplayCacheStats struct {
	Entries     int           `json:"entries"`
	MaxSize     int           `json:"max_size"`
	TTL         time.Duration `json:"ttl"`
	CleanupInt  time.Duration `json:"cleanup_interval"`
}

// Clear removes all entries from the cache.
func (rc *ReplayCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.entries = make(map[string]time.Time)
}
