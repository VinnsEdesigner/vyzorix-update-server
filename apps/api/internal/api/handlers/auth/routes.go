package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"

	"github.com/gin-gonic/gin"
)

// Dependencies holds all auth handler dependencies.
type Dependencies struct {
	AuthService    *auth.AuthService
	SessionManager *infraauth.SessionManager
	Config         config.Config
	GoogleVerifier *infraauth.GoogleTokenVerifier
	ClientService  *client.Service
	EmailService   *emailService.Service
	Lockout        *middleware.Lockout
	OperatorRepo   operator.Repository
}

// AllHandlers holds references to all auth handlers.
type AllHandlers struct {
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
}

// NewAllHandlers creates all auth handlers with proper dependencies.
func NewAllHandlers(deps *Dependencies) *AllHandlers {
	return &AllHandlers{
		Login:         NewLoginHandler(deps.AuthService),
		Register:      NewRegisterHandler(deps.AuthService, deps.EmailService),
		Logout:        NewLogoutHandler(deps.AuthService),
		Me:            NewMeHandler(deps.AuthService),
		EmailVerify:   NewEmailVerifyHandler(deps.AuthService, deps.EmailService),
		PasswordReset: NewPasswordResetHandler(deps.AuthService, deps.EmailService),
		MFA:           NewMFAHandler(deps.AuthService, deps.OperatorRepo),
		OAuth:         NewOAuthHandler(deps.AuthService, deps.SessionManager, deps.Config, deps.GoogleVerifier),
		Settings:      NewSettingsHandler(deps.AuthService),
		Admin:         NewAdminHandler(deps.AuthService),
		ClientCreds:   NewClientCredentialsHandler(deps.AuthService, deps.ClientService),
		Lockout:       NewLockoutHandler(deps.AuthService, deps.Lockout),
	}
}

// RegisterRoutes registers all auth routes under the given router group.
func (h *AllHandlers) RegisterRoutes(rg *gin.RouterGroup, cookieAuth *middleware.CookieAuth) {
	// Public auth endpoints
	rg.POST("/login", h.Login.Handle)
	rg.POST("/register", h.Register.Handle)
	rg.POST("/forgot-password", h.PasswordReset.ForgotPassword)
	rg.POST("/reset-password", h.PasswordReset.ResetPassword)

	// Email verification
	rg.POST("/verify-email", h.EmailVerify.VerifyEmail)
	rg.POST("/resend-verification", h.EmailVerify.ResendVerification)
	rg.POST("/cancel-verification", h.EmailVerify.CancelVerification)
	rg.GET("/poll-verification", h.EmailVerify.PollVerification)
	rg.POST("/resend-password-reset", h.PasswordReset.ResendPasswordReset)

	// OAuth endpoints
	rg.GET("/google", h.OAuth.GoogleLogin)
	rg.GET("/google/callback", h.OAuth.GoogleCallback)
	rg.GET("/github", h.OAuth.GitHubLogin)
	rg.GET("/github/callback", h.OAuth.GitHubCallback)

	// Authenticated endpoints (require session cookie)
	authenticated := rg.Group("")
	authenticated.Use(cookieAuth.Middleware())
	{
		authenticated.GET("/me", h.Me.Handle)
		authenticated.PATCH("/me", h.Settings.UpdateName)
		authenticated.PATCH("/me/settings", h.Settings.UpdateSettings)
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
	{
		adminLockout.POST("/unlock/:operator_id", h.Lockout.UnlockAccount)
	}

	// MFA endpoints (require authentication)
	mfa := rg.Group("/mfa")
	mfa.Use(cookieAuth.Middleware())
	{
		mfa.GET("/status", h.MFA.GetMFAStatus)
		mfa.POST("/enroll", h.MFA.EnrollMFA)
		mfa.POST("/verify-setup", h.MFA.VerifySetupMFA)
		mfa.POST("/enable", h.MFA.EnableMFA)
		mfa.POST("/disable", h.MFA.DisableMFA)
		mfa.POST("/verify-backup", h.MFA.VerifyBackupCode)
		mfa.POST("/regenerate-backup-codes", h.MFA.RegenerateBackupCodes)
	}

	// Client credentials (require authentication)
	clientCreds := rg.Group("/client-credentials")
	clientCreds.Use(cookieAuth.Middleware())
	{
		clientCreds.POST("", h.ClientCreds.Create)
		clientCreds.GET("", h.ClientCreds.List)
		clientCreds.GET("/:clientId", h.ClientCreds.Get)
		clientCreds.DELETE("/:clientId", h.ClientCreds.Delete)
	}
}
