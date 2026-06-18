// Package config provides configuration loading for the application.
package config

import (
	"os"
	"strconv"
)

// TOTPConfig holds TOTP/MFA configuration.
type TOTPConfig struct {
	Enabled     bool
	Issuer      string
	Digits      int
	Period      int
	BackupCodes int
}

// LoadTOTPConfig loads TOTP configuration from environment variables.
func LoadTOTPConfig() TOTPConfig {
	digits := 6
	if v := os.Getenv("TOTP_DIGITS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 6 && n <= 8 {
			digits = n
		}
	}

	period := 30
	if v := os.Getenv("TOTP_PERIOD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 15 && n <= 60 {
			period = n
		}
	}

	backupCodes := 10
	if v := os.Getenv("MFA_BACKUP_CODES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 5 && n <= 20 {
			backupCodes = n
		}
	}

	issuer := "Vyzorix"
	if v := os.Getenv("TOTP_ISSUER"); v != "" {
		issuer = v
	}

	return TOTPConfig{
		Enabled:     os.Getenv("MFA_ENABLED") == "true",
		Issuer:      issuer,
		Digits:      digits,
		Period:      period,
		BackupCodes: backupCodes,
	}
}