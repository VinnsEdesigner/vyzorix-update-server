// Package auth provides authentication utilities.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/lockout"
)

// Re-export from new location for backward compatibility.
type (
LockoutConfig  = lockout.Config
AccountLockout = lockout.Handler
LockoutReason  = lockout.Reason
)

var (
ErrInvalidPassword       = lockout.ErrInvalidPassword
ErrTokenGenerationFailed = lockout.ErrTokenGenerationFailed
)

func DefaultLockoutConfig() lockout.Config {
return lockout.DefaultConfig()
}

func NewAccountLockout(storage lockout.Storage, config lockout.Config) *lockout.Handler {
return lockout.New(storage, config)
}

func FakeHash(a, b string) bool {
return lockout.FakeHash(a, b)
}

func IsValidPassword(pwd string) bool {
return lockout.IsValidPassword(pwd)
}

func GenerateFakeToken() string {
return lockout.GenerateFakeToken()
}
