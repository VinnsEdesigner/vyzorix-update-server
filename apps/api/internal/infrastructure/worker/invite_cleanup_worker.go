package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// InvitationCleanupWorker runs periodically to mark expired pending invitations.
// Records are transitioned to 'expired' rather than deleted so the audit trail survives.
type InvitationCleanupWorker struct {
	invitationRepo organization.InvitationRepository
	logger         *slog.Logger
	stopCh         chan struct{}
	doneCh         chan struct{}
	interval       time.Duration
}

// NewInvitationCleanupWorker creates a new invitation cleanup worker.
func NewInvitationCleanupWorker(invitationRepo organization.InvitationRepository, logger *slog.Logger, interval time.Duration) *InvitationCleanupWorker {
	return &InvitationCleanupWorker{
		invitationRepo: invitationRepo,
		logger:         logger,
		interval:       interval,
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}
}

// Start begins the background worker loop.
func (w *InvitationCleanupWorker) Start() {
	go w.run()
	w.logger.Info("invitation cleanup worker started", "interval", w.interval.String())
}

// Stop gracefully stops the worker.
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

	expired, err := w.invitationRepo.ExpirePending(ctx)
	if err != nil {
		w.logger.Error("failed to expire pending invitations", "error", err)
		return
	}

	if expired > 0 {
		w.logger.Info("expired pending invitations", "expired_count", expired)
	}
}
