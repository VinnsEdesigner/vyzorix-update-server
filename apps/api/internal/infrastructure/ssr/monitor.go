// Package ssr provides modular SSR server management components.
package ssr

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// RestartCallback is called when SSR needs to be restarted.
type RestartCallback func()

// Monitor monitors SSR process health and triggers auto-restart on crashes.
type Monitor struct {
	process        *ProcessManager
	logger         *slog.Logger
	interval       time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	mu             sync.RWMutex
	running        bool
	lastState      ProcessState
	restartCb      RestartCallback
	restartBackoff time.Duration
	lastRestart    time.Time
}

// NewMonitor creates a new SSR process monitor.
func NewMonitor(process *ProcessManager, logger *slog.Logger, interval time.Duration) *Monitor {
	return &Monitor{
		process:        process,
		logger:         logger,
		interval:       interval,
		lastState:      ProcessStateStopped,
		restartBackoff: 10 * time.Second, // Minimum time between restarts
	}
}

// SetRestartCallback sets the function to call when SSR needs restart.
func (m *Monitor) SetRestartCallback(cb RestartCallback) {
	m.mu.Lock()
	m.restartCb = cb
	m.mu.Unlock()
}

// SetCancel sets the cancel function for the monitor context.
func (m *Monitor) SetCancel(cancel context.CancelFunc) {
	m.mu.Lock()
	m.cancel = cancel
	m.mu.Unlock()
}

// Start begins monitoring the SSR process.
func (m *Monitor) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.running = true
	m.mu.Unlock()

	m.wg.Add(1)
	go m.run()

	m.logger.Info("SSR monitor started", "interval", m.interval)
}

// Stop stops the monitor.
func (m *Monitor) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	if m.cancel != nil {
		m.cancel()
	}
	m.mu.Unlock()

	m.wg.Wait()
	m.logger.Info("SSR monitor stopped")
}

// run is the main monitoring loop.
func (m *Monitor) run() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.check()
		}
	}
}

// check performs a single health check and triggers restart if needed.
func (m *Monitor) check() {
	state := m.process.State()

	// Log state transitions.
	if state != m.lastState {
		m.logger.Info("SSR state changed", "from", m.lastState, "to", state)
		m.lastState = state
	}

	health := m.process.HealthCheck()

	// Check if process crashed (process exited but we think it's running).
	shouldRestart := false

	switch state {
	case ProcessStateCrashed:
		m.logger.Warn("SSR process crashed, triggering restart")
		shouldRestart = true
	case ProcessStateRunning, ProcessStateStarting:
		if !health.Healthy {
			// Process running but HTTP health check failing - might be hung.
			m.logger.Warn("SSR process unhealthy (HTTP check failed)", "status", health.String())
			shouldRestart = true
		}
	case ProcessStateStopped, ProcessStateStopping:
		// Process is stopping or stopped - no restart needed.
	default:
		// Unknown state - log but don't restart.
		m.logger.Debug("SSR unknown state", "state", state)
	}

	// Trigger restart if needed and not in backoff period.
	if shouldRestart {
		m.mu.RLock()
		cb := m.restartCb
		lastRestart := m.lastRestart
		backoff := m.restartBackoff
		m.mu.RUnlock()

		if cb != nil {
			if time.Since(lastRestart) > backoff {
				m.mu.Lock()
				m.lastRestart = time.Now()
				m.mu.Unlock()

				m.logger.Info("Calling restart callback for SSR")
				go cb() // Call restart in goroutine to avoid deadlock
			} else {
				m.logger.Warn("SSR restart skipped (backoff period)", "backoff", backoff)
			}
		}
	}
}
