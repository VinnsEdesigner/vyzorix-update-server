// Package wire provides dependency injection using google/wire.

package wire

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/admin"
	authhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	cmdhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/command"
	devicehandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/device"
	updaterhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updater"
	updateshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updates"
	websockethandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	eventapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/event"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	appnotification "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/notification"
	updatesapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/github"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/logging"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	infranotification "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/notification"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	passwordpkg "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/password"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// ProvideConfig returns the application config.
func ProvideConfig() (config.Config, error) {
	return config.Load()
}

// ProvideLogger returns the application logger.
func ProvideLogger() *slog.Logger {
	return logging.NewFromEnv()
}

// ProvideSQLite opens the database connection.
func ProvideSQLite(cfg config.Config) (*storage.SQLite, error) {
	sqliteCfg := storage.DefaultConfig(cfg.DatabaseURL)
	return storage.Open(sqliteCfg)
}

// ProvideDB returns the underlying *sql.DB from SQLite.
func ProvideDB(s *storage.SQLite) *sql.DB {
	return s.DB()
}

// ProvideAuditLogger creates the audit logger.
func ProvideAuditLogger(db *sql.DB, log *slog.Logger) *audit.Logger {
	auditRepo := audit.NewRepository(db)
	return audit.NewLogger(auditRepo, log, audit.DefaultLoggerConfig())
}

// ProvidePasswordHasher creates the password hasher.
func ProvidePasswordHasher() *passwordpkg.Argon2idHasher {
	return passwordpkg.NewArgon2idHasher()
}

// ProvideSessionManager creates the session manager.
func ProvideSessionManager(cfg config.Config) *infraauth.SessionManager {
	return infraauth.NewSessionManager(cfg.SessionSecret)
}

// ProvideGoogleVerifier creates the Google token verifier.
func ProvideGoogleVerifier(cfg config.Config) *infraauth.GoogleTokenVerifier {
	return infraauth.NewGoogleTokenVerifier(cfg.GoogleOAuthClientID)
}

// ProvideOperatorRepository creates the operator repository.
func ProvideOperatorRepository(db *sql.DB) *storage.OperatorRepository {
	return storage.NewOperatorRepository(db)
}

// ProvideDeviceRepository creates the device repository.
func ProvideDeviceRepository(db *sql.DB) *storage.DeviceRepository {
	return storage.NewDeviceRepository(db)
}

// ProvideCommandRepository creates the command repository.
func ProvideCommandRepository(db *sql.DB) *storage.CommandRepository {
	return storage.NewCommandRepository(db)
}

// ProvideSessionRepository creates the session repository.
func ProvideSessionRepository(db *sql.DB) *storage.SessionRepository {
	return storage.NewSessionRepository(db)
}

// ProvideClientRepository creates the client repository.
func ProvideClientRepository(db *sql.DB) *storage.ClientRepository {
	return storage.NewClientRepository(db)
}

// ProvideTelemetryRepository creates the telemetry repository.
func ProvideTelemetryRepository(db *sql.DB) *storage.TelemetryRepository {
	return storage.NewTelemetryRepository(db)
}

// ProvideEventRepository creates the event repository.
func ProvideEventRepository(db *sql.DB) *storage.EventRepository {
	return storage.NewEventRepository(db)
}

// ProvideUpdatesStorage creates the updates storage.
func ProvideUpdatesStorage(db *sql.DB) *storage.UpdatesStorage {
	return storage.NewUpdatesStorage(db)
}

// ProvideEventProcessor creates the event processor and wires it to the hub.
func ProvideEventProcessor(
	eventRepo *storage.EventRepository,
	deviceRepo *storage.DeviceRepository,
	operatorRepo *storage.OperatorRepository,
	hubResult *HubResult,
	log *slog.Logger,
) *eventapp.Processor {
	// Create broadcaster first
	broadcaster := eventapp.NewBroadcaster(log)
	hubResult.Hub.SetDashboardBroadcaster(broadcaster)

	// Create processor with broadcaster
	processor := eventapp.NewProcessor(eventRepo, deviceRepo, broadcaster, log)

	// Wire operator repository for per-operator threshold fetching
	processor.SetOperatorRepo(operatorRepo)

	hubResult.Hub.SetEventProcessor(processor)

	return processor
}

