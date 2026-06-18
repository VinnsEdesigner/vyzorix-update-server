// Package middleware provides HTTP middleware.
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Turnstile errors.
var (
	ErrTurnstileInvalid   = errors.New("invalid turnstile token")
	ErrTurnstileExpired   = errors.New("turnstile token expired")
	ErrTurnstileMismatch  = errors.New("turnstile token mismatch")
	ErrTurnstileFailed    = errors.New("turnstile verification failed")
)

// TurnstileConfig holds Turnstile configuration.
type TurnstileConfig struct {
	Enabled  bool
	Secret   string
	SiteKey  string
	CacheTTL time.Duration
}

// getEnvBool returns a boolean from environment variable.
func getEnvBool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	switch strings.ToLower(v) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off", "":
		return false
	default:
		return defaultVal
	}
}

// LoadTurnstileConfig loads Turnstile configuration from environment variables.
func LoadTurnstileConfig() TurnstileConfig {
	return TurnstileConfig{
		Enabled:  getEnvBool("TURNSTILE_ENABLED", false),
		Secret:   os.Getenv("TURNSTILE_SECRET"),
		SiteKey:  os.Getenv("TURNSTILE_SITE_KEY"),
		CacheTTL: 5 * time.Minute,
	}
}

// TurnstileResponse represents the Turnstile verification response.
type TurnstileResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
	Action      string    `json:"action"`
	CData       string    `json:"cdata"`
}

// TurnstileCache caches verification results to avoid repeated API calls.
type TurnstileCache struct {
	mu       sync.RWMutex
	cache    map[string]turnstileCacheEntry
	maxSize  int
	ttl      time.Duration
}

type turnstileCacheEntry struct {
	result  bool
	expires time.Time
}

// NewTurnstileCache creates a new Turnstile cache.
func NewTurnstileCache(ttl time.Duration, maxSize int) *TurnstileCache {
	return &TurnstileCache{
		cache:   make(map[string]turnstileCacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get returns cached verification result.
func (c *TurnstileCache) Get(token string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[token]
	if !exists {
		return false, false
	}
	if time.Now().After(entry.expires) {
		return false, false
	}
	return entry.result, true
}

// Set stores a verification result in cache.
func (c *TurnstileCache) Set(token string, result bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if full.
	if len(c.cache) >= c.maxSize {
		// Remove 20% oldest entries.
		evictCount := c.maxSize / 5
		removed := 0
		for k, v := range c.cache {
			if removed >= evictCount {
				break
			}
			if time.Now().After(v.expires) {
				delete(c.cache, k)
				removed++
			}
		}
	}

	c.cache[token] = turnstileCacheEntry{
		result:  result,
		expires: time.Now().Add(c.ttl),
	}
}

// TurnstileVerifier handles Cloudflare Turnstile token verification.
type TurnstileVerifier struct {
	Config TurnstileConfig
	Cache  *TurnstileCache
	Client *http.Client
}

// NewTurnstileVerifier creates a new Turnstile verifier.
func NewTurnstileVerifier(config TurnstileConfig) *TurnstileVerifier {
	return &TurnstileVerifier{
		Config: config,
		Cache:  NewTurnstileCache(config.CacheTTL, 100000),
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Verify verifies a Turnstile token with Cloudflare's API.
func (v *TurnstileVerifier) Verify(ctx context.Context, token, remoteIP string) error {
	// Check cache first.
	if result, found := v.Cache.Get(token); found {
		if !result {
			return ErrTurnstileInvalid
		}
		return nil
	}

	// Skip verification if disabled.
	if !v.Config.Enabled {
		v.Cache.Set(token, true)
		return nil
	}

	// Build request to Turnstile verify endpoint.
	formData := "secret=" + v.Config.Secret + "&response=" + token
	if remoteIP != "" {
		formData += "&remoteip=" + remoteIP
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		strings.NewReader(formData))
	if err != nil {
		return ErrTurnstileFailed
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.Client.Do(req)
	if err != nil {
		return ErrTurnstileFailed
	}
	defer func() { _ = resp.Body.Close() }()

	var result TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ErrTurnstileFailed
	}

	// Cache the result.
	v.Cache.Set(token, result.Success)

	if !result.Success {
		if len(result.ErrorCodes) > 0 {
			switch result.ErrorCodes[0] {
			case "timeout", "expired":
				return ErrTurnstileExpired
			case "invalid-input-response":
				return ErrTurnstileMismatch
			}
		}
		return ErrTurnstileInvalid
	}

	return nil
}

// TurnstileMiddleware returns a Gin middleware that verifies Turnstile tokens.
// for protected endpoints (signup, login, password reset, sensitive actions).
func TurnstileMiddleware(verifier *TurnstileVerifier) func(c *gin.Context) {
	return func(c *gin.Context) {
		// Skip if Turnstile is disabled.
		if !verifier.Config.Enabled {
			c.Next()
			return
		}

		// Get token from form or header.
		token := c.PostForm("turnstile_token")
		if token == "" {
			token = c.GetHeader("X-Turnstile-Token")
		}

		// Get client IP.
		remoteIP := c.ClientIP()

		// Verify token.
		if err := verifier.Verify(c.Request.Context(), token, remoteIP); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "verification_failed",
				"message": "Security verification failed, please try again",
			})
			return
		}

		c.Next()
	}
}

// TurnstileProtectedPaths returns a list of paths that require Turnstile verification.
func TurnstileProtectedPaths() []string {
	return []string{
		"/v1/auth/register",
		"/v1/auth/login",
		"/v1/auth/forgot-password",
		"/v1/auth/resend-password-reset",
	}
}

// ShouldVerifyTurnstile checks if a path requires Turnstile verification.
func ShouldVerifyTurnstile(path string) bool {
	protected := TurnstileProtectedPaths()
	for _, p := range protected {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}