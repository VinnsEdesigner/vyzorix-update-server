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
	APIKeys                       map[string]string
	APIKeyPrefix                  string
	GitHubReleaseRepo             string
	GitHubOAuthClientSecret       string
	FirebaseCreds                 string
	SessionSecret                 string
	FrontendURL                   string
	BinDir                        string
	Port                          string
	BaseURL                       string
	EmailFrom                     string
	DataDir                       string
	GoogleOAuthClientSecret       string
	GoogleOAuthClientID           string
	FirebaseAppID                 string
	GitHubReleaseToken            string
	DeviceSecret                  string
	GitHubWebhookSecret           string
	JWTSecret                     string
	GitHubOAuthClientID           string
	Env                           string
	DatabaseURL                   string
	DatabaseBackend               string // auto | sqlite | turso
	TursoDatabaseURL              string
	TursoAuthToken                string
	DatabaseMaxOpenConns          int
	DatabaseMaxIdleConns          int
	DatabaseConnMaxLifetime       time.Duration
	DatabaseConnMaxIdleTime       time.Duration
	DatabaseRequestTimeout        time.Duration
	DatabaseHealthCheckPeriod     time.Duration
	ResendAPIKey                  string
	ServerAPIToken                string
	EmailFromName                 string
	PublicDir                     string
	AuditLogSeparateDBPath        string
	AuditLogPath                  string
	AllowedOrigins                []string
	DiagnosticsConfig             DiagnosticsConfig
	MonthlyKeyLimit               int
	SessionMaxAge                 int
	DeviceDeletionIntervalMinutes int
	PasswordResetTokenExpiry      time.Duration
	EmailVerifyTokenExpiry        time.Duration
	JWTDuration                   time.Duration
	MaxKeyNameLength              int
	HMACWindow                    time.Duration
	NonceCacheTTL                 time.Duration
	EnableGraphQL                 bool
	AuditLogSeparateDB            bool
	RequireKeyName                bool
	AllowKeyRenaming              bool
	EnableUsageTracking           bool
	DeviceDeletionEnabled         bool
	EnforceHMAC                   bool
	RateLimitPerMin               int
	AuthRateLimitMin              int
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
	jwtDuration, err := parseJWTDuration()
	if err != nil {
		return Config{}, err
	}

	sessionMaxAge, err := parseSessionMaxAge()
	if err != nil {
		return Config{}, err
	}

	emailVerifyExpiry, err := parseEmailVerifyExpiry()
	if err != nil {
		return Config{}, err
	}

	passwordResetExpiry, err := parsePasswordResetExpiry()
	if err != nil {
		return Config{}, err
	}

	c := Config{
		Port:                      get("PORT", "3000"),
		Env:                       get("NODE_ENV", get("GO_ENV", "development")),
		DatabaseURL:               get("DATABASE_URL", "./data/vyzorix.db"),
		DatabaseBackend:           get("DATABASE_BACKEND", "auto"),
		TursoDatabaseURL:          os.Getenv("TURSO_DB_URL"),
		TursoAuthToken:            os.Getenv("TURSO_AUTH_TOKEN"),
		DatabaseMaxOpenConns:      getenvInt("DATABASE_MAX_OPEN_CONNS", 16),
		DatabaseMaxIdleConns:      getenvInt("DATABASE_MAX_IDLE_CONNS", 8),
		DatabaseConnMaxLifetime:   parseDurationEnv("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute),
		DatabaseConnMaxIdleTime:   parseDurationEnv("DATABASE_CONN_MAX_IDLE_TIME", 5*time.Minute),
		DatabaseRequestTimeout:    parseDurationEnv("DATABASE_REQUEST_TIMEOUT", 15*time.Second),
		DatabaseHealthCheckPeriod: parseDurationEnv("DATABASE_HEALTH_CHECK_PERIOD", 30*time.Second),
		DataDir:                   get("VYZORIX_API_DIR", "./data"),
		BinDir:                    get("VYZORIX_BIN_DIR", "./bin"),
		PublicDir:                 get("VYZORIX_PUBLIC_DIR", "./public"),
		FirebaseCreds:             os.Getenv("FIREBASE_CREDENTIALS"),
		ServerAPIToken:            os.Getenv("SERVER_API_TOKEN"),
		APIKeyPrefix:              get("API_KEY_PREFIX", "vxyz"),
		MonthlyKeyLimit:           20,
		MaxKeyNameLength:          64,
		RequireKeyName:            true,
		AllowKeyRenaming:          true,
		EnableUsageTracking:       true,
		JWTSecret:                 os.Getenv("JWT_SECRET"),
		SessionSecret:             os.Getenv("SESSION_SECRET"),
		SessionMaxAge:             sessionMaxAge,
		AllowedOrigins:            splitCSV(get("ALLOWED_ORIGINS", "*")),
		HMACWindow:                30 * time.Second,
		NonceCacheTTL:             1 * time.Hour,
		GoogleOAuthClientID:       os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret:   os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GitHubOAuthClientID:       os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		GitHubOAuthClientSecret:   os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
		GitHubReleaseRepo:         os.Getenv("GITHUB_RELEASE_REPO"),
		GitHubReleaseToken:        os.Getenv("GITHUB_RELEASE_TOKEN"),
		GitHubWebhookSecret:       os.Getenv("GITHUB_WEBHOOK_SECRET"),
		BaseURL:                   get("BASE_URL", "http://localhost:3000"),
		FrontendURL:               get("FRONTEND_URL", "http://localhost:5173"),
		ResendAPIKey:              os.Getenv("RESEND_API_KEY"),
		EmailFrom:                 get("EMAIL_FROM", "noreply@vyzorix.app"),
		EmailFromName:             get("EMAIL_FROM_NAME", "Vyzorix"),
		JWTDuration:               jwtDuration,
		EmailVerifyTokenExpiry:    emailVerifyExpiry,
		PasswordResetTokenExpiry:  passwordResetExpiry,
		EnableGraphQL:             getBool("ENABLE_GRAPHQL", true),
		AuditLogPath:              get("AUDIT_LOG_PATH", "./data/audit.log"),
		AuditLogSeparateDB:        getBool("AUDIT_LOG_SEPARATE_DB", false),
		AuditLogSeparateDBPath:    get("AUDIT_LOG_SEPARATE_DB_PATH", "./data/audit/audit.db"),
		DiagnosticsConfig:         LoadDiagnosticsConfig(),
		RateLimitPerMin:           getenvInt("RATE_LIMIT_REQUESTS", 100),
		AuthRateLimitMin:          getenvInt("AUTH_RATE_LIMIT_REQUESTS", 60),
	}

	c.APIKeys = loadAPIKeys()

	enforceDefault := strings.EqualFold(c.Env, "production")
	c.EnforceHMAC = getBool("ENFORCE_HMAC", enforceDefault)

	hmacWindow, err := parseHMACWindow()
	if err != nil {
		return c, err
	}
	c.HMACWindow = hmacWindow

	if err := validateRequiredSecrets(&c); err != nil {
		return c, err
	}

	if err := validateProductionConfig(&c); err != nil {
		return c, err
	}

	if err := validateDatabaseConfig(&c); err != nil {
		return c, err
	}

	if err := validatePort(&c); err != nil {
		return c, err
	}

	if err := parseDeviceDeletionConfig(&c); err != nil {
		return c, err
	}

	c.DeviceSecret = os.Getenv("DEVICE_SECRET")
	c.FirebaseAppID = os.Getenv("FIREBASE_APP_ID")

	return c, nil
}

