package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/admin"
	authhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	cmdhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/command"
	confirmationhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/confirmation"
	dashboardhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/dashboard"
	devicehandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/device"
	diagnosticshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/diagnostics"
	inboxhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/inbox"
	organizationhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/organization"
	updaterhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updater"
	updateshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updates"
	websockethandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/wire"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/confirmation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	diagnosticsapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/inbox"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	orgapplication "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	updatesapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	domaincommand "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/appcheck"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	infraMetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
)

// ServerConfig holds the server configuration.
type ServerConfig struct {
	FCMNotifier           fcm.Notifier
	OperatorRepo          operator.Repository
	OAuthStateRepo        authhandlers.OAuthStateProvider
	AuthService           *auth.AuthService
	Lockout               *middleware.Lockout
	AuthLimiter           *middleware.RateLimiter
	IPIntelligence        *middleware.IPIntelligence
	Log                   *slog.Logger
	SessionManager        *infraauth.SessionManager
	GoogleVerifier        *infraauth.GoogleTokenVerifier
	Hub                   *hub.Hub
	CommandService        *command.Service
	ConfirmationService   *confirmation.Service
	EmailService          *emailService.Service
	DB                    *storage.SQLite
	ClientService         *client.Service
	DeviceService         *device.Service
	Metrics               *infraMetrics.Metrics
	AuditLogger           *audit.Logger
	RateLimiter           *middleware.RateLimiter
	UpdatesService        *updatesapp.Service
	APIKeyService         *keys.APIKeyService
	AppCheckVerifier      *appcheck.Verifier
	DeviceSettingsService *device.DeviceSettingsService
	OrgSettingsService    *orgapplication.OrganizationSettingsService
	Config                config.Config
}

// Server is the main API server.
type Server struct {
	encryptKeyFn                func(clientID string) ([]byte, bool)
	authHandlers                *authhandlers.AllHandlers
	hub                         *hub.Hub
	engine                      *gin.Engine
	log                         *slog.Logger
	deviceStatusHandler         *devicehandlers.StatusHandler
	sessionManager              *infraauth.SessionManager
	rateLimiter                 *middleware.RateLimiter
	authLimiter                 *middleware.RateLimiter
	cookieAuth                  *middleware.CookieAuth
	signatureVerifier           *middleware.SignatureVerifier
	lockout                     *middleware.Lockout
	csrfProtector               *middleware.CSRFProtector
	turnstileVerifier           *middleware.TurnstileVerifier
	revocationList              *infraauth.RevocationList
	ipIntelligence              *middleware.IPIntelligence
	hmacVerifier                *cryptohmac.Verifier
	sessionSignatureVerifier    *cryptohmac.Verifier
	mwFactory                   *middleware.MiddlewareFactory
	db                          *storage.SQLite
	dashboardRateLimiter        *middleware.DashboardRateLimiterMiddleware
	deviceRegRateLimiter        *middleware.DeviceRegistrationRateLimiterMiddleware
	AuditLogger                 *audit.Logger
	deviceUpdaterHandler        *devicehandlers.UpdaterHandler
	deviceListHandler           *devicehandlers.ListHandler
	devicesHandler              *devicehandlers.DevicesHandler
	commandHandler              *cmdhandlers.ExecuteHandler
	confirmationHandler         *confirmationhandlers.Handler
	streamHandler               *websockethandlers.StreamHandler
	telemetryHistoryHandler     *handlers.TelemetryHistoryHandler
	connectionStatusHandler     *handlers.ConnectionStatusHandler
	adminClientsHandler         *admin.ClientsHandler
	metricsHandler              *infraMetrics.MetricsHandler
	commandHistoryHandler       *cmdhandlers.HistoryHandler
	deviceLogsHandler           *devicehandlers.LogsHandler
	deviceEventsHandler         *devicehandlers.EventsHandler
	deviceMetricsHandler        *devicehandlers.MetricsHandler
	deviceTelemetryHandler      *devicehandlers.TelemetryHandler
	dashboardStatsHandler       *dashboardhandlers.StatsHandler
	updaterHandler              *updaterhandlers.Handler
	updatesHandler              *updateshandlers.UpdatesHandler
	inboxHandler                *inboxhandlers.Handler
	inboxService                *inbox.Service
	deviceConfirmHandler        *devicehandlers.ConfirmHandler
	diagnosticsHandler          *diagnosticshandlers.Handler
	diagnosticsInspectHandler   *diagnosticshandlers.InspectHandler
	diagnosticsTimelineHandler  *diagnosticshandlers.TimelineHandler
	apiKeysHandler              *authhandlers.Handler
	superAdminAPIKeys           *admin.SuperAdminHandler
	tenantAPIKeyAuth            *middleware.TenantAPIKeyAuth
	apiKeyRateLimiter           *middleware.InMemoryRateLimiter
	organizationHandler         *organizationhandlers.OrganizationHandler
	organizationSettingsHandler *organizationhandlers.SettingsHandler
	deviceSettingsHandler       *devicehandlers.SettingsHandler
	invitationHandler           *organizationhandlers.InvitationHandler
	memberHandler               *organizationhandlers.MemberHandler
	transferHandler             *devicehandlers.TransferHandler
	DeviceRepo                  *storage.DeviceRepository
	InvitationService           *orgapplication.InvitationService
	idempotencyRepo             *storage.IdempotencyRepository
	config                      config.Config
}

