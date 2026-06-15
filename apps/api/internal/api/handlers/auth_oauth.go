package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	security "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

// GoogleLoginRedirect sends the browser to Google's OAuth consent screen.
// GET /v1/auth/google.
func (ac *AuthController) GoogleLoginRedirect(c *gin.Context) {
	if ac.config.GoogleOAuthClientID == "" || ac.config.GoogleOAuthClientSecret == "" {
		c.JSON(501, models.ErrorResponse{Error: "not_configured", Message: "Google OAuth is not configured on this server"})
		return
	}
	// Build the Google OAuth authorization URL
	// The callback will come back to /v1/auth/google/callback
	frontendURL := ac.config.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	callbackURL := ac.config.BaseURL + "/v1/auth/google/callback"
	googleURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&state=%s",
		url.QueryEscape(ac.config.GoogleOAuthClientID),
		url.QueryEscape(callbackURL),
		url.QueryEscape("openid email profile"),
		url.QueryEscape(frontendURL),
	)
	c.Redirect(http.StatusTemporaryRedirect, googleURL)
}

// GoogleCallback handles the OAuth callback from Google.
// GET /v1/auth/google/callback.
func (ac *AuthController) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state") // frontend URL to redirect back to
	if code == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_callback", Message: "missing authorization code from Google"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Exchange the code for tokens
	tokenURL := "https://oauth2.googleapis.com/token"
	tokenReq := map[string]string{
		"code":          code,
		"client_id":     ac.config.GoogleOAuthClientID,
		"client_secret": ac.config.GoogleOAuthClientSecret,
		"redirect_uri":  ac.config.BaseURL + "/v1/auth/google/callback",
		"grant_type":    "authorization_code",
	}
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := PostJSON(ctx, tokenURL, tokenReq, &tokenResp); err != nil {
		ac.log.Warn("google callback: token exchange failed", "err", err)
		c.JSON(502, models.ErrorResponse{Error: "oauth_error", Message: "failed to exchange code with Google"})
		return
	}

	// Verify the ID token using Google's public keys (cryptographically secure)
	googleClaims, err := ac.googleVer.Verify(tokenResp.IDToken)
	if err != nil {
		ac.log.Warn("google callback: ID token verification failed", "err", err)
		c.JSON(502, models.ErrorResponse{Error: "oauth_error", Message: "invalid identity token from Google"})
		return
	}

	if googleClaims.Email == "" {
		c.JSON(400, models.ErrorResponse{Error: "oauth_error", Message: "Google did not return an email address"})
		return
	}

	// Find or create the operator
	op, err := ac.store.GetOperatorByGoogleID(ctx, googleClaims.Sub)
	if err != nil {
		ac.log.Warn("google callback: db lookup failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
		return
	}

	isNew := false
	if op == nil {
		// Check if this is the first operator (bootstrap)
		count, err := ac.store.OperatorCount(ctx)
		if err != nil {
			ac.log.Warn("google callback: failed to count operators", "err", err)
		}
		role := models.RoleOperator
		if count == 0 {
			role = models.RoleSuperAdmin
			ac.log.Info("google callback: bootstrapping first operator", "email", googleClaims.Email)
		}
		op = &models.Operator{
			ID:        GenerateID(),
			Email:     googleClaims.Email,
			Name:      googleClaims.Name,
			Role:      role,
			GoogleID:  googleClaims.Sub,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := ac.store.CreateOperator(ctx, op); err != nil {
			ac.log.Warn("google callback: create operator failed", "email", googleClaims.Email, "err", err)
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
			return
		}
		isNew = true
	} else if op.GoogleID == "" {
		// Existing operator linking their Google account
		if err := ac.store.UpdateOperatorGoogleID(ctx, op.ID, googleClaims.Sub); err != nil {
			ac.log.Warn("google callback: link failed", "err", err)
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
			return
		}
	}

	// Set HttpOnly session cookie instead of exposing token in URL
	cookie, err := ac.session.CreateSessionCookieWithExpiry(op.ID, ac.config.SessionMaxAge)
	if err != nil {
		ac.log.Warn("google callback: failed to create session cookie", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
		return
	}
	http.SetCookie(c.Writer, cookie)

	// Redirect to frontend - cookie is set, no token in URL
	frontendURL := ac.config.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Use state parameter if provided (frontend URL), otherwise default to /auth/callback
	redirectTarget := frontendURL
	if state != "" {
		redirectTarget = state
	}

	// Redirect to /auth/callback with success indicator (cookie is already set)
	// This allows the callback page to show appropriate toast messages
	redirectURL := fmt.Sprintf("%s/auth/callback?oauth=success&new=%t", redirectTarget, isNew)

	ac.log.Info("google callback: login success", "email", op.Email, "role", op.Role)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GitHubLoginRedirect sends the browser to GitHub's OAuth consent screen.
// GET /v1/auth/github.
func (ac *AuthController) GitHubLoginRedirect(c *gin.Context) {
	if ac.config.GitHubOAuthClientID == "" || ac.config.GitHubOAuthClientSecret == "" {
		c.JSON(501, models.ErrorResponse{Error: "not_configured", Message: "GitHub OAuth is not configured on this server"})
		return
	}

	// Build the GitHub OAuth authorization URL
	// The callback will come back to /v1/auth/github/callback
	callbackURL := ac.config.BaseURL + "/v1/auth/github/callback"

	state := c.Query("state")
	if state == "" {
		// Generate CSRF state
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		state = hex.EncodeToString(b)
	}

	githubURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		url.QueryEscape(ac.config.GitHubOAuthClientID),
		url.QueryEscape(callbackURL),
		url.QueryEscape("read:user user:email"),
		url.QueryEscape(state),
	)
	c.Redirect(http.StatusTemporaryRedirect, githubURL)
}

// GitHubCallback handles the OAuth callback from GitHub.
// GET /v1/auth/github/callback.
func (ac *AuthController) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state") // frontend URL to redirect back to
	if code == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_callback", Message: "missing authorization code from GitHub"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	callbackURL := ac.config.BaseURL + "/v1/auth/github/callback"

	// Exchange the code for an access token
	tokenResp, err := security.ExchangeGitHubCode(ctx, code, security.GitHubOAuthConfig{
		ClientID:     ac.config.GitHubOAuthClientID,
		ClientSecret: ac.config.GitHubOAuthClientSecret,
		RedirectURI:  callbackURL,
	})
	if err != nil {
		ac.log.Warn("github callback: token exchange failed", "err", err)
		c.JSON(502, models.ErrorResponse{Error: "oauth_error", Message: "failed to exchange code with GitHub"})
		return
	}

	// Fetch GitHub user profile
	ghUser, err := security.FetchGitHubUserProfile(ctx, tokenResp.AccessToken)
	if err != nil {
		ac.log.Warn("github callback: failed to fetch user profile", "err", err)
		c.JSON(502, models.ErrorResponse{Error: "oauth_error", Message: "failed to retrieve GitHub user profile"})
		return
	}

	// Fetch user emails to get a verified email
	var email string
	if ghUser.Email == "" {
		emails, err := security.FetchGitHubEmails(ctx, tokenResp.AccessToken)
		if err != nil {
			ac.log.Warn("github callback: failed to fetch emails", "err", err)
		}
		email = security.GetPrimaryEmail(emails)
		// If still no email, use a fallback (GitHub allows login without email)
		if email == "" {
			email = ghUser.Login + "@github.noreply.vyzorix.internal"
		}
	} else {
		email = ghUser.Email
	}

	// Generate a stable GitHub ID string
	githubID := fmt.Sprintf("gh_%d", ghUser.ID)

	// Find or create the operator
	op, err := ac.store.GetOperatorByGitHubID(ctx, githubID)
	if err != nil {
		ac.log.Warn("github callback: db lookup failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
		return
	}

	isNew := false
	if op == nil {
		// Check if this is the first operator (bootstrap)
		count, err := ac.store.OperatorCount(ctx)
		if err != nil {
			ac.log.Warn("github callback: failed to count operators", "err", err)
		}
		role := models.RoleOperator
		if count == 0 {
			role = models.RoleSuperAdmin
			ac.log.Info("github callback: bootstrapping first operator", "email", email)
		}

		name := ghUser.Name
		if name == "" {
			name = ghUser.Login
		}

		op = &models.Operator{
			ID:        GenerateID(),
			Email:     email,
			Name:      name,
			Role:      role,
			GitHubID:  githubID,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := ac.store.CreateOperator(ctx, op); err != nil {
			ac.log.Warn("github callback: create operator failed", "email", email, "err", err)
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
			return
		}
		isNew = true
	} else if op.GitHubID == "" {
		// Existing operator linking their GitHub account
		if err := ac.store.UpdateOperatorGitHubID(ctx, op.ID, githubID); err != nil {
			ac.log.Warn("github callback: link failed", "err", err)
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
			return
		}
	}

	// Set HttpOnly session cookie
	cookie, err := ac.session.CreateSessionCookieWithExpiry(op.ID, ac.config.SessionMaxAge)
	if err != nil {
		ac.log.Warn("github callback: failed to create session cookie", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
		return
	}
	http.SetCookie(c.Writer, cookie)

	// Redirect to frontend
	frontendURL := ac.config.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	if state != "" {
		frontendURL = state
	}

	redirectURL := fmt.Sprintf("%s/auth/callback?oauth=success&new=%t&provider=github", frontendURL, isNew)
	ac.log.Info("github callback: login success", "email", op.Email, "role", op.Role)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}