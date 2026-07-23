// Package middleware provides HTTP middleware.
package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	"github.com/gin-gonic/gin"
)

// Request signing errors.
var (
	ErrMissingHeaders         = errors.New("missing required signature headers")
	ErrInvalidTimestamp       = errors.New("invalid timestamp format")
	ErrTimestampOutsideWindow = errors.New("timestamp outside allowed window")
	ErrReplayDetected         = errors.New("replay detected")
	ErrInvalidSignature       = errors.New("invalid signature")
	ErrInvalidEncryptedBody   = errors.New("invalid encrypted body")
	ErrUnknownClient          = errors.New("unknown or inactive client")
	ErrDecryptionFailed       = errors.New("decryption failed")
		ErrClientIDMismatch       = errors.New("client ID does not match device IMEI in request path")
)

// SigningConfig holds request signing configuration.
type SigningConfig struct {
	TimestampWindow       time.Duration
	MaxCacheSize          int
	GracePeriod           time.Duration
	Enabled               bool
	AllowUnsignedFallback bool
}

// DefaultSigningConfig returns the default signing configuration.
// Request signing is ENABLED by default for production infraauth.
func DefaultSigningConfig() SigningConfig {
	return SigningConfig{
		Enabled:               true,
		TimestampWindow:       5 * time.Minute,
		MaxCacheSize:          100000,
		GracePeriod:           24 * time.Hour,
		AllowUnsignedFallback: false,
	}
}

// LoadSigningConfig loads signing configuration from environment variables.
func LoadSigningConfig() SigningConfig {
	cfg := DefaultSigningConfig()
	cfg.Enabled = getEnvBool("REQUEST_SIGNING_ENABLED", true)
	cfg.AllowUnsignedFallback = getEnvBool("ALLOW_UNSIGNED_FALLBACK", false)

	return cfg
}

// SignatureVerifier verifies request signatures.
type SignatureVerifier struct {
	ClientSecret func(clientID string) (string, bool)
	Now          func() time.Time
	replayCache  *ReplayCache
	Config       SigningConfig
}

// ReplayCache is a thread-safe cache for preventing replay attacks.
// It stores signatures seen within the timestamp window.
// Uses a map with O(1) lookups and periodic cleanup for O(1) amortized insertions.
type ReplayCache struct {
	seen       map[string]time.Time
	window     time.Duration
	maxSize    int
	mu         sync.Mutex
	lastCleanup time.Time
}

// NewReplayCache creates a new replay cache with the given window and max size.
func NewReplayCache(window time.Duration, maxSize int) *ReplayCache {
	return &ReplayCache{
		seen:        make(map[string]time.Time),
		window:      window,
		maxSize:     maxSize,
		lastCleanup: time.Now(),
	}
}

// Use checks if a signature has been seen before. If not, it marks it as seen.
// Returns true if the signature is new (not a replay), false if it's a replay.
// Uses O(1) map operations with periodic cleanup to avoid O(N) eviction spikes.
func (c *ReplayCache) Use(signature string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-c.window)

	// Check if signature was already used (O(1) lookup).
	if _, exists := c.seen[signature]; exists {
		return false // Replay detected.
	}

	// Periodic cleanup: only clean when cache is getting full or every minute.
	// This spreads the O(N) cleanup cost over many operations instead of doing it on every call.
	if len(c.seen) >= c.maxSize || now.Sub(c.lastCleanup) > time.Minute {
		c.cleanupLocked(cutoff)
		c.lastCleanup = now

		// If still at capacity after cleanup, remove oldest entries directly.
		// Use simple truncation instead of iterating all entries.
		if len(c.seen) >= c.maxSize {
			evictCount := c.maxSize / 10 // Remove 10%.
			removed := 0
			for sig := range c.seen {
				if removed >= evictCount {
					break
				}
				delete(c.seen, sig)
				removed++
			}
		}
	}

	// Mark this signature as seen (O(1) insertion).
	c.seen[signature] = now

	return true
}

// cleanupLocked removes expired entries. Caller must hold c.mu.
func (c *ReplayCache) cleanupLocked(cutoff time.Time) {
	for sig, t := range c.seen {
		if t.Before(cutoff) {
			delete(c.seen, sig)
		}
	}
}

// Size returns the current cache size.
func (c *ReplayCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.seen)
}

// Clear empties the cache.
func (c *ReplayCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seen = make(map[string]time.Time)
}

