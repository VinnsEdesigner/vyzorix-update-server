// Package main provides the Vyzorix API server entry point.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	passwordpkg "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/password"
	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/github"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/logging"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/ssr"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
)

func main() {
	env := getEnv()
	PrintBanner(env)

	log := logging.NewFromEnv()

	cfg, err := config.Load()
	if err != nil {
		log.Error("configuration failed", "err", err)
		os.Exit(1)
	}

	db, err := initDatabase(cfg)
	if err != nil {
		log.Error("database initialization failed", "err", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Error("database close failed", "err", closeErr)
		}
	}()

	operatorRepo := storage.NewOperatorRepository(db.DB())
	deviceRepo := storage.NewDeviceRepository(db.DB())
	commandRepo := storage.NewCommandRepository(db.DB())

	_, sessionManager, googleVerifier := initSecurity(cfg)

	authService, deviceService, clientService, commandService, emailSvc,
		fcmNotifier, telemetryRepo, wsHub, _, _ := initServices(cfg, log, db, deviceRepo, operatorRepo, commandRepo)

	rateLimiter, authLimiter, accountLockout := initMiddleware()

	ssrConfig, ssrManager := initSSR(log)

	apiServer := createAPIServer(cfg, log, db, operatorRepo, sessionManager, googleVerifier,
		authService, deviceService, clientService, commandService, emailSvc,
		rateLimiter, authLimiter, accountLockout, wsHub, fcmNotifier)

	if cfg.EnableGraphQL {
		// Create dashboard services for GraphQL
		historyService := cmdapp.NewHistoryService(commandRepo, deviceRepo)
		logsRepo := storage.NewLogsRepository(db.DB())
		metricsRepo := storage.NewMetricsRepository(db.DB())
		logsSvc := logs.NewService(logsRepo, log)
		metricsSvc := appmetrics.NewService(metricsRepo)
		dashboardSvc := dashboard.NewService(deviceRepo, commandRepo)

		// Create UpdatesService for GraphQL
		updatesStorage := storage.NewUpdatesStorage(db.DB())
		versionsStatusSvc := updatesapp.NewVersionsStatusService(updatesStorage)
		versionsListSvc := updatesapp.NewVersionsListService(updatesStorage)
		changelogSvc := updatesapp.NewChangelogService(updatesStorage)
		exportSvc := updatesapp.NewExportService(updatesStorage)
		updatesHistorySvc := updatesapp.NewHistoryService(updatesStorage)

		// GitHub sync
		var githubSyncSvc *github.SyncService
		if cfg.GitHubReleaseToken != "" && cfg.GitHubReleaseRepo != "" {
			githubClient := github.NewClient("VinnsEdesigner", cfg.GitHubReleaseRepo, cfg.GitHubReleaseToken)
			githubSyncSvc = github.NewSyncService(githubClient, updatesStorage, log)
		}
		syncSvc := updatesapp.NewSyncService(updatesStorage, githubSyncSvc)

		// PushService needs Hub, FCM, CommandService, DeviceService
		pushSvc := updatesapp.NewPushService(updatesStorage, deviceService, wsHub, fcmNotifier, commandService, log)

		// Create main UpdatesService
		updatesSvc := updatesapp.NewService(
			updatesStorage,
			versionsStatusSvc,
			versionsListSvc,
			changelogSvc,
			exportSvc,
			pushSvc,
			updatesHistorySvc,
			syncSvc,
		)

		if regErr := apiServer.RegisterGraphQL(
			deviceService, commandService, historyService, dashboardSvc,
			logsSvc, metricsSvc, telemetryRepo, logsRepo, metricsRepo, wsHub,
			updatesSvc,
		); regErr != nil {
			log.Error("failed to register GraphQL", "err", regErr)
		}
	}

	startServer(cfg.Port, log, apiServer, ssrConfig, ssrManager)
}

