package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	controllers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/fcm"
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

func main() {
	// Print welcome banner
	env := "development"
	if os.Getenv("NODE_ENV") == "production" || os.Getenv("GIN_MODE") == "release" {
		env = "production"
	}
	printBanner(env)

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		log.Error("configuration failed", "err", err)
		os.Exit(1)
	}
	if dir := filepath.Dir(cfg.DatabaseURL); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Error("failed to create database directory", "dir", dir, "err", err)
		}
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Error("failed to create data directory", "dir", cfg.DataDir, "err", err)
	}
	if err := os.MkdirAll(cfg.BinDir, 0o755); err != nil {
		log.Error("failed to create bin directory", "dir", cfg.BinDir, "err", err)
	}
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

	notifier, err := fcm.Init(log, cfg.FirebaseCreds)
	if err != nil {
		log.Error("fcm init failed", "err", err)
		os.Exit(1)
	}

	h := hub.New(log, st)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go h.Run(ctx)

	srv := controllers.New(log, cfg, st, h, notifier)
	addr := ":" + cfg.Port
	log.Info("vyzorix update server starting", "addr", addr, "db", cfg.DatabaseURL, "env", cfg.Env, "enforceHMAC", cfg.EnforceHMAC)

	go func() {
		if err := http.ListenAndServe(addr, srv.Routes()); err != nil && err != http.ErrServerClosed {
			_, _ = fmt.Fprintf(os.Stderr, "server failed: %v\n", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("server shutting down")
}
