package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/admin"
	adminkeyshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/admin"
	authapikeyshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	authhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	cmdhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/command"
	dashboardhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/dashboard"
	devicehandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/device"
	diagnosticshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/diagnostics"
	inboxhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/inbox"
	updaterhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updater"
	updateshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updates"
	websockethandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/wire"
	keys "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	diagnosticsapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/inbox"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	updatesapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
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
	FCMNotifier    fcm.Notifier
	OperatorRepo   operator.Repository
	OAuthStateRepo authhandlers.OAuthStateProvider
	EmailService   *emailService.Service
	ClientService  *client.Service
	AuthLimiter    *middleware.RateLimiter
	IPIntelligence *middleware.IPIntelligence
	Log            *slog.Logger
	SessionManager *infraauth.SessionManager
	GoogleVerifier *infraauth.GoogleTokenVerifier
	Hub            *hub.Hub
	CommandService *command.Service
	AuthService    *auth.AuthService
	DB             *storage.SQLite
	Lockout        *middleware.Lockout
	DeviceService  *device.Service
	Metrics        *infraMetrics.Metrics
	AuditLogger    *audit.Logger
	RateLimiter    *middleware.RateLimiter
	UpdatesService *updatesapp.Service
	PushService    *updatesapp.PushService
	Config         config.Config
	APIKeyService  *keys.Service
}

// Server is the main API server.
type Server struct {
	AuditLogger                *audit.Logger
	revocationList             *infraauth.RevocationList
	log                        *slog.Logger
	apiKeyAuth                 *middleware.TenantAPIKeyAuth
	authHandlers               *authhandlers.AllHandlers
	hub                        *hub.Hub
	deviceStatusHandler        *devicehandlers.StatusHandler
	sessionManager             *infraauth.SessionManager
	rateLimiter                *middleware.RateLimiter
	authLimiter                *middleware.RateLimiter
	cookieAuth                 *middleware.CookieAuth
	signatureVerifier          *middleware.SignatureVerifier
	lockout                    *middleware.Lockout
	csrfProtector              *middleware.CSRFProtector
	turnstileVerifier          *middleware.TurnstileVerifier
	deviceRegisterHandler      *devicehandlers.RegisterHandler
	ipIntelligence             *middleware.IPIntelligence
	hmacVerifier               *cryptohmac.Verifier
	mwFactory                  *middleware.MiddlewareFactory
	db                         *storage.SQLite
	dashboardRateLimiter       *middleware.DashboardRateLimiterMiddleware
	deviceRegRateLimiter       *middleware.DeviceRegistrationRateLimiterMiddleware
	engine                     *gin.Engine
	encryptKeyFn               func(clientID string) ([]byte, bool)
	dashboardStatsHandler      *dashboardhandlers.StatsHandler
	deviceListHandler          *devicehandlers.ListHandler
	devicesHandler             *devicehandlers.DevicesHandler
	commandHandler             *cmdhandlers.ExecuteHandler
	streamHandler              *websockethandlers.StreamHandler
	telemetryHistoryHandler    *handlers.TelemetryHistoryHandler
	connectionStatusHandler    *handlers.ConnectionStatusHandler
	adminClientsHandler        *admin.ClientsHandler
	updaterHandler             *updaterhandlers.Handler
	metricsHandler             *infraMetrics.MetricsHandler
	commandHistoryHandler      *cmdhandlers.HistoryHandler
	deviceLogsHandler          *devicehandlers.LogsHandler
	deviceMetricsHandler       *devicehandlers.MetricsHandler
	deviceTelemetryHandler     *devicehandlers.TelemetryHandler
	deviceUpdaterHandler       *devicehandlers.UpdaterHandler
	updatesHandler             *updateshandlers.UpdatesHandler
	inboxHandler               *inboxhandlers.Handler
	deviceConfirmHandler       *devicehandlers.ConfirmHandler
	diagnosticsHandler         *diagnosticshandlers.Handler
	diagnosticsInspectHandler  *diagnosticshandlers.InspectHandler
	diagnosticsTimelineHandler *diagnosticshandlers.TimelineHandler
	config                     config.Config
	apiKeysHandler             *authapikeyshandlers.Handler
	superAdminAPIKeys          *adminkeyshandlers.SuperAdminHandler
}

