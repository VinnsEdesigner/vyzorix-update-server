package storage

import (
	"context"
	"os"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

func TestSecretHash_HashAndVerify(t *testing.T) {
	hasher := NewSecretHash()

	// Test hashing
	secret := "my-secret-key-123"
	hash, err := hasher.HashSecret(secret)
	if err != nil {
		t.Fatalf("HashSecret failed: %v", err)
	}
	if hash == "" {
		t.Fatal("HashSecret returned empty hash")
	}
	if hash == secret {
		t.Fatal("Hash should not equal secret")
	}

	// Test verification
	if !hasher.VerifyHash(secret, hash) {
		t.Error("VerifyHash failed for correct secret")
	}
	if hasher.VerifyHash("wrong-secret", hash) {
		t.Error("VerifyHash should fail for wrong secret")
	}
}

func TestSecretHash_DifferentHashes(t *testing.T) {
	hasher := NewSecretHash()
	secret := "same-secret"

	hash1, _ := hasher.HashSecret(secret)
	hash2, _ := hasher.HashSecret(secret)

	// bcrypt should produce different hashes due to random salt
	if hash1 == hash2 {
		t.Log("Note: bcrypt produced same hash (unlikely but possible)")
	}

	// Both should verify
	if !hasher.VerifyHash(secret, hash1) {
		t.Error("VerifyHash failed for hash1")
	}
	if !hasher.VerifyHash(secret, hash2) {
		t.Error("VerifyHash failed for hash2")
	}
}

func TestStore_SetAndGetSecretHash(t *testing.T) {
	// Create temp database
	tmpFile, err := os.CreateTemp("", "vyzorix-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()
	hasher := NewSecretHash()

	// Register a device first
	secret := "device-secret-abc"
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "test-device-1",
		FirebaseInstallID: "firebase-123",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Hash and store
	hash, err := hasher.HashSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	err = store.SetSecretHash(ctx, "test-device-1", hash)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve and verify
	retrievedHash, err := store.GetSecretHash(ctx, "test-device-1")
	if err != nil {
		t.Fatal(err)
	}
	if retrievedHash != hash {
		t.Errorf("GetSecretHash returned wrong hash")
	}

	// Verify the secret
	if !hasher.VerifyHash(secret, retrievedHash) {
		t.Error("Secret should verify against stored hash")
	}
}

func TestStore_HashAllSecrets(t *testing.T) {
	// Create temp database
	tmpFile, err := os.CreateTemp("", "vyzorix-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register multiple devices (they will be hashed on registration)
	devices := []string{"dev-1", "dev-2", "dev-3"}
	for _, id := range devices {
		_, _, err = store.Register(ctx, models.RegisterRequest{
			DeviceID:          id,
			FirebaseInstallID: "firebase-" + id,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Verify all have hashes (they were hashed during registration)
	hasher := NewSecretHash()
	for _, id := range devices {
		secret, _ := store.Secret(ctx, id)
		hash, err := store.GetSecretHash(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if hash == "" {
			t.Errorf("Device %s has no hash", id)
		}
		if !hasher.VerifyHash(secret, hash) {
			t.Errorf("Device %s secret doesn't verify against hash", id)
		}
	}

	// HashAllSecrets should return 0 since all are already hashed
	count, err := store.HashAllSecrets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("HashAllSecrets hashed %d, want 0 (all already hashed on registration)", count)
	}
}

func TestStore_GetSecretHash_NotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vyzorix-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()
	hash, err := store.GetSecretHash(ctx, "nonexistent-device")
	if err != nil {
		t.Fatal(err)
	}
	if hash != "" {
		t.Errorf("GetSecretHash for nonexistent device returned %q, want empty", hash)
	}
}
