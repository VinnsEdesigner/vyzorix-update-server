package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix-update-server/config"
	"github.com/VinnsEdesigner/vyzorix-update-server/models"
	"github.com/VinnsEdesigner/vyzorix-update-server/security"
	"github.com/VinnsEdesigner/vyzorix-update-server/storage"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthController handles operator authentication: login, register, logout, me, Google OAuth.
type AuthController struct {
	log       *slog.Logger
	config   config.Config
	store    *storage.Store
	jwt      *security.JWTManager
}

// NewAuthController creates a new auth controller.
func NewAuthController(log *slog.Logger, cfg config.Config, store *storage.Store) *AuthController {
	jwtManager := security.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration, "vyzorix-update-server")
	return &AuthController{
		log:     log,
		config:  cfg,
		store:   store,
		jwt:     jwtManager,
	}
}

// Login authenticates an operator with email and password, returning a JWT.
func (ac *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "email and password are required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	op, err := ac.store.GetOperatorByEmail(ctx, req.Email)
	if err != nil {
		ac.log.Warn("login: db error", "email", req.Email, "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
		return
	}
	if op == nil {
		c.JSON(401, models.ErrorResponse{Error: "invalid_credentials", Message: "invalid email or password"})
		return
	}
	if op.PasswordHash == "" {
		c.JSON(401, models.ErrorResponse{Error: "invalid_credentials", Message: "this account uses Google sign-in; please use that method"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(op.PasswordHash), []byte(req.Password)); err != nil {
		ac.log.Warn("login: bad password", "email", req.Email)
		c.JSON(401, models.ErrorResponse{Error: "invalid_credentials", Message: "invalid email or password"})
		return
	}

	ac.issueToken(c, ctx, op)
}

// Register creates the first operator in the system.
// Subsequent registrations require a super_admin token.
func (ac *AuthController) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" || req.Name == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "email, password, and name are required"})
		return
	}
	if len(req.Password) < 8 {
		c.JSON(400, models.ErrorResponse{Error: "bad_password", Message: "password must be at least 8 characters"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	count, err := ac.store.OperatorCount(ctx)
	if err != nil {
		ac.log.Warn("register: count failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "registration failed"})
		return
	}

	// Determine role: first operator gets super_admin, all others need a valid admin JWT
	role := models.RoleOperator
	if count == 0 {
		role = models.RoleSuperAdmin
		ac.log.Info("register: bootstrapping first operator", "email", req.Email)
	} else {
		// Subsequent registrations require a super_admin JWT
		authOp := getOperatorFromContext(c)
		if authOp == nil || authOp.Role != string(models.RoleSuperAdmin) {
			c.JSON(403, models.ErrorResponse{Error: "forbidden", Message: "only a super_admin can invite new operators"})
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ac.log.Warn("register: bcrypt failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "registration failed"})
		return
	}

	op := &models.Operator{
		ID:           generateID(),
		Email:        req.Email,
		Name:         strings.TrimSpace(req.Name),
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := ac.store.CreateOperator(ctx, op); err != nil {
		if isUniqueViolation(err) {
			c.JSON(409, models.ErrorResponse{Error: "email_conflict", Message: "an account with this email already exists"})
			return
		}
		ac.log.Warn("register: create failed", "email", req.Email, "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "registration failed"})
		return
	}

	ac.log.Info("register: operator created", "email", req.Email, "role", role)
	ac.issueToken(c, ctx, op)
}

// Me returns the operator profile for the authenticated caller.
func (ac *AuthController) Me(c *gin.Context) {
	op := getOperatorFromContext(c)
	if op == nil {
		c.JSON(401, models.ErrorResponse{Error: "unauthorized", Message: "authentication required"})
		return
	}
	c.JSON(200, op.ToResponse())
}

// UpdateName updates the display name for the authenticated operator.
// Email and role are server-controlled and cannot be changed via this endpoint.
func (ac *AuthController) UpdateName(c *gin.Context) {
	op := getOperatorFromContext(c)
	if op == nil {
		c.JSON(401, models.ErrorResponse{Error: "unauthorized", Message: "authentication required"})
		return
	}

	var req models.UpdateNameRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "name cannot be empty"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := ac.store.UpdateOperatorName(ctx, op.ID, name); err != nil {
		ac.log.Warn("updateName: failed", "operatorID", op.ID, "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "update failed"})
		return
	}

	// Re-fetch to return the updated operator
	updated, err := ac.store.GetOperatorByEmail(ctx, op.Email)
	if err != nil || updated == nil {
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "update succeeded but fetch failed"})
		return
	}

	ac.log.Info("updateName: success", "operatorID", op.ID, "name", name)
	c.JSON(200, updated.ToResponse())
}

// Logout revokes the current session.
func (ac *AuthController) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(200, map[string]string{"ok": true})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	hash := security.HashToken(token)
	_ = ac.store.DeleteSession(ctx, hash)
	c.JSON(200, map[string]string{"ok": true})
}

// GoogleLoginRedirect sends the browser to Google's OAuth consent screen.
// GET /v1/auth/google
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
// GET /v1/auth/google/callback
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
		IDToken     string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := postJSON(ctx, tokenURL, tokenReq, &tokenResp); err != nil {
		ac.log.Warn("google callback: token exchange failed", "err", err)
		c.JSON(502, models.ErrorResponse{Error: "oauth_error", Message: "failed to exchange code with Google"})
		return
	}

	// Decode the ID token to get user info
	// In production, verify the ID token signature with Google's public keys
	var idToken struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := decodeJWTPayload(tokenResp.IDToken, &idToken); err != nil {
		ac.log.Warn("google callback: id_token decode failed", "err", err)
		c.JSON(502, models.ErrorResponse{Error: "oauth_error", Message: "failed to decode identity token from Google"})
		return
	}

	if idToken.Email == "" {
		c.JSON(400, models.ErrorResponse{Error: "oauth_error", Message: "Google did not return an email address"})
		return
	}

	// Find or create the operator
	op, err := ac.store.GetOperatorByGoogleID(ctx, idToken.Sub)
	if err != nil {
		ac.log.Warn("google callback: db lookup failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
		return
	}

	isNew := false
	if op == nil {
		// Check if this is the first operator (bootstrap)
		count, _ := ac.store.OperatorCount(ctx)
		role := models.RoleOperator
		if count == 0 {
			role = models.RoleSuperAdmin
			ac.log.Info("google callback: bootstrapping first operator", "email", idToken.Email)
		}
		op = &models.Operator{
			ID:        generateID(),
			Email:     idToken.Email,
			Name:      idToken.Name,
			Role:      role,
			GoogleID:  idToken.Sub,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := ac.store.CreateOperator(ctx, op); err != nil {
			ac.log.Warn("google callback: create operator failed", "email", idToken.Email, "err", err)
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
			return
		}
		isNew = true
	} else if op.GoogleID == "" {
		// Existing operator linking their Google account
		if err := ac.store.UpdateOperatorGoogleID(ctx, op.ID, idToken.Sub); err != nil {
			ac.log.Warn("google callback: link failed", "err", err)
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
			return
		}
	}

	// Issue JWT
	token, expiresAt, err := ac.jwt.Generate(op.ID, op.Email, op.Name, string(op.Role))
	if err != nil {
		ac.log.Warn("google callback: jwt failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
		return
	}

	// Redirect to frontend with token
	frontendURL := ac.config.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	// Use state as redirect target if provided, otherwise default dashboard
	redirectTarget := "/"
	if state != "" {
		redirectTarget = state
	}
	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s&isNew=%t", frontendURL, url.QueryEscape(token), isNew)

	ac.log.Info("google callback: login success", "email", op.Email, "role", op.Role)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// issueToken creates a session in the DB and responds with the JWT.
func (ac *AuthController) issueToken(c *gin.Context, ctx context.Context, op *models.Operator) {
	token, expiresAt, err := ac.jwt.Generate(op.ID, op.Email, op.Name, string(op.Role))
	if err != nil {
		ac.log.Warn("issueToken: jwt failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
		return
	}

	sess := &models.Session{
		ID:         generateID(),
		OperatorID: op.ID,
		TokenHash:  security.HashToken(token),
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now().UTC(),
		UserAgent:  c.GetHeader("User-Agent"),
		IPAddress:  c.ClientIP(),
	}
	if err := ac.store.CreateSession(ctx, sess); err != nil {
		ac.log.Warn("issueToken: session create failed", "err", err)
		// Non-fatal: token is still valid, session just won't be revokable
	}

	c.JSON(200, models.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt.UnixMilli(),
		Operator:  op.ToResponse(),
	})
}

// ─── Helpers ───────────────────────────────────────────────────────────────────────

func getOperatorFromContext(c *gin.Context) *models.Operator {
	v, exists := c.Get("operator")
	if !exists {
		return nil
	}
	op, ok := v.(*models.Operator)
	if !ok {
		return nil
	}
	return op
}

func generateID() string {
	b := make([]byte, 16)
	_ = rand.Read(b)
	return hex.EncodeToString(b)
}

func isUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "UNIQUE constraint"))
}

// postJSON performs an HTTP POST with a JSON body.
func postJSON(ctx context.Context, url string, body any, resp any) error {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(reqBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	rb, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(rb))
	}
	return json.Unmarshal(rb, resp)
}

// decodeJWTPayload extracts the payload from a JWT without signature verification.
// In production, fetch Google's public keys and verify the signature.
// For Phase 1.5, we trust the token format and extract the claims directly.
func decodeJWTPayload(token string, out any) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format")
	}
	payload, err := base64RawURLDecode(parts[1])
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, out)
}

func base64RawURLDecode(s string) ([]byte, error) {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	// Replace URL-safe chars
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	return []byte(s), nil
}
