package command

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/outcome"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
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
	repo          command.Repository
	deviceRepo    device.Repository
	fcmNotifier   fcm.Notifier
	hub           *ws.Hub
	commandSigner *cryptohmac.CommandSigner
	log           *slog.Logger
	stopCh        chan struct{}
	guard         *command.DispatchGuard
	cfg           OutboxConfig
	wg            sync.WaitGroup
	mu            sync.Mutex
	running       bool
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
		repo:          repo,
		deviceRepo:    deviceRepo,
		hub:           hub,
		fcmNotifier:   fcmNotifier,
		commandSigner: cryptohmac.NewCommandSigner(),
		cfg:           cfg,
		log:           log,
		guard:         command.NewDispatchGuard(1),
		stopCh:        make(chan struct{}),
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
	metrics.Get().UpdateCommandQueue(len(cmds))

	for _, cmd := range cmds {
		select {
		case <-o.stopCh:
			return
		default:
			o.processCommand(cmd)
		}
	}
}

// processCommand attempts to deliver a single command and emits delivery metrics.
func (o *Outbox) processCommand(cmd *command.Command) bool {
	ctx := context.Background()

	// Take the device's dispatch slot before sending anything. If it is already
	// held, an earlier attempt to this device is still in flight and may yet
	// execute; starting a second send would let two irreversible commands race.
	if !o.guard.Acquire(cmd.DeviceID) {
		o.log.Warn("dispatch already in flight for device; deferring",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID)
		return false
	}
	defer o.guard.Release(cmd.DeviceID)

	owned := command.DispatchResource(cmd.DeviceID)

	// Try WebSocket delivery first if device is online.
	if o.hub != nil && o.hub.Online(cmd.DeviceID) {
		sent, err := o.deliverViaWebSocket(ctx, cmd)
		if sent && err == nil {
			return true
		}
		if failure := command.ClassifyDelivery(command.TransportResult{Sent: sent, Acked: false, Err: err}); failure != nil {
			o.handleFailedDeliveryWith(cmd, failure.WithOwned(owned))
			return false
		}
	}

	// Try FCM delivery.
	if o.fcmNotifier != nil && o.deviceRepo != nil {
		sent, err := o.deliverViaFCM(ctx, cmd)
		if sent && err == nil {
			return true
		}
		if failure := command.ClassifyDelivery(command.TransportResult{Sent: sent, Acked: false, Err: err}); failure != nil {
			o.handleFailedDeliveryWith(cmd, failure.WithOwned(owned))
			return false
		}
	}

	o.handleFailedDelivery(cmd)
	return false
}

