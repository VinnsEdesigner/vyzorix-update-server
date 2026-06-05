package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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
	"github.com/VinnsEdesigner/vyzorix-update-server/services"
	"github.com/VinnsEdesigner/vyzorix-update-server/storage"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthController handles operator authentication: login, register, logout, me, Google OAuth.
type AuthController struct {
	log        *slog.Logger
	config     config.Config
	store      *storage.Store
	jwt        *security.JWTManager
	googleVer  *security.GoogleTokenVerifier
	emailSvc   *services.EmailService
}

// NewAuthController creates a new auth controller.
func NewAuthController(log *slog.Logger, cfg config.Config, store *storage.Store) *AuthController {
	jwtManager := security.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration, "vyzorix-update-server")
	googleVer := security.NewGoogleTokenVerifier(cfg.GoogleOAuthClientID)
	emailSvc := services.NewEmailService()
	return &AuthController{
		log:       log,
		config:    cfg,
		store:     store,
		jwt:       jwtManager,
		googleVer: googleVer,
		emailSvc:  emailSvc,
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
	var req models.OperatorRegisterRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" || req.Name == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "email, password, and name are required"})
		return
	}

	// Validate password complexity
	if err := security.ValidatePassword(req.Password, security.DefaultPasswordPolicy); err != nil {
		ac.log.Warn("register: weak password", "email", req.Email)
		c.JSON(400, models.ErrorResponse{Error: "bad_password", Message: err.Error()})
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
		if authOp == nil || authOp.Role != models.RoleSuperAdmin {
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
		EmailVerified: false, // Not verified until email verification completes
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

	// Send verification email
	ac.sendVerificationEmail(ctx, op)

	ac.log.Info("register: operator created", "email", req.Email, "role", role)
	c.JSON(201, models.MessageResponse{
		Message: "Registration successful. Please check your email to verify your account.",
	})
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
		c.JSON(200, map[string]any{"ok": true})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	hash := security.HashToken(token)
	_ = ac.store.DeleteSession(ctx, hash)
	c.JSON(200, map[string]any{"ok": true})
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
	_ = c.Query("state") // frontend URL to redirect back to (reserved for future use)
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
		count, _ := ac.store.OperatorCount(ctx)
		role := models.RoleOperator
		if count == 0 {
			role = models.RoleSuperAdmin
			ac.log.Info("google callback: bootstrapping first operator", "email", googleClaims.Email)
		}
		op = &models.Operator{
			ID:        generateID(),
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

	// Issue JWT
	token, _, err := ac.jwt.Generate(op.ID, op.Email, op.Name, string(op.Role))
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

// sendVerificationEmail creates a verification token and sends the verification email.
func (ac *AuthController) sendVerificationEmail(ctx context.Context, op *models.Operator) {
	if !ac.emailSvc.IsConfigured() {
		ac.log.Warn("sendVerificationEmail: email service not configured, skipping", "email", op.Email)
		return
	}

	// Generate verification token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		ac.log.Warn("sendVerificationEmail: failed to generate token", "err", err)
		return
	}
	token := hex.EncodeToString(tokenBytes)
	tokenHash := security.HashToken(token) // Store hash in DB

	// Create verification record
	ev := &storage.EmailVerification{
		ID:         generateID(),
		OperatorID: op.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().UTC().Add(ac.config.EmailVerifyTokenExpiry),
		CreatedAt:  time.Now().UTC(),
	}

	if err := ac.store.CreateEmailVerification(ctx, ev); err != nil {
		ac.log.Warn("sendVerificationEmail: failed to store token", "err", err)
		return
	}

	// Send email (async, don't block registration)
	go func() {
		if err := ac.emailSvc.SendVerificationEmail(context.Background(), op.Email, op.Name, token); err != nil {
			ac.log.Error("sendVerificationEmail: failed to send email", "email", op.Email, "err", err)
		} else {
			ac.log.Info("sendVerificationEmail: sent", "email", op.Email)
		}
	}()
}

// VerifyEmail handles email verification requests.
// POST /v1/auth/verify-email
func (ac *AuthController) VerifyEmail(c *gin.Context) {
	var req models.VerifyEmailRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "token is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Find verification token
	tokenHash := security.HashToken(token)
	ev, err := ac.store.GetEmailVerificationByTokenHash(ctx, tokenHash)
	if err != nil {
		ac.log.Warn("verifyEmail: db error", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "verification failed"})
		return
	}

	if ev == nil {
		c.JSON(400, models.ErrorResponse{Error: "invalid_token", Message: "invalid or expired verification token"})
		return
	}

	// Check if expired
	if time.Now().UTC().After(ev.ExpiresAt) {
		c.JSON(400, models.ErrorResponse{Error: "token_expired", Message: "verification token has expired"})
		return
	}

	// Mark email as verified
	if err := ac.store.SetOperatorEmailVerified(ctx, ev.OperatorID, true); err != nil {
		ac.log.Warn("verifyEmail: failed to set verified", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "verification failed"})
		return
	}

	// Delete the verification token (single use)
	_ = ac.store.DeleteEmailVerification(ctx, ev.ID)

	ac.log.Info("verifyEmail: success", "operatorID", ev.OperatorID)
	c.JSON(200, models.EmailVerifiedResponse{Verified: true})
}

