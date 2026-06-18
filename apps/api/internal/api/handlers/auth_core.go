// Package controllers provides HTTP handlers.
package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	services "github.com/VinnsEdesigner/vyzorix/apps/api/internal"
	security "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
)

// AuthController handles operator authentication: login, register, logout, me, Google OAuth.
type AuthController struct {
	log       *slog.Logger
	store     *storage.Store
	session   *security.SessionManager
	googleVer *security.GoogleTokenVerifier
	emailSvc  *services.EmailService
	config    config.Config
}

// NewAuthController creates a new auth controller.
func NewAuthController(log *slog.Logger, cfg config.Config, store *storage.Store) *AuthController {
	googleVer := security.NewGoogleTokenVerifier(cfg.GoogleOAuthClientID)
	emailSvc := services.NewEmailService()

	// Initialize SessionManager - uses SESSION_SECRET, falls back to JWTSecret.
	sessionSecret := cfg.SessionSecret
	if sessionSecret == "" {
		sessionSecret = cfg.JWTSecret
	}
	sessionManager := security.NewSessionManager(sessionSecret)

	return &AuthController{
		log:       log,
		config:    cfg,
		store:     store,
		session:   sessionManager,
		googleVer: googleVer,
		emailSvc:  emailSvc,
	}
}

// Login authenticates an operator with email and password, setting an HttpOnly session cookie.
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

	if err := storage.VerifyPassword(req.Password, op.PasswordHash); err != nil {
		ac.log.Warn("login: bad password", "email", req.Email)
		c.JSON(401, models.ErrorResponse{Error: "invalid_credentials", Message: "invalid email or password"})
		return
	}

	// Set HttpOnly session cookie.
	cookie, err := ac.session.CreateSessionCookieWithExpiry(op.ID, ac.config.SessionMaxAge)
	if err != nil {
		ac.log.Warn("login: failed to create session cookie", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "login failed"})
		return
	}
	http.SetCookie(c.Writer, cookie)

	// Return operator (no JWT token in body - using cookie instead).
	c.JSON(200, op.ToResponse())
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

	// Validate password complexity.
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

	// Determine role: first operator gets super_admin, all others require super_admin auth.
	role := models.RoleOperator
	if count == 0 {
		role = models.RoleSuperAdmin
		ac.log.Info("register: bootstrapping first operator", "email", req.Email)
	} else {
		// Subsequent registrations require a super_admin session cookie.
		authOp := GetOperatorFromContext(c)
		if authOp == nil || authOp.Role != models.RoleSuperAdmin {
			c.JSON(403, models.ErrorResponse{Error: "forbidden", Message: "only a super_admin can invite new operators"})
			return
		}
	}

	hash, err := storage.HashPassword(req.Password)
	if err != nil {
		ac.log.Warn("register: password hash failed", "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "registration failed"})
		return
	}

	op := &models.Operator{
		ID:            GenerateID(),
		Email:         req.Email,
		Name:          strings.TrimSpace(req.Name),
		PasswordHash:  hash,
		Role:          role,
		EmailVerified: false, // Not verified until email verification completes
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := ac.store.CreateOperator(ctx, op); err != nil {
		if IsUniqueViolation(err) {
			c.JSON(409, models.ErrorResponse{Error: "email_conflict", Message: "an account with this email already exists"})
			return
		}
		ac.log.Warn("register: create failed", "email", req.Email, "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "registration failed"})
		return
	}

	// Send verification email.
	ac.sendVerificationEmail(ctx, op)

	ac.log.Info("register: operator created", "email", req.Email, "role", role)
	c.JSON(201, models.MessageResponse{
		Message: "Registration successful. Please check your email to verify your account.",
	})
}

// Me returns the operator profile for the authenticated caller.
func (ac *AuthController) Me(c *gin.Context) {
	op := GetOperatorFromContext(c)
	if op == nil {
		c.JSON(401, models.ErrorResponse{Error: "unauthorized", Message: "authentication required"})
		return
	}
	c.JSON(200, op.ToResponse())
}

// Logout clears the session cookie.
func (ac *AuthController) Logout(c *gin.Context) {
	// Clear the session cookie.
	clearCookie := ac.session.ClearSessionCookie()
	http.SetCookie(c.Writer, clearCookie)
	c.JSON(200, map[string]any{"ok": true})
}

// GetOperatorFromContext extracts the authenticated operator from the Gin context.
func GetOperatorFromContext(c *gin.Context) *models.Operator {
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

// GenerateID generates a unique ID string.
func GenerateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("gen-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// IsUniqueViolation checks if an error is a SQLite unique constraint violation.
func IsUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "UNIQUE constraint"))
}