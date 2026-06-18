// Package config provides configuration management for the Vyzorix API.
package config

import (
	"os"
	"strconv"
)

// SigningConfig holds request signing configuration.
type SigningConfig struct {
	// Enabled enables/disables request signing.
	Enabled bool

	// TimestampWindow is the allowed time window for timestamps (seconds).
	TimestampWindow int

	// MaxCacheSize is the maximum replay cache size.
	MaxCacheSize int

	// GracePeriod is the grace period for key rotation (seconds).
	GracePeriod int

	// AllowUnsignedFallback allows unsigned requests in emergencies.
	AllowUnsignedFallback bool
}

// LoadSigningConfig loads signing configuration from environment.
func LoadSigningConfig() SigningConfig {
	cfg := SigningConfig{
		Enabled:              getEnvBool("REQUEST_SIGNING_ENABLED", false),
		TimestampWindow:       getEnvInt("SIGNING_TIMESTAMP_WINDOW", 300),
		MaxCacheSize:          getEnvInt("SIGNING_MAX_CACHE_SIZE", 100000),
		GracePeriod:           getEnvInt("SIGNING_GRACE_PERIOD", 86400),
		AllowUnsignedFallback: getEnvBool("ALLOW_UNSIGNED_FALLBACK", false),
	}

	// Ensure sensible defaults
	if cfg.TimestampWindow < 60 {
		cfg.TimestampWindow = 60
	}
	if cfg.TimestampWindow > 3600 {
		cfg.TimestampWindow = 3600
	}
	if cfg.MaxCacheSize < 1000 {
		cfg.MaxCacheSize = 1000
	}
	if cfg.MaxCacheSize > 1000000 {
		cfg.MaxCacheSize = 1000000
	}

	return cfg
}

// getEnvBool gets a boolean from environment.
func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}

// getEnvInt gets an integer from environment.
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
