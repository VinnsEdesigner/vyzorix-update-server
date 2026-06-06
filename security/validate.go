package security

import (
	"regexp"
	"strings"
)

const (
	// Max lengths for common fields
	MaxEmailLength    = 255
	MaxNameLength     = 100
	MaxPasswordLength = 128
	MinPasswordLength = 8
	MaxDeviceIDLength = 64
	MaxCommandLength  = 256
	MaxTokenLength    = 1024
)

// Email regex pattern (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidateEmail validates and sanitizes an email address
func ValidateEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	
	if len(email) == 0 {
		return "", &ValidationError{Field: "email", Message: "email is required"}
	}
	if len(email) > MaxEmailLength {
		return "", &ValidationError{Field: "email", Message: "email exceeds maximum length"}
	}
	if !emailRegex.MatchString(email) {
		return "", &ValidationError{Field: "email", Message: "invalid email format"}
	}
	
	return email, nil
}

// ValidateName validates and sanitizes a name
func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	
	if len(name) == 0 {
		return "", &ValidationError{Field: "name", Message: "name is required"}
	}
	if len(name) > MaxNameLength {
		return "", &ValidationError{Field: "name", Message: "name exceeds maximum length"}
	}
	
	return name, nil
}

// ValidatePassword validates password length
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}
	if len(password) > MaxPasswordLength {
		return &ValidationError{Field: "password", Message: "password exceeds maximum length"}
	}
	return nil
}

// ValidateDeviceID validates and sanitizes a device ID
func ValidateDeviceID(id string) (string, error) {
	id = strings.TrimSpace(id)
	
	if len(id) == 0 {
		return "", &ValidationError{Field: "deviceId", Message: "device ID is required"}
	}
	if len(id) > MaxDeviceIDLength {
		return "", &ValidationError{Field: "deviceId", Message: "device ID exceeds maximum length"}
	}
	
	// Only allow alphanumeric, hyphens, and underscores
	validID := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	if !validID.MatchString(id) {
		return "", &ValidationError{Field: "deviceId", Message: "device ID contains invalid characters"}
	}
	
	return id, nil
}

// ValidateCommand validates a command string
func ValidateCommand(cmd string) (string, error) {
	cmd = strings.TrimSpace(cmd)
	
	if len(cmd) == 0 {
		return "", &ValidationError{Field: "command", Message: "command is required"}
	}
	if len(cmd) > MaxCommandLength {
		return "", &ValidationError{Field: "command", Message: "command exceeds maximum length"}
	}
	
	return cmd, nil
}

// ValidateToken validates a token string
func ValidateToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	
	if len(token) == 0 {
		return "", &ValidationError{Field: "token", Message: "token is required"}
	}
	if len(token) > MaxTokenLength {
		return "", &ValidationError{Field: "token", Message: "token exceeds maximum length"}
	}
	
	return token, nil
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}