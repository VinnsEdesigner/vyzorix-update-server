package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// OAuth error codes for frontend handling.
const (
	ErrOAuthEmailRequired    = "email_required"
	ErrOAuthEmailNotVerified = "email_not_verified"
	ErrOAuthLoginFailed      = "login_failed"
	ErrOAuthStateInvalid     = "state_invalid"
	ErrOAuthExchangeFailed   = "token_exchange_failed"
	ErrOAuthConfigMissing    = "oauth_not_configured"
)

// OAuthErrorDetails holds structured error information for the frontend.
type OAuthErrorDetails struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Provider   string `json:"provider,omitempty"`
	HelpURL    string `json:"helpUrl,omitempty"`
	Retryable  bool   `json:"retryable"`
}

// OAuthHandler handles OAuth endpoints.
type OAuthHandler struct {
	oauthStateRepo OAuthStateProvider
	authService    *appsvc.AuthService
	sessionMgr     *infraauth.SessionManager
	googleVer      *infraauth.GoogleTokenVerifier
	logger         *slog.Logger
	presenter      *response.Presenter
	config         config.Config
}

// OAuthStateProvider interface for OAuth state operations.
// 8: This interface allows persisting OAuth state to prevent CSRF attacks.
type OAuthStateProvider interface {
	// Create stores a new OAuth state and returns the state ID.
	Create(ctx context.Context, state, redirectURL, provider string) (string, error)
	// Validate retrieves and validates an OAuth state, returning the redirect URL and state.
	Validate(ctx context.Context, state string) (redirectURL string, stateID string, err error)
	// Delete removes an OAuth state after use.
	Delete(ctx context.Context, stateID string) error
}

// NewOAuthHandler creates a new OAuthHandler.
func NewOAuthHandler(authService *appsvc.AuthService, sessionMgr *infraauth.SessionManager, cfg config.Config, googleVer *infraauth.GoogleTokenVerifier, presenter *response.Presenter) *OAuthHandler {
	return &OAuthHandler{
		authService: authService,
		sessionMgr:  sessionMgr,
		config:      cfg,
		googleVer:   googleVer,
		presenter:   presenter,
	}
}

// WithOAuthStateRepo sets the OAuth state repository for persistent state storage.
func (h *OAuthHandler) WithOAuthStateRepo(repo OAuthStateProvider) *OAuthHandler {
	h.oauthStateRepo = repo
	return h
}

// WithLogger sets the logger for the handler.
func (h *OAuthHandler) WithLogger(logger *slog.Logger) *OAuthHandler {
	h.logger = logger
	return h
}

// getOAuthRedirectURL builds a redirect URL with error parameters for OAuth errors.
func (h *OAuthHandler) getOAuthRedirectURL(redirectURL string, err OAuthErrorDetails) string {
	baseURL := redirectURL
	if baseURL == "" {
		baseURL = h.config.FrontendURL
		if baseURL == "" {
			baseURL = "http://localhost:5173"
		}
	}

	errURL := fmt.Sprintf("%s/auth/callback?oauth=error&code=%s&message=%s",
		baseURL,
		url.QueryEscape(err.Code),
		url.QueryEscape(err.Message),
	)

	if err.Provider != "" {
		errURL += "&provider=" + url.QueryEscape(err.Provider)
	}
	if !err.Retryable {
		errURL += "&retryable=false"
	}

	return errURL
}

// getDefaultRedirectURL returns the default redirect URL.
func (h *OAuthHandler) getDefaultRedirectURL(redirectURL string) string {
	baseURL := redirectURL
	if baseURL == "" {
		baseURL = h.config.FrontendURL
		if baseURL == "" {
			baseURL = "http://localhost:5173"
		}
	}
	return baseURL
}

