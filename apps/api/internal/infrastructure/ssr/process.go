// Package ssr provides modular SSR server management components.
package ssr

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// ProcessState represents the current state of an SSR process.
type ProcessState int

// Process state constants.
const (
	// ProcessStateStopped indicates the process is not running.
	ProcessStateStopped ProcessState = iota
	// ProcessStateStarting indicates the process is starting up.
	ProcessStateStarting
	// ProcessStateRunning indicates the process is running and healthy.
	ProcessStateRunning
	// ProcessStateStopping indicates the process is shutting down.
	ProcessStateStopping
	// ProcessStateCrashed indicates the process has crashed.
	ProcessStateCrashed
)

// String returns a string representation of the process state.
func (s ProcessState) String() string {
	switch s {
	case ProcessStateStopped:
		return "stopped"
	case ProcessStateStarting:
		return "starting"
	case ProcessStateRunning:
		return "running"
	case ProcessStateStopping:
		return "stopping"
	case ProcessStateCrashed:
		return "crashed"
	default:
		return "unknown"
	}
}

// ProcessManager handles SSR subprocess lifecycle with resilience.
type ProcessManager struct {
	startTime  time.Time
	logger     *slog.Logger
	cmd        *exec.Cmd
	cancelFunc context.CancelFunc
	config     Config
	state      ProcessState
	mu         sync.RWMutex
	ready      bool
}

// NewProcessManager creates a new SSR process manager.
func NewProcessManager(config Config, logger *slog.Logger) *ProcessManager {
	return &ProcessManager{
		config: config,
		logger: logger,
		state:  ProcessStateStopped,
	}
}

// State returns the current process state.
func (pm *ProcessManager) State() ProcessState {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.state
}

// IsReady returns whether the process is ready to serve requests.
func (pm *ProcessManager) IsReady() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.ready
}

// StartTime returns when the process started.
func (pm *ProcessManager) StartTime() time.Time {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.startTime
}

// Uptime returns how long the process has been running.
func (pm *ProcessManager) Uptime() time.Duration {
	if pm.startTime.IsZero() {
		return 0
	}

	return time.Since(pm.startTime)
}

// Start launches the SSR subprocess with retry logic.
func (pm *ProcessManager) Start(scriptPath string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.state == ProcessStateRunning || pm.state == ProcessStateStarting {
		pm.logger.Info("SSR process already running or starting")
		return nil
	}

	return pm.startInternalLocked(scriptPath)
}

// Restart stops any existing process and starts fresh.
func (pm *ProcessManager) Restart(scriptPath string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.logger.Info("Restarting SSR process")
	pm.killProcess()
	pm.state = ProcessStateStopped

	return pm.startInternalLocked(scriptPath)
}

// startInternalLocked is the internal start logic (must be called with lock held).
func (pm *ProcessManager) startInternalLocked(scriptPath string) error {
	pm.logger.Info("Starting SSR process", "script", scriptPath, "retries", pm.config.SSRRetryAttempts)

	var lastErr error

	for attempt := 1; attempt <= pm.config.SSRRetryAttempts; attempt++ {
		pm.state = ProcessStateStarting

		if attempt > 1 {
			backoff := time.Duration(attempt*pm.config.SSRRetryBackoff) * time.Second
			pm.logger.Info("SSR retry backoff", "attempt", attempt, "backoff", backoff)
			time.Sleep(backoff)
		}

		pm.logger.Info("SSR start attempt", "attempt", attempt, "of", pm.config.SSRRetryAttempts)

		if err := pm.startProcess(scriptPath); err != nil {
			lastErr = err
			pm.logger.Error("SSR start attempt failed", "attempt", attempt, "err", err)

			continue
		}

		if pm.waitForReady() {
			pm.state = ProcessStateRunning
			pm.startTime = time.Now()
			pm.logger.Info("SSR process started successfully", "pid", pm.cmd.Process.Pid)

			return nil
		}

		// Process didn't become ready, clean up.
		pm.killProcess()

		lastErr = errors.New("SSR process did not become ready")
	}

	pm.state = ProcessStateStopped

	return fmt.Errorf("SSR failed after %d attempts: %w", pm.config.SSRRetryAttempts, lastErr)
}

// startProcess starts the SSR subprocess.
func (pm *ProcessManager) startProcess(scriptPath string) error {
	//
	pm.cmd = exec.Command("node", scriptPath)
	pm.cmd.Stdout = os.Stdout
	pm.cmd.Stderr = os.Stderr
	pm.cmd.Env = append(os.Environ(),
		"NODE_ENV=production",
		"SSR_MODE=production",
	)

	if err := pm.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	// Create cancellation context for monitoring.
	pm.cancelFunc = func() {}

	return nil
}

