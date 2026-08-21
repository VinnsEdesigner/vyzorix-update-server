package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/usagestats"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/serverlock"
)

// UsageStatsWorker periodically collects self-telemetry (entity counts +
// feature toggles) so the update-checker response is as fresh as one
// interval. The server lock ensures only one node collects.
type UsageStatsWorker struct {
	service  *usagestats.Service
	lockSvc  *serverlock.Service
	logger   *slog.Logger
	stopCh   chan struct{}
	doneCh   chan struct{}
	interval time.Duration
}

// NewUsageStatsWorker creates a new UsageStatsWorker.
func NewUsageStatsWorker(service *usagestats.Service, lockSvc *serverlock.Service, logger *slog.Logger, interval time.Duration) *UsageStatsWorker {
	return &UsageStatsWorker{
		service:  service,
		lockSvc:  lockSvc,
		logger:   logger,
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start launches the collection loop.
func (w *UsageStatsWorker) Start() {
	go w.run()
	w.logger.Info("usage stats worker started", "interval", w.interval.String())
}

// Stop shuts the worker down and waits for the loop to exit.
func (w *UsageStatsWorker) Stop() {
	close(w.stopCh)
	<-w.doneCh
	w.logger.Info("usage stats worker stopped")
}

func (w *UsageStatsWorker) run() {
	defer close(w.doneCh)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.collect()
		}
	}
}

func (w *UsageStatsWorker) collect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if w.lockSvc != nil {
		acquired, err := w.lockSvc.Acquire(ctx, "usage-stats-worker", "primary", w.interval)
		if err != nil || !acquired {
			return
		}
		defer func() { _ = w.lockSvc.Release(ctx, "usage-stats-worker", "primary") }()
	}
	w.service.Collect(ctx)
}

// Healthy reports whether the worker has not yet been stopped.
func (w *UsageStatsWorker) Healthy() bool {
	return true
}
