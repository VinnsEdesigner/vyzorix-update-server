package config

import "testing"

const (
	goodSecretA = "9f2c1a7e5b3d8046c1e9f7a2b4d60813579bdf02468ace13579bdf02468ace13"
	goodSecretB = "3b7d1f9a2c4e6081d5f7a9b1c3e50729486ace0246813579bdf013579bdf0246"
	goodSecretC = "c1e3f5a7b9d0246813579bdf02468ace1357911bdf3579ac02468ace13579bd1"
)

// validProd returns a Config that passes every rule, so each test can break one
// thing and assert exactly that failure.
func validProd() *Config {
	return &Config{
		Env:                      "production",
		Port:                     "3000",
		BindAddr:                 "0.0.0.0",
		DatabaseURL:              "/data/vyzorix.db",
		DatabaseBackend:          "sqlite",
		DatabaseRequestTimeout:   15e9,
		JWTSecret:                goodSecretA,
		SessionSecret:            goodSecretB,
		ServerAPIToken:           goodSecretC,
		APIKeys:                  map[string]string{"primary": "vxyz-key"},
		JWTDuration:              168e9,
		SessionMaxAge:            3600,
		RateLimitPerMin:          100,
		AuthRateLimitMin:         60,
		EmailVerifyTokenExpiry:   1e12,
		PasswordResetTokenExpiry: 1e12,
	}
}

func fieldsWith(errs []ValidationError, field string) int {
	n := 0
	for _, e := range errs {
		if e.Field == field {
			n++
		}
	}
	return n
}

func TestValidateConfig_AcceptsValidProduction(t *testing.T) {
	if errs := ValidateConfig(validProd()); len(errs) != 0 {
		t.Fatalf("valid production config should pass, got: %v", errs)
	}
}

func TestValidateConfig_RequiresCoreInputs(t *testing.T) {
	c := validProd()
	c.DatabaseURL = ""
	c.APIKeys = nil

	errs := ValidateConfig(c)
	if fieldsWith(errs, "DATABASE_URL") == 0 {
		t.Error("missing DATABASE_URL should be reported")
	}
	if fieldsWith(errs, "API_KEY_*") == 0 {
		t.Error("missing API key should be reported")
	}
}

// TestValidateConfig_RejectsPlaceholders is the core fix over the old
// length-only check: a long but publicly-known placeholder must be refused.
func TestValidateConfig_RejectsPlaceholders(t *testing.T) {
	for _, placeholder := range []string{
		"change-me-32-chars-minimum!!!!!!",
		"CHANGE-ME-32-CHARS-MINIMUM!!!!!!",
		"replace-with-a-random-64-character-api-key-xxxxx",
		"your-secret-here-please-change-me-32-chars",
		"example-jwt-secret-value-for-testing-123456",
	} {
		c := validProd()
		c.JWTSecret = placeholder
		if errs := ValidateConfig(c); fieldsWith(errs, "JWT_SECRET") == 0 {
			t.Errorf("placeholder %q must be rejected", placeholder)
		}
	}
}

func TestValidateConfig_RejectsShortAndLowEntropy(t *testing.T) {
	c := validProd()
	c.JWTSecret = "tooshort"
	if errs := ValidateConfig(c); fieldsWith(errs, "JWT_SECRET") == 0 {
		t.Error("a short secret must be rejected")
	}

	c = validProd()
	c.JWTSecret = "abababababababababababababababababababab"
	if errs := ValidateConfig(c); fieldsWith(errs, "JWT_SECRET") == 0 {
		t.Error("a low-diversity secret must be rejected")
	}
}

func TestValidateConfig_RejectsReusedSecret(t *testing.T) {
	c := validProd()
	c.SessionSecret = c.JWTSecret
	if errs := ValidateConfig(c); fieldsWith(errs, "SESSION_SECRET") == 0 {
		t.Error("reusing JWT_SECRET as SESSION_SECRET must be rejected")
	}
}

