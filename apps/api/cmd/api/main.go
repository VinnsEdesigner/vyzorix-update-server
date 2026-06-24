// Package main provides the Vyzorix API server entry point.
// This wires up the NEW clean architecture: Domain → Application → Infrastructure.
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
	infraauthinfra "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/auth"
	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/logging"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/ssr"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
)

// ANSI color codes for terminal output.
const (
	cyan    = "\033[36m"
	magenta = "\033[35m"
	yellow  = "\033[33m"
	red     = "\033[31m"
	green   = "\033[32m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	reset   = "\033[0m"
)

// printBanner prints the VYZORIX ASCII art banner.
func printBanner(mode string) {
	banner := []string{
		magenta + bold + "+-------------------------------------------------------------+" + reset,
		magenta + bold + "|   _   _           _        ____                           |" + reset,
		magenta + bold + "|  |_| |_|   ___   | |__    |  _|  ___  ___                 |" + reset,
		magenta + bold + "|  | | | |  / _ \\  | '_ \\  | |_  / _ \\/ __|                |" + reset,
		magenta + bold + "|  | |_| | | (_) | | |_) | |  _|  __/\\__ \\                |" + reset,
		magenta + bold + "|  |___|_|  \\___/  |_.__/   |_|   \\___||___/               |" + reset,
		magenta + bold + "|                                                              |" + reset,
		magenta + bold + "|                    GOLANG SERVER v1.0.0                      |" + reset,
		magenta + bold + "+-------------------------------------------------------------+" + reset,
	}

	for _, line := range banner {
		fmt.Println(line)
	}

	// Print mode indicator
	modeColor := yellow
	if mode == "production" {
		modeColor = red
	}
	fmt.Printf("  %sMode:%s %s%s[%s]%s\n", dim, reset, modeColor, bold, mode, reset)
	fmt.Printf("  %s%s\n", dim, "============================================================")
}

// printSection prints a section header.
func printSection(label string) {
	fmt.Printf("\n%s[%s]%s\n", cyan, label, reset)
}

// printStatus prints a status line with color coding.
func printStatus(label, value string, _ bool) {
	fmt.Printf("  %s✓%s %s: %s%s%s\n", green, reset, label, green, value, reset)
}

// printWarning prints a warning message.
func printWarning(label, value string) {
	fmt.Printf("  %s⚠%s %s: %s%s%s\n", yellow, reset, label, yellow, value, reset)
}

func main() {
	env := getEnv()
	printBanner(env)

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
		if regErr := apiServer.RegisterGraphQL(deviceService, commandService, telemetryRepo, wsHub); regErr != nil {
			log.Error("failed to register GraphQL", "err", regErr)
		}
	}

	startServer(cfg.Port, log, apiServer, ssrConfig, ssrManager)
}

func getEnv() string {
	if os.Getenv("NODE_ENV") == "production" || os.Getenv("GIN_MODE") == "release" {
		return "production"
	}
	return "development"
}

func initDatabase(cfg config.Config) (*storage.SQLite, error) {
	printSection("Database")

	sqliteCfg := storage.DefaultConfig(cfg.DatabaseURL)
	db, err := storage.Open(sqliteCfg)
	if err != nil {
		return nil, err
	}
	printStatus("Database", "Connected to "+cfg.DatabaseURL, false)

	printStatus("Repositories", "Initialized", false)

	return db, nil
}

func initSecurity(cfg config.Config) (*infraauthinfra.Argon2idHasher, *infraauth.SessionManager, *infraauth.GoogleTokenVerifier) {
	printSection("Security")

	passwordHasher := infraauthinfra.NewArgon2idHasher()
	printStatus("PasswordHasher", "Argon2id initialized", false)

	sessionManager := infraauth.NewSessionManager(cfg.SessionSecret)
	printStatus("SessionManager", "Initialized", false)

	googleVerifier := infraauth.NewGoogleTokenVerifier(cfg.GoogleOAuthClientID)
	printStatus("GoogleVerifier", "Initialized", false)

	return passwordHasher, sessionManager, googleVerifier
}

