// Package middleware provides HTTP middleware for the Vyzorix API.
package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
)

// ReplayProtection provides thread-safe replay attack prevention.
// It maintains an in-memory cache of recently seen signatures to detect.
// and reject replay attacks within the configured time window.
type ReplayProtection struct {
	mu     sync.RWMutex
	cache  map[string]time.Time
	window time.Duration
	max    int
}

// NewReplayProtection creates a new replay protection cache.
// The cache is initialized with the provided configuration and will.
// automatically evict entries older than the timestamp window.
func NewReplayProtection(cfg config.SigningConfig) *ReplayProtection {
	return &ReplayProtection{
		cache:  make(map[string]time.Time),
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

	// Add to cache with current timestamp.
	rp.cache[key] = time.Now()

	// Evict old entries and trim to max size if needed.
	rp.evictOldLocked()

	return false
}

// buildKey creates a cache key from client ID and signature using SHA-256 hashing.
// This ensures consistent key generation and hides the raw signature data.
func (rp *ReplayProtection) buildKey(clientID, signature string) string {
	h := sha256.New()
	h.Write([]byte(clientID + ":" + signature))
	return hex.EncodeToString(h.Sum(nil))
}

// evictOldLocked removes entries older than the configured window.
// Caller must hold the lock. This is a no-op if cache is empty.
func (rp *ReplayProtection) evictOldLocked() {
	if len(rp.cache) == 0 {
		return
	}

	now := time.Now()
	window := rp.window

	// Collect keys to delete to avoid modifying map during iteration.
	var expiredKeys []string
	for key, timestamp := range rp.cache {
		if now.Sub(timestamp) > window {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Delete all expired entries.
	for _, key := range expiredKeys {
		delete(rp.cache, key)
	}

	// If still over capacity, trim to max using proper sorting.
	if len(rp.cache) > rp.max {
		rp.trimToMaxLocked()
	}
}

// trimToMaxLocked removes the oldest entries to bring cache under max size.
// Uses stable sort by timestamp (oldest first) to ensure consistent eviction.
// Caller must hold the lock.
func (rp *ReplayProtection) trimToMaxLocked() {
	if len(rp.cache) <= rp.max {
		return
	}

	// Create a sortable slice of entries.
	type cacheEntry struct {
		key       string
		timestamp time.Time
	}

	entries := make([]cacheEntry, 0, len(rp.cache))
	for key, timestamp := range rp.cache {
		entries = append(entries, cacheEntry{key: key, timestamp: timestamp})
	}

	// Sort by timestamp ascending (oldest first) using stable sort.
	// This ensures consistent ordering when timestamps are equal.
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].timestamp.Before(entries[j].timestamp)
	})

	// Calculate how many entries to remove.
	entriesToRemove := len(entries) - rp.max

	// Remove the oldest entries (they're at the beginning after sorting).
	for i := 0; i < entriesToRemove; i++ {
		delete(rp.cache, entries[i].key)
	}
}

// Clear resets the cache, removing all entries.
func (rp *ReplayProtection) Clear() {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.cache = make(map[string]time.Time)
}

// Size returns the current number of entries in the cache.
func (rp *ReplayProtection) Size() int {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return len(rp.cache)
}

// Stats returns cache statistics for monitoring purposes.
type ReplayStats struct {
	CacheSize  int           `json:"cacheSize"`
	WindowSecs int           `json:"windowSecs"`
	MaxSize    int           `json:"maxSize"`
	UtilPct    float64       `json:"utilPct"`
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
