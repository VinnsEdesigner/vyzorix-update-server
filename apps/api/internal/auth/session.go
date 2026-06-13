// Package security provides authentication utilities including session cookie management.
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Session cookie configuration.
const (
	CookieName       = "vyz_session"
	CookieMaxAge     = 86400 // 24 hours in seconds
	CookiePath       = "/"
	EncryptionKeyLen = 32 // AES-256 requires 32 bytes
)

var (
	ErrInvalidCookie    = errors.New("invalid session cookie")
	ErrExpiredCookie    = errors.New("session cookie expired")
	ErrDecryptionFailed = errors.New("cookie decryption failed")
)

// SessionManager handles encrypted session cookies for HttpOnly cookie auth.
type SessionManager struct {
	encryptionKey []byte
}

// NewSessionManager creates a new session manager with the given secret.
// The secret is hashed to produce a 32-byte AES-256 key.
func NewSessionManager(secret string) *SessionManager {
	h := sha256.New()
	h.Write([]byte(secret))
	return &SessionManager{
		encryptionKey: h.Sum(nil),
	}
}

// EncryptOperatorID encrypts an operator ID for storage in a cookie value.
func (sm *SessionManager) EncryptOperatorID(operatorID string) (string, error) {
	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce (12 bytes for GCM)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the operator ID
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(operatorID), nil)

	// Encode to base64 for safe cookie storage
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// DecryptOperatorID decrypts an operator ID from a cookie value.
func (sm *SessionManager) DecryptOperatorID(cookieValue string) (string, error) {
	// Decode from base64
	ciphertext, err := base64.RawURLEncoding.DecodeString(cookieValue)
	if err != nil {
		return "", fmt.Errorf("%w: invalid base64 encoding", ErrDecryptionFailed)
	}

	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("%w: ciphertext too short", ErrDecryptionFailed)
	}

	// Split nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return string(plaintext), nil
}

// CreateSessionCookie creates an HttpOnly session cookie for the given operator.
func (sm *SessionManager) CreateSessionCookie(operatorID string) (*http.Cookie, error) {
	encryptedID, err := sm.EncryptOperatorID(operatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt session: %w", err)
	}

	return &http.Cookie{
		Name:     CookieName,
		Value:    encryptedID,
		Path:     CookiePath,
		MaxAge:   CookieMaxAge,
		HttpOnly: true,
		Secure:   true, // HTTPS only in production
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// CreateSessionCookieWithExpiry creates an HttpOnly session cookie with custom expiry.
func (sm *SessionManager) CreateSessionCookieWithExpiry(operatorID string, maxAge int) (*http.Cookie, error) {
	encryptedID, err := sm.EncryptOperatorID(operatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt session: %w", err)
	}

	return &http.Cookie{
		Name:     CookieName,
		Value:    encryptedID,
		Path:     CookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// ClearSessionCookie creates an expired cookie to clear the session.
func (sm *SessionManager) ClearSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     CookiePath,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

// ExtractSessionFromCookie extracts the operator ID from a session cookie value.
// Returns empty string if the cookie is invalid or decryption fails.
func (sm *SessionManager) ExtractSessionFromCookie(cookieValue string) (string, error) {
	if cookieValue == "" {
		return "", ErrInvalidCookie
	}
	return sm.DecryptOperatorID(cookieValue)
}

// HashOperatorID creates a SHA-256 hash of an operator ID for database lookups.
// This is used to store session hashes in the database for tracking/revocation.
func HashOperatorID(operatorID string) string {
	h := sha256.Sum256([]byte(operatorID))
	return hex.EncodeToString(h[:])
}
