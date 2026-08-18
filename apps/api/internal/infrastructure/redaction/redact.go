// Package redaction provides utilities for sanitizing sensitive data in logs.
package redaction

import (
	"regexp"
	"strings"
	"sync"
)

// Common secret patterns used throughout the application.
var (
	// API key patterns.
	apiKeyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(api[_-]?key|apikey|api[_-]?secret)\s*[=:]\s*["']?([\w!@#$%^&*()_+=\-]{8,})["']?`),
		regexp.MustCompile(`(?i)(secret[_-]?key|secret|private[_-]?key)\s*[=:]\s*["']?([\w!@#$%^&*()_+=\-]{8,})["']?`),
		regexp.MustCompile(`(?i)(bearer|token|access[_-]?token|refresh[_-]?token)\s+(?:bearer\s+)?([\w!@#$%^&*()_+=\-]{20,})`),
		regexp.MustCompile(`(?i)authorization\s*:\s*(?:bearer\s+)?([\S]{20,})`),
		regexp.MustCompile(`(?i)auth\s*:\s*(?:bearer\s+)?([\S]{20,})`),
	}

	// JWT patterns.
	jwtPatterns = []*regexp.Regexp{
		regexp.MustCompile(`eyJ[A-Za-z0-9-_]+\.eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+`),
		regexp.MustCompile(`(?i)jwt\s*[=:]\s*["']?([\w!@#$%^&*()_+=\-.]{20,})["']?`),
	}

	// Password patterns.
	passwordPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*["']?([^\s"']{4,})["']?`),
		regexp.MustCompile(`(?i)pass\s*[=:]\s*["']?([^\s"']{4,})["']?`),
	}

	// Database connection string patterns.
	dbPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(postgres|mysql|mongodb|redis|sqlite)://[^@]+:[^@]+@`),
		regexp.MustCompile(`(?i)connection[_-]?string\s*[=:]\s*["']?[^\s"']+["']?`),
	}

	// Private key patterns.
	privateKeyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
		regexp.MustCompile(`(?i)private[_-]?key[_-]?pem\s*[=:]\s*["']?[^\s"']+["']?`),
	}

	// Generic credential patterns.
	genericCredentialPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)credential\s*[=:]\s*["']?[^\s"']+["']?`),
		regexp.MustCompile(`(?i)secret\s*[=:]\s*["']?[^\s"']{8,}["']?`),
		regexp.MustCompile(`(?i)access[_-]?key[_-]?id\s*[=:]\s*["']?([\w]{16,})["']?`),
		regexp.MustCompile(`(?i)aws[_-]?secret[_-]?access[_-]?key\s*[=:]\s*["']?([\w]{32,})["']?`),
	}

	// Email address patterns — mask the local-part to protect PII in logs.
	emailPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b[\w.+-]+@[\w.-]+\.[a-zA-Z]{2,}\b`),
	}
)

// RedactionConfig holds configuration for the redaction process.
type RedactionConfig struct {
	// EnableAPIKeys redacts API key patterns (default: true).
	EnableAPIKeys bool
	// EnableJWTs redacts JWT tokens (default: true).
	EnableJWTs bool
	// EnablePasswords redacts password patterns (default: true).
	EnablePasswords bool
	// EnableDBConnStrings redacts database connection strings (default: true).
	EnableDBConnStrings bool
	// EnablePrivateKeys redacts private key markers (default: true).
	EnablePrivateKeys bool
	// EnableGenericCredentials redacts generic credential patterns (default: true).
	EnableGenericCredentials bool
	// EnableEmails masks email addresses to protect PII in logs (default: true).
	EnableEmails bool
	// MaskLength is the length of the mask to use (default: 8).
	MaskLength int
}

// DefaultConfig returns the default redaction configuration.
func DefaultConfig() *RedactionConfig {
	return &RedactionConfig{
		EnableAPIKeys:            true,
		EnableJWTs:               true,
		EnablePasswords:          true,
		EnableDBConnStrings:      true,
		EnablePrivateKeys:        true,
		EnableGenericCredentials: true,
		EnableEmails:             true,
		MaskLength:               8,
	}
}

// DefaultRedactor is the default redactor using default configuration.
var DefaultRedactor = NewRedactor(DefaultConfig())

// Redactor handles sanitization of sensitive data.
type Redactor struct {
	config   *RedactionConfig
	maskPool sync.Pool
}

// NewRedactor creates a new Redactor with the given configuration.
func NewRedactor(config *RedactionConfig) *Redactor {
	return &Redactor{
		config: config,
		maskPool: sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
	}
}

// mask returns a string of asterisks of the configured length.
func (r *Redactor) mask() string {
	return strings.Repeat("*", r.config.MaskLength)
}

// Redact sanitizes sensitive data in the input string.
// Returns the sanitized string with secrets replaced by [REDACTED].
func (r *Redactor) Redact(input string) string {
	if input == "" {
		return ""
	}

	result := input

	// Apply patterns in order of sensitivity (most sensitive first).
	if r.config.EnablePrivateKeys {
		result = r.redactPrivateKeys(result)
	}
	if r.config.EnableDBConnStrings {
		result = r.redactDBConnections(result)
	}
	if r.config.EnableJWTs {
		result = r.redactJWTs(result)
	}
	if r.config.EnableAPIKeys {
		result = r.redactAPIKeys(result)
	}
	if r.config.EnablePasswords {
		result = r.redactPasswords(result)
	}
	if r.config.EnableGenericCredentials {
		result = r.redactGenericCredentials(result)
	}
	if r.config.EnableEmails {
		result = r.redactEmails(result)
	}

	return result
}

// redactEmails masks the local-part of email addresses so PII like.
// "admin@vyzorix.test" becomes "[EMAIL:REDACTED]" in logs.
func (r *Redactor) redactEmails(input string) string {
	result := input
	for _, pattern := range emailPatterns {
		result = pattern.ReplaceAllString(result, "[EMAIL:REDACTED]")
	}
	return result
}

// RedactStruct recursively redacts all string fields in a struct/map.
// Panics if struct fields are not exported.
func (r *Redactor) RedactStruct(v any) any {
	switch val := v.(type) {
	case string:
		return r.Redact(val)
	case map[string]any:
		result := make(map[string]any)
		for k, v := range val {
			result[k] = r.RedactStruct(v)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = r.RedactStruct(v)
		}
		return result
	default:
		return v
	}
}

// redactAPIKeys redacts API key and token patterns.
func (r *Redactor) redactAPIKeys(input string) string {
	result := input
	for _, pattern := range apiKeyPatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return r.mask() + "=[REDACTED]"
		})
	}
	return result
}

// redactJWTs redacts JWT tokens.
func (r *Redactor) redactJWTs(input string) string {
	result := input
	for _, pattern := range jwtPatterns {
		result = pattern.ReplaceAllString(result, "[JWT:"+r.mask()+"]")
	}
	return result
}

// redactPasswords redacts password patterns.
func (r *Redactor) redactPasswords(input string) string {
	result := input
	for _, pattern := range passwordPatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return r.mask() + "=[REDACTED]"
		})
	}
	return result
}

// redactDBConnections redacts database connection strings.
func (r *Redactor) redactDBConnections(input string) string {
	result := input
	for _, pattern := range dbPatterns {
		result = pattern.ReplaceAllString(result, "[DB_CONNECTION:REDACTED]")
	}
	return result
}

// redactPrivateKeys redacts private key markers.
func (r *Redactor) redactPrivateKeys(input string) string {
	result := input
	for _, pattern := range privateKeyPatterns {
		result = pattern.ReplaceAllString(result, "[PRIVATE_KEY:REDACTED]")
	}
	return result
}

// redactGenericCredentials redacts generic credential patterns.
func (r *Redactor) redactGenericCredentials(input string) string {
	result := input
	for _, pattern := range genericCredentialPatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return r.mask() + "=[REDACTED]"
		})
	}
	return result
}

// RedactMapKey redacts a map's values by key name.
// Only values whose keys match sensitive patterns are redacted.
func (r *Redactor) RedactMapKey(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}

	sensitiveKeys := []string{
		"password", "secret", "token", "api_key", "apikey",
		"auth", "credential", "private_key", "jwt", "bearer",
	}

	result := make(map[string]string)
	for k, v := range input {
		isSensitive := false
		lowerKey := strings.ToLower(k)
		for _, sensitive := range sensitiveKeys {
			if strings.Contains(lowerKey, sensitive) {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			result[k] = "[REDACTED]"
		} else {
			result[k] = v
		}
	}
	return result
}
