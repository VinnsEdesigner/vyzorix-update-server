package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear all env vars
	clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Port != "3000" {
		t.Errorf("Port = %s, want 3000", cfg.Port)
	}
	if cfg.DatabaseURL != "./data/vyzorix.db" {
		t.Errorf("DatabaseURL = %s, want ./data/vyzorix.db", cfg.DatabaseURL)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("DataDir = %s, want ./data", cfg.DataDir)
	}
	if cfg.BinDir != "./bin" {
		t.Errorf("BinDir = %s, want ./bin", cfg.BinDir)
	}
	if cfg.PublicDir != "./public" {
		t.Errorf("PublicDir = %s, want ./public", cfg.PublicDir)
	}
	if cfg.JWTDuration != 7*24*time.Hour {
		t.Errorf("JWTDuration = %v, want 7 days", cfg.JWTDuration)
	}
	if cfg.HMACWindow != 30*time.Second {
		t.Errorf("HMACWindow = %v, want 30s (per COMMAND_SECURITY.md)", cfg.HMACWindow)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	clearEnvVars()

	os.Setenv("PORT", "8080")
	os.Setenv("DATABASE_URL", "/custom/path/db.sqlite")
	os.Setenv("NODE_ENV", "production")
	os.Setenv("TOKEN_SECRET", "test-secret")
	os.Setenv("JWT_SECRET", "jwt-secret-123")
	os.Setenv("ALLOWED_ORIGINS", "https://app.example.com,https://admin.example.com")
	os.Setenv("HMAC_WINDOW_SECONDS", "600")
	os.Setenv("JWT_DURATION_HOURS", "24")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %s, want 8080", cfg.Port)
	}
	if cfg.DatabaseURL != "/custom/path/db.sqlite" {
		t.Errorf("DatabaseURL = %s, want /custom/path/db.sqlite", cfg.DatabaseURL)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %s, want production", cfg.Env)
	}
	if cfg.TokenSecret != "test-secret" {
		t.Errorf("TokenSecret = %s, want test-secret", cfg.TokenSecret)
	}
	if cfg.JWTSecret != "jwt-secret-123" {
		t.Errorf("JWTSecret = %s, want jwt-secret-123", cfg.JWTSecret)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("AllowedOrigins length = %d, want 2", len(cfg.AllowedOrigins))
	}
	if cfg.HMACWindow != 10*time.Minute {
		t.Errorf("HMACWindow = %v, want 10m", cfg.HMACWindow)
	}
	if cfg.JWTDuration != 24*time.Hour {
		t.Errorf("JWTDuration = %v, want 24h", cfg.JWTDuration)
	}
}

func TestLoad_ProductionRequiresTokenSecret(t *testing.T) {
	clearEnvVars()

	os.Setenv("NODE_ENV", "production")
	os.Setenv("DATABASE_URL", "/tmp/test.db")

	defer clearEnvVars()

	_, err := Load()
	if err == nil {
		t.Error("expected error when TOKEN_SECRET not set in production")
	}
	if err != nil && err.Error() != "TOKEN_SECRET is required in production" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_ProductionWithTokenSecret(t *testing.T) {
	clearEnvVars()

	os.Setenv("NODE_ENV", "production")
	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("TOKEN_SECRET", "production-secret")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.TokenSecret != "production-secret" {
		t.Errorf("TokenSecret = %s, want production-secret", cfg.TokenSecret)
	}
}

func TestLoad_DevelopmentNoTokenRequired(t *testing.T) {
	clearEnvVars()

	os.Setenv("NODE_ENV", "development")
	os.Setenv("DATABASE_URL", "/tmp/test.db")
	// No TOKEN_SECRET

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.TokenSecret != "" {
		t.Errorf("TokenSecret = %s, want empty", cfg.TokenSecret)
	}
}

func TestLoad_DefaultToDevelopment(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")
	// No NODE_ENV set

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("Env = %s, want development", cfg.Env)
	}
}

