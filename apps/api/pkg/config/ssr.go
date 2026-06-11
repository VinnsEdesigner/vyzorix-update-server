// Package config provides SSR configuration
package config

import "os"

// SSRConfig holds SSR server configuration
// This allows the Go server to proxy requests to the Node.js SSR server
type SSRConfig struct {
	// EnableSSR enables SSR mode
	EnableSSR bool `json:"enableSsr" yaml:"enableSsr"`

	// SSRServerURL is the URL of the Node.js SSR server
	// Default: http://localhost:3001
	SSRServerURL string `json:"ssrServerUrl" yaml:"ssrServerUrl"`

	// SSRPort is the port the SSR server listens on
	// Default: 3001
	SSRPort string `json:"ssrPort" yaml:"ssrPort"`
}

// DefaultSSRConfig returns default SSR configuration
func DefaultSSRConfig() SSRConfig {
	return SSRConfig{
		EnableSSR:    false, // Disabled by default for safety
		SSRServerURL: "http://localhost:3001",
		SSRPort:      "3001",
	}
}

// LoadSSRConfig loads SSR configuration from environment
func LoadSSRConfig() SSRConfig {
	config := DefaultSSRConfig()

	if v := os.Getenv("SSR_ENABLED"); v != "" {
		config.EnableSSR = v == "true" || v == "1" || v == "yes"
	}

	if v := os.Getenv("SSR_SERVER_URL"); v != "" {
		config.SSRServerURL = v
	}

	if v := os.Getenv("SSR_PORT"); v != "" {
		config.SSRPort = v
	}

	return config
}
