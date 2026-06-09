// Package security provides authentication utilities.
package security

import (
	"fmt"
	"strings"
	"unicode"
)

// PasswordPolicy defines the requirements for a valid password.
type PasswordPolicy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

// DefaultPasswordPolicy is the standard password policy.
// Requires: 8+ chars, uppercase, lowercase, digit, and special character.
// Use for system/app passwords that need maximum security.
var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:      8,
	MaxLength:      128,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: true,
}

// UserPasswordPolicy is a user-friendly password policy for personal account passwords.
// This policy is designed for scenarios where users set passwords for their own
// email accounts or personal services, which may have different restrictions than
// system-generated passwords.
//
// Requirements:
// - Minimum 12 characters (longer passwords provide security without requiring special chars)
// - At least 1 uppercase letter
// - At least 1 lowercase letter
// - At least 1 number
//
// Note: Special characters are NOT required because:
// 1. Many email providers don't allow special characters in passwords
// 2. Mobile keyboards make special characters harder to type
// 3. Password managers typically generate alphanumeric-only passwords
// 4. Length is more important than special characters for security
//
// Users can still use special characters if they want, but they're not required.
var UserPasswordPolicy = PasswordPolicy{
	MinLength:      12, // Longer minimum for passwords without special chars
	MaxLength:      128,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: false, // Not required for user passwords
}

// PasswordError represents validation failures for a password.
type PasswordError struct {
	Missing []string
}

func (e *PasswordError) Error() string {
	return "password does not meet requirements: " + strings.Join(e.Missing, "; ")
}

// ValidatePassword checks a password against the given policy.
// Returns nil if valid, or a PasswordError with details about failures.
func ValidatePassword(password string, policy PasswordPolicy) error {
	var missing []string

	if len(password) < policy.MinLength {
		missing = append(missing, fmt.Sprintf("minimum %d characters", policy.MinLength))
	}

	if len(password) > policy.MaxLength {
		missing = append(missing, fmt.Sprintf("maximum %d characters", policy.MaxLength))
	}

	if policy.RequireUpper && !containsUpper(password) {
		missing = append(missing, "at least 1 uppercase letter (A-Z)")
	}

	if policy.RequireLower && !containsLower(password) {
		missing = append(missing, "at least 1 lowercase letter (a-z)")
	}

	if policy.RequireDigit && !containsDigit(password) {
		missing = append(missing, "at least 1 number (0-9)")
	}

	if policy.RequireSpecial && !containsSpecial(password) {
		missing = append(missing, "at least 1 special character (!@#$%^&*()_+-=)")
	}

	if len(missing) > 0 {
		return &PasswordError{Missing: missing}
	}

	return nil
}

// containsUpper checks if the string contains at least one uppercase letter.
func containsUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

// containsLower checks if the string contains at least one lowercase letter.
func containsLower(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

// containsDigit checks if the string contains at least one digit.
func containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// containsSpecial checks if the string contains at least one special character.
func containsSpecial(s string) bool {
	specialChars := "!@#$%^&*()_+-="
	for _, r := range s {
		if strings.ContainsRune(specialChars, r) {
			return true
		}
	}
	return false
}

// PasswordStrength returns a score from 0-5 based on password complexity.
// This is for informational purposes only; validation should use ValidatePassword.
func PasswordStrength(password string) int {
	score := 0

	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if len(password) >= 16 {
		score++
	}
	if containsUpper(password) && containsLower(password) {
		score++
	}
	if containsDigit(password) || containsSpecial(password) {
		score++
	}

	return score
}
