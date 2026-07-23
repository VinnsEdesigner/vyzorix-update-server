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
	APIKeys                  map[string]string
	ResendAPIKey             string
	ServerAPIToken              string
	GitHubOAuthClientSecret  string
	FirebaseCreds            string
	SessionSecret            string
	FrontendURL              string
	BinDir                   string
	Port                     string
	BaseURL                  string
	EmailFrom                string
	DataDir                  string
	GoogleOAuthClientSecret  string
	GoogleOAuthClientID      string
	PublicDir                string
	GitHubReleaseToken       string
	GitHubReleaseRepo        string
	GitHubWebhookSecret      string
	JWTSecret                string
	GitHubOAuthClientID      string
	Env                      string
	DatabaseURL              string
	APIKeyPrefix             string
	EmailFromName            string
	AllowedOrigins           []string
	DiagnosticsConfig        DiagnosticsConfig
	NonceCacheTTL            time.Duration
	MonthlyKeyLimit          int
	MaxKeyNameLength         int
	SessionMaxAge            int
	HMACWindow               time.Duration
	PasswordResetTokenExpiry time.Duration
	EmailVerifyTokenExpiry   time.Duration
	JWTDuration              time.Duration
	EnableUsageTracking      bool
	AllowKeyRenaming         bool
	EnforceHMAC              bool
	EnableGraphQL            bool
	RequireKeyName           bool
	AuditLogPath             string
	AuditLogSeparateDB       bool   
	AuditLogSeparateDBPath   string 
	DeviceSecret             string
	FirebaseAppID            string // Firebase App ID for App Check validation.
	DeviceDeletionEnabled     bool   // Enable background worker for device deletion.
	DeviceDeletionIntervalMinutes int // Interval in minutes for device deletion worker.
}