// NewSignatureVerifier creates a new signature verifier.
func NewSignatureVerifier(config SigningConfig, clientSecret func(clientID string) (string, bool)) *SignatureVerifier {
	return &SignatureVerifier{
		Config:       config,
		ClientSecret: clientSecret,
		Now:          time.Now,
		replayCache:  NewReplayCache(config.TimestampWindow, config.MaxCacheSize),
	}
}

// ReadAndVerifyHTTP reads the body and verifies the request signature.
// Returns the decrypted body if successful.
func (v *SignatureVerifier) ReadAndVerifyHTTP(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	verifiedBody, err := v.Verify(r.Method, r.URL.RequestURI(), body, r.Header)
	if err != nil {
		return nil, err
	}

	// Restore verified body (decrypted if encrypted) for downstream handlers.
	r.Body = io.NopCloser(bytes.NewReader(verifiedBody))

	return verifiedBody, nil
}

// Verify verifies a request signature.
// Returns the body used for verification (decrypted if encrypted, original otherwise).
func (v *SignatureVerifier) Verify(method, path string, body []byte, h http.Header) ([]byte, error) {
	// Extract headers.
	clientID := h.Get("X-Client-ID")
	timestamp := h.Get("X-Timestamp")
	signature := h.Get("X-Signature")
	encryptedBody := h.Get("X-Encrypted-Body")

	// Check required headers.
	if clientID == "" || timestamp == "" || signature == "" {
		return nil, ErrMissingHeaders
	}

	// Parse and validate timestamp.
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, ErrInvalidTimestamp
	}

	now := v.Now()
	requestTime := time.Unix(ts, 0)

	// Check timestamp window.
	if requestTime.Before(now.Add(-v.Config.TimestampWindow)) || requestTime.After(now.Add(v.Config.TimestampWindow)) {
		return nil, ErrTimestampOutsideWindow
	}

	// Get client secret.
	secret, ok := v.ClientSecret(clientID)
	if !ok || secret == "" {
		return nil, ErrUnknownClient
	}

	// For requests with encrypted body, decrypt first.
	var bodyToVerify []byte

	if encryptedBody != "" {
		decrypted, err := v.decryptBody(encryptedBody, secret)
		if err != nil {
			return nil, ErrInvalidEncryptedBody
		}

		bodyToVerify = decrypted
	} else {
		// Use original body if not encrypted.
		bodyToVerify = body
	}

	// Verify signature.
	expectedSig := v.computeSignature(method, path, timestamp, bodyToVerify, secret)
	if !hmac.Equal([]byte(expectedSig), []byte(signature)) {
		return nil, ErrInvalidSignature
	}

	// Check replay cache (only if not using encrypted body replay protection).
	sigKey := clientID + ":" + signature
	if !v.replayCache.Use(sigKey) {
		return nil, ErrReplayDetected
	}

	return bodyToVerify, nil
}

// computeSignature computes the HMAC-SHA512 signature.
// Format: t={timestamp},v1={hmac_signature}.
func (v *SignatureVerifier) computeSignature(method, path, timestamp string, body []byte, secret string) string {
	// Compute SHA512 of body.
	bodyHash := sha512.Sum512(body)
	bodyHashHex := hex.EncodeToString(bodyHash[:])

	// Build string to sign: method\npath\ntimestamp\nbodyHash.
	stringToSign := method + "\n" + path + "\n" + timestamp + "\n" + bodyHashHex

	// Compute HMAC-SHA512.
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	signature := hex.EncodeToString(mac.Sum(nil))

	return "t=" + timestamp + ",v1=" + signature
}

// decryptBody decrypts an AES-256-GCM encrypted body.
// The encrypted body is base64 encoded, with the nonce prepended.
func (v *SignatureVerifier) decryptBody(encryptedBase64, secret string) ([]byte, error) {
	// Decode from base64.
	encrypted, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// Use shared crypto package for decryption.
	return crypto.DecryptAES256GCM(secret, encrypted)
}

