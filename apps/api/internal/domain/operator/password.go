// Package operator provides domain entities and validation for operators.
package operator

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
var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:      8,
	MaxLength:      128,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: true,
}

// UserPasswordPolicy is a user-friendly password policy for personal accounts.
var UserPasswordPolicy = PasswordPolicy{
	MinLength:      12,
	MaxLength:      128,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: false,
}

// Policy is an alias for PasswordPolicy.
type Policy = PasswordPolicy
type Error = PasswordError

var DefaultPolicy = DefaultPasswordPolicy
var UserPolicy = UserPasswordPolicy

// ValidatePassword validates a password against the given policy.
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

// PasswordError represents validation failures for a password.
type PasswordError struct {
	Missing []string
}

func (e *PasswordError) Error() string {
	return "password does not meet requirements: " + strings.Join(e.Missing, "; ")
}

func containsUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}

	return false
}

func containsLower(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}

	return false
}

func containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}

	return false
}

func containsSpecial(s string) bool {
	specialChars := "!@#$%^&*()_+-="
	for _, r := range s {
		if strings.ContainsRune(specialChars, r) {
			return true
		}
	}

	return false
}

// Strength returns a score from 0-5 based on password complexity.
func Strength(password string) int {
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
