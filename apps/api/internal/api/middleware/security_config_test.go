// Package middleware provides tests for security configuration.
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
)

// ============================================================================.
// 4.1 SIGNING - Tests using config.SigningConfig.
// ============================================================================.

func TestSigningConfig(t *testing.T) {
	t.Run("default_signing_config", func(t *testing.T) {
		cfg := DefaultSigningConfig()
		assert.True(t, cfg.Enabled)
		assert.Equal(t, 5*time.Minute, cfg.TimestampWindow)
		assert.Equal(t, 100000, cfg.MaxCacheSize)
	})

	t.Run("signing_config_from_pkg", func(t *testing.T) {
		cfg := config.SigningConfig{
			Enabled:         true,
			TimestampWindow: 300, // 5 minutes in seconds
			MaxCacheSize:    1000,
		}
		assert.True(t, cfg.Enabled)
		assert.Equal(t, 300, cfg.TimestampWindow)
	})
}

func TestReplayProtectionFromConfig(t *testing.T) {
	t.Run("replay_protection_with_config", func(t *testing.T) {
		cfg := config.SigningConfig{
			Enabled:         true,
			TimestampWindow: 300,
			MaxCacheSize:    100,
		}
		rp := NewReplayProtection(cfg)
		assert.NotNil(t, rp)
	})
}

// ============================================================================.
// 4.2 SECURITY HEADERS - Tests.
// ============================================================================.

func TestSecurityHeadersComplete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("all_security_headers", func(t *testing.T) {
		r := gin.New()
		r.Use(SecurityHeaders())
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
		assert.NotEmpty(t, w.Header().Get("Strict-Transport-Security"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	})

	t.Run("relaxed_headers", func(t *testing.T) {
		r := gin.New()
		r.Use(SecurityHeadersRelaxed())
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		csp := w.Header().Get("Content-Security-Policy")
		assert.Contains(t, csp, "'unsafe-inline'")
	})
}

// ============================================================================.
// 4.3 TURNSTILE - Tests.
// ============================================================================.

func TestTurnstileProtectedPaths(t *testing.T) {
	testCases := []struct {
		path      string
		protected bool
	}{
		{"/v1/auth/register", true},
		{"/v1/auth/login", true},
// DEPRECATED: // DEPRECATED: 		{"/v1/device/register", false},
		{"/health", false},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := ShouldVerifyTurnstile(tc.path)
			assert.Equal(t, tc.protected, result)
		})
	}
}

func TestTurnstileCache(t *testing.T) {
	t.Run("cache_set_get", func(t *testing.T) {
		cache := NewTurnstileCache(5*time.Minute, 100)
		cache.Set("token1", true)
		result, found := cache.Get("token1")
		assert.True(t, found)
		assert.True(t, result)
	})

	t.Run("cache_miss", func(t *testing.T) {
		cache := NewTurnstileCache(5*time.Minute, 100)
		_, found := cache.Get("nonexistent")
		assert.False(t, found)
	})
}

// ============================================================================.
// 4.6 CSRF - Tests.
// ============================================================================.

func TestCSRFProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("token_generated", func(t *testing.T) {
		cfg := CSRFConfig{
			Enabled:     true,
			Secret:      "test-secret",
			TokenLength: 32,
		}
		protector := NewCSRFProtector(cfg)

		r := gin.New()
		r.Use(protector.Middleware())
		r.GET("/form", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/form", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// GET requests are exempt from CSRF.
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid_token_rejected", func(t *testing.T) {
		cfg := CSRFConfig{
			Enabled: true,
			Secret:  "test-secret",
		}
		protector := NewCSRFProtector(cfg)

		r := gin.New()
		r.Use(protector.Middleware())
		r.POST("/submit", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodPost, "/submit", nil)
		req.Header.Set("X-CSRF-Token", "invalid")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("safe_methods_exempt", func(t *testing.T) {
		cfg := CSRFConfig{Enabled: true, Secret: "test"}
		protector := NewCSRFProtector(cfg)

		r := gin.New()
		r.Use(protector.Middleware())
		r.GET("/fetch", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/fetch", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCSRFTokenStore(t *testing.T) {
	t.Run("generate_token", func(t *testing.T) {
		store := NewCSRFTokenStore()
		token, err := store.Generate("session1", 3600)
		assert.NoError(t, err)
		assert.NotEmpty(t, token.Token)
	})

	t.Run("validate_token", func(t *testing.T) {
		store := NewCSRFTokenStore()
		token, _ := store.Generate("session1", 3600)
		valid := store.Validate("session1", token.Token)
		assert.True(t, valid)
	})

	t.Run("invalidate_token", func(t *testing.T) {
		store := NewCSRFTokenStore()
		token, _ := store.Generate("session1", 3600)
		store.Invalidate("session1")
		valid := store.Validate("session1", token.Token)
		assert.False(t, valid)
	})
}

// ============================================================================.
// 4.9 ACCOUNT LOCKOUT - Tests.
// ============================================================================.

func TestAccountLockout(t *testing.T) {
	t.Run("default_lockout_config", func(t *testing.T) {
		cfg := DefaultLockoutConfig()
		assert.False(t, cfg.Enabled)
		assert.Equal(t, 5, cfg.MaxAttempts)
		assert.Equal(t, time.Hour, cfg.LockoutDuration)
	})

	t.Run("lockout_after_attempts", func(t *testing.T) {
		cfg := LockoutConfig{
			Enabled:            true,
			MaxAttempts:        3,
			LockoutDuration:    15 * time.Minute,
			MaxLockoutDuration: 24 * time.Hour,
		}
		lockout := NewLockout(cfg)

		email := "test@example.com"
		for i := 0; i < 3; i++ {
			err := lockout.RecordFailedAttempt(email)
			assert.NoError(t, err)
		}

		err := lockout.RecordFailedAttempt(email)
		assert.Error(t, err)
		assert.Equal(t, ErrAccountLocked, err)
	})

	t.Run("successful_login_clears", func(t *testing.T) {
		cfg := LockoutConfig{Enabled: true, MaxAttempts: 3}
		lockout := NewLockout(cfg)
		email := "test@example.com"

		lockout.RecordFailedAttempt(email)
		lockout.RecordFailedAttempt(email)
		lockout.RecordSuccessfulAttempt(email)

		err := lockout.RecordFailedAttempt(email)
		assert.NoError(t, err)
	})

	t.Run("lockout_middleware", func(t *testing.T) {
		cfg := LockoutConfig{Enabled: true, MaxAttempts: 2, LockoutDuration: time.Hour, MaxLockoutDuration: 24 * time.Hour}
		lockout := NewLockout(cfg)
		email := "locked@example.com"

		lockout.RecordFailedAttempt(email)
		lockout.RecordFailedAttempt(email)
		lockout.RecordFailedAttempt(email)

		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(LockoutMiddleware(lockout))
		r.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodPost, "/login?email="+email, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})
}

// ============================================================================.
// 4.12 HEALTH - Tests.
// ============================================================================.

func TestHealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("liveness", func(t *testing.T) {
		r := gin.New()
		r.GET("/health/live", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("readiness", func(t *testing.T) {
		r := gin.New()
		r.GET("/health/ready", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
		})

		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ============================================================================.
// 4.15 CI/CD SECURITY - Config Tests.
// ============================================================================.

func TestSecurityConfigurations(t *testing.T) {
	t.Run("signing_config", func(t *testing.T) {
		cfg := DefaultSigningConfig()
		assert.NotZero(t, cfg.TimestampWindow)
	})

	t.Run("csrf_config", func(t *testing.T) {
		cfg := DefaultCSRFConfig()
		assert.NotZero(t, cfg.TokenLength)
	})

	t.Run("lockout_config", func(t *testing.T) {
		cfg := DefaultLockoutConfig()
		assert.NotZero(t, cfg.MaxAttempts)
	})
}

// ============================================================================.
// 4.16 DOCUMENTATION - Tests.
// ============================================================================.

func TestSecurityDocumentation(t *testing.T) {
	t.Run("security_headers_funcs", func(t *testing.T) {
		assert.NotNil(t, SecurityHeaders())
		assert.NotNil(t, SecurityHeadersRelaxed())
	})

	t.Run("csrf_config_funcs", func(t *testing.T) {
		assert.NotNil(t, DefaultCSRFConfig())
	})
}
