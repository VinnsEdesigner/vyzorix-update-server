// Package middleware provides HTTP middleware for the Vyzorix API.
package middleware

import (
	"bytes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
)

// ResponseEncryptionConfig holds response encryption configuration.
type ResponseEncryptionConfig struct {
	Enabled bool
}

// DefaultResponseEncryptionConfig returns default config with encryption ENABLED.
func DefaultResponseEncryptionConfig() ResponseEncryptionConfig {
	return ResponseEncryptionConfig{
		Enabled: true, // Mandatory for production
	}
}

// LoadResponseEncryptionConfig loads from environment.
func LoadResponseEncryptionConfig() ResponseEncryptionConfig {
	return ResponseEncryptionConfig{
		Enabled: getEnvBool("RESPONSE_ENCRYPTION_ENABLED", true),
	}
}

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

// EncryptedResponseWriter wraps an http.ResponseWriter to encrypt ALL responses.
type EncryptedResponseWriter struct {
	http.ResponseWriter
	encryptor  *ResponseEncryptor
	buf        *bytes.Buffer
	headerSet  bool
	statusCode int
}

// NewEncryptedResponseWriter creates a response writer that encrypts all responses.
func NewEncryptedResponseWriter(w http.ResponseWriter, enc *ResponseEncryptor) *EncryptedResponseWriter {
	return &EncryptedResponseWriter{
		ResponseWriter: w,
		encryptor:  enc,
		buf:        bytes.NewBuffer(nil),
		statusCode: http.StatusOK,
	}
}

// Header implements http.ResponseWriter.
func (er *EncryptedResponseWriter) Header() http.Header {
	return er.ResponseWriter.Header()
}

// WriteHeader implements http.ResponseWriter.
func (er *EncryptedResponseWriter) WriteHeader(statusCode int) {
	er.statusCode = statusCode
	// Don't write headers yet - wait for Write() to flush encrypted
}

// Write encrypts the data before writing.
// This is MANDATORY - all responses are encrypted.
func (er *EncryptedResponseWriter) Write(b []byte) (int, error) {
	if !er.headerSet {
		er.headerSet = true
		// Set encryption header
		er.ResponseWriter.Header().Set("X-Content-Encryption", "AES-256-GCM")
		er.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
	}
	er.buf.Write(b)
	return len(b), nil
}

// Flush flushes the encrypted response.
func (er *EncryptedResponseWriter) Flush() {
	if !er.headerSet {
		er.headerSet = true
		er.ResponseWriter.Header().Set("X-Content-Encryption", "AES-256-GCM")
		er.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
	}

	if er.buf.Len() == 0 {
		er.ResponseWriter.WriteHeader(er.statusCode)
		return
	}

	// Encrypt the entire response body
	nonce, ciphertext, err := er.encryptor.Encrypt(er.buf.Bytes())
	if err != nil {
		// Can't encrypt - write error as plain text
		er.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		er.ResponseWriter.Write([]byte("encryption failed"))
		return
	}

	// Combine nonce + ciphertext
	encrypted := append(nonce, ciphertext...)

	// Set headers
	er.ResponseWriter.Header().Set("X-Encryption-Nonce", base64.StdEncoding.EncodeToString(nonce))
	er.ResponseWriter.Header().Set("X-Content-Length", fmt.Sprintf("%d", len(ciphertext)))
	er.ResponseWriter.Header().Set("Content-Length", fmt.Sprintf("%d", len(encrypted)))

	er.ResponseWriter.WriteHeader(er.statusCode)
	er.ResponseWriter.Write(encrypted)
}

// MandatoryEncryptionMiddleware returns a middleware that encrypts ALL responses.
// This is REQUIRED for all signed endpoints per the PRD.
func MandatoryEncryptionMiddleware(getKey func(clientID string) ([]byte, bool)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client ID from signature verification (set by RequestSigningMiddleware)
		clientID := c.GetHeader("X-Client-ID")
		if clientID == "" {
			c.Next()
			return
		}

		// Get the encryption key for this client
		key, ok := getKey(clientID)
		if !ok || len(key) != 32 {
			c.Next()
			return
		}

		// Create encryptor
		encryptor, err := NewResponseEncryptor(key)
		if err != nil {
			c.Next()
			return
		}

		// Wrap the response writer
		wrapper := NewEncryptedResponseWriter(c.Writer, encryptor)
		c.Writer = wrapper

		c.Next()

		// Flush encrypted response
		wrapper.Flush()
	}
}

// ResponseEncryptMiddleware is the standard response encryption middleware.
// Use this for endpoints that require encrypted responses.
func ResponseEncryptMiddleware(getKey func(clientID string) ([]byte, bool)) gin.HandlerFunc {
	return MandatoryEncryptionMiddleware(getKey)
}

// JSONEncryptResponse is a helper to encrypt a JSON response.
func JSONEncryptResponse(encryptor *ResponseEncryptor, status int, obj interface{}) (int, []byte, string, error) {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return http.StatusInternalServerError, nil, "", err
	}

	nonce, ciphertext, err := encryptor.Encrypt(jsonData)
	if err != nil {
		return http.StatusInternalServerError, nil, "", err
	}

	nonceB64 := base64.StdEncoding.EncodeToString(nonce)
	ciphertextB64 := base64.StdEncoding.EncodeToString(ciphertext)

	return status, ciphertext, nonceB64, nil
}
