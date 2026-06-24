// Package crypto provides cryptographic utilities.
package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// HTTP request verification errors.
var (
	ErrMissingHeaders = MissingError{HTTPHeader: "X-Vyzorix-Nonce, X-Vyzorix-Timestamp, X-Vyzorix-Signature"}
	ErrBadTimestamp   = &BadFormatError{HTTPHeader: "X-Vyzorix-Timestamp", Message: "must be Unix milliseconds"}
	ErrReplayedNonce  = ReplayedError{HTTPHeader: "X-Vyzorix-Nonce"}
	ErrUnknownDevice  = DeviceNotFoundError{Message: "device secret not found"}
	ErrBadSignature   = SignatureInvalidError{Message: "HMAC signature verification failed"}
)

// MissingError is returned when required HTTP headers are missing.
type MissingError struct {
	HTTPHeader string
}

func (e MissingError) Error() string {
	return "missing required header(s): " + e.HTTPHeader
}

// BadFormatError is returned when an HTTP header has invalid format.
type BadFormatError struct {
	HTTPHeader string
	Message    string
}

func (e BadFormatError) Error() string {
	return "invalid " + e.HTTPHeader + ": " + e.Message
}

// TimestampExpiredError is returned when the timestamp is outside the allowed window.
type TimestampExpiredError struct {
	Window time.Duration
}

func (e TimestampExpiredError) Error() string {
	return "timestamp outside replay window of " + e.Window.String()
}

// ReplayedError is returned when a nonce has been replayed.
type ReplayedError struct {
	HTTPHeader string
}

func (e ReplayedError) Error() string {
	return "replayed " + e.HTTPHeader
}

// DeviceNotFoundError is returned when device secret is not found.
type DeviceNotFoundError struct {
	Message string
}

func (e DeviceNotFoundError) Error() string {
	return e.Message
}

// SignatureInvalidError is returned when HMAC signature verification fails.
type SignatureInvalidError struct {
	Message string
}

func (e SignatureInvalidError) Error() string {
	return e.Message
}

// NonceCache provides thread-safe replay protection for HTTP request nonces.
// It tracks seen nonces and rejects duplicates within a configurable time window.
type NonceCache struct {
	seen   map[string]time.Time
	window time.Duration
	mu     sync.Mutex
}

// NewNonceCache creates a new NonceCache with the specified time window.
func NewNonceCache(window time.Duration) *NonceCache {
	return &NonceCache{
		seen:   make(map[string]time.Time),
		window: window,
	}
}

// Use checks if a nonce has been seen before and marks it as seen if not.
// Returns true if the nonce is new (not a replay), false if it's a replay.
func (c *NonceCache) Use(nonce string, now time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := now.Add(-c.window)
	for k, t := range c.seen {
		if t.Before(cutoff) {
			delete(c.seen, k)
		}
	}

	if _, ok := c.seen[nonce]; ok {
		return false
	}

	c.seen[nonce] = now

	return true
}

// Len returns the current number of nonces in the cache.
func (c *NonceCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.seen)
}

// Clear removes all nonces from the cache.
func (c *NonceCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seen = make(map[string]time.Time)
}

// SecretProvider is a function that returns the secret for a device.
type SecretProvider func(deviceID string) (string, bool)

// Verifier provides HTTP request verification for API endpoint protection.
type Verifier struct {
	Secret SecretProvider
	Nonces *NonceCache
	Window time.Duration
}

// NewVerifier creates a new Verifier with the given secret provider and nonce cache.
func NewVerifier(secret SecretProvider, nonces *NonceCache, window time.Duration) *Verifier {
	return &Verifier{
		Secret: secret,
		Nonces: nonces,
		Window: window,
	}
}

// ReadAndVerifyHTTP reads the request body and verifies the HTTP request.
// It restores the body after reading so it can be read again.
func (v *Verifier) ReadAndVerifyHTTP(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	return body, v.Verify(r.Method, r.URL.RequestURI(), "", body, r.Header)
}

// ReadAndVerify reads the request body and verifies the HTTP request with the device ID.
// It restores the body after reading so it can be read again.
func (v *Verifier) ReadAndVerify(r *http.Request, deviceID string) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	return body, v.Verify(r.Method, r.URL.RequestURI(), deviceID, body, r.Header)
}

// Verify verifies an HTTP request's method, path, deviceID, body, and headers.
// Returns nil if verification succeeds, or an error if verification fails.
func (v *Verifier) Verify(method, path, deviceID string, body []byte, h http.Header) error {
	nonce := h.Get("X-Vyzorix-Nonce")
	ts := h.Get("X-Vyzorix-Timestamp")
	sig := h.Get("X-Vyzorix-Signature")

	if nonce == "" || ts == "" || sig == "" {
		return ErrMissingHeaders
	}

	milli, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return ErrBadTimestamp
	}

	now := time.Now()
	t := time.UnixMilli(milli)

	if t.Before(now.Add(-v.Window)) || t.After(now.Add(v.Window)) {
		return &TimestampExpiredError{Window: v.Window}
	}

	if v.Nonces != nil && !v.Nonces.Use(deviceID+":"+nonce, now) {
		return ErrReplayedNonce
	}

	secret, ok := v.Secret(deviceID)
	if !ok || secret == "" {
		return ErrUnknownDevice
	}

	mac := hmac.New(sha512.New, []byte(secret))
	_, _ = mac.Write([]byte(method + "\n" + path + "\n" + nonce + "\n" + ts + "\n"))
	_, _ = mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return ErrBadSignature
	}

	return nil
}
