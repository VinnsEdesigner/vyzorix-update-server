package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port           string
	Env            string
	DatabaseURL    string
	DataDir        string
	BinDir         string
	PublicDir      string
	FirebaseCreds  string
	TokenSecret    string
	JWTSecret      string
	JWTDuration    time.Duration
	AllowedOrigins []string
	EnforceHMAC   bool
	HMACWindow     time.Duration
	// Google OAuth — required for Google sign-in
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string
	// Deployment URLs — used for OAuth redirect construction
	BaseURL     string
	FrontendURL string
}

func (c Config) EnforceHMAC() bool   { return c.EnforceHMAC }
func (c Config) DataDir() string     { return c.DataDir }
func (c Config) BinDir() string     { return c.BinDir }
func (c Config) PublicDir() string  { return c.PublicDir }
func (c Config) BaseURL() string    { return c.BaseURL }

func Load() (Config, error) {
	jwtDuration := 7 * 24 * time.Hour // default 7 days
	if v := os.Getenv("JWT_DURATION_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			jwtDuration = time.Duration(n) * time.Hour
		}
	}

	c := Config{
		Port:           get("PORT", "3000"),
		Env:            get("NODE_ENV", get("GO_ENV", "development")),
		DatabaseURL:    get("DATABASE_URL", "./data/vyzorix.db"),
		DataDir:        get("VYZORIX_API_DIR", "./data"),
		BinDir:         get("VYZORIX_BIN_DIR", "./bin"),
		PublicDir:      get("VYZORIX_PUBLIC_DIR", "./public"),
		FirebaseCreds:  os.Getenv("FIREBASE_CREDENTIALS"),
		TokenSecret:    os.Getenv("TOKEN_SECRET"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTDuration:    jwtDuration,
		AllowedOrigins: splitCSV(get("ALLOWED_ORIGINS", "*")),
		HMACWindow:     5 * time.Minute,
		GoogleOAuthClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		BaseURL:     get("BASE_URL", "http://localhost:3000"),
		FrontendURL: get("FRONTEND_URL", "http://localhost:5173"),
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
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	if c.Env == "production" && c.TokenSecret == "" {
		return c, fmt.Errorf("TOKEN_SECRET is required in production")
	}
	return c, nil
}

func get(k, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
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
