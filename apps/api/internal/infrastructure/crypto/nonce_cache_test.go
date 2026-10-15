package crypto

import (
	"testing"
	"time"
)

func TestNonceCache_RejectsReplay(t *testing.T) {
	c := NewNonceCache(time.Minute)
	now := time.Now()

	if !c.Use("nonce-1", now) {
		t.Fatal("first use of a nonce must be accepted")
	}
	if c.Use("nonce-1", now) {
		t.Fatal("second use of the same nonce must be rejected as a replay")
	}
}

func TestNonceCache_DistinctNoncesAccepted(t *testing.T) {
	c := NewNonceCache(time.Minute)
	now := time.Now()

	for i := 0; i < 100; i++ {
		if !c.Use(string(rune('a'+i%26))+string(rune('0'+i%10))+string(rune(i)), now) {
			t.Fatalf("distinct nonce %d should be accepted", i)
		}
	}
}

// TestNonceCache_StaysBounded proves the cache cannot grow without limit when
// flooded with distinct nonces - the unbounded-growth DoS the previous
// implementation allowed.
func TestNonceCache_StaysBounded(t *testing.T) {
	c := NewNonceCache(time.Minute)
	c.maxSize = 1000
	now := time.Now()

	// Insert far more distinct nonces than the cap, all within the window so none
	// are expired and the size cap (not the TTL) is what bounds the cache.
	for i := 0; i < 20000; i++ {
		c.Use("nonce-"+string(rune(i%256))+string(rune(i/256)), now)
	}

	if got := c.Len(); got > c.maxSize {
		t.Fatalf("cache grew past its bound: len=%d maxSize=%d", got, c.maxSize)
	}
}

// TestNonceCache_ExpiredEntriesEvicted proves entries older than the window are
// swept, so a nonce is only remembered for the window duration.
func TestNonceCache_ExpiredEntriesEvicted(t *testing.T) {
	c := NewNonceCache(50 * time.Millisecond)
	c.maxSize = 100000
	base := time.Now()

	for i := 0; i < 50; i++ {
		c.Use("old-"+string(rune(i)), base)
	}
	before := c.Len()
	if before == 0 {
		t.Fatal("expected entries to be recorded")
	}

	// Advance well past the window and force a sweep by exceeding the cleanup
	// interval.
	later := base.Add(2 * time.Minute)
	c.Use("fresh", later)

	if got := c.Len(); got >= before {
		t.Fatalf("expired entries should have been evicted: before=%d after=%d", before, got)
	}
}

func TestNonceCache_Clear(t *testing.T) {
	c := NewNonceCache(time.Minute)
	now := time.Now()
	c.Use("a", now)
	c.Use("b", now)

	c.Clear()

	if c.Len() != 0 {
		t.Fatalf("Clear should empty the cache, got len=%d", c.Len())
	}
}
