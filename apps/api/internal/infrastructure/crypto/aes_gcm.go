package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"io"
)

const (
	AES256KeySize = 32
	NonceSize     = 12
)

var (
	ErrCiphertextTooShort  = errors.New("ciphertext too short")
	ErrDecryptionFailed    = errors.New("decryption failed")
	ErrKeyDerivationFailed = errors.New("key derivation failed")
	ErrEncryptionFailed    = errors.New("encryption failed")
)

// DeriveKey derives a 32-byte key from any secret using SHA-512.
func DeriveKey(secret string) []byte {
	key := sha512.Sum512([]byte(secret))
	return key[:AES256KeySize]
}

// EncryptAES256GCM encrypts plaintext using AES-256-GCM.
// Returns nonce || ciphertext.
func EncryptAES256GCM(secret string, plaintext []byte) ([]byte, error) {
	key := DeriveKey(secret)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrKeyDerivationFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, ErrEncryptionFailed
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptAES256GCM decrypts ciphertext using AES-256-GCM.
// ciphertext must be nonce || actual_ciphertext.
func DecryptAES256GCM(secret string, ciphertext []byte) ([]byte, error) {
	key := DeriveKey(secret)
	return DecryptAES256GCMWithKey(key, ciphertext)
}

// EncryptAES256GCMWithKey encrypts plaintext using AES-256-GCM with a pre-derived key.
// Returns nonce || ciphertext.
func EncryptAES256GCMWithKey(key []byte, plaintext []byte) ([]byte, error) {
	if len(key) != AES256KeySize {
		return nil, ErrKeyDerivationFailed
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrKeyDerivationFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, ErrEncryptionFailed
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptAES256GCMWithKey decrypts ciphertext using AES-256-GCM with a pre-derived key.
// ciphertext must be nonce || actual_ciphertext.
func DecryptAES256GCMWithKey(key []byte, ciphertext []byte) ([]byte, error) {
	if len(key) != AES256KeySize {
		return nil, ErrKeyDerivationFailed
	}

	if len(ciphertext) < NonceSize {
		return nil, ErrCiphertextTooShort
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrKeyDerivationFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	nonce := ciphertext[:NonceSize]

	plaintext, err := gcm.Open(nil, nonce, ciphertext[NonceSize:], nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// NewGCMCipher creates a new GCM cipher from a 32-byte key.
// This is useful when you already have a derived key and want to reuse it.
func NewGCMCipher(key []byte) (cipher.AEAD, error) {
	if len(key) != AES256KeySize {
		return nil, ErrKeyDerivationFailed
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrKeyDerivationFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	return gcm, nil
}

// GenerateNonce generates a random nonce of the specified size.
func GenerateNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, ErrEncryptionFailed
	}

	return nonce, nil
}
