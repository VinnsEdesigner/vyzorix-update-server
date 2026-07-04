package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// Dependencies holds all auth handler dependencies.
type Dependencies struct {
	OperatorRepo        operator.Repository
	AuthService         *auth.AuthService
	SessionManager      *infraauth.SessionManager
	GoogleVerifier      *infraauth.GoogleTokenVerifier
	ClientService       *client.Service
	EmailService        *emailService.Service
	Lockout             *middleware.Lockout
	AuditLogger         *audit.Logger
	IPIntelligence      *middleware.IPIntelligence
	Presenter           *response.Presenter
	Config              config.Config
	SettingsRateLimiter *middleware.SettingsRateLimiterMiddleware
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
		Login:         NewLoginHandler(deps.AuthService, deps.Presenter),
		Register:      NewRegisterHandler(deps.AuthService, deps.EmailService, deps.Presenter),
		Logout:        NewLogoutHandler(deps.AuthService, deps.Presenter),
		Me:            NewMeHandler(deps.AuthService, deps.Presenter),
		EmailVerify:   NewEmailVerifyHandler(deps.AuthService, deps.EmailService, deps.Presenter),
		PasswordReset: NewPasswordResetHandler(deps.AuthService, deps.EmailService, deps.Presenter),
		MFA:           NewMFAHandler(deps.AuthService, deps.OperatorRepo, deps.Presenter),
		OAuth:         NewOAuthHandler(deps.AuthService, deps.SessionManager, deps.Config, deps.GoogleVerifier, deps.Presenter),
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
		publicAuth.POST("/resend-verification", middleware.POST(), h.EmailVerify.ResendVerification)
		publicAuth.POST("/cancel-verification", middleware.POST(), h.EmailVerify.CancelVerification)
		publicAuth.GET("/poll-verification", middleware.GET(), h.EmailVerify.PollVerification)
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
		authenticated.GET("/admin/operators", h.Admin.ListOperators)
		authenticated.POST("/admin/operators", h.Admin.CreateOperator)
		authenticated.GET("/admin/operators/:id", h.Admin.GetOperator)
		authenticated.PATCH("/admin/operators/:id", h.Admin.UpdateOperator)
		authenticated.DELETE("/admin/operators/:id", h.Admin.DeleteOperator)
	}

	// Admin lockout management
	adminLockout := rg.Group("/admin/lockout")
	adminLockout.Use(cookieAuth.Middleware())
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
		mfa.POST("/verify", h.MFA.VerifyMFA) // Main MFA verification during login
		mfa.POST("/enable", h.MFA.EnableMFA)
		mfa.POST("/disable", h.MFA.DisableMFA)
		mfa.POST("/verify-backup", h.MFA.VerifyBackupCode)
		mfa.POST("/regenerate-backup-codes", h.MFA.RegenerateBackupCodes)
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
	}
}
