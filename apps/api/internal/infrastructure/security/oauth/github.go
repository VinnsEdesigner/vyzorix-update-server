// Package oauth provides OAuth utilities for GitHub and Google OAuth flows.
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitHubTokenResponse represents the token response from GitHub.
type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// GitHubUserInfo represents the GitHub user profile.
type GitHubUserInfo struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GitHubEmailInfo represents a GitHub email address.
type GitHubEmailInfo struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// GitHubConfig holds GitHub OAuth configuration.
type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// ExchangeGitHubCode exchanges an authorization code for a GitHub access token.
func ExchangeGitHubCode(ctx context.Context, code string, config GitHubConfig) (*GitHubTokenResponse, error) {
	data := url.Values{
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
		"code":          {code},
		"redirect_uri":  {config.RedirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request access token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub access token request returned status %d", resp.StatusCode)
	}

	var tokenResp GitHubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub access token response: %w", err)
	}

	return &tokenResp, nil
}

// FetchGitHubUserProfile retrieves the GitHub user profile using the access token.
func FetchGitHubUserProfile(ctx context.Context, accessToken string) (*GitHubUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call GitHub user API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub user API returned status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub user response: %w", err)
	}

	return &userInfo, nil
}

// FetchGitHubEmails retrieves all email addresses for the GitHub user.
func FetchGitHubEmails(ctx context.Context, accessToken string) ([]GitHubEmailInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call GitHub emails API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub emails API returned status %d: %s", resp.StatusCode, string(body))
	}

	var emails []GitHubEmailInfo
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub emails response: %w", err)
	}

	return emails, nil
}

// GetPrimaryEmail finds the primary verified email from the email list.
func GetPrimaryEmail(emails []GitHubEmailInfo) string {
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email
		}
	}
	if len(emails) > 0 {
		return emails[0].Email
	}
	return ""
}
