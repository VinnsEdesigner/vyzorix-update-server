package security

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	if sm == nil {
		t.Fatal("NewSessionManager returned nil")
	}
	if len(sm.encryptionKey) != 32 {
		t.Errorf("encryptionKey length = %d, want 32", len(sm.encryptionKey))
	}
}

func TestNewSessionManager_Deterministic(t *testing.T) {
	sm1 := NewSessionManager("same-secret")
	sm2 := NewSessionManager("same-secret")
	if len(sm1.encryptionKey) != len(sm2.encryptionKey) {
		t.Error("same secret should produce same key length")
	}
	for i := range sm1.encryptionKey {
		if sm1.encryptionKey[i] != sm2.encryptionKey[i] {
			t.Error("same secret should produce identical key")
			break
		}
	}
}

func TestNewSessionManager_DifferentSecrets(t *testing.T) {
	sm1 := NewSessionManager("secret-one")
	sm2 := NewSessionManager("secret-two")
	for i := range sm1.encryptionKey {
		if sm1.encryptionKey[i] == sm2.encryptionKey[i] {
			t.Error("different secrets should produce different keys")
			break
		}
	}
}

func TestEncryptOperatorID(t *testing.T) {
	sm := NewSessionManager("test-secret-key-for-encryption")
	operatorID := "op_1234567890abcdef"

	encrypted, err := sm.EncryptOperatorID(operatorID)
	if err != nil {
		t.Fatalf("EncryptOperatorID() error = %v", err)
	}

	if encrypted == "" {
		t.Error("encrypted value should not be empty")
	}

	if encrypted == operatorID {
		t.Error("encrypted value should differ from original")
	}

	// Encrypted value should be base64 encoded (RawURLEncoding)
	if strings.ContainsAny(encrypted, "+/=") {
		t.Error("RawURLEncoding should not contain +, /, or =")
	}
}

func TestEncryptOperatorID_Uniqueness(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	operatorID := "op_same_id"

	// Encrypt same operator ID multiple times - should get different ciphertexts
	// due to random nonce
	encrypted1, err := sm.EncryptOperatorID(operatorID)
	if err != nil {
		t.Fatalf("EncryptOperatorID() error = %v", err)
	}

	encrypted2, err := sm.EncryptOperatorID(operatorID)
	if err != nil {
		t.Fatalf("EncryptOperatorID() error = %v", err)
	}

	if encrypted1 == encrypted2 {
		t.Error("encrypting same ID twice should produce different ciphertexts (random nonce)")
	}

	// But both should decrypt to the same value
	decrypted1, err := sm.DecryptOperatorID(encrypted1)
	if err != nil {
		t.Fatalf("DecryptOperatorID() error = %v", err)
	}

	decrypted2, err := sm.DecryptOperatorID(encrypted2)
	if err != nil {
		t.Fatalf("DecryptOperatorID() error = %v", err)
	}

	if decrypted1 != operatorID || decrypted2 != operatorID {
		t.Errorf("both should decrypt to %q", operatorID)
	}
}

func TestDecryptOperatorID(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	originalID := "op_abcdef1234567890"

	encrypted, err := sm.EncryptOperatorID(originalID)
	if err != nil {
		t.Fatalf("EncryptOperatorID() error = %v", err)
	}

	decrypted, err := sm.DecryptOperatorID(encrypted)
	if err != nil {
		t.Fatalf("DecryptOperatorID() error = %v", err)
	}

	if decrypted != originalID {
		t.Errorf("decrypted = %q, want %q", decrypted, originalID)
	}
}

func TestDecryptOperatorID_EmptyString(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	_, err := sm.DecryptOperatorID("")
	if err == nil {
		t.Error("DecryptOperatorID(\"\") should return error")
	}
}

func TestDecryptOperatorID_InvalidBase64(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	_, err := sm.DecryptOperatorID("not-valid-base64!!!")
	if err == nil {
		t.Error("DecryptOperatorID with invalid base64 should return error")
	}
	if !errors.Is(err, ErrDecryptionFailed) {
		t.Errorf("error = %v, want ErrDecryptionFailed", err)
	}
}

