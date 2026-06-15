// Package config provides SSR configuration.
package config

import "os"

// SSRConfig holds SSR server configuration.
// This allows the Go server to proxy requests to the Node.js SSR server.
type SSRConfig struct {
	// EnableSSR enables SSR mode.
	EnableSSR bool `json:"enableSsr" yaml:"enableSsr"`

	// SSRServerURL is the URL of the Node.js SSR server.
	// Default: http://localhost:3001.
	SSRServerURL string `json:"ssrServerUrl" yaml:"ssrServerUrl"`

	// SSRPort is the port the SSR server listens on.
	// Default: 3001.
	SSRPort string `json:"ssrPort" yaml:"ssrPort"`

	// SSRAutoStart enables automatic SSR server startup by the Go server.
	// When true, Go server will spawn the SSR subprocess automatically.
	SSRAutoStart bool `json:"ssrAutoStart" yaml:"ssrAutoStart"`

	// SSRBuildTimeout is the timeout for waiting for SSR server to be ready.
	// Default: 30 seconds.
	SSRBuildTimeout int `json:"ssrBuildTimeout" yaml:"ssrBuildTimeout"`

	// SSRAutoBuild enables automatic web app build if not present.
	// Default: true (always build if needed for SSR).
	SSRAutoBuild bool `json:"ssrAutoBuild" yaml:"ssrAutoBuild"`

	// SSRHealthCheckInterval is the interval for health checks to SSR server.
	// Default: 5 seconds.
	SSRHealthCheckInterval int `json:"ssrHealthCheckInterval" yaml:"ssrHealthCheckInterval"`

	// SSRRetryAttempts is the number of retry attempts for SSR server startup.
	// Default: 3.
	SSRRetryAttempts int `json:"ssrRetryAttempts" yaml:"ssrRetryAttempts"`
}

// DefaultSSRConfig returns default SSR configuration.
// SSR is ENABLED by default for all environments to ensure consistent behavior.
func DefaultSSRConfig() SSRConfig {
	return SSRConfig{
		EnableSSR:              true, // SSR enabled by default - always use SSR when available
		SSRServerURL:           "http://localhost:3001",
		SSRPort:                "3001",
		SSRAutoStart:           true, // Auto-start SSR subprocess by default
		SSRBuildTimeout:        60,   // 60 seconds to wait for SSR server (increased for build time)
		SSRAutoBuild:           true, // Always build web app if needed for SSR
		SSRHealthCheckInterval: 5,    // Check SSR health every 5 seconds
		SSRRetryAttempts:       3,    // Retry SSR startup up to 3 times
	}
}

// LoadSSRConfig loads SSR configuration from environment.
func LoadSSRConfig() SSRConfig {
	config := DefaultSSRConfig()

	// SSR_ENABLED=false explicitly disables SSR (for pure SPA mode)
	if v := os.Getenv("SSR_ENABLED"); v != "" {
		config.EnableSSR = v == "true" || v == "1" || v == "yes"
	}

	if v := os.Getenv("SSR_SERVER_URL"); v != "" {
		config.SSRServerURL = v
	}

	if v := os.Getenv("SSR_PORT"); v != "" {
		config.SSRPort = v
	}

	// SSR_AUTO_START=false can disable auto-start (run SSR manually)
	if v := os.Getenv("SSR_AUTO_START"); v != "" {
		config.SSRAutoStart = v == "true" || v == "1" || v == "yes"
	}

	if v := os.Getenv("SSR_BUILD_TIMEOUT"); v != "" {
		if timeout := parseInt(v); timeout > 0 {
			config.SSRBuildTimeout = timeout
		}
	}

	// SSR_AUTO_BUILD=false disables automatic web app build
	if v := os.Getenv("SSR_AUTO_BUILD"); v != "" {
		config.SSRAutoBuild = v == "true" || v == "1" || v == "yes"
	}

	if v := os.Getenv("SSR_HEALTH_CHECK_INTERVAL"); v != "" {
		if interval := parseInt(v); interval > 0 {
			config.SSRHealthCheckInterval = interval
		}
	}

	if v := os.Getenv("SSR_RETRY_ATTEMPTS"); v != "" {
		if retries := parseInt(v); retries > 0 {
			config.SSRRetryAttempts = retries
		}
	}

	return config
}

// parseInt parses an integer from a string.
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
