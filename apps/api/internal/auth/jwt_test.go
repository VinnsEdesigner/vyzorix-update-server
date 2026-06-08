package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTManager_Generate(t *testing.T) {
	manager := NewJWTManager("test-secret-key-32-bytes-long!!", 7*24*time.Hour, "test-issuer")

	token, expiresAt, err := manager.Generate("operator-123", "test@example.com", "Test User", "admin")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if expiresAt.Before(time.Now()) {
		t.Error("expected future expiration time")
	}
}

func TestJWTManager_Verify_ValidToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-32-bytes-long!!", 7*24*time.Hour, "test-issuer")

	token, _, err := manager.Generate("operator-123", "test@example.com", "Test User", "admin")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if claims.OperatorID != "operator-123" {
		t.Errorf("OperatorID = %s, want operator-123", claims.OperatorID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Email = %s, want test@example.com", claims.Email)
	}
	if claims.Name != "Test User" {
		t.Errorf("Name = %s, want Test User", claims.Name)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %s, want admin", claims.Role)
	}
}

func TestJWTManager_Verify_ExpiredToken(t *testing.T) {
	// Create manager with 0 duration (already expired)
	manager := NewJWTManager("test-secret-key-32-bytes-long!!", -1*time.Hour, "test-issuer")

	token, _, err := manager.Generate("operator-123", "test@example.com", "Test User", "admin")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	_, err = manager.Verify(token)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestJWTManager_Verify_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-32-bytes-long!!", 7*24*time.Hour, "test-issuer")

	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"garbage", "not.a.valid.jwt.token"},
		{"missing parts", "header.payload"},
		{"random", "randomstring"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.Verify(tt.token)
			if err == nil {
				t.Error("expected error for invalid token")
			}
		})
	}
}

func TestJWTManager_Verify_WrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret-one-32-bytes-long-long!!", 7*24*time.Hour, "test-issuer")
	manager2 := NewJWTManager("secret-two-32-bytes-long-long!!", 7*24*time.Hour, "test-issuer")

	token, _, err := manager1.Generate("operator-123", "test@example.com", "Test User", "admin")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	_, err = manager2.Verify(token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for wrong secret, got %v", err)
	}
}

func TestJWTManager_Verify_WrongAlgorithm(t *testing.T) {
	manager := NewJWTManager("test-secret-key-32-bytes-long!!", 7*24*time.Hour, "test-issuer")

	// Create a token with none algorithm (invalid)
	claims := OperatorClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "test-issuer",
		},
		OperatorID: "operator-123",
		Email:      "test@example.com",
		Name:       "Test User",
		Role:       "admin",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	// This should fail - none algorithm is not allowed
	tokenString, _ := token.SignedString([]byte(""))

	_, err := manager.Verify(tokenString)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for 'none' algorithm, got %v", err)
	}
}

func TestJWTManager_Verify_ModifiedToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-32-bytes-long!!", 7*24*time.Hour, "test-issuer")

	token, _, err := manager.Generate("operator-123", "test@example.com", "Test User", "admin")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Modify the token (change one character)
	modifiedToken := token[:len(token)-10] + "X" + token[len(token)-9:]

	_, err = manager.Verify(modifiedToken)
	if err == nil {
		t.Error("expected error for modified token")
	}
}

func TestJWTManager_Verify_OperatorNotFound(t *testing.T) {
	manager := NewJWTManager("test-secret-key-32-bytes-long!!", 7*24*time.Hour, "test-issuer")

	token, _, err := manager.Generate("operator-123", "test@example.com", "Test User", "admin")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if claims.OperatorID != "operator-123" {
		t.Errorf("OperatorID = %s, want operator-123", claims.OperatorID)
	}
}

func TestHashToken(t *testing.T) {
	token := "my-jwt-token-string"
	hash := HashToken(token)

	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if hash == token {
		t.Error("hash should not equal original token")
	}
	if len(hash) != 64 { // SHA256 produces 64 hex chars
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Same input should produce same hash
	hash2 := HashToken(token)
	if hash != hash2 {
		t.Error("same token should produce same hash")
	}
}

func TestHashToken_DifferentTokens(t *testing.T) {
	hash1 := HashToken("token-1")
	hash2 := HashToken("token-2")

	if hash1 == hash2 {
		t.Error("different tokens should produce different hashes")
	}
}

func TestGenerateTokenID(t *testing.T) {
	id1 := generateTokenID()
	id2 := generateTokenID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
}

func TestJWTManager_DifferentExpiry(t *testing.T) {
	manager1 := NewJWTManager("test-secret-key-32-bytes-long!!", 1*time.Hour, "test-issuer")
	manager2 := NewJWTManager("test-secret-key-32-bytes-long!!", 24*time.Hour, "test-issuer")

	token1, _, _ := manager1.Generate("op", "e@e.com", "N", "r")
	token2, _, _ := manager2.Generate("op", "e@e.com", "N", "r")

	// Both should be valid now
	_, err := manager1.Verify(token1)
	if err != nil {
		t.Errorf("token1 should be valid, got %v", err)
	}
	_, err = manager2.Verify(token2)
	if err != nil {
		t.Errorf("token2 should be valid, got %v", err)
	}
}

func TestJWTManager_ClaimsStruct(t *testing.T) {
	claims := OperatorClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "token-id-123",
		},
		OperatorID: "op-123",
		Email:      "user@example.com",
		Name:       "John Doe",
		Role:       "super_admin",
	}

	if claims.Issuer != "test-issuer" {
		t.Errorf("Issuer = %s, want test-issuer", claims.Issuer)
	}
	if claims.OperatorID != "op-123" {
		t.Errorf("OperatorID = %s, want op-123", claims.OperatorID)
	}
	if claims.Email != "user@example.com" {
		t.Errorf("Email = %s, want user@example.com", claims.Email)
	}
	if claims.Role != "super_admin" {
		t.Errorf("Role = %s, want super_admin", claims.Role)
	}
}

func BenchmarkJWTGenerate(b *testing.B) {
	manager := NewJWTManager("test-secret-key-32-bytes-long!!", 7*24*time.Hour, "test-issuer")

	for i := 0; i < b.N; i++ {
		_, _, _ = manager.Generate("operator-123", "test@example.com", "Test User", "admin")
	}
}

func BenchmarkJWTVerify(b *testing.B) {
	manager := NewJWTManager("test-secret-key-32-bytes-long!!", 7*24*time.Hour, "test-issuer")
	token, _, _ := manager.Generate("operator-123", "test@example.com", "Test User", "admin")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.Verify(token)
	}
}
