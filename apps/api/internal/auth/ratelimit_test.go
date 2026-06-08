package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 3)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.Allow("test-key") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.Allow("test-key") {
		t.Error("4th request should be denied")
	}
}

func TestRateLimiter_AllowDifferentKeys(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 2)

	// Different keys should have separate limits
	if !rl.Allow("key1") {
		t.Error("key1 request 1 should be allowed")
	}
	if !rl.Allow("key1") {
		t.Error("key1 request 2 should be allowed")
	}
	if rl.Allow("key1") {
		t.Error("key1 request 3 should be denied")
	}

	// key2 should still have its own quota
	if !rl.Allow("key2") {
		t.Error("key2 request 1 should be allowed")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 2)

	// Use up the limit
	if !rl.Allow("test-key") {
		t.Error("request 1 should be allowed")
	}
	if !rl.Allow("test-key") {
		t.Error("request 2 should be allowed")
	}
	if rl.Allow("test-key") {
		t.Error("request 3 should be denied")
	}

	// Wait for window to reset
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	if !rl.Allow("test-key") {
		t.Error("request after window reset should be allowed")
	}
}

func TestRateLimiter_GetRemaining(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 5)

	if remaining := rl.GetRemaining("test"); remaining != 5 {
		t.Errorf("initial remaining should be 5, got %d", remaining)
	}

	rl.Allow("test")
	rl.Allow("test")

	if remaining := rl.GetRemaining("test"); remaining != 3 {
		t.Errorf("remaining should be 3, got %d", remaining)
	}

	// Non-existent key
	if remaining := rl.GetRemaining("nonexistent"); remaining != 5 {
		t.Errorf("non-existent key should have 5 remaining, got %d", remaining)
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 2)

	// Use up the limit
	rl.Allow("test-key")
	rl.Allow("test-key")
	if rl.Allow("test-key") {
		t.Error("request should be denied")
	}

	// Reset
	rl.Reset("test-key")

	// Should be allowed again
	if !rl.Allow("test-key") {
		t.Error("request after reset should be allowed")
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 2)

	router := gin.New()
	router.Use(rl.Middleware(RateLimitConfig{
		KeyFunc: DefaultKeyFunc,
	}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// 3rd request should be rate limited
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("request 3: expected status 429, got %d", w.Code)
	}
}

func TestRateLimiter_MiddlewareHeaders(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 5)

	router := gin.New()
	router.Use(rl.Middleware(RateLimitConfig{
		KeyFunc: DefaultKeyFunc,
	}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w, req)

	// Check rate limit headers
	if limit := w.Header().Get("X-RateLimit-Limit"); limit != "5" {
		t.Errorf("X-RateLimit-Limit should be 5, got %s", limit)
	}
	if remaining := w.Header().Get("X-RateLimit-Remaining"); remaining != "4" {
		t.Errorf("X-RateLimit-Remaining should be 4, got %s", remaining)
	}
}

func TestRateLimiter_MiddlewareCustomKey(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 2)

	router := gin.New()
	router.Use(rl.Middleware(RateLimitConfig{
		KeyFunc: func(c *gin.Context) string {
			return c.GetHeader("X-User-ID")
		},
	}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test with different user IDs - each should have its own limit
	// User 1: 2 requests allowed
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "user1")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("user1 request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// User 1's 3rd request should be denied
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "user1")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("user1 request 3: expected status 429, got %d", w.Code)
	}

	// User 2 should have its own quota (not affected by user1's limit)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "user2")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("user2 request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// User 2's 3rd request should be denied
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "user2")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("user2 request 3: expected status 429, got %d", w.Code)
	}
}

func TestMultiWindowLimiter(t *testing.T) {
	ml := NewMultiWindowLimiter(map[string]struct {
		Window time.Duration
		Max    int
	}{
		"minute": {Window: time.Minute, Max: 3},
		"hour":   {Window: time.Hour, Max: 10},
	})

	// Simulate 3 requests (minute limit)
	for i := 0; i < 3; i++ {
		if !ml.limiters["minute"].Allow("test") {
			t.Errorf("minute request %d should be allowed", i+1)
		}
	}

	// 4th should be denied by minute limiter
	if ml.limiters["minute"].Allow("test") {
		t.Error("minute request 4 should be denied")
	}

	// But hour limiter should still allow
	if !ml.limiters["hour"].Allow("test") {
		t.Error("hour request should still be allowed")
	}
}

func TestMultiWindowLimiter_Middleware(t *testing.T) {
	ml := NewMultiWindowLimiter(map[string]struct {
		Window time.Duration
		Max    int
	}{
		"minute": {Window: time.Minute, Max: 2},
	})

	router := gin.New()
	router.Use(ml.Middleware(DefaultKeyFunc))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// 3rd request should be rate limited
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("request 3: expected status 429, got %d", w.Code)
	}
}

func TestDefaultKeyFunc(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.RemoteAddr = "192.168.1.100:12345"

	key := DefaultKeyFunc(c)
	if key != "192.168.1.100" {
		t.Errorf("expected IP '192.168.1.100', got '%s'", key)
	}
}