// ProvideNotificationService creates the notification service.
func ProvideNotificationService(
	operatorRepo *storage.OperatorRepository,
	emailSvc *emailService.Service,
	webhookClient *webhook.Client,
	auditRepo *infranotification.Repository,
	log *slog.Logger,
) *appnotification.Service {
	return appnotification.NewService(operatorRepo, emailSvc, webhookClient, auditRepo, log)
}

// WireNotificationServiceToProcessor wires the notification service to the event processor.
func WireNotificationServiceToProcessor(
	processor *eventapp.Processor,
	notificationSvc *appnotification.Service,
) {
	processor.SetNotificationService(notificationSvc)
}

// ProvideEmailVerificationRepository creates the email verification repository.
func ProvideEmailVerificationRepository(db *sql.DB) *storage.EmailVerificationRepository {
	return storage.NewEmailVerificationRepository(db)
}

// ProvidePasswordResetRepository creates the password reset repository.
func ProvidePasswordResetRepository(db *sql.DB) *storage.PasswordResetRepository {
	return storage.NewPasswordResetRepository(db)
}

// ProvideAuthService creates the auth service.
func ProvideAuthService(
	operatorRepo *storage.OperatorRepository,
	sessionRepo *storage.SessionRepository,
	emailVerifyRepo *storage.EmailVerificationRepository,
	passwordResetRepo *storage.PasswordResetRepository,
	hasher *passwordpkg.Argon2idHasher,
) *auth.AuthService {
	sessionTTL := 7 * 24 * time.Hour
	return auth.NewAuthService(
		operatorRepo,
		sessionRepo,
		emailVerifyRepo,
		passwordResetRepo,
		hasher,
		sessionTTL,
	)
}

// ProvideDeviceService creates the device service.
func ProvideDeviceService(deviceRepo *storage.DeviceRepository, operatorRepo *storage.OperatorRepository, log *slog.Logger) *device.Service {
	return device.NewService(deviceRepo, operatorRepo, log)
}

// ProvideClientService creates the client service.
func ProvideClientService(clientRepo *storage.ClientRepository) *client.Service {
	return client.NewService(clientRepo)
}

// ProvideCommandService creates the command service.
func ProvideCommandService(commandRepo *storage.CommandRepository, deviceRepo *storage.DeviceRepository) *cmdapp.Service {
	return cmdapp.NewService(commandRepo, deviceRepo)
}

// ProvideEmailService creates the email service.
func ProvideEmailService() *emailService.Service {
	return emailService.NewService()
}

// ProvideMetrics creates the metrics service.
func ProvideMetrics() *metrics.Metrics {
	return metrics.New()
}

// HubResult holds the WebSocket hub components.
type HubResult struct {
	Hub        *hub.Hub
	MessageQueue *hub.MessageQueue
	RateLimiter *hub.RateLimiter
}

// ProvideWebSocketHub creates the WebSocket hub with message queue and rate limiter.
func ProvideWebSocketHub(
	log *slog.Logger,
	deviceRepo *storage.DeviceRepository,
	telemetryRepo *storage.TelemetryRepository,
	db *sql.DB,
) *HubResult {
	hubConfig := &hub.HubConfig{
		MessageQueue: hub.DefaultMessageQueueConfig(),
		RateLimiter:  hub.DefaultRateLimiterConfig(),
	}

	h := hub.New(log, deviceRepo, telemetryRepo, nil, hubConfig)

	mq := hub.NewMessageQueue(log, db, hub.DefaultMessageQueueConfig())
	mq.Start(context.Background())
	h.SetMessageQueue(mq)

	rl := hub.NewRateLimiter(log, hub.DefaultRateLimiterConfig())
	rl.StartCleanup(context.Background())
	h.SetRateLimiter(rl)

	return &HubResult{Hub: h, MessageQueue: mq, RateLimiter: rl}
}

