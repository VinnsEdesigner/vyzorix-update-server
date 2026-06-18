// Package middleware provides HTTP middleware.
package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

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
	Enabled    bool
	Secret    string
	TokenLength int
	CookieName string
	HeaderName string
	MaxAge    int
}

// DefaultCSRFConfig returns the default CSRF configuration.
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		Enabled:    false,
		Secret:    "",
		TokenLength: 32,
		CookieName: "_csrf",
		HeaderName: "X-CSRF-Token",
		MaxAge:    0,
	}
}

// CSRFToken represents a CSRF token with metadata.
type CSRFToken struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt *time.Time
}

// CSRFTokenStore manages CSRF tokens in memory.
type CSRFTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*CSRFToken
}

// NewCSRFTokenStore creates a new CSRF token store.
func NewCSRFTokenStore() *CSRFTokenStore {
	return &CSRFTokenStore{
		tokens: make(map[string]*CSRFToken),
	}
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
	Config CSRFConfig
	Store  *CSRFTokenStore
	Secret []byte
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
func (p *CSRFProtector) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !p.Config.Enabled {
			c.Next()
			return
		}

		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		headerToken := c.GetHeader(p.Config.HeaderName)
		if headerToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "CSRF token required",
			})
			return
		}

		token, valid := p.verifyToken(headerToken)
		if !valid {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Invalid CSRF token",
			})
			return
		}

		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "CSRF token required",
			})
			return
		}

		if !p.Store.Validate(sessionID, token) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Invalid CSRF token",
			})
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

	secure := os.Getenv("NODE_ENV") == "production"

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
		Enabled:    enabled,
		Secret:     secret,
		TokenLength: 32,
		CookieName: "_csrf",
		HeaderName: "X-CSRF-Token",
		MaxAge:    maxAge,
	}
}