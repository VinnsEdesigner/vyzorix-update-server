// Package middleware provides HTTP middleware for the Vyzorix API.
package middleware

import (
	"bytes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
)

// ResponseEncryptor handles response encryption.
type ResponseEncryptor struct {
	cipher    cipher.AEAD
	nonceSize int
}

// NewResponseEncryptor creates a new response encryptor with the given key.
func NewResponseEncryptor(key []byte) (*ResponseEncryptor, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes for AES-256")
	}

	// Use shared crypto package to create the cipher.
	gcm, err := crypto.NewGCMCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	return &ResponseEncryptor{
		cipher:    gcm,
		nonceSize: crypto.NonceSize,
	}, nil
}

// Encrypt encrypts the data using AES-256-GCM.
func (re *ResponseEncryptor) Encrypt(plaintext []byte) (nonce, ciphertext []byte, err error) {
	// Use shared crypto package for nonce generation.
	nonce, err = crypto.GenerateNonce(re.nonceSize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext = re.cipher.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

// EncryptToBase64 encrypts and returns base64-encoded result.
func (re *ResponseEncryptor) EncryptToBase64(plaintext []byte) (string, string, error) {
	nonce, ciphertext, err := re.Encrypt(plaintext)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(nonce),
		base64.StdEncoding.EncodeToString(ciphertext), nil
}

// EncryptedResponseWriter wraps an http.ResponseWriter to encrypt responses.
type EncryptedResponseWriter struct {
	http.ResponseWriter
	encryptor *ResponseEncryptor
	written   bool
	buf       *bytes.Buffer
}

// NewEncryptedResponseWriter creates a response writer that encrypts the body.
func NewEncryptedResponseWriter(w http.ResponseWriter, enc *ResponseEncryptor) *EncryptedResponseWriter {
	return &EncryptedResponseWriter{
		ResponseWriter: w,
		encryptor: enc,
		buf:       bytes.NewBuffer(nil),
	}
}

// Write encrypts the data before writing.
func (er *EncryptedResponseWriter) Write(b []byte) (int, error) {
	if !er.written {
		er.written = true
		// Set encryption headers.
		er.Header().Set("X-Content-Encryption", "AES-256-GCM")
	}
	er.buf.Write(b)
	return len(b), nil
}

// WriteEncrypted flushes the encrypted buffer to the response.
func (er *EncryptedResponseWriter) WriteEncrypted() error {
	if !er.written {
		return nil
	}

	nonce, ciphertext, err := er.encryptor.Encrypt(er.buf.Bytes())
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	er.Header().Set("X-Encryption-Nonce", base64.StdEncoding.EncodeToString(nonce))
	er.Header().Set("Content-Length", fmt.Sprintf("%d", len(ciphertext)))
	er.buf.Reset()

	_, err = er.ResponseWriter.Write(ciphertext)
	return err
}

// ShouldEncrypt checks if the request wants encrypted response.
func ShouldEncrypt(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("X-Encrypt-Response"), "true")
}
