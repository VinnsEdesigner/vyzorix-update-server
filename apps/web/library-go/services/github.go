package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type GitHubUserInfo struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GitHubEmailInfo struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// ExchangeGitHubCode trades the short-lived OAuth authorization code for an access token
func ExchangeGitHubCode(code, clientID, clientSecret, redirectURI string) (*GitHubTokenResponse, error) {
	reqBody := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(reqBody.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed requesting GitHub access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub access token request returned status %d", resp.StatusCode)
	}

	var tokenResp GitHubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub access token response: %w", err)
	}

	return &tokenResp, nil
}

// FetchGitHubUserProfile retrieves profile info for the user
func FetchGitHubUserProfile(accessToken string) (*GitHubUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed calling GitHub profile endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub user profile API returned status %d", resp.StatusCode)
	}

	var uInfo GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&uInfo); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub user profile response: %w", err)
	}

	return &uInfo, nil
}

// FetchGitHubEmails retrieves email records for the user from GitHub API
func FetchGitHubEmails(accessToken string) ([]GitHubEmailInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed calling GitHub user emails API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub emails API returned status %d", resp.StatusCode)
	}

	var emailList []GitHubEmailInfo
	if err := json.NewDecoder(resp.Body).Decode(&emailList); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub emails list response: %w", err)
	}

	return emailList, nil
}