// deliverViaWebSocket attempts to deliver command via WebSocket with
// confirmation. It reports whether the frame was written to the socket, and any
// error encountered. A written-but-unacknowledged frame is the uncertain case
// the caller must classify before retrying.
func (o *Outbox) deliverViaWebSocket(ctx context.Context, cmd *command.Command) (bool, error) {
	if o.hub == nil {
		return false, nil
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

	// Sign the frame with the device's command secret (Domain B: server→device).
	dev, err := o.deviceRepo.FindByID(ctx, cmd.DeviceID)
	if err != nil {
		o.log.Warn("failed to load device for command signing",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID,
			"err", err)
		return false, err
	}
	if dev.CommandSecretHash == "" {
		o.log.Warn("device has no command secret; cannot sign outbox command",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID)
		return false, nil
	}
	if signErr := cryptohmac.SignCommandFrame(o.commandSigner, &frame, cmd.DeviceID, dev.CommandSecretHash); signErr != nil {
		o.log.Warn("failed to sign outbox command frame",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID,
			"err", signErr)
		return false, signErr
	}

	// Use delivery confirmation with 5 second timeout.
	confirmed, err := o.hub.SendWithDeliveryConfirmation(cmd.DeviceID, frame, 5*time.Second)
	if err != nil {
		o.log.Warn("websocket delivery error",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID,
			"err", err)
		// The frame was written to the socket before the acknowledgement
		// window elapsed, so the device may already have executed it.
		return true, err
	}

	if confirmed {
		if err := o.markDelivered(cmd); err != nil {
			o.log.Error("failed to mark command delivered",
				"dispatchID", cmd.DispatchID,
				"deviceID", cmd.DeviceID,
				"err", err)
			// Don't return true - delivery confirmation failed to persist.
			return true, err
		}
		o.log.Info("command delivered via websocket",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID)
		return true, nil
	}

	return true, command.ErrAckTimeout
}

// deliverViaFCM attempts to deliver command via FCM push notification. FCM is
// fire-and-forget: a successful send means the device was woken, not that it
// received the command, so the caller treats this as an uncertain outcome.
func (o *Outbox) deliverViaFCM(ctx context.Context, cmd *command.Command) (bool, error) {
	if o.fcmNotifier == nil || o.deviceRepo == nil {
		return false, nil
	}

	// Get device to retrieve FCM token.
	dev, err := o.deviceRepo.FindByID(ctx, cmd.DeviceID)
	if err != nil {
		o.log.Warn("device not found for FCM delivery",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID)
		return false, err
	}

	fcmToken := dev.FCMToken
	if fcmToken == "" {
		o.log.Debug("device has no FCM token",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID)
		return false, nil
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
		return false, err
	}

	o.log.Info("command queued via FCM",
		"dispatchID", cmd.DispatchID,
		"deviceID", cmd.DeviceID)

	// FCM is fire-and-forget, so we mark as queued - device will poll for commands.
	// The command remains in pending status until device confirms receipt.
	return true, nil
}

// markDelivered marks a command as successfully delivered.
// Returns an error if the database update fails.
func (o *Outbox) markDelivered(cmd *command.Command) error {
	ctx := context.Background()
	if err := o.repo.MarkDelivered(ctx, cmd.DispatchID); err != nil {
		return err
	}
	metrics.Get().RecordCommandOutcome("delivered")
	metrics.Get().RecordCommandDelivery(time.Since(cmd.CreatedAt))
	if o.hub != nil {
		o.hub.PublishCommandStatusForDevice(ctx, cmd.DispatchID, cmd.DeviceID, "delivered")
	}
	return nil
}

// handleFailedDelivery handles a command that couldn't be delivered.
// Implements retry with exponential backoff up to MaxRetries.
func (o *Outbox) handleFailedDelivery(cmd *command.Command) {
	o.handleFailedDeliveryWith(cmd, nil)
}

// handleFailedDeliveryWith applies the typed retry policy for a failed attempt.
// When failure is nil the attempt is treated as a clean, effect-free failure.
func (o *Outbox) handleFailedDeliveryWith(cmd *command.Command, failure *outcome.Failure) {
	ctx := context.Background()
	if cmd.IsExpired() {
		errMsg := "command expired before delivery"
		if err := o.repo.MarkFailed(ctx, cmd.DispatchID, errMsg); err != nil {
			o.log.Error("failed to mark expired command as failed",
				"dispatchID", cmd.DispatchID,
				"err", err)
		}
		metrics.Get().RecordCommandOutcome("failed")
		o.log.Warn("command marked as failed - expired",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID)
		return
	}

	cmd.RetryCount++
	cmd.MaxRetries = o.cfg.MaxRetries

	tier := command.LookupRiskProfile(cmd.Command).Tier
	plan := command.PlanDelivery(failure, tier, cmd.RetryCount, cmd.MaxRetries)

	if plan.Reconcile {
		errMsg := "delivery outcome uncertain for a destructive command; manual reconciliation required"
		if err := o.repo.MarkFailed(ctx, cmd.DispatchID, errMsg); err != nil {
			o.log.Error("failed to mark uncertain command for reconciliation",
				"dispatchID", cmd.DispatchID,
				"err", err)
		}
		metrics.Get().RecordCommandOutcome("reconcile")
		o.log.Warn("command needs reconciliation - not retrying",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID,
			"command", cmd.Command,
			"tier", string(tier),
			"reason", plan.Reason)
		return
	}

	if !plan.Retry {
		errMsg := "max delivery retries exceeded"
		if err := o.repo.MarkFailed(ctx, cmd.DispatchID, errMsg); err != nil {
			o.log.Error("failed to mark command as failed",
				"dispatchID", cmd.DispatchID,
				"err", err)
		}
		metrics.Get().RecordCommandOutcome("failed")
		o.log.Warn("command marked as failed after max retries",
			"dispatchID", cmd.DispatchID,
			"deviceID", cmd.DeviceID,
			"retries", cmd.RetryCount)
		return
	}

	delay := o.cfg.RetryBaseDelay * time.Duration(1<<(cmd.RetryCount-1))
	nextRetry := time.Now().Add(delay)
	cmd.NextRetryAt = &nextRetry

	if err := o.repo.UpdateRetryInfo(ctx, cmd.ID, cmd.RetryCount, cmd.MaxRetries, cmd.NextRetryAt); err != nil {
		o.log.Error("failed to persist retry info",
			"dispatchID", cmd.DispatchID,
			"err", err)
	}

	o.log.Warn("command delivery failed, will retry",
		"dispatchID", cmd.DispatchID,
		"deviceID", cmd.DeviceID,
		"retryCount", cmd.RetryCount,
		"nextRetryIn", delay,
		"reason", plan.Reason)
}

// IsRunning returns whether the outbox worker is currently running.
func (o *Outbox) IsRunning() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.running
}
