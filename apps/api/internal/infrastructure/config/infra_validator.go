// Package config provides configuration validation from environment variables.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// ValidationError represents a configuration validation error with field context.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ConfigValidator validates configuration values including format and ranges.
type ConfigValidator struct {
	errs   []ValidationError
	config Config
	ssrCfg SSRConfig
}

// NewValidator creates a new config validator with the given configuration.
func NewValidator(cfg Config) *ConfigValidator {
	return &ConfigValidator{config: cfg, errs: make([]ValidationError, 0)}
}

// NewValidatorWithSSR creates a new config validator with SSR configuration.
func NewValidatorWithSSR(cfg Config, ssrCfg SSRConfig) *ConfigValidator {
	return &ConfigValidator{config: cfg, ssrCfg: ssrCfg, errs: make([]ValidationError, 0)}
}

// Validate runs all validation checks and returns all errors found.
func (v *ConfigValidator) Validate() []ValidationError {
	v.validatePort()
	v.validateSecrets()
	v.validateOrigins()
	v.validateURLs()
	v.validateSecuritySettings()
	v.validateTurnstile()
	v.validateSigning()
	v.validateCSRF()
	v.validateSession()
	v.validateRateLimits()
	v.validateTimeWindows()
	v.validateSSR()

	return v.errs
}

// validatePort validates PORT format and range.
func (v *ConfigValidator) validatePort() {
	port := v.config.Port
	if port == "" {
		v.addError("PORT", "port is required")
		return
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		v.addError("PORT", fmt.Sprintf("must be a valid number, got: %q", port))
		return
	}

	if p < 1 || p > 65535 {
		v.addError("PORT", fmt.Sprintf("must be between 1 and 65535, got: %d", p))
	}
}

// validateSecrets validates all secret values meet minimum requirements.
func (v *ConfigValidator) validateSecrets() {
	// SERVER_API_TOKEN: min 32 chars in production
	if v.config.ServerAPIToken == "" {
		if v.config.Env == "production" {
			v.addError("SERVER_API_TOKEN", "is required in production")
		}
	} else if v.config.Env == "production" && len(v.config.ServerAPIToken) < 32 {
		v.addError("SERVER_API_TOKEN", "must be at least 32 characters in production")
	}

	// JWT_SECRET: min 32 chars in production
	if v.config.JWTSecret == "" {
		if v.config.Env == "production" {
			v.addError("JWT_SECRET", "is required in production")
		}
	} else if v.config.Env == "production" && len(v.config.JWTSecret) < 32 {
		v.addError("JWT_SECRET", "must be at least 32 characters in production")
	}

	// SESSION_SECRET: must be exactly 32 chars (AES-256 key)
	if v.config.SessionSecret != "" && len(v.config.SessionSecret) != 32 {
		v.addError("SESSION_SECRET", "must be exactly 32 characters for AES-256 encryption")
	}

	// CSRF_SECRET: min 32 chars if enabled
	csrfSecret := os.Getenv("CSRF_SECRET")
	if getBool("CSRF_ENABLED", false) && len(csrfSecret) < 32 {
		v.addError("CSRF_SECRET", "must be at least 32 characters when CSRF is enabled")
	}

	// TURNSTILE_SECRET: required if enabled
	turnstileSecret := os.Getenv("TURNSTILE_SECRET")
	if getBool("TURNSTILE_ENABLED", false) && turnstileSecret == "" {
		v.addError("TURNSTILE_SECRET", "is required when Turnstile is enabled")
	}
}

// validateOrigins validates ALLOWED_ORIGINS format.
func (v *ConfigValidator) validateOrigins() {
	if len(v.config.AllowedOrigins) == 0 {
		return
	}

	for _, origin := range v.config.AllowedOrigins {
		if origin == "*" {
			if v.config.Env == "production" {
				v.addError("ALLOWED_ORIGINS", "wildcard '*' is not allowed in production")
			}

			continue
		}

		if _, err := url.Parse(origin); err != nil {
			v.addError("ALLOWED_ORIGINS", fmt.Sprintf("invalid origin format: %q", origin))
		}
	}
}

// validateURLs validates BASE_URL and FRONTEND_URL formats.
func (v *ConfigValidator) validateURLs() {
	if v.config.BaseURL != "" {
		if _, err := url.Parse(v.config.BaseURL); err != nil {
			v.addError("BASE_URL", fmt.Sprintf("must be a valid URL, got: %q", v.config.BaseURL))
		}
	}

	if v.config.FrontendURL != "" {
		if _, err := url.Parse(v.config.FrontendURL); err != nil {
			v.addError("FRONTEND_URL", fmt.Sprintf("must be a valid URL, got: %q", v.config.FrontendURL))
		}
	}
}

// validateSecuritySettings validates security-related settings.
func (v *ConfigValidator) validateSecuritySettings() {
	_ = v.config.EnforceHMAC
}

// validateTurnstile validates Turnstile configuration.
func (v *ConfigValidator) validateTurnstile() {
	turnstileEnabled := getBool("TURNSTILE_ENABLED", false)
	turnstileSecret := os.Getenv("TURNSTILE_SECRET")

	if turnstileEnabled && turnstileSecret == "" {
		v.addError("TURNSTILE_SECRET", "is required when TURNSTILE_ENABLED=true")
	}
}