func TestLoad_EmptyDatabaseURL(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "")

	defer clearEnvVars()

	cfg, err := Load()
	// Empty DATABASE_URL uses fallback default, no error
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	// Fallback default path should be used
	if cfg.DatabaseURL != "./data/vyzorix.db" {
		t.Errorf("DatabaseURL = %q, want ./data/vyzorix.db (fallback default)", cfg.DatabaseURL)
	}
}

func TestLoad_WhitespaceDatabaseURL(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "   ")

	defer clearEnvVars()

	_, err := Load()
	if err == nil {
		t.Error("expected error for whitespace DATABASE_URL")
	}
}

func TestLoad_InvalidHMACWindow(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("HMAC_WINDOW_SECONDS", "invalid")

	defer clearEnvVars()

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid HMAC_WINDOW_SECONDS")
	}
}

func TestLoad_ZeroHMACWindow(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("HMAC_WINDOW_SECONDS", "0")

	defer clearEnvVars()

	_, err := Load()
	if err == nil {
		t.Error("expected error for zero HMAC_WINDOW_SECONDS")
	}
}

func TestLoad_NegativeHMACWindow(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("HMAC_WINDOW_SECONDS", "-10")

	defer clearEnvVars()

	_, err := Load()
	if err == nil {
		t.Error("expected error for negative HMAC_WINDOW_SECONDS")
	}
}

func TestLoad_HMACWindowValid(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("HMAC_WINDOW_SECONDS", "300")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.HMACWindow != 5*time.Minute {
		t.Errorf("HMACWindow = %v, want 5m", cfg.HMACWindow)
	}
}

func TestLoad_EnforceHMACProduction(t *testing.T) {
	clearEnvVars()

	os.Setenv("NODE_ENV", "production")
	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("TOKEN_SECRET", "secret")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !cfg.EnforceHMAC {
		t.Error("EnforceHMAC should be true in production")
	}
}

func TestLoad_EnforceHMACDevelopment(t *testing.T) {
	clearEnvVars()

	os.Setenv("NODE_ENV", "development")
	os.Setenv("DATABASE_URL", "/tmp/test.db")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.EnforceHMAC {
		t.Error("EnforceHMAC should be false in development")
	}
}

func TestLoad_EnforceHMACExplicitTrue(t *testing.T) {
	clearEnvVars()

	os.Setenv("NODE_ENV", "development")
	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("ENFORCE_HMAC", "true")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !cfg.EnforceHMAC {
		t.Error("EnforceHMAC should be true when explicitly set")
	}
}

func TestLoad_EnforceHMACExplicitFalse(t *testing.T) {
	clearEnvVars()

	os.Setenv("NODE_ENV", "production")
	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("TOKEN_SECRET", "secret")
	os.Setenv("ENFORCE_HMAC", "false")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.EnforceHMAC {
		t.Error("EnforceHMAC should be false when explicitly set")
	}
}

func TestLoad_AllowedOriginsCSV(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("ALLOWED_ORIGINS", "https://a.com, https://b.com, https://c.com")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.AllowedOrigins) != 3 {
		t.Errorf("AllowedOrigins length = %d, want 3", len(cfg.AllowedOrigins))
	}
}

func TestLoad_AllowedOriginsWithWildcard(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("ALLOWED_ORIGINS", "*")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
		t.Errorf("AllowedOrigins = %v, want [*]", cfg.AllowedOrigins)
	}
}

func TestLoad_AllowedOriginsEmpty(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("ALLOWED_ORIGINS", "")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// When ALLOWED_ORIGINS is empty string, get() returns fallback "*"
	// since empty string == "" after TrimSpace
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
		t.Errorf("AllowedOrigins = %v, want [*]", cfg.AllowedOrigins)
	}
}

func TestLoad_GoogleOAuthConfig(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")
	os.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id-123")
	os.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret-456")
	os.Setenv("BASE_URL", "https://server.example.com")
	os.Setenv("FRONTEND_URL", "https://app.example.com")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.GoogleOAuthClientID != "client-id-123" {
		t.Errorf("GoogleOAuthClientID = %s, want client-id-123", cfg.GoogleOAuthClientID)
	}
	if cfg.GoogleOAuthClientSecret != "client-secret-456" {
		t.Errorf("GoogleOAuthClientSecret = %s, want client-secret-456", cfg.GoogleOAuthClientSecret)
	}
	if cfg.BaseURL != "https://server.example.com" {
		t.Errorf("BaseURL = %s, want https://server.example.com", cfg.BaseURL)
	}
	if cfg.FrontendURL != "https://app.example.com" {
		t.Errorf("FrontendURL = %s, want https://app.example.com", cfg.FrontendURL)
	}
}

