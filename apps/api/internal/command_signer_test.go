package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

func TestCommandSigner_SignCommand(t *testing.T) {
	signer := NewCommandSigner()

	frame := &models.CommandFrame{
		DispatchID: "tx-123",
		Command:    "FORCE_SPEAKER",
		Timestamp:  time.Now().UnixMilli(),
		Args:       json.RawMessage("{}"),
	}
	deviceID := "device-456"
	secret := "abcd1234efgh5678901234567890123456789012345678901234567890123456"

	nonce, hmacHex, err := signer.SignCommand(frame, deviceID, secret)
	if err != nil {
		t.Fatalf("SignCommand() error = %v", err)
	}

	// Nonce should be 32 hex chars (16 bytes)
	if len(nonce) != 32 {
		t.Errorf("nonce length = %d, want 32", len(nonce))
	}

	// HMAC should be 64 hex chars (32 bytes SHA256)
	if len(hmacHex) != 64 {
		t.Errorf("hmac length = %d, want 64", len(hmacHex))
	}

	// HMAC should be valid hex
	for _, c := range hmacHex {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("hmac contains non-hex char: %c", c)
		}
	}
}

func TestCommandSigner_SignCommand_Deterministic(t *testing.T) {
	signer := NewCommandSigner()

	frame := &models.CommandFrame{
		DispatchID: "tx-fixed",
		Command:    "REBOOT",
		Timestamp:  1748260800000, // Fixed timestamp in ms
		Args:       json.RawMessage("{}"),
	}
	deviceID := "device-fixed"
	secret := "testsecret123456789012345678901234567890123456789012345678"

	// Sign twice with same frame - nonces should be different but HMAC computation same
	nonce1, hmac1, err := signer.SignCommand(frame, deviceID, secret)
	if err != nil {
		t.Fatalf("SignCommand() error = %v", err)
	}

	nonce2, hmac2, err := signer.SignCommand(frame, deviceID, secret)
	if err != nil {
		t.Fatalf("SignCommand() error = %v", err)
	}

	// Nonces should be different (random)
	if nonce1 == nonce2 {
		t.Error("nonces should be different (random)")
	}

	// HMAC values should be different because nonce changes
	if hmac1 == hmac2 {
		t.Error("hmac values should be different due to different nonces")
	}
}

func TestCommandSigner_ValidateCommandHMAC(t *testing.T) {
	signer := NewCommandSigner()

	frame := &models.CommandFrame{
		DispatchID: "tx-123",
		Command:    "FORCE_SPEAKER",
		Timestamp:  time.Now().UnixMilli(),
		Args:       json.RawMessage("{}"),
	}
	deviceID := "device-456"
	secret := "testsecret123456789012345678901234567890123456789012345678"

	// Sign the command
	nonce, hmacHex, err := signer.SignCommand(frame, deviceID, secret)
	if err != nil {
		t.Fatalf("SignCommand() error = %v", err)
	}
	frame.Nonce = nonce
	frame.Signature = hmacHex

	// Validation should pass
	if !signer.ValidateCommandHMAC(frame, deviceID, secret) {
		t.Error("ValidateCommandHMAC() should pass for valid signature")
	}
}

func TestCommandSigner_ValidateCommandHMAC_InvalidSecret(t *testing.T) {
	signer := NewCommandSigner()

	frame := &models.CommandFrame{
		DispatchID: "tx-123",
		Command:    "FORCE_SPEAKER",
		Timestamp:  time.Now().UnixMilli(),
		Args:       json.RawMessage("{}"),
	}
	deviceID := "device-456"
	secret := "correctsecret1234567890123456789012345678901234567890123456"
	wrongSecret := "wrongsecret12345678901234567890123456789012345678901234567"

	nonce, hmacHex, err := signer.SignCommand(frame, deviceID, secret)
	if err != nil {
		t.Fatalf("SignCommand() error = %v", err)
	}
	frame.Nonce = nonce
	frame.Signature = hmacHex

	// Validation should fail with wrong secret
	if signer.ValidateCommandHMAC(frame, deviceID, wrongSecret) {
		t.Error("ValidateCommandHMAC() should fail for wrong secret")
	}
}

