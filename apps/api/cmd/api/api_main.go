// Package main provides the Vyzorix API server entry point.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/wire"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	diagnosticsapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	appoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/provisioning"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	notificationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
	orgdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/logging"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/serverlock"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/ssr"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	infrawebhook "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/worker"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env.example file for default production configuration.
	// Try multiple paths: Docker (/app), local dev from apps/api (../../), local from root (.)
	for _, envPath := range []string{".env.example", "../../.env.example", "../.env.example"} {
		if err := godotenv.Load(envPath); err == nil {
			break
		}
	}

	// Determine environment mode from environment (before config load).
	envMode := getEnv()

	// Load config first to get the actual environment mode.
	cfg, err := config.Load()
	if err != nil {
		// Print banner with detected environment mode before exiting.
		PrintBanner(envMode)
		log := logging.NewFromEnv()
		log.Error("configuration failed", "err", err)
		os.Exit(1)
	}

	// Use config's Env if available, otherwise use detected mode.
	if cfg.Env != "" {
		envMode = cfg.Env
	}
	PrintBanner(envMode)

	log := logging.NewFromEnv()

	// Wire up all dependencies using google/wire.
	PrintSection("Dependency Injection")
	PrintStatus("Wire", "Injecting dependencies...")

	server, err := wire.Injector(cfg)
	if err != nil {
		log.Error("wire injection failed", "err", err)
		os.Exit(1)
	}

	PrintStatus("Wire", "All dependencies injected")

	// Extract components from wire result.
	deps := server.Dependencies
	result := server.Result

	// Initialize SSR if enabled.
	ssrConfig := config.LoadSSRConfig()
	var ssrManager *ssr.Manager
	initSSR(log, ssrConfig, &ssrManager)

	// Create API server using wire outputs.
	apiServer := api.NewServerWithDeps(&api.ServerConfigWithDeps{
		Config:          cfg,
		Log:             deps.Log,
		DB:              deps.DB,
		Engine:          result.Engine,
		Middleware:      result.MiddlewareSet,
		HandlerSet:      result.HandlerSet,
		SessionManager:  deps.SessionManager,
		Hub:             deps.Hub,
		AuditLogger:     deps.AuditLogger,
		UpdatesService:  deps.UpdatesService,
		APIKeyService:   deps.APIKeyService,
		DeviceRepo:      deps.DeviceRepo,
		IdempotencyRepo: deps.IdempotencyRepo,
	})

	// Register GraphQL if enabled.
	if cfg.EnableGraphQL {
		PrintSection("GraphQL")
		PrintStatus("GraphQL", "Registering...")

		// Create GraphQL-specific services not in main wire graph.
		db := deps.DB.DB()
		deviceRepo := deps.DeviceService.DeviceRepo()
		commandRepo := deps.CommandService.CommandRepo()

		historyService := command.NewHistoryService(commandRepo, deviceRepo)
		logsRepo := storage.NewLogsRepository(db)
		metricsRepo := storage.NewMetricsRepository(db)
		logsSvc := logs.NewService(logsRepo, deps.Log)
		// Get repositories for hierarchical threshold resolution.
		var deviceSettingsRepo devicedomain.DeviceSettingsRepository
		var orgSettingsRepo orgdomain.OrganizationSettingsRepository
		if deps.DeviceSettingsService != nil {
			deviceSettingsRepo = deps.DeviceSettingsService.SettingsRepo()
		}
		if deps.OrgSettingsService != nil {
			orgSettingsRepo = deps.OrgSettingsService.SettingsRepo()
		}
		metricsSvc := appmetrics.NewService(metricsRepo, deviceSettingsRepo, orgSettingsRepo)
		dashboardSvc := dashboard.NewService(deviceRepo, commandRepo, logsRepo)
		diagnosticsRepo := storage.NewDiagnosticsRepository(db)
		diagnosticsSvc := diagnosticsapp.NewService(diagnosticsRepo, deviceRepo, deps.Hub, cfg.DiagnosticsConfig)

		// Create settings-related services for GraphQL.
		settingsService := auth.NewClientSettingsService(deps.OperatorRepo)
		notificationSvc := appoperator.NewNotificationService(deps.OperatorRepo)
		webhookClient := infrawebhook.NewClient(10 * time.Second)

		if regErr := apiServer.RegisterGraphQL(
			deps.DeviceService,
			deps.DeviceSettingsService,
			deps.CommandService,
			historyService,
			dashboardSvc,
			logsSvc,
			metricsSvc,
			deps.TelemetryRepo,
			logsRepo,
			metricsRepo,
			deps.Hub,
			deps.UpdatesService,
			diagnosticsSvc,
			deps.OperatorRepo,
			settingsService,
			notificationSvc,
			webhookClient,
			result.HandlerSet.OrgService,
			result.HandlerSet.OrgSettingsService,
			result.HandlerSet.MemberService,
			result.HandlerSet.InvitationService,
			apiServer.ContactPointHandler.Service(),
			apiServer.ContactPointHandler.Dispatcher(),
			apiServer.ServiceAccountHandler.Service(),
		); regErr != nil {
			log.Error("failed to register GraphQL", "err", regErr)
			PrintWarning("GraphQL", "Registration failed")
		} else {
			PrintStatus("GraphQL", "Registered")
		}
	}

	startServer(&cfg, log, apiServer, ssrConfig, ssrManager)
}

