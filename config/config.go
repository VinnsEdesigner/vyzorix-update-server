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
	AllowedOrigins []string
	EnforceHMAC    bool
	HMACWindow     time.Duration
}

func Load() (Config, error) {
	c := Config{
		Port:           get("PORT", "3000"),
		Env:            get("NODE_ENV", get("GO_ENV", "development")),
		DatabaseURL:    get("DATABASE_URL", "./data/vyzorix.db"),
		DataDir:        get("VYZORIX_API_DIR", "./api/v1"),
		BinDir:         get("VYZORIX_BIN_DIR", "./bin"),
		PublicDir:      get("VYZORIX_PUBLIC_DIR", "./public"),
		FirebaseCreds:  os.Getenv("FIREBASE_CREDENTIALS"),
		TokenSecret:    os.Getenv("TOKEN_SECRET"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AllowedOrigins: splitCSV(get("ALLOWED_ORIGINS", "*")),
		HMACWindow:     5 * time.Minute,
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