// ProvideFCMNotifier creates the FCM notifier.
func ProvideFCMNotifier(log *slog.Logger, cfg config.Config) fcm.Notifier {
	if cfg.FirebaseCreds == "" {
		return nil
	}
	fcmClient, err := fcm.Init(log, cfg.FirebaseCreds)
	if err != nil {
		return nil
	}
	return fcm.NewEnhancedNotifier(fcmClient, fcm.DefaultFCMConfig())
}

// ProvideMiddlewareFactory creates the middleware factory.
func ProvideMiddlewareFactory(
	log *slog.Logger,
	sessionManager *infraauth.SessionManager,
	authService *auth.AuthService,
	clientService *client.Service,
	cfg config.Config,
) *middleware.MiddlewareFactory {
	return middleware.NewMiddlewareFactory(
		log,
		sessionManager,
		authService,
		clientService,
		middleware.FactoryConfig{
			AllowedOrigins:   cfg.AllowedOrigins,
			EnforceHMAC:      cfg.EnforceHMAC,
			HMACWindow:       cfg.HMACWindow,
			PublicDir:        cfg.PublicDir,
			JWTSecret:        cfg.JWTSecret,
			RateLimitPerMin:  100,
			AuthRateLimitMin: 5,
		},
	)
}

// ProvideRateLimiter creates the rate limiter.
func ProvideRateLimiter() *middleware.RateLimiter {
	return middleware.NewRateLimiter(100, time.Minute)
}

// ProvideLockout creates the lockout middleware.
func ProvideLockout() *middleware.Lockout {
	return middleware.NewLockout(middleware.LoadLockoutConfig())
}

// ProvideIPIntelligence creates the IP intelligence middleware.
func ProvideIPIntelligence(factory *middleware.MiddlewareFactory) *middleware.IPIntelligence {
	return factory.IPIntelligence()
}

// ProvideHmacVerifier creates the HMAC verifier.
func ProvideHmacVerifier(factory *middleware.MiddlewareFactory) *cryptohmac.Verifier {
	return factory.GetHmacVerifier()
}

// ProvideCookieAuth creates the cookie auth middleware.
func ProvideCookieAuth(sessionManager *infraauth.SessionManager, authService *auth.AuthService) *middleware.CookieAuth {
	return middleware.NewCookieAuth(sessionManager, authService)
}

// ProvidePresenter creates the response presenter.
func ProvidePresenter(
	authService *auth.AuthService,
	auditLogger *audit.Logger,
	ipIntelligence *middleware.IPIntelligence,
) *response.Presenter {
	return response.NewPresenter(authService, auditLogger, ipIntelligence)
}

// ProvideUpdatesService creates the updates service with all sub-services.
func ProvideUpdatesService(
	updatesStorage *storage.UpdatesStorage,
	deviceService *device.Service,
	hubResult *HubResult,
	fcmNotifier fcm.Notifier,
	commandService *cmdapp.Service,
	log *slog.Logger,
	cfg config.Config,
) *updatesapp.Service {
	versionsStatusSvc := updatesapp.NewVersionsStatusService(updatesStorage)
	versionsListSvc := updatesapp.NewVersionsListService(updatesStorage)
	changelogSvc := updatesapp.NewChangelogService(updatesStorage)
	exportSvc := updatesapp.NewExportService(updatesStorage)
	historySvc := updatesapp.NewHistoryService(updatesStorage)

	var githubSyncSvc *github.SyncService
	if cfg.GitHubReleaseToken != "" && cfg.GitHubReleaseRepo != "" {
		githubClient := github.NewClient("VinnsEdesigner", cfg.GitHubReleaseRepo, cfg.GitHubReleaseToken)
		githubSyncSvc = github.NewSyncService(githubClient, updatesStorage, log)
	}
	syncSvc := updatesapp.NewSyncService(updatesStorage, githubSyncSvc)
	pushSvc := updatesapp.NewPushService(updatesStorage, deviceService, hubResult.Hub, fcmNotifier, commandService, log)

	return updatesapp.NewService(
		updatesStorage,
		versionsStatusSvc,
		versionsListSvc,
		changelogSvc,
		exportSvc,
		pushSvc,
		historySvc,
		syncSvc,
	)
}

