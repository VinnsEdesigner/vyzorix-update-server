// Package auth provides authentication utilities including session revocation.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"time"

"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/revocation"
)

// Re-export from new location for backward compatibility.
type (
RevocationReason = revocation.Reason
RevocationEntry  = revocation.Entry
RevocationList   = revocation.List
)

const (
RevokeReasonLogout         = revocation.ReasonLogout
RevokeReasonPasswordChange = revocation.ReasonPasswordChange
RevokeReasonAdmin          = revocation.ReasonAdmin
RevokeReasonSecurity       = revocation.ReasonSecurity
RevokeReasonExpired        = revocation.ReasonExpired
)

func NewRevocationList(maxEntries int, ttl time.Duration) *revocation.List {
return revocation.New(maxEntries, ttl)
}

func DefaultRevocationList() *revocation.List {
return revocation.Default()
}