func TestCommandSigner_ValidateCommandHMAC_WrongDeviceID(t *testing.T) {
	signer := NewCommandSigner()

	frame := &models.CommandFrame{
		DispatchID: "tx-123",
		Command:    "FORCE_SPEAKER",
		Timestamp:  time.Now().UnixMilli(),
		Args:       json.RawMessage("{}"),
	}
	deviceID := "device-456"
	wrongDeviceID := "device-789"
	secret := "testsecret123456789012345678901234567890123456789012345678"

	nonce, hmacHex, err := signer.SignCommand(frame, deviceID, secret)
	if err != nil {
		t.Fatalf("SignCommand() error = %v", err)
	}
	frame.Nonce = nonce
	frame.Signature = hmacHex

	// Validation should fail with wrong device ID
	if signer.ValidateCommandHMAC(frame, wrongDeviceID, secret) {
		t.Error("ValidateCommandHMAC() should fail for wrong device ID")
	}
}

func TestCommandSigner_ValidateCommandHMAC_TamperedFrame(t *testing.T) {
	signer := NewCommandSigner()

	frame := &models.CommandFrame{
		DispatchID: "tx-123",
		Command:    "FORCE_SPEAKER",
		Timestamp:  time.Now().UnixMilli(),
		Args:       json.RawMessage("{}"),
	}
	deviceID := "device-456"
	secret := "testsecret123456789012345678901234567890123456789012345678"

	nonce, hmacHex, err := signer.SignCommand(frame, deviceID, secret)
	if err != nil {
		t.Fatalf("SignCommand() error = %v", err)
	}
	frame.Nonce = nonce
	frame.Signature = hmacHex

	// Tamper with the frame after signing
	frame.Command = "DANGER_COMMAND"

	// Validation should fail
	if signer.ValidateCommandHMAC(frame, deviceID, secret) {
		t.Error("ValidateCommandHMAC() should fail for tampered frame")
	}
}

