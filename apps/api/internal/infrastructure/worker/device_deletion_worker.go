package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
)

// DeviceDeletionWorker runs periodically to delete devices whose.
// deletion_scheduled_at has passed.
type DeviceDeletionWorker struct {
	deviceRepo device.Repository
	logger     *slog.Logger
	stopCh     chan struct{}
	doneCh     chan struct{}
	interval   time.Duration
}

// NewDeviceDeletionWorker creates a new device deletion worker.
func NewDeviceDeletionWorker(deviceRepo device.Repository, logger *slog.Logger, interval time.Duration) *DeviceDeletionWorker {
	return &DeviceDeletionWorker{
		deviceRepo: deviceRepo,
		logger:     logger,
		interval:   interval,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

// Start begins the background worker loop.
func (w *DeviceDeletionWorker) Start() {
	go w.run()
	w.logger.Info("device deletion worker started",
		"interval", w.interval.String(),
	)
}

// Stop gracefully stops the worker.
func (w *DeviceDeletionWorker) Stop() {
	close(w.stopCh)
	<-w.doneCh
	w.logger.Info("device deletion worker stopped")
}

func (w *DeviceDeletionWorker) run() {
	defer close(w.doneCh)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processDeletions()
		}
	}
}

func (w *DeviceDeletionWorker) processDeletions() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deleted, err := w.deviceRepo.DeleteScheduled(ctx)
	if err != nil {
		w.logger.Error("failed to process scheduled deletions",
			"error", err,
		)
		return
	}

	if deleted > 0 {
		w.logger.Info("processed scheduled device deletions",
			"deleted_count", deleted,
		)
	}
}
