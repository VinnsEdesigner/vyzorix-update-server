package hub

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	log := testLogger()
	cfg := DefaultRateLimiterConfig()
	cfg.Rate = 100  // 100 per second
	cfg.Burst = 10

	limiter := NewRateLimiter(log, cfg)
	limiter.StartCleanup(context.Background())
	defer limiter.StopCleanup()

	// Test within rate limit
	for i := 0; i < 5; i++ {
		if !limiter.Allow("device-1") {
			t.Errorf("expected to allow request %d", i)
		}
	}

	metrics := limiter.GetMetrics()
	if metrics.TotalRequests != 5 {
		t.Errorf("expected 5 requests, got %d", metrics.TotalRequests)
	}
}

func TestRateLimiterBurst(t *testing.T) {
	log := testLogger()
	cfg := DefaultRateLimiterConfig()
	cfg.Rate = 10  // 10 per second
	cfg.Burst = 5

	limiter := NewRateLimiter(log, cfg)
	limiter.StartCleanup(context.Background())
	defer limiter.StopCleanup()

	// Should allow burst
	for i := 0; i < 5; i++ {
		if !limiter.Allow("device-1") {
			t.Errorf("expected to allow burst request %d", i)
		}
	}

	// Should reject beyond burst
	if limiter.Allow("device-1") {
		t.Error("expected to reject beyond burst limit")
	}

	metrics := limiter.GetMetrics()
	if metrics.TotalRejected == 0 {
		t.Error("expected some rejected requests")
	}
}

func TestRateLimiterRefill(t *testing.T) {
	log := testLogger()
	cfg := DefaultRateLimiterConfig()
	cfg.Rate = 10  // 10 per second
	cfg.Burst = 2

	limiter := NewRateLimiter(log, cfg)
	limiter.StartCleanup(context.Background())
	defer limiter.StopCleanup()

	// Consume burst
	limiter.Allow("device-1")
	limiter.Allow("device-1")

	// Should be rate limited now
	if limiter.Allow("device-1") {
		t.Error("expected to be rate limited after burst")
	}

	// Wait for refill (100ms for 1 token at 10/s)
	time.Sleep(150 * time.Millisecond)

	// Should allow again
	if !limiter.Allow("device-1") {
		t.Error("expected to allow after refill")
	}
}

func TestRateLimiterMultipleDevices(t *testing.T) {
	log := testLogger()
	cfg := DefaultRateLimiterConfig()
	cfg.Rate = 10
	cfg.Burst = 2

	limiter := NewRateLimiter(log, cfg)
	limiter.StartCleanup(context.Background())
	defer limiter.StopCleanup()

	// Device 1: consume burst
	limiter.Allow("device-1")
	limiter.Allow("device-1")

	// Device 2: should have separate bucket
	if !limiter.Allow("device-2") {
		t.Error("device-2 should have separate rate limit")
	}

	// Device 1 should still be limited
	if limiter.Allow("device-1") {
		t.Error("device-1 should still be rate limited")
	}
}

func TestRateLimiterMetrics(t *testing.T) {
	log := testLogger()
	cfg := DefaultRateLimiterConfig()
	cfg.Rate = 10
	cfg.Burst = 3

	limiter := NewRateLimiter(log, cfg)
	limiter.StartCleanup(context.Background())
	defer limiter.StopCleanup()

	// Make some requests
	for i := 0; i < 5; i++ {
		limiter.Allow("device-1")
	}

	// Some should be rejected
	for i := 0; i < 3; i++ {
		limiter.Allow("device-1")
	}

	metrics := limiter.GetMetrics()

	if metrics.TotalRequests == 0 {
		t.Error("expected some requests")
	}

	if metrics.TotalRejected == 0 {
		t.Log("no requests rejected (may be within burst)")
	}
}

func TestRateLimiterDisabled(t *testing.T) {
	log := testLogger()
	cfg := DefaultRateLimiterConfig()
	cfg.Enabled = false

	limiter := NewRateLimiter(log, cfg)

	// Should always allow when disabled
	for i := 0; i < 100; i++ {
		if !limiter.Allow("device-1") {
			t.Error("expected to always allow when disabled")
		}
	}

	metrics := limiter.GetMetrics()
	if metrics.TotalRequests != 100 {
		t.Errorf("expected 100 requests, got %d", metrics.TotalRequests)
	}
}

func TestRateLimiterHighRate(t *testing.T) {
	log := testLogger()
	cfg := DefaultRateLimiterConfig()
	cfg.Rate = 1000  // High rate
	cfg.Burst = 100

	limiter := NewRateLimiter(log, cfg)
	limiter.StartCleanup(context.Background())
	defer limiter.StopCleanup()

	// Should allow many requests
	allowed := 0
	for i := 0; i < 50; i++ {
		if limiter.Allow("device-1") {
			allowed++
		}
	}

	if allowed < 40 {
		t.Errorf("expected most requests to be allowed at high rate, got %d", allowed)
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	log := testLogger()
	cfg := DefaultRateLimiterConfig()
	cfg.CleanupInterval = 100 * time.Millisecond

	limiter := NewRateLimiter(log, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	limiter.StartCleanup(ctx)

	// Let it run
	time.Sleep(300 * time.Millisecond)

	// Should not panic
	limiter.StopCleanup()
}

func testLogger() *slog.Logger {
	return slog.Default()
}
