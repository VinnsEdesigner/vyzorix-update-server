// Package main provides the Vyzorix API server entry point.
// This wires up the NEW clean architecture: Domain → Application → Infrastructure.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
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
func printStatus(label, value string, isError bool) {
	color := green
	if isError {
		color = red
	}
	fmt.Printf("  %s✓%s %s: %s%s%s\n", green, reset, label, color, value, reset)
}

// printWarning prints a warning message.
func printWarning(label, value string) {
	fmt.Printf("  %s⚠%s %s: %s%s%s\n", yellow, reset, label, yellow, value, reset)
}

func main() {
	// Print welcome banner
	env := "development"
	if os.Getenv("NODE_ENV") == "production" || os.Getenv("GIN_MODE") == "release" {
		env = "production"
	}
	printBanner(env)

	// Use structured logging with PII redaction
	log := logging.NewFromEnv()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("configuration failed", "err", err)
		os.Exit(1)
	}

	// ============================================================
	// STEP 1: Infrastructure Layer - SQLite + Repositories
	// ============================================================
	printSection("Database")

	sqliteCfg := storage.DefaultConfig(cfg.DatabaseURL)
	db, err := storage.Open(sqliteCfg)
	if err != nil {
		log.Error("database connection failed", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("database close failed", "err", err)
		}
	}()
	printStatus("Database", "Connected to "+cfg.DatabaseURL, false)

	// Create repository instances (implement domain interfaces)
	operatorRepo := storage.NewOperatorRepository(db.DB())
	sessionRepo := storage.NewSessionRepository(db.DB())
	emailVerifyRepo := storage.NewEmailVerificationRepository(db.DB())
	passwordResetRepo := storage.NewPasswordResetRepository(db.DB())
	deviceRepo := storage.NewDeviceRepository(db.DB())
	clientRepo := storage.NewClientRepository(db.DB())
	commandRepo := storage.NewCommandRepository(db.DB())
	printStatus("Repositories", "Initialized (operator, session, device, client, command, email, password_reset)", false)

	// ============================================================
	// STEP 2: Infrastructure - Password Hasher & Session Manager
	// ============================================================
	printSection("Security")

	passwordHasher := infraauthinfra.NewArgon2idHasher()
	printStatus("PasswordHasher", "Argon2id initialized", false)

	sessionManager := infraauth.NewSessionManager(cfg.SessionSecret)
	printStatus("SessionManager", "Initialized", false)

	// Google token verifier for OAuth
	googleVerifier := infraauth.NewGoogleTokenVerifier(cfg.GoogleOAuthClientID)
	printStatus("GoogleVerifier", "Initialized", false)

	// ============================================================
	// STEP 3: Application Layer - Services
	// ============================================================
	printSection("Services")

	sessionTTL := 7 * 24 * time.Hour // 7 days

	authService := appsvc.NewAuthService(
		operatorRepo,
		sessionRepo,
		emailVerifyRepo,
		passwordResetRepo,
		passwordHasher,
		sessionTTL,
	)
	printStatus("AuthService", "Initialized", false)

	deviceService := device.NewService(
		deviceRepo,
		operatorRepo,
	)
	printStatus("DeviceService", "Initialized", false)

	clientService := client.NewService(clientRepo)
	printStatus("ClientService", "Initialized", false)

	commandService := cmdapp.NewService(commandRepo, deviceRepo)
	printStatus("CommandService", "Initialized", false)

	// Email service (optional - only sends if RESEND_API_KEY is configured)
	emailSvc := emailService.NewService()
	if cfg.ResendAPIKey != "" {
		printStatus("EmailService", "Resend configured", false)
	} else {
		printWarning("EmailService", "RESEND_API_KEY not configured (email disabled)")
	}

	// FCM notifier for push notifications (optional - only if FIREBASE_CREDENTIALS is configured)
	// Using EnhancedNotifier with retry logic and metrics
	var fcmNotifier fcm.Notifier
	if fcmClient, err := fcm.Init(log, cfg.FirebaseCreds); err == nil {
		enhancedFCM := fcm.NewEnhancedNotifier(fcmClient, fcm.DefaultFCMConfig())
		fcmNotifier = enhancedFCM
		printStatus("FCM", "Initialized (Enhanced: 3 retries, metrics, token validation)", false)
	} else {
		printWarning("FCM", "Firebase not configured (push notifications disabled)")
	}

	// Create telemetry repository for WebSocket hub
	telemetryRepo := storage.NewTelemetryRepository(db.DB())
	printStatus("TelemetryRepository", "Initialized", false)

	// WebSocket hub configuration with new features
	hubConfig := &hub.HubConfig{
		MessageQueue: hub.DefaultMessageQueueConfig(),
		RateLimiter:  hub.DefaultRateLimiterConfig(),
	}

	// Create WebSocket hub
	wsHub := hub.New(log, deviceRepo, telemetryRepo, nil, hubConfig)
	printStatus("WebSocketHub", "Initialized", false)

	// Initialize message queue for offline command delivery
	messageQueue := hub.NewMessageQueue(log, db.DB(), hub.DefaultMessageQueueConfig())
	messageQueue.Start(context.Background())
	wsHub.SetMessageQueue(messageQueue)
	printStatus("MessageQueue", "Initialized (offline command queue enabled)", false)

	// Initialize rate limiter for WebSocket connections
	wsRateLimiter := hub.NewRateLimiter(log, hub.DefaultRateLimiterConfig())
	wsRateLimiter.StartCleanup(context.Background())
	wsHub.SetRateLimiter(wsRateLimiter)
	printStatus("WebSocketRateLimiter", "Initialized (100 msg/s, 200 burst)", false)

	// ============================================================
	// STEP 4: API Layer - Router + Middleware
	// ============================================================
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	authLimiter := middleware.NewRateLimiter(5, time.Minute) // Stricter for auth endpoints
	printStatus("RateLimiters", "Initialized (100/min general, 5/min auth)", false)

	// Account lockout configuration
	lockoutConfig := middleware.LoadLockoutConfig()
	accountLockout := middleware.NewLockout(lockoutConfig)
	printStatus("AccountLockout", fmt.Sprintf("Enabled: %v, MaxAttempts: %d", lockoutConfig.Enabled, lockoutConfig.MaxAttempts), false)

	// ============================================================
	// STEP 5: SSR Server (Server-Side Rendering)
	// SSR is ENABLED BY DEFAULT - native mode for both dev and prod
	// ============================================================
	ssrConfig := config.LoadSSRConfig()
	var ssrManager *ssr.Manager // Will be nil if SSR is disabled

	if ssrConfig.EnableSSR {
		printSection("SSR Configuration")
		printStatus("SSR Mode", "Enabled (Native SSR mode)", false)
		printStatus("SSR Auto-Start", strconv.FormatBool(ssrConfig.SSRAutoStart), false)
		printStatus("SSR Auto-Build", strconv.FormatBool(ssrConfig.SSRAutoBuild), false)
		printStatus("SSR URL", ssrConfig.SSRServerURL, false)
		printStatus("SSR Timeout", strconv.Itoa(ssrConfig.SSRBuildTimeout)+"s", false)
		printStatus("SSR Retries", strconv.Itoa(ssrConfig.SSRRetryAttempts), false)

		// Create SSR Manager (webDir and publicDir)
		ssrManager = ssr.NewManager(ssrConfig, log, "", "")

		// Start SSR server with auto-build
		if err := ssrManager.Start(); err != nil {
			log.Error("SSR server failed to start", "err", err)
			printWarning("SSR Server", fmt.Sprintf("Failed after %d attempts, using SPA fallback", ssrConfig.SSRRetryAttempts))
		} else if ssrManager.IsReady() {
			printStatus("SSR Server", "Ready on "+ssrConfig.SSRServerURL, false)
			printStatus("Frontend Mode", "SSR (Node.js)", false)
		} else {
			printWarning("SSR Server", "Started but not ready, using SPA fallback")
			printStatus("Frontend Mode", "SPA (Static HTML fallback)", false)
		}
	} else {
		printSection("SSR Configuration")
		printStatus("SSR Mode", "Disabled (SPA fallback only)", false)
		printStatus("Frontend Mode", "SPA (Static HTML)", false)
	}

	printSection("Server")
	printStatus("Go Server", "Starting on "+cfg.Port, false)
	printStatus("Environment", env, false)

	// Initialize metrics
	appMetrics := metrics.New()

	// Initialize audit logger
	auditRepo := audit.NewRepository(db.DB())
	auditLogger := audit.NewLogger(auditRepo, log, audit.DefaultLoggerConfig())

	apiServer := api.NewServer(&api.ServerConfig{
		AuthService:     authService,
		DeviceService:   deviceService,
		ClientService:   clientService,
		CommandService:  commandService,
		RateLimiter:     rateLimiter,
		AuthLimiter:     authLimiter,
		Config:          cfg,
		Log:             log,
		SessionManager:  sessionManager,
		GoogleVerifier:  googleVerifier,
		EmailService:    emailSvc,
		Hub:             wsHub,
		FCMNotifier:     fcmNotifier,
		DB:              db,
		Lockout:         accountLockout,
		OperatorRepo:    operatorRepo,
		Metrics:         appMetrics,
		AuditLogger:     auditLogger,
	})

	// ============================================================
	// STEP 5: HTTP Server
	// ============================================================
	addr := ":" + cfg.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      apiServer.Routes(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
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

	// Stop SSR Manager (handles graceful shutdown of Node.js subprocess)
	if ssrConfig.EnableSSR && ssrManager != nil {
		_ = ssrManager.Stop()
		log.Info("SSR server stopped")
	}
}