// NewServer creates a new API server with wired-up dependencies.
func NewServer(cfg *ServerConfig) *Server {
	if cfg.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.GinPanicRecovery(cfg.Log))

	// Wire middleware using wire package.
	mwSet := wire.WireMiddleware(wire.MiddlewareConfig{
		Log:              cfg.Log,
		SessionManager:   cfg.SessionManager,
		AuthService:      cfg.AuthService,
		ClientService:    cfg.ClientService,
		AllowedOrigins:   cfg.Config.AllowedOrigins,
		EnforceHMAC:      cfg.Config.EnforceHMAC,
		HMACWindow:       cfg.Config.HMACWindow,
		PublicDir:        cfg.Config.PublicDir,
		JWTSecret:        cfg.Config.JWTSecret,
		RateLimitPerMin:  cfg.Config.RateLimitPerMin,
		AuthRateLimitMin: cfg.Config.AuthRateLimitMin,
		APIKeyService:    cfg.APIKeyService,
		AuditLogger:      cfg.AuditLogger,
	})

	s := &Server{
		engine:                   engine,
		mwFactory:                mwSet.Factory,
		rateLimiter:              cfg.RateLimiter,
		authLimiter:              cfg.AuthLimiter,
		config:                   cfg.Config,
		log:                      cfg.Log,
		cookieAuth:               middleware.NewCookieAuth(cfg.SessionManager, cfg.AuthService),
		signatureVerifier:        mwSet.SignatureVerifier,
		lockout:                  mwSet.Lockout,
		csrfProtector:            mwSet.CSRFProtector,
		turnstileVerifier:        mwSet.TurnstileVerifier,
		revocationList:           mwSet.RevocationList,
		ipIntelligence:           mwSet.IPIntelligence,
		hmacVerifier:             mwSet.HmacVerifier,
		sessionSignatureVerifier: mwSet.SessionSignVerifier,
		encryptKeyFn:             mwSet.EncryptKeyFn,
		sessionManager:           cfg.SessionManager,
		db:                       cfg.DB,
		hub:                      cfg.Hub,
		tenantAPIKeyAuth:         mwSet.TenantAPIKeyAuth,
		apiKeyRateLimiter:        mwSet.APIKeyRateLimiter,
	}

	// Create presenter and wire handlers.
	presenter := response.NewPresenter(cfg.AuthService, cfg.AuditLogger, cfg.IPIntelligence)
	s.wireHandlers(cfg, presenter, mwSet)

	// Start Hub if available.
	if cfg.Hub != nil {
		go cfg.Hub.Run(context.Background())
	}

	// Initialize metrics handler.
	if cfg.Metrics != nil {
		s.metricsHandler = infraMetrics.NewMetricsHandler(cfg.Metrics)
		s.engine.Use(infraMetrics.Middleware(cfg.Metrics))
	}

	s.setupRoutes()

	// Set audit logger.
	s.AuditLogger = cfg.AuditLogger

	return s
}

