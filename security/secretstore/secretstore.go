package secretstore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// SecretStore manages encrypted storage of per-device command secrets.
// Per DEVICE_REGISTRATION.md §6.1: secrets stored in data/secrets/<deviceId>.bin
// encrypted with AES-GCM using a master key from VYZORIX_SECRET_MASTER_KEY env.
type SecretStore struct {
	encryptor *AESGCMEncryptor
	baseDir   string
	masterKey []byte
	mu        sync.RWMutex
}

// NewSecretStore creates a new SecretStore.
// baseDir is the directory where secrets will be stored.
// masterKeyBase64 is the base64-encoded AES-256 key (32 bytes).
func NewSecretStore(baseDir string, masterKeyBase64 string) (*SecretStore, error) {
	masterKey, err := base64.StdEncoding.DecodeString(masterKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid master key encoding: %w", err)
	}

	if len(masterKey) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes (256 bits), got %d", len(masterKey))
	}

	encryptor, err := NewAESGCMEncryptor(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	// Ensure base directory exists with proper permissions
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create secrets directory: %w", err)
	}

	return &SecretStore{
		baseDir:   baseDir,
		masterKey: masterKey,
		encryptor: encryptor,
	}, nil
}

// Get retrieves the raw command secret for a device.
// Returns error if secret is not found.
func (s *SecretStore) Get(deviceID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	secret, err := s.getSecret(deviceID)
	if err != nil {
		return "", fmt.Errorf("failed to get secret for device %s: %w", deviceID, err)
	}
	return secret, nil
}

// Set stores the raw command secret for a device, encrypting it first.
func (s *SecretStore) Set(deviceID string, secret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	encrypted, err := s.encryptor.Encrypt([]byte(secret))
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	secretPath := s.secretPath(deviceID)
	if err := os.WriteFile(secretPath, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write secret file: %w", err)
	}

	return nil
}

// Delete removes the stored secret for a device.
func (s *SecretStore) Delete(deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	secretPath := s.secretPath(deviceID)
	if err := os.Remove(secretPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// Exists checks if a secret exists for the given device.
func (s *SecretStore) Exists(deviceID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, err := os.Stat(s.secretPath(deviceID))
	return err == nil
}

// List returns all device IDs that have stored secrets.
func (s *SecretStore) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets directory: %w", err)
	}

	var deviceIDs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".bin" {
			deviceID := entry.Name()[:len(entry.Name())-4] // Remove .bin extension
			deviceIDs = append(deviceIDs, deviceID)
		}
	}
	return deviceIDs, nil
}

func (s *SecretStore) secretPath(deviceID string) string {
	return filepath.Join(s.baseDir, deviceID+".bin")
}

func (s *SecretStore) getSecret(deviceID string) (string, error) {
	secretPath := s.secretPath(deviceID)

	encrypted, err := os.ReadFile(secretPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound
		}
		return "", err
	}

	decrypted, err := s.encryptor.Decrypt(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}

	return string(decrypted), nil
}

// ErrNotFound indicates the requested secret does not exist.
var ErrNotFound = errors.New("secret not found")

// Errors
var (
	ErrInvalidKey    = errors.New("invalid encryption key")
	ErrDecryption    = errors.New("decryption failed")
	ErrEncryption    = errors.New("encryption failed")
	ErrCorruptedData = errors.New("corrupted encrypted data")
)

// AESGCMEncryptor provides AES-GCM encryption/decryption.
type AESGCMEncryptor struct {
	gcm cipher.AEAD
	key []byte
}

// NewAESGCMEncryptor creates a new AES-GCM encryptor with the given 256-bit key.
func NewAESGCMEncryptor(key []byte) (*AESGCMEncryptor, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AESGCMEncryptor{
		key: key,
		gcm: gcm,
	}, nil
}

// NonceSize is the size of the GCM nonce.
const NonceSize = 12 // AES-GCM standard nonce size

// Encrypt encrypts plaintext using AES-GCM with a random nonce.
// Returns ciphertext with nonce prepended.
func (e *AESGCMEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("%w: failed to generate nonce: %v", ErrEncryption, err)
	}

	ciphertext := e.gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext that was encrypted with Encrypt.
// Expects nonce prepended to ciphertext.
func (e *AESGCMEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	nonceSize := NonceSize
	if len(ciphertext) < nonceSize {
		return nil, ErrCorruptedData
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryption, err)
	}

	return plaintext, nil
}
