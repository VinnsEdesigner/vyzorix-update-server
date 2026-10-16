// Package config provides configuration validation from environment variables.
package config

import (
	"fmt"
	"math"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ValidationError is a single configuration problem, with the field it belongs to.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ConfigValidator validates a resolved Config.
//
// It validates the Config value the application will actually use, not the raw
// environment: reading os.Getenv here would let the validator disagree with the
// running config (defaults, parsing, derived values). Every rule therefore reads
// from cfg.
type ConfigValidator struct {
	cfg    *Config
	ssrCfg *SSRConfig
	errs   []ValidationError
}

// ValidateConfig validates cfg and returns every problem found, so an operator
// can fix them in one pass rather than one restart at a time.
func ValidateConfig(cfg *Config) []ValidationError {
	v := &ConfigValidator{cfg: cfg}
	v.validateRequired()
	v.validatePort()
	v.validateSecrets()
	v.validateDatabase()
	v.validateOrigins()
	v.validateURLs()
	v.validateSession()
	v.validateRateLimits()
	v.validateMetrics()
	v.validateTimeWindows()
	return v.errs
}

// ValidateConfigWithSSR additionally validates SSR configuration.
func ValidateConfigWithSSR(cfg *Config, ssrCfg SSRConfig) []ValidationError {
	v := &ConfigValidator{cfg: cfg, ssrCfg: &ssrCfg}
	errs := ValidateConfig(cfg)
	v.validateSSR()
	return append(errs, v.errs...)
}

// Err returns a single error joining all problems, or nil when the config is valid.
func Err(errs []ValidationError) error {
	if len(errs) == 0 {
		return nil
	}
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, e.Error())
	}
	return fmt.Errorf("configuration validation failed: %s", strings.Join(msgs, "; "))
}

// Secret-strength thresholds. 32 characters of real entropy is the conventional
// bar for an HMAC/JWT signing key; a random 32-char hex string is ~128 bits.
const (
	minSecretLength        = 32
	minSecretDistinctRunes = 10
	minSecretEntropyBits   = 100
)

// placeholderFragments mark a value as a shipped default rather than a generated
// secret. Comparison is case-insensitive.
var placeholderFragments = []string{
	"change-me", "changeme", "change_me",
	"replace-me", "replace_me", "replace-with", "replacewith",
	"your-secret", "yoursecret", "your_secret",
	"example", "placeholder", "dummy", "sample",
	"todo", "fixme", "xxx", "password", "secret123",
	"insecure", "default-secret", "test-secret", "dev-secret",
}

// distinctRunes counts unique runes in s.
func distinctRunes(s string) int {
	set := make(map[rune]struct{}, len(s))
	for _, r := range s {
		set[r] = struct{}{}
	}
	return len(set)
}

// shannonEntropyBits returns the total Shannon entropy of s in bits: per-symbol
// entropy times symbol count. Penalises low diversity and heavy repetition.
func shannonEntropyBits(s string) float64 {
	if s == "" {
		return 0
	}
	counts := make(map[rune]int, len(s))
	for _, r := range s {
		counts[r]++
	}
	total := float64(len([]rune(s)))
	var perSymbol float64
	for _, n := range counts {
		p := float64(n) / total
		perSymbol -= p * math.Log2(p)
	}
	return perSymbol * total
}

// isValidCIDROrIP reports whether raw is a CIDR block or a bare IP address.
func isValidCIDROrIP(raw string) bool {
	if _, _, err := net.ParseCIDR(raw); err == nil {
		return true
	}
	return net.ParseIP(raw) != nil
}

func (v *ConfigValidator) addf(field, format string, args ...any) {
	v.errs = append(v.errs, ValidationError{Field: field, Message: fmt.Sprintf(format, args...)})
}

// validateRequired checks the inputs the server cannot start without, in any
// environment.
func (v *ConfigValidator) validateRequired() {
	if strings.TrimSpace(v.cfg.DatabaseURL) == "" {
		v.addf("DATABASE_URL", "is required")
	}
	if len(v.cfg.APIKeys) == 0 {
		v.addf("API_KEY_*", "at least one API key is required (e.g. API_KEY_primary)")
	}
}

func (v *ConfigValidator) validatePort() {
	port, err := strconv.Atoi(v.cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		v.addf("PORT", "must be an integer between 1 and 65535, got %q", v.cfg.Port)
	}
	if v.cfg.BindAddr == "" {
		v.addf("BIND_ADDR", "must not be empty")
	}
}