// parseJWTDuration parses JWT_DURATION_HOURS environment variable.
func parseJWTDuration() (time.Duration, error) {
	v := os.Getenv("JWT_DURATION_HOURS")
	if v == "" {
		return 7 * 24 * time.Hour, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("JWT_DURATION_HOURS must be a valid integer: %w", err)
	}
	if n <= 0 {
		return 0, errors.New("JWT_DURATION_HOURS must be positive")
	}
	if n > 8760 {
		return 0, errors.New("JWT_DURATION_HOURS exceeds maximum of 8760 (1 year)")
	}
	return time.Duration(n) * time.Hour, nil
}

// parseSessionMaxAge parses SESSION_MAX_AGE_SECONDS environment variable.
func parseSessionMaxAge() (int, error) {
	v := os.Getenv("SESSION_MAX_AGE_SECONDS")
	if v == "" {
		return 86400, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("SESSION_MAX_AGE_SECONDS must be a valid integer: %w", err)
	}
	if n <= 0 {
		return 0, errors.New("SESSION_MAX_AGE_SECONDS must be positive")
	}
	if n > 604800 {
		return 0, errors.New("SESSION_MAX_AGE_SECONDS exceeds maximum of 604800 (7 days)")
	}
	return n, nil
}

// parseEmailVerifyExpiry parses EMAIL_VERIFY_TOKEN_EXPIRY_HOURS environment variable.
func parseEmailVerifyExpiry() (time.Duration, error) {
	v := os.Getenv("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS")
	if v == "" {
		return 24 * time.Hour, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS must be a valid integer: %w", err)
	}
	if n <= 0 {
		return 0, errors.New("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS must be positive")
	}
	if n > 168 {
		return 0, errors.New("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS exceeds maximum of 168 (7 days)")
	}
	return time.Duration(n) * time.Hour, nil
}

// parsePasswordResetExpiry parses PASSWORD_RESET_TOKEN_EXPIRY_MINUTES environment variable.
func parsePasswordResetExpiry() (time.Duration, error) {
	v := os.Getenv("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES")
	if v == "" {
		return time.Hour, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES must be a valid integer: %w", err)
	}
	if n <= 0 {
		return 0, errors.New("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES must be positive")
	}
	if n > 1440 {
		return 0, errors.New("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES exceeds maximum of 1440 (24 hours)")
	}
	return time.Duration(n) * time.Minute, nil
}

