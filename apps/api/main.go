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

	"github.com/VinnsEdesigner/vyzorix/apps/api/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/controllers"
	"github.com/VinnsEdesigner/vyzorix/apps/api/hub"
	"github.com/VinnsEdesigner/vyzorix/apps/api/services/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/storage"
)

func main() {
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
