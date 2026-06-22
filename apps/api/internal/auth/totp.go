// Package auth provides TOTP-based multi-factor authentication.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/totp"
)

// Re-export from new location for backward compatibility.
type (
TOTPConfig = totp.Config
TOTP       = totp.TOTP
)

const TOTPDigits = totp.Digits

func DefaultTOTPConfig() totp.Config {
return totp.DefaultConfig()
}

func GenerateSecret() (string, error) {
return totp.GenerateSecret()
}

func NewTOTP(secret string, cfg totp.Config) *totp.TOTP {
return totp.New(secret, cfg)
}

func GenerateBackupCodes(count int) ([]string, error) {
return totp.GenerateBackupCodes(count)
}

func ValidateBackupCode(stored []string, code string) int {
return totp.ValidateBackupCode(stored, code)
}

func RemoveBackupCode(codes []string, index int) []string {
return totp.RemoveBackupCode(codes, index)
}