// TestValidateConfig_SessionSecretAnyLength documents the corrected rule: the
// session key is sha512(secret) truncated, so any length >= 32 is valid (the old
// validator wrongly demanded exactly 32).
func TestValidateConfig_SessionSecretAnyLength(t *testing.T) {
	c := validProd()
	c.SessionSecret = goodSecretA + goodSecretB // 128 chars
	if errs := ValidateConfig(c); len(errs) != 0 {
		t.Fatalf("a long SESSION_SECRET is valid, got: %v", errs)
	}
}

func TestValidateConfig_OptionalSecretsSkippedWhenUnset(t *testing.T) {
	c := validProd()
	c.DeviceSecret = ""
	if errs := ValidateConfig(c); len(errs) != 0 {
		t.Fatalf("unset DEVICE_SECRET is fine, got: %v", errs)
	}

	c.DeviceSecret = "change-me-32-chars-minimum!!!!!!"
	if errs := ValidateConfig(c); fieldsWith(errs, "DEVICE_SECRET") == 0 {
		t.Error("a weak DEVICE_SECRET must be rejected when set")
	}
}

func TestValidateConfig_RejectsBadPortAndOrigin(t *testing.T) {
	c := validProd()
	c.Port = "99999"
	c.AllowedOrigins = []string{"not a url"}
	errs := ValidateConfig(c)
	if fieldsWith(errs, "PORT") == 0 {
		t.Error("out-of-range PORT should be reported")
	}
	if fieldsWith(errs, "ALLOWED_ORIGINS") == 0 {
		t.Error("an invalid origin should be reported")
	}
}

func TestValidateConfig_RejectsWildcardOriginInProduction(t *testing.T) {
	c := validProd()
	c.AllowedOrigins = []string{"*"}
	if errs := ValidateConfig(c); fieldsWith(errs, "ALLOWED_ORIGINS") == 0 {
		t.Error("wildcard origin must be rejected in production")
	}
}

func TestValidateConfig_RejectsBadMetricsCIDR(t *testing.T) {
	c := validProd()
	c.MetricsAllowedCIDRs = []string{"10.0.0.0/8", "not-a-cidr"}
	errs := ValidateConfig(c)
	if fieldsWith(errs, "METRICS_ALLOWED_CIDRS") != 1 {
		t.Fatalf("expected exactly one CIDR error, got: %v", errs)
	}
}

func TestValidateConfig_TursoRequiresCredentials(t *testing.T) {
	c := validProd()
	c.DatabaseBackend = "turso"
	errs := ValidateConfig(c)
	if fieldsWith(errs, "TURSO_DB_URL") == 0 {
		t.Error("turso without TURSO_DB_URL should be reported")
	}
	if fieldsWith(errs, "TURSO_AUTH_TOKEN") == 0 {
		t.Error("turso without TURSO_AUTH_TOKEN should be reported")
	}
}

// TestValidateConfig_SecretStrengthIsProductionOnly proves development is not
// blocked by secret-strength rules, so it can boot with throwaway values.
func TestValidateConfig_SecretStrengthIsProductionOnly(t *testing.T) {
	c := validProd()
	c.Env = "development"
	c.JWTSecret = "change-me-32-chars-minimum!!!!!!"
	c.SessionSecret = "change-me-32-chars-minimum!!!!!!"
	if errs := ValidateConfig(c); len(errs) != 0 {
		t.Fatalf("development must not enforce secret strength, got: %v", errs)
	}
}

// TestValidateConfig_ReportsAllProblems proves the validator accumulates errors
// instead of stopping at the first, so one pass finds everything.
func TestValidateConfig_ReportsAllProblems(t *testing.T) {
	c := validProd()
	c.Port = "0"
	c.JWTSecret = "change-me"
	c.AllowedOrigins = []string{"*"}

	errs := ValidateConfig(c)
	if len(errs) < 3 {
		t.Fatalf("expected several errors at once, got %d: %v", len(errs), errs)
	}
}
