// Package middleware provides HTTP middleware.
package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/idempotency"
	"github.com/gin-gonic/gin"
)

// IdempotencyConfig holds configuration for idempotency middleware.
type IdempotencyConfig struct {
	Repository idempotency.Repository
	HeaderName string
	Paths      []string
	TTL        time.Duration
}

// DefaultIdempotencyConfig returns the default idempotency configuration.
func DefaultIdempotencyConfig() IdempotencyConfig {
	return IdempotencyConfig{
		TTL:        24 * time.Hour,
		HeaderName: "Idempotency-Key",
		Paths:      []string{"/v1/device/inbox"},
	}
}

// GinIdempotency returns a Gin middleware that handles idempotency keys.
// This implements Bug 45 fix - enterprise-grade idempotency.
func GinIdempotency(config IdempotencyConfig) func(c *gin.Context) {
	if config.TTL == 0 {
		config.TTL = 24 * time.Hour
	}
	if config.HeaderName == "" {
		config.HeaderName = "Idempotency-Key"
	}

	return func(c *gin.Context) {
		// Only apply to configured paths
		pathSupported := false
		for _, p := range config.Paths {
			if c.Request.URL.Path == p {
				pathSupported = true
				break
			}
		}
		if !pathSupported {
			c.Next()
			return
		}

		// Only apply to POST/PATCH/PUT methods
		if c.Request.Method != http.MethodPost &&
			c.Request.Method != http.MethodPatch &&
			c.Request.Method != http.MethodPut {
			c.Next()
			return
		}

		// Get idempotency key from header
		idempotencyKey := c.GetHeader(config.HeaderName)
		if idempotencyKey == "" {
			// No idempotency key provided - proceed normally (not an error)
			c.Next()
			return
		}

		// Validate idempotency key format
		if len(idempotencyKey) < 8 || len(idempotencyKey) > 128 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"code":    "INVALID_IDEMPOTENCY_KEY",
				"message": "Idempotency-Key must be between 8 and 128 characters",
			})
			return
		}

		// Check if we have a cached response
		if config.Repository != nil {
			record, err := config.Repository.Get(c.Request.Context(), idempotencyKey)
			if err != nil {
				// Log error but continue without idempotency
				c.Next()
				return
			}

			if record != nil {
				// Return cached response
				c.Header("X-Idempotency-Replay", "true")
				c.Header("Content-Type", record.ContentType)
				c.Header("X-Idempotency-Recorded-At", record.CreatedAt.Format(time.RFC3339))
				c.Data(record.StatusCode, record.ContentType, record.ResponseBody)
				c.Abort()
				return
			}
		}

		// Capture request body for hashing
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create response recorder
		recorder := &responseRecorder{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = recorder

		// Process request
		c.Next()

		// Store response if successful (2xx status codes)
		if config.Repository != nil && recorder.statusCode >= 200 && recorder.statusCode < 300 {
			body := recorder.body.Bytes()
			hash := sha256.Sum256(body)

			record := &idempotency.IdempotencyRecord{
				ID:           idempotencyKey,
				Method:       c.Request.Method,
				Path:         c.Request.URL.Path,
				Hash:         hex.EncodeToString(hash[:]),
				StatusCode:   recorder.statusCode,
				ResponseBody: body,
				CreatedAt:    time.Now(),
				ExpiresAt:    time.Now().Add(config.TTL),
				ContentType:  recorder.contentType,
				ClientIP:     c.ClientIP(),
				UserAgent:    c.Request.UserAgent(),
			}

			// Store asynchronously to not delay response
			go func() {
				_ = config.Repository.Create(c.Request.Context(), record)
			}()
		}
	}
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
	// Capture Content-Type when it's set
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
	// Allow alphanumeric, hyphens, underscores
	for _, c := range key {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' && c != '_' {
			return fmt.Errorf("idempotency key contains invalid characters")
		}
	}
	return nil
}
