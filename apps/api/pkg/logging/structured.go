// Package logging provides structured logging with PII redaction.
package logging

import (
	"context"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

// Config holds logging configuration.
type Config struct {
	Level      string
	RedactPII  bool
	RedactKeys []string
}

// DefaultConfig returns the default logging configuration.
func DefaultConfig() Config {
	return Config{
		Level:     "info",
		RedactPII: true,
		RedactKeys: []string{
			"password",
			"secret",
			"token",
			"apikey",
			"api_key",
			"auth",
			"credential",
			"private",
			"ssn",
			"credit_card",
			"email",
			"phone",
		},
	}
}

// NewLogger creates a new structured logger with PII redaction.
func NewLogger(cfg Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if os.Getenv("ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(&redactingHandler{
		Handler:    handler,
		redactKeys: cfg.RedactKeys,
		enabled:    cfg.RedactPII,
	})
}

// redactingHandler wraps a slog handler to redact PII from log output.
type redactingHandler struct {
	slog.Handler
	redactKeys []string
	enabled    bool
}

func (h *redactingHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.enabled {
		h.redactRecord(r)
	}
	return h.Handler.Handle(ctx, r)
}

func (h *redactingHandler) redactRecord(r slog.Record) {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	ccRegex := regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`)
	ssnRegex := regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	phoneRegex := regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`)
	ipRegex := regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)

	var newAttrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		newAttrs = append(newAttrs, h.redactAttr(a, emailRegex, ccRegex, ssnRegex, phoneRegex, ipRegex))
		return true
	})

	for _, a := range newAttrs {
		r.AddAttrs(a)
	}
}

func (h *redactingHandler) redactAttr(a slog.Attr, emailRegex, ccRegex, ssnRegex, phoneRegex, ipRegex *regexp.Regexp) slog.Attr {
	key := strings.ToLower(a.Key)
	for _, redactKey := range h.redactKeys {
		if strings.Contains(key, redactKey) {
			return slog.String(a.Key, "[REDACTED]")
		}
	}

	if a.Value.Kind() == slog.KindString {
		value := a.Value.String()
		value = emailRegex.ReplaceAllString(value, "[EMAIL]")
		value = ccRegex.ReplaceAllString(value, "[CARD]")
		value = ssnRegex.ReplaceAllString(value, "[SSN]")
		value = phoneRegex.ReplaceAllString(value, "[PHONE]")
		value = ipRegex.ReplaceAllString(value, "[IP]")
		return slog.String(a.Key, value)
	}

	return a
}

// NewProductionLogger creates a logger optimized for production.
func NewProductionLogger() *slog.Logger {
	return NewLogger(Config{
		Level:     "info",
		RedactPII: true,
		RedactKeys: []string{
			"password", "secret", "token", "apikey", "api_key",
			"auth", "credential", "private", "authorization",
			"ssn", "credit_card", "card_number", "cvv",
		},
	})
}

// NewDevelopmentLogger creates a logger for development.
func NewDevelopmentLogger() *slog.Logger {
	return NewLogger(Config{
		Level:     "debug",
		RedactPII: false,
		RedactKeys: []string{},
	})
}