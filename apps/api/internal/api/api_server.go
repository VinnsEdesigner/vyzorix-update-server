package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/admin"
	alerthandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/alert"
	annotationhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/annotation"
	authhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	channelhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/channel"
	cmdhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/command"
	cvhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/configversion"
	confirmationhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/confirmation"
	dashboardhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/dashboard"
	devicehandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/device"
	diagnosticshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/diagnostics"
	inboxhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/inbox"
	notificationhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/notification"
	organizationhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/organization"
	sahandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/serviceaccount"
	updaterhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updater"
	updateshandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updates"
	websockethandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/wire"
	alertapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/alert"
	appannotation "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/annotation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	appconfigversion "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/configversion"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/confirmation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	diagnosticsapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/inbox"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	lifecycle "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/lifecycle"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	appnotification "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/notifications"
	orgapplication "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	searchsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/search"
	saapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/serviceaccount"
	updatesapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	usagestats "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/usagestats"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/featuremgmt"
	notificationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/appcheck"
	infracache "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/cache"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	infraMetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	infrawebhook "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/worker"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	wschannel "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws/channel"

	"github.com/gin-gonic/gin"
)

// ServerConfig holds the server configuration consumed by the wiring helpers
// (wireOperationalHandlers, wireAuthorizationWiring, wireInboxHandler, ...).
// It is built internally by NewServerWithDeps from the injected dependencies.
type ServerConfig struct {
	AlertConfig           config.AlertConfig
	Version               string
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

// LifecycleManager returns the worker manager for handler registration.
func (s *Server) LifecycleManager() *lifecycle.Manager {
	return s.lifecycleManager
}

// Server is the main API server.
type Server struct {
	ContactPointHandler         *notificationhandlers.Handler
	channelHandler              *channelhandlers.Handler
	ServiceAccountHandler       *sahandlers.Handler
	AnnotationHandler           *annotationhandlers.Handler
	configVersionHandler        *cvhandlers.Handler
	cache                       *infracache.Cache
	lifecycleManager            *lifecycle.Manager
	FCMRetryWorker              *worker.FCMRetryWorker
	statsSvc                    *usagestats.Service
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
	DB                          *storage.SQLite
	dashboardRateLimiter        *middleware.DashboardRateLimiterMiddleware
	deviceRegRateLimiter        *middleware.DeviceRegistrationRateLimiterMiddleware
	AuditLogger                 *audit.Logger
	deviceUpdaterHandler        *devicehandlers.UpdaterHandler
	deviceListHandler           *devicehandlers.ListHandler
	devicesHandler              *devicehandlers.DevicesHandler
	commandHandler              *cmdhandlers.ExecuteHandler
	confirmationHandler         *confirmationhandlers.Handler
	commandAuthorizer           *command.Authorizer
	grantRepo                   *storage.GrantRepository
	groupRepo                   *storage.DeviceGroupRepository
	InvitationStorage           *storage.InvitationStorage
	features                    *featuremgmt.Manager
	supportBundleHandler        *admin.SupportBundleHandler
	updateCheckerHandler        *admin.UpdateCheckerHandler
	searchService               *searchsvc.Service
	alertHandler                *alerthandlers.Handler
	AlertEvaluator              *alertapp.Evaluator
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

// Routes returns the Gin engine for serving.
func (s *Server) Routes() http.Handler {
	return s.engine
}

// wireOperationalHandlers wires the DB-backed operational surface that both
// construction paths need: the response cache, search/support-bundle, the
// dashboard/log/event/metrics/telemetry/stats handlers, inbox + confirm,
// diagnostics, the update checker, and the feature handlers.
//
// Ordering matters: the cache is created before the feature wiring so
// SetDashboardCache/SetHistoryCache actually take effect, and the alert
// evaluator is built before the annotator is attached to it.
func (s *Server) wireOperationalHandlers(cfg *ServerConfig) {
	if cfg.Config.CacheTTLSeconds > 0 {
		s.cache = infracache.New(time.Duration(cfg.Config.CacheTTLSeconds) * time.Second)
	}

	if cfg.DB != nil {
		s.wireFeatureHandlers(cfg, response.NewPresenter(nil, cfg.AuditLogger, nil))
	}

	// The dashboard/device rate limiter is dereferenced by the log, metrics,
	// telemetry and diagnostics RegisterRoutes helpers, so it must exist before
	// those handlers are wired (and before setupRoutes runs).
	s.dashboardRateLimiter = middleware.NewDashboardRateLimiterMiddleware(nil)

	s.wireDashboardHandlers(cfg)
	s.createUpdatesHandler(cfg)
	s.wireInboxHandler(cfg)
	s.wireConfirmHandler(cfg)
	s.wireDiagnosticsHandler(cfg)

	// Wire the telemetry-history response cache. The handler itself comes from the
	// injected HandlerSet (or the legacy path), so attach the cache here, after the
	// cache and the handler both exist.
	if s.cache != nil && s.telemetryHistoryHandler != nil {
		s.telemetryHistoryHandler.SetHistoryCache(s.cache.Section("telemetry"))
	}

	if cfg.DB != nil {
		s.initUpdateChecker(cfg)
	}
}

// wireAuthorizationWiring initializes the shared authorization plumbing used by
// both the REST route guards and the GraphQL resolver: the scoped permission
// grant / device-group repositories, the lifecycle registry, the feature
// manager, and the command risk gate. Both construction paths call this so the
// two never diverge.
func (s *Server) wireAuthorizationWiring(cfg *ServerConfig) {
	s.lifecycleManager = lifecycle.NewManager()

	// Scoped permission grants + device groups, wired for route-level
	// authorization (requireScope) and team-based device access.
	if cfg.DB != nil {
		permCache := permission.NewCache(30 * time.Second)
		s.grantRepo = storage.NewGrantRepositoryWithCache(cfg.DB.DB(), permCache)
		s.groupRepo = storage.NewDeviceGroupRepository(cfg.DB.DB())
		s.InvitationStorage = storage.NewInvitationStorage(cfg.DB.DB())
	}

	// Shared command risk gate. It gates REST command execution and the GraphQL
	// command mutation identically so neither transport can skip the MFA or
	// confirmation requirements.
	s.commandAuthorizer = command.NewAuthorizerFromService(cfg.ConfirmationService)

	// Feature manager: all new authz features default ON in production,
	// individually overridable via FEATURE_* env vars.
	allToggles := map[featuremgmt.Feature]bool{
		featuremgmt.ScopedRBAC:      true,
		featuremgmt.DeviceGroups:    true,
		featuremgmt.ServerDrivenOrg: true,
		featuremgmt.ActionSets:      true,
		featuremgmt.ScopeResolvers:  true,
		featuremgmt.PermissionCache: true,
	}
	s.features = featuremgmt.NewManager(allToggles)

	if cfg.Metrics != nil {
		for toggle := range allToggles {
			infraMetrics.Get().UpdateFeatureToggle(string(toggle), s.features.IsEnabled(toggle))
		}
	}
}

// initUpdateChecker builds the admin update-checker handler and its self-telemetry
// service. It is registered unconditionally by setupAdminRoutes, so it must be
// non-nil whenever the DB is available or the route panics.
func (s *Server) initUpdateChecker(cfg *ServerConfig) {
	collector := usagestats.NewCollector(cfg.DB.DB())
	statsSvc := usagestats.NewService(collector, s.features)
	version := cfg.Version
	if version == "" {
		version = "dev"
	}
	s.updateCheckerHandler = admin.NewUpdateCheckerHandler(version, "VinnsEdesigner/vyzorix-update-server")
	s.updateCheckerHandler.SetUsageStats(statsSvc)
	s.statsSvc = statsSvc
}

// wireFeatureHandlers wires the DB-backed feature handlers: search, support
// bundles, annotations, config versions, alerts + notifications, and service
// accounts. s.cache must already be initialized.
func (s *Server) wireFeatureHandlers(cfg *ServerConfig, presenter *response.Presenter) {
	s.searchService = searchsvc.NewService(cfg.DB.DB())
	s.supportBundleHandler = admin.NewSupportBundleHandler(cfg.DB)

	annotationRepo := storage.NewAnnotationRepository(cfg.DB.DB())
	annotationService := appannotation.NewService(annotationRepo)
	s.AnnotationHandler = annotationhandlers.NewHandler(annotationService)
	// Note: the alert evaluator is built later in this function, so the annotator
	// is attached there (not here) - otherwise it would be attached to nil.
	if cfg.UpdatesService != nil && cfg.UpdatesService.GetPushService() != nil {
		cfg.UpdatesService.GetPushService().SetAnnotator(appannotation.NewAlertAnnotator(annotationService))
		if s.cache != nil {
			cfg.UpdatesService.GetPushService().SetDashboardCache(s.cache.Section("dashboard"))
		}
	}

	configVersionRepo := storage.NewConfigVersionRepository(cfg.DB.DB())
	configVersionSvc := appconfigversion.NewService(configVersionRepo)
	if cfg.OrgSettingsService != nil {
		cfg.OrgSettingsService.SetVersionService(configVersionSvc)
	}
	if cfg.DeviceSettingsService != nil {
		cfg.DeviceSettingsService.SetVersionService(configVersionSvc)
	}
	s.configVersionHandler = cvhandlers.NewHandler(configVersionSvc, cfg.OrgSettingsService)

	alertRules := storage.NewAlertRuleRepository(cfg.DB.DB())
	alertStates := storage.NewAlertStateRepository(cfg.DB.DB())
	alertHistory := storage.NewAlertHistoryRepository(cfg.DB.DB())
	alertService := alertapp.NewService(alertRules, alertStates, alertHistory)
	cpRepo := storage.NewContactPointRepository(cfg.DB.DB())
	cpDeliveries := storage.NewDeliveryRepository(cfg.DB.DB())
	notificationSvc := appnotification.NewService(cpRepo)
	channels := map[notificationdomain.ChannelType]notificationdomain.Channel{
		notificationdomain.ChannelTypeEmail:   appnotification.NewEmailChannel(emailService.NewService()),
		notificationdomain.ChannelTypeWebhook: appnotification.NewWebhookChannel(infrawebhook.NewClient(10 * time.Second)),
		notificationdomain.ChannelTypeSlack:   appnotification.NewSlackChannel(infrawebhook.NewClient(10 * time.Second)),
	}
	dispatcher := appnotification.NewDispatcher(cpRepo, cpDeliveries, channels, cfg.Log)
	alertNotifier := appnotification.NewAlertNotifierAdapter(dispatcher)

	s.AlertEvaluator = alertapp.NewEvaluator(
		alertRules,
		alertStates,
		alertHistory,
		alertapp.NewMetricSource(cfg.DB.DB(), int64(cfg.AlertConfig.MetricWindowSecs)),
		alertNotifier,
		cfg.Hub,
		cfg.Log,
	)
	s.AlertEvaluator.SetAnnotator(appannotation.NewAlertAnnotator(annotationService))
	if s.cache != nil {
		s.AlertEvaluator.SetDashboardCache(s.cache.Section("dashboard"))
	}
	channelMgr := wschannel.NewManager(wschannel.NewMembershipAuthorizer(storage.NewMemberStorage(cfg.DB.DB())))
	adapter := wschannel.NewMessageMapAdapter(channelMgr)
	s.AlertEvaluator.SetStreamPublisher(adapter)
	s.alertHandler = alerthandlers.NewHandler(alertService, s.AlertEvaluator)
	hubBridge := wschannel.NewHubBridge(channelMgr, cfg.Hub)
	s.channelHandler = channelhandlers.NewHandler(hubBridge, presenter)

	s.ContactPointHandler = notificationhandlers.NewHandler(notificationSvc, dispatcher)

	saRepo := storage.NewServiceAccountRepository(cfg.DB.DB())
	saTokenRepo := storage.NewServiceAccountTokenRepository(cfg.DB.DB())
	saService := saapp.NewService(saRepo, saTokenRepo)
	saService.SetAuditLogger(cfg.AuditLogger)
	if cfg.DB != nil {
		saService.SetMemberRepository(storage.NewMemberStorage(cfg.DB.DB()))
	}
	if limit := os.Getenv("SA_TOKEN_LIMIT"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil && n > 0 {
			saService.SetMaxTokens(n)
		}
	}
	s.ServiceAccountHandler = sahandlers.NewHandler(saService)
	if s.tenantAPIKeyAuth != nil {
		s.tenantAPIKeyAuth.SetServiceAccountAuth(saService)
	}
}

// wireDashboardHandlers creates and assigns dashboard command handler instances.
func (s *Server) wireDashboardHandlers(cfg *ServerConfig) {
	logsRepo, metricsRepo, eventRepo := s.createStorageRepositories(cfg)
	historySvc, logsSvc, metricsSvc, dashboardSvc := s.createDashboardServices(cfg, logsRepo, metricsRepo)
	s.createDashboardHandlers(cfg, historySvc, logsSvc, metricsSvc, eventRepo, dashboardSvc)
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

func (s *Server) wireConfirmHandler(cfg *ServerConfig) {
	if cfg.DeviceService == nil || cfg.DB == nil {
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
	AlertConfig         config.AlertConfig
	Version             string
	Log                 *slog.Logger
	DB                  *storage.SQLite
	Engine              *gin.Engine
	Middleware          *wire.MiddlewareSet
	HandlerSet          *wire.HandlerSet
	SessionManager      *infraauth.SessionManager
	Hub                 *hub.Hub
	AuditLogger         *audit.Logger
	UpdatesService      *updatesapp.Service
	APIKeyService       *keys.APIKeyService
	CommandService      *command.Service
	ConfirmationService *confirmation.Service
	DeviceRepo          *storage.DeviceRepository
	IdempotencyRepo     *storage.IdempotencyRepository
	FCMRetryWorker      *worker.FCMRetryWorker
	Metrics             *infraMetrics.Metrics
	Config              config.Config
}

// NewServerWithDeps creates a Server using pre-wired dependencies from wire.
func NewServerWithDeps(cfg *ServerConfigWithDeps) *Server {
	if cfg.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ProvideMiddlewareSet does not wire the API key middleware, so backfill
	// TenantAPIKeyAuth here to ensure API key auth is enforced on tenant routes.
	if cfg.Middleware != nil && cfg.Middleware.TenantAPIKeyAuth == nil && cfg.APIKeyService != nil && cfg.AuditLogger != nil {
		cfg.Middleware.TenantAPIKeyAuth = middleware.NewTenantAPIKeyAuth(cfg.APIKeyService, cfg.AuditLogger)
		if cfg.Middleware.APIKeyRateLimiter == nil {
			cfg.Middleware.APIKeyRateLimiter = middleware.NewInMemoryRateLimiter(100, time.Minute)
		}
	}

	// ProvideMiddlewareSet omits the request-signing verifier and the response
	// encryption key function (only WireMiddleware produced them). Without them
	// device request signing is silently skipped and response encryption panics
	// on any request carrying X-Client-ID, so recover both from the factory.
	if cfg.Middleware != nil && cfg.Middleware.Factory != nil {
		if cfg.Middleware.SignatureVerifier == nil {
			cfg.Middleware.SignatureVerifier = cfg.Middleware.Factory.GetSignatureVerifier()
		}
		if cfg.Middleware.EncryptKeyFn == nil {
			cfg.Middleware.EncryptKeyFn = cfg.Middleware.Factory.GetEncryptionKeyFn()
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
		signatureVerifier:        cfg.Middleware.SignatureVerifier,
		encryptKeyFn:             cfg.Middleware.EncryptKeyFn,
		sessionManager:           cfg.SessionManager,
		DB:                       cfg.DB,
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

	// This is the single server constructor: it wires everything the API needs
	// from the injected dependencies, so the surface cannot drift between callers.
	serverCfg := &ServerConfig{
		DB:                    cfg.DB,
		Version:               cfg.Version,
		DeviceService:         cfg.HandlerSet.DeviceService,
		CommandService:        cfg.CommandService,
		ConfirmationService:   cfg.ConfirmationService,
		FCMNotifier:           cfg.HandlerSet.FCMNotifier,
		Config:                cfg.Config,
		Log:                   cfg.Log,
		AuditLogger:           cfg.AuditLogger,
		Metrics:               cfg.Metrics,
		UpdatesService:        cfg.UpdatesService,
		DeviceSettingsService: cfg.HandlerSet.DeviceSettingsService,
		OrgSettingsService:    cfg.HandlerSet.OrgSettingsService,
		AlertConfig:           cfg.AlertConfig,
		AppCheckVerifier:      cfg.HandlerSet.AppCheckVerifier,
	}
	s.wireAuthorizationWiring(serverCfg)
	s.wireOperationalHandlers(serverCfg)

	// Create updater handler for OTA update distribution.
	s.updaterHandler = updaterhandlers.NewHandler(cfg.Log, cfg.Config)

	// Store deviceRepo for the deletion worker in api_main.go.
	s.DeviceRepo = cfg.DeviceRepo
	s.FCMRetryWorker = cfg.FCMRetryWorker

	// Metrics handler + request middleware, so /metrics is registered and request
	// metrics are recorded.
	if cfg.Metrics != nil {
		s.metricsHandler = infraMetrics.NewMetricsHandler(cfg.Metrics)
		s.engine.Use(infraMetrics.Middleware(cfg.Metrics))
	}

	// Start the WebSocket hub event loop; without it device registration,
	// broadcasts and telemetry never run.
	if cfg.Hub != nil {
		go cfg.Hub.Run(context.Background())
	}

	s.setupRoutes()

	return s
}

// UsageStatsService returns the self-telemetry snapshot service.
func (s *Server) UsageStatsService() *usagestats.Service {
	return s.statsSvc
}
