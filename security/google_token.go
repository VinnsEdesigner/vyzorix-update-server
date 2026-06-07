package security

import (
	"context"
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidGoogleToken     = errors.New("invalid Google ID token")
	ErrGoogleTokenExpired     = errors.New("Google ID token expired")
	ErrGoogleTokenBadIssuer   = errors.New("Google ID token wrong issuer")
	ErrGoogleTokenBadAudience = errors.New("Google ID token wrong audience")
)

// Google JWKS endpoint
const googleJWKSURL = "https://www.googleapis.com/oauth2/v3/certs"

// Google valid issuers
var googleIssuers = []string{"https://accounts.google.com", "accounts.google.com"}

// GoogleTokenVerifier verifies Google ID tokens using Google's public keys.
type GoogleTokenVerifier struct {
	lastFetch time.Time
	client    *http.Client
	keys      map[string]*rsa.PublicKey
	jwksURL   string
	audience  string
	cacheTTL  time.Duration
	keysMu    sync.RWMutex
}

// NewGoogleTokenVerifier creates a new verifier for Google ID tokens.
func NewGoogleTokenVerifier(audience string) *GoogleTokenVerifier {
	return &GoogleTokenVerifier{
		client:   &http.Client{Timeout: 10 * time.Second},
		jwksURL:  googleJWKSURL,
		keys:     make(map[string]*rsa.PublicKey),
		cacheTTL: 1 * time.Hour,
		audience: audience,
	}
}

// Verify verifies a Google ID token and returns the claims if valid.
func (v *GoogleTokenVerifier) Verify(token string) (*GoogleClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidGoogleToken
	}

	// Parse header to get key ID
	var header struct {
		Kid string `json:"kid"`
		Alg string `json:"alg"`
	}
	headerBytes, err := base64RawURLDecode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid header", ErrInvalidGoogleToken)
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("%w: invalid header", ErrInvalidGoogleToken)
	}

	if header.Alg != "RS256" {
		return nil, fmt.Errorf("%w: expected RS256, got %s", ErrInvalidGoogleToken, header.Alg)
	}

	// Get the public key
	key, err := v.getKey(header.Kid)
	if err != nil {
		return nil, err
	}

	// Verify signature
	if err := v.verifySignature(parts[0]+"."+parts[1], parts[2], key); err != nil {
		return nil, fmt.Errorf("%w: signature verification failed", ErrInvalidGoogleToken)
	}

	// Parse claims
	claimsBytes, err := base64RawURLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid claims", ErrInvalidGoogleToken)
	}
	var claims GoogleClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("%w: invalid claims", ErrInvalidGoogleToken)
	}

	// Verify claims
	if err := v.verifyClaims(&claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

// GoogleClaims represents the claims in a Google ID token.
type GoogleClaims struct {
	Iss           string `json:"iss"`
	Azp           string `json:"azp"`
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Iat           int64  `json:"iat"`
	Exp           int64  `json:"exp"`
	EmailVerified bool   `json:"email_verified"`
}

// getKey retrieves the public key for the given key ID.
func (v *GoogleTokenVerifier) getKey(kid string) (*rsa.PublicKey, error) {
	// Try cache first
	v.keysMu.RLock()
	if key, ok := v.keys[kid]; ok && time.Since(v.lastFetch) < v.cacheTTL {
		v.keysMu.RUnlock()
		return key, nil
	}
	v.keysMu.RUnlock()

	// Refresh keys
	if err := v.refreshKeys(); err != nil {
		// If refresh fails but we have a cached key, use it
		v.keysMu.RLock()
		if key, ok := v.keys[kid]; ok {
			v.keysMu.RUnlock()
			return key, nil
		}
		v.keysMu.RUnlock()
		return nil, err
	}

	v.keysMu.RLock()
	defer v.keysMu.RUnlock()
	if key, ok := v.keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("%w: key not found: %s", ErrInvalidGoogleToken, kid)
}

// refreshKeys fetches the JWKS from Google.
func (v *GoogleTokenVerifier) refreshKeys() error {
	resp, err := v.client.Get(v.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	v.keysMu.Lock()
	defer v.keysMu.Unlock()

	for _, k := range jwks.Keys {
		if k.Kty != "RSA" || k.Use != "sig" {
			continue
		}
		pubKey, err := parseRSAPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		v.keys[k.Kid] = pubKey
	}

	v.lastFetch = time.Now()
	return nil
}

// verifySignature verifies the JWT signature using RSA SHA256.
func (v *GoogleTokenVerifier) verifySignature(signingInput, signature string, key *rsa.PublicKey) error {
	sigBytes, err := base64RawURLDecode(signature)
	if err != nil {
		return err
	}

	hasher := crypto.SHA256.New()
	hasher.Write([]byte(signingInput))

	return rsa.VerifyPKCS1v15(key, crypto.SHA256, hasher.Sum(nil), sigBytes)
}

// verifyClaims verifies the token claims.
func (v *GoogleTokenVerifier) verifyClaims(claims *GoogleClaims) error {
	now := time.Now().Unix()

	// Check expiration
	if claims.Exp < now {
		return ErrGoogleTokenExpired
	}

	// Check issued at (allow 5 minute clock skew)
	if claims.Iat > now+300 {
		return fmt.Errorf("%w: token issued in the future", ErrInvalidGoogleToken)
	}

	// Check issuer
	validIssuer := false
	for _, iss := range googleIssuers {
		if claims.Iss == iss {
			validIssuer = true
			break
		}
	}
	if !validIssuer {
		return ErrGoogleTokenBadIssuer
	}

	// Check audience
	if v.audience != "" && claims.Aud != v.audience {
		return ErrGoogleTokenBadAudience
	}

	return nil
}

// parseRSAPublicKey parses RSA public key from modulus and exponent base64 strings.
func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64RawURLDecode(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64RawURLDecode(eStr)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

// base64RawURLDecode decodes a base64url string without padding.
// Handles the URL-safe alphabet (- for +, _ for /) by converting to standard base64 first.
func base64RawURLDecode(s string) ([]byte, error) {
	// Replace URL-safe chars with standard base64 chars
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	// Add padding if needed for standard base64
	// len % 4 == 0: no padding
	// len % 4 == 1: need 3 padding chars (invalid for base64, but we handle it)
	// len % 4 == 2: need 2 padding chars
	// len % 4 == 3: need 1 padding char
	switch len(s) % 4 {
	case 1:
		s += "==="
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.StdEncoding.DecodeString(s)
}

// DecodeGoogleIDToken is a convenience function that verifies and decodes a Google ID token.
// For production use, prefer GoogleTokenVerifier.Verify() for caching and efficiency.
func DecodeGoogleIDToken(token string, audience string) (*GoogleClaims, error) {
	verifier := NewGoogleTokenVerifier(audience)
	return verifier.Verify(token)
}

// GetGoogleUserInfo fetches user info from Google's userinfo endpoint using an access token.
func GetGoogleUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("userinfo endpoint returned status %d: failed to read body", resp.StatusCode)
		}
		return nil, fmt.Errorf("userinfo endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// GoogleUserInfo represents user info from Google's userinfo endpoint.
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}
