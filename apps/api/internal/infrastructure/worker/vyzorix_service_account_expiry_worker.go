package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/serviceaccount"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/serverlock"
)

// ServiceAccountExpiryWorker sweeps expired service account tokens and
// revokes them, mirroring the pattern of DeviceDeletionWorker.
type ServiceAccountExpiryWorker struct {
	tokens  serviceaccount.TokenRepository
	lockSvc *serverlock.Service
	logger  *slog.Logger
	stopCh  chan struct{}
	doneCh  chan struct{}
	interval time.Duration
}

// NewServiceAccountExpiryWorker creates a new worker.
func NewServiceAccountExpiryWorker(tokens serviceaccount.TokenRepository, lockSvc *serverlock.Service, logger *slog.Logger, interval time.Duration) *ServiceAccountExpiryWorker {
	return &ServiceAccountExpiryWorker{
		tokens:   tokens,
		lockSvc:  lockSvc,
		logger:   logger,
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start launches the expiry loop.
func (w *ServiceAccountExpiryWorker) Start() {
	go w.run()
	w.logger.Info("service account expiry worker started", "interval", w.interval.String())
}

// Stop shuts the worker down.
func (w *ServiceAccountExpiryWorker) Stop() {
	close(w.stopCh)
	<-w.doneCh
	w.logger.Info("service account expiry worker stopped")
}

func (w *ServiceAccountExpiryWorker) run() {
	defer close(w.doneCh)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.sweep()
		}
	}
}

func (w *ServiceAccountExpiryWorker) sweep() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if w.lockSvc != nil {
		acquired, err := w.lockSvc.Acquire(ctx, "service-account-expiry-worker", "primary", w.interval)
		if err != nil || !acquired {
			return
		}
		defer func() { _ = w.lockSvc.Release(ctx, "service-account-expiry-worker", "primary") }()
	}

	expired, err := w.tokens.ListExpired(ctx, time.Now())
	if err != nil {
		w.logger.Error("failed to list expired service account tokens", "error", err)
		return
	}

	for _, token := range expired {
		if err := w.tokens.Revoke(ctx, token.ID, time.Now()); err != nil {
			w.logger.Error("failed to revoke expired token", "token_id", token.ID, "error", err)
		}
	}
	if len(expired) > 0 {
		w.logger.Info("service account token sweep complete", "revoked", len(expired))
	}
}
