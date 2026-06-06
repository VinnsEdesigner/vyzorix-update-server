package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// Test NonceCache

func TestNonceCache_NewNonce(t *testing.T) {
	cache := NewNonceCache(5 * time.Minute)
	now := time.Now()

	if !cache.Use("nonce-1", now) {
		t.Error("expected new nonce to be accepted")
	}
}

func TestNonceCache_ReuseRejected(t *testing.T) {
	cache := NewNonceCache(5 * time.Minute)
	now := time.Now()

	if !cache.Use("nonce-1", now) {
		t.Fatal("first use should succeed")
	}
	if cache.Use("nonce-1", now) {
		t.Error("expected replay to be rejected")
	}
}

func TestNonceCache_ExpiredNonceCleanup(t *testing.T) {
	cache := NewNonceCache(2 * time.Second)
	oldTime := time.Now().Add(-5 * time.Second)

	// Use a nonce with old timestamp
	cache.seen["old-nonce"] = oldTime

	// New nonce at current time should work
	now := time.Now()
	if !cache.Use("new-nonce", now) {
		t.Error("expected new nonce to be accepted")
	}

	// Old nonce should be cleaned up and accepted
	if !cache.Use("old-nonce", now) {
		t.Error("expected expired nonce to be cleaned up and accepted")
	}
}

func TestNonceCache_DifferentDeviceIDs(t *testing.T) {
	cache := NewNonceCache(5 * time.Minute)
	now := time.Now()

	// Same nonce but different device IDs should both work
	if !cache.Use("device-a:nonce-1", now) {
		t.Fatal("first use should succeed")
	}
	if !cache.Use("device-b:nonce-1", now) {
		t.Fatal("same nonce different device should succeed")
	}
}

// Test Verifier

