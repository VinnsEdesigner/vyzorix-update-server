package auth

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/ratelimit"

	"github.com/gin-gonic/gin"
)

// Dependencies holds all auth handler dependencies.
type Dependencies struct {
	OperatorRepo        operator.Repository
	OAuthStateRepo      OAuthStateProvider
	Lockout             *middleware.Lockout
	GoogleVerifier      *infraauth.GoogleTokenVerifier
	ClientService       *client.Service
	EmailService        *emailService.Service
	SessionManager      *infraauth.SessionManager
	AuditLogger         *audit.Logger
	IPIntelligence      *middleware.IPIntelligence
	Presenter           *response.Presenter
	SettingsRateLimiter *middleware.SettingsRateLimiterMiddleware
	AuthService         *auth.AuthService
	Config              config.Config
}

// AllHandlers holds references to all auth handlers.
type AllHandlers struct {
	AuthService   *auth.AuthService
	Login         *LoginHandler
	Register      *RegisterHandler
	Logout        *LogoutHandler
	Me            *MeHandler
	EmailVerify   *EmailVerifyHandler
	PasswordReset *PasswordResetHandler
	Refresh       *RefreshHandler // MISSING endpoint - added
	MFA           *MFAHandler
	OAuth         *OAuthHandler
	Settings      *SettingsHandler
	Admin         *AdminHandler
	ClientCreds   *ClientCredentialsHandler
	Lockout       *LockoutHandler
	Sessions      *SessionsHandler
}

// NewAllHandlers creates all auth handlers with proper dependencies.
func NewAllHandlers(deps *Dependencies) *AllHandlers {
	// Create settings rate limiter if not provided
	settingsRateLimiter := deps.SettingsRateLimiter
	if settingsRateLimiter == nil {
		settingsRateLimiter = middleware.NewSettingsRateLimiterMiddleware(nil)
	}

	return &AllHandlers{
		AuthService:   deps.AuthService,
		Login:         NewLoginHandler(deps.AuthService, deps.Presenter, deps.EmailService),
		Register:      NewRegisterHandler(deps.AuthService, deps.EmailService, deps.Presenter),
		Logout:        NewLogoutHandler(deps.AuthService, deps.Presenter),
		Me:            NewMeHandler(deps.AuthService, deps.Presenter),
		EmailVerify:   NewEmailVerifyHandler(deps.AuthService, deps.EmailService, deps.Presenter),
		PasswordReset: NewPasswordResetHandler(deps.AuthService, deps.EmailService, deps.Presenter),
		Refresh:       NewRefreshHandler(deps.AuthService, deps.Presenter), // MISSING endpoint - added
		MFA:           NewMFAHandler(deps.AuthService, deps.OperatorRepo, deps.Presenter),
		OAuth:         NewOAuthHandler(deps.AuthService, deps.SessionManager, deps.Config, deps.GoogleVerifier, deps.Presenter).WithOAuthStateRepo(deps.OAuthStateRepo),
		Settings:      NewSettingsHandler(deps.AuthService, deps.OperatorRepo, deps.Presenter, settingsRateLimiter),
		Admin:         NewAdminHandler(deps.AuthService, deps.Presenter),
		ClientCreds:   NewClientCredentialsHandler(deps.AuthService, deps.ClientService, deps.Presenter),
		Lockout:       NewLockoutHandler(deps.AuthService, deps.Lockout, deps.Presenter),
		Sessions:      NewSessionsHandler(deps.AuthService, deps.SessionManager, deps.Presenter),
	}
}