// NewServer creates a new API server with wired-up dependencies.
func NewServer(cfg *ServerConfig) *Server {
	if cfg.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.GinPanicRecovery(cfg.Log))

	// Wire middleware using wire package
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
		RateLimitPerMin:  100,
		AuthRateLimitMin: 5,
	})

	s := &Server{
		engine:            engine,
		mwFactory:         mwSet.Factory,
		rateLimiter:       cfg.RateLimiter,
		authLimiter:       cfg.AuthLimiter,
		config:            cfg.Config,
		log:               cfg.Log,
		cookieAuth:        middleware.NewCookieAuth(cfg.SessionManager, cfg.AuthService),
		signatureVerifier: mwSet.SignatureVerifier,
		lockout:           mwSet.Lockout,
		csrfProtector:     mwSet.CSRFProtector,
		turnstileVerifier: mwSet.TurnstileVerifier,
		revocationList:    mwSet.RevocationList,
		ipIntelligence:    mwSet.IPIntelligence,
		hmacVerifier:      mwSet.HmacVerifier,
		encryptKeyFn:      mwSet.EncryptKeyFn,
		sessionManager:    cfg.SessionManager,
		db:                cfg.DB,
		hub:               cfg.Hub,
	}

	// Create presenter and wire handlers
	presenter := response.NewPresenter(cfg.AuthService, cfg.AuditLogger, cfg.IPIntelligence)
	s.wireHandlers(cfg, presenter, mwSet)

	// Initialize API keys handler
	if cfg.APIKeyService != nil {
		s.apiKeysHandler = authapikeyshandlers.NewHandler(cfg.APIKeyService)
		s.superAdminAPIKeys = adminkeyshandlers.NewSuperAdminHandler(cfg.APIKeyService)
	}

	// Start Hub if available
	if cfg.Hub != nil {
		go cfg.Hub.Run(context.Background())
	}

	// Initialize metrics handler
	if cfg.Metrics != nil {
		s.metricsHandler = infraMetrics.NewMetricsHandler(cfg.Metrics)
		s.engine.Use(infraMetrics.Middleware(cfg.Metrics))
	}

	s.setupRoutes()

	// Set audit logger
	s.AuditLogger = cfg.AuditLogger

	return s
}

// wireHandlers creates and assigns all handler instances.
func (s *Server) wireHandlers(cfg *ServerConfig, presenter *response.Presenter, mwSet *wire.MiddlewareSet) {
	// Auth handlers
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

	// Device handlers
	s.deviceRegisterHandler = devicehandlers.NewRegisterHandler(cfg.DeviceService)
	s.deviceStatusHandler = devicehandlers.NewStatusHandler(cfg.DeviceService)
	s.deviceUpdaterHandler = devicehandlers.NewUpdaterHandler(cfg.DeviceService)
	s.deviceListHandler = devicehandlers.NewListHandler(cfg.DeviceService, cfg.Hub)

	// Command handler
	s.commandHandler = cmdhandlers.NewExecuteHandler(cfg.CommandService, cfg.DeviceService, cfg.Hub, cfg.FCMNotifier)

	// WebSocket handler
	s.streamHandler = websockethandlers.NewStreamHandler(cfg.Log, cfg.Config, cfg.Hub, *mwSet.HmacVerifier, cfg.AuditLogger)

	// Telemetry history handler
	s.telemetryHistoryHandler = handlers.NewTelemetryHistoryHandler(
		cfg.Log,
		storage.NewTelemetryRepository(cfg.DB.DB()),
		nil,
	)

	// Connection status handler
	s.connectionStatusHandler = handlers.NewConnectionStatusHandler(cfg.Log, cfg.Hub)

	// Admin handlers
	s.adminClientsHandler = admin.NewClientsHandler(cfg.ClientService)

	// Updater handlers
	s.updaterHandler = updaterhandlers.NewHandler(cfg.Log, cfg.Config)

	// Dashboard command handlers - wire up new handlers
	s.wireDashboardHandlers(cfg)

	// Initialize dashboard rate limiter
	s.dashboardRateLimiter = middleware.NewDashboardRateLimiterMiddleware(nil)
}

// Routes returns the Gin engine for serving.
func (s *Server) Routes() http.Handler {
	return s.engine
}

