package command

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	ws "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
)

// OutboxConfig holds configuration for the command outbox worker.
type OutboxConfig struct {
	PollInterval   time.Duration // How often to poll for pending commands.
	MaxRetries     int           // Maximum retry attempts before marking failed.
	RetryBaseDelay time.Duration // Base delay for exponential backoff.
	BatchSize      int           // Number of commands to process per poll.
}

// DefaultOutboxConfig returns the default outbox configuration.
func DefaultOutboxConfig() OutboxConfig {
	return OutboxConfig{
		PollInterval:   1 * time.Second,
		MaxRetries:     5,
		RetryBaseDelay: 5 * time.Second,
		BatchSize:      50,
	}
}

// Outbox processes pending commands from the outbox table.
// It implements the transactional outbox pattern: commands are written to the DB.
// atomically with their status, then processed asynchronously by this worker.
type Outbox struct {
	repo        command.Repository
	deviceRepo  device.Repository
	fcmNotifier fcm.Notifier
	hub         *ws.Hub
	log         *slog.Logger
	stopCh      chan struct{}
	cfg         OutboxConfig
	wg          sync.WaitGroup
	mu          sync.Mutex
	running     bool
}

// NewOutbox creates a new command outbox worker.
func NewOutbox(
	repo command.Repository,
	deviceRepo device.Repository,
	hub *ws.Hub,
	fcmNotifier fcm.Notifier,
	cfg OutboxConfig,
	log *slog.Logger,
) *Outbox {
	if cfg.PollInterval == 0 {
		cfg = DefaultOutboxConfig()
	}
	return &Outbox{
		repo:        repo,
		deviceRepo:  deviceRepo,
		hub:         hub,
		fcmNotifier: fcmNotifier,
		cfg:         cfg,
		log:         log,
		stopCh:      make(chan struct{}),
	}
}

// Start begins processing pending commands in the background.
func (o *Outbox) Start() {
	o.mu.Lock()
	if o.running {
		o.mu.Unlock()
		return
	}
	o.running = true
	o.mu.Unlock()

	o.wg.Add(1)
	go o.run()
	o.log.Info("command outbox worker started")
}

// Stop gracefully stops the outbox worker.
func (o *Outbox) Stop() {
	o.mu.Lock()
	if !o.running {
		o.mu.Unlock()
		return
	}
	o.running = false
	o.mu.Unlock()

	close(o.stopCh)
	o.wg.Wait()
	o.log.Info("command outbox worker stopped")
}

// run is the main processing loop.
func (o *Outbox) run() {
	defer o.wg.Done()

	ticker := time.NewTicker(o.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-o.stopCh:
			return
		case <-ticker.C:
			o.processPendingCommands()
		}
	}
}

// processPendingCommands fetches and processes pending commands.
func (o *Outbox) processPendingCommands() {
	ctx := context.Background()

	// Get pending commands from the outbox using FindPending.
	cmds, err := o.repo.FindPending(ctx, o.cfg.BatchSize)
	if err != nil {
		o.log.Error("failed to fetch pending commands", "err", err)
		return
	}

	for _, cmd := range cmds {
		select {
		case <-o.stopCh:
			return
		default:
			o.processCommand(cmd)
		}
	}
}

// processCommand attempts to deliver a single command.
func (o *Outbox) processCommand(cmd *command.Command) bool {
	ctx := context.Background()

	// Try WebSocket delivery first if device is online.
	if o.hub != nil && o.hub.Online(cmd.DeviceID) {
		if o.deliverViaWebSocket(ctx, cmd) {
			return true
		}
	}

	// Try FCM delivery.
	if o.fcmNotifier != nil && o.deviceRepo != nil {
		if o.deliverViaFCM(ctx, cmd) {
			return true
		}
	}

	// Delivery failed - handle retry or mark as failed.
	o.handleFailedDelivery(cmd)
	return false
}