// waitForReady waits for the SSR server to respond to health checks.
func (pm *ProcessManager) waitForReady() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	timeout := time.Duration(pm.config.SSRBuildTimeout) * time.Second
	deadline := time.Now().Add(timeout)

	pm.logger.Info("Waiting for SSR to be ready", "timeout", timeout)

	healthEndpoints := []string{
		"/health",
		"/healthz",
		"/",
	}

	for time.Now().Before(deadline) {
		if pm.cmd.ProcessState != nil && pm.cmd.ProcessState.Exited() {
			pm.logger.Error("SSR process exited unexpectedly", "code", pm.cmd.ProcessState.ExitCode())
			return false
		}

		for _, endpoint := range healthEndpoints {
			url := pm.config.SSRServerURL + endpoint

			resp, err := client.Get(url)
			if err == nil {
				_ = resp.Body.Close()

				if resp.StatusCode < 500 {
					pm.logger.Info("SSR health check passed", "endpoint", endpoint, "status", resp.StatusCode)
					pm.ready = true

					return true
				}
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return false
}

// killProcess forcefully terminates the SSR subprocess.
func (pm *ProcessManager) killProcess() {
	if pm.cmd == nil || pm.cmd.Process == nil {
		return
	}

	if pm.cancelFunc != nil {
		pm.cancelFunc()
	}

	_ = pm.cmd.Process.Kill()
	_ = pm.cmd.Wait()
	pm.cmd = nil
	pm.ready = false
}

// Stop gracefully shuts down the SSR subprocess.
func (pm *ProcessManager) Stop() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.state != ProcessStateRunning && pm.state != ProcessStateStarting {
		return nil
	}

	pm.state = ProcessStateStopping
	pm.logger.Info("Stopping SSR process")

	if pm.cancelFunc != nil {
		pm.cancelFunc()
	}

	if pm.cmd == nil || pm.cmd.Process == nil {
		pm.state = ProcessStateStopped
		return nil
	}

	// Try graceful shutdown first.
	if err := pm.cmd.Process.Signal(syscall.SIGTERM); err == nil {
		done := make(chan struct{})
		go func() {
			_ = pm.cmd.Wait()

			close(done)
		}()

		select {
		case <-done:
			pm.logger.Info("SSR process stopped gracefully")
		case <-time.After(5 * time.Second):
			pm.logger.Warn("SSR graceful shutdown timed out, forcing")
			_ = pm.cmd.Process.Kill()
			_ = pm.cmd.Wait()
		}
	} else {
		_ = pm.cmd.Process.Kill()
		_ = pm.cmd.Wait()
	}

	pm.cmd = nil
	pm.ready = false
	pm.startTime = time.Time{}
	pm.state = ProcessStateStopped

	return nil
}

// HealthCheck performs a health check on the SSR process.
func (pm *ProcessManager) HealthCheck() *HealthStatus {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	status := &HealthStatus{
		State:  pm.state.String(),
		Ready:  pm.ready,
		Uptime: pm.Uptime(),
		PID:    0,
	}

	if pm.cmd != nil && pm.cmd.Process != nil {
		status.PID = pm.cmd.Process.Pid
	}

	// Check if process has exited unexpectedly (crashed).
	if pm.cmd != nil && pm.cmd.ProcessState != nil && pm.cmd.ProcessState.Exited() {
		status.Healthy = false
		status.State = ProcessStateCrashed.String()
		status.Error = fmt.Sprintf("process exited with code %d", pm.cmd.ProcessState.ExitCode())

		return status
	}

	if pm.state != ProcessStateRunning || !pm.ready {
		return status
	}

	// Perform HTTP health check.
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(pm.config.SSRServerURL + "/health")
	if err != nil {
		status.Healthy = false
		status.Error = err.Error()

		return status
	}

	_ = resp.Body.Close()

	status.Healthy = resp.StatusCode < 500
	status.HTTPStatus = resp.StatusCode

	return status
}

// HealthStatus represents the health of an SSR process.
type HealthStatus struct {
	State      string        `json:"state"`
	Error      string        `json:"error,omitempty"`
	Uptime     time.Duration `json:"uptime"`
	PID        int           `json:"pid"`
	HTTPStatus int           `json:"httpStatus"`
	Ready      bool          `json:"ready"`
	Healthy    bool          `json:"healthy"`
}

// String returns a string representation of the health status.
func (h *HealthStatus) String() string {
	var parts []string
	parts = append(parts, "state="+h.State)
	parts = append(parts, "ready="+strconv.FormatBool(h.Ready))
	parts = append(parts, "healthy="+strconv.FormatBool(h.Healthy))
	parts = append(parts, "uptime="+h.Uptime.String())

	if h.PID > 0 {
		parts = append(parts, "pid="+strconv.Itoa(h.PID))
	}

	if h.Error != "" {
		parts = append(parts, "error="+h.Error)
	}

	return "{" + joinStrings(parts, " ") + "}"
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}

	return result
}
