package security

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

// Test base64RawURLDecode
func TestBase64RawURLDecode(t *testing.T) {
	// Test inputs - URL-safe base64 without padding (function adds padding internally)
	// "a" -> YQ (4 chars, divisible by 4, no padding needed)
	// "ab" -> YWI (4 chars, divisible by 4, no padding needed) 
	// "abc" -> YWJj (4 chars, divisible by 4, no padding needed)
	// "abcd" -> YWJjZA== (8 chars with padding)
	// "Hello" -> SGVsbG8= (7 chars, needs 1 padding char)
	// "HelloWorld" -> SGVsbG8gV29ybGQ= (17 chars, needs 3 padding chars)
	tests := []struct {
		name      string
		input     string
		expected  string
		shouldErr bool
	}{
		{"single char", "YQ", "a", false},
		{"two chars", "YWI", "ab", false},
		{"three chars", "YWJj", "abc", false},
		{"hello", "SGVsbG8=", "Hello", false},
		{"world", "V29ybGQ=", "World", false},
		{"abcd with padding", "YWJjZA==", "abcd", false},
		{"helloworld with padding", "SGVsbG9Xb3JsZA==", "HelloWorld", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := base64RawURLDecode(tt.input)
			if tt.shouldErr && err == nil {
				t.Error("expected error")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldErr && string(result) != tt.expected {
				t.Errorf("got %q, want %q", string(result), tt.expected)
			}
		})
	}
}

func TestBase64RawURLDecode_InvalidInput(t *testing.T) {
	_, err := base64RawURLDecode("!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

// Test GoogleClaims struct
func TestGoogleClaims(t *testing.T) {
	claims := GoogleClaims{
		Iss:           "https://accounts.google.com",
		Azp:           "client-id.apps.googleusercontent.com",
		Aud:           "client-id.apps.googleusercontent.com",
		Sub:           "google-user-123",
		Email:         "user@example.com",
		EmailVerified: true,
		Name:          "Test User",
		Picture:       "https://example.com/photo.jpg",
		Iat:           time.Now().Unix(),
		Exp:           time.Now().Add(time.Hour).Unix(),
	}

	if claims.Iss != "https://accounts.google.com" {
		t.Errorf("Iss = %s, want https://accounts.google.com", claims.Iss)
	}
	if claims.Sub != "google-user-123" {
		t.Errorf("Sub = %s, want google-user-123", claims.Sub)
	}
	if claims.Email != "user@example.com" {
		t.Errorf("Email = %s, want user@example.com", claims.Email)
	}
	if !claims.EmailVerified {
		t.Error("EmailVerified should be true")
	}
}

// Test parseRSAPublicKey (we need valid RSA components)
func TestParseRSAPublicKey(t *testing.T) {
	// This is a test RSA key pair for testing purposes
	// Modulus (n) and exponent (e) in base64url format
	// This is NOT a real Google key - just for testing parsing logic
	
	// Simple test with minimal values
	nStr := base64.RawURLEncoding.EncodeToString([]byte{0x00, 0x80}) // Small modulus
	eStr := base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}) // 65537 exponent
	
	key, err := parseRSAPublicKey(nStr, eStr)
	if err != nil {
		t.Fatalf("parseRSAPublicKey failed: %v", err)
	}
	if key.E != 65537 {
		t.Errorf("E = %d, want 65537", key.E)
	}
}