func initDatabase(cfg config.Config) (*storage.SQLite, error) {
	PrintSection("Database")

	sqliteCfg := storage.DefaultConfig(cfg.DatabaseURL)
	db, err := storage.Open(sqliteCfg)
	if err != nil {
		return nil, err
	}
	PrintStatus("Database", "Connected to "+cfg.DatabaseURL)

	PrintStatus("Repositories", "Initialized")

	return db, nil
}

func initSecurity(cfg config.Config) (*passwordpkg.Argon2idHasher, *infraauth.SessionManager, *infraauth.GoogleTokenVerifier) {
	PrintSection("Security")

	passwordHasher := passwordpkg.NewArgon2idHasher()
	PrintStatus("PasswordHasher", "Argon2id initialized")

	sessionManager := infraauth.NewSessionManager(cfg.SessionSecret)
	PrintStatus("SessionManager", "Initialized")

	googleVerifier := infraauth.NewGoogleTokenVerifier(cfg.GoogleOAuthClientID)
	PrintStatus("GoogleVerifier", "Initialized")

	return passwordHasher, sessionManager, googleVerifier
}

func initServices(cfg config.Config, log *slog.Logger, db *storage.SQLite,
	deviceRepo *storage.DeviceRepository, operatorRepo *storage.OperatorRepository,
	commandRepo *storage.CommandRepository) (*appsvc.AuthService, *device.Service, *client.Service,
	*cmdapp.Service, *emailService.Service, fcm.Notifier, *storage.TelemetryRepository,
	*hub.Hub, *hub.MessageQueue, *hub.RateLimiter) {
	PrintSection("Services")

	sessionTTL := 7 * 24 * time.Hour

	authService := appsvc.NewAuthService(
		operatorRepo,
		storage.NewSessionRepository(db.DB()),
		storage.NewEmailVerificationRepository(db.DB()),
		storage.NewPasswordResetRepository(db.DB()),
		passwordpkg.NewArgon2idHasher(),
		sessionTTL,
	)
	PrintStatus("AuthService", "Initialized")

	deviceService := device.NewService(deviceRepo, operatorRepo)
	PrintStatus("DeviceService", "Initialized")

	clientService := client.NewService(storage.NewClientRepository(db.DB()))
	PrintStatus("ClientService", "Initialized")

	commandService := cmdapp.NewService(commandRepo, deviceRepo)
	PrintStatus("CommandService", "Initialized")

	emailSvc := emailService.NewService()
	if cfg.ResendAPIKey != "" {
		PrintStatus("EmailService", "Resend configured")
	} else {
		PrintWarning("EmailService", "RESEND_API_KEY not configured")
	}

	var fcmNotifier fcm.Notifier
	if fcmClient, err := fcm.Init(log, cfg.FirebaseCreds); err == nil {
		fcmNotifier = fcm.NewEnhancedNotifier(fcmClient, fcm.DefaultFCMConfig())
		PrintStatus("FCM", "Initialized")
	} else {
		PrintWarning("FCM", "Firebase not configured")
	}

	telemetryRepo := storage.NewTelemetryRepository(db.DB())
	PrintStatus("TelemetryRepository", "Initialized")

	hubConfig := &hub.HubConfig{
		MessageQueue: hub.DefaultMessageQueueConfig(),
		RateLimiter:  hub.DefaultRateLimiterConfig(),
	}
	wsHub := hub.New(log, deviceRepo, telemetryRepo, nil, hubConfig)
	PrintStatus("WebSocketHub", "Initialized")

	messageQueue := hub.NewMessageQueue(log, db.DB(), hub.DefaultMessageQueueConfig())
	messageQueue.Start(context.Background())
	wsHub.SetMessageQueue(messageQueue)
	PrintStatus("MessageQueue", "Initialized")

	wsRateLimiter := hub.NewRateLimiter(log, hub.DefaultRateLimiterConfig())
	wsRateLimiter.StartCleanup(context.Background())
	wsHub.SetRateLimiter(wsRateLimiter)
	PrintStatus("WebSocketRateLimiter", "Initialized")

	return authService, deviceService, clientService, commandService, emailSvc,
		fcmNotifier, telemetryRepo, wsHub, messageQueue, wsRateLimiter
}

