package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/serverlock"
)

type InvitationCleanupWorker struct {
	invitationRepo organization.InvitationRepository
	lockSvc        *serverlock.Service
	logger         *slog.Logger
	stopCh         chan struct{}
	doneCh         chan struct{}
	interval       time.Duration
}

func NewInvitationCleanupWorker(invitationRepo organization.InvitationRepository, lockSvc *serverlock.Service, logger *slog.Logger, interval time.Duration) *InvitationCleanupWorker {
	return &InvitationCleanupWorker{
		invitationRepo: invitationRepo,
		lockSvc:        lockSvc,
		logger:         logger,
		interval:       interval,
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}
}

func (w *InvitationCleanupWorker) Start() {
	go w.run()
	w.logger.Info("invitation cleanup worker started", "interval", w.interval.String())
}

func (w *InvitationCleanupWorker) Stop() {
	close(w.stopCh)
	<-w.doneCh
	w.logger.Info("invitation cleanup worker stopped")
}

func (w *InvitationCleanupWorker) run() {
	defer close(w.doneCh)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.expirePending()
		}
	}
}

func (w *InvitationCleanupWorker) expirePending() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if w.lockSvc != nil {
		acquired, err := w.lockSvc.Acquire(ctx, "invite-cleanup-worker", "primary", w.interval)
		if err != nil || !acquired {
			return
		}
		defer func() { _ = w.lockSvc.Release(ctx, "invite-cleanup-worker", "primary") }()
	}

	expired, err := w.invitationRepo.ExpirePending(ctx)
	if err != nil {
		w.logger.Error("failed to expire pending invitations", "error", err)
		return
	}
	if expired > 0 {
		w.logger.Info("expired pending invitations", "expired_count", expired)
	}
}
