// Package auth provides authentication utilities.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/password"
)

// Re-export from new location for backward compatibility.
type (
PasswordPolicy = password.Policy
PasswordError  = password.Error
)

var (
DefaultPasswordPolicy = password.DefaultPolicy
UserPasswordPolicy    = password.UserPolicy
)

func ValidatePassword(pwd string, policy password.Policy) error {
return password.Validate(pwd, policy)
}

func PasswordStrength(pwd string) int {
return password.Strength(pwd)
}
