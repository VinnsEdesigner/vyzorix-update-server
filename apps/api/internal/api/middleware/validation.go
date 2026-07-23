// Package middleware provides HTTP middleware.
package middleware

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// =============================================================================.
// Validation Schemas (Zod-style patterns for Go).
// =============================================================================.

// ValidationError represents a validation error with field context.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	if len(ve) == 1 {
		return ve[0].Message
	}

	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, e.Message)
	}

	return strings.Join(msgs, "; ")
}

// =============================================================================.
// Regex Patterns.
// =============================================================================.

var (
	// EmailPattern validates email format (RFC 5322 simplified).
	EmailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// UUIDv7Pattern validates UUID format.
	UUIDv7Pattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-7[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

	// UUIDPattern validates any UUID format.
	UUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	IMEIPattern = regexp.MustCompile(`^[0-9]{15}$`)

	// PasswordPattern validates password meets policy.
	PasswordPattern = regexp.MustCompile(`^.{12,128}$`)

	// NamePattern validates display names (2-100 chars, letters/numbers/spaces).
	NamePattern = regexp.MustCompile(`^[a-zA-Z0-9\s\-'.]{2,100}$`)

	// DeviceIDPattern validates device ID format.
	DeviceIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-]{1,64}$`)

	// TokenPattern validates hex tokens (64-256 chars).
	TokenPattern = regexp.MustCompile(`^[a-fA-F0-9]{64,256}$`)

	// DispatchIDPattern validates dispatch ID format.
	DispatchIDPattern = regexp.MustCompile(`^[a-fA-F0-9]{32,128}$`)

	// CommandPattern validates command name format.
	CommandPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,63}$`)
)

// =============================================================================.
// Schema Validators.
// =============================================================================.

// LoginSchema validates login requests.
type LoginSchema struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *LoginSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	email := strings.TrimSpace(strings.ToLower(s.Email))
	if email == "" {
		errs = append(errs, ValidationError{Field: "email", Message: "email is required"})
	} else if !EmailPattern.MatchString(email) {
		errs = append(errs, ValidationError{Field: "email", Message: "invalid email format"})
	}

	if s.Password == "" {
		errs = append(errs, ValidationError{Field: "password", Message: "password is required"})
	}

	return errs
}

// RegisterSchema validates registration requests.
type RegisterSchema struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

func (s *RegisterSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	email := strings.TrimSpace(strings.ToLower(s.Email))
	if email == "" {
		errs = append(errs, ValidationError{Field: "email", Message: "email is required"})
	} else if !EmailPattern.MatchString(email) {
		errs = append(errs, ValidationError{Field: "email", Message: "invalid email format"})
	}

	if s.Password == "" {
		errs = append(errs, ValidationError{Field: "password", Message: "password is required"})
	} else if len(s.Password) < 12 {
		errs = append(errs, ValidationError{Field: "password", Message: "password must be at least 12 characters"})
	} else if len(s.Password) > 128 {
		errs = append(errs, ValidationError{Field: "password", Message: "password must be at most 128 characters"})
	}

	name := strings.TrimSpace(s.Name)
	if name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "name is required"})
	} else if !NamePattern.MatchString(name) {
		errs = append(errs, ValidationError{Field: "name", Message: "name must be 2-100 characters (letters, numbers, spaces, hyphens, apostrophes)"})
	}

	// Role is optional, but if provided must be valid.
	if s.Role != "" && s.Role != "operator" && s.Role != "super_admin" {
		errs = append(errs, ValidationError{Field: "role", Message: "role must be 'operator' or 'super_admin'"})
	}

	return errs
}

// ForgotPasswordSchema validates forgot password requests.
type ForgotPasswordSchema struct {
	Email string `json:"email"`
}

func (s *ForgotPasswordSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	email := strings.TrimSpace(strings.ToLower(s.Email))
	if email == "" {
		errs = append(errs, ValidationError{Field: "email", Message: "email is required"})
	} else if !EmailPattern.MatchString(email) {
		errs = append(errs, ValidationError{Field: "email", Message: "invalid email format"})
	}

	return errs
}

// ResetPasswordSchema validates password reset requests.
type ResetPasswordSchema struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (s *ResetPasswordSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	token := strings.TrimSpace(s.Token)
	if token == "" {
		errs = append(errs, ValidationError{Field: "token", Message: "token is required"})
	} else if !TokenPattern.MatchString(token) {
		errs = append(errs, ValidationError{Field: "token", Message: "invalid token format"})
	}

	if s.NewPassword == "" {
		errs = append(errs, ValidationError{Field: "new_password", Message: "new_password is required"})
	} else if len(s.NewPassword) < 12 {
		errs = append(errs, ValidationError{Field: "new_password", Message: "password must be at least 12 characters"})
	}

	return errs
}

