// Package revocation provides session revocation management.
package revocation

import (
	"sort"
	"sync"
	"time"
)

// Reason describes why a session was revoked.
type Reason string

const (
	ReasonLogout         Reason = "user_logout"
	ReasonPasswordChange Reason = "password_change"
	ReasonAdmin          Reason = "admin_revoke"
	ReasonSecurity       Reason = "security_event"
	ReasonExpired        Reason = "session_expired"
)

// Entry represents a revoked session entry.
type Entry struct {
	TokenHash   string
	RevokedAt   time.Time
	Reason      Reason
	Description string
}

// List manages session revocation with thread-safe operations.
type List struct {
	mu         sync.RWMutex
	revoked    map[string]*Entry
	maxEntries int
	ttl        time.Duration
}

// New creates a new revocation list with specified limits.
func New(maxEntries int, ttl time.Duration) *List {
	return &List{
		revoked:    make(map[string]*Entry),
		maxEntries: maxEntries,
		ttl:        ttl,
	}
}

// Revoke adds a token hash to the revocation list.
func (r *List) Revoke(tokenHash string, reason Reason, description string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.revoked[tokenHash] = &Entry{
		TokenHash:   tokenHash,
		RevokedAt:   time.Now(),
		Reason:      reason,
		Description: description,
	}

	if len(r.revoked) > r.maxEntries {
		r.evictOldEntries()
	}
}

// IsRevoked checks if a token hash is in the revocation list.
func (r *List) IsRevoked(tokenHash string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.revoked[tokenHash]
	if !exists {
		return false
	}

	if time.Since(entry.RevokedAt) > r.ttl {
		return false
	}

	return true
}

// RevokeMultiple revokes multiple token hashes at once.
func (r *List) RevokeMultiple(tokenHashes []string, reason Reason, description string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, hash := range tokenHashes {
		r.revoked[hash] = &Entry{
			TokenHash:   hash,
			RevokedAt:   now,
			Reason:      reason,
			Description: description,
		}
	}

	if len(r.revoked) > r.maxEntries {
		r.evictOldEntries()
	}
}

// Remove removes a token hash from the revocation list.
func (r *List) Remove(tokenHash string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.revoked, tokenHash)
}

// Size returns the current number of revoked entries.
func (r *List) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.revoked)
}

// Clear removes all entries from the revocation list.
func (r *List) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revoked = make(map[string]*Entry)
}

func (r *List) evictOldEntries() {
	now := time.Now()
	cutoff := now.Add(-r.ttl)

	for hash, entry := range r.revoked {
		if entry.RevokedAt.Before(cutoff) {
			delete(r.revoked, hash)
		}
	}

	if len(r.revoked) > r.maxEntries {
		evictCount := r.maxEntries / 10
		if evictCount < 1 {
			evictCount = 1
		}

		type entryWithTime struct {
			hash      string
			revokedAt time.Time
		}
		entries := make([]entryWithTime, 0, len(r.revoked))
		for hash, entry := range r.revoked {
			entries = append(entries, entryWithTime{hash, entry.RevokedAt})
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].revokedAt.Before(entries[j].revokedAt)
		})

		for i := 0; i < evictCount && i < len(entries); i++ {
			delete(r.revoked, entries[i].hash)
		}
	}
}

// Default creates a revocation list with sensible defaults (100k entries, 24h TTL).
func Default() *List {
	return New(100000, 24*time.Hour)
}
