// Package security provides cryptographic verification utilities.
package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	ErrMissing         = errors.New("missing HMAC headers")
	ErrBadTimestamp    = errors.New("bad timestamp")
	ErrTimestampWindow = errors.New("timestamp outside replay window")
	ErrReplayedNonce   = errors.New("replayed nonce")
	ErrUnknownDevice   = errors.New("unknown device secret")
	ErrBadSignature    = errors.New("bad signature")
)

type NonceCache struct {
	seen   map[string]time.Time
	window time.Duration
	mu     sync.Mutex
}

func NewNonceCache(window time.Duration) *NonceCache {
	return &NonceCache{seen: map[string]time.Time{}, window: window}
}

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

type Verifier struct {
	Secret func(deviceID string) (string, bool)
	Nonces *NonceCache
	Window time.Duration
}

func (v Verifier) ReadAndVerifyHTTP(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, v.Verify(r.Method, r.URL.RequestURI(), "", body, r.Header)
}

func (v Verifier) ReadAndVerify(r *http.Request, deviceID string) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, v.Verify(r.Method, r.URL.RequestURI(), deviceID, body, r.Header)
}

func (v Verifier) Verify(method, path, deviceID string, body []byte, h http.Header) error {
	nonce, ts, sig := h.Get("X-Vyzorix-Nonce"), h.Get("X-Vyzorix-Timestamp"), h.Get("X-Vyzorix-Signature")
	if nonce == "" || ts == "" || sig == "" {
		return ErrMissing
	}
	milli, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return ErrBadTimestamp
	}
	now := time.Now()
	t := time.UnixMilli(milli)
	if t.Before(now.Add(-v.Window)) || t.After(now.Add(v.Window)) {
		return ErrTimestampWindow
	}
	if v.Nonces != nil && !v.Nonces.Use(deviceID+":"+nonce, now) {
		return ErrReplayedNonce
	}
	secret, ok := v.Secret(deviceID)
	if !ok || secret == "" {
		return ErrUnknownDevice
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(method + "\n" + path + "\n" + nonce + "\n" + ts + "\n"))
	_, _ = mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return ErrBadSignature
	}
	return nil
}
