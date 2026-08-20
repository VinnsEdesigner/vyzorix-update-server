package worker

import (
	"context"
	"log/slog"
	"time"

	alertapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/alert"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/serverlock"
)

// AlertEvaluationWorker ticks on an interval and evaluates all enabled alert
// rules. The server lock ensures only one instance evaluates in a multi-node
// deployment.
type AlertEvaluationWorker struct {
	evaluator *alertapp.Evaluator
	lockSvc   *serverlock.Service
	logger    *slog.Logger
	stopCh    chan struct{}
	doneCh    chan struct{}
	interval  time.Duration
}

// NewAlertEvaluationWorker creates a new AlertEvaluationWorker.
func NewAlertEvaluationWorker(evaluator *alertapp.Evaluator, lockSvc *serverlock.Service, logger *slog.Logger, interval time.Duration) *AlertEvaluationWorker {
	return &AlertEvaluationWorker{
		evaluator: evaluator,
		lockSvc:   lockSvc,
		logger:    logger,
		interval:  interval,
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

// Start launches the evaluation loop.
func (w *AlertEvaluationWorker) Start() {
	go w.run()
	w.logger.Info("alert evaluation worker started", "interval", w.interval.String())
}

// Stop shuts the worker down and waits for the loop to exit.
func (w *AlertEvaluationWorker) Stop() {
	close(w.stopCh)
	<-w.doneCh
	w.logger.Info("alert evaluation worker stopped")
}

func (w *AlertEvaluationWorker) run() {
	defer close(w.doneCh)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.evaluate()
		}
	}
}

func (w *AlertEvaluationWorker) evaluate() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if w.lockSvc != nil {
		acquired, err := w.lockSvc.Acquire(ctx, "alert-evaluation-worker", "primary", w.interval)
		if err != nil || !acquired {
			return
		}
		defer func() { _ = w.lockSvc.Release(ctx, "alert-evaluation-worker", "primary") }()
	}

	transitions, err := w.evaluator.EvaluateAll(ctx, time.Now())
	if err != nil {
		w.logger.Error("alert evaluation failed", "error", err)
		return
	}
	if transitions > 0 {
		w.logger.Info("alert evaluation complete", "transitions", transitions)
	}
}
