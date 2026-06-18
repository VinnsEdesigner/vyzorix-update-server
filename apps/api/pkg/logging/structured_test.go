// Package logging provides tests for structured logging.
package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// 4.13 STRUCTURED LOGGING - Tests (using slog)
// ============================================================================

func TestStructuredLogging(t *testing.T) {
	t.Run("production_logger", func(t *testing.T) {
		logger := NewProductionLogger()
		assert.NotNil(t, logger)
	})

	t.Run("development_logger", func(t *testing.T) {
		logger := NewDevelopmentLogger()
		assert.NotNil(t, logger)
	})

	t.Run("default_config", func(t *testing.T) {
		cfg := DefaultConfig()
		assert.Equal(t, "info", cfg.Level)
		assert.True(t, cfg.RedactPII)
		assert.NotEmpty(t, cfg.RedactKeys)
	})

	t.Run("custom_config", func(t *testing.T) {
		cfg := Config{
			Level:     "debug",
			RedactPII: true,
			RedactKeys: []string{"password", "token"},
		}
		assert.Equal(t, "debug", cfg.Level)
		assert.True(t, cfg.RedactPII)
	})
}

func TestRedactor(t *testing.T) {
	t.Run("redact_map_password_key", func(t *testing.T) {
		redactor := NewRedactor()
		input := map[string]interface{}{
			"username": "john",
			"password": "secret123",
		}
		result := redactor.RedactMap(input)
		assert.Equal(t, "john", result["username"])
		assert.Equal(t, "[REDACTED]", result["password"])
	})

	t.Run("fields_to_redact", func(t *testing.T) {
		assert.NotEmpty(t, FieldsToRedact)
		assert.Contains(t, FieldsToRedact, "password")
		assert.Contains(t, FieldsToRedact, "secret")
		assert.Contains(t, FieldsToRedact, "token")
		assert.Contains(t, FieldsToRedact, "api_key")
	})
}

func TestIsSensitiveKey(t *testing.T) {
	t.Run("sensitive_keys", func(t *testing.T) {
		sensitive := []string{
			"password",
			"secret",
			"token",
			"api_key",
			"apikey",
			"private",
			"credential",
		}
		for _, key := range sensitive {
			assert.True(t, isSensitiveKey(key), "Expected %s to be sensitive", key)
		}
	})

	t.Run("non_sensitive_keys", func(t *testing.T) {
		nonSensitive := []string{
			"username",
			"name",
			"email",
			"id",
		}
		for _, key := range nonSensitive {
			assert.False(t, isSensitiveKey(key), "Expected %s to not be sensitive", key)
		}
	})
}