// GoogleLogin handles GET /v1/auth/google.
func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	if h.config.GoogleOAuthClientID == "" || h.config.GoogleOAuthClientSecret == "" {
		h.presenter.NotImplemented(c, "Google OAuth is not configured on this server")
		return
	}

	// OAuth state repository is required for CSRF protection.
	if h.oauthStateRepo == nil {
		h.presenter.InternalError(c, "OAuth state repository is required for CSRF protection")
		return
	}

	frontendURL := h.config.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Generate random state for CSRF protection.
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		h.presenter.InternalError(c, "failed to generate OAuth state")
		return
	}
	state := hex.EncodeToString(stateBytes)

	// Persist OAuth state to database for CSRF validation.
	if _, err := h.oauthStateRepo.Create(c.Request.Context(), state, frontendURL, "google"); err != nil {
		h.presenter.InternalError(c, "failed to create OAuth state")
		return
	}

	callbackURL := h.config.BaseURL + "/v1/auth/google/callback"
	googleURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&state=%s",
		url.QueryEscape(h.config.GoogleOAuthClientID),
		url.QueryEscape(callbackURL),
		url.QueryEscape("openid email profile"),
		url.QueryEscape(state),
	)
	c.Redirect(http.StatusTemporaryRedirect, googleURL)
}

