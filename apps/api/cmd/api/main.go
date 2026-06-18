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
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/fcm"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
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

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

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
	var fcmNotifier fcm.Notifier
	if fcmClient, err := fcm.Init(log, cfg.FirebaseCreds); err == nil {
		fcmNotifier = fcmClient
		printStatus("FCM", "Initialized", false)
	} else {
		printWarning("FCM", "Firebase not configured (push notifications disabled)")
	}

	// WebSocket hub for device connections
	wsHub := hub.New(log, nil) // Storage is optional for hub
	printStatus("WebSocketHub", "Initialized", false)

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

	printSection("Server")
	printStatus("Go Server", "Starting on "+cfg.Port, false)
	printStatus("Environment", env, false)

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
//     api.NewServer(config) ─► api.Routes()
//
// The flow: Infrastructure → Domain Interfaces ← Application Services ← API Handlers
