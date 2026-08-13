// Package middleware provides HTTP middleware.
package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/idempotency"
	"github.com/gin-gonic/gin"
)

// IdempotencyConfig holds configuration for idempotency middleware.
type IdempotencyConfig struct {
	Repository    idempotency.Repository
	HeaderName   string
	Paths        []string       // Explicit paths to apply idempotency (exact match).
	PathPrefixes []string       // Path prefixes to apply idempotency (prefix match).
	PathPatterns []string       // Regex patterns to apply idempotency.
	ExcludedPaths []string       // Paths to exclude from idempotency.
	TTL          time.Duration
	Enabled      bool
}

// DefaultIdempotencyConfig returns the default idempotency configuration.
func DefaultIdempotencyConfig() IdempotencyConfig {
	return IdempotencyConfig{
		TTL:    24 * time.Hour,
		HeaderName: "X-Idempotency-Key",
		Enabled:     true,
		PathPrefixes: []string{"/v1/", "/api/v1/"},
		ExcludedPaths: []string{
			"/v1/auth/login",
			"/v1/auth/register",
			"/v1/auth/csrf-token",
			"/v1/auth/refresh",
			"/v1/auth/logout",
			"/health",
			"/healthz",
			"/metrics",
		},
	}
}

// compiledPattern holds a compiled regex for path matching.
type compiledPattern struct {
	pattern *regexp.Regexp
	raw     string
}

// GinIdempotency returns a Gin middleware that handles idempotency keys.
// It applies idempotency handling to POST/PUT/PATCH requests on configured paths.
func GinIdempotency(config IdempotencyConfig) func(c *gin.Context) {
	if config.TTL == 0 {
		config.TTL = 24 * time.Hour
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-Idempotency-Key"
	}

	// Pre-compile regex patterns for efficiency.
	compiledPatterns := compilePatterns(config.PathPatterns)

	return func(c *gin.Context) {
		if !config.Enabled {
			c.Next()
			return
		}

		if !isIdempotentMethod(c.Request.Method) {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		if isPathExcluded(path, config.ExcludedPaths) {
			c.Next()
			return
		}

		if !isPathSupported(path, config.Paths, config.PathPrefixes, compiledPatterns) {
			c.Next()
			return
		}

		idempotencyKey := c.GetHeader(config.HeaderName)
		if idempotencyKey == "" {
			c.Next()
			return
		}

		if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"code":    "INVALID_IDEMPOTENCY_KEY",
				"message": err.Error(),
			})
			return
		}

		handleIdempotentRequest(c, config, path, idempotencyKey)
	}
}

// compilePatterns pre-compiles regex patterns for efficiency.
func compilePatterns(patterns []string) []compiledPattern {
	var compiled []compiledPattern
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err == nil {
			compiled = append(compiled, compiledPattern{pattern: re, raw: p})
		}
	}
	return compiled
}

// isIdempotentMethod returns true if the method should be idempotent.
func isIdempotentMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPatch || method == http.MethodPut || method == http.MethodDelete
}

// isPathExcluded returns true if the path should be excluded from idempotency.
func isPathExcluded(path string, excludedPaths []string) bool {
	for _, excl := range excludedPaths {
		if excl == path || strings.HasPrefix(path, excl+"/") {
			return true
		}
	}
	return false
}