// ChangePasswordSchema validates change password requests.
type ChangePasswordSchema struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (s *ChangePasswordSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	if s.OldPassword == "" {
		errs = append(errs, ValidationError{Field: "old_password", Message: "old_password is required"})
	}

	if s.NewPassword == "" {
		errs = append(errs, ValidationError{Field: "new_password", Message: "new_password is required"})
	} else if len(s.NewPassword) < 12 {
		errs = append(errs, ValidationError{Field: "new_password", Message: "password must be at least 12 characters"})
	}

	if s.OldPassword != "" && s.NewPassword != "" && s.OldPassword == s.NewPassword {
		errs = append(errs, ValidationError{Field: "new_password", Message: "new_password must be different from old_password"})
	}

	return errs
}

// DEPRECATED: DeviceRegisterSchema - /v1/device/register endpoint removed. Use /v1/device/inbox instead.
// type DeviceRegisterSchema struct {.
// 	DeviceID          string `json:"deviceId"`.
// 	FirebaseInstallID string `json:"firebaseInstallId"`.
// 	FCMToken          string `json:"fcmToken"`.
// 	AppVersion        string `json:"appVersion"`.
// 	DeviceClass       string `json:"deviceClass"`.
// }.
//
// func (s *DeviceRegisterSchema) Validate() ValidationErrors {.
// 	var errs ValidationErrors.
//
// 	if s.DeviceID == "" {.
// 		errs = append(errs, ValidationError{Field: "deviceId", Message: "deviceId is required"}).
// 	} else if !DeviceIDPattern.MatchString(s.DeviceID) {.
// 		errs = append(errs, ValidationError{Field: "deviceId", Message: "deviceId must be 1-64 alphanumeric characters with underscores or hyphens"}).
// 	}.
//
// 	if s.FirebaseInstallID == "" {.
// 		errs = append(errs, ValidationError{Field: "firebaseInstallId", Message: "firebaseInstallId is required"}).
// 	}.
//
// 	if s.AppVersion != "" && len(s.AppVersion) > 32 {.
// 		errs = append(errs, ValidationError{Field: "appVersion", Message: "appVersion must be at most 32 characters"}).
// 	}.
//
// 	if s.DeviceClass != "" && len(s.DeviceClass) > 64 {.
// 		errs = append(errs, ValidationError{Field: "deviceClass", Message: "deviceClass must be at most 64 characters"}).
// 	}.
//
// 	return errs.
// }.

// DeviceStatusSchema validates device status update requests.
type DeviceStatusSchema struct {
	Online   *bool  `json:"online"`
	DeviceID string `json:"deviceId"`
}

func (s *DeviceStatusSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	if s.DeviceID == "" {
		errs = append(errs, ValidationError{Field: "deviceId", Message: "deviceId is required"})
	} else if !DeviceIDPattern.MatchString(s.DeviceID) {
		errs = append(errs, ValidationError{Field: "deviceId", Message: "deviceId must be 1-64 alphanumeric characters"})
	}

	if s.Online == nil {
		errs = append(errs, ValidationError{Field: "online", Message: "online is required"})
	}

	return errs
}

// CommandExecuteSchema validates command execution requests.
type CommandExecuteSchema struct {
	Args       map[string]interface{} `json:"args"`
	DeviceID   string                 `json:"deviceId"`
	Command    string                 `json:"command"`
	DispatchID string                 `json:"dispatchId"`
	Nonce      string                 `json:"nonce"`
	Signature  string                 `json:"signature"`
	Timestamp  int64                  `json:"timestamp"`
}

func (s *CommandExecuteSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	if s.DeviceID == "" {
		errs = append(errs, ValidationError{Field: "deviceId", Message: "deviceId is required"})
	} else if !DeviceIDPattern.MatchString(s.DeviceID) {
		errs = append(errs, ValidationError{Field: "deviceId", Message: "deviceId must be 1-64 alphanumeric characters"})
	}

	if s.Command == "" {
		errs = append(errs, ValidationError{Field: "command", Message: "command is required"})
	} else if !CommandPattern.MatchString(s.Command) {
		errs = append(errs, ValidationError{Field: "command", Message: "command must start with letter and contain only alphanumeric and underscores (max 64 chars)"})
	}

	if s.DispatchID != "" && !DispatchIDPattern.MatchString(s.DispatchID) {
		errs = append(errs, ValidationError{Field: "dispatchId", Message: "invalid dispatchId format"})
	}

	return errs
}

// ClientSchema validates API client requests.
type ClientSchema struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

func (s *ClientSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	name := strings.TrimSpace(s.Name)
	if name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "name is required"})
	} else if len(name) < 2 || len(name) > 100 {
		errs = append(errs, ValidationError{Field: "name", Message: "name must be 2-100 characters"})
	}

	if s.Description != "" && len(s.Description) > 500 {
		errs = append(errs, ValidationError{Field: "description", Message: "description must be at most 500 characters"})
	}

	// Validate permissions.
	validPermissions := map[string]bool{
		"device:read":    true,
		"device:write":   true,
		"command:read":   true,
		"command:write":  true,
		"telemetry:read": true,
	}
	for _, p := range s.Permissions {
		if !validPermissions[p] {
			errs = append(errs, ValidationError{Field: "permissions", Message: "invalid permission: " + p})
		}
	}

	return errs
}

