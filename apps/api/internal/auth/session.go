// Package auth provides authentication utilities.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/session"
)

// Re-export from new location for backward compatibility.
type SessionManager = session.Manager

const (
CookieName   = session.CookieName
CookieMaxAge = session.CookieMaxAge
CookiePath   = session.CookiePath
)

var (
ErrInvalidCookie    = session.ErrInvalidCookie
ErrExpiredCookie    = session.ErrExpiredCookie
ErrDecryptionFailed = session.ErrDecryptionFailed
)

func NewSessionManager(secret string) *session.Manager {
return session.NewManager(secret)
}

func HashOperatorID(id string) string {
return session.HashOperatorID(id)
}

