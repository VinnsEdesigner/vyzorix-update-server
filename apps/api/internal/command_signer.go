// Package services provides business logic services.
package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// CommandSigner handles HMAC signing of commands per DEVICE_REGISTRATION.md §5
// and COMMAND_SECURITY.md §3.
type CommandSigner struct{}

// NewCommandSigner creates a new CommandSigner.
func NewCommandSigner() *CommandSigner {
	return &CommandSigner{}
}

// SignCommand generates a nonce and HMAC signature for a command frame.
// Returns the nonce, HMAC hex string, and any error.
//
// Canonical string format (per COMMAND_SECURITY.md §3):
// {dispatchId}|{deviceId}|{command}|{timestamp_unix_ms}|{nonce}|{args}
func (s *CommandSigner) SignCommand(frame *models.CommandFrame, deviceID, secret string) (nonce string, hmacHex string, err error) {
	// Generate 16 random bytes → 32 hex chars
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce = hex.EncodeToString(nonceBytes)

	// Build canonical string
	canonical := BuildCanonicalString(frame, deviceID, nonce)

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	hmacHex = hex.EncodeToString(mac.Sum(nil))

	return nonce, hmacHex, nil
}

// BuildCanonicalString constructs the canonical message string for HMAC computation.
// Format: {dispatchId}|{deviceId}|{command}|{timestamp_unix_ms}|{nonce}|{args}.
func BuildCanonicalString(frame *models.CommandFrame, deviceID, nonce string) string {
	// Timestamp is already in Unix milliseconds (int64)
	argsStr := string(frame.Args)
	if argsStr == "" {
		argsStr = "{}"
	}
	return frame.DispatchID + "|" + deviceID + "|" + frame.Command + "|" +
		strconv.FormatInt(frame.Timestamp, 10) + "|" + nonce + "|" + argsStr
}

// ValidateCommandHMAC validates a command frame's HMAC signature.
// Returns true if valid, false otherwise.
func (s *CommandSigner) ValidateCommandHMAC(frame *models.CommandFrame, deviceID, secret string) bool {
	// Recompute canonical string with the frame's nonce and timestamp
	argsStr := string(frame.Args)
	if argsStr == "" {
		argsStr = "{}"
	}
	canonical := frame.DispatchID + "|" + deviceID + "|" + frame.Command + "|" +
		strconv.FormatInt(frame.Timestamp, 10) + "|" + frame.Nonce + "|" + argsStr

	// Compute expected HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	expected := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison
	return hmac.Equal([]byte(frame.Signature), []byte(expected))
}

// ValidateTimestamp checks if the command timestamp is within the ±30s window.
func (s *CommandSigner) ValidateTimestamp(frame *models.CommandFrame, maxDriftMs int64) bool {
	if maxDriftMs == 0 {
		maxDriftMs = 30_000 // Default ±30 seconds
	}

	// Timestamp is already Unix milliseconds
	nowMs := time.Now().UnixMilli()
	drift := nowMs - frame.Timestamp
	if drift < 0 {
		drift = -drift
	}
	return drift <= maxDriftMs
}

// ValidateConnectHMAC validates WebSocket connection HMAC per DEVICE_REGISTRATION.md §4.1.
// Format: HMAC over "CONNECT:<deviceId>:<timestamp>:<nonce>".
func (s *CommandSigner) ValidateConnectHMAC(deviceID, timestamp, nonce, providedHmac, secret string) bool {
	canonical := "CONNECT:" + deviceID + ":" + timestamp + ":" + nonce

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(providedHmac), []byte(expected))
}

// GenerateNonce generates a cryptographically random 16-byte nonce (32 hex chars).
func (s *CommandSigner) GenerateNonce() (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", errors.New("failed to generate nonce")
	}
	return hex.EncodeToString(nonceBytes), nil
}

// GenerateTimestamp generates a Unix timestamp string for connect HMAC.
func (s *CommandSigner) GenerateTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// GenerateTimestampMs generates a Unix timestamp in milliseconds.
func (s *CommandSigner) GenerateTimestampMs() int64 {
	return time.Now().UnixMilli()
}

// HashSecret creates a bcrypt hash of the command secret for secure storage.
// Cost factor is 12 (default bcrypt cost), providing strong protection against brute force.
func (s *CommandSigner) HashSecret(secret string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), 12)
	if err != nil {
		// Fall back to SHA256-based hash if bcrypt fails (should never happen)
		return s.fallbackHash(secret)
	}
	return string(hash)
}

// fallbackHash provides a SHA256-based fallback if bcrypt fails.
func (s *CommandSigner) fallbackHash(secret string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		// Fallback to time-based salt if crypto/rand fails
		salt = []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
	}
	saltHex := hex.EncodeToString(salt)
	mac := sha256.New()
	mac.Write(salt)
	mac.Write([]byte(secret))
	hash := mac.Sum(nil)
	return saltHex + ":" + hex.EncodeToString(hash)
}

// VerifySecretHash verifies a secret against its bcrypt hash.
func (s *CommandSigner) VerifySecretHash(secret, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret))
	if err != nil {
		// Try fallback hash format for backward compatibility
		return s.verifyFallbackHash(secret, hash)
	}
	return true
}

// verifyFallbackHash verifies using the old SHA256-based hash format.
func (s *CommandSigner) verifyFallbackHash(secret, hash string) bool {
	parts := splitHash(hash)
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	mac := sha256.New()
	mac.Write(salt)
	mac.Write([]byte(secret))
	expected := mac.Sum(nil)
	return hmac.Equal([]byte(hex.EncodeToString(expected)), []byte(parts[1]))
}

func splitHash(hash string) []string {
	for i := 0; i < len(hash); i++ {
		if hash[i] == ':' {
			return []string{hash[:i], hash[i+1:]}
		}
	}
	return nil
}