// OAuthCallbackSchema validates OAuth callback requests.
type OAuthCallbackSchema struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func (s *OAuthCallbackSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	if s.Code == "" {
		errs = append(errs, ValidationError{Field: "code", Message: "code is required"})
	}

	if s.State == "" {
		errs = append(errs, ValidationError{Field: "state", Message: "state is required"})
	}

	return errs
}

// MFASchema validates MFA verification requests.
type MFASchema struct {
	Code string `json:"code"`
}

func (s *MFASchema) Validate() ValidationErrors {
	var errs ValidationErrors

	if s.Code == "" {
		errs = append(errs, ValidationError{Field: "code", Message: "code is required"})
	} else if len(s.Code) < 6 || len(s.Code) > 8 {
		errs = append(errs, ValidationError{Field: "code", Message: "code must be 6-8 characters"})
	}

	return errs
}

// RefreshTokenSchema validates refresh token requests.
type RefreshTokenSchema struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *RefreshTokenSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	if s.RefreshToken == "" {
		errs = append(errs, ValidationError{Field: "refresh_token", Message: "refresh_token is required"})
	} else if !TokenPattern.MatchString(s.RefreshToken) {
		errs = append(errs, ValidationError{Field: "refresh_token", Message: "invalid refresh_token format"})
	}

	return errs
}

// UpdateSettingsSchema validates settings update requests.
type UpdateSettingsSchema struct {
	Name       *string `json:"name"`
	Email      *string `json:"email"`
	Role       *string `json:"role"`
	StrictHmac *bool   `json:"strictHmac"`
}

func (s *UpdateSettingsSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	if s.Name != nil {
		name := strings.TrimSpace(*s.Name)
		if name == "" {
			errs = append(errs, ValidationError{Field: "name", Message: "name cannot be empty"})
		} else if !NamePattern.MatchString(name) {
			errs = append(errs, ValidationError{Field: "name", Message: "name must be 2-100 characters (letters, numbers, spaces, hyphens, apostrophes)"})
		}
	}

	if s.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*s.Email))
		if email == "" {
			errs = append(errs, ValidationError{Field: "email", Message: "email cannot be empty"})
		} else if !EmailPattern.MatchString(email) {
			errs = append(errs, ValidationError{Field: "email", Message: "invalid email format"})
		}
	}

	if s.Role != nil && *s.Role != "operator" && *s.Role != "super_admin" {
		errs = append(errs, ValidationError{Field: "role", Message: "role must be 'operator' or 'super_admin'"})
	}

	return errs
}

// FCMTokenUpdateSchema validates FCM token update requests.
type FCMTokenUpdateSchema struct {
	FCMToken string `json:"fcmToken"`
}

func (s *FCMTokenUpdateSchema) Validate() ValidationErrors {
	var errs ValidationErrors

	if s.FCMToken == "" {
		errs = append(errs, ValidationError{Field: "fcmToken", Message: "fcmToken is required"})
	} else if len(s.FCMToken) > 500 {
		errs = append(errs, ValidationError{Field: "fcmToken", Message: "fcmToken is too long"})
	}

	return errs
}

// =============================================================================.
// Middleware Factory.
// =============================================================================.

// ValidationMiddleware creates a Gin middleware for validating request bodies.
func ValidationMiddleware(schema Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength == 0 {
			c.Next()
			return
		}

		if err := c.ShouldBindJSON(schema); err != nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error":   "bad_request",
				"message": "invalid request body format",
			})

			return
		}

		if errs := schema.Validate(); errs.HasErrors() {
			c.AbortWithStatusJSON(400, gin.H{
				"error":   "bad_request",
				"message": errs.Error(),
				"errors":  errs,
			})

			return
		}

		c.Next()
	}
}

// ValidationMiddlewareFunc creates a Gin middleware using a validation function.
func ValidationMiddlewareFunc(validateFn func(Validator) ValidationErrors) func(Validator) gin.HandlerFunc {
	return func(schema Validator) gin.HandlerFunc {
		return func(c *gin.Context) {
			if c.Request.ContentLength == 0 {
				c.Next()
				return
			}

			if err := c.ShouldBindJSON(schema); err != nil {
				c.AbortWithStatusJSON(400, gin.H{
					"error":   "bad_request",
					"message": "invalid request body format",
				})

				return
			}

			if errs := validateFn(schema); errs.HasErrors() {
				c.AbortWithStatusJSON(400, gin.H{
					"error":   "bad_request",
					"message": errs.Error(),
					"errors":  errs,
				})

				return
			}

			c.Next()
		}
	}
}

// ValidateRequest creates a middleware that validates request body against a schema.
// This is a helper for common validation patterns.
func ValidateRequest(schema Validator) gin.HandlerFunc {
	return ValidationMiddleware(schema)
}

// Validator defines the interface for request validators.
type Validator interface {
	Validate() ValidationErrors
}