// DEPENDENCY WIRING DIAGRAM (Clean Architecture):
//
//   main()
//     │
//     ├─► infrastructure/storage.Open() ─► SQLite (DB connection)
//     │                                        │
//     │   ┌─────────────────────────────────────┴─────────────────────────────────────┐
//     │   │                                                                         │
//     │   ▼                                                                         ▼
//     │   storage.NewOperatorRepository() ◄────────────── storage.NewSessionRepository()
//     │   (implements operator.Repository)                 (implements session.Repository)
//     │                                                                         │
//     │   storage.NewEmailVerificationRepository() ◄── storage.NewPasswordResetRepository()
//     │   (implements email_verification.Repository)      (implements password_reset.Repository)
//     │                                                                         │
//     │   storage.NewDeviceRepository() ◄────────────── storage.NewCommandRepository()
//     │                                                                         │
//     └─────────────────────────────────────────────────────────────────────────────┘
//
//     ▼
//     infraauth.NewArgon2idHasher() ─► implements auth.PasswordHasher interface
//
//     ▼
//     auth.NewAuthService(
//       operatorRepo,
//       sessionRepo,
//       emailVerifyRepo,
//       passwordResetRepo,
//       passwordHasher,
//       sessionTTL,
//     )
//
//     ▼
//     device.NewService(deviceRepo, operatorRepo)
//
//     ▼
//     ssr.NewManager(ssrConfig, log) ─► Node.js SSR subprocess (auto-builds web app)
//         │
//         ├─► WebPublicDir: served static files
//         ├─► WebDistDir: SSR build output
//         └─► Health monitoring + auto-recovery
//
//     ▼
//     api.NewServer(config) ─► api.Routes() + SSRProxy middleware
//         │
//         └─► SSRProxy: proxies /dashboard/* to Node.js SSR, falls back to static HTML
//
// The flow: Infrastructure → Domain Interfaces ← Application Services ← API Handlers
//           └────────────────────────── SSR ────────────────────────────────┘