// GoogleCallback handles GET /v1/auth/google/callback.
func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		h.presenter.BadRequest(c, "missing authorization code from Google")
		return
	}

	if state == "" {
		h.presenter.BadRequest(c, "missing state parameter")
		return
	}

	// 8: Validate state from database if repository is configured.
	var redirectURL string
	if h.oauthStateRepo == nil {
		h.presenter.InternalError(c, "OAuth state repository not configured")
		return
	}

	var err error
	redirectURL, state, err = h.oauthStateRepo.Validate(c.Request.Context(), state)
	if err != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthStateInvalid,
			Message:   "OAuth state expired or invalid. Please try signing in again.",
			Provider:  "google",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}
	// Delete the state after successful validation (one-time use).
	_ = h.oauthStateRepo.Delete(c.Request.Context(), state)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Exchange code for tokens.
	tokenURL := "https://oauth2.googleapis.com/token"
	tokenReq := map[string]string{
		"code":          code,
		"client_id":     h.config.GoogleOAuthClientID,
		"client_secret": h.config.GoogleOAuthClientSecret,
		"redirect_uri":  h.config.BaseURL + "/v1/auth/google/callback",
		"grant_type":    "authorization_code",
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		IDToken     string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if postErr := postJSON(ctx, tokenURL, tokenReq, &tokenResp); postErr != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthExchangeFailed,
			Message:   "Failed to exchange code with Google. Please try again.",
			Provider:  "google",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Verify ID token.
	googleClaims, err := h.googleVer.Verify(tokenResp.IDToken)
	if err != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthExchangeFailed,
			Message:   "Failed to verify your Google identity. Please try again.",
			Provider:  "google",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Check for email requirement - Google requires verified email.
	if googleClaims.Email == "" {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthEmailRequired,
			Message:   "Your Google account must have a verified email address to sign up. Please add an email to your Google Account and verify it, then try again.",
			Provider:  "google",
			HelpURL:   "https://support.google.com/accounts/answer/97060",
			Retryable: false,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	if !googleClaims.EmailVerified {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthEmailNotVerified,
			Message:   "Your Google email must be verified to sign up. Please verify your email in Google Account settings, then try again.",
			Provider:  "google",
			HelpURL:   "https://support.google.com/accounts/answer/97060",
			Retryable: false,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Find or create operator via application service.
	result, err := h.authService.FindOrCreateGoogleOperator(ctx, googleClaims.Sub, googleClaims.Email, googleClaims.Name)
	if err != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthLoginFailed,
			Message:   "Failed to create or access your account. Please try again or contact support.",
			Provider:  "google",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Create session (validates operator was found/created).
	session, err := h.authService.CreateSession(ctx, result.Operator.ID)
	if err != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthLoginFailed,
			Message:   "Failed to create session. Please try again.",
			Provider:  "google",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Set session cookie with session ID.
	cookie, err := h.sessionMgr.CreateCookieWithExpiry(session.ID, h.config.SessionMaxAge)
	if err != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthLoginFailed,
			Message:   "Failed to create session. Please try again.",
			Provider:  "google",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	http.SetCookie(c.Writer, cookie)

	// Redirect to frontend with success.
	baseURL := h.getDefaultRedirectURL(redirectURL)
	successURL := fmt.Sprintf("%s/auth/callback?oauth=success&new=%t", baseURL, result.IsNew)
	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

// GitHubLogin handles GET /v1/auth/github.
func (h *OAuthHandler) GitHubLogin(c *gin.Context) {
	if h.config.GitHubOAuthClientID == "" || h.config.GitHubOAuthClientSecret == "" {
		h.presenter.NotImplemented(c, "GitHub OAuth is not configured on this server")
		return
	}

	// OAuth state repository is required for CSRF protection.
	if h.oauthStateRepo == nil {
		h.presenter.InternalError(c, "OAuth state repository is required for CSRF protection")
		return
	}

	frontendURL := h.config.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	callbackURL := h.config.BaseURL + "/v1/auth/github/callback"

	state := c.Query("state")
	if state == "" {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		state = hex.EncodeToString(b)
	}

	// Persist OAuth state to database for CSRF validation.
	if _, err := h.oauthStateRepo.Create(c.Request.Context(), state, frontendURL, "github"); err != nil {
		h.presenter.InternalError(c, "failed to create OAuth state")
		return
	}

	githubURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		url.QueryEscape(h.config.GitHubOAuthClientID),
		url.QueryEscape(callbackURL),
		url.QueryEscape("read:user user:email"),
		url.QueryEscape(state),
	)
	c.Redirect(http.StatusTemporaryRedirect, githubURL)
}

// GitHubCallback handles GET /v1/auth/github/callback.
func (h *OAuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		h.presenter.BadRequest(c, "missing authorization code from GitHub")
		return
	}

	if state == "" {
		h.presenter.BadRequest(c, "missing state parameter")
		return
	}

	redirectURL, err := h.validateOAuthState(c, state)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tokenResp, err := h.exchangeGitHubCode(ctx, c, code, redirectURL)
	if err != nil {
		return
	}

	ghUser, err := h.fetchGitHubUser(ctx, c, tokenResp.AccessToken, redirectURL)
	if err != nil {
		return
	}

	emails, _ := infraauth.FetchGitHubEmails(ctx, tokenResp.AccessToken)
	email := h.extractGitHubEmail(ghUser, emails)
	if email == "" {
		h.redirectWithOAuthError(c, redirectURL, ErrOAuthEmailRequired, "Your GitHub account must have a public or verified email address to sign up.")
		return
	}

	if !h.hasVerifiedEmail(email, emails) {
		h.redirectWithOAuthError(c, redirectURL, ErrOAuthEmailNotVerified, "Your GitHub email must be verified to sign up.")
		return
	}

	result, err := h.findOrCreateOperator(ctx, ghUser, email)
	if err != nil {
		h.redirectWithOAuthError(c, redirectURL, ErrOAuthLoginFailed, "Failed to create or access your account.")
		return
	}

	session, err := h.authService.CreateSession(ctx, result.Operator.ID)
	if err != nil {
		h.redirectWithOAuthError(c, redirectURL, ErrOAuthLoginFailed, "Failed to create session.")
		return
	}

	h.setSessionCookie(c, session.ID, redirectURL)
	h.redirectToSuccess(c, redirectURL, result.IsNew)
}

// validateOAuthState validates the OAuth state and returns redirect URL.
func (h *OAuthHandler) validateOAuthState(c *gin.Context, state string) (string, error) {
	if h.oauthStateRepo == nil {
		h.presenter.InternalError(c, "OAuth state repository not configured")
		return "", errors.New("oauth state repo not configured")
	}

	redirectURL, state, err := h.oauthStateRepo.Validate(c.Request.Context(), state)
	if err != nil {
		h.redirectWithOAuthError(c, redirectURL, ErrOAuthStateInvalid, "OAuth state expired or invalid.")
		return "", err
	}
	_ = h.oauthStateRepo.Delete(c.Request.Context(), state)
	return redirectURL, nil
}

// exchangeGitHubCode exchanges the GitHub code for an access token.
func (h *OAuthHandler) exchangeGitHubCode(ctx context.Context, c *gin.Context, code, redirectURL string) (*infraauth.GitHubTokenResponse, error) {
	callbackURL := h.config.BaseURL + "/v1/auth/github/callback"
	tokenResp, err := infraauth.ExchangeGitHubCode(ctx, code, infraauth.GitHubOAuthConfig{
		ClientID:     h.config.GitHubOAuthClientID,
		ClientSecret: h.config.GitHubOAuthClientSecret,
		RedirectURI:  callbackURL,
	})
	if err != nil {
		h.redirectWithOAuthError(c, redirectURL, ErrOAuthExchangeFailed, "Failed to exchange code with GitHub.")
		return nil, err
	}
	return tokenResp, nil
}

// fetchGitHubUser fetches the GitHub user profile.
func (h *OAuthHandler) fetchGitHubUser(ctx context.Context, c *gin.Context, accessToken, redirectURL string) (*infraauth.GitHubUserInfo, error) {
	ghUser, err := infraauth.FetchGitHubUserProfile(ctx, accessToken)
	if err != nil {
		h.redirectWithOAuthError(c, redirectURL, ErrOAuthLoginFailed, "Failed to retrieve GitHub user profile.")
		return nil, err
	}
	return ghUser, nil
}

// extractGitHubEmail extracts the email from GitHub user or emails list.
func (h *OAuthHandler) extractGitHubEmail(ghUser *infraauth.GitHubUserInfo, emails []infraauth.GitHubEmailInfo) string {
	if ghUser.Email != "" {
		return ghUser.Email
	}
	return infraauth.GetPrimaryEmail(emails)
}

// hasVerifiedEmail checks if the email is verified.
func (h *OAuthHandler) hasVerifiedEmail(email string, emails []infraauth.GitHubEmailInfo) bool {
	for _, e := range emails {
		if e.Email == email && e.Verified {
			return true
		}
	}
	return false
}

// findOrCreateOperator finds or creates the operator.
func (h *OAuthHandler) findOrCreateOperator(ctx context.Context, ghUser *infraauth.GitHubUserInfo, email string) (*appsvc.OAuthResult, error) {
	githubID := fmt.Sprintf("gh_%d", ghUser.ID)
	name := ghUser.Name
	if name == "" {
		name = ghUser.Login
	}
	return h.authService.FindOrCreateGitHubOperator(ctx, githubID, email, name)
}

// redirectWithOAuthError redirects to the error URL.
func (h *OAuthHandler) redirectWithOAuthError(c *gin.Context, redirectURL, code, message string) {
	errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
		Code:      code,
		Message:   message,
		Provider:  "github",
		HelpURL:   "https://github.com/settings/emails",
		Retryable: code == ErrOAuthLoginFailed || code == ErrOAuthExchangeFailed || code == ErrOAuthStateInvalid,
	})
	c.Redirect(http.StatusTemporaryRedirect, errURL)
}

// setSessionCookie sets the session cookie.
func (h *OAuthHandler) setSessionCookie(c *gin.Context, sessionID, redirectURL string) {
	cookie, err := h.sessionMgr.CreateCookieWithExpiry(sessionID, h.config.SessionMaxAge)
	if err != nil {
		h.redirectWithOAuthError(c, redirectURL, ErrOAuthLoginFailed, "Failed to create session.")
		return
	}
	http.SetCookie(c.Writer, cookie)
}

// redirectToSuccess redirects to the success URL.
func (h *OAuthHandler) redirectToSuccess(c *gin.Context, redirectURL string, isNew bool) {
	baseURL := h.getDefaultRedirectURL(redirectURL)
	successURL := fmt.Sprintf("%s/auth/callback?oauth=success&new=%t&provider=github", baseURL, isNew)
	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

// postJSON performs an HTTP POST with a JSON body and parses the JSON response.
func postJSON(ctx context.Context, url string, body map[string]string, resp interface{}) error {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(reqBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	httpResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", httpResp.StatusCode)
	}

	if err := json.NewDecoder(httpResp.Body).Decode(resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