// EncryptBody encrypts a body using AES-256-GCM.
// Returns base64 encoded ciphertext with prepended nonce.
func (v *SignatureVerifier) EncryptBody(plaintext []byte, secret string) (string, error) {
	// Use shared crypto package for encryption.
	ciphertext, err := crypto.EncryptAES256GCM(secret, plaintext)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// SignRequest signs a request for client-side use.
// Returns the signature headers to add to the request.
func SignRequest(method, path string, body []byte, clientID, clientSecret string) (http.Header, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Encrypt body using shared crypto package.
	var encryptedBody string

	if len(body) > 0 {
		ciphertext, err := crypto.EncryptAES256GCM(clientSecret, body)
		if err != nil {
			return nil, err
		}

		encryptedBody = base64.StdEncoding.EncodeToString(ciphertext)
	}

	// Compute signature.
	bodyHash := sha512.Sum512(body)
	stringToSign := method + "\n" + path + "\n" + timestamp + "\n" + hex.EncodeToString(bodyHash[:])
	mac := hmac.New(sha512.New, []byte(clientSecret))
	mac.Write([]byte(stringToSign))
	signature := "t=" + timestamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))

	// Build headers.
	headers := http.Header{}
	headers.Set("X-Client-ID", clientID)
	headers.Set("X-Timestamp", timestamp)
	headers.Set("X-Signature", signature)

	if encryptedBody != "" {
		headers.Set("X-Encrypted-Body", encryptedBody)
	}

	return headers, nil
}

// IsSigningRequiredPath checks if a path requires request signing.
// Per PRD: Only truly public endpoints are exempt; everything else requires signing.
func IsSigningRequiredPath(path string) bool {
	// Exempt ONLY these truly public endpoints:.
	// Health checks - no auth required, just liveness probes.
	if path == "/health/live" ||
		path == "/health/ready" ||
		path == "/healthz" ||
		path == "/health" {
		return false
	}

	// Static assets - no sensitive data.
	if strings.HasPrefix(path, "/assets/") ||
		path == "/favicon.ico" {
		return false
	}

	// Auth endpoints use their own session/cookie auth mechanism.
	// These are exempt because they establish the session used for signing.
	if strings.HasPrefix(path, "/v1/auth/") {
		return false
	}

	// Device status is public - used by devices to check in.
	if path == "/v1/device/status" {
		return false
	}

	// API info - no sensitive data.
	if path == "/api/v1/version" || path == "/api/v1/changelog" {
		return false
	}

	// ALL other endpoints REQUIRE request signing.
	// This includes:.
	// - /metrics (Prometheus metrics - sensitive operational data).
	// - /health/secure (security status - requires signing).
	// - All /v1/device/* endpoints except register.
	// - All /v1/command/* endpoints.
	// - All admin endpoints.
	return true
}

// RequestSigningMiddleware returns a Gin middleware that verifies request signatures.
// for protected API endpoints. It checks if signing is enabled and if the path requires.
// signature verification before processing.
func RequestSigningMiddleware(verifier *SignatureVerifier) func(c *gin.Context) {
	return func(c *gin.Context) {
		// Skip if signing is disabled.
		if !verifier.Config.Enabled {
			c.Next()
			return
		}

		// Skip if path doesn't require signing.
		if !IsSigningRequiredPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract clientID from header for IMEI verification after successful auth.
		clientID := c.GetHeader("X-Client-ID")

		// Read body and verify signature.
		body, err := verifier.ReadAndVerifyHTTP(c.Request)
		if err != nil {
			// Handle errors based on type.
			switch {
			case errors.Is(err, ErrMissingHeaders):
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":   "SIGN_001",
					"message": "Missing required signature headers",
				})
			case errors.Is(err, ErrInvalidTimestamp):
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":   "SIGN_002",
					"message": "Invalid timestamp format",
				})
			case errors.Is(err, ErrTimestampOutsideWindow):
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":   "SIGN_003",
					"message": "Request timestamp outside allowed window",
				})
			case errors.Is(err, ErrUnknownClient):
				// Return generic error to prevent client ID enumeration.
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":   "SIGN_003",
					"message": "Signature verification failed",
				})
			case errors.Is(err, ErrInvalidSignature):
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":   "SIGN_003",
					"message": "Signature verification failed",
				})
			case errors.Is(err, ErrReplayDetected):
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":   "SIGN_006",
					"message": "Replay detected",
				})
			case errors.Is(err, ErrInvalidEncryptedBody):
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "SIGN_007",
					"message": "Invalid encrypted body",
				})
			default:
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":   "SIGN_001",
					"message": "Signature verification failed",
				})
			}

			return
		}
		// This prevents requests where the clientID header is missing/empty.
		imei := c.Param("imei")
		if imei != "" && clientID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "SIGN_008",
				"message": "Client ID header is required for device endpoints",
			})
			return
		}

		// This prevents a compromised device from forging requests for another device's IMEI.
		if imei != "" && clientID != imei {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "SIGN_009",
				"message": "Client ID does not match device IMEI in request path",
			})
			return
		}

		// Store the decrypted body in context for handlers.
		c.Set("signed_body", body)
		c.Next()
	}
}