// validateSecrets checks presence and strength of the signing/authentication
// secrets. Both are production-only so local development can boot with throwaway
// values.
func (v *ConfigValidator) validateSecrets() {
	// Secret strength is a production requirement: local development must be able
	// to boot with throwaway values. A set secret is validated wherever it is set,
	// so a weak secret is still caught before it can reach production.
	if v.cfg.Env != "production" {
		return
	}

	required := []struct {
		name  string
		value string
	}{
		{"JWT_SECRET", v.cfg.JWTSecret},
		{"SESSION_SECRET", v.cfg.SessionSecret},
		{"SERVER_API_TOKEN", v.cfg.ServerAPIToken},
	}
	optional := []struct {
		name  string
		value string
	}{
		{"DEVICE_SECRET", v.cfg.DeviceSecret},
	}

	for _, s := range required {
		if s.value == "" {
			v.addf(s.name, "is required in production")
			continue
		}
		v.checkSecretStrength(s.name, s.value)
	}
	for _, s := range optional {
		if s.value != "" {
			v.checkSecretStrength(s.name, s.value)
		}
	}

	// Reusing one secret for two purposes means a single leak compromises both.
	seen := make(map[string]string)
	for _, s := range append(append([]struct {
		name  string
		value string
	}{}, required...), optional...) {
		if s.value == "" {
			continue
		}
		if prev, dup := seen[s.value]; dup {
			v.addf(s.name, "must not reuse the value of %s", prev)
		}
		seen[s.value] = s.name
	}
}

// checkSecretStrength rejects a secret that is too short, a known placeholder, or
// too low-entropy to be a real key.
func (v *ConfigValidator) checkSecretStrength(name, value string) {
	if len(value) < minSecretLength {
		v.addf(name, "must be at least %d characters (got %d); generate one with `openssl rand -hex 32`",
			minSecretLength, len(value))
		return
	}

	lower := strings.ToLower(value)
	for _, frag := range placeholderFragments {
		if strings.Contains(lower, frag) {
			v.addf(name, "looks like a placeholder (contains %q); generate one with `openssl rand -hex 32`", frag)
			return
		}
	}

	if n := distinctRunes(value); n < minSecretDistinctRunes {
		v.addf(name, "has only %d distinct characters; generate one with `openssl rand -hex 32`", n)
		return
	}

	if bits := shannonEntropyBits(value); bits < minSecretEntropyBits {
		v.addf(name, "has only ~%.0f bits of entropy (need >= %d); generate one with `openssl rand -hex 32`",
			bits, minSecretEntropyBits)
	}
}

// validateDatabase checks the database backend selection and its connection settings.
func (v *ConfigValidator) validateDatabase() {
	switch v.cfg.ResolvedDatabaseBackend() {
	case "turso":
		if v.cfg.TursoDatabaseURL == "" {
			v.addf("TURSO_DB_URL", "is required when DATABASE_BACKEND resolves to turso")
		}
		if v.cfg.TursoAuthToken == "" {
			v.addf("TURSO_AUTH_TOKEN", "is required when DATABASE_BACKEND resolves to turso")
		}
		if u := v.cfg.TursoDatabaseURL; u != "" {
			ok := strings.HasPrefix(u, "libsql://") || strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")
			if !ok {
				v.addf("TURSO_DB_URL", "must use the libsql://, https://, or http:// scheme")
			} else if v.cfg.Env == "production" && strings.HasPrefix(u, "http://") {
				v.addf("TURSO_DB_URL", "must use libsql:// or https:// in production (http:// is plaintext)")
			}
		}
	case "sqlite":
		if strings.TrimSpace(v.cfg.DatabaseURL) == "" {
			v.addf("DATABASE_URL", "is required when DATABASE_BACKEND resolves to sqlite")
		}
	default:
		v.addf("DATABASE_BACKEND", "must be one of auto, sqlite, turso (got %q)", v.cfg.DatabaseBackend)
	}

	if v.cfg.DatabaseMaxOpenConns < 0 {
		v.addf("DATABASE_MAX_OPEN_CONNS", "must be >= 0")
	}
	if v.cfg.DatabaseMaxIdleConns < 0 {
		v.addf("DATABASE_MAX_IDLE_CONNS", "must be >= 0")
	}
	if v.cfg.DatabaseRequestTimeout <= 0 {
		v.addf("DATABASE_REQUEST_TIMEOUT", "must be positive")
	}
}

