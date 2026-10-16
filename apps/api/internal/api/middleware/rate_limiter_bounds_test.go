package middleware

import (
	"testing"
	"time"
)

// TestRateLimiter_BoundedMap proves the token-bucket map cannot grow without limit
// when flooded with distinct keys.
func TestRateLimiter_BoundedMap(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	rl.MaxBuckets = 100
	defer rl.Stop()

	for i := 0; i < 5000; i++ {
		rl.Allow("spoofed-ip-" + itoaTest(i))
	}

	if got := len(rl.buckets); got > rl.MaxBuckets {
		t.Fatalf("bucket map grew past its cap: %d > %d", got, rl.MaxBuckets)
	}
}

// TestRateLimiter_StillLimitsAfterEviction proves the cap does not disable limiting
// for a normal client: the same key is still rate limited.
func TestRateLimiter_StillLimitsAfterEviction(t *testing.T) {
	rl := NewRateLimiter(3, time.Hour) // effectively no refill within the test
	rl.MaxBuckets = 10
	defer rl.Stop()

	allowed := 0

	for i := 0; i < 10; i++ {
		if rl.Allow("client-a") {
			allowed++
		}
	}
	if allowed != 3 {
		t.Fatalf("expected 3 allowed requests, got %d", allowed)
	}
}

// TestInMemoryRateLimiter_NoBoundaryBurst is the regression guard for the fixed
// window: with a sliding window the quota cannot be doubled across a boundary.
func TestInMemoryRateLimiter_NoBoundaryBurst(t *testing.T) {
	// Short window so the test is fast; limit 3.
	lim := NewInMemoryRateLimiter(3, 200*time.Millisecond)

	// Exhaust the quota.
	for i := 0; i < 3; i++ {
		if !lim.Allow("k").Allowed {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	if lim.Allow("k").Allowed {
		t.Fatal("4th request in the window must be denied")
	}

	// A fixed window anchored to a calendar boundary would allow the full quota
	// again immediately after the boundary, doubling the rate. A sliding window
	// must keep denying until the window from the FIRST request has elapsed.
	if lim.Allow("k").Allowed {
		t.Fatal("still inside the sliding window; must remain denied")
	}

	// After the window passes, requests are allowed again.
	time.Sleep(220 * time.Millisecond)
	if !lim.Allow("k").Allowed {
		t.Fatal("a new window should allow requests again")
	}
}

func itoaTest(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}