// deliverViaWebSocket attempts to deliver command via WebSocket with confirmation.
func (o *Outbox) deliverViaWebSocket(_ctx context.Context, cmd *command.Command) bool {
	if o.hub == nil {
		return false
	}

	// Build command frame.
	var args json.RawMessage
	if cmd.Args != nil {
		args = cmd.Args
	} else {
		args = []byte("{}")
	}

	frame := command.CommandFrame{
		Type:       cmd.Command,
		Command:    cmd.Command,
		DispatchID: cmd.DispatchID,
		Args:       args,
		Timestamp:  time.Now().UnixMilli(),
	}

	// Use delivery confirmation with 5 second timeout.
	confirmed, err := o.hub.SendWithDeliveryConfirmation(cmd.DeviceID, frame, 5*time.Second)
	if err != nil {
		o.log.Warn("websocket delivery error",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID,
			"err", err)
		return false
	}

	if confirmed {
		if err := o.markDelivered(cmd); err != nil {
			o.log.Error("failed to mark command delivered",
				"dispatchID", cmd.DispatchID,
				"deviceID", cmd.DeviceID,
				"err", err)
			// Don't return true - delivery confirmation failed to persist.
			return false
		}
		o.log.Info("command delivered via websocket",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID)
		return true
	}

	return false
}

// deliverViaFCM attempts to deliver command via FCM push notification.
func (o *Outbox) deliverViaFCM(ctx context.Context, cmd *command.Command) bool {
	if o.fcmNotifier == nil || o.deviceRepo == nil {
		return false
	}

	// Get device to retrieve FCM token.
	dev, err := o.deviceRepo.FindByID(ctx, cmd.DeviceID)
	if err != nil {
		o.log.Warn("device not found for FCM delivery",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID)
		return false
	}

	fcmToken := dev.FCMToken
	if fcmToken == "" {
		o.log.Debug("device has no FCM token",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID)
		return false
	}

	// Build and send FCM wake notification.
	wake := fcm.SilentWake{
		Token:      fcmToken,
		Command:    cmd.Command,
		DispatchID: cmd.DispatchID,
		DeviceID:   cmd.DeviceID,
	}

	if err := o.fcmNotifier.SendSilentWake(ctx, wake); err != nil {
		o.log.Warn("FCM delivery failed",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID,
			"err", err)
		return false
	}

	o.log.Info("command queued via FCM",
		"dispatchID", cmd.DispatchID,
		"deviceID", cmd.DeviceID)

	// FCM is fire-and-forget, so we mark as queued - device will poll for commands.
	// The command remains in pending status until device confirms receipt.
	return true
}

// markDelivered marks a command as successfully delivered.
// Returns an error if the database update fails.
func (o *Outbox) markDelivered(cmd *command.Command) error {
	ctx := context.Background()
	if err := o.repo.MarkDelivered(ctx, cmd.DispatchID); err != nil {
		return err
	}
	return nil
}

// handleFailedDelivery handles a command that couldn't be delivered.
// Implements retry with exponential backoff up to MaxRetries.
func (o *Outbox) handleFailedDelivery(cmd *command.Command) {
	ctx := context.Background()
	if cmd.IsExpired() {
		errMsg := "command expired before delivery"
		if err := o.repo.MarkFailed(ctx, cmd.DispatchID, errMsg); err != nil {
			o.log.Error("failed to mark expired command as failed",
				"dispatchID", cmd.DispatchID,
				"err", err)
		}
		o.log.Warn("command marked as failed - expired",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID)
		return
	}

	// Increment retry count and set max retries from config.
	cmd.RetryCount++
	cmd.MaxRetries = o.cfg.MaxRetries

	if cmd.RetryCount >= cmd.MaxRetries {
		// Mark as failed.
		errMsg := "max delivery retries exceeded"
		if err := o.repo.MarkFailed(ctx, cmd.DispatchID, errMsg); err != nil {
			o.log.Error("failed to mark command as failed",
				"dispatchID", cmd.DispatchID,
				"err", err)
		}
		o.log.Warn("command marked as failed after max retries",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID,
			"retries", cmd.RetryCount)
		return
	}

	// Calculate next retry time with exponential backoff.
	// Exponential backoff: baseDelay * 2^(retryCount-1).
	delay := o.cfg.RetryBaseDelay * time.Duration(1<<(cmd.RetryCount-1))
	nextRetry := time.Now().Add(delay)
	cmd.NextRetryAt = &nextRetry

	// Persist retry info to database for outbox pattern durability.
	if err := o.repo.UpdateRetryInfo(ctx, cmd.ID, cmd.RetryCount, cmd.MaxRetries, cmd.NextRetryAt); err != nil {
		o.log.Error("failed to persist retry info",
			"dispatchID", cmd.DispatchID,
			"err", err)
	}

	o.log.Warn("command delivery failed, will retry",
		"dispatchID", cmd.DispatchID,
		"deviceID", cmd.DeviceID,
		"retryCount", cmd.RetryCount,
		"nextRetryIn", delay)
}

// IsRunning returns whether the outbox worker is currently running.
func (o *Outbox) IsRunning() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.running
}