// wireDashboardHandlers creates and assigns dashboard command handler instances.
func (s *Server) wireDashboardHandlers(cfg *ServerConfig) {
	// Create repositories
	var logsRepo *storage.LogsRepository
	var metricsRepo *storage.MetricsRepository

	if cfg.DB != nil {
		logsRepo = storage.NewLogsRepository(cfg.DB.DB())
		metricsRepo = storage.NewMetricsRepository(cfg.DB.DB())
	}

	// Create services
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
		metricsSvc = appmetrics.NewService(metricsRepo, cfg.OperatorRepo)
	}

	if cfg.CommandService != nil && cfg.DeviceService != nil && logsRepo != nil {
		dashboardSvc = dashboard.NewService(cfg.DeviceService.DeviceRepo(), cfg.CommandService.CommandRepo(), logsRepo)
	}

	// Create handlers
	if historySvc != nil {
		s.commandHistoryHandler = cmdhandlers.NewHistoryHandler(historySvc, cfg.Log)
	}

	if logsSvc != nil && cfg.DeviceService != nil {
		s.deviceLogsHandler = devicehandlers.NewLogsHandler(logsSvc, cfg.DeviceService.DeviceRepo(), cfg.Log)
	}

	if metricsSvc != nil && cfg.DeviceService != nil {
		s.deviceMetricsHandler = devicehandlers.NewMetricsHandler(metricsSvc, cfg.DeviceService.DeviceRepo(), cfg.Log)
		s.deviceTelemetryHandler = devicehandlers.NewTelemetryHandler(metricsSvc, cfg.DeviceService.DeviceRepo(), cfg.Log)
	}

	if dashboardSvc != nil {
		s.dashboardStatsHandler = dashboardhandlers.NewStatsHandler(dashboardSvc, cfg.Log)
	}

	// Updates handler
	if cfg.UpdatesService != nil {
		updatesRateLimiters := middleware.NewUpdatesRateLimiterMiddleware(middleware.DefaultUpdatesRateLimits())
		s.updatesHandler = updateshandlers.NewUpdatesHandler(cfg.UpdatesService, cfg.PushService, updatesRateLimiters, cfg.AuditLogger, cfg.Config.GitHubWebhookSecret)
	}

	// Inbox handler
	s.wireInboxHandler(cfg)

	// Device confirm handler
	s.wireConfirmHandler(cfg)

	// Diagnostics handler
	s.wireDiagnosticsHandler(cfg)
}

// wireInboxHandler creates and assigns the inbox handler.
func (s *Server) wireInboxHandler(cfg *ServerConfig) {
	if cfg.DB == nil || cfg.DeviceService == nil {
		return
	}

	// Create inbox repository
	inboxRepo := storage.NewInboxRepository(cfg.DB.DB())
	regLogRepo := storage.NewRegistrationLogRepository(cfg.DB.DB())

	// Create inbox service with FCM notifier (SafeNotifier for graceful degradation)
	// and device service for creating devices on approval and device lookup
	var fcmNotifier inbox.FCMNotifier
	if cfg.FCMNotifier != nil {
		fcmNotifier = &fcm.SafeNotifier{Notifier: cfg.FCMNotifier}
	}
	inboxService := inbox.NewService(inboxRepo, regLogRepo, cfg.DeviceService, cfg.DeviceService, fcmNotifier, s.log)

	// Enable ACID transactions for enterprise production
	inboxService.WithTxManager(cfg.DB)

	// Create device registration rate limiter
	s.deviceRegRateLimiter = middleware.NewDeviceRegistrationRateLimiterMiddleware(nil)

	// Create handler
	s.inboxHandler = inboxhandlers.NewHandler(inboxService)
}

// wireConfirmHandler creates and assigns the device confirm handler.
func (s *Server) wireConfirmHandler(cfg *ServerConfig) {
	if cfg.DeviceService == nil {
		return
	}

	// Create confirm handler
	s.deviceConfirmHandler = devicehandlers.NewConfirmHandler(cfg.DeviceService)
}