// validateOrigins checks ALLOWED_ORIGINS entries are parseable origins, and that
// the wildcard is not used in production.
func (v *ConfigValidator) validateOrigins() {
	for _, origin := range v.cfg.AllowedOrigins {
		if origin == "*" {
			if v.cfg.Env == "production" {
				v.addf("ALLOWED_ORIGINS", "wildcard '*' is not allowed in production")
			}
			continue
		}
		if !isValidOrigin(origin) {
			v.addf("ALLOWED_ORIGINS", "invalid origin %q", origin)
		}
	}
}

// isValidOrigin reports whether raw is an absolute origin (scheme + host, no path).
func isValidOrigin(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != "" && (u.Path == "" || u.Path == "/")
}

func (v *ConfigValidator) validateURLs() {
	for _, u := range []struct {
		name  string
		value string
	}{
		{"BASE_URL", v.cfg.BaseURL},
		{"FRONTEND_URL", v.cfg.FrontendURL},
	} {
		if u.value == "" {
			continue
		}
		parsed, err := url.Parse(u.value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			v.addf(u.name, "must be an absolute URL, got %q", u.value)
		}
	}
}

func (v *ConfigValidator) validateSession() {
	switch {
	case v.cfg.SessionMaxAge <= 0:
		v.addf("SESSION_MAX_AGE_SECONDS", "must be a positive integer")
	case v.cfg.SessionMaxAge < 300:
		v.addf("SESSION_MAX_AGE_SECONDS", "should be at least 300 seconds (5 minutes)")
	}
}

func (v *ConfigValidator) validateRateLimits() {
	if v.cfg.RateLimitPerMin <= 0 {
		v.addf("RATE_LIMIT_REQUESTS", "must be a positive integer")
	}
	if v.cfg.AuthRateLimitMin <= 0 {
		v.addf("AUTH_RATE_LIMIT_REQUESTS", "must be a positive integer")
	}
}

// validateMetrics checks the scrape access control. A token that is set but weak
// is worse than none, since it looks like a control while being guessable.
func (v *ConfigValidator) validateMetrics() {
	for _, cidr := range v.cfg.MetricsAllowedCIDRs {
		if !isValidCIDROrIP(cidr) {
			v.addf("METRICS_ALLOWED_CIDRS", "%q is not a valid CIDR or IP address", cidr)
		}
	}
	if v.cfg.MetricsToken != "" && len(v.cfg.MetricsToken) < minSecretLength {
		v.addf("METRICS_TOKEN", "must be at least %d characters", minSecretLength)
	}
}

func (v *ConfigValidator) validateTimeWindows() {
	if v.cfg.JWTDuration <= 0 {
		v.addf("JWT_DURATION_HOURS", "must be a positive duration")
	}
	if v.cfg.EmailVerifyTokenExpiry < 5*time.Minute {
		v.addf("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS", "should be at least 5 minutes")
	}
	if v.cfg.PasswordResetTokenExpiry < 5*time.Minute {
		v.addf("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES", "should be at least 5 minutes")
	}
}

func (v *ConfigValidator) validateSSR() {
	if v.ssrCfg == nil || !v.ssrCfg.EnableSSR {
		return
	}

	if v.ssrCfg.SSRServerURL == "" {
		v.addf("SSR_SERVER_URL", "is required when SSR_ENABLED=true")
		return
	}
	parsed, err := url.Parse(v.ssrCfg.SSRServerURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		v.addf("SSR_SERVER_URL", "must be an absolute URL, got %q", v.ssrCfg.SSRServerURL)
	}
	if v.ssrCfg.SSRBuildTimeout < 30 || v.ssrCfg.SSRBuildTimeout > 600 {
		v.addf("SSR_BUILD_TIMEOUT", "should be between 30 and 600 seconds")
	}
	if v.ssrCfg.SSRHealthCheckInterval < 5 || v.ssrCfg.SSRHealthCheckInterval > 300 {
		v.addf("SSR_HEALTH_CHECK_INTERVAL", "should be between 5 and 300 seconds")
	}
	if v.ssrCfg.SSRRetryAttempts < 1 || v.ssrCfg.SSRRetryAttempts > 10 {
		v.addf("SSR_RETRY_ATTEMPTS", "should be between 1 and 10")
	}
}