func initServices(cfg config.Config, log *slog.Logger, db *storage.SQLite,
	deviceRepo *storage.DeviceRepository, operatorRepo *storage.OperatorRepository,
	commandRepo *storage.CommandRepository) (*appsvc.AuthService, *device.Service, *client.Service,
	*cmdapp.Service, *emailService.Service, fcm.Notifier, *storage.TelemetryRepository,
	*hub.Hub, *hub.MessageQueue, *hub.RateLimiter) {
	printSection("Services")

	sessionTTL := 7 * 24 * time.Hour

	authService := appsvc.NewAuthService(
		operatorRepo,
		storage.NewSessionRepository(db.DB()),
		storage.NewEmailVerificationRepository(db.DB()),
		storage.NewPasswordResetRepository(db.DB()),
		infraauthinfra.NewArgon2idHasher(),
		sessionTTL,
	)
	printStatus("AuthService", "Initialized", false)

	deviceService := device.NewService(deviceRepo, operatorRepo)
	printStatus("DeviceService", "Initialized", false)

	clientService := client.NewService(storage.NewClientRepository(db.DB()))
	printStatus("ClientService", "Initialized", false)

	commandService := cmdapp.NewService(commandRepo, deviceRepo)
	printStatus("CommandService", "Initialized", false)

	emailSvc := emailService.NewService()
	if cfg.ResendAPIKey != "" {
		printStatus("EmailService", "Resend configured", false)
	} else {
		printWarning("EmailService", "RESEND_API_KEY not configured")
	}

	var fcmNotifier fcm.Notifier
	if fcmClient, err := fcm.Init(log, cfg.FirebaseCreds); err == nil {
		fcmNotifier = fcm.NewEnhancedNotifier(fcmClient, fcm.DefaultFCMConfig())
		printStatus("FCM", "Initialized", false)
	} else {
		printWarning("FCM", "Firebase not configured")
	}

	telemetryRepo := storage.NewTelemetryRepository(db.DB())
	printStatus("TelemetryRepository", "Initialized", false)

	hubConfig := &hub.HubConfig{
		MessageQueue: hub.DefaultMessageQueueConfig(),
		RateLimiter:  hub.DefaultRateLimiterConfig(),
	}
	wsHub := hub.New(log, deviceRepo, telemetryRepo, nil, hubConfig)
	printStatus("WebSocketHub", "Initialized", false)

	messageQueue := hub.NewMessageQueue(log, db.DB(), hub.DefaultMessageQueueConfig())
	messageQueue.Start(context.Background())
	wsHub.SetMessageQueue(messageQueue)
	printStatus("MessageQueue", "Initialized", false)

	wsRateLimiter := hub.NewRateLimiter(log, hub.DefaultRateLimiterConfig())
	wsRateLimiter.StartCleanup(context.Background())
	wsHub.SetRateLimiter(wsRateLimiter)
	printStatus("WebSocketRateLimiter", "Initialized", false)

	return authService, deviceService, clientService, commandService, emailSvc,
		fcmNotifier, telemetryRepo, wsHub, messageQueue, wsRateLimiter
}

func initMiddleware() (*middleware.RateLimiter, *middleware.RateLimiter, *middleware.Lockout) {
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	authLimiter := middleware.NewRateLimiter(5, time.Minute)
	printStatus("RateLimiters", "Initialized", false)

	lockoutConfig := middleware.LoadLockoutConfig()
	accountLockout := middleware.NewLockout(lockoutConfig)
	printStatus("AccountLockout", fmt.Sprintf("Enabled: %v", lockoutConfig.Enabled), false)

	return rateLimiter, authLimiter, accountLockout
}

func initSSR(log *slog.Logger) (config.SSRConfig, *ssr.Manager) {
	ssrConfig := config.LoadSSRConfig()
	var ssrManager *ssr.Manager

	if ssrConfig.EnableSSR {
		printSection("SSR Configuration")
		printStatus("SSR Mode", "Enabled", false)
		printStatus("SSR Auto-Start", strconv.FormatBool(ssrConfig.SSRAutoStart), false)
		printStatus("SSR Auto-Build", strconv.FormatBool(ssrConfig.SSRAutoBuild), false)
		printStatus("SSR URL", ssrConfig.SSRServerURL, false)

		ssrManager = ssr.NewManager(ssrConfig, log, "", "")
		if err := ssrManager.Start(); err != nil {
			log.Error("SSR server failed to start", "err", err)
			printWarning("SSR Server", "Failed, using SPA fallback")
		} else if ssrManager.IsReady() {
			printStatus("SSR Server", "Ready", false)
		}
	} else {
		printSection("SSR Configuration")
		printStatus("SSR Mode", "Disabled", false)
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
	printSection("Server")
	printStatus("Go Server", "Starting on "+cfg.Port, false)

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
