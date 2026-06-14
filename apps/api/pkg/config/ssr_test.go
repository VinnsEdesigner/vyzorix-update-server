package config

import (
	"os"
	"testing"
)

func TestDefaultSSRConfig(t *testing.T) {
	config := DefaultSSRConfig()

	if config.EnableSSR != false {
		t.Errorf("EnableSSR = %v, want false", config.EnableSSR)
	}
	if config.SSRServerURL != "http://localhost:3001" {
		t.Errorf("SSRServerURL = %q, want \"http://localhost:3001\"", config.SSRServerURL)
	}
	if config.SSRPort != "3001" {
		t.Errorf("SSRPort = %q, want \"3001\"", config.SSRPort)
	}
}

func TestLoadSSRConfig_Defaults(t *testing.T) {
	// Clear any existing env vars
	clearSSREnv()

	config := LoadSSRConfig()

	if config.EnableSSR != false {
		t.Errorf("EnableSSR = %v, want false", config.EnableSSR)
	}
	if config.SSRServerURL != "http://localhost:3001" {
		t.Errorf("SSRServerURL = %q, want default", config.SSRServerURL)
	}
	if config.SSRPort != "3001" {
		t.Errorf("SSRPort = %q, want default", config.SSRPort)
	}
}

func TestLoadSSRConfig_EnableSSR_True(t *testing.T) {
	clearSSREnv()
	os.Setenv("SSR_ENABLED", "true")
	defer os.Unsetenv("SSR_ENABLED")

	config := LoadSSRConfig()

	if !config.EnableSSR {
		t.Error("EnableSSR should be true")
	}
}

func TestLoadSSRConfig_EnableSSR_1(t *testing.T) {
	clearSSREnv()
	os.Setenv("SSR_ENABLED", "1")
	defer os.Unsetenv("SSR_ENABLED")

	config := LoadSSRConfig()

	if !config.EnableSSR {
		t.Error("EnableSSR should be true for '1'")
	}
}

func TestLoadSSRConfig_EnableSSR_Yes(t *testing.T) {
	clearSSREnv()
	os.Setenv("SSR_ENABLED", "yes")
	defer os.Unsetenv("SSR_ENABLED")

	config := LoadSSRConfig()

	if !config.EnableSSR {
		t.Error("EnableSSR should be true for 'yes'")
	}
}

func TestLoadSSRConfig_EnableSSR_False(t *testing.T) {
	clearSSREnv()
	os.Setenv("SSR_ENABLED", "false")
	defer os.Unsetenv("SSR_ENABLED")

	config := LoadSSRConfig()

	if config.EnableSSR {
		t.Error("EnableSSR should be false")
	}
}

func TestLoadSSRConfig_EnableSSR_Invalid(t *testing.T) {
	clearSSREnv()
	os.Setenv("SSR_ENABLED", "invalid")
	defer os.Unsetenv("SSR_ENABLED")

	config := LoadSSRConfig()

	if config.EnableSSR {
		t.Error("EnableSSR should be false for invalid value")
	}
}

func TestLoadSSRConfig_CustomServerURL(t *testing.T) {
	clearSSREnv()
	customURL := "https://ssr.example.com:8080"
	os.Setenv("SSR_SERVER_URL", customURL)
	defer os.Unsetenv("SSR_SERVER_URL")

	config := LoadSSRConfig()

	if config.SSRServerURL != customURL {
		t.Errorf("SSRServerURL = %q, want %q", config.SSRServerURL, customURL)
	}
}

func TestLoadSSRConfig_CustomPort(t *testing.T) {
	clearSSREnv()
	customPort := "9000"
	os.Setenv("SSR_PORT", customPort)
	defer os.Unsetenv("SSR_PORT")

	config := LoadSSRConfig()

	if config.SSRPort != customPort {
		t.Errorf("SSRPort = %q, want %q", config.SSRPort, customPort)
	}
}

func TestLoadSSRConfig_AllCustom(t *testing.T) {
	clearSSREnv()
	os.Setenv("SSR_ENABLED", "true")
	os.Setenv("SSR_SERVER_URL", "https://custom-ssr.example.com")
	os.Setenv("SSR_PORT", "4000")
	defer clearSSREnv()

	config := LoadSSRConfig()

	if !config.EnableSSR {
		t.Error("EnableSSR should be true")
	}
	if config.SSRServerURL != "https://custom-ssr.example.com" {
		t.Errorf("SSRServerURL = %q, want custom URL", config.SSRServerURL)
	}
	if config.SSRPort != "4000" {
		t.Errorf("SSRPort = %q, want \"4000\"", config.SSRPort)
	}
}

func TestSSRConfig_Struct(t *testing.T) {
	config := SSRConfig{
		EnableSSR:    true,
		SSRServerURL: "http://localhost:3001",
		SSRPort:      "3001",
	}

	if !config.EnableSSR {
		t.Error("EnableSSR should be true")
	}
	if config.SSRServerURL != "http://localhost:3001" {
		t.Errorf("SSRServerURL = %q, want default URL", config.SSRServerURL)
	}
}

func clearSSREnv() {
	os.Unsetenv("SSR_ENABLED")
	os.Unsetenv("SSR_SERVER_URL")
	os.Unsetenv("SSR_PORT")
}
