// Package middleware provides HTTP middleware.
package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/gin-gonic/gin"
)

// CSRF errors.
var (
	ErrCSRFTokenMissing = errors.New("csrf token missing")
	ErrCSRFTokenInvalid = errors.New("csrf token invalid")
	ErrCSRFTokenExpired = errors.New("csrf token expired")
)

// CSRFConfig holds CSRF configuration.
type CSRFConfig struct {
	Secret            string
	CookieName        string
	HeaderName        string
	TokenLength       int
	MaxAge            int
	Enabled           bool
	AllowDoubleSubmit bool // Allow double-submit cookie pattern without session.
}

// DefaultCSRFConfig returns the default CSRF configuration.
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		Enabled:           true,
		Secret:            "csrf-secret-change-in-production",
		TokenLength:       32,
		CookieName:        "_csrf",
		HeaderName:        "X-CSRF-Token",
		MaxAge:            3600, // 1 hour.
		AllowDoubleSubmit: true, // Enable double-submit for endpoints without session.
	}
}

// CSRFToken represents a CSRF token with metadata.
type CSRFToken struct {
	CreatedAt time.Time
	ExpiresAt *time.Time
	Token     string
}

// CSRFTokenStore manages CSRF tokens in memory.
type CSRFTokenStore struct {
	tokens map[string]*CSRFToken
	stop   chan struct{}
	mu     sync.RWMutex
}

// NewCSRFTokenStore creates a new CSRF token store.
func NewCSRFTokenStore() *CSRFTokenStore {
	store := &CSRFTokenStore{
		tokens: make(map[string]*CSRFToken),
		stop:   make(chan struct{}),
	}
	go store.cleanupExpired()
	return store
}

// cleanupExpired periodically removes expired tokens from the store.
func (s *CSRFTokenStore) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.removeExpired()
		case <-s.stop:
			return
		}
	}
}

// removeExpired removes all expired tokens.
func (s *CSRFTokenStore) removeExpired() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for sessionID, token := range s.tokens {
		if token.ExpiresAt != nil && now.After(*token.ExpiresAt) {
			delete(s.tokens, sessionID)
		}
	}
}

// Stop stops the cleanup goroutine.
func (s *CSRFTokenStore) Stop() {
	close(s.stop)
}

// Generate creates a new CSRF token for a session.
func (s *CSRFTokenStore) Generate(sessionID string, maxAge int) (*CSRFToken, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}

	tokenStr := base64.URLEncoding.EncodeToString(tokenBytes)

	now := time.Now()
	csrfToken := &CSRFToken{
		Token:     tokenStr,
		CreatedAt: now,
	}

	if maxAge > 0 {
		expires := now.Add(time.Duration(maxAge) * time.Second)
		csrfToken.ExpiresAt = &expires
	}

	s.mu.Lock()
	s.tokens[sessionID] = csrfToken
	s.mu.Unlock()

	return csrfToken, nil
}

// Validate checks if a CSRF token is valid for a session.
func (s *CSRFTokenStore) Validate(sessionID, token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storedToken, exists := s.tokens[sessionID]
	if !exists {
		return false
	}

	if storedToken.ExpiresAt != nil && time.Now().After(*storedToken.ExpiresAt) {
		return false
	}

	return hmac.Equal([]byte(storedToken.Token), []byte(token))
}

// Invalidate removes a CSRF token for a session.
func (s *CSRFTokenStore) Invalidate(sessionID string) {
	s.mu.Lock()
	delete(s.tokens, sessionID)
	s.mu.Unlock()
}

// CSRFProtector provides CSRF protection middleware.
type CSRFProtector struct {
	Store  *CSRFTokenStore
	Secret []byte
	Config CSRFConfig
}

// NewCSRFProtector creates a new CSRF protector.
func NewCSRFProtector(config CSRFConfig) *CSRFProtector {
	return &CSRFProtector{
		Config: config,
		Store:  NewCSRFTokenStore(),
		Secret: []byte(config.Secret),
	}
}

