// Package main provides the Vyzorix API server with integrated SSR management.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	controllers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/fcm"
	ssrmod "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ssr"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
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

	ssrConfig := config.LoadSSRConfig()

	// Ensure required directories exist
	printSection("Initialization")

	dirs := []struct {
		name string
		dir  string
	}{
		{"Database", filepath.Dir(cfg.DatabaseURL)},
		{"Data", cfg.DataDir},
		{"Bin", cfg.BinDir},
		{"Public", cfg.PublicDir},
	}

	for _, d := range dirs {
		if d.dir != "." && d.dir != "" {
			//nolint:gosec // G301 - directory permissions are configurable and intentionally open
			if err := os.MkdirAll(d.dir, 0o755); err != nil {
				log.Error("failed to create directory", "name", d.name, "dir", d.dir, "err", err)
			} else {
				printStatus(d.name+" Directory", d.dir, false)
			}
		}
	}

	// Initialize database
	printSection("Database")
	st, err := storage.Open(cfg.DatabaseURL)
	if err != nil {
		log.Error("database init failed", "err", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Error("database close failed", "err", closeErr)
		}
	}()
	printStatus("Database", cfg.DatabaseURL, false)

	// Initialize FCM (optional - may fail without credentials)
	notifier, err := fcm.Init(log, cfg.FirebaseCreds)
	if err != nil {
		printWarning("FCM", "Firebase not configured (push notifications disabled)")
	} else {
		printStatus("FCM", "Initialized", false)
	}

	// Initialize WebSocket hub
	h := hub.New(log, st)

	// Setup signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go h.Run(ctx)

	// Initialize SSR Manager for modular SSR management.
	ssrManager := ssrmod.NewManager(ssrConfig, log, "", "")

	if ssrManager.IsEnabled() {
		printSection("SSR Configuration")
		printStatus("SSR Mode", "Enabled (SSR default)", false)
		printStatus("SSR Auto-Start", strconv.FormatBool(ssrManager.Config().SSRAutoStart), false)
		printStatus("SSR Auto-Build", strconv.FormatBool(ssrManager.Config().SSRAutoBuild), false)
		printStatus("SSR URL", ssrManager.Config().SSRServerURL, false)
		printStatus("SSR Timeout", strconv.Itoa(ssrManager.Config().SSRBuildTimeout)+"s", false)
		printStatus("SSR Retries", strconv.Itoa(ssrManager.Config().SSRRetryAttempts), false)

		// Start SSR with auto-build and monitoring.
		if err := ssrManager.Start(); err != nil {
			log.Error("SSR server failed to start", "err", err)
			printWarning("SSR Server", fmt.Sprintf("Failed after %d attempts, using SPA fallback", ssrManager.Config().SSRRetryAttempts))
		} else if ssrManager.IsReady() {
			printStatus("SSR Server", "Ready on "+ssrManager.Config().SSRServerURL, false)
		} else {
			printWarning("SSR Server", "Started but not ready, using SPA fallback")
		}
	} else {
		printSection("SSR Configuration")
		printStatus("SSR Mode", "Disabled (SPA fallback only)", false)
	}

	// Create HTTP server
	srv := controllers.New(log, cfg, st, h, notifier)
	addr := ":" + cfg.Port

	printSection("Server")
	printStatus("Go Server", "Starting on "+addr, false)
	printStatus("Environment", env, false)
	if ssrManager.IsEnabled() && ssrManager.IsReady() {
		printStatus("Frontend Mode", "SSR (Node.js)", false)
	} else {
		printStatus("Frontend Mode", "SPA (Static)", false)
	}

	log.Info("vyzorix update server starting",
		"addr", addr,
		"db", cfg.DatabaseURL,
		"env", cfg.Env,
		"enforceHMAC", cfg.EnforceHMAC,
		"ssrEnabled", ssrManager.IsEnabled(),
		"ssrReady", ssrManager.IsReady(),
	)

	// Start Go HTTP server
	server := &http.Server{
		Addr:         addr,
		Handler:      srv.Routes(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			_, _ = fmt.Fprintf(os.Stderr, "server failed: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Info("server shutting down")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", "err", err)
	}

	// Stop SSR Manager (handles monitoring and subprocess gracefully)
	_ = ssrManager.Stop()

	log.Info("server stopped", "ssr", ssrManager.String())
}