// ResendVerification resends the verification email.
// POST /v1/auth/resend-verification
func (ac *AuthController) ResendVerification(c *gin.Context) {
	var req models.ResendVerificationRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "email is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Find operator by email
	op, err := ac.store.GetOperatorByEmail(ctx, email)
	if err != nil {
		ac.log.Warn("resendVerification: db error", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "request failed"})
		return
	}

	// Always return success for security (don't reveal if email exists)
	if op == nil {
		ac.log.Info("resendVerification: email not found (silently)", "email", email)
		c.JSON(200, models.MessageResponse{Message: "If that email exists, a verification email has been sent."})
		return
	}

	// Check if already verified
	if op.EmailVerified {
		c.JSON(400, models.ErrorResponse{Error: "already_verified", Message: "this email is already verified"})
		return
	}

	// Delete old verification tokens
	_ = ac.store.DeleteEmailVerificationsByOperator(ctx, op.ID)

	// Send new verification email
	ac.sendVerificationEmail(ctx, op)

	ac.log.Info("resendVerification: sent", "email", email)
	c.JSON(200, models.MessageResponse{Message: "If that email exists, a verification email has been sent."})
}

// ForgotPassword handles password reset requests.
// POST /v1/auth/forgot-password
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "email is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Find operator by email
	op, err := ac.store.GetOperatorByEmail(ctx, email)
	if err != nil {
		ac.log.Warn("forgotPassword: db error", "err", err)
		// Still return success for security
		c.JSON(200, models.MessageResponse{Message: "If that email exists, a password reset link has been sent."})
		return
	}

	// Always return success for security (don't reveal if email exists)
	if op == nil {
		ac.log.Info("forgotPassword: email not found (silently)", "email", email)
		c.JSON(200, models.MessageResponse{Message: "If that email exists, a password reset link has been sent."})
		return
	}

	// Check if password-based account (Google-only accounts can't reset via email)
	if op.PasswordHash == "" {
		c.JSON(400, models.ErrorResponse{Error: "google_account", Message: "this account uses Google sign-in and cannot reset password via email"})
		return
	}

	// Delete old reset tokens
	_ = ac.store.DeletePasswordResetTokensByOperator(ctx, op.ID)

	// Generate reset token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		ac.log.Warn("forgotPassword: failed to generate token", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "request failed"})
		return
	}
	token := hex.EncodeToString(tokenBytes)
	tokenHash := security.HashToken(token)

	prt := &storage.PasswordResetToken{
		ID:         generateID(),
		OperatorID: op.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().UTC().Add(ac.config.PasswordResetTokenExpiry),
		CreatedAt:  time.Now().UTC(),
	}

	if err := ac.store.CreatePasswordResetToken(ctx, prt); err != nil {
		ac.log.Warn("forgotPassword: failed to store token", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "request failed"})
		return
	}

	// Send email (async)
	go func() {
		if err := ac.emailSvc.SendPasswordResetEmail(context.Background(), op.Email, op.Name, token); err != nil {
			ac.log.Error("forgotPassword: failed to send email", "email", op.Email, "err", err)
		} else {
			ac.log.Info("forgotPassword: sent", "email", op.Email)
		}
	}()

	c.JSON(200, models.MessageResponse{Message: "If that email exists, a password reset link has been sent."})
}

// ResetPassword handles password reset with a valid token.
// POST /v1/auth/reset-password
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}

	token := strings.TrimSpace(req.Token)
	newPassword := req.NewPassword

	if token == "" || newPassword == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "token and newPassword are required"})
		return
	}

	// Validate password complexity
	if err := security.ValidatePassword(newPassword, security.DefaultPasswordPolicy); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_password", Message: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Find reset token
	tokenHash := security.HashToken(token)
	prt, err := ac.store.GetPasswordResetTokenByHash(ctx, tokenHash)
	if err != nil {
		ac.log.Warn("resetPassword: db error", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "reset failed"})
		return
	}

	if prt == nil {
		c.JSON(400, models.ErrorResponse{Error: "invalid_token", Message: "invalid or expired reset token"})
		return
	}

	// Check if expired
	if time.Now().UTC().After(prt.ExpiresAt) {
		c.JSON(400, models.ErrorResponse{Error: "token_expired", Message: "reset token has expired"})
		return
	}

	// Check if already used
	if prt.UsedAt != nil {
		c.JSON(400, models.ErrorResponse{Error: "token_used", Message: "this reset token has already been used"})
		return
	}

	// Get operator and update password
	op, err := ac.store.GetOperatorByID(ctx, prt.OperatorID)
	if err != nil || op == nil {
		ac.log.Warn("resetPassword: operator not found", "operatorID", prt.OperatorID)
		c.JSON(400, models.ErrorResponse{Error: "invalid_token", Message: "invalid reset token"})
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		ac.log.Warn("resetPassword: bcrypt failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "reset failed"})
		return
	}

	// Update operator password (need to add this method)
	// For now, we'll use a direct approach via UpdateOperatorPassword
	if err := ac.store.UpdateOperatorPassword(ctx, op.ID, string(newHash)); err != nil {
		ac.log.Warn("resetPassword: update failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "reset failed"})
		return
	}

	// Mark token as used
	_ = ac.store.MarkPasswordResetTokenUsed(ctx, prt.ID)

	// Delete all sessions (force logout)
	_ = ac.store.DeleteAllSessionsForOperator(ctx, op.ID)

	// Send confirmation email
	go func() {
		if err := ac.emailSvc.SendPasswordChangedEmail(context.Background(), op.Email, op.Name); err != nil {
			ac.log.Error("resetPassword: failed to send confirmation email", "email", op.Email, "err", err)
		}
	}()

	ac.log.Info("resetPassword: success", "operatorID", op.ID)
	c.JSON(200, models.MessageResponse{Message: "Password reset successful. Please log in with your new password."})
}

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
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("gen-%d", time.Now().UnixNano())
	}
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
