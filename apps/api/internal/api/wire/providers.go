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
	updateshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updates"
	websockethandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/confirmation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	eventapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/event"
	keys "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	appnotification "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/notification"
	orgapplication "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	updatesapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/appcheck"
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
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/worker"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

const (
	// Default JWT expiry for access tokens.
	defaultJWTAccessExpiry = 15 * time.Minute
	// Default refresh token expiry.
	defaultRefreshTokenExpiry = 7 * 24 * time.Hour // 7 days.
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
//
// The backend is selected from config: when TURSO_DB_URL is set (or.
// DATABASE_BACKEND=turso) the server connects to Turso libSQL over HTTP for.
// production; otherwise it opens a local SQLite file for development. Both.
// paths run through the same migration registry and return the same.
// *storage.SQLite handle, so every repository works unchanged either way.
func ProvideSQLite(cfg config.Config, log *slog.Logger) (*storage.SQLite, error) {
	storageCfg := buildStorageConfig(cfg)
	return storage.OpenWithLogger(storageCfg, log)
}

// buildStorageConfig translates the app config into a storage.Config bound to.
// the resolved backend.
func buildStorageConfig(cfg config.Config) *storage.Config {
	switch cfg.ResolvedDatabaseBackend() {
	case "turso":
		return &storage.Config{
			Backend:           storage.BackendTurso,
			TursoURL:          cfg.TursoDatabaseURL,
			TursoAuthToken:    cfg.TursoAuthToken,
			MaxOpenConns:      cfg.DatabaseMaxOpenConns,
			MaxIdleConns:      cfg.DatabaseMaxIdleConns,
			ConnMaxLifetime:   cfg.DatabaseConnMaxLifetime,
			ConnMaxIdleTime:   cfg.DatabaseConnMaxIdleTime,
			RequestTimeout:    cfg.DatabaseRequestTimeout,
			HealthCheckPeriod: cfg.DatabaseHealthCheckPeriod,
		}
	default:
		sqliteCfg := storage.DefaultConfig(cfg.DatabaseURL)
		return sqliteCfg
	}
}

// ProvideDB returns the underlying *sql.DB from SQLite.
func ProvideDB(s *storage.SQLite) *sql.DB {
	return s.DB()
}

// ProvideIdempotencyRepository creates the idempotency key repository.
func ProvideIdempotencyRepository(db *sql.DB) *storage.IdempotencyRepository {
	return storage.NewIdempotencyRepository(db)
}

// ProvideConfirmationService creates the confirmation service backed by the.
// SQLite confirmation store. Returns nil when no DB is available, in which.
// case the command handler blocks risky commands (confirmations disabled).
func ProvideConfirmationService(db *storage.SQLite) *confirmation.Service {
	if db == nil {
		return nil
	}
	return confirmation.NewService(storage.NewConfirmationRepository(db.DB()))
}

// ProvideAuditLogger creates the audit logger with a dedicated file-based logger.

func ProvideAuditLogger(db *sql.DB, cfg config.Config) *audit.Logger {
	auditRepo := audit.NewRepository(db)
	auditLog := logging.NewAuditFileLogger(cfg.AuditLogPath)

	var separateRepo interface {
		Log(context.Context, *audit.Entry) error
	}
	if cfg.AuditLogSeparateDB && cfg.AuditLogSeparateDBPath != "" {
		sepRepo, err := audit.NewSeparateDBAuditRepository(audit.SeparateDBConfig{Path: cfg.AuditLogSeparateDBPath})
		if err != nil {
			auditLog.Warn("failed to create separate audit DB, continuing without it", "error", err)
			separateRepo = &audit.NoOpAuditRepo{}
		} else {
			separateRepo = sepRepo
			auditLog.Info("audit logging to separate database", "path", cfg.AuditLogSeparateDBPath)
		}
	} else {
		separateRepo = &audit.NoOpAuditRepo{}
	}

	return audit.NewLogger(auditRepo, auditLog, audit.DefaultLoggerConfig(), separateRepo)
}

// ProvidePasswordHasher creates the password hasher.
func ProvidePasswordHasher() *passwordpkg.Argon2idHasher {
	return passwordpkg.NewArgon2idHasher()
}

// ProvideSessionManager creates the session manager.
// The manager needs its session repository wired so that session-management.
// endpoints (list/revoke/concurrent) can query persisted sessions. Without.
// this, ListActiveSessions returns "session repository not configured".
// The storage repository speaks the domain/session types; the manager's.
// Repository interface speaks security/session types, so we adapt here.
func ProvideSessionManager(cfg config.Config, sessionRepo *storage.SessionRepository) *infraauth.SessionManager {
	m := infraauth.NewSessionManager(cfg.SessionSecret)
	m.SetRepository(newSessionRepoAdapter(sessionRepo))
	return m
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

// ProvideDiagnosticsRepository creates the diagnostics repository.
func ProvideDiagnosticsRepository(db *sql.DB) *storage.DiagnosticsRepository {
	return storage.NewDiagnosticsRepository(db)
}

// ProvideUpdatesStorage creates the updates storage.
func ProvideUpdatesStorage(db *sql.DB) *storage.UpdatesStorage {
	return storage.NewUpdatesStorage(db)
}

// ProvideEventProcessor creates the event processor and wires it to the hub.
func ProvideEventProcessor(
	eventRepo *storage.EventRepository,
	diagnosticsRepo *storage.DiagnosticsRepository,
	deviceRepo *storage.DeviceRepository,
	deviceSettingsRepo *storage.DeviceSettingsRepository,
	orgSettingsRepo *storage.OrganizationSettingsRepository,
	hubResult *HubResult,
	log *slog.Logger,
) *eventapp.Processor {
	// Create broadcaster first.
	broadcaster := eventapp.NewBroadcaster(log)
	hubResult.Hub.SetDashboardBroadcaster(broadcaster)

	// Create processor with broadcaster.
	processor := eventapp.NewProcessor(eventRepo, deviceRepo, broadcaster, log)

	// Wire repositories for hierarchical threshold resolution.
	processor.SetDeviceSettingsRepo(deviceSettingsRepo)
	processor.SetOrgSettingsRepo(orgSettingsRepo)

	// Wire diagnostics recorder for timeline events.
	processor.SetDiagnosticsRecorder(diagnosticsRepo)

	hubResult.Hub.SetEventProcessor(processor)

	return processor
}

// ProvideNotificationService creates the notification service.
func ProvideNotificationService(
	operatorRepo *storage.OperatorRepository,
	emailSvc *emailService.Service,
	webhookClient *webhook.Client,
	auditRepo *infranotification.Repository,
	cfg config.Config,
	log *slog.Logger,
) *appnotification.Service {
	return appnotification.NewService(operatorRepo, emailSvc, webhookClient, auditRepo, cfg.BaseURL, log)
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

// ProvideRefreshTokenRepository creates the refresh token repository.
func ProvideRefreshTokenRepository(db *sql.DB) *storage.RefreshTokenRepository {
	return storage.NewRefreshTokenRepository(db)
}

// ProvideJWTManager creates the JWT manager for access tokens.
func ProvideJWTManager(cfg config.Config) (*infraauth.JWTManager, error) {
	return infraauth.NewJWTManager(
		cfg.JWTSecret,
		defaultJWTAccessExpiry,
		"vyz-api",
	)
}

// ProvideAuthService creates the auth service with full refresh token support.
func ProvideAuthService(
	log *slog.Logger,
	operatorRepo *storage.OperatorRepository,
	sessionRepo *storage.SessionRepository,
	emailVerifyRepo *storage.EmailVerificationRepository,
	passwordResetRepo *storage.PasswordResetRepository,
	refreshTokenRepo *storage.RefreshTokenRepository,
	hasher *passwordpkg.Argon2idHasher,
	jwtManager *infraauth.JWTManager,
	memberRepo *storage.MemberStorage,
	invitationRepo *storage.InvitationStorage,
) *auth.AuthService {
	sessionTTL := 7 * 24 * time.Hour
	svc := auth.NewAuthServiceWithRefresh(
		operatorRepo,
		sessionRepo,
		emailVerifyRepo,
		passwordResetRepo,
		hasher,
		sessionTTL,
		refreshTokenRepo,
		defaultRefreshTokenExpiry,
		jwtManager,
		nil, // ldapConfig - not used currently.
	)
	svc.SetLogger(log)
	// Wire the member repository so membership-dependent flows work: listing the.
	// operator's organizations (GET /v1/auth/organizations, /v1/auth/me) and.
	// organization selection (POST /v1/auth/organizations/select). Without this,
	// these endpoints silently return empty / "not a member" even when the.
	// operator owns an organization.
	svc.SetMemberRepository(memberRepo)
	// Wire the invitation repository so the operator deletion flow.
	// (DELETE /v1/auth/admin/operators/:id) can cancel pending invitations.
	// Without this, s.invitationRepo is nil and DeleteOperator panics.
	svc.SetInvitationRepository(invitationRepo)
	return svc
}

// ProvideDeviceService creates the device service.

func ProvideDeviceService(deviceRepo *storage.DeviceRepository, operatorRepo *storage.OperatorRepository, txManager transaction.TxManager, log *slog.Logger, hub *hub.Hub) *device.Service {
	svc := device.NewService(deviceRepo, operatorRepo, log)
	svc.WithTxManager(txManager)
	svc.WithStatusUpdater(hub)
	return svc
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

// ProvideOrganizationRepository creates the organization repository.
func ProvideOrganizationRepository(db *sql.DB) *storage.OrganizationStorage {
	return storage.NewOrganizationStorage(db)
}

// ProvideMemberRepository creates the member repository.
func ProvideMemberRepository(db *sql.DB) *storage.MemberStorage {
	return storage.NewMemberStorage(db)
}

// ProvideInvitationRepository creates the invitation repository.
func ProvideInvitationRepository(db *sql.DB) *storage.InvitationStorage {
	return storage.NewInvitationStorage(db)
}

// ProvideTxManager creates a simple transaction manager for SQLite.
func ProvideTxManager(db *sql.DB) transaction.TxManager {
	return storage.NewTxManager(db)
}

// ProvideOrganizationService creates the organization service.
func ProvideOrganizationService(
	orgRepo *storage.OrganizationStorage,
	memberRepo *storage.MemberStorage,
	invitationRepo *storage.InvitationStorage,
	operatorRepo *storage.OperatorRepository,
	sessionRepo *storage.SessionRepository,
	deviceRepo *storage.DeviceRepository,
	telemetryRepo *storage.TelemetryRepository,
	commandRepo *storage.CommandRepository,
	txManager transaction.TxManager,
	log *slog.Logger,
) *orgapplication.OrganizationService {
	return orgapplication.NewOrganizationService(orgRepo, memberRepo, invitationRepo, operatorRepo, sessionRepo, deviceRepo, telemetryRepo, commandRepo, txManager, log)
}

// ProvideMemberService creates the member service.
func ProvideMemberService(
	memberRepo *storage.MemberStorage,
	orgRepo *storage.OrganizationStorage,
	authSvc *auth.AuthService,
	log *slog.Logger,
) *orgapplication.MemberService {
	return orgapplication.NewMemberService(memberRepo, orgRepo, authSvc, log)
}

// ProvideInvitationService creates the invitation service.
func ProvideInvitationService(
	invitationRepo *storage.InvitationStorage,
	orgRepo *storage.OrganizationStorage,
	memberRepo *storage.MemberStorage,
	sessionRepo *storage.SessionRepository,
	txManager transaction.TxManager,
	emailSvc *emailService.Service,
	log *slog.Logger,
	cfg config.Config,
) *orgapplication.InvitationService {
	baseURL := cfg.FrontendURL
	if baseURL == "" {
		baseURL = ""
	}
	return orgapplication.NewInvitationService(invitationRepo, orgRepo, memberRepo, sessionRepo, txManager, emailSvc, nil, log, baseURL)
}

// ProvideOrganizationSettingsRepository creates the organization settings repository.
func ProvideOrganizationSettingsRepository(db *sql.DB) *storage.OrganizationSettingsRepository {
	return storage.NewOrganizationSettingsRepository(db)
}

// ProvideOrganizationSettingsService creates the organization settings service.
func ProvideOrganizationSettingsService(
	settingsRepo *storage.OrganizationSettingsRepository,
	orgRepo *storage.OrganizationStorage,
	memberRepo *storage.MemberStorage,
) *orgapplication.OrganizationSettingsService {
	return orgapplication.NewOrganizationSettingsService(settingsRepo, orgRepo, memberRepo)
}

// ProvideDeviceSettingsRepository creates the device settings repository.
func ProvideDeviceSettingsRepository(db *sql.DB) *storage.DeviceSettingsRepository {
	return storage.NewDeviceSettingsRepository(db)
}

// ProvideDeviceSettingsService creates the device settings service.
func ProvideDeviceSettingsService(
	settingsRepo *storage.DeviceSettingsRepository,
	deviceRepo *storage.DeviceRepository,
	orgSettingsRepo *storage.OrganizationSettingsRepository,
) *device.DeviceSettingsService {
	return device.NewDeviceSettingsService(settingsRepo, deviceRepo, orgSettingsRepo)
}

// HubResult holds the WebSocket hub components.
type HubResult struct {
	Hub          *hub.Hub
	MessageQueue *hub.MessageQueue
	RateLimiter  *hub.RateLimiter
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

// ProvideFCMNotifier creates the FCM notifier with persistent retry support.

func ProvideFCMNotifier(log *slog.Logger, cfg config.Config, db *sql.DB) fcm.Notifier {
	if cfg.FirebaseCreds == "" {
		return nil
	}
	fcmClient, err := fcm.Init(log, cfg.FirebaseCreds)
	if err != nil {
		log.Error("failed to initialize FCM - malformed Firebase credentials",
			"error", err)
		return nil
	}

	notifier := fcm.NewEnhancedNotifier(fcmClient, fcm.DefaultFCMConfig())

	if db != nil {
		notifier = fcm.NewEnhancedNotifierWithDB(fcmClient, fcm.DefaultFCMConfig(), db)
		retryWorker := worker.NewFCMRetryWorker(db, notifier, log, 30*time.Second)
		retryWorker.Start()
		log.Info("FCM retry worker started", "interval", "30s")
	}

	return notifier
}

// ProvideAppCheckVerifier creates the Firebase App Check verifier.
// Returns nil if FIREBASE_CREDENTIALS is not configured.
func ProvideAppCheckVerifier(log *slog.Logger, cfg config.Config) (*appcheck.Verifier, error) {
	if cfg.FirebaseCreds == "" {
		log.Warn("Firebase credentials not configured, App Check disabled")
		//nolint:nilnil // Intentionally returns nil,nil when App Check is disabled (no credentials configured).
		return nil, nil
	}
	verifier, err := appcheck.NewVerifier(log, cfg.FirebaseCreds, cfg.FirebaseAppID)
	if err != nil {
		return nil, err
	}
	return verifier, nil
}

// ProvideDeviceDeletionWorker creates the background worker for scheduled device deletion.
// Returns nil if the worker is disabled.
func ProvideDeviceDeletionWorker(deviceRepo *storage.DeviceRepository, log *slog.Logger, cfg config.Config) *worker.DeviceDeletionWorker {
	if !cfg.DeviceDeletionEnabled {
		log.Info("device deletion worker disabled via config")
		return nil
	}
	interval := time.Duration(cfg.DeviceDeletionIntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = 5 * time.Minute // default 5 minutes.
	}
	w := worker.NewDeviceDeletionWorker(deviceRepo, log, interval)
	w.Start()
	log.Info("device deletion worker started", "interval", interval.String())
	return w
}

// ProvideCommandOutbox creates and starts the command outbox background worker.
// This implements the transactional outbox pattern for command delivery.
// The outbox polls for pending commands and delivers them via WebSocket or FCM.
func ProvideCommandOutbox(
	commandRepo *storage.CommandRepository,
	deviceRepo *storage.DeviceRepository,
	hub *hub.Hub,
	fcmNotifier fcm.Notifier,
	log *slog.Logger,
) *cmdapp.Outbox {
	cfg := cmdapp.DefaultOutboxConfig()
	outbox := cmdapp.NewOutbox(commandRepo, deviceRepo, hub, fcmNotifier, cfg, log)
	outbox.Start()
	log.Info("command outbox worker started",
		"pollInterval", cfg.PollInterval.String(),
		"batchSize", cfg.BatchSize,
		"maxRetries", cfg.MaxRetries)
	return outbox
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
			RateLimitPerMin:  cfg.RateLimitPerMin,
			AuthRateLimitMin: cfg.AuthRateLimitMin,
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

// ProvideSessionSignVerifier creates the session signature verifier for tenant routes.
func ProvideSessionSignVerifier(factory *middleware.MiddlewareFactory) *cryptohmac.Verifier {
	return factory.GetSessionSignatureVerifier()
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
	pushService *updatesapp.PushService,
	auditLogger *audit.Logger,
	cfg config.Config,
) *updateshandlers.UpdatesHandler {
	updatesRateLimiters := middleware.NewUpdatesRateLimiterMiddleware(nil)
	return updateshandlers.NewUpdatesHandler(
		updatesService,
		pushService,
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
	})

	hs.DeviceStatus = devicehandlers.NewStatusHandler(deviceService)
	hs.DeviceUpdater = devicehandlers.NewUpdaterHandler(deviceService)
	hs.DeviceList = devicehandlers.NewListHandler(deviceService, hubResult.Hub)
	hs.Command = cmdhandlers.NewExecuteHandler(commandService, deviceService, hubResult.Hub, fcmNotifier, cmdapp.NewAuthorizer(nil), auditLogger)
	hs.Stream = websockethandlers.NewStreamHandler(log, cfg, hubResult.Hub, *hmacVerifier, auditLogger)
	hs.TelemetryHistory = handlers.NewTelemetryHistoryHandler(log, storage.NewTelemetryRepository(db), storage.NewDeviceRepository(db), nil)
	hs.ConnectionStatus = handlers.NewConnectionStatusHandler(log, hubResult.Hub, storage.NewDeviceRepository(db))
	hs.AdminClients = admin.NewClientsHandler(clientService)
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
	// Create a separate rate limiter for auth with different limits.
	authLimiter := middleware.NewRateLimiter(5, time.Minute)
	return &MiddlewareSet{
		Factory:        factory,
		CookieAuth:     cookieAuth,
		Lockout:        lockout,
		IPIntelligence: ipIntelligence,
		HmacVerifier:   hmacVerifier,
		RateLimiter:    rateLimiter,
		AuthLimiter:    authLimiter,
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
	ProvideIdempotencyRepository,
	ProvideConfirmationService,
	ProvideUpdatesStorage,
	ProvideEventProcessor,
	ProvideEmailVerificationRepository,
	ProvidePasswordResetRepository,
	ProvideRefreshTokenRepository,
	ProvideOrganizationRepository,
	ProvideMemberRepository,
	ProvideInvitationRepository,
	ProvideTxManager,
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
	ProvideOrganizationService,
	ProvideMemberService,
	ProvideInvitationService,
	ProvideOrganizationSettingsRepository,
	ProvideOrganizationSettingsService,
	ProvideDeviceSettingsRepository,
	ProvideDeviceSettingsService,
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
	appCheckVerifier *appcheck.Verifier,
	deviceDeletionWorker *worker.DeviceDeletionWorker,
	commandOutbox *cmdapp.Outbox,
	factory *middleware.MiddlewareFactory,
	rateLimiter *middleware.RateLimiter,
	lockout *middleware.Lockout,
	ipIntelligence *middleware.IPIntelligence,
	updatesService *updatesapp.Service,
	apiKeyService *keys.APIKeyService,
	orgService *orgapplication.OrganizationService,
	memberService *orgapplication.MemberService,
	invitationService *orgapplication.InvitationService,
	orgSettingsService *orgapplication.OrganizationSettingsService,
	deviceSettingsService *device.DeviceSettingsService,
	idempotencyRepo *storage.IdempotencyRepository,
	confirmationService *confirmation.Service,
) *ServerDependencies {
	return &ServerDependencies{
		FCMNotifier:           fcmNotifier,
		AppCheckVerifier:      appCheckVerifier,
		DeviceDeletionWorker:  deviceDeletionWorker,
		CommandOutbox:         commandOutbox,
		DeviceRepo:            deviceRepo,
		OperatorRepo:          operatorRepo,
		RateLimiter:           rateLimiter,
		Hub:                   hubResult.Hub,
		AuthService:           authService,
		AuthLimiter:           middleware.NewRateLimiter(cfg.AuthRateLimitMin, time.Minute),
		IPIntelligence:        ipIntelligence,
		Log:                   log,
		SessionManager:        sessionManager,
		GoogleVerifier:        googleVerifier,
		EmailService:          emailService,
		EmailVerificationRepo: emailVerifyRepo,
		CommandService:        commandService,
		ClientService:         clientService,
		DB:                    sqlite,
		Lockout:               lockout,
		DeviceService:         deviceService,
		Metrics:               metrics,
		AuditLogger:           auditLogger,
		Config:                cfg,
		UpdatesStorage:        updatesStorage,
		UpdatesService:        updatesService,
		TelemetryRepo:         telemetryRepo,
		APIKeyService:         apiKeyService,
		OrgService:            orgService,
		MemberService:         memberService,
		InvitationService:     invitationService,
		OrgSettingsService:    orgSettingsService,
		DeviceSettingsService: deviceSettingsService,
		IdempotencyRepo:       idempotencyRepo,
		ConfirmationService:   confirmationService,
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