// wireHandlers creates and assigns all handler instances.
func (s *Server) wireHandlers(cfg *ServerConfig, presenter *response.Presenter, mwSet *wire.MiddlewareSet) {
	// Auth handlers.
	s.authHandlers = authhandlers.NewAllHandlers(&authhandlers.Dependencies{
		AuthService:    cfg.AuthService,
		SessionManager: cfg.SessionManager,
		Config:         cfg.Config,
		GoogleVerifier: cfg.GoogleVerifier,
		ClientService:  cfg.ClientService,
		EmailService:   cfg.EmailService,
		Lockout:        mwSet.Lockout,
		OperatorRepo:   cfg.OperatorRepo,
		AuditLogger:    cfg.AuditLogger,
		IPIntelligence: mwSet.IPIntelligence,
		Presenter:      presenter,
		OAuthStateRepo: cfg.OAuthStateRepo,
	})

	// Device handlers.
	// DEPRECATED: s.deviceRegisterHandler = devicehandlers.NewRegisterHandler(cfg.DeviceService) // /v1/device/register removed.
	s.deviceStatusHandler = devicehandlers.NewStatusHandler(cfg.DeviceService)
	s.deviceUpdaterHandler = devicehandlers.NewUpdaterHandler(cfg.DeviceService)
	s.deviceListHandler = devicehandlers.NewListHandler(cfg.DeviceService, cfg.Hub)

	// Command handler. Risk evaluator gates dangerous commands; audit logger
	// records every execution attempt. Fall back to a no-op if unset so the
	// handler never holds a nil dependency. When a confirmation service is
	// configured, the confirmation handler also serves as the command handler's
	// confirmation consumer; otherwise risky commands are blocked (425).
	var cmdAud cmdhandlers.AuditLogger = audit.NewNoOpLogger()
	if cfg.AuditLogger != nil {
		cmdAud = cfg.AuditLogger
	}
	riskEval := domaincommand.NewRiskEvaluator()
	var confirmConsumer cmdhandlers.ConfirmationConsumer
	if cfg.ConfirmationService != nil {
		s.confirmationHandler = confirmationhandlers.NewHandler(cfg.ConfirmationService, cfg.DeviceService, riskEval)
		confirmConsumer = s.confirmationHandler
	}
	s.commandHandler = cmdhandlers.NewExecuteHandler(cfg.CommandService, cfg.DeviceService, cfg.Hub, cfg.FCMNotifier, riskEval, cmdAud, confirmConsumer)

	// WebSocket handler.
	s.streamHandler = websockethandlers.NewStreamHandler(cfg.Log, cfg.Config, cfg.Hub, *mwSet.HmacVerifier, cfg.AuditLogger)

	// Telemetry history handler.
	s.telemetryHistoryHandler = handlers.NewTelemetryHistoryHandler(
		cfg.Log,
		storage.NewTelemetryRepository(cfg.DB.DB()),
		storage.NewDeviceRepository(cfg.DB.DB()),
		nil,
	)

	// Connection status handler.
	s.connectionStatusHandler = handlers.NewConnectionStatusHandler(cfg.Log, cfg.Hub, storage.NewDeviceRepository(cfg.DB.DB()))

	// Admin handlers.
	s.adminClientsHandler = admin.NewClientsHandler(cfg.ClientService)

	// API key handlers.
	if cfg.APIKeyService != nil {
		s.apiKeysHandler = authhandlers.NewHandler(cfg.APIKeyService, cfg.AuditLogger)
		s.superAdminAPIKeys = admin.NewSuperAdminHandler(cfg.APIKeyService, cfg.AuditLogger)
	}

	// Dashboard command handlers - wire up new handlers.
	s.wireDashboardHandlers(cfg)

	// Initialize dashboard rate limiter.
	s.dashboardRateLimiter = middleware.NewDashboardRateLimiterMiddleware(nil)
}

