package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	secret := "test-secret-key"
	plaintext := []byte("Hello, World!")

	ciphertext, err := EncryptAES256GCM(secret, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := DecryptAES256GCM(secret, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted != plaintext: got %x, want %x", decrypted, plaintext)
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	ciphertext, _ := EncryptAES256GCM("secret1", []byte("test"))

	_, err := DecryptAES256GCM("secret2", ciphertext)
	if err == nil {
		t.Error("Expected error with wrong key, got nil")
	}
}

func TestDecryptTruncatedCiphertext(t *testing.T) {
	_, err := DecryptAES256GCM("secret", []byte("short"))
	if err != ErrCiphertextTooShort {
		t.Errorf("Expected ErrCiphertextTooShort, got %v", err)
	}
}

func TestEncryptProducesDifferentOutput(t *testing.T) {
	secret := "test-secret"
	plaintext := []byte("same message")

	cipher1, _ := EncryptAES256GCM(secret, plaintext)
	cipher2, _ := EncryptAES256GCM(secret, plaintext)

	// Due to random nonce, same plaintext should produce different ciphertext.
	if bytes.Equal(cipher1, cipher2) {
		t.Error("Expected different ciphertext for same plaintext, got equal")
	}

	// But both should decrypt to the same plaintext.
	decrypted1, _ := DecryptAES256GCM(secret, cipher1)
	decrypted2, _ := DecryptAES256GCM(secret, cipher2)

	if !bytes.Equal(decrypted1, decrypted2) {
		t.Error("Expected same plaintext after decryption")
	}
}

func TestDeriveKey(t *testing.T) {
	key1 := DeriveKey("secret")
	key2 := DeriveKey("secret")
	key3 := DeriveKey("different")

	// Same secret should produce same key.
	if !bytes.Equal(key1, key2) {
		t.Error("Expected same key for same secret")
	}

	// Different secret should produce different key.
	if bytes.Equal(key1, key3) {
		t.Error("Expected different key for different secret")
	}

	// Key should be 32 bytes (AES256).
	if len(key1) != AES256KeySize {
		t.Errorf("Expected key length %d, got %d", AES256KeySize, len(key1))
	}
}

func TestNonceSize(t *testing.T) {
	if NonceSize != 12 {
		t.Errorf("Expected NonceSize 12 (96 bits for GCM), got %d", NonceSize)
	}
}

func TestEmptyPlaintext(t *testing.T) {
	secret := "test-secret"
	plaintext := []byte{}

	ciphertext, err := EncryptAES256GCM(secret, plaintext)
	if err != nil {
		t.Fatalf("Encrypt empty plaintext failed: %v", err)
	}

	decrypted, err := DecryptAES256GCM(secret, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt empty ciphertext failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted empty plaintext doesn't match original")
	}
}

func TestLargePlaintext(t *testing.T) {
	secret := "test-secret"
	// 1MB of data.
	plaintext := make([]byte, 1024*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	ciphertext, err := EncryptAES256GCM(secret, plaintext)
	if err != nil {
		t.Fatalf("Encrypt large plaintext failed: %v", err)
	}

	decrypted, err := DecryptAES256GCM(secret, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt large ciphertext failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted large plaintext doesn't match original")
	}
}