// wireDiagnosticsHandler creates and assigns the diagnostics handlers.
func (s *Server) wireDiagnosticsHandler(cfg *ServerConfig) {
	if cfg.DB == nil || cfg.DeviceService == nil {
		return
	}

	// Create diagnostics repository
	diagnosticsRepo := storage.NewDiagnosticsRepository(cfg.DB.DB())

	// Create diagnostics service
	diagnosticsService := diagnosticsapp.NewService(diagnosticsRepo, cfg.DeviceService.DeviceRepo(), s.hub, cfg.Config.DiagnosticsConfig)

	// Get rate limiters
	var inspectLimit, timelineLimit func(c *gin.Context)
	if s.dashboardRateLimiter != nil {
		inspectLimit = s.dashboardRateLimiter.DeviceInspectLimit()
		timelineLimit = s.dashboardRateLimiter.DeviceTimelineLimit()
	}

	// Create inspect handler
	s.diagnosticsInspectHandler = diagnosticshandlers.NewInspectHandler(diagnosticsService, inspectLimit)

	// Create timeline handler
	s.diagnosticsTimelineHandler = diagnosticshandlers.NewTimelineHandler(diagnosticsService, timelineLimit)

	// Create combined handler for backwards compatibility
	s.diagnosticsHandler = &diagnosticshandlers.Handler{
		InspectHandler:  s.diagnosticsInspectHandler,
		TimelineHandler: s.diagnosticsTimelineHandler,
	}
}

// Handlers are defined in server_handlers.go
// Routes are defined in server_routes.go
// GraphQL is defined in server_graphql.go

// ServerConfigWithDeps is the config for NewServerWithDeps using pre-wired dependencies.
type ServerConfigWithDeps struct {
	Log            *slog.Logger
	DB             *storage.SQLite
	Engine         *gin.Engine
	Middleware     *wire.MiddlewareSet
	HandlerSet     *wire.HandlerSet
	SessionManager *infraauth.SessionManager
	Hub            *hub.Hub
	AuditLogger    *audit.Logger
	UpdatesService *updatesapp.Service
	Config         config.Config
	APIKeyService  *keys.Service
}

// NewServerWithDeps creates a Server using pre-wired dependencies from wire.
func NewServerWithDeps(cfg *ServerConfigWithDeps) *Server {
	if cfg.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{
		engine:            cfg.Engine,
		mwFactory:         cfg.Middleware.Factory,
		rateLimiter:       cfg.Middleware.RateLimiter,
		authLimiter:       cfg.Middleware.AuthLimiter,
		config:            cfg.Config,
		log:               cfg.Log,
		cookieAuth:        cfg.Middleware.CookieAuth,
		lockout:           cfg.Middleware.Lockout,
		csrfProtector:     cfg.Middleware.CSRFProtector,
		turnstileVerifier: cfg.Middleware.TurnstileVerifier,
		revocationList:    cfg.Middleware.RevocationList,
		ipIntelligence:    cfg.Middleware.IPIntelligence,
		hmacVerifier:      cfg.Middleware.HmacVerifier,
		sessionManager:    cfg.SessionManager,
		db:                cfg.DB,
		hub:               cfg.Hub,
		AuditLogger:       cfg.AuditLogger,
		apiKeyAuth:        middleware.NewTenantAPIKeyAuth(cfg.APIKeyService, cfg.Config.APIKeyPrefix),
	}

	// Wire handlers from HandlerSet
	s.authHandlers = cfg.HandlerSet.Auth
	s.deviceRegisterHandler = cfg.HandlerSet.DeviceRegister
	s.deviceStatusHandler = cfg.HandlerSet.DeviceStatus
	s.deviceUpdaterHandler = cfg.HandlerSet.DeviceUpdater
	s.deviceListHandler = cfg.HandlerSet.DeviceList
	s.devicesHandler = cfg.HandlerSet.Devices
	s.commandHandler = cfg.HandlerSet.Command
	s.streamHandler = cfg.HandlerSet.Stream
	s.telemetryHistoryHandler = cfg.HandlerSet.TelemetryHistory
	s.connectionStatusHandler = cfg.HandlerSet.ConnectionStatus
	s.adminClientsHandler = cfg.HandlerSet.AdminClients
	s.updaterHandler = cfg.HandlerSet.Updater
	s.updatesHandler = cfg.HandlerSet.Updates

	s.setupRoutes()

	return s
}
