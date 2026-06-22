// Package auth provides authentication utilities.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"time"

"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/jwt"
)

// Re-export from new location for backward compatibility.
type (
JWTManager     = jwt.Manager
OperatorClaims = jwt.OperatorClaims
)

var (
ErrInvalidToken = jwt.ErrInvalidToken
ErrExpiredToken = jwt.ErrExpiredToken
)

func NewJWTManager(secret string, expiry time.Duration, issuer string) *jwt.Manager {
return jwt.NewManager(secret, expiry, issuer)
}
