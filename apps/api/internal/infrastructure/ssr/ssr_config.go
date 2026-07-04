// Package ssr provides modular SSR server management components.
package ssr

import "os"

// Config holds all SSR configuration with sensible defaults.
type Config struct {
	SSRServerURL           string
	SSRPort                string
	SSRBuildTimeout        int
	SSRHealthCheckInterval int
	SSRRetryAttempts       int
	SSRRetryBackoff        int
	EnableSSR              bool
	SSRAutoStart           bool
	SSRAutoBuild           bool
}

// DefaultConfig returns default SSR configuration.
func DefaultConfig() Config {
	return Config{
		EnableSSR:              true,
		SSRServerURL:           "http://localhost:3001",
		SSRPort:                "3001",
		SSRAutoStart:           true,
		SSRAutoBuild:           true,
		SSRBuildTimeout:        60,
		SSRHealthCheckInterval: 5,
		SSRRetryAttempts:       3,
		SSRRetryBackoff:        2,
	}
}

// LoadConfig loads SSR configuration from environment variables.
func LoadConfig() Config {
	cfg := DefaultConfig()

	if v := os.Getenv("SSR_ENABLED"); v != "" {
		cfg.EnableSSR = v == "true" || v == "1" || v == "yes"
	}

	if v := os.Getenv("SSR_SERVER_URL"); v != "" {
		cfg.SSRServerURL = v
	}

	if v := os.Getenv("SSR_PORT"); v != "" {
		cfg.SSRPort = v
	}

	if v := os.Getenv("SSR_AUTO_START"); v != "" {
		cfg.SSRAutoStart = v == "true" || v == "1" || v == "yes"
	}

	if v := os.Getenv("SSR_AUTO_BUILD"); v != "" {
		cfg.SSRAutoBuild = v == "true" || v == "1" || v == "yes"
	}

	if v := os.Getenv("SSR_BUILD_TIMEOUT"); v != "" {
		if n := parseInt(v); n > 0 {
			cfg.SSRBuildTimeout = n
		}
	}

	if v := os.Getenv("SSR_HEALTH_CHECK_INTERVAL"); v != "" {
		if n := parseInt(v); n > 0 {
			cfg.SSRHealthCheckInterval = n
		}
	}

	if v := os.Getenv("SSR_RETRY_ATTEMPTS"); v != "" {
		if n := parseInt(v); n > 0 {
			cfg.SSRRetryAttempts = n
		}
	}

	if v := os.Getenv("SSR_RETRY_BACKOFF"); v != "" {
		if n := parseInt(v); n > 0 {
			cfg.SSRRetryBackoff = n
		}
	}

	return cfg
}

// parseInt parses a non-negative integer from a string.
func parseInt(s string) int {
	n := 0

	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return 0
		}
	}

	return n
}
