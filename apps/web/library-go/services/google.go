package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
}

type GoogleUserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
}

// ExchangeGoogleCode contacts Google APIs to trade an authorization code for tokens
func ExchangeGoogleCode(code, clientID, clientSecret, redirectURI string) (*GoogleTokenResponse, error) {
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed token exchange HTTP post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google token exchange returned status %d", resp.StatusCode)
	}

	var tokenResp GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode Google token response: %w", err)
	}

	return &tokenResp, nil
}

// FetchGoogleUserProfile uses the access token to fetch the OIDC claims
func FetchGoogleUserProfile(accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed requesting userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google userinfo returned status %d", resp.StatusCode)
	}

	var uInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&uInfo); err != nil {
		return nil, fmt.Errorf("failed decoding userinfo json: %w", err)
	}

	return &uInfo, nil
}