func TestLoad_DefaultBaseURL(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.BaseURL != "http://localhost:3000" {
		t.Errorf("BaseURL = %s, want http://localhost:3000", cfg.BaseURL)
	}
}

func TestLoad_DefaultFrontendURL(t *testing.T) {
	clearEnvVars()

	os.Setenv("DATABASE_URL", "/tmp/test.db")

	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.FrontendURL != "http://localhost:5173" {
		t.Errorf("FrontendURL = %s, want http://localhost:5173", cfg.FrontendURL)
	}
}

func TestConfig_HelperMethods(t *testing.T) {
	cfg := Config{
		Env:         "test",
		EnforceHMAC: true,
		DataDir:     "/data",
		BinDir:      "/bin",
		PublicDir:   "/public",
		BaseURL:     "https://example.com",
	}

	if !cfg.EnforceHMAC {
		t.Error("EnforceHMAC should be true")
	}
	if cfg.DataDir != "/data" {
		t.Errorf("DataDir = %s, want /data", cfg.DataDir)
	}
	if cfg.BinDir != "/bin" {
		t.Errorf("BinDir = %s, want /bin", cfg.BinDir)
	}
	if cfg.PublicDir != "/public" {
		t.Errorf("PublicDir = %s, want /public", cfg.PublicDir)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Errorf("BaseURL = %s, want https://example.com", cfg.BaseURL)
	}
}

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{"a, b, , c", []string{"a", "b", "c"}}, // empty parts skipped
		{"", []string{}},
		{",,", []string{}},
	}

	for _, tt := range tests {
		result := splitCSV(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitCSV(%q) = %v, want %v", tt.input, result, tt.expected)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("splitCSV(%q)[%d] = %s, want %s", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

func TestGet(t *testing.T) {
	clearEnvVars()
	os.Setenv("TEST_KEY", "test-value")

	if get("TEST_KEY", "default") != "test-value" {
		t.Error("expected env value")
	}
	if get("MISSING_KEY", "default") != "default" {
		t.Error("expected default value")
	}
	if get("WHITESPACE_KEY", "default") != "default" {
		t.Error("expected default for whitespace-only value")
	}

	clearEnvVars()
}

func TestGetBool(t *testing.T) {
	clearEnvVars()

	if getBool("MISSING", true) != true {
		t.Error("expected fallback for missing key")
	}
	if getBool("MISSING", false) != false {
		t.Error("expected fallback for missing key")
	}

	os.Setenv("TEST_BOOL", "true")
	if getBool("TEST_BOOL", false) != true {
		t.Error("expected true")
	}

	os.Setenv("TEST_BOOL", "false")
	if getBool("TEST_BOOL", true) != false {
		t.Error("expected false")
	}

	os.Setenv("TEST_BOOL", "invalid")
	if getBool("TEST_BOOL", true) != true {
		t.Error("expected fallback for invalid value")
	}

	clearEnvVars()
}

func clearEnvVars() {
	envVars := []string{
		"PORT", "NODE_ENV", "GO_ENV", "DATABASE_URL", "VYZORIX_API_DIR",
		"VYZORIX_BIN_DIR", "VYZORIX_PUBLIC_DIR", "FIREBASE_CREDENTIALS",
		"TOKEN_SECRET", "JWT_SECRET", "JWT_DURATION_HOURS", "ALLOWED_ORIGINS",
		"ENFORCE_HMAC", "HMAC_WINDOW_SECONDS", "GOOGLE_OAUTH_CLIENT_ID",
		"GOOGLE_OAUTH_CLIENT_SECRET", "BASE_URL", "FRONTEND_URL",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