// isPathSupported returns true if the path should have idempotency applied.
func isPathSupported(path string, explicitPaths, prefixes []string, patterns []compiledPattern) bool {
	// If no constraints, apply to all.
	if len(explicitPaths) == 0 && len(prefixes) == 0 && len(patterns) == 0 {
		return true
	}

	for _, p := range explicitPaths {
		if p == path {
			return true
		}
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	for _, cp := range patterns {
		if cp.pattern.MatchString(path) {
			return true
		}
	}

	return false
}

// handleIdempotentRequest handles the idempotent request processing.
func handleIdempotencyCache(c *gin.Context, config IdempotencyConfig, key string) *idempotency.IdempotencyRecord {
	if config.Repository == nil {
		return nil
	}

	record, err := config.Repository.Get(c.Request.Context(), key)
	if err != nil {
		return nil
	}

	return record
}

// handleIdempotentRequest processes the idempotent request.
func handleIdempotentRequest(c *gin.Context, config IdempotencyConfig, path, key string) {
	// Check for cached response.
	if record := handleIdempotencyCache(c, config, key); record != nil {
		c.Header("X-Idempotency-Replay", "true")
		c.Header("Content-Type", record.ContentType)
		c.Header("X-Idempotency-Recorded-At", record.CreatedAt.Format(time.RFC3339))
		if record.TTLRemaining() > 0 {
			c.Header("X-Idempotency-Expires-In", record.TTLRemaining().String())
		}
		c.Data(record.StatusCode, record.ContentType, record.ResponseBody)
		c.Abort()
		return
	}

	// Capture request body.
	captureRequestBody(c)

	// Create response recorder.
	recorder := &responseRecorder{
		ResponseWriter: c.Writer,
		statusCode:     http.StatusOK,
		body:           bytes.NewBuffer(nil),
	}
	c.Writer = recorder

	// Process request.
	c.Next()

	// Store response if successful.
	storeIdempotencyRecord(c, config, path, key, recorder)
}

// captureRequestBody reads and restores the request body.
func captureRequestBody(c *gin.Context) []byte {
	if c.Request.Body == nil {
		return nil
	}
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes
}

// storeIdempotencyRecord stores the response for future replay.
func storeIdempotencyRecord(c *gin.Context, config IdempotencyConfig, path, key string, recorder *responseRecorder) {
	if config.Repository == nil {
		return
	}

	// Store records for successful responses (2xx) and also for 4xx client errors
	// to prevent duplicate processing of invalid requests.
	if recorder.statusCode < 200 || recorder.statusCode >= 500 {
		return
	}

	body := recorder.body.Bytes()
	hash := sha256.Sum256(body)

	record := &idempotency.IdempotencyRecord{
		ID:             key,
		Method:         c.Request.Method,
		Path:           path,
		Hash:           hex.EncodeToString(hash[:]),
		StatusCode:     recorder.statusCode,
		ResponseBody:   body,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(config.TTL),
		ContentType:    recorder.contentType,
		ClientIP:       c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		OrganizationID: GetOrganizationID(c),
	}

	go func() {
		// The request context (c.Request.Context()) is cancelled the moment the
		// response is written and the handler returns. Storage backends that honor
		// context cancellation — notably Turso libSQL over HTTP — would abort the
		// INSERT, silently dropping the idempotency record and defeating replay on
		// the next request. Use a detached context with a bounded timeout so the
		// record is durably persisted independent of the request lifecycle.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := config.Repository.Create(ctx, record); err != nil {
			fmt.Printf("idempotency: failed to store record key=%s: %v\n", key, err)
		}
	}()
}

// responseRecorder captures the response for idempotency storage.
type responseRecorder struct {
	gin.ResponseWriter
	body        *bytes.Buffer
	contentType string
	statusCode  int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	// Capture Content-Type when it's set.
	if ct := r.Header().Get("Content-Type"); ct != "" {
		r.contentType = ct
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// ValidateIdempotencyKey validates an idempotency key format.
// Returns an error description if invalid.
func ValidateIdempotencyKey(key string) error {
	if key == "" {
		return fmt.Errorf("idempotency key is required")
	}
	if len(key) < 8 {
		return fmt.Errorf("idempotency key must be at least 8 characters")
	}
	if len(key) > 128 {
		return fmt.Errorf("idempotency key must not exceed 128 characters")
	}
	// Allow alphanumeric, hyphens, underscores.
	for _, c := range key {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' && c != '_' {
			return fmt.Errorf("idempotency key contains invalid characters")
		}
	}
	return nil
}
