// Package config provides configuration loading from environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// SSRConfig holds SSR server configuration.
type SSRConfig struct {
	SSRServerURL           string
	SSRPort                string
	SSRBuildTimeout        int
	SSRHealthCheckInterval int
	SSRRetryAttempts       int
	EnableSSR              bool
	SSRAutoStart           bool
	SSRAutoBuild           bool
}

// SigningConfig holds request signing configuration.
type SigningConfig struct {
	TimestampWindow       int
	MaxCacheSize          int
	GracePeriod           int
	Enabled               bool
	AllowUnsignedFallback bool
}

// DefaultSigningConfig returns default signing configuration.
func DefaultSigningConfig() SigningConfig {
	return SigningConfig{
		Enabled:               false,
		TimestampWindow:       300,
		MaxCacheSize:          100000,
		GracePeriod:           86400,
		AllowUnsignedFallback: false,
	}
}

// LoadSigningConfig loads signing configuration from environment.
func LoadSigningConfig() SigningConfig {
	cfg := DefaultSigningConfig()

	if v := os.Getenv("REQUEST_SIGNING_ENABLED"); v != "" {
		cfg.Enabled = v == "true" || v == "1" || v == "yes"
	}

	if v := os.Getenv("SIGNING_TIMESTAMP_WINDOW"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.TimestampWindow = n
		}
	}

	if v := os.Getenv("SIGNING_MAX_CACHE_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxCacheSize = n
		}
	}

	if v := os.Getenv("SIGNING_GRACE_PERIOD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.GracePeriod = n
		}
	}

	if v := os.Getenv("ALLOW_UNSIGNED_FALLBACK"); v != "" {
		cfg.AllowUnsignedFallback = v == "true" || v == "1" || v == "yes"
	}

	return cfg
}

// Config holds all application configuration loaded from environment variables.
type Config struct {
	JWTSecret                string
	EmailFrom                string
	GitHubOAuthClientID      string
	GitHubOAuthClientSecret  string
	FirebaseCreds            string
	SessionSecret            string
	FrontendURL              string
	BinDir                   string
	Port                     string
	BaseURL                  string
	DatabaseURL              string
	DataDir                  string
	GoogleOAuthClientSecret  string
	GoogleOAuthClientID      string
	PublicDir                string
	GitHubReleaseToken       string
	GitHubReleaseRepo        string
	GitHubWebhookSecret      string
	ResendAPIKey             string
	Env                      string
	TokenSecret              string
	EmailFromName            string
	AllowedOrigins           []string
	PasswordResetTokenExpiry time.Duration
	HMACWindow               time.Duration
	SessionMaxAge            int
	NonceCacheTTL            time.Duration
	EmailVerifyTokenExpiry   time.Duration
	JWTDuration              time.Duration
	EnforceHMAC              bool
	EnableGraphQL            bool
	DiagnosticsConfig         DiagnosticsConfig
}

// DiagnosticsConfig holds configuration for the diagnostics API.
type DiagnosticsConfig struct {
	OfflineThresholdMinutes     int
	FCMTokenExpiryDays          int
	InspectionCacheTTLSeconds   int
	TelemetryRetentionDays       int
}

// DefaultDiagnosticsConfig returns default diagnostics configuration.
func DefaultDiagnosticsConfig() DiagnosticsConfig {
	return DiagnosticsConfig{
		OfflineThresholdMinutes:   5,
		FCMTokenExpiryDays:        30,
		InspectionCacheTTLSeconds: 10,
		TelemetryRetentionDays:    7,
	}
}

// LoadDiagnosticsConfig loads diagnostics configuration from environment.
func LoadDiagnosticsConfig() DiagnosticsConfig {
	cfg := DefaultDiagnosticsConfig()

	if v := os.Getenv("DIAG_OFFLINE_THRESHOLD_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.OfflineThresholdMinutes = n
		}
	}

	if v := os.Getenv("DIAG_FCM_TOKEN_EXPIRY_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.FCMTokenExpiryDays = n
		}
	}

	if v := os.Getenv("DIAG_INSPECTION_CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.InspectionCacheTTLSeconds = n
		}
	}

	if v := os.Getenv("DIAG_TELEMETRY_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.TelemetryRetentionDays = n
		}
	}

	return cfg
}

