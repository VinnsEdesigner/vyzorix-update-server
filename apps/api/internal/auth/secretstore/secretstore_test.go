package secretstore

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestNewSecretStore(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate valid 32-byte key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	keyBase64 := base64.StdEncoding.EncodeToString(key)

	store, err := NewSecretStore(tmpDir, keyBase64)
	if err != nil {
		t.Fatalf("NewSecretStore() error = %v", err)
	}
	if store == nil {
		t.Fatal("NewSecretStore() returned nil store")
	}
}

func TestNewSecretStore_InvalidKeyLength(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 16-byte key (invalid, need 32 bytes)
	shortKey := make([]byte, 16)
	keyBase64 := base64.StdEncoding.EncodeToString(shortKey)

	_, err = NewSecretStore(tmpDir, keyBase64)
	if err == nil {
		t.Error("NewSecretStore() should fail with invalid key length")
	}
}

func TestNewSecretStore_InvalidKeyEncoding(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Invalid base64
	_, err = NewSecretStore(tmpDir, "not-valid-base64!!!")
	if err == nil {
		t.Error("NewSecretStore() should fail with invalid base64")
	}
}

func TestSecretStore_SetAndGet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := newTestStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	deviceID := "device-123"
	secret := "my-super-secret-key-12345678901234567890123456789012"

	// Set secret
	if err := store.Set(deviceID, secret); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Get secret
	got, err := store.Get(deviceID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != secret {
		t.Errorf("Get() = %s, want %s", got, secret)
	}
}

func TestSecretStore_GetNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := newTestStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	_, err = store.Get("non-existent-device")
	if err == nil {
		t.Error("Get() should return error for non-existent device")
	}
}

func TestSecretStore_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := newTestStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	deviceID := "device-to-delete"
	secret := "secret-to-delete"

	// Set then delete
	if err := store.Set(deviceID, secret); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if err := store.Delete(deviceID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Should not exist
	if store.Exists(deviceID) {
		t.Error("Exists() should return false after delete")
	}
}

func TestSecretStore_Exists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := newTestStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	deviceID := "device-exists"
	secret := "secret-exists"

	if store.Exists(deviceID) {
		t.Error("Exists() should return false for new device")
	}

	if err := store.Set(deviceID, secret); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if !store.Exists(deviceID) {
		t.Error("Exists() should return true after Set()")
	}
}

func TestSecretStore_List(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := newTestStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	devices := []string{"device-1", "device-2", "device-3"}
	for _, id := range devices {
		if err := store.Set(id, "secret-for-"+id); err != nil {
			t.Fatalf("Set() error for %s: %v", id, err)
		}
	}

	listed, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(listed) != len(devices) {
		t.Errorf("List() returned %d devices, want %d", len(listed), len(devices))
	}

	// Check all expected devices are present
	for _, id := range devices {
		found := false
		for _, l := range listed {
			if l == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List() missing device %s", id)
		}
	}
}

func TestSecretStore_ListEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := newTestStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	listed, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(listed) != 0 {
		t.Errorf("List() returned %d devices, want 0", len(listed))
	}
}

func TestSecretStore_MultipleSecrets(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := newTestStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	secrets := map[string]string{
		"device-a": "secret-a-1234567890123456789012345678901234567890123",
		"device-b": "secret-b-1234567890123456789012345678901234567890123",
		"device-c": "secret-c-1234567890123456789012345678901234567890123",
	}

	for id, secret := range secrets {
		if err := store.Set(id, secret); err != nil {
			t.Fatalf("Set() error for %s: %v", id, err)
		}
	}

	// Verify each secret
	for id, expected := range secrets {
		got, err := store.Get(id)
		if err != nil {
			t.Fatalf("Get() error for %s: %v", id, err)
		}
		if got != expected {
			t.Errorf("Get(%s) = %s, want %s", id, got, expected)
		}
	}
}

func TestSecretStore_ConcurrentAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := newTestStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			deviceID := "device-concurrent"
			secret := "secret-for-concurrent"
			if err := store.Set(deviceID, secret); err != nil {
				t.Logf("Set error: %v", err)
			}
			if _, err := store.Get(deviceID); err != nil {
				t.Logf("Get error: %v", err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSecretStore_SecretFilePermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "secretstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := newTestStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	deviceID := "device-perm"
	secret := "secret-for-perm"

	if err := store.Set(deviceID, secret); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Check file permissions
	secretPath := filepath.Join(tmpDir, deviceID+".bin")
	info, err := os.Stat(secretPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// File should be readable only by owner (0600)
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("File permissions = %o, want 0600", mode)
	}
}

// Test AESGCMEncryptor

func TestAESGCMEncryptor_EncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	encryptor, err := NewAESGCMEncryptor(key)
	if err != nil {
		t.Fatalf("NewAESGCMEncryptor() error = %v", err)
	}

	plaintext := []byte("my-super-secret-message-1234567890")

	encrypted, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted = %s, want %s", decrypted, plaintext)
	}
}

func TestAESGCMEncryptor_InvalidKeyLength(t *testing.T) {
	shortKey := make([]byte, 16)

	_, err := NewAESGCMEncryptor(shortKey)
	if err == nil {
		t.Error("NewAESGCMEncryptor() should fail with invalid key length")
	}
}

func TestAESGCMEncryptor_DifferentCiphertexts(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	encryptor, err := NewAESGCMEncryptor(key)
	if err != nil {
		t.Fatalf("NewAESGCMEncryptor() error = %v", err)
	}

	plaintext := []byte("same-message-encrypted-twice")

	encrypted1, _ := encryptor.Encrypt(plaintext)
	encrypted2, _ := encryptor.Encrypt(plaintext)

	// Should produce different ciphertexts due to random nonce
	if string(encrypted1) == string(encrypted2) {
		t.Error("Encrypt() should produce different ciphertexts for same plaintext")
	}

	// Both should decrypt to same value
	decrypted1, _ := encryptor.Decrypt(encrypted1)
	decrypted2, _ := encryptor.Decrypt(encrypted2)

	if string(decrypted1) != string(decrypted2) {
		t.Error("Both ciphertexts should decrypt to same plaintext")
	}
}

func TestAESGCMEncryptor_DecryptInvalidCiphertext(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	encryptor, err := NewAESGCMEncryptor(key)
	if err != nil {
		t.Fatalf("NewAESGCMEncryptor() error = %v", err)
	}

	// Too short ciphertext
	_, err = encryptor.Decrypt([]byte("short"))
	if err == nil {
		t.Error("Decrypt() should fail for short ciphertext")
	}

	// Corrupted ciphertext (wrong tag)
	encrypted, _ := encryptor.Encrypt([]byte("test"))
	// Corrupt the last bytes (tag)
	corrupted := make([]byte, len(encrypted))
	copy(corrupted, encrypted)
	corrupted[len(corrupted)-1] ^= 0xFF

	_, err = encryptor.Decrypt(corrupted)
	if err == nil {
		t.Error("Decrypt() should fail for corrupted ciphertext")
	}
}

// Helper function to create a test store
func newTestStore(baseDir string) (*SecretStore, error) {
	// Use a fixed key for testing
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	keyBase64 := base64.StdEncoding.EncodeToString(key)
	return NewSecretStore(baseDir, keyBase64)
}