// validateSigning validates request signing configuration.
func (v *ConfigValidator) validateSigning() {
	signingEnabled := getBool("REQUEST_SIGNING_ENABLED", false)
	if !signingEnabled {
		return
	}

	// SIGNING_TIMESTAMP_WINDOW: must be 60-600 seconds
	if window := os.Getenv("SIGNING_TIMESTAMP_WINDOW"); window != "" {
		w, err := strconv.Atoi(window)
		if err != nil || w < 60 || w > 600 {
			v.addError("SIGNING_TIMESTAMP_WINDOW", "must be between 60 and 600 seconds")
		}
	}

	// SIGNING_MAX_CACHE_SIZE: must be positive
	if cacheSize := os.Getenv("SIGNING_MAX_CACHE_SIZE"); cacheSize != "" {
		cs, err := strconv.Atoi(cacheSize)
		if err != nil || cs <= 0 {
			v.addError("SIGNING_MAX_CACHE_SIZE", "must be a positive integer")
		}
	}

	// SIGNING_GRACE_PERIOD: must be positive
	if grace := os.Getenv("SIGNING_GRACE_PERIOD"); grace != "" {
		g, err := strconv.Atoi(grace)
		if err != nil || g <= 0 {
			v.addError("SIGNING_GRACE_PERIOD", "must be a positive integer (seconds)")
		}
	}
}

// validateCSRF validates CSRF configuration.
func (v *ConfigValidator) validateCSRF() {
	csrfEnabled := getBool("CSRF_ENABLED", false)
	if !csrfEnabled {
		return
	}

	csrfSecret := os.Getenv("CSRF_SECRET")
	if csrfSecret == "" {
		v.addError("CSRF_SECRET", "is required when CSRF_ENABLED=true")
	} else if len(csrfSecret) < 32 {
		v.addError("CSRF_SECRET", "must be at least 32 characters")
	}
}

// validateSession validates session configuration.
func (v *ConfigValidator) validateSession() {
	if v.config.SessionMaxAge <= 0 {
		v.addError("SESSION_MAX_AGE_SECONDS", "must be a positive integer")
	} else if v.config.SessionMaxAge < 300 {
		v.addError("SESSION_MAX_AGE_SECONDS", "should be at least 300 seconds (5 minutes)")
	}
}

// validateRateLimits validates rate limit configuration.
func (v *ConfigValidator) validateRateLimits() {
	validatePositiveInt := func(key string) {
		if val := os.Getenv(key); val != "" {
			n, err := strconv.Atoi(val)
			if err != nil || n <= 0 {
				v.addError(key, "must be a positive integer")
			}
		}
	}

	validatePositiveInt("RATE_LIMIT_REQUESTS")
	validatePositiveInt("RATE_LIMIT_WINDOW_SECONDS")
	validatePositiveInt("AUTH_RATE_LIMIT_REQUESTS")
	validatePositiveInt("AUTH_RATE_LIMIT_WINDOW_SECONDS")
}

// validateTimeWindows validates time window configurations.
func (v *ConfigValidator) validateTimeWindows() {
	if v.config.JWTDuration <= 0 {
		v.addError("JWT_DURATION_HOURS", "must be a positive duration")
	}

	if v.config.EmailVerifyTokenExpiry < 5*time.Minute {
		v.addError("EMAIL_VERIFY_TOKEN_EXPIRY_HOURS", "should be at least 5 minutes")
	}

	if v.config.PasswordResetTokenExpiry < 5*time.Minute {
		v.addError("PASSWORD_RESET_TOKEN_EXPIRY_MINUTES", "should be at least 5 minutes")
	}
}

// addError adds a validation error.
func (v *ConfigValidator) addError(field, message string) {
	v.errs = append(v.errs, ValidationError{Field: field, Message: message})
}

// Err returns validation errors if any.
func (v *ConfigValidator) Err() error {
	errs := v.Validate()
	if len(errs) > 0 {
		msgs := make([]string, 0, len(errs))
		for _, e := range errs {
			msgs = append(msgs, e.Error())
		}

		return fmt.Errorf("configuration validation failed: %s", strings.Join(msgs, "; "))
	}

	return nil
}

// validateSSR validates SSR server configuration.
func (v *ConfigValidator) validateSSR() {
	if !v.ssrCfg.EnableSSR {
		return
	}

	// SSR_SERVER_URL is required when SSR is enabled
	if v.ssrCfg.SSRServerURL == "" {
		v.addError("SSR_SERVER_URL", "is required when SSR_ENABLED=true")
		return
	}

	// Validate URL format
	if _, err := url.Parse(v.ssrCfg.SSRServerURL); err != nil {
		v.addError("SSR_SERVER_URL", fmt.Sprintf("must be a valid URL, got: %q", v.ssrCfg.SSRServerURL))
	}

	// SSR_BUILD_TIMEOUT: should be reasonable (30s - 10min)
	if v.ssrCfg.SSRBuildTimeout < 30 || v.ssrCfg.SSRBuildTimeout > 600 {
		v.addError("SSR_BUILD_TIMEOUT", "should be between 30 and 600 seconds")
	}

	// SSR_HEALTH_CHECK_INTERVAL: should be reasonable (5s - 5min)
	if v.ssrCfg.SSRHealthCheckInterval < 5 || v.ssrCfg.SSRHealthCheckInterval > 300 {
		v.addError("SSR_HEALTH_CHECK_INTERVAL", "should be between 5 and 300 seconds")
	}

	// SSR_RETRY_ATTEMPTS: should be reasonable (1 - 10)
	if v.ssrCfg.SSRRetryAttempts < 1 || v.ssrCfg.SSRRetryAttempts > 10 {
		v.addError("SSR_RETRY_ATTEMPTS", "should be between 1 and 10")
	}
}

// Validate runs validation on the given config and returns an error if invalid.
func Validate(cfg Config) error {
	validator := NewValidator(cfg)
	return validator.Err()
}

// ValidateWithSSR runs validation on the given config with SSR config.
func ValidateWithSSR(cfg Config, ssrCfg SSRConfig) error {
	validator := NewValidatorWithSSR(cfg, ssrCfg)
	return validator.Err()
}