func initSSR(log *slog.Logger, ssrConfig config.SSRConfig, ssrManager **ssr.Manager) {
	if ssrConfig.EnableSSR {
		PrintSection("SSR Configuration")
		PrintStatus("SSR Mode", "Enabled")
		PrintStatus("SSR Auto-Start", strconv.FormatBool(ssrConfig.SSRAutoStart))
		PrintStatus("SSR Auto-Build", strconv.FormatBool(ssrConfig.SSRAutoBuild))
		PrintStatus("SSR URL", ssrConfig.SSRServerURL)

		*ssrManager = ssr.NewManager(ssrConfig, log, "", "")
		if err := (*ssrManager).Start(); err != nil {
			log.Error("SSR server failed to start", "err", err)
			PrintWarning("SSR Server", "Failed, using SPA fallback")
		} else if (*ssrManager).IsReady() {
			PrintStatus("SSR Server", "Ready")
		}
	} else {
		PrintSection("SSR Configuration")
		PrintStatus("SSR Mode", "Disabled")
	}
}

// bootstrapSlackContactPoint creates a Slack contact point at boot when
// SLACK_WEBHOOK_URL is set. Idempotent via name/channel check.
func bootstrapSlackContactPoint(log *slog.Logger, db *storage.SQLite, slackCfg config.SlackConfig) {
	if db == nil || slackCfg.WebhookURL == "" {
		return
	}

	repo := storage.NewContactPointRepository(db.DB())
	ctx := context.Background()

	// Resolve org: SLACK_ORG_NAME if set, otherwise first org.
	var orgID string
	if slackCfg.OrgName != "" {
		row := db.DB().QueryRowContext(ctx, "SELECT id FROM organizations WHERE name = ?", slackCfg.OrgName)
		if err := row.Scan(&orgID); err != nil {
			log.Warn("org for slack contact point not found; create it via provisioning first", "org", slackCfg.OrgName)
			return
		}
	} else {
		row := db.DB().QueryRowContext(ctx, "SELECT id FROM organizations ORDER BY created_at LIMIT 1")
		if err := row.Scan(&orgID); err != nil {
			log.Warn("no orgs exist yet; slack contact point not created")
			return
		}
	}

	existing, err := repo.ListByOrg(ctx, orgID)
	if err != nil {
		log.Error("failed to list contact points for slack bootstrap", "error", err)
		return
	}
	for _, cp := range existing {
		if cp.Name == "slack-alerts" && cp.Channel == notificationdomain.ChannelTypeSlack {
			log.Info("slack contact point already exists, skipping")
			return
		}
	}

	point := &notificationdomain.ContactPoint{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		Name:      "slack-alerts",
		Channel:   notificationdomain.ChannelTypeSlack,
		Config:    map[string]string{"webhook": slackCfg.WebhookURL},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if slackCfg.Channel != "" {
		point.Config["channel"] = slackCfg.Channel
	}
	if slackCfg.Username != "" {
		point.Config["username"] = slackCfg.Username
	}

	if err := point.Validate(); err != nil {
		log.Error("slack contact point invalid", "error", err)
		return
	}
	if err := repo.Save(ctx, point); err != nil {
		log.Error("failed to save slack contact point", "error", err)
		return
	}
	log.Info("slack contact point registered", "org", orgID)
}

//nolint:gocyclo // startServer wires many optional services.
func startServer(cfg *config.Config, log *slog.Logger, apiServer *api.Server,
	ssrConfig config.SSRConfig, ssrManager *ssr.Manager) {
	PrintSection("Server")
	PrintStatus("Go Server", "Starting on "+cfg.Port)

	addr := ":" + cfg.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      apiServer.Routes(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start device deletion worker.
	var lockSvc *serverlock.Service
	if apiServer.DB != nil {
		lockSvc = serverlock.NewService(apiServer.DB.DB())
	}

	deviceDeletionWorker := worker.NewDeviceDeletionWorker(
		apiServer.DeviceRepo,
		lockSvc,
		log,
		1*time.Hour, // Run every hour.
	)
	deviceDeletionWorker.Start()

	// Start service account token expiry sweep (revokes tokens past expiry hourly).
	var saExpiryWorker *worker.ServiceAccountExpiryWorker
	if apiServer.DB != nil {
		saTokenRepo := storage.NewServiceAccountTokenRepository(apiServer.DB.DB())
		saExpiryWorker = worker.NewServiceAccountExpiryWorker(saTokenRepo, lockSvc, log, time.Hour)
		saExpiryWorker.Start()
	}

	// Start leak scan worker (scans public GitHub/GitLab code search daily).
	var saLeakWorker *worker.ServiceAccountLeakWorker
	if apiServer.DB != nil && apiServer.ServiceAccountHandler != nil {
		saLeakWorker = worker.NewServiceAccountLeakWorker(apiServer.ServiceAccountHandler.Service(), lockSvc, log, 24*time.Hour)
		saLeakWorker.Start()
	}

	// Start invitation cleanup worker (expires stale pending invitations every 10 minutes).
	var inviteCleanupWorker *worker.InvitationCleanupWorker
	if apiServer.InvitationStorage != nil {
		inviteCleanupWorker = worker.NewInvitationCleanupWorker(apiServer.InvitationStorage, lockSvc, log, 10*time.Minute)
		inviteCleanupWorker.Start()
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Bootstrap Slack contact point from SLACK_* env vars so alerts reach
	// Slack without manual provisioning.
	if cfg.SlackConfig.WebhookURL != "" {
		bootstrapSlackContactPoint(log, apiServer.DB, cfg.SlackConfig)
	}

	// Run provisioning if a provisioning file is configured.
	if provFile := os.Getenv("PROVISIONING_FILE"); provFile != "" {
		prov := provisioning.New(log)
		if apiServer.DB != nil {
			prov = prov.WithAlertRepository(storage.NewAlertRuleRepository(apiServer.DB.DB()))
			prov = prov.WithContactPointRepository(storage.NewContactPointRepository(apiServer.DB.DB()))
			prov = prov.WithServiceAccountRepository(storage.NewServiceAccountRepository(apiServer.DB.DB()))
		}
		if err := prov.LoadAndApply(context.Background(), provFile); err != nil {
			log.Error("provisioning failed", "err", err)
		} else {
			log.Info("provisioning complete")
		}
	}

	go func() {
		log.Info("starting server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	log.Info("shutting down")

	// Stop background workers first - they may be generating new requests.
	if saLeakWorker != nil {
		saLeakWorker.Stop()
	}
	if saExpiryWorker != nil {
		saExpiryWorker.Stop()
	}
	if inviteCleanupWorker != nil {
		inviteCleanupWorker.Stop()
	}
	deviceDeletionWorker.Stop()

	// Graceful drain: give in-flight requests time to complete.
	// Use a longer timeout for server shutdown to ensure graceful drain.
	drainCtx, drainCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer drainCancel()

	// Shutdown the HTTP server gracefully - this stops accepting new connections.
	// and waits for existing connections to finish their requests.
	if err := server.Shutdown(drainCtx); err != nil {
		log.Error("server graceful shutdown error", "err", err)
	} else {
		log.Info("server gracefully drained in-flight requests")
	}

	// Short timeout for remaining services.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown AuditLogger to flush pending audit log writes.
	if apiServer.AuditLogger != nil {
		if err := apiServer.AuditLogger.Shutdown(shutdownCtx); err != nil {
			log.Error("audit logger shutdown error", "err", err)
		} else {
			log.Info("audit logger shutdown complete")
		}
	}

	// Shutdown InvitationService to flush pending email goroutines.
	if apiServer.InvitationService != nil {
		if err := apiServer.InvitationService.Shutdown(shutdownCtx); err != nil {
			log.Error("invitation service shutdown error", "err", err)
		} else {
			log.Info("invitation service shutdown complete")
		}
	}

	if ssrConfig.EnableSSR && ssrManager != nil {
		_ = ssrManager.Stop()
		log.Info("SSR server stopped")
	}
}
