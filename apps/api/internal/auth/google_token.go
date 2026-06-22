// Package auth provides authentication utilities.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"context"

"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/oauth"
)

// Re-export from new location for backward compatibility.
type (
GoogleTokenVerifier = oauth.Verifier
GoogleClaims       = oauth.GoogleClaims
GoogleUserInfo    = oauth.GoogleUserInfo
)

var (
ErrInvalidGoogleToken     = oauth.ErrInvalidGoogleToken
ErrGoogleTokenExpired     = oauth.ErrGoogleTokenExpired
ErrGoogleTokenBadIssuer   = oauth.ErrGoogleTokenBadIssuer
ErrGoogleTokenBadAudience = oauth.ErrGoogleTokenBadAudience
)

func NewGoogleTokenVerifier(audience string) *oauth.Verifier {
return oauth.NewGoogleVerifier(audience)
}

func DecodeGoogleIDToken(token, audience string) (*oauth.GoogleClaims, error) {
return oauth.DecodeGoogleIDToken(token, audience)
}

func GetGoogleUserInfo(ctx context.Context, accessToken string) (*oauth.GoogleUserInfo, error) {
return oauth.GetGoogleUserInfo(ctx, accessToken)
}
