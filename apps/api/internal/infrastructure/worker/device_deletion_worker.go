package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/serverlock"
)

type DeviceDeletionWorker struct {
	deviceRepo device.Repository
	lockSvc    *serverlock.Service
	logger     *slog.Logger
	stopCh     chan struct{}
	doneCh     chan struct{}
	interval   time.Duration
}

func NewDeviceDeletionWorker(deviceRepo device.Repository, lockSvc *serverlock.Service, logger *slog.Logger, interval time.Duration) *DeviceDeletionWorker {
	return &DeviceDeletionWorker{
		deviceRepo: deviceRepo,
		lockSvc:    lockSvc,
		logger:     logger,
		interval:   interval,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

func (w *DeviceDeletionWorker) Start() {
	go w.run()
	w.logger.Info("device deletion worker started", "interval", w.interval.String())
}

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

	if w.lockSvc != nil {
		acquired, err := w.lockSvc.Acquire(ctx, "device-deletion-worker", "primary", w.interval)
		if err != nil || !acquired {
			return
		}
		defer func() { _ = w.lockSvc.Release(ctx, "device-deletion-worker", "primary") }()
	}

	deleted, err := w.deviceRepo.DeleteScheduled(ctx)
	if err != nil {
		w.logger.Error("failed to process scheduled deletions", "error", err)
		return
	}
	if deleted > 0 {
		w.logger.Info("processed scheduled device deletions", "deleted_count", deleted)
	}
}

// Healthy reports whether the worker has not yet been stopped.
func (w *DeviceDeletionWorker) Healthy() bool {
select {
case <-w.doneCh:
return false
default:
return true
}
}