func TestCommandSigner_ValidateTimestamp(t *testing.T) {
	signer := NewCommandSigner()
	nowMs := time.Now().UnixMilli()

	tests := []struct {
		name      string
		offsetMs  int64
		maxDrift  int64
		wantValid bool
	}{
		{"fresh timestamp", 0, 30_000, true},
		{"10s ago", 10_000, 30_000, true},
		{"29s ago", 29_000, 30_000, true},
		{"31s ago", 31_000, 30_000, false},
		{"1min ago", 60_000, 30_000, false},
		{"10s in future", -10_000, 30_000, true},
		{"31s in future", -31_000, 30_000, false},
		{"custom window 60s", 55_000, 60_000, true},
		{"custom window 60s too old", 65_000, 60_000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := &models.CommandFrame{
				Timestamp: nowMs + tt.offsetMs,
			}
			got := signer.ValidateTimestamp(frame, tt.maxDrift)
			if got != tt.wantValid {
				t.Errorf("ValidateTimestamp() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestCommandSigner_ValidateConnectHMAC(t *testing.T) {
	signer := NewCommandSigner()

	deviceID := "device-123"
	secret := "testsecret123456789012345678901234567890123456789012345678"

	// Generate connect parameters
	timestamp := signer.GenerateTimestamp()
	nonce, _ := signer.GenerateNonce()

	// Compute HMAC manually for the CONNECT format
	canonical := "CONNECT:" + deviceID + ":" + timestamp + ":" + nonce
	
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	expectedHmac := hex.EncodeToString(mac.Sum(nil))

	// Validate should pass with correct HMAC
	if !signer.ValidateConnectHMAC(deviceID, timestamp, nonce, expectedHmac, secret) {
		t.Error("ValidateConnectHMAC() should pass for valid signature")
	}
}

func TestCommandSigner_ValidateConnectHMAC_WrongSecret(t *testing.T) {
	signer := NewCommandSigner()

	deviceID := "device-123"
	secret := "correctsecret1234567890123456789012345678901234567890123456"
	wrongSecret := "wrongsecret12345678901234567890123456789012345678901234567"

	timestamp := signer.GenerateTimestamp()
	nonce, _ := signer.GenerateNonce()

	frame := &models.CommandFrame{
		DispatchID: nonce,
		Command:    "CONNECT",
		Timestamp:  time.Now().UnixMilli(),
		Args:       nil,
	}
	_, hmacHex, _ := signer.SignCommand(frame, deviceID, secret)

	// Validate should fail with wrong secret
	if signer.ValidateConnectHMAC(deviceID, timestamp, nonce, hmacHex, wrongSecret) {
		t.Error("ValidateConnectHMAC() should fail for wrong secret")
	}
}

func TestCommandSigner_GenerateNonce(t *testing.T) {
	signer := NewCommandSigner()

	nonce1, err := signer.GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce() error = %v", err)
	}

	nonce2, err := signer.GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce() error = %v", err)
	}

	if len(nonce1) != 32 {
		t.Errorf("nonce1 length = %d, want 32", len(nonce1))
	}

	if nonce1 == nonce2 {
		t.Error("GenerateNonce() should produce unique nonces")
	}
}

func TestCommandSigner_GenerateTimestamp(t *testing.T) {
	signer := NewCommandSigner()

	before := time.Now().Unix()
	ts := signer.GenerateTimestamp()
	after := time.Now().Unix()

	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		t.Fatalf("GenerateTimestamp() returned non-numeric: %v", err)
	}

	if tsInt < before || tsInt > after {
		t.Errorf("GenerateTimestamp() = %d, want between %d and %d", tsInt, before, after)
	}
}

func TestCommandSigner_GenerateTimestampMs(t *testing.T) {
	signer := NewCommandSigner()

	before := time.Now().UnixMilli()
	ts := signer.GenerateTimestampMs()
	after := time.Now().UnixMilli()

	if ts < before || ts > after {
		t.Errorf("GenerateTimestampMs() = %d, want between %d and %d", ts, before, after)
	}
}

func TestCommandSigner_HashSecret(t *testing.T) {
	signer := NewCommandSigner()

	secret := "testsecret123456789012345678901234567890123456789012345678"

	hash := signer.HashSecret(secret)

	// Hash should be bcrypt format (starts with $2a$, $2b$, etc.)
	if !strings.HasPrefix(hash, "$2") {
		t.Error("HashSecret() should return bcrypt hash format")
	}

	// Hash should be verifiable
	if !signer.VerifySecretHash(secret, hash) {
		t.Error("VerifySecretHash() should pass for correct secret")
	}
}

func TestCommandSigner_HashSecret_DifferentSecrets(t *testing.T) {
	signer := NewCommandSigner()

	secret1 := "secret1_1234567890123456789012345678901234567890123456789"
	secret2 := "secret2_1234567890123456789012345678901234567890123456789"

	hash1 := signer.HashSecret(secret1)
	hash2 := signer.HashSecret(secret2)

	// Hashes should be different
	if hash1 == hash2 {
		t.Error("HashSecret() should produce different hashes for different secrets")
	}

	// Cross-verification should fail
	if signer.VerifySecretHash(secret1, hash2) {
		t.Error("VerifySecretHash() should fail for mismatched secret/hash")
	}
}

func TestCommandSigner_VerifySecretHash_InvalidHash(t *testing.T) {
	signer := NewCommandSigner()

	// Invalid hash formats
	invalidHashes := []string{
		"nocolon",
		"invalidsalthex",
		"",
		"a",
	}

	for _, hash := range invalidHashes {
		if signer.VerifySecretHash("anyscret", hash) {
			t.Errorf("VerifySecretHash() should fail for invalid hash: %s", hash)
		}
	}
}

func TestBuildCanonicalString(t *testing.T) {
	frame := &models.CommandFrame{
		DispatchID: "tx-abc",
		Command:    "FORCE_SPEAKER",
		Timestamp:  1748260800000, // Fixed timestamp in ms
		Args:       json.RawMessage("{}"),
	}
	deviceID := "device-xyz"
	nonce := "a1b2c3d4e5f67890"

	canonical := BuildCanonicalString(frame, deviceID, nonce)

	expected := "tx-abc|device-xyz|FORCE_SPEAKER|1748260800000|a1b2c3d4e5f67890|{}"
	if canonical != expected {
		t.Errorf("BuildCanonicalString() = %s, want %s", canonical, expected)
	}
}

func TestBuildCanonicalString_ComplexArgs(t *testing.T) {
	frame := &models.CommandFrame{
		DispatchID: "tx-1",
		Command:    "SET_VOLUME",
		Timestamp:  1000000000,
		Args:       json.RawMessage(`{"level":50,"mute":false}`),
	}
	deviceID := "device-1"
	nonce := "noncenoncenoncenonce123456789012"

	canonical := BuildCanonicalString(frame, deviceID, nonce)

	expected := "tx-1|device-1|SET_VOLUME|1000000000|noncenoncenoncenonce123456789012|{\"level\":50,\"mute\":false}"
	if canonical != expected {
		t.Errorf("BuildCanonicalString() = %s, want %s", canonical, expected)
	}
}

func TestBuildCanonicalString_EmptyArgs(t *testing.T) {
	frame := &models.CommandFrame{
		DispatchID: "tx-1",
		Command:    "REBOOT",
		Timestamp:  1000000000,
		Args:       nil,
	}
	deviceID := "device-1"
	nonce := "abcd"

	canonical := BuildCanonicalString(frame, deviceID, nonce)

	// Empty args should default to {}
	expected := "tx-1|device-1|REBOOT|1000000000|abcd|{}"
	if canonical != expected {
		t.Errorf("BuildCanonicalString() = %s, want %s", canonical, expected)
	}
}

// Edge case tests

func TestCommandSigner_EmptySecret(t *testing.T) {
	signer := NewCommandSigner()

	frame := &models.CommandFrame{
		DispatchID: "tx-123",
		Command:    "FORCE_SPEAKER",
		Timestamp:  time.Now().UnixMilli(),
		Args:       json.RawMessage("{}"),
	}
	deviceID := "device-456"
	secret := ""

	// Should not panic
	_, _, err := signer.SignCommand(frame, deviceID, secret)
	if err != nil {
		t.Errorf("SignCommand() with empty secret error = %v", err)
	}
}

func TestCommandSigner_VeryLongNonce(t *testing.T) {
	signer := NewCommandSigner()

	frame := &models.CommandFrame{
		DispatchID: "tx-123",
		Command:    "FORCE_SPEAKER",
		Timestamp:  time.Now().UnixMilli(),
		Args:       json.RawMessage("{}"),
	}
	deviceID := "device-456"
	secret := "testsecret"

	// Sign multiple times to get different nonces
	nonces := make(map[string]bool)
	for i := 0; i < 100; i++ {
		nonce, _, err := signer.SignCommand(frame, deviceID, secret)
		if err != nil {
			t.Fatalf("SignCommand() error = %v", err)
		}
		nonces[nonce] = true
	}

	// All nonces should be unique
	if len(nonces) != 100 {
		t.Errorf("Expected 100 unique nonces, got %d", len(nonces))
	}
}

func TestCommandSigner_TimestampBoundary(t *testing.T) {
	signer := NewCommandSigner()
	nowMs := time.Now().UnixMilli()

	// Test exactly at boundary
	frame := &models.CommandFrame{
		Timestamp: nowMs + 30_000, // Exactly 30s in future
	}

	if !signer.ValidateTimestamp(frame, 30_000) {
		t.Error("Timestamp exactly at boundary should be valid")
	}

	frame.Timestamp = nowMs + 30_001 // One ms over
	if signer.ValidateTimestamp(frame, 30_000) {
		t.Error("Timestamp one ms over boundary should be invalid")
	}
}

func TestCommandSigner_HashVerification(t *testing.T) {
	signer := NewCommandSigner()

	// Test that we can generate, hash, verify cycle
	originalSecret := "my-super-secret-key-12345678901234567890123456789012"
	
	hash := signer.HashSecret(originalSecret)
	
	if !signer.VerifySecretHash(originalSecret, hash) {
		t.Error("Hash should verify correctly")
	}
	
	if signer.VerifySecretHash("wrong-secret", hash) {
		t.Error("Wrong secret should not verify")
	}
}