func TestDecryptOperatorID_TamperedCiphertext(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	operatorID := "op_test123"

	encrypted, err := sm.EncryptOperatorID(operatorID)
	if err != nil {
		t.Fatalf("EncryptOperatorID() error = %v", err)
	}

	// Tamper with the ciphertext
	tampered := encrypted[:len(encrypted)-4] + "XXXX"

	_, err = sm.DecryptOperatorID(tampered)
	if err == nil {
		t.Error("DecryptOperatorID with tampered ciphertext should return error")
	}
}

func TestDecryptOperatorID_WrongSecret(t *testing.T) {
	sm1 := NewSessionManager("secret-one")
	sm2 := NewSessionManager("secret-two")

	encrypted, err := sm1.EncryptOperatorID("op_test")
	if err != nil {
		t.Fatalf("EncryptOperatorID() error = %v", err)
	}

	_, err = sm2.DecryptOperatorID(encrypted)
	if err == nil {
		t.Error("decrypting with wrong secret should fail")
	}
}

func TestDecryptOperatorID_TooShortCiphertext(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	// GCM nonce is 12 bytes, so anything shorter than that should fail
	short := "YWJj" // "abc" in base64

	_, err := sm.DecryptOperatorID(short)
	if err == nil {
		t.Error("ciphertext shorter than nonce size should fail")
	}
}

func TestCreateSessionCookie(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	operatorID := "op_test123"

	cookie, err := sm.CreateSessionCookie(operatorID)
	if err != nil {
		t.Fatalf("CreateSessionCookie() error = %v", err)
	}

	if cookie.Name != CookieName {
		t.Errorf("cookie.Name = %q, want %q", cookie.Name, CookieName)
	}

	if cookie.Value == "" {
		t.Error("cookie value should not be empty")
	}

	if cookie.Path != CookiePath {
		t.Errorf("cookie.Path = %q, want %q", cookie.Path, CookiePath)
	}

	if cookie.MaxAge != CookieMaxAge {
		t.Errorf("cookie.MaxAge = %d, want %d", cookie.MaxAge, CookieMaxAge)
	}

	if !cookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}

	if !cookie.Secure {
		t.Error("cookie should be Secure")
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie.SameSite = %v, want %v", cookie.SameSite, http.SameSiteLaxMode)
	}
}

func TestCreateSessionCookieWithExpiry(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	operatorID := "op_test123"
	customMaxAge := 3600 // 1 hour

	cookie, err := sm.CreateSessionCookieWithExpiry(operatorID, customMaxAge)
	if err != nil {
		t.Fatalf("CreateSessionCookieWithExpiry() error = %v", err)
	}

	if cookie.MaxAge != customMaxAge {
		t.Errorf("cookie.MaxAge = %d, want %d", cookie.MaxAge, customMaxAge)
	}

	if cookie.Name != CookieName {
		t.Errorf("cookie.Name = %q, want %q", cookie.Name, CookieName)
	}

	if !cookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}
}

func TestCreateSessionCookieWithExpiry_ZeroMaxAge(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	cookie, err := sm.CreateSessionCookieWithExpiry("op_test", 0)
	if err != nil {
		t.Fatalf("CreateSessionCookieWithExpiry() error = %v", err)
	}
	if cookie.MaxAge != 0 {
		t.Errorf("cookie.MaxAge = %d, want 0", cookie.MaxAge)
	}
}

func TestClearSessionCookie(t *testing.T) {
	sm := NewSessionManager("test-secret-key")

	cookie := sm.ClearSessionCookie()

	if cookie.Name != CookieName {
		t.Errorf("cookie.Name = %q, want %q", cookie.Name, CookieName)
	}

	if cookie.Value != "" {
		t.Error("cleared cookie should have empty value")
	}

	if cookie.MaxAge != -1 {
		t.Errorf("cookie.MaxAge = %d, want -1", cookie.MaxAge)
	}

	if !cookie.Expires.IsZero() && cookie.Expires.Unix() != 0 {
		t.Error("cookie should expire at Unix epoch")
	}

	if !cookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}

	if !cookie.Secure {
		t.Error("cookie should be Secure")
	}
}

func TestExtractSessionFromCookie(t *testing.T) {
	sm := NewSessionManager("test-secret-key")
	operatorID := "op_extract123"

	encrypted, err := sm.EncryptOperatorID(operatorID)
	if err != nil {
		t.Fatalf("EncryptOperatorID() error = %v", err)
	}

	extracted, err := sm.ExtractSessionFromCookie(encrypted)
	if err != nil {
		t.Fatalf("ExtractSessionFromCookie() error = %v", err)
	}

	if extracted != operatorID {
		t.Errorf("extracted = %q, want %q", extracted, operatorID)
	}
}

