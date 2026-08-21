package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

// FCMRetryWorker runs periodically to retry failed FCM notifications.
type FCMRetryWorker struct {
	fcmNotifier fcm.Notifier
	db          *sql.DB
	logger      *slog.Logger
	stopCh      chan struct{}
	doneCh      chan struct{}
	interval    time.Duration
	maxRetries  int
	baseDelay   time.Duration
}

// NewFCMRetryWorker creates a new FCM retry worker.
func NewFCMRetryWorker(db *sql.DB, fcmNotifier fcm.Notifier, logger *slog.Logger, interval time.Duration) *FCMRetryWorker {
	return &FCMRetryWorker{
		db:          db,
		fcmNotifier: fcmNotifier,
		logger:      logger,
		interval:    interval,
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
		maxRetries:  5,
		baseDelay:   time.Minute,
	}
}

// Start begins the background worker loop.
func (w *FCMRetryWorker) Start() {
	go w.run()
	w.logger.Info("fcm retry worker started",
		"interval", w.interval.String(),
		"maxRetries", w.maxRetries,
	)
}

// Stop gracefully stops the worker.
func (w *FCMRetryWorker) Stop() {
	close(w.stopCh)
	<-w.doneCh
	w.logger.Info("fcm retry worker stopped")
}

func (w *FCMRetryWorker) run() {
	defer close(w.doneCh)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processRetries()
		}
	}
}

func (w *FCMRetryWorker) processRetries() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repo := storage.NewPendingFCMRepository(w.db)
	now := time.Now().UnixMilli()

	// Get pending notifications due for retry.
	notifications, err := w.getPendingNotifications(ctx, repo, now, 100)
	if err != nil {
		w.logger.Error("failed to get pending FCM notifications",
			"error", err,
		)
		return
	}

	for _, n := range notifications {
		w.processNotification(ctx, repo, &n)
	}
}

func (w *FCMRetryWorker) getPendingNotifications(ctx context.Context, _repo *storage.PendingFCMRepository, now int64, limit int) ([]storage.PendingFCMNotification, error) {
	query := `
		SELECT id, dispatch_id, device_id, token, command, priority, retry_count, next_retry_at, last_error, created_at, updated_at
		FROM pending_fcm
		WHERE next_retry_at <= ?
		ORDER BY next_retry_at ASC
		LIMIT ?
	`
	rows, err := w.db.QueryContext(ctx, query, now, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var notifications []storage.PendingFCMNotification
	for rows.Next() {
		var n storage.PendingFCMNotification
		err := rows.Scan(
			&n.ID, &n.DispatchID, &n.DeviceID, &n.Token, &n.Command, &n.Priority,
			&n.RetryCount, &n.NextRetryAt, &n.LastError, &n.CreatedAt, &n.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (w *FCMRetryWorker) processNotification(ctx context.Context, repo *storage.PendingFCMRepository, n *storage.PendingFCMNotification) {
	// Build the silent wake notification.
	wake := fcm.SilentWake{
		Token:      n.Token,
		Command:    n.Command,
		DispatchID: n.DispatchID,
		DeviceID:   n.DeviceID,
		Priority:   n.Priority,
	}

	// Attempt to send.
	err := w.fcmNotifier.SendSilentWake(ctx, wake)
	if err == nil {
		// Success - delete the pending notification.
		if deleteErr := repo.Delete(ctx, n.ID); deleteErr != nil {
			w.logger.Warn("failed to delete pending FCM notification after success",
				"id", n.ID,
				"error", deleteErr,
			)
		}
		w.logger.Info("fcm retry successful",
			"dispatchId", n.DispatchID,
			"deviceId", n.DeviceID,
			"retryCount", n.RetryCount,
		)
		return
	}

	// Failure - update retry count and schedule next attempt.
	n.RetryCount++
	n.LastError = err.Error()
	n.UpdatedAt = time.Now().UnixMilli()

	if n.RetryCount >= w.maxRetries {
		// Max retries exceeded - delete and log.
		if deleteErr := repo.Delete(ctx, n.ID); deleteErr != nil {
			w.logger.Warn("failed to delete maxed-out FCM notification",
				"id", n.ID,
				"error", deleteErr,
			)
		}
		w.logger.Error("fcm notification failed after max retries",
			"dispatchId", n.DispatchID,
			"deviceId", n.DeviceID,
			"retryCount", n.RetryCount,
			"lastError", err,
		)
		return
	}

	// Schedule exponential backoff: 1min, 2min, 4min, 8min, 16min.
	delay := w.baseDelay * time.Duration(1<<(n.RetryCount-1))
	n.NextRetryAt = time.Now().Add(delay).UnixMilli()

	if updateErr := repo.Update(ctx, n); updateErr != nil {
		w.logger.Warn("failed to update pending FCM notification",
			"id", n.ID,
			"error", updateErr,
		)
	}

	w.logger.Warn("fcm retry failed, scheduled for next attempt",
		"dispatchId", n.DispatchID,
		"deviceId", n.DeviceID,
		"retryCount", n.RetryCount,
		"nextRetryAt", time.UnixMilli(n.NextRetryAt).Format(time.RFC3339),
		"lastError", err,
	)
}

// PersistPendingNotification saves a failed FCM notification for later retry.
func PersistPendingNotification(ctx context.Context, db *sql.DB, wake fcm.SilentWake) error {
	repo := storage.NewPendingFCMRepository(db)
	notification := &storage.PendingFCMNotification{
		DispatchID:  wake.DispatchID,
		DeviceID:    wake.DeviceID,
		Token:       wake.Token,
		Command:     wake.Command,
		Priority:    wake.Priority,
		RetryCount:  0,
		NextRetryAt: time.Now().Add(time.Minute).UnixMilli(), // First retry in 1 minute.
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
	}
	return repo.Create(ctx, notification)
}

// Healthy reports whether the worker has not yet been stopped.
func (w *FCMRetryWorker) Healthy() bool {
select {
case <-w.doneCh:
return false
default:
return true
}
}