// ProvideUpdatesHandler creates the updates handler.
func ProvideUpdatesHandler(
	updatesService *updatesapp.Service,
	auditLogger *audit.Logger,
	cfg config.Config,
) *updateshandlers.UpdatesHandler {
	updatesRateLimiters := middleware.NewUpdatesRateLimiterMiddleware(nil)
	return updateshandlers.NewUpdatesHandler(
		updatesService,
		updatesRateLimiters,
		auditLogger,
		cfg.GitHubWebhookSecret,
	)
}

// ProvideGinEngine creates the Gin engine.
func ProvideGinEngine(cfg config.Config, log *slog.Logger) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(middleware.GinPanicRecovery(log))
	return engine
}

// ProvideHandlerSet creates all handlers.
func ProvideHandlerSet(
	cfg config.Config,
	authService *auth.AuthService,
	sessionManager *infraauth.SessionManager,
	googleVerifier *infraauth.GoogleTokenVerifier,
	clientService *client.Service,
	emailService *emailService.Service,
	lockout *middleware.Lockout,
	operatorRepo *storage.OperatorRepository,
	auditLogger *audit.Logger,
	ipIntelligence *middleware.IPIntelligence,
	presenter *response.Presenter,
	deviceService *device.Service,
	hubResult *HubResult,
	commandService *cmdapp.Service,
	fcmNotifier fcm.Notifier,
	log *slog.Logger,
	hmacVerifier *cryptohmac.Verifier,
	db *sql.DB,
	updatesHandler *updateshandlers.UpdatesHandler,
) *HandlerSet {
	hs := &HandlerSet{}

	hs.Auth = authhandlers.NewAllHandlers(&authhandlers.Dependencies{
		AuthService:    authService,
		SessionManager: sessionManager,
		Config:         cfg,
		GoogleVerifier: googleVerifier,
		ClientService:  clientService,
		EmailService:   emailService,
		Lockout:        lockout,
		OperatorRepo:   operatorRepo,
		AuditLogger:    auditLogger,
		IPIntelligence: ipIntelligence,
		Presenter:      presenter,
	})

	hs.DeviceRegister = devicehandlers.NewRegisterHandler(deviceService)
	hs.DeviceStatus = devicehandlers.NewStatusHandler(deviceService)
	hs.DeviceUpdater = devicehandlers.NewUpdaterHandler(deviceService)
	hs.DeviceList = devicehandlers.NewListHandler(deviceService, hubResult.Hub)
	hs.Command = cmdhandlers.NewExecuteHandler(commandService, deviceService, hubResult.Hub, fcmNotifier)
	hs.Stream = websockethandlers.NewStreamHandler(log, cfg, hubResult.Hub, *hmacVerifier, auditLogger)
	hs.TelemetryHistory = handlers.NewTelemetryHistoryHandler(log, storage.NewTelemetryRepository(db), nil)
	hs.ConnectionStatus = handlers.NewConnectionStatusHandler(log, hubResult.Hub)
	hs.AdminClients = admin.NewClientsHandler(clientService)
	hs.Updater = updaterhandlers.NewHandler(log, cfg)
	hs.Updates = updatesHandler

	return hs
}

// ProvideMiddlewareSet creates all middleware.
func ProvideMiddlewareSet(
	factory *middleware.MiddlewareFactory,
	cookieAuth *middleware.CookieAuth,
	lockout *middleware.Lockout,
	ipIntelligence *middleware.IPIntelligence,
	hmacVerifier *cryptohmac.Verifier,
	rateLimiter *middleware.RateLimiter,
) *MiddlewareSet {
	// Create a separate rate limiter for auth with different limits
	authLimiter := middleware.NewRateLimiter(5, time.Minute)
	return &MiddlewareSet{
		Factory:          factory,
		CookieAuth:       cookieAuth,
		Lockout:          lockout,
		IPIntelligence:   ipIntelligence,
		HmacVerifier:     hmacVerifier,
		RateLimiter:      rateLimiter,
		AuthLimiter:      authLimiter,
	}
}

