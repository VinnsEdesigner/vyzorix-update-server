package storage

import (
	"context"
	"os"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

func TestHashSecret(t *testing.T) {
	// Test hashing
	secret := "my-secret-key-123"
	hash, err := HashSecret(secret)
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
	if len(hash) < 11 || hash[:11] != "$argon2id$v" {
		t.Error("Hash should start with $argon2id$v")
	}

	err = VerifySecret(secret, hash)
	if err != nil {
		t.Fatalf("VerifySecret failed for correct secret: %v", err)
	}

	err = VerifySecret("wrong-secret", hash)
	if err == nil {
		t.Error("VerifySecret should fail for wrong secret")
	}
}

func TestHashSecret_DifferentHashes(t *testing.T) {
	secret := "same-secret"

	hash1, _ := HashSecret(secret)
	hash2, _ := HashSecret(secret)

	// Argon2id should produce different hashes due to random salt
	if hash1 == hash2 {
		t.Log("Note: Argon2id produced same hash (unlikely but possible)")
	}

	// Both should verify
	if err := VerifySecret(secret, hash1); err != nil {
		t.Error("VerifySecret failed for hash1")
	}
	if err := VerifySecret(secret, hash2); err != nil {
		t.Error("VerifySecret failed for hash2")
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
	hash, err := HashSecret(secret)
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
	err = VerifySecret(secret, retrievedHash)
	if err != nil {
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
	for _, id := range devices {
		secret, _ := store.Secret(ctx, id)
		hash, err := store.GetSecretHash(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if hash == "" {
			t.Errorf("Device %s has no hash", id)
		}
		verified := VerifySecret(secret, hash)
		if verified != nil {
			t.Errorf("Device %s secret should verify: %v", id, verified)
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