// Routes returns the Gin engine for serving.
func (s *Server) Routes() http.Handler {
	return s.engine
}

// wireDashboardHandlers creates and assigns dashboard command handler instances.
func (s *Server) wireDashboardHandlers(cfg *ServerConfig) {
	logsRepo, metricsRepo, eventRepo := s.createStorageRepositories(cfg)
	historySvc, logsSvc, metricsSvc, dashboardSvc := s.createDashboardServices(cfg, logsRepo, metricsRepo)
	s.createDashboardHandlers(cfg, historySvc, logsSvc, metricsSvc, eventRepo, dashboardSvc)
	s.createUpdatesHandler(cfg)
	s.wireInboxHandler(cfg)
	s.wireConfirmHandler(cfg)
	s.wireDiagnosticsHandler(cfg)
}

// createStorageRepositories creates storage repositories for dashboard.
func (s *Server) createStorageRepositories(cfg *ServerConfig) (*storage.LogsRepository, *storage.MetricsRepository, *storage.EventRepository) {
	if cfg.DB == nil {
		return nil, nil, nil
	}
	return storage.NewLogsRepository(cfg.DB.DB()),
		storage.NewMetricsRepository(cfg.DB.DB()),
		storage.NewEventRepository(cfg.DB.DB())
}

// createDashboardServices creates dashboard-related services.
func (s *Server) createDashboardServices(cfg *ServerConfig, logsRepo *storage.LogsRepository, metricsRepo *storage.MetricsRepository) (*command.HistoryService, *logs.Service, *appmetrics.Service, *dashboard.Service) {
	var historySvc *command.HistoryService
	var logsSvc *logs.Service
	var metricsSvc *appmetrics.Service
	var dashboardSvc *dashboard.Service

	if cfg.CommandService != nil && cfg.DeviceService != nil {
		historySvc = command.NewHistoryService(cfg.CommandService.CommandRepo(), cfg.DeviceService.DeviceRepo())
	}

	if logsRepo != nil {
		logsSvc = logs.NewService(logsRepo, cfg.Log)
	}

	if metricsRepo != nil {
		metricsSvc = s.createMetricsService(cfg, metricsRepo)
	}

	if cfg.CommandService != nil && cfg.DeviceService != nil && logsRepo != nil {
		dashboardSvc = dashboard.NewService(cfg.DeviceService.DeviceRepo(), cfg.CommandService.CommandRepo(), logsRepo)
	}

	return historySvc, logsSvc, metricsSvc, dashboardSvc
}

// createMetricsService creates the metrics service with threshold resolution.
func (s *Server) createMetricsService(cfg *ServerConfig, metricsRepo *storage.MetricsRepository) *appmetrics.Service {
	var deviceSettingsRepo devicedomain.DeviceSettingsRepository
	var orgSettingsRepo organization.OrganizationSettingsRepository
	if cfg.DeviceSettingsService != nil {
		deviceSettingsRepo = cfg.DeviceSettingsService.SettingsRepo()
	}
	if cfg.OrgSettingsService != nil {
		orgSettingsRepo = cfg.OrgSettingsService.SettingsRepo()
	}
	return appmetrics.NewService(metricsRepo, deviceSettingsRepo, orgSettingsRepo)
}

