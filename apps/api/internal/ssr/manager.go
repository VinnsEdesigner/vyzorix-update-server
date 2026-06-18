// Package ssr provides modular SSR server management components.
package ssr

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
)

// Manager orchestrates all SSR components.
type Manager struct {
	config    config.SSRConfig
	logger    *slog.Logger
	process   *ProcessManager
	monitor   *Monitor
	builder   *Builder
	started   bool
	stopped   bool
	ssrScript string
	mu        sync.RWMutex
}

// NewManager creates a new SSR manager with the given configuration.
// If webDir or publicDir is empty, they will be auto-discovered.
func NewManager(cfg config.SSRConfig, logger *slog.Logger, webDir, publicDir string) *Manager {
	// Discover web directory if not provided.
	if webDir == "" {
		webDir = discoverWebDir()
	}

	// Calculate public dir relative to api.
	if publicDir == "" {
		wd, _ := os.Getwd()
		publicDir = filepath.Join(wd, "public")
	}

	// Discover SSR script path.
	ssrScript := filepath.Join(filepath.Dir(os.Args[0]), "ssr-server.js")
	//nolint:gosec // G703 - path is validated and falls back to safe default.
	if _, err := os.Stat(ssrScript); os.IsNotExist(err) {
		ssrScript = "./ssr-server.js"
	}

	// Create internal config for process manager.
	internalCfg := Config{
		EnableSSR:              cfg.EnableSSR,
		SSRServerURL:           cfg.SSRServerURL,
		SSRPort:                cfg.SSRPort,
		SSRAutoStart:           cfg.SSRAutoStart,
		SSRAutoBuild:           cfg.SSRAutoBuild,
		SSRBuildTimeout:        cfg.SSRBuildTimeout,
		SSRHealthCheckInterval: cfg.SSRHealthCheckInterval,
		SSRRetryAttempts:       cfg.SSRRetryAttempts,
		SSRRetryBackoff:       2, // Default backoff
	}

	process := NewProcessManager(internalCfg, logger)
	monitor := NewMonitor(process, logger, time.Duration(cfg.SSRHealthCheckInterval)*time.Second)
	builder := NewBuilder(logger, webDir, publicDir)

	return &Manager{
		config:    cfg,
		logger:    logger,
		process:   process,
		monitor:   monitor,
		builder:   builder,
		ssrScript: ssrScript,
	}
}

// discoverWebDir attempts to find the web app directory.
func discoverWebDir() string {
	// Try common locations relative to executable.
	exeDir, _ := os.Executable()
	if exeDir != "" {
		// Try ../web from api directory.
		webDir := filepath.Join(filepath.Dir(exeDir), "..", "web")
		if _, err := os.Stat(webDir); err == nil {
			return webDir
		}
	}

	// Try ./web from current directory.
	if _, err := os.Stat("./web"); err == nil {
		return "./web"
	}

	// Fallback to ./web.
	return "./web"
}

// Config returns the SSR configuration.
func (m *Manager) Config() config.SSRConfig {
	return m.config
}

// IsEnabled returns whether SSR is enabled.
func (m *Manager) IsEnabled() bool {
	return m.config.EnableSSR
}

// IsReady returns whether the SSR process is ready.
func (m *Manager) IsReady() bool {
	return m.process.IsReady()
}

// State returns the current process state.
func (m *Manager) State() ProcessState {
	return m.process.State()
}

// HealthCheck returns the current health status.
func (m *Manager) HealthCheck() *HealthStatus {
	return m.process.HealthCheck()
}

// Start starts the SSR server with all configured options.
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.EnableSSR {
		m.logger.Info("SSR is disabled")
		return nil
	}

	if m.started && !m.stopped {
		m.logger.Info("SSR already started")
		return nil
	}

	m.logger.Info("Starting SSR manager",
		"autoStart", m.config.SSRAutoStart,
		"autoBuild", m.config.SSRAutoBuild,
		"retries", m.config.SSRRetryAttempts,
	)

	// Auto-build if enabled.
	if m.config.SSRAutoBuild {
		if err := m.builder.BuildIfNeeded(); err != nil {
			m.logger.Error("SSR auto-build failed", "err", err)
			if !m.config.SSRAutoStart {
				return nil // Not required to start
			}
			return err
		}
	}

	// Start process if enabled.
	if m.config.SSRAutoStart {
		if err := m.process.Start(m.ssrScript); err != nil {
			m.logger.Error("SSR process start failed", "err", err)
			return err
		}

		// Start monitoring with auto-restart callback.
		ctx, cancel := context.WithCancel(context.Background())
		m.monitor.SetCancel(cancel)
		m.monitor.SetRestartCallback(func() {
			m.mu.RLock()
			if m.stopped {
				m.mu.RUnlock()
				return
			}
			// Restart the process (stops existing and starts fresh).
			m.logger.Info("Auto-restarting SSR process...")
			if err := m.process.Restart(m.ssrScript); err != nil {
				m.logger.Error("SSR auto-restart failed", "err", err)
			}
			m.mu.RUnlock()
		})
		m.monitor.Start(ctx)
	}

	m.started = true
	m.stopped = false
	return nil
}

// Stop stops the SSR server gracefully.
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Stopping SSR manager")
	m.stopped = true

	// Stop monitoring first (this will cancel the context).
	m.monitor.Stop()

	// Stop process.
	if err := m.process.Stop(); err != nil {
		m.logger.Error("SSR process stop failed", "err", err)
		return err
	}

	m.started = false
	return nil
}

// String returns a string representation of the SSR manager status.
func (m *Manager) String() string {
	health := m.HealthCheck()
	return fmt.Sprintf("SSR{enabled=%v state=%s ready=%v healthy=%v}",
		m.config.EnableSSR, health.State, health.Ready, health.Healthy)
}
