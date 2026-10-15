package auth

import (
	"fmt"
	"testing"
)

// TestLoginLockout_StaysBounded proves the lockout map cannot grow without limit
// when an unauthenticated caller sprays unique emails. The map key is
// caller-supplied, so an unbounded map would be a memory DoS.
func TestLoginLockout_StaysBounded(t *testing.T) {
	l := NewLoginLockout()

	for i := 0; i < maxLockoutEntries*3; i++ {
		l.RecordFailed(fmt.Sprintf("spray-%d@example.test", i), "127.0.0.1", "ua")
	}

	l.mu.RLock()
	size := len(l.attempts)
	l.mu.RUnlock()

	if size > maxLockoutEntries {
		t.Fatalf("lockout map grew past its cap: size=%d cap=%d", size, maxLockoutEntries)
	}
}

// TestLoginLockout_StillLocks proves the eviction logic did not break the actual
// security behaviour: repeated failures for one account still lock it.
func TestLoginLockout_StillLocks(t *testing.T) {
	l := NewLoginLockout()
	const email = "victim@example.test"

	for i := 0; i < 5; i++ {
		l.RecordFailed(email, "127.0.0.1", "ua")
	}

	if !l.IsLocked(email) {
		t.Fatal("account should be locked after 5 failed attempts")
	}
}

// TestLoginLockout_ClearUnlocks proves a successful login still clears the counter.
func TestLoginLockout_ClearUnlocks(t *testing.T) {
	l := NewLoginLockout()
	const email = "victim@example.test"

	for i := 0; i < 5; i++ {
		l.RecordFailed(email, "127.0.0.1", "ua")
	}
	l.Clear(email)

	if l.IsLocked(email) {
		t.Fatal("Clear should unlock the account")
	}

	l.mu.RLock()
	_, exists := l.attempts[email]
	l.mu.RUnlock()
	if exists {
		t.Fatal("Clear should drop the entry entirely")
	}
}
