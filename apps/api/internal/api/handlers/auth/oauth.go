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

	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"

	"github.com/gin-gonic/gin"
)

// OAuthHandler handles OAuth endpoints.
type OAuthHandler struct {
	authService *appsvc.AuthService
	sessionMgr  *infraauth.SessionManager
	config      config.Config
	googleVer   *infraauth.GoogleTokenVerifier
	logger      *slog.Logger
}

// NewOAuthHandler creates a new OAuthHandler.
func NewOAuthHandler(authService *appsvc.AuthService, sessionMgr *infraauth.SessionManager, cfg config.Config, googleVer *infraauth.GoogleTokenVerifier) *OAuthHandler {
	return &OAuthHandler{
		authService: authService,
		sessionMgr:  sessionMgr,
		config:      cfg,
		googleVer:  googleVer,
	}
}

// WithLogger sets the logger for the handler.
func (h *OAuthHandler) WithLogger(logger *slog.Logger) *OAuthHandler {
	h.logger = logger
	return h
}

// GoogleLogin handles GET /v1/auth/google.
func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	if h.config.GoogleOAuthClientID == "" || h.config.GoogleOAuthClientSecret == "" {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not_configured", "message": "Google OAuth is not configured on this server"})
		return
	}

	frontendURL := h.config.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	callbackURL := h.config.BaseURL + "/v1/auth/google/callback"
	googleURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&state=%s",
		url.QueryEscape(h.config.GoogleOAuthClientID),
		url.QueryEscape(callbackURL),
		url.QueryEscape("openid email profile"),
		url.QueryEscape(frontendURL),
	)
	c.Redirect(http.StatusTemporaryRedirect, googleURL)
}

// GoogleCallback handles GET /v1/auth/google/callback.
func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_callback", "message": "missing authorization code from Google"})
		return
	}

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
	if err := postJSON(ctx, tokenURL, tokenReq, &tokenResp); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "oauth_error", "message": "failed to exchange code with Google"})
		return
	}

	// Verify ID token
	googleClaims, err := h.googleVer.Verify(tokenResp.IDToken)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "oauth_error", "message": "invalid identity token from Google"})
		return
	}

	if googleClaims.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "oauth_error", "message": "Google did not return an email address"})
		return
	}

	// Find or create operator via application service
	result, err := h.authService.FindOrCreateGoogleOperator(ctx, googleClaims.Sub, googleClaims.Email, googleClaims.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "login failed"})
		return
	}

	// Create session (validates operator was found/created)
	_, err = h.authService.CreateSession(ctx, result.Operator.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "login failed"})
		return
	}

	// Set session cookie
	cookie, err := h.sessionMgr.CreateSessionCookieWithExpiry(result.Operator.ID, h.config.SessionMaxAge)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "login failed"})
		return
	}
	http.SetCookie(c.Writer, cookie)

	// Redirect to frontend
	frontendURL := h.config.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	if state != "" {
		frontendURL = state
	}
	redirectURL := fmt.Sprintf("%s/auth/callback?oauth=success&new=%t", frontendURL, result.IsNew)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GitHubLogin handles GET /v1/auth/github.
func (h *OAuthHandler) GitHubLogin(c *gin.Context) {
	if h.config.GitHubOAuthClientID == "" || h.config.GitHubOAuthClientSecret == "" {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not_configured", "message": "GitHub OAuth is not configured on this server"})
		return
	}

	callbackURL := h.config.BaseURL + "/v1/auth/github/callback"

	state := c.Query("state")
	if state == "" {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		state = hex.EncodeToString(b)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_callback", "message": "missing authorization code from GitHub"})
		return
	}

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
		c.JSON(http.StatusBadGateway, gin.H{"error": "oauth_error", "message": "failed to exchange code with GitHub"})
		return
	}

	// Fetch GitHub user profile
	ghUser, err := infraauth.FetchGitHubUserProfile(ctx, tokenResp.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "oauth_error", "message": "failed to retrieve GitHub user profile"})
		return
	}

	// Fetch user emails
	var email string
	if ghUser.Email == "" {
		emails, err := infraauth.FetchGitHubEmails(ctx, tokenResp.AccessToken)
		if err != nil {
			if h.logger != nil {
				h.logger.Warn("Failed to fetch GitHub emails, using fallback", "err", err)
			}
		}
		email = infraauth.GetPrimaryEmail(emails)
		if email == "" {
			email = ghUser.Login + "@github.noreply.vyzorix.internal"
		}
	} else {
		email = ghUser.Email
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "login failed"})
		return
	}

	// Create session (validates operator was found/created)
	_, err = h.authService.CreateSession(ctx, result.Operator.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "login failed"})
		return
	}

	// Set session cookie
	cookie, err := h.sessionMgr.CreateSessionCookieWithExpiry(result.Operator.ID, h.config.SessionMaxAge)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "login failed"})
		return
	}
	http.SetCookie(c.Writer, cookie)

	// Redirect to frontend
	frontendURL := h.config.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	if state != "" {
		frontendURL = state
	}
	redirectURL := fmt.Sprintf("%s/auth/callback?oauth=success&new=%t&provider=github", frontendURL, result.IsNew)
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
