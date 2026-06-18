package config

import (
	"os"
	"testing"
	"time"
)

func TestValidator_ValidatePort(t *testing.T) {
	tests := []struct {
		name      string
		port      string
		wantError bool
	}{
		{"valid port", "3000", false},
		{"valid port min", "1", false},
		{"valid port max", "65535", false},
		{"empty port", "", true},
		{"invalid port string", "abc", true},
		{"port too low", "0", true},
		{"port too high", "65536", true},
		{"negative port", "-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Port: tt.port, Env: "development"}
			v := NewValidator(cfg)
			v.validatePort()
			hasError := len(v.errs) > 0
			if hasError != tt.wantError {
				t.Errorf("port %q: got error=%v, want error=%v", tt.port, hasError, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateSecrets(t *testing.T) {
	tests := []struct {
		name           string
		env            string
		tokenSecret    string
		jwtSecret      string
		sessionSecret  string
		wantErrorCount int
	}{
		{"production missing TOKEN_SECRET", "production", "", "somejwtsomesecret32charslongenough!!", "", 1},
		{"production missing JWT_SECRET", "production", "sometokenvaluemorethan32chars!!!!", "", "", 1},
		{"production short TOKEN_SECRET", "production", "short", "somejwtsomesecret32charslongenough!!", "", 1},
		{"production short JWT_SECRET", "production", "sometokenvaluemorethan32chars!!!!", "short", "", 1},
		{"production valid secrets", "production", "sometokenvaluemorethan32chars!!!!", "somejwtsomesecret32charslongenough!!", "", 0},
		{"production SESSION_SECRET wrong length", "production", "sometokenvaluemorethan32chars!!!!", "somejwtsomesecret32charslongenough!!", "exactly32characterslong", 1}, // 27 chars - too short
		{"development missing secrets ok", "development", "", "", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Port:         "3000",
				Env:          tt.env,
				TokenSecret:  tt.tokenSecret,
				JWTSecret:    tt.jwtSecret,
				SessionSecret: tt.sessionSecret,
			}
			v := NewValidator(cfg)
			v.validateSecrets()
			if len(v.errs) != tt.wantErrorCount {
				t.Errorf("got %d errors, want %d: %v", len(v.errs), tt.wantErrorCount, v.errs)
			}
		})
	}
}

func TestValidator_ValidateOrigins(t *testing.T) {
	tests := []struct {
		name           string
		origins        []string
		env            string
		wantErrorCount int
	}{
		{"valid single origin", []string{"https://example.com"}, "production", 0},
		{"valid multiple origins", []string{"https://example.com", "https://app.example.com"}, "production", 0},
		{"wildcard in development ok", []string{"*"}, "development", 0},
		{"wildcard in production error", []string{"*"}, "production", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{AllowedOrigins: tt.origins, Env: tt.env}
			v := NewValidator(cfg)
			v.validateOrigins()
			if len(v.errs) != tt.wantErrorCount {
				t.Errorf("got %d errors, want %d: %v", len(v.errs), tt.wantErrorCount, v.errs)
			}
		})
	}
}

func TestValidator_ValidateURLs(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		frontendURL string
		wantErrors  int
	}{
		{"valid URLs", "https://api.example.com", "https://app.example.com", 0},
		{"empty URLs ok", "", "", 0},
		{"invalid base URL", "http://[invalid/", "", 1},
		{"invalid frontend URL", "", "http://[invalid/", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{BaseURL: tt.baseURL, FrontendURL: tt.frontendURL}
			v := NewValidator(cfg)
			v.validateURLs()
			if len(v.errs) != tt.wantErrors {
				t.Errorf("got %d errors, want %d: %v", len(v.errs), tt.wantErrors, v.errs)
			}
		})
	}
}

