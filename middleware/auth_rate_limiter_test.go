package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewAuthRateLimiter(t *testing.T) {
	limiter := NewAuthRateLimiter(DefaultAuthRateLimits)

	if limiter.LoginLimiter == nil {
		t.Error("LoginLimiter should be initialized")
	}
	if limiter.RegisterLimiter == nil {
		t.Error("RegisterLimiter should be initialized")
	}
	if limiter.PasswordResetLimiter == nil {
		t.Error("PasswordResetLimiter should be initialized")
	}
}

func TestAuthRateLimiter_LoginLimits(t *testing.T) {
	limiter := NewAuthRateLimiter(DefaultAuthRateLimits)

	// Login should allow 10 requests per minute
	for i := 0; i < 10; i++ {
		if !limiter.LoginLimiter.Allow("192.168.1.1") {
			t.Errorf("login request %d should be allowed", i+1)
		}
	}

	// 11th request should be denied
	if limiter.LoginLimiter.Allow("192.168.1.1") {
		t.Error("11th login request should be denied")
	}
}

func TestAuthRateLimiter_RegisterLimits(t *testing.T) {
	limiter := NewAuthRateLimiter(DefaultAuthRateLimits)

	// Register should allow 5 requests per minute
	for i := 0; i < 5; i++ {
		if !limiter.RegisterLimiter.Allow("192.168.1.1") {
			t.Errorf("register request %d should be allowed", i+1)
		}
	}

	// 6th request should be denied
	if limiter.RegisterLimiter.Allow("192.168.1.1") {
		t.Error("6th register request should be denied")
	}
}

func TestAuthRateLimiter_PasswordResetLimits(t *testing.T) {
	limiter := NewAuthRateLimiter(DefaultAuthRateLimits)

	// Password reset should allow 3 requests per minute
	for i := 0; i < 3; i++ {
		if !limiter.PasswordResetLimiter.Allow("192.168.1.1") {
			t.Errorf("password reset request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if limiter.PasswordResetLimiter.Allow("192.168.1.1") {
		t.Error("4th password reset request should be denied")
	}
}

func TestAuthRateLimiter_DifferentIPs(t *testing.T) {
	limiter := NewAuthRateLimiter(DefaultAuthRateLimits)

	// Different IPs should have separate limits
	if !limiter.LoginLimiter.Allow("192.168.1.1") {
		t.Error("IP1 login should be allowed")
	}
	if !limiter.LoginLimiter.Allow("192.168.1.2") {
		t.Error("IP2 login should be allowed")
	}

	// IP1 should be at limit (1 used)
	// IP2 should have capacity left (1 used)
	if limiter.LoginLimiter.Allow("192.168.1.1") {
		t.Error("IP1 second login should be denied")
	}
	if !limiter.LoginLimiter.Allow("192.168.1.2") {
		t.Error("IP2 second login should be allowed")
	}
}

func TestAuthRateLimiter_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewAuthRateLimiter(DefaultAuthRateLimits)

	t.Run("LoginMiddleware", func(t *testing.T) {
		m := limiter.LoginMiddleware()

		// First 10 requests should succeed
		for i := 0; i < 10; i++ {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/v1/auth/login", nil)
			c.Request.RemoteAddr = "192.168.1.1:12345"

			m(c)

			if w.Code != http.StatusOK {
				t.Errorf("request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
			}
		}

		// 11th request should be rate limited
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/auth/login", nil)
		c.Request.RemoteAddr = "192.168.1.1:12345"

		m(c)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("rate limited request: got status %d, want %d", w.Code, http.StatusTooManyRequests)
		}

		body := w.Body.String()
		if body == "" {
			t.Error("expected error message in response body")
		}
	})

	t.Run("RegisterMiddleware", func(t *testing.T) {
		m := limiter.RegisterMiddleware()

		// First 5 requests should succeed
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/v1/auth/register", nil)
			c.Request.RemoteAddr = "192.168.1.1:12345"

			m(c)

			if w.Code != http.StatusOK {
				t.Errorf("request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
			}
		}

		// 6th request should be rate limited
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/auth/register", nil)
		c.Request.RemoteAddr = "192.168.1.1:12345"

		m(c)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("rate limited request: got status %d, want %d", w.Code, http.StatusTooManyRequests)
		}
	})

	t.Run("PasswordResetMiddleware", func(t *testing.T) {
		m := limiter.PasswordResetMiddleware()

		// First 3 requests should succeed
		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/v1/auth/forgot-password", nil)
			c.Request.RemoteAddr = "192.168.1.1:12345"

			m(c)

			if w.Code != http.StatusOK {
				t.Errorf("request %d: got status %d, want %d", i+1, w.Code, http.StatusOK)
			}
		}

		// 4th request should be rate limited
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/auth/forgot-password", nil)
		c.Request.RemoteAddr = "192.168.1.1:12345"

		m(c)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("rate limited request: got status %d, want %d", w.Code, http.StatusTooManyRequests)
		}
	})
}

