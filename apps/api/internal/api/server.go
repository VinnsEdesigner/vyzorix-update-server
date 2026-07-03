package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/admin"
	authhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	cmdhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/command"
	dashboardhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/dashboard"
	devicehandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/device"
	updaterhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updater"
	websockethandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
        updateshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/wire"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
        updatesapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
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
	RateLimiter    *middleware.RateLimiter
	Hub            *hub.Hub
	AuthService    *auth.AuthService
	AuthLimiter    *middleware.RateLimiter
	IPIntelligence *middleware.IPIntelligence
	Log            *slog.Logger
	SessionManager *infraauth.SessionManager
	GoogleVerifier *infraauth.GoogleTokenVerifier
	EmailService   *emailService.Service
	CommandService *command.Service
	ClientService  *client.Service
	DB             *storage.SQLite
	Lockout        *middleware.Lockout
	DeviceService  *device.Service
	Metrics        *infraMetrics.Metrics
	AuditLogger    *audit.Logger
	Config         config.Config
	UpdatesService *updatesapp.Service
}

// Server is the main API server.
type Server struct {
	encryptKeyFn            func(clientID string) ([]byte, bool)
	authHandlers            *authhandlers.AllHandlers
	hub                     *hub.Hub
	engine                  *gin.Engine
	log                     *slog.Logger
	deviceStatusHandler     *devicehandlers.StatusHandler
	sessionManager          *infraauth.SessionManager
	rateLimiter             *middleware.RateLimiter
	authLimiter             *middleware.RateLimiter
	cookieAuth              *middleware.CookieAuth
	signatureVerifier       *middleware.SignatureVerifier
	lockout                 *middleware.Lockout
	csrfProtector           *middleware.CSRFProtector
	turnstileVerifier       *middleware.TurnstileVerifier
	revocationList          *infraauth.RevocationList
	ipIntelligence          *middleware.IPIntelligence
	hmacVerifier            *cryptohmac.Verifier
	mwFactory               *middleware.MiddlewareFactory
	db                      *storage.SQLite
	dashboardRateLimiter    *middleware.DashboardRateLimiterMiddleware
	AuditLogger             *audit.Logger
	deviceRegisterHandler   *devicehandlers.RegisterHandler
	deviceUpdaterHandler    *devicehandlers.UpdaterHandler
	deviceListHandler       *devicehandlers.ListHandler
	commandHandler          *cmdhandlers.ExecuteHandler
	streamHandler           *websockethandlers.StreamHandler
	telemetryHistoryHandler *handlers.TelemetryHistoryHandler
	connectionStatusHandler *handlers.ConnectionStatusHandler
	adminClientsHandler     *admin.ClientsHandler
	updaterHandler          *updaterhandlers.Handler
	metricsHandler          *infraMetrics.MetricsHandler
	commandHistoryHandler   *cmdhandlers.HistoryHandler
	deviceLogsHandler       *devicehandlers.LogsHandler
	deviceMetricsHandler    *devicehandlers.MetricsHandler
	deviceTelemetryHandler  *devicehandlers.TelemetryHandler
	dashboardStatsHandler   *dashboardhandlers.StatsHandler
        updatesHandler          *updateshandlers.UpdatesHandler
	config                  config.Config
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
		metricsSvc = appmetrics.NewService(metricsRepo)
	}

	if cfg.CommandService != nil && cfg.DeviceService != nil {
		dashboardSvc = dashboard.NewService(cfg.DeviceService.DeviceRepo(), cfg.CommandService.CommandRepo())
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
        if cfg.UpdatesService != nil {
                updatesRateLimiters := middleware.NewUpdatesRateLimiterMiddleware(middleware.DefaultUpdatesRateLimits())
                s.updatesHandler = updateshandlers.NewUpdatesHandler(cfg.UpdatesService, updatesRateLimiters, cfg.AuditLogger)
        }
}

// Handlers are defined in server_handlers.go
// Routes are defined in server_routes.go
// GraphQL is defined in server_graphql.go
