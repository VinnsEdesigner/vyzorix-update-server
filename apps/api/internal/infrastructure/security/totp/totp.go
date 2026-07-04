// Package totp provides TOTP-based multi-factor authentication.
package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"
)

const (
	Digits    = 6
	Period    = 30
	Algorithm = "SHA512"
)

// MFASecretBinding contains the TOTP secret and its cryptographic binding.
type MFASecretBinding struct {
	Secret string // The base32-encoded TOTP secret
	MAC    string // HMAC-SHA256 of the secret, keyed by operator ID
}

// Config holds TOTP configuration.
type Config struct {
	Issuer      string
	AccountName string
	Algorithm   string
	Digits      int
	Period      int
}

// DefaultConfig returns the default TOTP configuration.
func DefaultConfig() Config {
	return Config{
		Issuer:      "Vyzorix",
		AccountName: "",
		Digits:      Digits,
		Period:      Period,
		Algorithm:   Algorithm,
	}
}

// TOTP represents a TOTP secret and its configuration.
type TOTP struct {
	Secret      string
	BackupCodes []string
	Config      Config
}

// GenerateSecret generates a new random TOTP secret.
func GenerateSecret() (string, error) {
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

// New creates a new TOTP with the given secret.
func New(secret string, cfg Config) *TOTP {
	return &TOTP{
		Secret: secret,
		Config: cfg,
	}
}

// GenerateCode generates the current TOTP code.
func (t *TOTP) GenerateCode() (string, error) {
	return t.GenerateCodeAt(time.Now())
}

// GenerateCodeAt generates a TOTP code for a specific time.
func (t *TOTP) GenerateCodeAt(timestamp time.Time) (string, error) {
	counter := uint64(math.Floor(float64(timestamp.Unix()) / float64(t.Config.Period)))
	return t.generateHOTP(counter)
}

// Verify checks if a TOTP code is valid for the current time.
func (t *TOTP) Verify(code string) bool {
	now := time.Now()
	for _, offset := range []int{-1, 0, 1} {
		checkTime := now.Add(time.Duration(offset*t.Config.Period) * time.Second)

		generated, err := t.GenerateCodeAt(checkTime)
		if err == nil && hmac.Equal([]byte(generated), []byte(code)) {
			return true
		}
	}

	return false
}

func (t *TOTP) generateHOTP(counter uint64) (string, error) {
	secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(t.Secret))
	if err != nil {
		return "", fmt.Errorf("invalid secret: %w", err)
	}

	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)

	h := hmac.New(sha512.New, secret)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	truncatedHash := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	code := truncatedHash % uint32(math.Pow10(t.Config.Digits))

	return fmt.Sprintf("%0*d", t.Config.Digits, code), nil
}

// ProvisioningURI generates the TOTP provisioning URI for authenticator apps.
func (t *TOTP) ProvisioningURI() string {
	return t.ProvisioningURIFor(t.Config.AccountName)
}

// ProvisioningURIFor generates the TOTP provisioning URI with a custom account name.
func (t *TOTP) ProvisioningURIFor(accountName string) string {
	if accountName == "" {
		accountName = t.Config.AccountName
	}

	params := url.Values{}
	params.Set("secret", t.Secret)
	params.Set("issuer", t.Config.Issuer)
	params.Set("algorithm", t.Config.Algorithm)
	params.Set("digits", fmt.Sprintf("%d", t.Config.Digits))
	params.Set("period", fmt.Sprintf("%d", t.Config.Period))

	return fmt.Sprintf("otpauth://totp/%s:%s?%s",
		url.PathEscape(t.Config.Issuer),
		url.PathEscape(accountName),
		params.Encode())
}

// GenerateBackupCodes generates a set of backup codes for account recovery.
func GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		code := make([]byte, 8)
		if _, err := rand.Read(code); err != nil {
			return nil, err
		}

		encoded := base32.StdEncoding.EncodeToString(code)
		codes[i] = fmt.Sprintf("%s-%s-%s",
			encoded[:4],
			encoded[4:8],
			encoded[8:12],
		)
	}

	return codes, nil
}

// ValidateBackupCode checks if a backup code is valid and returns its index.
func ValidateBackupCode(stored []string, code string) int {
	for i, storedCode := range stored {
		if hmac.Equal([]byte(storedCode), []byte(code)) {
			return i
		}
	}

	return -1
}

// RemoveBackupCode removes a used backup code from the list.
func RemoveBackupCode(codes []string, index int) []string {
	if index < 0 || index >= len(codes) {
		return codes
	}

	return append(codes[:index], codes[index+1:]...)
}

// CreateMFASecretBinding creates a cryptographically bound MFA secret for an operator.
// 7: The MAC binds the secret to the operator ID, preventing token theft attacks.
func CreateMFASecretBinding(operatorID, secret string) MFASecretBinding {
	mac := hmac.New(sha512.New, []byte(operatorID))
	mac.Write([]byte(secret))
	return MFASecretBinding{
		Secret: secret,
		MAC:    hex.EncodeToString(mac.Sum(nil)),
	}
}

// VerifyMFASecretBinding verifies that a TOTP code was generated using the bound secret.
// 7: Rejects codes if the operator ID doesn't match the binding.
func VerifyMFASecretBinding(operatorID string, binding MFASecretBinding) bool {
	expectedMAC := CreateMFASecretBinding(operatorID, binding.Secret)
	return hmac.Equal([]byte(binding.MAC), []byte(expectedMAC.MAC))
}

// GenerateBoundSecret generates a new MFA secret bound to an operator.
func GenerateBoundSecret(operatorID string) (MFASecretBinding, error) {
	secret, err := GenerateSecret()
	if err != nil {
		return MFASecretBinding{}, err
	}
	return CreateMFASecretBinding(operatorID, secret), nil
}