func TestAuthRateLimiter_GetStatus(t *testing.T) {
	limiter := NewAuthRateLimiter(DefaultAuthRateLimits)

	// Use some tokens
	limiter.LoginLimiter.Allow("192.168.1.1")
	limiter.LoginLimiter.Allow("192.168.1.1")
	limiter.RegisterLimiter.Allow("192.168.1.1")

	status := limiter.GetStatus("192.168.1.1")

	if status.LoginLimit != 10 {
		t.Errorf("LoginLimit = %d, want 10", status.LoginLimit)
	}
	if status.LoginRemaining > 10 || status.LoginRemaining < 0 {
		t.Errorf("LoginRemaining = %d, expected between 0 and 10", status.LoginRemaining)
	}
	if status.RegisterLimit != 5 {
		t.Errorf("RegisterLimit = %d, want 5", status.RegisterLimit)
	}
	if status.PasswordResetLimit != 3 {
		t.Errorf("PasswordResetLimit = %d, want 3", status.PasswordResetLimit)
	}
}

func TestAuthRateLimiter_GetStatus_NewIP(t *testing.T) {
	limiter := NewAuthRateLimiter(DefaultAuthRateLimits)

	// New IP should have full capacity
	status := limiter.GetStatus("192.168.1.100")

	if status.LoginRemaining != 10 {
		t.Errorf("LoginRemaining for new IP = %d, want 10", status.LoginRemaining)
	}
	if status.RegisterRemaining != 5 {
		t.Errorf("RegisterRemaining for new IP = %d, want 5", status.RegisterRemaining)
	}
	if status.PasswordResetRemaining != 3 {
		t.Errorf("PasswordResetRemaining for new IP = %d, want 3", status.PasswordResetRemaining)
	}
}

func TestStrictAuthRateLimits(t *testing.T) {
	limiter := NewAuthRateLimiter(StrictAuthRateLimits)

	// Strict should allow fewer requests
	if limiter.LoginLimiter.Capacity != 5 {
		t.Errorf("Strict login limit = %d, want 5", limiter.LoginLimiter.Capacity)
	}
	if limiter.RegisterLimiter.Capacity != 3 {
		t.Errorf("Strict register limit = %d, want 3", limiter.RegisterLimiter.Capacity)
	}
	if limiter.PasswordResetLimiter.Capacity != 2 {
		t.Errorf("Strict password reset limit = %d, want 2", limiter.PasswordResetLimiter.Capacity)
	}
}

func TestAuthRateLimiter_Refill(t *testing.T) {
	limiter := NewAuthRateLimiter(AuthRateLimitConfig{
		LoginLimit:      2,
		RegisterLimit:   2,
		PasswordResetLimit: 2,
		RefillDuration:  100 * time.Millisecond,
	})

	// Exhaust login tokens
	limiter.LoginLimiter.Allow("192.168.1.1")
	limiter.LoginLimiter.Allow("192.168.1.1")

	// Should be denied
	if limiter.LoginLimiter.Allow("192.168.1.1") {
		t.Error("should be rate limited after exhausting tokens")
	}

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	if !limiter.LoginLimiter.Allow("192.168.1.1") {
		t.Error("should be allowed after token refill")
	}
}

func TestDefaultAuthRateLimits_Values(t *testing.T) {
	if DefaultAuthRateLimits.LoginLimit != 10 {
		t.Errorf("Default login limit = %d, want 10", DefaultAuthRateLimits.LoginLimit)
	}
	if DefaultAuthRateLimits.RegisterLimit != 5 {
		t.Errorf("Default register limit = %d, want 5", DefaultAuthRateLimits.RegisterLimit)
	}
	if DefaultAuthRateLimits.PasswordResetLimit != 3 {
		t.Errorf("Default password reset limit = %d, want 3", DefaultAuthRateLimits.PasswordResetLimit)
	}
	if DefaultAuthRateLimits.RefillDuration != time.Minute {
		t.Errorf("Default refill duration = %v, want 1m", DefaultAuthRateLimits.RefillDuration)
	}
}

func TestStrictAuthRateLimits_Values(t *testing.T) {
	if StrictAuthRateLimits.LoginLimit != 5 {
		t.Errorf("Strict login limit = %d, want 5", StrictAuthRateLimits.LoginLimit)
	}
	if StrictAuthRateLimits.RegisterLimit != 3 {
		t.Errorf("Strict register limit = %d, want 3", StrictAuthRateLimits.RegisterLimit)
	}
	if StrictAuthRateLimits.PasswordResetLimit != 2 {
		t.Errorf("Strict password reset limit = %d, want 2", StrictAuthRateLimits.PasswordResetLimit)
	}
}