func TestVerifier_MissingHeaders(t *testing.T) {
	v := Verifier{
		Secret: func(id string) (string, bool) { return "secret", true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	tests := []struct {
		name   string
		header http.Header
	}{
		{"missing nonce", http.Header{"X-Vyzorix-Timestamp": {"123"}, "X-Vyzorix-Signature": {"sig"}}},
		{"missing timestamp", http.Header{"X-Vyzorix-Nonce": {"n"}, "X-Vyzorix-Signature": {"sig"}}},
		{"missing signature", http.Header{"X-Vyzorix-Nonce": {"n"}, "X-Vyzorix-Timestamp": {"123"}}},
		{"all missing", http.Header{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Verify("POST", "/test", "device1", nil, tt.header)
			if err != ErrMissing {
				t.Errorf("expected ErrMissing, got %v", err)
			}
		})
	}
}

func TestVerifier_BadTimestampFormat(t *testing.T) {
	v := Verifier{
		Secret: func(id string) (string, bool) { return "secret", true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	header := http.Header{
		"X-Vyzorix-Nonce":     {"nonce-1"},
		"X-Vyzorix-Timestamp": {"not-a-number"},
		"X-Vyzorix-Signature": {"sig"},
	}

	err := v.Verify("POST", "/test", "device1", nil, header)
	if err == nil || err.Error() != "bad timestamp" {
		t.Errorf("expected 'bad timestamp' error, got %v", err)
	}
}

func TestVerifier_StaleTimestamp(t *testing.T) {
	v := Verifier{
		Secret: func(id string) (string, bool) { return "secret", true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	stale := time.Now().Add(-10 * time.Minute).UnixMilli()
	header := http.Header{
		"X-Vyzorix-Nonce":     {"nonce-1"},
		"X-Vyzorix-Timestamp": {strconv.FormatInt(stale, 10)},
		"X-Vyzorix-Signature": {"sig"},
	}

	err := v.Verify("POST", "/test", "device1", nil, header)
	if err == nil || err.Error() != "timestamp outside replay window" {
		t.Errorf("expected 'timestamp outside replay window', got %v", err)
	}
}

func TestVerifier_FutureTimestamp(t *testing.T) {
	v := Verifier{
		Secret: func(id string) (string, bool) { return "secret", true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	// 10 minutes in the future
	future := time.Now().Add(10 * time.Minute).UnixMilli()
	header := http.Header{
		"X-Vyzorix-Nonce":     {"nonce-1"},
		"X-Vyzorix-Timestamp": {strconv.FormatInt(future, 10)},
		"X-Vyzorix-Signature": {"sig"},
	}

	err := v.Verify("POST", "/test", "device1", nil, header)
	if err == nil || err.Error() != "timestamp outside replay window" {
		t.Errorf("expected 'timestamp outside replay window', got %v", err)
	}
}

func TestVerifier_UnknownDevice(t *testing.T) {
	v := Verifier{
		Secret: func(id string) (string, bool) { return "", false },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	header := http.Header{
		"X-Vyzorix-Nonce":     {"nonce-1"},
		"X-Vyzorix-Timestamp": {ts},
		"X-Vyzorix-Signature": {"sig"},
	}

	err := v.Verify("POST", "/test", "unknown-device", nil, header)
	if err == nil || err.Error() != "unknown device secret" {
		t.Errorf("expected 'unknown device secret', got %v", err)
	}
}

func TestVerifier_BadSignature(t *testing.T) {
	secret := "test-secret"
	v := Verifier{
		Secret: func(id string) (string, bool) { return secret, true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	header := http.Header{
		"X-Vyzorix-Nonce":     {"nonce-1"},
		"X-Vyzorix-Timestamp": {ts},
		"X-Vyzorix-Signature": {"bad-signature"},
	}

	err := v.Verify("POST", "/test", "device1", nil, header)
	if err == nil || err.Error() != "bad signature" {
		t.Errorf("expected 'bad signature', got %v", err)
	}
}

func TestVerifier_ValidSignature(t *testing.T) {
	secret := "test-secret"
	v := Verifier{
		Secret: func(id string) (string, bool) { return secret, true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	nonce := "nonce-1"
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	body := []byte(`{"test":"data"}`)

	// Build valid signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("POST\n/test\n" + nonce + "\n" + ts + "\n"))
	mac.Write(body)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	header := http.Header{
		"X-Vyzorix-Nonce":     {nonce},
		"X-Vyzorix-Timestamp": {ts},
		"X-Vyzorix-Signature": {sig},
	}

	err := v.Verify("POST", "/test", "device1", body, header)
	if err != nil {
		t.Errorf("expected valid signature to pass, got %v", err)
	}
}

func TestVerifier_ValidSignatureWithEmptyBody(t *testing.T) {
	secret := "test-secret"
	v := Verifier{
		Secret: func(id string) (string, bool) { return secret, true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	nonce := "nonce-1"
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Build valid signature for empty body
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("GET\n/test\n" + nonce + "\n" + ts + "\n"))
	mac.Write(nil)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	header := http.Header{
		"X-Vyzorix-Nonce":     {nonce},
		"X-Vyzorix-Timestamp": {ts},
		"X-Vyzorix-Signature": {sig},
	}

	err := v.Verify("GET", "/test", "device1", nil, header)
	if err != nil {
		t.Errorf("expected valid signature to pass, got %v", err)
	}
}

func TestVerifier_ReplayRejected(t *testing.T) {
	secret := "test-secret"
	v := Verifier{
		Secret: func(id string) (string, bool) { return secret, true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	nonce := "unique-nonce"
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Build valid signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("POST\n/test\n" + nonce + "\n" + ts + "\n"))
	mac.Write(nil)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	header := http.Header{
		"X-Vyzorix-Nonce":     {nonce},
		"X-Vyzorix-Timestamp": {ts},
		"X-Vyzorix-Signature": {sig},
	}

	// First call should succeed
	err := v.Verify("POST", "/test", "device1", nil, header)
	if err != nil {
		t.Fatalf("first call should succeed, got %v", err)
	}

	// Second call with same nonce should fail
	err = v.Verify("POST", "/test", "device1", nil, header)
	if err == nil || err.Error() != "replayed nonce" {
		t.Errorf("expected 'replayed nonce', got %v", err)
	}
}

func TestVerifier_DifferentDeviceAllowsSameNonce(t *testing.T) {
	secret := "test-secret"
	v := Verifier{
		Secret: func(id string) (string, bool) { return secret, true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	nonce := "shared-nonce"
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("POST\n/test\n" + nonce + "\n" + ts + "\n"))
	mac.Write(nil)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	header := http.Header{
		"X-Vyzorix-Nonce":     {nonce},
		"X-Vyzorix-Timestamp": {ts},
		"X-Vyzorix-Signature": {sig},
	}

	// First device
	err := v.Verify("POST", "/test", "device-a", nil, header)
	if err != nil {
		t.Fatalf("device-a should succeed, got %v", err)
	}

	// Second device with same nonce should also succeed
	err = v.Verify("POST", "/test", "device-b", nil, header)
	if err != nil {
		t.Errorf("device-b should succeed (different device), got %v", err)
	}
}

// Test ReadAndVerifyHTTP

func TestReadAndVerifyHTTP_Success(t *testing.T) {
	secret := "test-secret"
	v := Verifier{
		Secret: func(id string) (string, bool) { return secret, true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	body := []byte(`{"test":"data"}`)
	nonce := "nonce-1"
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("POST\n/test\n" + nonce + "\n" + ts + "\n"))
	mac.Write(body)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("X-Vyzorix-Nonce", nonce)
	req.Header.Set("X-Vyzorix-Timestamp", ts)
	req.Header.Set("X-Vyzorix-Signature", sig)

	resultBody, err := v.ReadAndVerifyHTTP(req)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !bytes.Equal(resultBody, body) {
		t.Errorf("body mismatch: got %s, want %s", resultBody, body)
	}
}

func TestReadAndVerifyHTTP_RequestBodyConsumed(t *testing.T) {
	secret := "test-secret"
	v := Verifier{
		Secret: func(id string) (string, bool) { return secret, true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	body := []byte(`{"test":"data"}`)
	nonce := "nonce-1"
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("POST\n/test\n" + nonce + "\n" + ts + "\n"))
	mac.Write(body)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("X-Vyzorix-Nonce", nonce)
	req.Header.Set("X-Vyzorix-Timestamp", ts)
	req.Header.Set("X-Vyzorix-Signature", sig)

	// First call
	v.ReadAndVerifyHTTP(req)

	// Second call should also work (body is reset)
	nonce = "nonce-2"
	ts = strconv.FormatInt(time.Now().UnixMilli(), 10)

	mac = hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("POST\n/test\n" + nonce + "\n" + ts + "\n"))
	mac.Write(body)
	sig = base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("X-Vyzorix-Nonce", nonce)
	req.Header.Set("X-Vyzorix-Timestamp", ts)
	req.Header.Set("X-Vyzorix-Signature", sig)

	_, err := v.ReadAndVerifyHTTP(req)
	if err != nil {
		t.Fatalf("expected second call to succeed, got %v", err)
	}
}

// Edge cases

func TestVerifier_EmptySecret(t *testing.T) {
	v := Verifier{
		Secret: func(id string) (string, bool) { return "", true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	header := http.Header{
		"X-Vyzorix-Nonce":     {"nonce-1"},
		"X-Vyzorix-Timestamp": {ts},
		"X-Vyzorix-Signature": {"sig"},
	}

	err := v.Verify("POST", "/test", "device1", nil, header)
	if err == nil || err.Error() != "unknown device secret" {
		t.Errorf("expected 'unknown device secret' for empty secret, got %v", err)
	}
}

func TestVerifier_WithinWindowBoundary(t *testing.T) {
	secret := "test-secret"
	v := Verifier{
		Secret: func(id string) (string, bool) { return secret, true },
		Nonces: NewNonceCache(5 * time.Minute),
		Window: 5 * time.Minute,
	}

	// Exactly at the window boundary (4 minutes ago)
	withinWindow := time.Now().Add(-4 * time.Minute).UnixMilli()
	nonce := "nonce-1"
	ts := strconv.FormatInt(withinWindow, 10)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("POST\n/test\n" + nonce + "\n" + ts + "\n"))
	mac.Write(nil)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	header := http.Header{
		"X-Vyzorix-Nonce":     {nonce},
		"X-Vyzorix-Timestamp": {ts},
		"X-Vyzorix-Signature": {sig},
	}

	err := v.Verify("POST", "/test", "device1", nil, header)
	if err != nil {
		t.Errorf("expected within window to pass, got %v", err)
	}
}

func TestVerifier_NilNonces(t *testing.T) {
	secret := "test-secret"
	v := Verifier{
		Secret: func(id string) (string, bool) { return secret, true },
		Nonces: nil, // No nonce checking
		Window: 5 * time.Minute,
	}

	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	nonce := "nonce-1"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("POST\n/test\n" + nonce + "\n" + ts + "\n"))
	mac.Write(nil)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	header := http.Header{
		"X-Vyzorix-Nonce":     {nonce},
		"X-Vyzorix-Timestamp": {ts},
		"X-Vyzorix-Signature": {sig},
	}

	// Should succeed without nonce cache
	err := v.Verify("POST", "/test", "device1", nil, header)
	if err != nil {
		t.Errorf("expected nil nonces to pass, got %v", err)
	}

	// Same nonce again should also work (no nonce cache)
	err = v.Verify("POST", "/test", "device1", nil, header)
	if err != nil {
		t.Errorf("expected replay to work without nonce cache, got %v", err)
	}
}

// Helper for manual testing
func _signRequest(secretHex, method, path, body, nonce, ts string) string {
	key, _ := hex.DecodeString(secretHex)
	canonical := method + "\n" + path + "\n" + nonce + "\n" + ts + "\n"
	if body != "" {
		canonical += body
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(canonical))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func BenchmarkNonceCache(b *testing.B) {
	cache := NewNonceCache(5 * time.Minute)
	now := time.Now()

	for i := 0; i < b.N; i++ {
		nonce := fmt.Sprintf("nonce-%d", i)
		cache.Use(nonce, now)
	}
}