// createDashboardHandlers creates dashboard-related handlers.
func (s *Server) createDashboardHandlers(cfg *ServerConfig, historySvc *command.HistoryService, logsSvc *logs.Service, metricsSvc *appmetrics.Service, eventRepo *storage.EventRepository, dashboardSvc *dashboard.Service) {
	if historySvc != nil && cfg.DeviceService != nil {
		s.commandHistoryHandler = cmdhandlers.NewHistoryHandler(historySvc, cfg.DeviceService.DeviceRepo(), cfg.Log)
	}

	if logsSvc != nil && cfg.DeviceService != nil {
		s.deviceLogsHandler = devicehandlers.NewLogsHandler(logsSvc, cfg.DeviceService.DeviceRepo(), cfg.Log)
	}

	if eventRepo != nil && cfg.DeviceService != nil {
		s.deviceEventsHandler = devicehandlers.NewEventsHandler(eventRepo, cfg.DeviceService.DeviceRepo(), cfg.Log)
	}

	if metricsSvc != nil && cfg.DeviceService != nil {
		s.deviceMetricsHandler = devicehandlers.NewMetricsHandler(metricsSvc, cfg.DeviceService.DeviceRepo(), cfg.Log)
		s.deviceTelemetryHandler = devicehandlers.NewTelemetryHandler(metricsSvc, cfg.DeviceService.DeviceRepo(), cfg.Log)
	}

	if dashboardSvc != nil {
		s.dashboardStatsHandler = dashboardhandlers.NewStatsHandler(dashboardSvc, cfg.Log)
	}
}

// createUpdatesHandler creates the updates handler.
func (s *Server) createUpdatesHandler(cfg *ServerConfig) {
	if cfg.UpdatesService == nil {
		return
	}
	s.updaterHandler = updaterhandlers.NewHandler(cfg.Log, cfg.Config)
	updatesRateLimiters := middleware.NewUpdatesRateLimiterMiddleware(middleware.DefaultUpdatesRateLimits())
	s.updatesHandler = updateshandlers.NewUpdatesHandler(cfg.UpdatesService, cfg.UpdatesService.GetPushService(), updatesRateLimiters, cfg.AuditLogger, cfg.Config.GitHubWebhookSecret)
}

// wireInboxHandler creates and assigns the inbox handler.
func (s *Server) wireInboxHandler(cfg *ServerConfig) {
	if cfg.DB == nil || cfg.DeviceService == nil {
		return
	}

	// Create inbox repository.
	inboxRepo := storage.NewInboxRepository(cfg.DB.DB())
	regLogRepo := storage.NewRegistrationLogRepository(cfg.DB.DB())

	// Create inbox service with FCM notifier (SafeNotifier for graceful degradation).
	// and device service for creating devices on approval and device lookup.
	var fcmNotifier inbox.FCMNotifier
	if cfg.FCMNotifier != nil {
		fcmNotifier = &fcm.SafeNotifier{Notifier: cfg.FCMNotifier}
	}
	inboxService := inbox.NewService(inboxRepo, regLogRepo, cfg.DeviceService, cfg.DeviceService, fcmNotifier, s.log)

	// Store service for GraphQL resolver wiring.
	s.inboxService = inboxService

	// Enable ACID transactions for enterprise production.
	inboxService.WithTxManager(cfg.DB)

	// Create device registration rate limiter.
	s.deviceRegRateLimiter = middleware.NewDeviceRegistrationRateLimiterMiddleware(nil)

	// Create handler with device attestation.
	if cfg.AppCheckVerifier != nil && cfg.AppCheckVerifier.Enabled() {
		s.log.Info("using Firebase App Check for device attestation")
		s.inboxHandler = inboxhandlers.NewHandlerWithAppCheck(
			inboxService,
			cfg.Config.DeviceSecret,
			cfg.AppCheckVerifier,
		)
	} else if cfg.Config.DeviceSecret != "" {
		s.log.Warn("using HMAC-SHA256 for device attestation - consider enabling Firebase App Check")
		s.inboxHandler = inboxhandlers.NewHandlerWithAttestation(inboxService, cfg.Config.DeviceSecret)
	} else {
		s.log.Warn("no device attestation configured - /v1/device/inbox is public")
		s.inboxHandler = inboxhandlers.NewHandler(inboxService, cfg.Config.DeviceSecret)
	}
}