// RegisterRoutes registers all auth routes under the given router group.
func (h *AllHandlers) RegisterRoutes(rg *gin.RouterGroup, cookieAuth *middleware.CookieAuth) {
	// Public auth endpoints with NoCache and POST-only restriction
	publicAuth := rg.Group("")
	publicAuth.Use(middleware.NoCache())
	{
		// Login with validation
		publicAuth.POST("/login", middleware.POST(),
			middleware.ValidationMiddleware(&middleware.LoginSchema{}),
			h.Login.Handle,
		)

		// Register with validation
		publicAuth.POST("/register", middleware.POST(),
			middleware.ValidationMiddleware(&middleware.RegisterSchema{}),
			h.Register.Handle,
		)

		// Password reset with validation
		publicAuth.POST("/forgot-password", middleware.POST(),
			middleware.ValidationMiddleware(&middleware.ForgotPasswordSchema{}),
			h.PasswordReset.ForgotPassword,
		)
		publicAuth.POST("/reset-password", middleware.POST(),
			middleware.ValidationMiddleware(&middleware.ResetPasswordSchema{}),
			h.PasswordReset.ResetPassword,
		)
		publicAuth.POST("/resend-password-reset", middleware.POST(),
			middleware.ValidationMiddleware(&middleware.ForgotPasswordSchema{}),
			h.PasswordReset.ResendPasswordReset,
		)

		// Email verification
		publicAuth.POST("/verify-email", middleware.POST(), h.EmailVerify.VerifyEmail)
		publicAuth.GET("/verify-email", middleware.GET(), h.EmailVerify.VerifyEmailGet)
		publicAuth.POST("/resend-verification", middleware.POST(), h.EmailVerify.ResendVerification)
		publicAuth.GET("/resend-verification", middleware.GET(), h.EmailVerify.ResendVerificationGet)
		publicAuth.POST("/cancel-verification", middleware.POST(), h.EmailVerify.CancelVerification)
		publicAuth.GET("/poll-verification", middleware.GET(), h.EmailVerify.PollVerification)

		// Token refresh - MISSING endpoint from bug analysis
		publicAuth.POST("/refresh", h.Refresh.Handle)
	}

	// OAuth endpoints (GET only)
	rg.GET("/google", h.OAuth.GoogleLogin)
	rg.GET("/google/callback", h.OAuth.GoogleCallback)
	rg.GET("/github", h.OAuth.GitHubLogin)
	rg.GET("/github/callback", h.OAuth.GitHubCallback)

	// Authenticated endpoints (require session cookie)
	authenticated := rg.Group("")
	authenticated.Use(cookieAuth.Middleware())
	authenticated.Use(middleware.NoCache())
	{
		authenticated.GET("/me", h.Me.Handle)
		authenticated.PATCH("/me", h.Settings.UpdateName)
		authenticated.GET("/me/settings", h.Settings.GetSettings)
		authenticated.PATCH("/me/settings", h.Settings.UpdateSettings)
		authenticated.GET("/me/thresholds", h.Settings.GetThresholds)
		authenticated.PATCH("/me/thresholds", h.Settings.UpdateThresholds)
		authenticated.GET("/me/notifications", h.Settings.GetNotifications)
		authenticated.PATCH("/me/notifications", h.Settings.UpdateNotifications)
		authenticated.POST("/me/notifications/webhook/test", h.Settings.TestWebhook)
		authenticated.POST("/me/notifications/webhook/rotate", h.Settings.RotateWebhookSecret)
		authenticated.POST("/logout", h.Logout.Handle)
		authenticated.GET("/lockout/status", h.Lockout.GetLockoutStatus)
	}

	// SuperAdmin-only operator management routes
	adminOperators := rg.Group("/admin")
	adminOperators.Use(cookieAuth.Middleware())
	adminOperators.Use(middleware.RequireSuperAdmin())
	adminOperators.Use(middleware.NoCache())
	{
		adminOperators.GET("/operators", h.Admin.ListOperators)
		adminOperators.POST("/operators", h.Admin.CreateOperator)
		adminOperators.GET("/operators/:id", h.Admin.GetOperator)
		adminOperators.PATCH("/operators/:id", h.Admin.UpdateOperator)
		adminOperators.DELETE("/operators/:id", h.Admin.DeleteOperator)
	}

	// Admin lockout management (SuperAdmin only)
	adminLockout := rg.Group("/admin/lockout")
	adminLockout.Use(cookieAuth.Middleware())
	adminLockout.Use(middleware.RequireSuperAdmin())
	adminLockout.Use(middleware.NoCache())
	{
		adminLockout.POST("/unlock/:operator_id", h.Lockout.UnlockAccount)
	}

	// MFA endpoints (require authentication)
	mfa := rg.Group("/mfa")
	mfa.Use(cookieAuth.Middleware())
	mfa.Use(middleware.NoCache())
	{
		mfa.GET("/status", h.MFA.GetMFAStatus)
		mfa.POST("/enroll", h.MFA.EnrollMFA)
		mfa.POST("/verify-setup", h.MFA.VerifySetupMFA)
		mfa.POST("/enable", h.MFA.EnableMFA)
		mfa.POST("/disable", h.MFA.DisableMFA)
		mfa.POST("/verify-backup", h.MFA.VerifyBackupCode)
		mfa.POST("/regenerate-backup-codes", h.MFA.RegenerateBackupCodes)

		// MFA Verify - CRITICAL: rate limited by operator_id to prevent brute force
		// 3: Also apply lockout middleware to check for locked accounts
		// 3 attempts per minute per operator_id
		mfa.POST("/verify", h.Lockout.Middleware(), mfaRateLimitMiddleware(), h.MFA.VerifyMFA)
	}

	// Session management endpoints
	sessions := rg.Group("/sessions")
	sessions.Use(cookieAuth.Middleware())
	sessions.Use(middleware.NoCache())
	{
		sessions.GET("", h.Sessions.ListSessions)
		sessions.GET("/concurrent", h.Sessions.CheckConcurrent)
		sessions.DELETE("/:id", h.Sessions.RevokeSession)
		sessions.DELETE("", h.Sessions.RevokeAllExceptCurrent)
		sessions.POST("/revoke-all", h.Sessions.RevokeAllDevices)
		sessions.GET("/:id", h.Sessions.GetSession) // GET specific session
	}

	// Client credentials (require authentication)
	clientCreds := rg.Group("/client-credentials")
	clientCreds.Use(cookieAuth.Middleware())
	clientCreds.Use(middleware.NoCache())
	{
		clientCreds.POST("", h.ClientCreds.Create)
		clientCreds.GET("", h.ClientCreds.List)
		clientCreds.GET("/:clientId", h.ClientCreds.Get)
		clientCreds.DELETE("/:clientId", h.ClientCreds.Delete)
		clientCreds.PATCH("/:clientId", h.ClientCreds.Update) // Update client
		clientCreds.POST("/:clientId/rotate-secret", h.ClientCreds.RotateSecret) // Rotate secret
	}
}

// mfaRateLimitMiddleware returns a rate limiting middleware for MFA verify endpoint.
// Uses operator_id as key to prevent brute force attacks on TOTP codes.
// Allows 3 attempts per minute per operator_id.
func mfaRateLimitMiddleware() gin.HandlerFunc {
	return ratelimit.MFAVerifyLimiter.Middleware(ratelimit.Config{
		KeyFunc: func(c *gin.Context) string {
			// Extract operator_id from request body for rate limiting
			// Must restore body after reading since Gin doesn't support body rewind
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			var req struct {
				OperatorID string `json:"operator_id"`
			}
			if err := json.Unmarshal(bodyBytes, &req); err == nil && req.OperatorID != "" {
				return "mfa_verify:" + req.OperatorID
			}
			// Fallback to IP if operator_id not available
			return "mfa_verify:ip:" + c.ClientIP()
		},
	})
}
