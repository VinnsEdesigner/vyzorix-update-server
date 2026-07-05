// Package session provides session cookie management.
package session

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	CookieName       = "vyz_session"
	CookieMaxAge     = 86400
	CookiePath       = "/"
	EncryptionKeyLen = 32
)

var (
	ErrInvalidCookie    = errors.New("invalid session cookie")
	ErrExpiredCookie    = errors.New("session cookie expired")
	ErrDecryptionFailed = errors.New("cookie decryption failed")
)

// Repository defines the interface for session persistence.
type Repository interface {
	FindByID(ctx context.Context, id string) (*Session, error)
	FindByOperatorID(ctx context.Context, operatorID string) ([]*Session, error)
	Create(ctx context.Context, s *Session) error
	Delete(ctx context.Context, id string) error
	DeleteByOperatorID(ctx context.Context, operatorID string) error
	DeleteExpired(ctx context.Context) (int, error)
	Extend(ctx context.Context, id string, newExpiry time.Time) error
	RevokeAllOperatorSessions(ctx context.Context, operatorID string) error
}

// Session represents an authentication session.
type Session struct {
	ID         string
	OperatorID string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	IPAddress  string
	UserAgent  string
}

// Manager handles encrypted session cookies for HttpOnly cookie auth.
type Manager struct {
	sessionRepo   Repository
	encryptionKey []byte
}

// NewManager creates a new session manager with the given secret.
func NewManager(secret string) *Manager {
	h := sha512.New()
	h.Write([]byte(secret))
	fullHash := h.Sum(nil)
	key := make([]byte, EncryptionKeyLen)
	copy(key, fullHash)

	return &Manager{
		encryptionKey: key,
	}
}

// SetRepository sets the session repository for session management.
func (sm *Manager) SetRepository(repo Repository) {
	sm.sessionRepo = repo
}

// EncryptOperatorID encrypts an operator ID for storage in a cookie value.
func (sm *Manager) EncryptOperatorID(operatorID string) (string, error) {
	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(operatorID), nil)

	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// DecryptOperatorID decrypts an operator ID from a cookie value.
func (sm *Manager) DecryptOperatorID(cookieValue string) (string, error) {
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

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return string(plaintext), nil
}

// CreateCookie creates an HttpOnly session cookie for the given operator.
func (sm *Manager) CreateCookie(operatorID string) (*http.Cookie, error) {
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
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// CreateCookieWithExpiry creates an HttpOnly session cookie with custom expiry.
func (sm *Manager) CreateCookieWithExpiry(operatorID string, maxAge int) (*http.Cookie, error) {
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

// ClearCookie creates an expired cookie to clear the session.
func (sm *Manager) ClearCookie() *http.Cookie {
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

// ExtractFromCookie extracts the operator ID from a session cookie value.
func (sm *Manager) ExtractFromCookie(cookieValue string) (string, error) {
	if cookieValue == "" {
		return "", ErrInvalidCookie
	}

	return sm.DecryptOperatorID(cookieValue)
}

// HashOperatorID creates a SHA-512 hash of an operator ID for database lookups.
// Uses SHA-512 for stronger security.
func HashOperatorID(operatorID string) string {
	h := sha512.Sum512([]byte(operatorID))
	return hex.EncodeToString(h[:])
}

// ListActiveSessions returns all active (non-expired) sessions for an operator.
func (sm *Manager) ListActiveSessions(ctx context.Context, operatorID string) ([]*Session, error) {
	if sm.sessionRepo == nil {
		return nil, errors.New("session repository not configured")
	}

	sessions, err := sm.sessionRepo.FindByOperatorID(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	// Filter out expired sessions
	active := make([]*Session, 0, len(sessions))
	now := time.Now()
	for _, s := range sessions {
		if s.ExpiresAt.After(now) {
			active = append(active, s)
		}
	}

	return active, nil
}

// RevokeSession revokes a specific session by ID.
func (sm *Manager) RevokeSession(ctx context.Context, sessionID string) error {
	if sm.sessionRepo == nil {
		return errors.New("session repository not configured")
	}

	return sm.sessionRepo.Delete(ctx, sessionID)
}
