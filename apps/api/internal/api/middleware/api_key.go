// Package middleware provides HTTP middleware.
package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth provides API key authentication with rate limiting and audit logging.
type APIKeyAuth struct {
	auditLogger     func(keyID, clientIP, path, method string, success bool)
	rateLimiter     *authRateLimiter
	validKeys       map[string]string // key -> key_id (for audit)
	keyPrefix       string
	mu              sync.RWMutex
}

// authRateLimiter tracks failed attempts per IP.
type authRateLimiter struct {
	attempts map[string]*rateLimitEntry
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

type rateLimitEntry struct {
	firstSeen time.Time
	count     int
}

// NewAPIKeyAuth creates a new APIKeyAuth middleware.
// apiKeys is a map of key_id -> key_value (supports rotation).
func NewAPIKeyAuth(apiKeys map[string]string, keyPrefix string) *APIKeyAuth {
	auth := &APIKeyAuth{
		validKeys:   apiKeys,
		keyPrefix:   keyPrefix,
		rateLimiter: newAuthRateLimiter(5, 15*time.Minute), // 5 failures per 15 min
	}

	return auth
}

// SetAuditLogger sets the audit logging function.
func (a *APIKeyAuth) SetAuditLogger(logger func(keyID, clientIP, path, method string, success bool)) {
	a.auditLogger = logger
}

// Middleware returns a Gin middleware that validates the API key.
// The API key can be provided via X-API-Key header or Authorization header with Bearer prefix.
func (a *APIKeyAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Check rate limit first
		if a.rateLimiter.isLimited(clientIP) {
			a.logAudit("rate_limited", clientIP, path, method, false)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "too many failed attempts",
			})
			c.Abort()
			return
		}

		// Extract API key from headers
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = auth[7:]
			}
		}

		// Empty key check
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid API key",
			})
			c.Abort()
			return
		}

		// Validate API key using constant-time comparison
		keyID := a.findKeyID(apiKey)
		if keyID == "" {
			// Key not found - record failed attempt
			a.rateLimiter.recordFailure(clientIP)
			a.logAudit("unknown", clientIP, path, method, false)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid API key",
			})
			c.Abort()
			return
		}

		// Success - reset rate limit and log
		a.rateLimiter.reset(clientIP)
		a.logAudit(keyID, clientIP, path, method, true)

		// Store key ID in context for downstream use
		c.Set("api_key_id", keyID)

		c.Next()
	}
}

// findKeyID finds the key ID for a given API key value.
func (a *APIKeyAuth) findKeyID(apiKey string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for keyID, key := range a.validKeys {
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(key)) == 1 {
			return keyID
		}
	}
	return ""
}

// SkipCheck returns a middleware that skips API key check for specific paths.
func (a *APIKeyAuth) SkipCheck(paths ...string) gin.HandlerFunc {
	pathMap := make(map[string]bool)
	for _, p := range paths {
		pathMap[p] = true
	}

	return func(c *gin.Context) {
		if pathMap[c.Request.URL.Path] {
			c.Next()
			return
		}
		a.Middleware()(c)
	}
}

// logAudit logs authentication attempts.
func (a *APIKeyAuth) logAudit(keyID, clientIP, path, method string, success bool) {
	if a.auditLogger != nil {
		a.auditLogger(keyID, clientIP, path, method, success)
	}
}

// newAuthRateLimiter creates a new rate limiter for auth failures.
func newAuthRateLimiter(limit int, window time.Duration) *authRateLimiter {
	return &authRateLimiter{
		attempts: make(map[string]*rateLimitEntry),
		limit:    limit,
		window:    window,
	}
}

// isLimited checks if an IP is rate limited.
func (r *authRateLimiter) isLimited(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.attempts[ip]
	if !exists {
		return false
	}

	// Check if window has expired
	if time.Since(entry.firstSeen) > r.window {
		delete(r.attempts, ip)
		return false
	}

	return entry.count >= r.limit
}

// recordFailure records a failed attempt for an IP.
func (r *authRateLimiter) recordFailure(ip string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.attempts[ip]
	if !exists {
		r.attempts[ip] = &rateLimitEntry{
			count:     1,
			firstSeen: time.Now(),
		}
		return
	}

	// Reset if window expired
	if time.Since(entry.firstSeen) > r.window {
		entry.count = 1
		entry.firstSeen = time.Now()
		return
	}

	entry.count++
}

// reset clears the rate limit for an IP.
func (r *authRateLimiter) reset(ip string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.attempts, ip)
}

