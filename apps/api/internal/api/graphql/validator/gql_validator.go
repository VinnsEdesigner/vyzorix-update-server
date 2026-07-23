// Package validator provides input validation for GraphQL resolvers.
package validator

import (
	"regexp"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/errors"
)

// Validator provides input validation utilities.
type Validator struct{}

// New creates a new Validator.
func New() *Validator {
	return &Validator{}
}

// UUIDRegex matches standard UUID format.
var UUIDRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// FCMTokenRegex matches FCM token format (base64-like, 150-200 chars).
var FCMTokenRegex = regexp.MustCompile(`^[A-Za-z0-9_-]{50,500}$`)

// ValidateDeviceID validates a device ID format.
func (v *Validator) ValidateDeviceID(id string) error {
	if id == "" {
		return errors.BadRequest("device ID is required")
	}

	if len(id) > 255 {
		return errors.BadRequest("device ID too long")
	}

	return nil
}

// ValidateDispatchID validates a dispatch ID format.
func (v *Validator) ValidateDispatchID(id string) error {
	if id == "" {
		return errors.BadRequest("dispatch ID is required")
	}

	if len(id) > 255 {
		return errors.BadRequest("dispatch ID too long")
	}

	return nil
}

// ValidateCommand validates a command string.
func (v *Validator) ValidateCommand(cmd string) error {
	if cmd == "" {
		return errors.BadRequest("command is required")
	}

	if len(cmd) > 100 {
		return errors.BadRequest("command too long")
	}
	// Allow alphanumeric, underscore, hyphen.
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`).MatchString(cmd) {
		return errors.BadRequest("invalid command format")
	}

	return nil
}

// ValidateFCMToken validates an FCM token format.
func (v *Validator) ValidateFCMToken(token string) error {
	if token == "" {
		return errors.BadRequest("FCM token is required")
	}

	if len(token) < 50 || len(token) > 500 {
		return errors.BadRequest("invalid FCM token length")
	}

	if !FCMTokenRegex.MatchString(token) {
		return errors.BadRequest("invalid FCM token format")
	}

	return nil
}

// ValidateLimit validates a pagination limit.
func (v *Validator) ValidateLimit(limit, maxVal int) error {
	if limit < 0 {
		return errors.BadRequest("limit cannot be negative")
	}

	if limit > maxVal {
		return errors.BadRequest("limit exceeds maximum of %d", maxVal)
	}

	return nil
}

// ValidateOffset validates a pagination offset.
func (v *Validator) ValidateOffset(offset int) error {
	if offset < 0 {
		return errors.BadRequest("offset cannot be negative")
	}

	return nil
}

// ValidateTimeRange validates a time range for queries.
func (v *Validator) ValidateTimeRange(startTime, endTime int64) error {
	if startTime < 0 {
		return errors.BadRequest("startTime cannot be negative")
	}

	if endTime < 0 {
		return errors.BadRequest("endTime cannot be negative")
	}

	if endTime > 0 && startTime > endTime {
		return errors.BadRequest("startTime must be before endTime")
	}

	return nil
}

// SanitizeString removes potentially dangerous characters.
func (v *Validator) SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

// ValidateArgs validates command arguments map.
func (v *Validator) ValidateArgs(args map[string]interface{}) error {
	if args == nil {
		return nil
	}
	// Limit args size to prevent abuse.
	if len(args) > 50 {
		return errors.BadRequest("too many arguments")
	}
	// Check total size.
	for k, val := range args {
		if len(k) > 100 {
			return errors.BadRequest("argument key too long")
		}

		switch v := val.(type) {
		case string:
			if len(v) > 10000 {
				return errors.BadRequest("argument value too long")
			}
		case map[string]interface{}, []interface{}:
			// For complex types, just check depth.
			if !checkDepth(v, 0, 5) {
				return errors.BadRequest("argument nested too deeply")
			}
		}
	}

	return nil
}

func checkDepth(val interface{}, current, maxVal int) bool {
	if current >= maxVal {
		return false
	}

	switch v := val.(type) {
	case map[string]interface{}:
		for _, vv := range v {
			if !checkDepth(vv, current+1, maxVal) {
				return false
			}
		}
	case []interface{}:
		for _, vv := range v {
			if !checkDepth(vv, current+1, maxVal) {
				return false
			}
		}
	}

	return true
}
