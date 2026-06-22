// Package request_signer provides request signing utilities for API authentication.
package request_signer

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Signer signs API requests for authenticated clients.
type Signer struct {
	clientID     string
	clientSecret []byte
}

// New creates a new request signer.
func New(clientID, clientSecret string) (*Signer, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	if len(clientSecret) < 32 {
		return nil, fmt.Errorf("client secret must be at least 32 bytes")
	}

	return &Signer{
		clientID:     clientID,
		clientSecret: []byte(clientSecret),
	}, nil
}

// Sign creates a signed request with HMAC-SHA256.
func (rs *Signer) Sign(method, path string, body []byte, timestamp int64) (string, error) {
	encryptedBody, err := rs.encryptBody(body)
	if err != nil {
		return "", fmt.Errorf("body encryption failed: %w", err)
	}

	bodyHash := sha256.Sum256(encryptedBody)
	stringToSign := fmt.Sprintf("%s\n%s\n%d\n%x",
		strings.ToUpper(method),
		path,
		timestamp,
		bodyHash,
	)

	h := hmac.New(sha256.New, rs.clientSecret)
	h.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("t=%d,v1=%s", timestamp, signature), nil
}

// Headers returns the headers needed for a signed request.
func (rs *Signer) Headers(method, path string, body []byte) (map[string]string, error) {
	timestamp := time.Now().Unix()

	signature, err := rs.Sign(method, path, body, timestamp)
	if err != nil {
		return nil, err
	}

	encryptedBody := base64.StdEncoding.EncodeToString(body)

	return map[string]string{
		"X-Client-ID":      rs.clientID,
		"X-Timestamp":      fmt.Sprintf("%d", timestamp),
		"X-Signature":      signature,
		"X-Encrypted-Body": encryptedBody,
		"Content-Type":     "application/json",
	}, nil
}

func (rs *Signer) encryptBody(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, nil
	}

	block, err := aes.NewCipher(rs.clientSecret[:32])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptBody decrypts the response body.
func (rs *Signer) DecryptBody(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, nil
	}

	block, err := aes.NewCipher(rs.clientSecret[:32])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