// WireInjector is the provider set for dependency injection.
// Note: config.Config is passed as an argument to Injector, not provided here.
var WireInjector = wire.NewSet(
	ProvideLogger,
	ProvideSQLite,
	ProvideDB,
	ProvideAuditLogger,
	ProvidePasswordHasher,
	ProvideSessionManager,
	ProvideGoogleVerifier,
	ProvideOperatorRepository,
	ProvideDeviceRepository,
	ProvideCommandRepository,
	ProvideSessionRepository,
	ProvideClientRepository,
	ProvideTelemetryRepository,
	ProvideEventRepository,
	ProvideUpdatesStorage,
	ProvideEventProcessor,
	ProvideEmailVerificationRepository,
	ProvidePasswordResetRepository,
	ProvideAuthService,
	ProvideDeviceService,
	ProvideClientService,
	ProvideCommandService,
	ProvideEmailService,
	ProvideMetrics,
	ProvideWebSocketHub,
	ProvideFCMNotifier,
	ProvideMiddlewareFactory,
	ProvideRateLimiter,
	ProvideLockout,
	ProvideIPIntelligence,
	ProvideHmacVerifier,
	ProvideCookieAuth,
	ProvidePresenter,
	ProvideUpdatesService,
	ProvideUpdatesHandler,
	ProvideGinEngine,
	ProvideHandlerSet,
	ProvideMiddlewareSet,
	ProvideServerDependencies,
	ProvideServerResult,
	ProvideServer,
	wire.Bind(new(operator.Repository), new(*storage.OperatorRepository)),
)

// ProvideServerDependencies creates ServerDependencies from all required components.
func ProvideServerDependencies(
	cfg config.Config,
	log *slog.Logger,
	sqlite *storage.SQLite,
	auditLogger *audit.Logger,
	sessionManager *infraauth.SessionManager,
	googleVerifier *infraauth.GoogleTokenVerifier,
	operatorRepo *storage.OperatorRepository,
	deviceRepo *storage.DeviceRepository,
	commandRepo *storage.CommandRepository,
	sessionRepo *storage.SessionRepository,
	clientRepo *storage.ClientRepository,
	telemetryRepo *storage.TelemetryRepository,
	updatesStorage *storage.UpdatesStorage,
	emailVerifyRepo *storage.EmailVerificationRepository,
	passwordResetRepo *storage.PasswordResetRepository,
	hasher *passwordpkg.Argon2idHasher,
	authService *auth.AuthService,
	deviceService *device.Service,
	clientService *client.Service,
	commandService *cmdapp.Service,
	emailService *emailService.Service,
	metrics *metrics.Metrics,
	hubResult *HubResult,
	fcmNotifier fcm.Notifier,
	factory *middleware.MiddlewareFactory,
	rateLimiter *middleware.RateLimiter,
	lockout *middleware.Lockout,
	ipIntelligence *middleware.IPIntelligence,
	updatesService *updatesapp.Service,
) *ServerDependencies {
	return &ServerDependencies{
		FCMNotifier:     fcmNotifier,
		OperatorRepo:    operatorRepo,
		RateLimiter:     rateLimiter,
		Hub:             hubResult.Hub,
		AuthService:     authService,
		AuthLimiter:     middleware.NewRateLimiter(5, time.Minute),
		IPIntelligence:  ipIntelligence,
		Log:             log,
		SessionManager:  sessionManager,
		GoogleVerifier:  googleVerifier,
		EmailService:    emailService,
		CommandService:  commandService,
		ClientService:   clientService,
		DB:              sqlite,
		Lockout:         lockout,
		DeviceService:   deviceService,
		Metrics:         metrics,
		AuditLogger:     auditLogger,
		Config:          cfg,
		UpdatesStorage:  updatesStorage,
		UpdatesService:  updatesService,
		TelemetryRepo:   telemetryRepo,
	}
}

// ProvideServerResult provides the wired server result by calling WireServer.
func ProvideServerResult(deps *ServerDependencies) *ServerResult {
	return WireServer(*deps)
}

// ProvideServer combines ServerDependencies and ServerResult into a single struct.
func ProvideServer(deps *ServerDependencies, result *ServerResult) *Server {
	return &Server{
		Dependencies: deps,
		Result:       result,
	}
}
