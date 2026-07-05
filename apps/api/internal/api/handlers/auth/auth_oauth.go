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

// GoogleLogin handles GET /v1/auth/google.
func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	if h.config.GoogleOAuthClientID == "" || h.config.GoogleOAuthClientSecret == "" {
		h.presenter.NotImplemented(c, "Google OAuth is not configured on this server")
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

	// 8: Persist OAuth state to database if repository is configured
	if h.oauthStateRepo != nil {
		if _, err := h.oauthStateRepo.Create(c.Request.Context(), state, frontendURL, "google"); err != nil {
			h.presenter.InternalError(c, "failed to create OAuth state")
			return
		}
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
		h.presenter.BadRequest(c, "invalid or expired OAuth state")
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
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if postErr := postJSON(ctx, tokenURL, tokenReq, &tokenResp); postErr != nil {
			h.presenter.BadGateway(c, "failed to exchange code with Google")
			return
		}

	// Verify ID token
	googleClaims, err := h.googleVer.Verify(tokenResp.IDToken)
	if err != nil {
		h.presenter.BadGateway(c, "invalid identity token from Google")
		return
	}

	if googleClaims.Email == "" {
		h.presenter.BadRequest(c, "Google did not return an email address")
		return
	}

	// Find or create operator via application service
	result, err := h.authService.FindOrCreateGoogleOperator(ctx, googleClaims.Sub, googleClaims.Email, googleClaims.Name)
	if err != nil {
		h.presenter.InternalError(c, "login failed")
		return
	}

	// Create session (validates operator was found/created)
	_, err = h.authService.CreateSession(ctx, result.Operator.ID)
	if err != nil {
		h.presenter.InternalError(c, "login failed")
		return
	}

	// Set session cookie
	cookie, err := h.sessionMgr.CreateCookieWithExpiry(result.Operator.ID, h.config.SessionMaxAge)
	if err != nil {
		h.presenter.InternalError(c, "login failed")
		return
	}

	http.SetCookie(c.Writer, cookie)

	// Redirect to frontend with stored redirect URL
	if redirectURL == "" {
		redirectURL = h.config.FrontendURL
		if redirectURL == "" {
			redirectURL = "http://localhost:5173"
		}
	}

	redirectURL = fmt.Sprintf("%s/auth/callback?oauth=success&new=%t", redirectURL, result.IsNew)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GitHubLogin handles GET /v1/auth/github.
func (h *OAuthHandler) GitHubLogin(c *gin.Context) {
	if h.config.GitHubOAuthClientID == "" || h.config.GitHubOAuthClientSecret == "" {
		h.presenter.NotImplemented(c, "GitHub OAuth is not configured on this server")
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

	// 8: Persist OAuth state to database if repository is configured
	if h.oauthStateRepo != nil {
		if _, err := h.oauthStateRepo.Create(c.Request.Context(), state, frontendURL, "github"); err != nil {
			h.presenter.InternalError(c, "failed to create OAuth state")
			return
		}
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
		h.presenter.BadRequest(c, "invalid or expired OAuth state")
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
		h.presenter.BadGateway(c, "failed to exchange code with GitHub")
		return
	}

	// Fetch GitHub user profile
	ghUser, err := infraauth.FetchGitHubUserProfile(ctx, tokenResp.AccessToken)
	if err != nil {
		h.presenter.BadGateway(c, "failed to retrieve GitHub user profile")
		return
	}

	// Fetch user emails
	email := ghUser.Email
	if email == "" {
		email = fetchGitHubEmailWithFallback(ctx, tokenResp.AccessToken, ghUser.Login, h.logger)
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
		h.presenter.InternalError(c, "login failed")
		return
	}

	// Create session (validates operator was found/created)
	_, err = h.authService.CreateSession(ctx, result.Operator.ID)
	if err != nil {
		h.presenter.InternalError(c, "login failed")
		return
	}

	// Set session cookie
	cookie, err := h.sessionMgr.CreateCookieWithExpiry(result.Operator.ID, h.config.SessionMaxAge)
	if err != nil {
		h.presenter.InternalError(c, "login failed")
		return
	}

	http.SetCookie(c.Writer, cookie)

	// Redirect to frontend with stored redirect URL if available
	if redirectURL == "" {
		redirectURL = h.config.FrontendURL
		if redirectURL == "" {
			redirectURL = "http://localhost:5173"
		}
	}

	redirectURL = fmt.Sprintf("%s/auth/callback?oauth=success&new=%t&provider=github", redirectURL, result.IsNew)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
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

// fetchGitHubEmailWithFallback fetches the primary email or generates a fallback.
func fetchGitHubEmailWithFallback(ctx context.Context, accessToken, login string, logger *slog.Logger) string {
	emails, fetchErr := infraauth.FetchGitHubEmails(ctx, accessToken)
	email := infraauth.GetPrimaryEmail(emails)
	if email == "" {
		email = login + "@github.noreply.vyzorix.internal"
		if fetchErr != nil {
			logger.Warn("Failed to fetch GitHub emails, using fallback", "err", fetchErr)
		}
		return email
	}
	if fetchErr != nil {
		logger.Warn("Failed to fetch GitHub emails, using fallback", "err", fetchErr)
	}
	return email
}
