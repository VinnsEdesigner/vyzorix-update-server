package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
	"sync"
	"time"

	apikeyapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/api_key"
	apikeydomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/api_key"
	"github.com/gin-gonic/gin"
)

// TenantAPIKeyAuth handles multi-tenant API key authentication with scope enforcement.
type TenantAPIKeyAuth struct {
	service     *apikeyapp.Service
	rateLimiter *APIKeyRateLimiter
	auditLogger func(keyID, operatorID, clientIP, path, method string, success bool)
}

// NewTenantAPIKeyAuth creates a new TenantAPIKeyAuth middleware.
func NewTenantAPIKeyAuth(service *apikeyapp.Service, rateLimit int) *TenantAPIKeyAuth {
	return &TenantAPIKeyAuth{
		service:     service,
		rateLimiter: NewAPIKeyRateLimiter(rateLimit),
	}
}

// SetAuditLogger sets the audit logging function.
func (t *TenantAPIKeyAuth) SetAuditLogger(logger func(keyID, operatorID, clientIP, path, method string, success bool)) {
	t.auditLogger = logger
}

// Middleware returns a Gin middleware that validates the API key and enforces scope.
func (t *TenantAPIKeyAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Extract API key from headers
		apiKeyValue := c.GetHeader("X-API-Key")
		if apiKeyValue == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKeyValue = auth[7:]
			}
		}

		// Empty key check
		if apiKeyValue == "" {
			t.logAudit("", "", clientIP, path, method, false)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "api_key_required",
				"message": "API key is required",
			})
			c.Abort()
			return
		}

		// Validate the API key
		ctx := c.Request.Context()
		key, err := t.service.ValidateKey(ctx, apiKeyValue)
		if err != nil {
			status := http.StatusUnauthorized
			errCode := "invalid_api_key"

			switch err {
			case apikeydomain.ErrAPIKeyExpired:
				errCode = "expired_api_key"
			case apikeydomain.ErrAPIKeyRevoked:
				errCode = "revoked_api_key"
			case apikeydomain.ErrAPIKeyInactive:
				errCode = "inactive_api_key"
			}

			t.logAudit("", "", clientIP, path, method, false)
			c.JSON(status, gin.H{
				"error":   errCode,
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Check rate limit
		rateLimitResult := t.rateLimiter.Allow(key.ID)
		if !rateLimitResult.Allowed {
			t.logAudit(key.ID, key.OperatorID, clientIP, path, method, false)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "rate limit exceeded",
				"retry_after": rateLimitResult.RetryAfter,
			})
			c.Abort()
			return
		}

		// Enforce scope based on HTTP method
		requiredScope := getRequiredScope(method)
		if !hasScope(key.Scope, requiredScope) {
			t.logAudit(key.ID, key.OperatorID, clientIP, path, method, false)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "insufficient_scope",
				"message": "insufficient scope for this operation",
			})
			c.Abort()
			return
		}

		// Success - log audit and increment usage
		t.logAudit(key.ID, key.OperatorID, clientIP, path, method, true)

		// Increment request counter asynchronously
		go func() {
			bgCtx := context.Background()
			t.service.IncrementUsage(bgCtx, key.ID)
		}()

		// Store key info in context for downstream use
		c.Set("api_key_id", key.ID)
		c.Set("api_key_operator_id", key.OperatorID)
		c.Set("api_key_scope", string(key.Scope))

		c.Next()
	}
}

// ScopeEnforcement returns a middleware that enforces scope based on HTTP method.
func (t *TenantAPIKeyAuth) ScopeEnforcement(requiredScope apikeydomain.Scope) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopeVal, exists := c.Get("api_key_scope")
		if !exists {
			c.Next()
			return
		}

		scope := apikeydomain.Scope(scopeVal.(string))
		if !hasScope(scope, requiredScope) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "insufficient_scope",
				"message": "insufficient scope for this operation",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SkipCheck returns a middleware that skips API key check for specific paths.
func (t *TenantAPIKeyAuth) SkipCheck(paths ...string) gin.HandlerFunc {
	pathMap := make(map[string]bool)
	for _, p := range paths {
		pathMap[p] = true
	}

	return func(c *gin.Context) {
		if pathMap[c.Request.URL.Path] {
			c.Next()
			return
		}
		t.Middleware()(c)
	}
}

// logAudit logs authentication attempts.
func (t *TenantAPIKeyAuth) logAudit(keyID, operatorID, clientIP, path, method string, success bool) {
	if t.auditLogger != nil {
		t.auditLogger(keyID, operatorID, clientIP, path, method, success)
	}
}

// getRequiredScope returns the minimum scope required for an HTTP method.
func getRequiredScope(method string) apikeydomain.Scope {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS":
		return apikeydomain.ScopeRead
	case "POST", "PUT", "PATCH":
		return apikeydomain.ScopeWrite
	case "DELETE":
		return apikeydomain.ScopeAdmin
	default:
		return apikeydomain.ScopeAdmin
	}
}

// hasScope checks if a scope satisfies a required scope.
func hasScope(scope, required apikeydomain.Scope) bool {
	switch required {
	case apikeydomain.ScopeRead:
		return true
	case apikeydomain.ScopeWrite:
		return scope == apikeydomain.ScopeWrite || scope == apikeydomain.ScopeAdmin
	case apikeydomain.ScopeAdmin:
		return scope == apikeydomain.ScopeAdmin
	default:
		return false
	}
}

// APIKeyRateLimiter implements rate limiting for API keys.
type APIKeyRateLimiter struct {
	mu     sync.RWMutex
	data   map[string]*apikeyRateLimitEntry
	limit  int
	window time.Duration
	clock  func() time.Time
}

type apikeyRateLimitEntry struct {
	count     int64
	windowEnd time.Time
}

// NewAPIKeyRateLimiter creates a new API key rate limiter.
func NewAPIKeyRateLimiter(limit int) *APIKeyRateLimiter {
	return &APIKeyRateLimiter{
		data:   make(map[string]*apikeyRateLimitEntry),
		limit:  limit,
		window: time.Minute,
		clock:  time.Now,
	}
}

// RateLimitResult represents the result of a rate limit check.
type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	ResetAt    time.Time
	RetryAfter int
}

// Allow checks if a request is allowed under the rate limit.
func (r *APIKeyRateLimiter) Allow(keyID string) *RateLimitResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.clock()
	windowEnd := now.Add(r.window).Truncate(r.window).Add(r.window)

	entry, exists := r.data[keyID]
	if !exists || now.After(entry.windowEnd) {
		entry = &apikeyRateLimitEntry{
			windowEnd: windowEnd,
		}
		r.data[keyID] = entry
	}

	entry.count++

	remaining := r.limit - int(entry.count)
	if remaining < 0 {
		remaining = 0
	}

	if entry.count > int64(r.limit) {
		retryAfter := int(entry.windowEnd.Sub(now).Seconds())
		if retryAfter < 0 {
			retryAfter = 0
		}
		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetAt:    entry.windowEnd,
			RetryAfter: retryAfter,
		}
	}

	return &RateLimitResult{
		Allowed:    true,
		Remaining:  remaining,
		ResetAt:    entry.windowEnd,
		RetryAfter: 0,
	}
}

// Reset clears the rate limit for a key.
func (r *APIKeyRateLimiter) Reset(keyID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, keyID)
}

// ConstantTimeCompareString compares two strings in constant time.
func ConstantTimeCompareString(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
