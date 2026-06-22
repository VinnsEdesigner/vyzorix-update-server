// Package auth provides authentication utilities.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/validate"
)

// Re-export from new location for backward compatibility.
type (
EmailValidator         = validate.Validator
ValidationError        = validate.Error
EmailValidationResult  = validate.Result
EmailValidationError   = validate.ValidationError
)

const (
MaxEmailLength    = validate.MaxEmailLength
MaxNameLength     = validate.MaxNameLength
MaxPasswordLength = validate.MaxPasswordLength
MinPasswordLength = validate.MinPasswordLength
MaxDeviceIDLength = validate.MaxDeviceIDLength
MaxCommandLength  = validate.MaxCommandLength
MaxTokenLength    = validate.MaxTokenLength
)

var (
DefaultEmailValidator = validate.DefaultValidator()
StrictEmailValidator = validate.StrictValidator()
)

func ValidateEmail(email string) (string, error) {
return validate.Email(email)
}

func ValidateEmailFull(email string, validator *validate.Validator) *validate.Result {
return validate.EmailFull(email, validator)
}

func ValidateName(name string) (string, error) {
return validate.Name(name)
}

func ValidatePasswordLength(pwd string) error {
return validate.PasswordLength(pwd)
}

func ValidateDeviceID(id string) (string, error) {
return validate.DeviceID(id)
}

func ValidateCommand(cmd string) (string, error) {
return validate.Command(cmd)
}

func ValidateToken(token string) (string, error) {
return validate.Token(token)
}

func SanitizeString(s string, maxLen int) string {
return validate.Sanitize(s, maxLen)
}

func ContainsInvalidUTF8(s string) bool {
return validate.ContainsInvalidUTF8(s)
}

func ContainsControlCharacters(s string) bool {
return validate.ContainsControlCharacters(s)
}

func ExtractDomain(email string) string {
return validate.ExtractDomain(email)
}

func IsEmailDomainDisposable(email string) bool {
return validate.IsDisposableDomain(email)
}

func ClearMXCache() {
validate.ClearMXCache()
}

func NormalizeEmailForComparison(email string) string {
return validate.NormalizeEmail(email)
}

func ValidateEmailURI(uri string) error {
return validate.EmailURI(uri)
}