// wireConfirmHandler creates and assigns the device confirm handler.
func (s *Server) wireConfirmHandler(cfg *ServerConfig) {
	if cfg.DeviceService == nil {
		return
	}

	// Create inbox repository for cleanup after device confirmation.
	inboxRepo := storage.NewInboxRepository(cfg.DB.DB())

	// Create confirm handler with inbox cleanup capability.
	s.deviceConfirmHandler = devicehandlers.NewConfirmHandlerWithCleanup(cfg.DeviceService, inboxRepo)
}

// wireDiagnosticsHandler creates and assigns the diagnostics handlers.
func (s *Server) wireDiagnosticsHandler(cfg *ServerConfig) {
	if cfg.DB == nil || cfg.DeviceService == nil {
		return
	}

	// Create diagnostics repository.
	diagnosticsRepo := storage.NewDiagnosticsRepository(cfg.DB.DB())

	// Create diagnostics service.
	diagnosticsService := diagnosticsapp.NewService(diagnosticsRepo, cfg.DeviceService.DeviceRepo(), s.hub, cfg.Config.DiagnosticsConfig)

	// Get rate limiters.
	var inspectLimit, timelineLimit func(c *gin.Context)
	if s.dashboardRateLimiter != nil {
		inspectLimit = s.dashboardRateLimiter.DeviceInspectLimit()
		timelineLimit = s.dashboardRateLimiter.DeviceTimelineLimit()
	}

	// Create inspect handler.
	s.diagnosticsInspectHandler = diagnosticshandlers.NewInspectHandler(diagnosticsService, inspectLimit)

	// Create timeline handler.
	s.diagnosticsTimelineHandler = diagnosticshandlers.NewTimelineHandler(diagnosticsService, timelineLimit)

	// Create combined handler for backwards compatibility.
	s.diagnosticsHandler = &diagnosticshandlers.Handler{
		InspectHandler:  s.diagnosticsInspectHandler,
		TimelineHandler: s.diagnosticsTimelineHandler,
	}
}

// Handlers are defined in server_handlers.go.
// Routes are defined in server_routes.go.
// GraphQL is defined in server_graphql.go.

// ServerConfigWithDeps is the config for NewServerWithDeps using pre-wired dependencies.
type ServerConfigWithDeps struct {
	Log             *slog.Logger
	DB              *storage.SQLite
	Engine          *gin.Engine
	Middleware      *wire.MiddlewareSet
	HandlerSet      *wire.HandlerSet
	SessionManager  *infraauth.SessionManager
	Hub             *hub.Hub
	AuditLogger     *audit.Logger
	UpdatesService  *updatesapp.Service
	APIKeyService   *keys.APIKeyService
	DeviceRepo      *storage.DeviceRepository
	IdempotencyRepo *storage.IdempotencyRepository
	Config          config.Config
}

