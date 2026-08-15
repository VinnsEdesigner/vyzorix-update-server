// Package middleware provides HTTP middleware for the Vyzorix API.
package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
)

// ReplayProtection provides thread-safe replay attack prevention.
// It maintains an in-memory cache of recently seen signatures to detect.
// and reject replay attacks within the configured time window.
// Uses O(1) eviction by maintaining insertion order in a slice.
type ReplayProtection struct {
	cache  map[string]time.Time
	order  []string
	window time.Duration
	max    int
	mu     sync.RWMutex
}

// NewReplayProtection creates a new replay protection cache.
// The cache is initialized with the provided configuration and will.
// automatically evict entries older than the timestamp window.
func NewReplayProtection(cfg config.SigningConfig) *ReplayProtection {
	return &ReplayProtection{
		cache:  make(map[string]time.Time),
		order:  make([]string, 0, cfg.MaxCacheSize),
		window: time.Duration(cfg.TimestampWindow) * time.Second,
		max:    cfg.MaxCacheSize,
	}
}

// IsReplay checks if the signature has been used recently.
// Returns true if this is a replay attack (signature was already used).
// If the signature is new, it is added to the cache and false is returned.
func (rp *ReplayProtection) IsReplay(clientID, signature string) bool {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	key := rp.buildKey(clientID, signature)
	if _, exists := rp.cache[key]; exists {
		return true
	}

	now := time.Now()

	// Add to cache and order tracking.
	rp.cache[key] = now
	rp.order = append(rp.order, key)

	// O(1) eviction if over capacity - remove oldest entry.
	if len(rp.cache) > rp.max {
		oldestKey := rp.order[0]
		delete(rp.cache, oldestKey)
		rp.order = rp.order[1:]
	}

	// Evict expired entries (front of order slice may contain expired entries).
	rp.evictExpiredLocked(now)

	return false
}

// buildKey creates a cache key from client ID and signature using SHA-256 hashing.
// This ensures consistent key generation and hides the raw signature data.
func (rp *ReplayProtection) buildKey(clientID, signature string) string {
	h := sha256.New()
	h.Write([]byte(clientID + ":" + signature))

	return hex.EncodeToString(h.Sum(nil))
}

// evictExpiredLocked removes entries older than the configured window.
// Uses the ordered slice for efficient O(N) eviction of expired entries from front.
func (rp *ReplayProtection) evictExpiredLocked(now time.Time) {
	if len(rp.order) == 0 {
		return
	}

	window := rp.window
	cutoff := now.Add(-window)

	// Find the first non-expired entry index.
	firstValidIdx := 0
	for i, key := range rp.order {
		if ts, exists := rp.cache[key]; exists && ts.After(cutoff) {
			firstValidIdx = i
			break
		}
		firstValidIdx = i + 1
	}

	// Remove all expired entries from front.
	if firstValidIdx > 0 {
		rp.order = rp.order[firstValidIdx:]
	}
}

// Clear resets the cache, removing all entries.
func (rp *ReplayProtection) Clear() {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.cache = make(map[string]time.Time)
	rp.order = make([]string, 0, rp.max)
}

// Size returns the current number of entries in the cache.
func (rp *ReplayProtection) Size() int {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	return len(rp.cache)
}

// ReplayStats holds replay protection cache statistics.
type ReplayStats struct {
	CacheSize  int     `json:"cacheSize"`
	WindowSecs int     `json:"windowSecs"`
	MaxSize    int     `json:"maxSize"`
	UtilPct    float64 `json:"utilPct"`
}

// Stats returns current replay protection statistics.
func (rp *ReplayProtection) Stats() ReplayStats {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	size := len(rp.cache)
	maxSize := rp.max

	var utilPct float64

	if maxSize > 0 {
		utilPct = float64(size) / float64(maxSize) * 100
	}

	return ReplayStats{
		CacheSize:  size,
		WindowSecs: int(rp.window.Seconds()),
		MaxSize:    maxSize,
		UtilPct:    utilPct,
	}
}