// DiagnosticsConfig holds configuration for the diagnostics API.
type DiagnosticsConfig struct {
	OfflineThresholdMinutes   int
	FCMTokenExpiryDays        int
	InspectionCacheTTLSeconds int
	TelemetryRetentionDays    int
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

// rather than failing on first request.
func Load() (Config, error) {
	jwtDuration := 7 * 24 * time.Hour

	if v := os.Getenv("JWT_DURATION_HOURS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("JWT_DURATION_HOURS must be a valid integer: %w", err)
		}
		if n <= 0 {
			return Config{}, errors.New("JWT_DURATION_HOURS must be positive")
		}
		if n > 8760 { // 1 year.
			return Config{}, errors.New("JWT_DURATION_HOURS exceeds maximum of 8760 (1 year)")
		}
		jwtDuration = time.Duration(n) * time.Hour
	}

	sessionMaxAge := 86400

	if v := os.Getenv("SESSION_MAX_AGE_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("SESSION_MAX_AGE_SECONDS must be a valid integer: %w", err)
		}
		if n <= 0 {
			return Config{}, errors.New("SESSION_MAX_AGE_SECONDS must be positive")
		}
		if n > 604800 { // 7 days.
			return Config{}, errors.New("SESSION_MAX_AGE_SECONDS exceeds maximum of 604800 (7 days)")
		}
		sessionMaxAge = n
	}

	emailVerifyExpiry := 24 * time.Hour

	if v := os.Getenv("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS must be a valid integer: %w", err)
		}
		if n <= 0 {
			return Config{}, errors.New("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS must be positive")
		}
		if n > 168 { // 7 days.
			return Config{}, errors.New("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS exceeds maximum of 168 (7 days)")
		}
		emailVerifyExpiry = time.Duration(n) * time.Hour
	}

	passwordResetExpiry := time.Hour

	if v := os.Getenv("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES must be a valid integer: %w", err)
		}
		if n <= 0 {
			return Config{}, errors.New("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES must be positive")
		}
		if n > 1440 { // 24 hours.
			return Config{}, errors.New("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES exceeds maximum of 1440 (24 hours)")
		}
		passwordResetExpiry = time.Duration(n) * time.Minute
	}

	c := Config{
		Port:                     get("PORT", "3000"),
		Env:                      get("NODE_ENV", get("GO_ENV", "development")),
		DatabaseURL:              get("DATABASE_URL", "./data/vyzorix.db"),
		DataDir:                  get("VYZORIX_API_DIR", "./data"),
		BinDir:                   get("VYZORIX_BIN_DIR", "./bin"),
		PublicDir:                get("VYZORIX_PUBLIC_DIR", "./public"),
		FirebaseCreds:            os.Getenv("FIREBASE_CREDENTIALS"),
		ServerAPIToken:              os.Getenv("SERVER_API_TOKEN"),
		APIKeyPrefix:             get("API_KEY_PREFIX", "vxyz"),
		MonthlyKeyLimit:          20,
		MaxKeyNameLength:         64,
		RequireKeyName:           true,
		AllowKeyRenaming:         true,
		EnableUsageTracking:      true,
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
		EnableGraphQL:            getBool("ENABLE_GRAPHQL", true), // Enabled by default.
		AuditLogPath:             get("AUDIT_LOG_PATH", "./data/audit.log"),
		
		AuditLogSeparateDB:       getBool("AUDIT_LOG_SEPARATE_DB", false),
		AuditLogSeparateDBPath:   get("AUDIT_LOG_SEPARATE_DB_PATH", "./data/audit/audit.db"),
		DiagnosticsConfig:        LoadDiagnosticsConfig(),
	}

	// Load API keys (supports multiple for rotation).
	// Format: API_KEY_<id>=<key_value> (e.g., API_KEY_primary=abc123, API_KEY_backup=def456).
	c.APIKeys = loadAPIKeys()

	enforceDefault := strings.EqualFold(c.Env, "production")
	c.EnforceHMAC = getBool("ENFORCE_HMAC", enforceDefault)

	if v := os.Getenv("HMAC_WINDOW_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return c, fmt.Errorf("invalid HMAC_WINDOW_SECONDS: %q", v)
		}
		if n > 300 { // 5 minutes max.
			return c, fmt.Errorf("HMAC_WINDOW_SECONDS exceeds maximum of 300 (5 minutes)")
		}

		c.HMACWindow = time.Duration(n) * time.Second
	}

	
	// Collect all missing or invalid required values.
	var missingVars []string

	if strings.TrimSpace(c.DatabaseURL) == "" {
		missingVars = append(missingVars, "DATABASE_URL")
	}

	if len(c.APIKeys) == 0 {
		missingVars = append(missingVars, "at least one API_KEY_* (e.g., API_KEY_primary=<key>)")
	}

	// JWT_SECRET is required regardless of environment.
	// It is used for signing tokens and cannot have a safe default.
	if c.JWTSecret == "" {
		missingVars = append(missingVars, "JWT_SECRET")
	}

	// SESSION_SECRET is required regardless of environment.
	// It is used for session encryption and cannot have a safe default.
	if c.SessionSecret == "" {
		missingVars = append(missingVars, "SESSION_SECRET")
	}

	// Return all missing variables at once for clear diagnosis.
	if len(missingVars) > 0 {
		return c, fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	// Production-specific validations.
	if c.Env == "production" {
		if c.ServerAPIToken == "" {
			return c, errors.New("SERVER_API_TOKEN is required in production")
		}
		if len(c.JWTSecret) < 32 {
			return c, errors.New("JWT_SECRET must be at least 32 characters in production")
		}
		if len(c.SessionSecret) < 32 {
			return c, errors.New("SESSION_SECRET must be at least 32 characters in production")
		}
	}

	// Validate port is a valid number.
	if port, err := strconv.Atoi(c.Port); err != nil || port < 1 || port > 65535 {
		return c, fmt.Errorf("invalid PORT: %q (must be between 1 and 65535)", c.Port)
	}

	// Device deletion worker configuration.
	c.DeviceDeletionEnabled = getBool("DEVICE_DELETION_ENABLED", false)
	c.DeviceDeletionIntervalMinutes = getenvInt("DEVICE_DELETION_INTERVAL_MINUTES", 5)

	// Validate DeviceDeletionIntervalMinutes.
	if c.DeviceDeletionIntervalMinutes < 1 {
		return c, errors.New("DEVICE_DELETION_INTERVAL_MINUTES must be at least 1")
	}
	if c.DeviceDeletionIntervalMinutes > 60 {
		return c, errors.New("DEVICE_DELETION_INTERVAL_MINUTES exceeds maximum of 60 (1 hour)")
	}

	// Device secret for device attestation (HMAC-SHA256).
	c.DeviceSecret = os.Getenv("DEVICE_SECRET")

	// Firebase App ID for App Check validation.
	c.FirebaseAppID = os.Getenv("FIREBASE_APP_ID")

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

func getenvInt(k string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return n
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

// loadAPIKeys loads API keys from environment variables.
// Supports multiple keys for rotation: API_KEY_<id>=<value>.
// Example: API_KEY_primary=abc123, API_KEY_backup=def456.
func loadAPIKeys() map[string]string {
	keys := make(map[string]string)
	prefix := "API_KEY_"

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}
		// Parse KEY=VALUE.
		if idx := strings.Index(env, "="); idx > 0 {
			keyID := env[len(prefix):idx]
			keyValue := env[idx+1:]
			if keyID != "" && keyValue != "" {
				keys[keyID] = keyValue
			}
		}
	}

	return keys
}