func TestValidator_ValidateSigning(t *testing.T) {
	tests := []struct {
		name        string
		signingEnv  string
		windowEnv   string
		cacheEnv    string
		graceEnv    string
		wantErrors  int
	}{
		{"signing disabled", "false", "", "", "", 0},
		{"signing enabled defaults ok", "true", "", "", "", 0},
		{"valid window", "true", "300", "", "", 0},
		{"invalid window too low", "true", "30", "", "", 1},
		{"invalid window too high", "true", "900", "", "", 1},
		{"invalid cache size", "true", "", "0", "", 1},
		{"invalid grace period", "true", "", "", "-1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("REQUEST_SIGNING_ENABLED", tt.signingEnv)
			if tt.windowEnv != "" {
				os.Setenv("SIGNING_TIMESTAMP_WINDOW", tt.windowEnv)
			} else {
				os.Unsetenv("SIGNING_TIMESTAMP_WINDOW")
			}
			if tt.cacheEnv != "" {
				os.Setenv("SIGNING_MAX_CACHE_SIZE", tt.cacheEnv)
			} else {
				os.Unsetenv("SIGNING_MAX_CACHE_SIZE")
			}
			if tt.graceEnv != "" {
				os.Setenv("SIGNING_GRACE_PERIOD", tt.graceEnv)
			} else {
				os.Unsetenv("SIGNING_GRACE_PERIOD")
			}

			cfg := Config{Env: "development"}
			v := NewValidator(cfg)
			v.validateSigning()

			if len(v.errs) != tt.wantErrors {
				t.Errorf("got %d errors, want %d: %v", len(v.errs), tt.wantErrors, v.errs)
			}

			os.Unsetenv("REQUEST_SIGNING_ENABLED")
			os.Unsetenv("SIGNING_TIMESTAMP_WINDOW")
			os.Unsetenv("SIGNING_MAX_CACHE_SIZE")
			os.Unsetenv("SIGNING_GRACE_PERIOD")
		})
	}
}

func TestValidator_ValidateCSRF(t *testing.T) {
	tests := []struct {
		name        string
		csrfEnv     string
		secretEnv   string
		wantErrors  int
	}{
		{"csrf disabled", "false", "", 0},
		{"csrf enabled no secret", "true", "", 1},
		{"csrf enabled short secret", "true", "short", 1},
		{"csrf enabled valid secret", "true", "this-is-a-32-char-secret-key!!!!", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CSRF_ENABLED", tt.csrfEnv)
			os.Setenv("CSRF_SECRET", tt.secretEnv)

			cfg := Config{}
			v := NewValidator(cfg)
			v.validateCSRF()

			if len(v.errs) != tt.wantErrors {
				t.Errorf("got %d errors, want %d: %v", len(v.errs), tt.wantErrors, v.errs)
			}

			os.Unsetenv("CSRF_ENABLED")
			os.Unsetenv("CSRF_SECRET")
		})
	}
}

func TestValidator_ValidateSession(t *testing.T) {
	tests := []struct {
		name           string
		sessionMaxAge  int
		wantErrorCount int
	}{
		{"valid session max age", 86400, 0},
		{"too short", 60, 1},
		{"zero", 0, 1},
		{"negative", -1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{SessionMaxAge: tt.sessionMaxAge}
			v := NewValidator(cfg)
			v.validateSession()
			if len(v.errs) != tt.wantErrorCount {
				t.Errorf("got %d errors, want %d: %v", len(v.errs), tt.wantErrorCount, v.errs)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{Field: "PORT", Message: "must be a valid number"}
	if got := err.Error(); got != "PORT: must be a valid number" {
		t.Errorf("ValidationError.Error() = %q, want %q", got, "PORT: must be a valid number")
	}
}

func TestValidateConfig(t *testing.T) {
	// Valid config should pass
	cfg := Config{
		Port:                       "3000",
		Env:                        "development",
		TokenSecret:                "sometokenvaluemorethan32chars!!!!",
		JWTSecret:                  "somejwtsomesecret32charslongenough!!",
		SessionMaxAge:              86400,
		AllowedOrigins:             []string{"http://localhost:5173"},
		JWTDuration:                168 * time.Hour,
		EmailVerifyTokenExpiry:     24 * time.Hour,
		PasswordResetTokenExpiry:   60 * time.Minute,
	}
	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("expected valid config to pass, got error: %v", err)
	}

	// Invalid config should fail
	invalidCfg := Config{Port: "invalid"}
	err = ValidateConfig(invalidCfg)
	if err == nil {
		t.Error("expected invalid config to fail")
	}
}

func TestValidateConfig_Production(t *testing.T) {
	os.Setenv("TOKEN_SECRET", "short")
	os.Setenv("JWT_SECRET", "also-short")

	cfg := Config{
		Port:           "3000",
		Env:            "production",
		TokenSecret:    "short",
		JWTSecret:      "also-short",
		SessionMaxAge:  86400,
		AllowedOrigins: []string{"*"},
	}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Error("expected production config with issues to fail")
	}

	os.Unsetenv("TOKEN_SECRET")
	os.Unsetenv("JWT_SECRET")
}