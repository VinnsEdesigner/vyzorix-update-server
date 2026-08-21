package worker

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	alertapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/alert"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/serverlock"
)

// AlertEvaluationWorker ticks on an interval and evaluates all enabled alert
// rules. The server lock ensures only one instance evaluates in a multi-node
// deployment. A per-instance random jitter offsets the first tick so rule
// startup and restarts across nodes never hammer the database in unison.
type AlertEvaluationWorker struct {
	evaluator *alertapp.Evaluator
	lockSvc   *serverlock.Service
	logger    *slog.Logger
	stopCh    chan struct{}
	doneCh    chan struct{}
	interval  time.Duration
	jitter    time.Duration
}

// NewAlertEvaluationWorker creates a new AlertEvaluationWorker. The jitter
// is sampled once so every node offsets its first tick by a random fraction
// of the interval.
func NewAlertEvaluationWorker(evaluator *alertapp.Evaluator, lockSvc *serverlock.Service, logger *slog.Logger, interval time.Duration) *AlertEvaluationWorker {
	jitter := time.Duration(0)
	if interval > 0 {
		jitter = time.Duration(rand.Int63n(int64(interval)))
	}
	return &AlertEvaluationWorker{
		evaluator: evaluator,
		lockSvc:   lockSvc,
		logger:    logger,
		interval:  interval,
		jitter:    jitter,
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

	// First tick fires (interval - jitter) after start so nodes offset from
	// each other; steady state then ticks at the full interval.
	initial := time.NewTimer(w.interval - w.jitter)
	defer initial.Stop()
	select {
	case <-w.stopCh:
		return
	case <-initial.C:
		w.evaluate()
	}

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

	// Backoff-retry: a single transient evaluation failure would keep every
	// rule waiting until the next tick; one bounded retry recovers most.
	err := w.evaluateOnce(ctx)
	if err != nil {
		w.logger.Warn("alert evaluation attempt failed, retrying", "error", err)
		err = w.evaluateOnce(ctx)
		if err != nil {
			w.logger.Error("alert evaluation failed", "error", err)
			return
		}
	}
}

func (w *AlertEvaluationWorker) evaluateOnce(ctx context.Context) error {
	transitions, err := w.evaluator.EvaluateAll(ctx, time.Now())
	if err != nil {
		return err
	}
	if transitions > 0 {
		w.logger.Info("alert evaluation complete", "transitions", transitions)
	}
	return nil
}

// Healthy reports whether the worker has not yet been stopped.
func (w *AlertEvaluationWorker) Healthy() bool {
	select {
	case <-w.doneCh:
		return false
	default:
		return true
	}
}
