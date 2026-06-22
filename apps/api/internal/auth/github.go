// Package auth provides GitHub OAuth utilities.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"context"

"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/oauth"
)

// Re-export from new location for backward compatibility.
type (
GitHubTokenResponse = oauth.GitHubTokenResponse
GitHubUserInfo     = oauth.GitHubUserInfo
GitHubEmailInfo    = oauth.GitHubEmailInfo
GitHubOAuthConfig  = oauth.GitHubConfig
)

func ExchangeGitHubCode(ctx context.Context, code string, config oauth.GitHubConfig) (*oauth.GitHubTokenResponse, error) {
return oauth.ExchangeGitHubCode(ctx, code, config)
}

func FetchGitHubUserProfile(ctx context.Context, accessToken string) (*oauth.GitHubUserInfo, error) {
return oauth.FetchGitHubUserProfile(ctx, accessToken)
}

func FetchGitHubEmails(ctx context.Context, accessToken string) ([]oauth.GitHubEmailInfo, error) {
return oauth.FetchGitHubEmails(ctx, accessToken)
}

func GetPrimaryEmail(emails []oauth.GitHubEmailInfo) string {
return oauth.GetPrimaryEmail(emails)
}