// parseHMACWindow parses HMAC_WINDOW_SECONDS environment variable.
func parseHMACWindow() (time.Duration, error) {
	v := os.Getenv("HMAC_WINDOW_SECONDS")
	if v == "" {
		return 30 * time.Second, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid HMAC_WINDOW_SECONDS: %q", v)
	}
	if n > 300 {
		return 0, fmt.Errorf("HMAC_WINDOW_SECONDS exceeds maximum of 300 (5 minutes)")
	}
	return time.Duration(n) * time.Second, nil
}

// validateRequiredSecrets checks for required secrets.
func validateRequiredSecrets(c *Config) error {
	var missingVars []string

	if strings.TrimSpace(c.DatabaseURL) == "" {
		missingVars = append(missingVars, "DATABASE_URL")
	}
	if len(c.APIKeys) == 0 {
		missingVars = append(missingVars, "at least one API_KEY_* (e.g., API_KEY_primary=<key>)")
	}
	if c.JWTSecret == "" {
		missingVars = append(missingVars, "JWT_SECRET")
	}
	if c.SessionSecret == "" {
		missingVars = append(missingVars, "SESSION_SECRET")
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missingVars)
	}
	return nil
}

// validateProductionConfig validates production-specific configuration.
func validateProductionConfig(c *Config) error {
	if c.Env != "production" {
		return nil
	}
	if c.ServerAPIToken == "" {
		return errors.New("SERVER_API_TOKEN is required in production")
	}
	if len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters in production")
	}
	if len(c.SessionSecret) < 32 {
		return errors.New("SESSION_SECRET must be at least 32 characters in production")
	}
	return nil
}

// ResolvedDatabaseBackend determines which storage backend will be used given
// the configured DATABASE_BACKEND and the presence of Turso credentials.
// "auto" selects turso when TURSO_DB_URL is set, otherwise sqlite.
func (c *Config) ResolvedDatabaseBackend() string {
	switch strings.ToLower(strings.TrimSpace(c.DatabaseBackend)) {
	case "sqlite":
		return "sqlite"
	case "turso":
		return "turso"
	default: // "auto" or unset
		if c.TursoDatabaseURL != "" {
			return "turso"
		}
		return "sqlite"
	}
}

// validateDatabaseConfig validates the database backend selection and the
// credentials required by the chosen backend.
func validateDatabaseConfig(c *Config) error {
	backend := c.ResolvedDatabaseBackend()

	switch backend {
	case "turso":
		if c.TursoDatabaseURL == "" {
			return errors.New("DATABASE_BACKEND=turso requires TURSO_DB_URL to be set")
		}
		if c.TursoAuthToken == "" {
			return errors.New("DATABASE_BACKEND=turso requires TURSO_AUTH_TOKEN to be set")
		}
		if !strings.HasPrefix(c.TursoDatabaseURL, "libsql://") &&
			!strings.HasPrefix(c.TursoDatabaseURL, "https://") &&
			!strings.HasPrefix(c.TursoDatabaseURL, "http://") {
			return errors.New("TURSO_DB_URL must use the libsql://, https://, or http:// scheme")
		}
		if c.Env == "production" && strings.HasPrefix(c.TursoDatabaseURL, "http://") {
			return errors.New("TURSO_DB_URL must use libsql:// or https:// in production (http:// is plaintext)")
		}
	case "sqlite":
		if strings.TrimSpace(c.DatabaseURL) == "" {
			return errors.New("DATABASE_BACKEND=sqlite requires DATABASE_URL to be set")
		}
	default:
		return fmt.Errorf("invalid DATABASE_BACKEND %q (want auto, sqlite, or turso)", c.DatabaseBackend)
	}

	if c.DatabaseMaxOpenConns < 0 {
		return errors.New("DATABASE_MAX_OPEN_CONNS must be >= 0")
	}
	if c.DatabaseMaxIdleConns < 0 {
		return errors.New("DATABASE_MAX_IDLE_CONNS must be >= 0")
	}
	if c.DatabaseRequestTimeout <= 0 {
		return errors.New("DATABASE_REQUEST_TIMEOUT must be positive")
	}
	return nil
}

// validatePort validates the port configuration.
func validatePort(c *Config) error {
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid PORT: %q (must be between 1 and 65535)", c.Port)
	}
	return nil
}

// parseDeviceDeletionConfig parses device deletion configuration.
func parseDeviceDeletionConfig(c *Config) error {
	c.DeviceDeletionEnabled = getBool("DEVICE_DELETION_ENABLED", false)
	c.DeviceDeletionIntervalMinutes = getenvInt("DEVICE_DELETION_INTERVAL_MINUTES", 5)

	if c.DeviceDeletionIntervalMinutes < 1 {
		return errors.New("DEVICE_DELETION_INTERVAL_MINUTES must be at least 1")
	}
	if c.DeviceDeletionIntervalMinutes > 60 {
		return errors.New("DEVICE_DELETION_INTERVAL_MINUTES exceeds maximum of 60 (1 hour)")
	}
	return nil
}

func get(k, fallback string) string {
	v := os.Getenv(k)
	if v != "" {
		return strings.TrimSpace(v)
	}

	return fallback
}

// parseDurationEnv parses a Go duration string (e.g. "30s", "5m", "1h") with a
// fallback default. Falls back when unset or unparseable.
func parseDurationEnv(k string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
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