// Load reads configuration from environment variables and returns a Config struct.
func Load() (Config, error) {
	jwtDuration := 7 * 24 * time.Hour

	if v := os.Getenv("JWT_DURATION_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			jwtDuration = time.Duration(n) * time.Hour
		}
	}

	sessionMaxAge := 86400

	if v := os.Getenv("SESSION_MAX_AGE_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			sessionMaxAge = n
		}
	}

	emailVerifyExpiry := 24 * time.Hour

	if v := os.Getenv("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			emailVerifyExpiry = time.Duration(n) * time.Hour
		}
	}

	passwordResetExpiry := time.Hour

	if v := os.Getenv("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			passwordResetExpiry = time.Duration(n) * time.Minute
		}
	}

	c := Config{
		Port:                     get("PORT", "3000"),
		Env:                      get("NODE_ENV", get("GO_ENV", "development")),
		DatabaseURL:              get("DATABASE_URL", "./data/vyzorix.db"),
		DataDir:                  get("VYZORIX_API_DIR", "./data"),
		BinDir:                   get("VYZORIX_BIN_DIR", "./bin"),
		PublicDir:                get("VYZORIX_PUBLIC_DIR", "./public"),
		FirebaseCreds:            os.Getenv("FIREBASE_CREDENTIALS"),
		TokenSecret:              os.Getenv("TOKEN_SECRET"),
		JWTSecret:                os.Getenv("JWT_SECRET"),
		SessionSecret:            os.Getenv("SESSION_SECRET"),
		SessionMaxAge:            sessionMaxAge,
		AllowedOrigins:           splitCSV(get("ALLOWED_ORIGINS", "*")),
		HMACWindow:               30 * time.Second,
		NonceCacheTTL:            1 * time.Hour,
		GoogleOAuthClientID:      os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret:  os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GitHubOAuthClientID:      os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		GitHubOAuthClientSecret:  os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
		GitHubReleaseRepo:        os.Getenv("GITHUB_RELEASE_REPO"),
		GitHubReleaseToken:       os.Getenv("GITHUB_RELEASE_TOKEN"),
			GitHubWebhookSecret:      os.Getenv("GITHUB_WEBHOOK_SECRET"),
		BaseURL:                  get("BASE_URL", "http://localhost:3000"),
		FrontendURL:              get("FRONTEND_URL", "http://localhost:5173"),
		ResendAPIKey:             os.Getenv("RESEND_API_KEY"),
		EmailFrom:                get("EMAIL_FROM", "noreply@vyzorix.app"),
		EmailFromName:            get("EMAIL_FROM_NAME", "Vyzorix"),
		JWTDuration:              jwtDuration,
		EmailVerifyTokenExpiry:   emailVerifyExpiry,
		PasswordResetTokenExpiry: passwordResetExpiry,
		EnableGraphQL:            getBool("ENABLE_GRAPHQL", true), // Enabled by default
		DiagnosticsConfig:         LoadDiagnosticsConfig(),
	}

	enforceDefault := strings.EqualFold(c.Env, "production")
	c.EnforceHMAC = getBool("ENFORCE_HMAC", enforceDefault)

	if v := os.Getenv("HMAC_WINDOW_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return c, fmt.Errorf("invalid HMAC_WINDOW_SECONDS: %q", v)
		}

		c.HMACWindow = time.Duration(n) * time.Second
	}

	if strings.TrimSpace(c.DatabaseURL) == "" {
		return c, errors.New("DATABASE_URL is required")
	}

	if c.Env == "production" && c.TokenSecret == "" {
		return c, errors.New("TOKEN_SECRET is required in production")
	}

	return c, nil
}

func get(k, fallback string) string {
	v := os.Getenv(k)
	if v != "" {
		return strings.TrimSpace(v)
	}

	return fallback
}

func getBool(k string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return fallback
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}

	return b
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}

	return out
}

// DefaultSSRConfig returns default SSR configuration.
func DefaultSSRConfig() SSRConfig {
	return SSRConfig{
		EnableSSR:              true,
		SSRServerURL:           "http://localhost:3001",
		SSRPort:                "3001",
		SSRAutoStart:           true,
		SSRBuildTimeout:        60,
		SSRAutoBuild:           true,
		SSRHealthCheckInterval: 5,
		SSRRetryAttempts:       3,
	}
}

// LoadSSRConfig loads SSR configuration from environment.
func LoadSSRConfig() SSRConfig {
	cfg := DefaultSSRConfig()

	cfg.EnableSSR = parseBoolEnv("SSR_ENABLED", cfg.EnableSSR)
	cfg.SSRServerURL = parseStringEnv("SSR_SERVER_URL", cfg.SSRServerURL)
	cfg.SSRPort = parseStringEnv("SSR_PORT", cfg.SSRPort)
	cfg.SSRAutoStart = parseBoolEnv("SSR_AUTO_START", cfg.SSRAutoStart)
	cfg.SSRBuildTimeout = parseIntEnv("SSR_BUILD_TIMEOUT", cfg.SSRBuildTimeout)
	cfg.SSRAutoBuild = parseBoolEnv("SSR_AUTO_BUILD", cfg.SSRAutoBuild)
	cfg.SSRHealthCheckInterval = parseIntEnv("SSR_HEALTH_CHECK_INTERVAL", cfg.SSRHealthCheckInterval)
	cfg.SSRRetryAttempts = parseIntEnv("SSR_RETRY_ATTEMPTS", cfg.SSRRetryAttempts)

	return cfg
}

func parseStringEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseBoolEnv(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1" || v == "yes"
	}
	return fallback
}

func parseIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}
