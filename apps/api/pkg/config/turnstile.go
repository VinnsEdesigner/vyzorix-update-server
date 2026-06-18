// Package config provides configuration management for the Vyzorix API.
package config

import "os"

// TurnstileConfig holds Cloudflare Turnstile configuration.
type TurnstileConfig struct {
	// Secret is the Cloudflare Turnstile secret key.
	Secret string

	// Enabled enables/disables Turnstile verification.
	Enabled bool

	// SiteVerifyURL is the Turnstile verification endpoint.
	SiteVerifyURL string
}

// LoadTurnstileConfig loads Turnstile configuration from environment.
func LoadTurnstileConfig() TurnstileConfig {
	cfg := TurnstileConfig{
		Enabled:       getEnvBool("TURNSTILE_ENABLED", false),
		SiteVerifyURL: "https://challenges.cloudflare.com/turnstile/v0/siteverify",
	}

	// Load secret if enabled
	if cfg.Enabled {
		cfg.Secret = os.Getenv("TURNSTILE_SECRET")
	}

	return cfg
}

// IsConfigured returns true if Turnstile is properly configured.
func (c TurnstileConfig) IsConfigured() bool {
	return c.Enabled && c.Secret != ""
}
