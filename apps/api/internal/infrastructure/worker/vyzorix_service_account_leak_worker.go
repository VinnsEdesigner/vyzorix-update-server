package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/serviceaccount"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/serverlock"
)

// ServiceAccountLeakWorker periodically scans public GitHub and GitLab code
// search for leaked service account tokens and emails org admins.
type ServiceAccountLeakWorker struct {
	service *serviceaccount.Service
	lockSvc *serverlock.Service
	logger  *slog.Logger
	stopCh  chan struct{}
	doneCh  chan struct{}
	interval time.Duration
}

// NewServiceAccountLeakWorker creates a new worker.
func NewServiceAccountLeakWorker(service *serviceaccount.Service, lockSvc *serverlock.Service, logger *slog.Logger, interval time.Duration) *ServiceAccountLeakWorker {
	return &ServiceAccountLeakWorker{
		service:  service,
		lockSvc:  lockSvc,
		logger:   logger,
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start launches the scan loop.
func (w *ServiceAccountLeakWorker) Start() {
	go w.run()
	w.logger.Info("service account leak worker started", "interval", w.interval.String())
}

// Stop shuts the worker down.
func (w *ServiceAccountLeakWorker) Stop() {
	close(w.stopCh)
	<-w.doneCh
	w.logger.Info("service account leak worker stopped")
}

func (w *ServiceAccountLeakWorker) run() {
	defer close(w.doneCh)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.scan()
		}
	}
}

func (w *ServiceAccountLeakWorker) scan() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if w.lockSvc != nil {
		acquired, err := w.lockSvc.Acquire(ctx, "service-account-leak-worker", "primary", w.interval)
		if err != nil || !acquired {
			return
		}
		defer func() { _ = w.lockSvc.Release(ctx, "service-account-leak-worker", "primary") }()
	}

	// Scan across all orgs. Uses the service's ScanForLeaks; org listing comes
	// through the service's repository.
	orgs, err := w.service.ListAllOrgs(ctx)
	if err != nil {
		w.logger.Error("failed to list orgs for leak scan", "error", err)
		return
	}
	for _, orgID := range orgs {
		if _, err := w.service.ScanForLeaks(ctx, orgID); err != nil {
			w.logger.Error("leak scan failed", "org_id", orgID, "error", err)
		}
	}
}