func TestExtractSessionFromCookie_Empty(t *testing.T) {
	sm := NewSessionManager("test-secret-key")

	_, err := sm.ExtractSessionFromCookie("")
	if err != ErrInvalidCookie {
		t.Errorf("error = %v, want %v", err, ErrInvalidCookie)
	}
}

func TestHashOperatorID(t *testing.T) {
	operatorID := "op_test123"
	hash := HashOperatorID(operatorID)

	if hash == "" {
		t.Error("hash should not be empty")
	}

	// SHA-256 produces 64 hex characters
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Hash should be deterministic
	hash2 := HashOperatorID(operatorID)
	if hash != hash2 {
		t.Error("same operator ID should produce same hash")
	}
}

func TestHashOperatorID_Uniqueness(t *testing.T) {
	hash1 := HashOperatorID("op_one")
	hash2 := HashOperatorID("op_two")

	if hash1 == hash2 {
		t.Error("different operator IDs should produce different hashes")
	}
}

func TestHashOperatorID_NotReversible(t *testing.T) {
	operatorID := "op_secret123"
	hash := HashOperatorID(operatorID)

	// Hash should not contain the original ID
	if strings.Contains(hash, operatorID) {
		t.Error("hash should not contain the original operator ID")
	}
}

func TestCookieConstants(t *testing.T) {
	if CookieName != "vyz_session" {
		t.Errorf("CookieName = %q, want \"vyz_session\"", CookieName)
	}

	if CookieMaxAge != 86400 {
		t.Errorf("CookieMaxAge = %d, want 86400 (24 hours)", CookieMaxAge)
	}

	if CookiePath != "/" {
		t.Errorf("CookiePath = %q, want \"/\"", CookiePath)
	}

	if EncryptionKeyLen != 32 {
		t.Errorf("EncryptionKeyLen = %d, want 32 (AES-256)", EncryptionKeyLen)
	}
}

func TestSessionManager_EmptySecret(t *testing.T) {
	sm := NewSessionManager("")
	if sm == nil {
		t.Fatal("NewSessionManager with empty secret should not return nil")
	}
	if len(sm.encryptionKey) != 32 {
		t.Errorf("encryptionKey length = %d, want 32", len(sm.encryptionKey))
	}

	// Should still be able to encrypt/decrypt with empty secret
	encrypted, err := sm.EncryptOperatorID("op_test")
	if err != nil {
		t.Fatalf("EncryptOperatorID() error = %v", err)
	}

	decrypted, err := sm.DecryptOperatorID(encrypted)
	if err != nil {
		t.Fatalf("DecryptOperatorID() error = %v", err)
	}

	if decrypted != "op_test" {
		t.Errorf("decrypted = %q, want \"op_test\"", decrypted)
	}
}

func TestSessionManager_LongOperatorID(t *testing.T) {
	sm := NewSessionManager("test-secret")
	// Test with a very long operator ID
	longID := strings.Repeat("a", 1000)

	encrypted, err := sm.EncryptOperatorID(longID)
	if err != nil {
		t.Fatalf("EncryptOperatorID() error = %v", err)
	}

	decrypted, err := sm.DecryptOperatorID(encrypted)
	if err != nil {
		t.Fatalf("DecryptOperatorID() error = %v", err)
	}

	if decrypted != longID {
		t.Error("long operator ID should round-trip correctly")
	}
}

func TestSessionManager_SpecialCharacters(t *testing.T) {
	sm := NewSessionManager("test-secret")
	testIDs := []string{
		"op_with-dash",
		"op_with_underscore",
		"op_with.dots",
		"op_with@email",
		"op_with_unicode_🦀",
	}

	for _, id := range testIDs {
		encrypted, err := sm.EncryptOperatorID(id)
		if err != nil {
			t.Errorf("EncryptOperatorID(%q) error = %v", id, err)
			continue
		}

		decrypted, err := sm.DecryptOperatorID(encrypted)
		if err != nil {
			t.Errorf("DecryptOperatorID() error = %v", err)
			continue
		}

		if decrypted != id {
			t.Errorf("round-trip for %q failed: got %q", id, decrypted)
		}
	}
}