// NewServerWithDeps creates a Server using pre-wired dependencies from wire.
func NewServerWithDeps(cfg *ServerConfigWithDeps) *Server {
	if cfg.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ProvideMiddlewareSet does not wire the API key middleware (it is normally
	// produced by WireMiddleware, which is only called from the legacy NewServer
	// constructor). Backfill TenantAPIKeyAuth here so the dependency-injected
	// server (used by api_main.go) also enforces API key auth on tenant routes.
	if cfg.Middleware != nil && cfg.Middleware.TenantAPIKeyAuth == nil && cfg.APIKeyService != nil && cfg.AuditLogger != nil {
		cfg.Middleware.TenantAPIKeyAuth = middleware.NewTenantAPIKeyAuth(cfg.APIKeyService, cfg.AuditLogger)
		if cfg.Middleware.APIKeyRateLimiter == nil {
			cfg.Middleware.APIKeyRateLimiter = middleware.NewInMemoryRateLimiter(100, time.Minute)
		}
	}

	s := &Server{
		engine:                   cfg.Engine,
		mwFactory:                cfg.Middleware.Factory,
		rateLimiter:              cfg.Middleware.RateLimiter,
		authLimiter:              cfg.Middleware.AuthLimiter,
		config:                   cfg.Config,
		log:                      cfg.Log,
		cookieAuth:               cfg.Middleware.CookieAuth,
		lockout:                  cfg.Middleware.Lockout,
		csrfProtector:            cfg.Middleware.CSRFProtector,
		turnstileVerifier:        cfg.Middleware.TurnstileVerifier,
		revocationList:           cfg.Middleware.RevocationList,
		ipIntelligence:           cfg.Middleware.IPIntelligence,
		hmacVerifier:             cfg.Middleware.HmacVerifier,
		sessionSignatureVerifier: cfg.Middleware.SessionSignVerifier,
		sessionManager:           cfg.SessionManager,
		db:                       cfg.DB,
		hub:                      cfg.Hub,
		AuditLogger:              cfg.AuditLogger,
		tenantAPIKeyAuth:         cfg.Middleware.TenantAPIKeyAuth,
		apiKeyRateLimiter:        cfg.Middleware.APIKeyRateLimiter,
		idempotencyRepo:          cfg.IdempotencyRepo,
	}

	// Wire handlers from HandlerSet.
	s.authHandlers = cfg.HandlerSet.Auth
	// DEPRECATED: s.deviceRegisterHandler = cfg.HandlerSet.DeviceRegister // /v1/device/register removed.
	s.deviceStatusHandler = cfg.HandlerSet.DeviceStatus
	s.deviceUpdaterHandler = cfg.HandlerSet.DeviceUpdater
	s.deviceListHandler = cfg.HandlerSet.DeviceList
	s.devicesHandler = cfg.HandlerSet.Devices
	s.commandHandler = cfg.HandlerSet.Command
	s.confirmationHandler = cfg.HandlerSet.Confirmation
	s.streamHandler = cfg.HandlerSet.Stream
	s.telemetryHistoryHandler = cfg.HandlerSet.TelemetryHistory
	s.connectionStatusHandler = cfg.HandlerSet.ConnectionStatus
	s.adminClientsHandler = cfg.HandlerSet.AdminClients
	s.updatesHandler = cfg.HandlerSet.Updates

	// Organization handlers.
	s.organizationHandler = cfg.HandlerSet.Organization
	s.organizationSettingsHandler = cfg.HandlerSet.OrgSettings
	s.deviceSettingsHandler = cfg.HandlerSet.DeviceSettings
	s.invitationHandler = cfg.HandlerSet.Invitation
	s.memberHandler = cfg.HandlerSet.Member
	s.transferHandler = cfg.HandlerSet.Transfer

	// API key handlers.
	if cfg.APIKeyService != nil {
		s.apiKeysHandler = authhandlers.NewHandler(cfg.APIKeyService, cfg.AuditLogger)
		s.superAdminAPIKeys = admin.NewSuperAdminHandler(cfg.APIKeyService, cfg.AuditLogger)
	}

	// Wire inbox and confirm handlers using ServerConfig.
	// Note: FCMNotifier and AppCheckVerifier are passed via HandlerSet.
	serverCfg := &ServerConfig{
		DB:               cfg.DB,
		DeviceService:    cfg.HandlerSet.DeviceService,
		FCMNotifier:      cfg.HandlerSet.FCMNotifier,
		Config:           cfg.Config,
		Log:              cfg.Log,
		AppCheckVerifier: cfg.HandlerSet.AppCheckVerifier,
	}
	s.wireInboxHandler(serverCfg)
	s.wireConfirmHandler(serverCfg)

	// Create updater handler for OTA update distribution.
	s.updaterHandler = updaterhandlers.NewHandler(cfg.Log, cfg.Config)

	// Store deviceRepo for the deletion worker in api_main.go.
	s.DeviceRepo = cfg.DeviceRepo

	s.setupRoutes()

	return s
}
