package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(5, time.Minute)

	// First 5 requests should pass
	for i := 0; i < 5; i++ {
		if !limiter.Allow("192.168.1.1") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 6th request should be denied
	if limiter.Allow("192.168.1.1") {
		t.Error("6th request should be denied")
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)

	// Different IPs should have separate buckets
	if !limiter.Allow("192.168.1.1") {
		t.Error("IP1 request 1 should be allowed")
	}
	if !limiter.Allow("192.168.1.2") {
		t.Error("IP2 request 1 should be allowed")
	}
	if !limiter.Allow("192.168.1.1") {
		t.Error("IP1 request 2 should be allowed")
	}
	// IP1 should now be rate limited
	if limiter.Allow("192.168.1.1") {
		t.Error("IP1 request 3 should be denied")
	}
	// IP2 should still work
	if !limiter.Allow("192.168.1.2") {
		t.Error("IP2 request 2 should be allowed")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	limiter := NewRateLimiter(2, 100*time.Millisecond)

	// Exhaust tokens
	limiter.Allow("192.168.1.1")
	limiter.Allow("192.168.1.1")

	// Should be denied
	if limiter.Allow("192.168.1.1") {
		t.Error("should be rate limited")
	}

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	if !limiter.Allow("192.168.1.1") {
		t.Error("should be allowed after refill")
	}
}

func TestRateLimiter_MaxCapacity(t *testing.T) {
	limiter := NewRateLimiter(5, time.Minute)

	// Fill up
	for i := 0; i < 10; i++ {
		limiter.Allow("192.168.1.1")
	}

	// Wait for some refill
	time.Sleep(200 * time.Millisecond)

	// Should not exceed capacity
	allowed := 0
	for i := 0; i < 100; i++ {
		if limiter.Allow("192.168.1.1") {
			allowed++
		}
	}

	if allowed > 5 {
		t.Errorf("allowed %d requests, max capacity is 5", allowed)
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)

	handler := limiter.Middleware()

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.RemoteAddr = "192.168.1.1:12345"

		handler(c)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}

	// Third request should be rate limited
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "192.168.1.1:12345"

	handler(c)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("rate limited request: got status %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiter_Middleware_DifferentIPs(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	handler := limiter.Middleware()

	// IP1
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = httptest.NewRequest("GET", "/test", nil)
	c1.Request.RemoteAddr = "192.168.1.1:12345"
	handler(c1)

	// IP2 (different, should be allowed)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/test", nil)
	c2.Request.RemoteAddr = "192.168.1.2:12345"
	handler(c2)

	if w1.Code != http.StatusOK {
		t.Errorf("IP1 first request: got %d, want %d", w1.Code, http.StatusOK)
	}
	if w2.Code != http.StatusOK {
		t.Errorf("IP2 first request: got %d, want %d", w2.Code, http.StatusOK)
	}

	// IP1 again should be blocked
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest("GET", "/test", nil)
	c3.Request.RemoteAddr = "192.168.1.1:12346"
	handler(c3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 second request: got %d, want %d", w3.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiter_ResponseBody(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	handler := limiter.Middleware()

	// Exhaust
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = httptest.NewRequest("GET", "/test", nil)
	c1.Request.RemoteAddr = "192.168.1.1:12345"
	handler(c1)

	// Should be rate limited
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("GET", "/test", nil)
	c2.Request.RemoteAddr = "192.168.1.1:12346"
	handler(c2)

	body := w2.Body.String()
	if body == "" {
		t.Error("expected error message in response body")
	}
}

func TestRateLimiter_Capacity(t *testing.T) {
	limiter := NewRateLimiter(100, time.Minute)

	if limiter.Capacity != 100 {
		t.Errorf("Capacity = %d, want 100", limiter.Capacity)
	}
}

func TestRateLimiter_RefillDuration(t *testing.T) {
	limiter := NewRateLimiter(10, 5*time.Minute)

	if limiter.Refill != 5*time.Minute {
		t.Errorf("Refill = %v, want 5m", limiter.Refill)
	}
}

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(50, 10*time.Second)

	if limiter.Capacity != 50 {
		t.Errorf("Capacity = %d, want 50", limiter.Capacity)
	}
	if limiter.Refill != 10*time.Second {
		t.Errorf("Refill = %v, want 10s", limiter.Refill)
	}
	if limiter.buckets == nil {
		t.Error("buckets map should be initialized")
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewRateLimiter(100, time.Minute)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				limiter.Allow("192.168.1.1")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have consumed 100 tokens
	// Next request should be denied
	if limiter.Allow("192.168.1.1") {
		t.Error("should be rate limited after concurrent access")
	}
}

func TestRateLimiter_ZeroCapacity(t *testing.T) {
	limiter := NewRateLimiter(0, time.Minute)

	// No requests should be allowed
	if limiter.Allow("192.168.1.1") {
		t.Error("zero capacity should deny all")
	}
}

func TestRateLimiter_EmptyKey(t *testing.T) {
	limiter := NewRateLimiter(5, time.Minute)

	// Should work with empty key
	if !limiter.Allow("") {
		t.Error("empty key should be allowed")
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	limiter := NewRateLimiter(1000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("192.168.1.1")
	}
}

func BenchmarkRateLimiter_Middleware(b *testing.B) {
	limiter := NewRateLimiter(1000, time.Minute)
	handler := limiter.Middleware()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.RemoteAddr = "192.168.1.1:12345"
		handler(c)
	}
}