func initMiddleware() (*middleware.RateLimiter, *middleware.RateLimiter, *middleware.Lockout) {
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	authLimiter := middleware.NewRateLimiter(5, time.Minute)
	PrintStatus("RateLimiters", "Initialized")

	lockoutConfig := middleware.LoadLockoutConfig()
	accountLockout := middleware.NewLockout(lockoutConfig)
	PrintStatus("AccountLockout", fmt.Sprintf("Enabled: %v", lockoutConfig.Enabled))

	return rateLimiter, authLimiter, accountLockout
}

func initSSR(log *slog.Logger) (config.SSRConfig, *ssr.Manager) {
	ssrConfig := config.LoadSSRConfig()
	var ssrManager *ssr.Manager

	if ssrConfig.EnableSSR {
		PrintSection("SSR Configuration")
		PrintStatus("SSR Mode", "Enabled")
		PrintStatus("SSR Auto-Start", strconv.FormatBool(ssrConfig.SSRAutoStart))
		PrintStatus("SSR Auto-Build", strconv.FormatBool(ssrConfig.SSRAutoBuild))
		PrintStatus("SSR URL", ssrConfig.SSRServerURL)

		ssrManager = ssr.NewManager(ssrConfig, log, "", "")
		if err := ssrManager.Start(); err != nil {
			log.Error("SSR server failed to start", "err", err)
			PrintWarning("SSR Server", "Failed, using SPA fallback")
		} else if ssrManager.IsReady() {
			PrintStatus("SSR Server", "Ready")
		}
	} else {
		PrintSection("SSR Configuration")
		PrintStatus("SSR Mode", "Disabled")
	}

	return ssrConfig, ssrManager
}

func createAPIServer(cfg config.Config, log *slog.Logger, db *storage.SQLite,
	operatorRepo *storage.OperatorRepository, sessionManager *infraauth.SessionManager,
	googleVerifier *infraauth.GoogleTokenVerifier, authService *appsvc.AuthService,
	deviceService *device.Service, clientService *client.Service,
	commandService *cmdapp.Service, emailSvc *emailService.Service,
	rateLimiter *middleware.RateLimiter, authLimiter *middleware.RateLimiter,
	accountLockout *middleware.Lockout, wsHub *hub.Hub, fcmNotifier fcm.Notifier) *api.Server {
	PrintSection("Server")
	PrintStatus("Go Server", "Starting on "+cfg.Port)

	appMetrics := metrics.New()
	auditRepo := audit.NewRepository(db.DB())
	auditLogger := audit.NewLogger(auditRepo, log, audit.DefaultLoggerConfig())

	return api.NewServer(&api.ServerConfig{
		AuthService:    authService,
		DeviceService:  deviceService,
		ClientService:  clientService,
		CommandService: commandService,
		RateLimiter:    rateLimiter,
		AuthLimiter:    authLimiter,
		Config:         cfg,
		Log:            log,
		SessionManager:  sessionManager,
		GoogleVerifier:  googleVerifier,
		EmailService:    emailSvc,
		Hub:             wsHub,
		FCMNotifier:     fcmNotifier,
		DB:             db,
		Lockout:         accountLockout,
		OperatorRepo:   operatorRepo,
		Metrics:        appMetrics,
		AuditLogger:    auditLogger,
	})
}

func startServer(port string, log *slog.Logger, apiServer *api.Server,
	ssrConfig config.SSRConfig, ssrManager *ssr.Manager) {
	addr := ":" + port
	server := &http.Server{
		Addr:         addr,
		Handler:      apiServer.Routes(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("starting server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "err", err)
	}

	if ssrConfig.EnableSSR && ssrManager != nil {
		_ = ssrManager.Stop()
		log.Info("SSR server stopped")
	}
}
