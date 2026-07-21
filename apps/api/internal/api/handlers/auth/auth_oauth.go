package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

// OAuth error codes for frontend handling
const (
	ErrOAuthEmailRequired    = "email_required"
	ErrOAuthEmailNotVerified = "email_not_verified"
	ErrOAuthLoginFailed      = "login_failed"
	ErrOAuthStateInvalid     = "state_invalid"
	ErrOAuthExchangeFailed   = "token_exchange_failed"
	ErrOAuthConfigMissing    = "oauth_not_configured"
)

// OAuthErrorDetails holds structured error information for the frontend
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
	// Create stores a new OAuth state and returns the state ID
	Create(ctx context.Context, state, redirectURL, provider string) (string, error)
	// Validate retrieves and validates an OAuth state, returning the redirect URL and state
	Validate(ctx context.Context, state string) (redirectURL string, stateID string, err error)
	// Delete removes an OAuth state after use
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

// getOAuthRedirectURL builds a redirect URL with error parameters for OAuth errors
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

// getDefaultRedirectURL returns the default redirect URL
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

	// OAuth state repository is required for CSRF protection
	if h.oauthStateRepo == nil {
		h.presenter.InternalError(c, "OAuth state repository is required for CSRF protection")
		return
	}

	frontendURL := h.config.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Generate random state for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		h.presenter.InternalError(c, "failed to generate OAuth state")
		return
	}
	state := hex.EncodeToString(stateBytes)

	// Persist OAuth state to database for CSRF validation
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

	// 8: Validate state from database if repository is configured
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
	// Delete the state after successful validation (one-time use)
	_ = h.oauthStateRepo.Delete(c.Request.Context(), state)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Exchange code for tokens
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

	// Verify ID token
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

	// Check for email requirement - Google requires verified email
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

	// Find or create operator via application service
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

	// Create session (validates operator was found/created)
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

	// Set session cookie with session ID
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

	// Redirect to frontend with success
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

	// OAuth state repository is required for CSRF protection
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

	// Persist OAuth state to database for CSRF validation
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

	// 8: Validate state from database - REQUIRED for CSRF protection
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
			Provider:  "github",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}
	// Delete the state after successful validation (one-time use)
	_ = h.oauthStateRepo.Delete(c.Request.Context(), state)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	callbackURL := h.config.BaseURL + "/v1/auth/github/callback"

	// Exchange code for access token
	tokenResp, err := infraauth.ExchangeGitHubCode(ctx, code, infraauth.GitHubOAuthConfig{
		ClientID:     h.config.GitHubOAuthClientID,
		ClientSecret: h.config.GitHubOAuthClientSecret,
		RedirectURI:  callbackURL,
	})
	if err != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthExchangeFailed,
			Message:   "Failed to exchange code with GitHub. Please try again.",
			Provider:  "github",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Fetch GitHub user profile
	ghUser, err := infraauth.FetchGitHubUserProfile(ctx, tokenResp.AccessToken)
	if err != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthLoginFailed,
			Message:   "Failed to retrieve GitHub user profile. Please try again.",
			Provider:  "github",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Fetch user emails - REQUIRED for sign up
	emails, fetchErr := infraauth.FetchGitHubEmails(ctx, tokenResp.AccessToken)
	if fetchErr != nil {
		h.logger.Warn("Failed to fetch GitHub emails", "err", fetchErr)
	}

	email := ghUser.Email
	if email == "" {
		email = infraauth.GetPrimaryEmail(emails)
	}

	// If no email found, redirect with error - email is REQUIRED
	if email == "" {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthEmailRequired,
			Message:   "Your GitHub account must have a public or verified email address to sign up. Please go to GitHub Settings > Emails and make an email public or verified, then try again.",
			Provider:  "github",
			HelpURL:   "https://github.com/settings/emails",
			Retryable: false,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Check if email is verified
	hasVerifiedEmail := false
	for _, e := range emails {
		if e.Email == email && e.Verified {
			hasVerifiedEmail = true
			break
		}
	}

	// Reject unverified emails to prevent account takeover
	if !hasVerifiedEmail {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthEmailNotVerified,
			Message:   "Your GitHub email must be verified to sign up. Please verify your email in GitHub Settings, then try again.",
			Provider:  "github",
			HelpURL:   "https://github.com/settings/emails",
			Retryable: false,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Generate stable GitHub ID
	githubID := fmt.Sprintf("gh_%d", ghUser.ID)

	// Get name
	name := ghUser.Name
	if name == "" {
		name = ghUser.Login
	}

	// Find or create operator
	result, err := h.authService.FindOrCreateGitHubOperator(ctx, githubID, email, name)
	if err != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthLoginFailed,
			Message:   "Failed to create or access your account. Please try again or contact support.",
			Provider:  "github",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Create session (validates operator was found/created)
	session, err := h.authService.CreateSession(ctx, result.Operator.ID)
	if err != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthLoginFailed,
			Message:   "Failed to create session. Please try again.",
			Provider:  "github",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	// Set session cookie with session ID
	cookie, err := h.sessionMgr.CreateCookieWithExpiry(session.ID, h.config.SessionMaxAge)
	if err != nil {
		errURL := h.getOAuthRedirectURL(redirectURL, OAuthErrorDetails{
			Code:      ErrOAuthLoginFailed,
			Message:   "Failed to create session. Please try again.",
			Provider:  "github",
			Retryable: true,
		})
		c.Redirect(http.StatusTemporaryRedirect, errURL)
		return
	}

	http.SetCookie(c.Writer, cookie)

	// Redirect to frontend with success
	baseURL := h.getDefaultRedirectURL(redirectURL)
	successURL := fmt.Sprintf("%s/auth/callback?oauth=success&new=%t&provider=github", baseURL, result.IsNew)
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
