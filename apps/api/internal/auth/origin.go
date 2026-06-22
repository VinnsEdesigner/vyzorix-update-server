// Package auth provides authentication utilities.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/origin"
)

// Re-export from new location for backward compatibility.
type OriginValidator = origin.Validator

func NewOriginValidator(origins []string) *origin.Validator {
return origin.NewValidator(origins)
}