// Test JWT structure validation (without actual Google verification)
func TestJWTPayloadExtraction(t *testing.T) {
	// Create a fake JWT payload for testing structure
	payload := map[string]interface{}{
		"iss":    "https://accounts.google.com",
		"azp":    "test-client-id",
		"aud":    "test-client-id",
		"sub":    "12345",
		"email":  "test@example.com",
		"email_verified": true,
		"name":   "Test User",
		"picture": "https://example.com/pic.jpg",
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	payloadBytes, _ := json.Marshal(payload)
	encoded := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Decode it back
	decoded, err := base64RawURLDecode(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	var claims GoogleClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Email = %s, want test@example.com", claims.Email)
	}
	if claims.Sub != "12345" {
		t.Errorf("Sub = %s, want 12345", claims.Sub)
	}
}

// Test claims verification logic
func TestVerifyClaimsLogic(t *testing.T) {
	verifier := &GoogleTokenVerifier{
		audience: "test-client-id",
	}

	tests := []struct {
		name      string
		claims    GoogleClaims
		wantError bool
		errType   error
	}{
		{
			name: "valid claims",
			claims: GoogleClaims{
				Iss:           "https://accounts.google.com",
				Aud:           "test-client-id",
				Sub:           "12345",
				Email:         "test@example.com",
				EmailVerified: true,
				Iat:           time.Now().Unix() - 60,
				Exp:           time.Now().Unix() + 3600,
			},
			wantError: false,
		},
		{
			name: "expired token",
			claims: GoogleClaims{
				Iss:           "https://accounts.google.com",
				Aud:           "test-client-id",
				Sub:           "12345",
				Email:         "test@example.com",
				EmailVerified: true,
				Iat:           time.Now().Unix() - 7200,
				Exp:           time.Now().Unix() - 3600, // Expired 1 hour ago
			},
			wantError: true,
			errType:   ErrGoogleTokenExpired,
		},
		{
			name: "wrong issuer",
			claims: GoogleClaims{
				Iss:           "https://evil.com",
				Aud:           "test-client-id",
				Sub:           "12345",
				Email:         "test@example.com",
				EmailVerified: true,
				Iat:           time.Now().Unix() - 60,
				Exp:           time.Now().Unix() + 3600,
			},
			wantError: true,
			errType:   ErrGoogleTokenBadIssuer,
		},
		{
			name: "wrong audience",
			claims: GoogleClaims{
				Iss:           "https://accounts.google.com",
				Aud:           "wrong-client-id",
				Sub:           "12345",
				Email:         "test@example.com",
				EmailVerified: true,
				Iat:           time.Now().Unix() - 60,
				Exp:           time.Now().Unix() + 3600,
			},
			wantError: true,
			errType:   ErrGoogleTokenBadAudience,
		},
		{
			name: "future iat (clock skew)",
			claims: GoogleClaims{
				Iss:           "https://accounts.google.com",
				Aud:           "test-client-id",
				Sub:           "12345",
				Email:         "test@example.com",
				EmailVerified: true,
				Iat:           time.Now().Unix() + 600, // 10 minutes in future
				Exp:           time.Now().Unix() + 3600,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier.verifyClaims(&tt.claims)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.errType != nil && err != tt.errType {
				t.Errorf("expected %v, got %v", tt.errType, err)
			}
		})
	}
}

// Test token with no audience requirement
func TestVerifyClaims_NoAudienceCheck(t *testing.T) {
	verifier := &GoogleTokenVerifier{
		audience: "", // No audience check
	}

	claims := GoogleClaims{
		Iss:           "https://accounts.google.com",
		Aud:           "any-client-id-works",
		Sub:           "12345",
		Email:         "test@example.com",
		EmailVerified: true,
		Iat:           time.Now().Unix() - 60,
		Exp:           time.Now().Unix() + 3600,
	}

	err := verifier.verifyClaims(&claims)
	if err != nil {
		t.Errorf("expected no error with empty audience, got %v", err)
	}
}

// Test valid Google issuers
func TestGoogleIssuers(t *testing.T) {
	validIssuers := []string{"https://accounts.google.com", "accounts.google.com"}

	for _, iss := range validIssuers {
		found := false
		for _, v := range googleIssuers {
			if v == iss {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("issuer %s should be valid", iss)
		}
	}
}

// Test token with different algorithms
func TestVerify_InvalidAlgorithm(t *testing.T) {
	verifier := NewGoogleTokenVerifier("test-audience")

	// Create a mock token with a non-RS256 algorithm
	header := map[string]string{
		"alg": "RS384", // Wrong algorithm
		"kid": "test-key",
		"typ": "JWT",
	}
	headerBytes, _ := json.Marshal(header)
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)

	payload := map[string]string{
		"sub": "12345",
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	token := encodedHeader + "." + encodedPayload + ".signature"

	_, err := verifier.Verify(token)
	if err == nil {
		t.Error("expected error for non-RS256 algorithm")
	}
}

// Test malformed JWT
func TestVerify_MalformedJWT(t *testing.T) {
	verifier := NewGoogleTokenVerifier("test-audience")

	tests := []struct {
		name  string
		token string
	}{
		{"no dots", "notajwt"},
		{"one dot", "header.payload"},
		{"four parts", "a.b.c.d"},
		{"empty parts", ".."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := verifier.Verify(tt.token)
			if err == nil {
				t.Error("expected error for malformed JWT")
			}
		})
	}
}

// Test GoogleUserInfo struct
func TestGoogleUserInfo(t *testing.T) {
	info := GoogleUserInfo{
		ID:            "google-id-123",
		Email:         "user@example.com",
		VerifiedEmail: true,
		Name:          "Test User",
		Picture:       "https://example.com/photo.jpg",
	}

	if info.ID != "google-id-123" {
		t.Errorf("ID = %s, want google-id-123", info.ID)
	}
	if info.Email != "user@example.com" {
		t.Errorf("Email = %s, want user@example.com", info.Email)
	}
	if !info.VerifiedEmail {
		t.Error("VerifiedEmail should be true")
	}
}

func BenchmarkGoogleTokenVerify(b *testing.B) {
	verifier := NewGoogleTokenVerifier("test-audience")
	
	// Create a mock token for benchmarking
	header := map[string]interface{}{
		"alg": "RS256",
		"kid": "test-key",
		"typ": "JWT",
	}
	headerBytes, _ := json.Marshal(header)
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)

	payload := map[string]interface{}{
		"iss":    "https://accounts.google.com",
		"azp":    "test-audience",
		"aud":    "test-audience",
		"sub":    "12345",
		"email":  "test@example.com",
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	token := encodedHeader + "." + encodedPayload + ".fake-signature"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		verifier.Verify(token) //nolint:errcheck // benchmark only
	}
}

func TestParseRSAPublicKey_InvalidBase64(t *testing.T) {
	_, err := parseRSAPublicKey("!!!invalid", "AQAB")
	if err == nil {
		t.Error("expected error for invalid modulus base64")
	}
}