// signToken creates an HMAC signature of the token.
func (p *CSRFProtector) signToken(token string) string {
	mac := hmac.New(sha512.New, p.Secret)
	mac.Write([]byte(token))

	return token + "." + base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

// verifyToken verifies the HMAC signature of a token.
func (p *CSRFProtector) verifyToken(signed string) (string, bool) {
	for i := len(signed) - 1; i >= 0; i-- {
		if signed[i] == '.' {
			token := signed[:i]

			sig, err := base64.URLEncoding.DecodeString(signed[i+1:])
			if err != nil {
				return "", false
			}

			mac := hmac.New(sha512.New, p.Secret)
			mac.Write([]byte(token))
			expected := mac.Sum(nil)

			if hmac.Equal([]byte(sig), expected) {
				return token, true
			}

			return "", false
		}
	}

	return "", false
}

// Middleware returns a Gin middleware that validates CSRF tokens.
// Implements double-submit cookie pattern for endpoints without sessions,.
// and session-based validation for authenticated endpoints.
func (p *CSRFProtector) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add cache control headers early to prevent caching of sensitive responses.
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		if !p.Config.Enabled {
			c.Next()
			return
		}

		// Skip CSRF for safe HTTP methods.
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Check for X-Requested-With header as additional cross-origin protection.
		// AJAX requests typically send this header automatically.
		// Note: We don't block requests without this header - the double-submit check provides protection.
		_ = c.GetHeader("X-Requested-With")

		headerToken := c.GetHeader(p.Config.HeaderName)
		if headerToken == "" {
			responses.RespondStructuredAbort(c, http.StatusForbidden,

				"CSRF token required",
			)
			return
		}

		// Verify the token signature.
		token, valid := p.verifyToken(headerToken)
		if !valid {
			responses.RespondStructuredAbort(c, http.StatusForbidden,

				"Invalid CSRF token",
			)
			return
		}

		// Check for session cookie.
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			// No session - use double-submit cookie pattern.
			// For this, we just need the cookie to exist and be valid.
			// The cookie is set when the token is generated, and the client.
			// must echo back the same token in the header.
			if p.Config.AllowDoubleSubmit {
				// Double-submit: verify the cookie value matches the header value.
				// Both should contain the same signed token.
				// Note: cookie values may be URL-encoded, so we need to decode for comparison.
				cookieToken, cookieErr := c.Cookie(p.Config.CookieName)
				if cookieErr == nil {
					// URL-decode the cookie value for comparison.
					decodedCookie, decodeErr := url.QueryUnescape(cookieToken)
					if decodeErr == nil && decodedCookie == headerToken {
						// Token matches - it's a valid double-submit.
						c.Next()
						return
					}
					// Also try direct comparison in case it's not encoded.
					if cookieToken == headerToken {
						c.Next()
						return
					}
				}
			}
			responses.RespondStructuredAbort(c, http.StatusForbidden,

				"CSRF token required",
			)
			return
		}

		// Session exists - use session-based validation.
		if !p.Store.Validate(sessionID, token) {
			responses.RespondStructuredAbort(c, http.StatusForbidden,

				"Invalid CSRF token",
			)
			return
		}

		c.Next()
	}
}

// GetToken returns the current CSRF token for a session and sets the cookie.
func (p *CSRFProtector) GetToken(c *gin.Context) (string, error) {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		return "", ErrCSRFTokenMissing
	}

	token, err := p.Store.Generate(sessionID, p.Config.MaxAge)
	if err != nil {
		return "", err
	}

	signed := p.signToken(token.Token)

	secure := os.Getenv("GIN_MODE") == "release"

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		p.Config.CookieName,
		signed,
		p.Config.MaxAge,
		"/",
		"",
		secure,
		true,
	)

	return signed, nil
}

// GetTokenForPublicEndpoint generates a CSRF token for public endpoints (no session required).
// This implements the double-submit cookie pattern for endpoints like registration and login.
func (p *CSRFProtector) GetTokenForPublicEndpoint(c *gin.Context) (string, error) {
	// Generate a random token.
	tokenBytes := make([]byte, p.Config.TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)
	signed := p.signToken(token)

	secure := os.Getenv("GIN_MODE") == "release"

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		p.Config.CookieName,
		signed,
		p.Config.MaxAge,
		"/",
		"",
		secure,
		true,
	)

	return signed, nil
}

// LoadCSRFConfig loads CSRF configuration from environment variables.
func LoadCSRFConfig() CSRFConfig {
	enabled := os.Getenv("CSRF_ENABLED") == "true"
	secret := os.Getenv("CSRF_SECRET")
	maxAge := 0

	if v := os.Getenv("CSRF_TOKEN_MAX_AGE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxAge = n
		}
	}

	return CSRFConfig{
		Enabled:     enabled,
		Secret:      secret,
		TokenLength: 32,
		CookieName:  "_csrf",
		HeaderName:  "X-CSRF-Token",
		MaxAge:      maxAge,
	}
}
