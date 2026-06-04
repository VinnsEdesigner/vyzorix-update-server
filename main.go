package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/VinnsEdesigner/vyzorix-update-server/config"
	"github.com/VinnsEdesigner/vyzorix-update-server/controllers"
	"github.com/VinnsEdesigner/vyzorix-update-server/hub"
	"github.com/VinnsEdesigner/vyzorix-update-server/services/fcm"
	"github.com/VinnsEdesigner/vyzorix-update-server/storage"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		log.Error("configuration failed", "err", err)
		os.Exit(1)
	}
	if dir := filepath.Dir(cfg.DatabaseURL); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	_ = os.MkdirAll(cfg.DataDir, 0o755)
	_ = os.MkdirAll(cfg.BinDir, 0o755)
	st, err := storage.Open(cfg.DatabaseURL)
	if err != nil {
		log.Error("database init failed", "err", err)
		os.Exit(1)
	}
	defer st.Close()
	notifier, err := fcm.Init(log, cfg.FirebaseCreds)
	if err != nil {
		log.Error("fcm init failed", "err", err)
		os.Exit(1)
	}
	h := hub.New(log, st)
	srv := controllers.New(log, cfg, st, h, notifier)
	addr := ":" + cfg.Port
	log.Info("vyzorix real update server starting", "addr", addr, "db", cfg.DatabaseURL, "env", cfg.Env, "enforceHMAC", cfg.EnforceHMAC)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil && err != http.ErrServerClosed {
		_, _ = fmt.Fprintf(os.Stderr, "server failed: %v\n", err)
		os.Exit(1)
	}
}
