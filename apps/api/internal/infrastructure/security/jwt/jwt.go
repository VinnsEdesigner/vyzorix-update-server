// Package jwt provides JWT token management for authentication.
package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken        = errors.New("invalid token")
	ErrExpiredToken        = errors.New("token has expired")
	ErrTokenIDGeneration   = errors.New("failed to generate token ID")
)

// OperatorClaims are the JWT claims for an authenticated operator.
type OperatorClaims struct {
	jwt.RegisteredClaims
	OperatorID string `json:"oid"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       string `json:"role"`
}

// Manager handles signing and verifying JWT tokens for operator authentication.
type Manager struct {
	issuer string
	secret []byte
	expiry time.Duration
}

// NewManager creates a new JWT manager.
func NewManager(secret string, expiry time.Duration, issuer string) *Manager {
	h := sha256.New()
	h.Write([]byte(secret))

	return &Manager{
		secret: h.Sum(nil),
		expiry: expiry,
		issuer: issuer,
	}
}

// Generate creates a new JWT token for an operator.
func (m *Manager) Generate(operatorID, email, name, role string) (string, time.Time, error) {
	tokenID, err := generateTokenID()
	if err != nil {
		return "", time.Time{}, ErrTokenIDGeneration
	}

	expiresAt := time.Now().Add(m.expiry)
	claims := OperatorClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
			ID:        tokenID,
		},
		OperatorID: operatorID,
		Email:      email,
		Name:       name,
		Role:       role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)

	return signed, expiresAt, err
}

// Verify parses and validates a JWT token, returning the claims if valid.
func (m *Manager) Verify(tokenString string) (*OperatorClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &OperatorClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}

		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*OperatorClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// HashToken stores a SHA-256 hash of a token rather than the raw token.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func generateTokenID() (string, error) {
	b := make([]byte, 16)
	n, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("crypto/rand read failed: %w", err)
	}
	if n != len(b) {
		return "", fmt.Errorf("crypto/rand returned unexpected count: %d", n)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
