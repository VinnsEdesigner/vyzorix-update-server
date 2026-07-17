// Package wire provides dependency injection utilities for the API server.
//
// This package centralizes all dependency wiring to keep server.go minimal.
// Use WireServer() to create a fully configured Server instance.
package wire

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	orgapplication "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	updatesapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	infraConfig "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
	"log/slog"
)

// ServerDependencies contains all dependencies needed to create a Server.
type ServerDependencies struct {
	FCMNotifier       fcm.Notifier
	OperatorRepo      operator.Repository
	EmailService      *emailService.Service
	CommandService    *command.Service
	AuthService      *auth.AuthService
	AuthLimiter      *middleware.RateLimiter
	IPIntelligence    *middleware.IPIntelligence
	Log               *slog.Logger
	SessionManager    *infraauth.SessionManager
	GoogleVerifier    *infraauth.GoogleTokenVerifier
	RateLimiter      *middleware.RateLimiter
	Hub               *hub.Hub
	ClientService     *client.Service
	DB               *storage.SQLite
	Lockout          *middleware.Lockout
	DeviceService    *device.Service
	Metrics          *metrics.Metrics
	AuditLogger      *audit.Logger
	TelemetryRepo    *storage.TelemetryRepository
	UpdatesStorage   *storage.UpdatesStorage
	UpdatesService   *updatesapp.Service
	APIKeyService    *keys.APIKeyService
	Config           infraConfig.Config
	OrgService       *orgapplication.OrganizationService
	MemberService    *orgapplication.MemberService
	InvitationService *orgapplication.InvitationService
}

// ServerResult contains the fully wired server components.
type ServerResult struct {
	Engine            *gin.Engine
	MiddlewareSet     *MiddlewareSet
	HandlerSet        *HandlerSet
	Presenter         *response.Presenter
	HmacVerifier      *cryptohmac.Verifier
	EncryptKeyFn      func(clientID string) ([]byte, bool)
	CookieAuth        *middleware.CookieAuth
	SignatureVerifier *middleware.SignatureVerifier
	Lockout           *middleware.Lockout
	CSRFProtector     *middleware.CSRFProtector
	TurnstileVerifier *middleware.TurnstileVerifier
	RevocationList    *infraauth.RevocationList
	IPIntelligence    *middleware.IPIntelligence
	SessionManager    *infraauth.SessionManager
}

// Server combines ServerDependencies and ServerResult for wire-compatible return.
type Server struct {
	Dependencies *ServerDependencies
	Result       *ServerResult
}

// WireServer wires all server dependencies.
func WireServer(deps ServerDependencies) *ServerResult {
	result := &ServerResult{}

	// Wire middleware first
	mwCfg := MiddlewareConfig{
		Log:              deps.Log,
		SessionManager:   deps.SessionManager,
		AuthService:      deps.AuthService,
		ClientService:    deps.ClientService,
		AllowedOrigins:   deps.Config.AllowedOrigins,
		EnforceHMAC:      deps.Config.EnforceHMAC,
		HMACWindow:       deps.Config.HMACWindow,
		PublicDir:        deps.Config.PublicDir,
		JWTSecret:        deps.Config.JWTSecret,
		RateLimitPerMin:  100,
		AuthRateLimitMin: 5,
	}
	result.MiddlewareSet = WireMiddleware(mwCfg)

	// Get middleware instances
	result.HmacVerifier = result.MiddlewareSet.HmacVerifier
	result.EncryptKeyFn = result.MiddlewareSet.EncryptKeyFn
	result.CookieAuth = result.MiddlewareSet.CookieAuth
	result.SignatureVerifier = result.MiddlewareSet.SignatureVerifier
	result.Lockout = result.MiddlewareSet.Lockout
	result.CSRFProtector = result.MiddlewareSet.CSRFProtector
	result.TurnstileVerifier = result.MiddlewareSet.TurnstileVerifier
	result.RevocationList = result.MiddlewareSet.RevocationList
	result.IPIntelligence = result.MiddlewareSet.IPIntelligence
	result.SessionManager = deps.SessionManager

	// Create presenter
	result.Presenter = response.NewPresenter(deps.AuthService, deps.AuditLogger, deps.IPIntelligence)

	// Wire handlers
	handlerDeps := HandlerDependencies{
		AuthService:       deps.AuthService,
		SessionManager:    deps.SessionManager,
		Config:            deps.Config,
		GoogleVerifier:    deps.GoogleVerifier,
		ClientService:     deps.ClientService,
		EmailService:      deps.EmailService,
		Lockout:           deps.Lockout,
		OperatorRepo:      deps.OperatorRepo,
		AuditLogger:       deps.AuditLogger,
		IPIntelligence:    deps.IPIntelligence,
		Presenter:         result.Presenter,
		DeviceService:     deps.DeviceService,
		Hub:               deps.Hub,
		CommandService:    deps.CommandService,
		FCMNotifier:       deps.FCMNotifier,
		Log:               deps.Log,
		HmacVerifier:      result.HmacVerifier,
		DB:                deps.DB,
		UpdatesStorage:    deps.UpdatesStorage,
		OrgService:        deps.OrgService,
		MemberService:     deps.MemberService,
		InvitationService: deps.InvitationService,
	}
	result.HandlerSet = WireHandlers(handlerDeps)

	// Create Gin engine
	if deps.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	result.Engine = gin.New()
	result.Engine.Use(middleware.GinPanicRecovery(deps.Log))

	return result
}
