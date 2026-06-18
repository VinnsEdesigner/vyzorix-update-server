// Package security provides authentication utilities including session revocation.
package auth

import (
	"sort"
	"sync"
	"time"
)

// RevocationReason describes why a session was revoked.
type RevocationReason string

const (
	RevokeReasonLogout         RevocationReason = "user_logout"
	RevokeReasonPasswordChange RevocationReason = "password_change"
	RevokeReasonAdmin          RevocationReason = "admin_revoke"
	RevokeReasonSecurity       RevocationReason = "security_event"
	RevokeReasonExpired        RevocationReason = "session_expired"
)

// RevocationEntry represents a revoked session entry.
type RevocationEntry struct {
	TokenHash   string
	RevokedAt   time.Time
	Reason      RevocationReason
	Description string
}

// RevocationList manages session revocation with thread-safe operations.
type RevocationList struct {
	mu         sync.RWMutex
	revoked    map[string]*RevocationEntry
	maxEntries int
	ttl        time.Duration
}

// NewRevocationList creates a new revocation list with specified limits.
func NewRevocationList(maxEntries int, ttl time.Duration) *RevocationList {
	return &RevocationList{
		revoked:    make(map[string]*RevocationEntry),
		maxEntries: maxEntries,
		ttl:        ttl,
	}
}

// Revoke adds a token hash to the revocation list.
func (r *RevocationList) Revoke(tokenHash string, reason RevocationReason, description string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.revoked[tokenHash] = &RevocationEntry{
		TokenHash:   tokenHash,
		RevokedAt:   time.Now(),
		Reason:      reason,
		Description: description,
	}

	// Evict old entries if over capacity.
	if len(r.revoked) > r.maxEntries {
		r.evictOldEntries()
	}
}

// IsRevoked checks if a token hash is in the revocation list.
func (r *RevocationList) IsRevoked(tokenHash string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.revoked[tokenHash]
	if !exists {
		return false
	}

	// Check if entry has expired (beyond TTL).
	if time.Since(entry.RevokedAt) > r.ttl {
		return false
	}

	return true
}

// RevokeMultiple revokes multiple token hashes at once.
func (r *RevocationList) RevokeMultiple(tokenHashes []string, reason RevocationReason, description string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, hash := range tokenHashes {
		r.revoked[hash] = &RevocationEntry{
			TokenHash:   hash,
			RevokedAt:   now,
			Reason:      reason,
			Description: description,
		}
	}

	// Evict old entries if over capacity.
	if len(r.revoked) > r.maxEntries {
		r.evictOldEntries()
	}
}

// Remove removes a token hash from the revocation list (if it was added erroneously).
func (r *RevocationList) Remove(tokenHash string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.revoked, tokenHash)
}

// Size returns the current number of revoked entries.
func (r *RevocationList) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.revoked)
}

// Clear removes all entries from the revocation list.
func (r *RevocationList) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revoked = make(map[string]*RevocationEntry)
}

// evictOldEntries removes expired entries and oldest if still over capacity.
// Must be called with mu held.
func (r *RevocationList) evictOldEntries() {
	now := time.Now()
	cutoff := now.Add(-r.ttl)

	// First, remove all expired entries.
	for hash, entry := range r.revoked {
		if entry.RevokedAt.Before(cutoff) {
			delete(r.revoked, hash)
		}
	}

	// If still over capacity, remove oldest entries (10%).
	if len(r.revoked) > r.maxEntries {
		evictCount := r.maxEntries / 10
		if evictCount < 1 {
			evictCount = 1
		}

		// Collect entries and sort by revocation time (oldest first).
		type entryWithTime struct {
			hash      string
			revokedAt time.Time
		}
		entries := make([]entryWithTime, 0, len(r.revoked))
		for hash, entry := range r.revoked {
			entries = append(entries, entryWithTime{hash, entry.RevokedAt})
		}

		// Sort by revokedAt ascending (oldest first).
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].revokedAt.Before(entries[j].revokedAt)
		})

		// Remove evictCount oldest entries.
		for i := 0; i < evictCount && i < len(entries); i++ {
			delete(r.revoked, entries[i].hash)
		}
	}
}

// DefaultRevocationList creates a revocation list with sensible defaults.
// Max 100,000 entries, 24-hour TTL.
func DefaultRevocationList() *RevocationList {
	return NewRevocationList(100000, 24*time.Hour)
}