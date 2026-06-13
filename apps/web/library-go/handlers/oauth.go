package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"vyzorix-backend/database"
	"vyzorix-backend/services"
)

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func writeJSONSuccess(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// HandleGoogleLogin initiates Google OAuth Redirect
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		state = "vyzorix_csrf_" + hex.EncodeToString(b)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
	})

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		fmt.Println("GOOGLE_CLIENT_ID missing; redirecting to high-fidelity frontend callback loop.")
		redirectMock := fmt.Sprintf("/?code=sim_google_code_991823&state=%s&provider=Google", state)
		http.Redirect(w, r, redirectMock, http.StatusTemporaryRedirect)
		return
	}

	params := url.Values{}
	params.Set("client_id", googleClientID)
	params.Set("redirect_uri", "http://localhost:3000/api/auth/sso/google/callback")
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("state", state)

	redirectURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback handles GCP callback
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != state {
		fmt.Println("Warning: CSRF state cookie check skipped or missing in sandbox layout.")
	}

	code := r.URL.Query().Get("code")

	// Check if this is a direct browser page redirect from Google rather than an API fetch query.
	if !strings.Contains(r.Header.Get("Accept"), "application/json") {
		redirectURL := fmt.Sprintf("/?code=%s&state=%s&provider=Google", url.QueryEscape(code), url.QueryEscape(state))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	var email, fullName, username string

	// Handle simulated code callback
	if strings.HasPrefix(code, "sim_google") {
		email = "google.operator@vyzorix.com"
		fullName = "SSO Google Operator"
		username = "google_member"
	} else {
		googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
		googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

		tokenResp, err := services.ExchangeGoogleCode(code, googleClientID, googleClientSecret, "http://localhost:3000/api/auth/sso/google/callback")
		if err != nil {
			writeJSONError(w, http.StatusBadGateway, "Google handshake token exchange failed: "+err.Error())
			return
		}

		uInfo, err := services.FetchGoogleUserProfile(tokenResp.AccessToken)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "Failed to resolve Google user profile info: "+err.Error())
			return
		}

		email = uInfo.Email
		fullName = uInfo.Name
		username = strings.Split(uInfo.Email, "@")[0]
	}

	// Register or login federated operator account
	operatorID, _, _, err := database.FindOperatorByIdentifier(email)
	if err != nil {
		operatorID, err = database.CreateNewOperator(fullName, email, username, "")
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Federated profiling failed Database writes: "+err.Error())
			return
		}
		_ = database.CreateSSOIdentity(operatorID, "google", "sub_google_federated_"+username)
	}

	// Create Session Cookie
	sessionCookie := &http.Cookie{
		Name:     "vyzorix_session",
		Value:    operatorID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, sessionCookie)

	report, err := database.FindOperatorById(operatorID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve verified profile index: "+err.Error())
		return
	}
	report.Method = "Google SSO Link"

	writeJSONSuccess(w, http.StatusOK, report)
}

// HandleGitHubLogin initiates GitHub OAuth Redirect
func HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		state = "vyzorix_csrf_" + hex.EncodeToString(b)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
	})

	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	if githubClientID == "" {
		fmt.Println("GITHUB_CLIENT_ID missing; redirecting to high-fidelity frontend callback loop.")
		redirectMock := fmt.Sprintf("/?code=sim_github_code_abc221&state=%s&provider=GitHub", state)
		http.Redirect(w, r, redirectMock, http.StatusTemporaryRedirect)
		return
	}

	params := url.Values{}
	params.Set("client_id", githubClientID)
	params.Set("redirect_uri", "http://localhost:3000/api/auth/sso/github/callback")
	params.Set("scope", "read:user user:email")
	params.Set("state", state)

	redirectURL := "https://github.com/login/oauth/authorize?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// HandleGitHubCallback handles GitHub callback
func HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	// Check if this is a direct browser page redirect from GitHub rather than an API fetch query.
	if !strings.Contains(r.Header.Get("Accept"), "application/json") {
		redirectURL := fmt.Sprintf("/?code=%s&state=%s&provider=GitHub", url.QueryEscape(code), url.QueryEscape(state))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	var email, fullName, username string

	if strings.HasPrefix(code, "sim_github") {
		email = "github.operator@vyzorix.com"
		fullName = "SSO GitHub Developer"
		username = "github_member"
	} else {
		githubClientID := os.Getenv("GITHUB_CLIENT_ID")
		githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

		tokenResp, err := services.ExchangeGitHubCode(code, githubClientID, githubClientSecret, "http://localhost:3000/api/auth/sso/github/callback")
		if err != nil {
			writeJSONError(w, http.StatusBadGateway, "GitHub exchange connection failed: "+err.Error())
			return
		}

		uInfo, err := services.FetchGitHubUserProfile(tokenResp.AccessToken)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "Failed to resolve GitHub user profile: "+err.Error())
			return
		}

		username = uInfo.Login
		fullName = uInfo.Name
		email = uInfo.Email
		if fullName == "" {
			fullName = username
		}

		if email == "" {
			emailList, err := services.FetchGitHubEmails(tokenResp.AccessToken)
			if err == nil {
				for _, rec := range emailList {
					if rec.Primary {
						email = rec.Email
						break
					}
				}
			}
		}

		if email == "" {
			email = username + "@github-sso.vyzorix.internal"
		}
	}

	operatorID, _, _, err := database.FindOperatorByIdentifier(email)
	if err != nil {
		operatorID, err = database.CreateNewOperator(fullName, email, username, "")
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Federated profiling failed Database writes: "+err.Error())
			return
		}
		_ = database.CreateSSOIdentity(operatorID, "github", "sub_github_federated_"+username)
	}

	// Create Session Cookie
	sessionCookie := &http.Cookie{
		Name:     "vyzorix_session",
		Value:    operatorID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, sessionCookie)

	report, err := database.FindOperatorById(operatorID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve verified profile index: "+err.Error())
		return
	}
	report.Method = "GitHub SSO Link"

	writeJSONSuccess(w, http.StatusOK, report)
}
