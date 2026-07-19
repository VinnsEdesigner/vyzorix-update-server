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
	appoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	diagnosticsapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/logging"
	infrawebhook "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/ssr"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
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

	// Wire up all dependencies using google/wire
	PrintSection("Dependency Injection")
	PrintStatus("Wire", "Injecting dependencies...")

	server, err := wire.Injector(cfg)
	if err != nil {
		log.Error("wire injection failed", "err", err)
		os.Exit(1)
	}

	PrintStatus("Wire", "All dependencies injected")

	// Extract components from wire result
	deps := server.Dependencies
	result := server.Result

	// Initialize SSR if enabled
	ssrConfig := config.LoadSSRConfig()
	var ssrManager *ssr.Manager
	initSSR(log, ssrConfig, &ssrManager)

	// Create API server using wire outputs
	apiServer := api.NewServerWithDeps(&api.ServerConfigWithDeps{
		Config:         cfg,
		Log:            deps.Log,
		DB:             deps.DB,
		Engine:         result.Engine,
		Middleware:     result.MiddlewareSet,
		HandlerSet:     result.HandlerSet,
		SessionManager: deps.SessionManager,
		Hub:            deps.Hub,
		AuditLogger:    deps.AuditLogger,
		UpdatesService: deps.UpdatesService,
	})

	// Register GraphQL if enabled
	if cfg.EnableGraphQL {
		PrintSection("GraphQL")
		PrintStatus("GraphQL", "Registering...")

		// Create GraphQL-specific services not in main wire graph
		db := deps.DB.DB()
		deviceRepo := deps.DeviceService.DeviceRepo()
		commandRepo := deps.CommandService.CommandRepo()

		historyService := command.NewHistoryService(commandRepo, deviceRepo)
		logsRepo := storage.NewLogsRepository(db)
		metricsRepo := storage.NewMetricsRepository(db)
		logsSvc := logs.NewService(logsRepo, deps.Log)
		// Get repositories for hierarchical threshold resolution
		var deviceSettingsRepo device.DeviceSettingsRepository
		var orgSettingsRepo organization.OrganizationSettingsRepository
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

		// Create settings-related services for GraphQL
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
		); regErr != nil {
			log.Error("failed to register GraphQL", "err", regErr)
			PrintWarning("GraphQL", "Registration failed")
		} else {
			PrintStatus("GraphQL", "Registered")
		}
	}

	startServer(cfg.Port, log, apiServer, ssrConfig, ssrManager)
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

func startServer(port string, log *slog.Logger, apiServer *api.Server,
	ssrConfig config.SSRConfig, ssrManager *ssr.Manager) {
	PrintSection("Server")
	PrintStatus("Go Server", "Starting on "+port)

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
