// Package logging provides structured logging with PII redaction.
package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// NewAuditFileLogger creates a dedicated logger that writes audit events to a separate file.
// Audit logs are always JSON formatted for machine parsing and are never redacted
// (unlike application logs which may redact PII).
func NewAuditFileLogger(path string) *slog.Logger {
	if path == "" {
		path = "audit.log"
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		// Fall back to stdout if file cannot be opened
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	handler := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}

// LoggerConfig holds logging configuration.
type LoggerConfig struct {
	Level      string
	RedactKeys []string
	RedactPII  bool
}

// DefaultLoggerConfig returns the default logging configuration.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
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
func NewLogger(cfg LoggerConfig) *slog.Logger {
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
	if os.Getenv("ENV") == "production" || os.Getenv("NODE_ENV") == "production" {
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

// Handle implements slog.Handler, redacting PII before logging.
func (h *redactingHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.enabled {
		h.redactRecord(r)
	}

	return h.Handler.Handle(ctx, r)
}

// redactRecord redacts PII from log record attributes.
func (h *redactingHandler) redactRecord(r slog.Record) {
	redactor := NewRedactor()

	var newAttrs []slog.Attr

	r.Attrs(func(a slog.Attr) bool {
		newAttrs = append(newAttrs, h.redactAttr(a, redactor))
		return true
	})

	for _, a := range newAttrs {
		r.AddAttrs(a)
	}
}

// redactAttr redacts sensitive keys and PII from an attribute.
func (h *redactingHandler) redactAttr(a slog.Attr, redactor *Redactor) slog.Attr {
	key := strings.ToLower(a.Key)
	for _, redactKey := range h.redactKeys {
		if strings.Contains(key, redactKey) {
			return slog.String(a.Key, "[REDACTED]")
		}
	}

	if a.Value.Kind() == slog.KindString {
		value := redactor.Redact(a.Value.String())
		return slog.String(a.Key, value)
	}

	return a
}

// NewProductionLogger creates a logger optimized for production.
func NewProductionLogger() *slog.Logger {
	return NewLogger(LoggerConfig{
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
	return NewLogger(LoggerConfig{
		Level:      "debug",
		RedactPII:  false,
		RedactKeys: []string{},
	})
}

// NewFromEnv creates a logger based on the ENV environment variable.
func NewFromEnv() *slog.Logger {
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("NODE_ENV")
	}

	if env == "production" {
		return NewProductionLogger()
	}

	return NewDevelopmentLogger()
}
