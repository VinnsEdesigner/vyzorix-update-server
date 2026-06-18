// Package logging provides structured logging utilities.
package logging

import (
	"regexp"
	"strings"
)

// Redactor redacts sensitive information from logs.
type Redactor struct {
	patterns []*regexp.Regexp
}

// NewRedactor creates a new log redactor.
func NewRedactor() *Redactor {
	r := &Redactor{
		patterns: []*regexp.Regexp{
			// Email addresses
			regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
			// Credit card numbers (basic pattern)
			regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
			// Social Security Numbers
			regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			// API keys (generic patterns)
			regexp.MustCompile(`(?i)(api[_-]?key|secret[_-]?key|access[_-]?token|auth[_-]?token|bearer)["\s:=]+["']?[a-zA-Z0-9_-]{20,}["']?`),
			// Passwords in JSON
			regexp.MustCompile(`(?i)"(password|passwd|pwd|secret)["\s:]+["'][^"']{4,}["']`),
			// JWT tokens
			regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+`),
			// Firebase credentials
			regexp.MustCompile(`"type"\s*:\s*"service_account"`),
		},
	}
	return r
}

// Redact replaces sensitive patterns with [REDACTED].
func (r *Redactor) Redact(s string) string {
	result := s
	for _, pattern := range r.patterns {
		result = pattern.ReplaceAllString(result, "[REDACTED]")
	}
	return result
}

// RedactMap redacts sensitive fields in a map.
func (r *Redactor) RedactMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		// Check if key is sensitive
		if isSensitiveKey(k) {
			result[k] = "[REDACTED]"
			continue
		}
		result[k] = r.redactValue(v)
	}
	return result
}

// redactValue recursively redacts values.
func (r *Redactor) redactValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return r.Redact(val)
	case map[string]interface{}:
		return r.RedactMap(val)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = r.redactValue(item)
		}
		return result
	default:
		return val
	}
}

// isSensitiveKey returns true if the key name suggests sensitive data.
func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	sensitive := []string{
		"password", "secret", "token", "key", "credential",
		"auth", "api_key", "apikey", "private", "ssn",
		"credit_card", "cvv", "pin",
	}
	for _, s := range sensitive {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// FieldsToRedact returns a list of field names to always redact.
var FieldsToRedact = []string{
	"password",
	"password_hash",
	"secret",
	"api_key",
	"client_secret",
	"token",
	"jwt",
	"bearer",
	"authorization",
	"cookie",
	"credentials",
	"firebase_credentials",
	"resend_api_key",
	"google_oauth_secret",
}
