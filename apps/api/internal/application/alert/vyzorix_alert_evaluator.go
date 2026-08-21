package alert

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/cache"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/uuid"
)

// DashboardBroadcaster is the websocket hub's broadcast surface; routes alert
// transitions to connected dashboards alongside webhook delivery.
type DashboardBroadcaster interface {
	BroadcastEvent(eventType string, data []byte) error
}

// Annotator marks alert transitions on the fleet timeline so operators can
// correlate rollouts with failure spikes.
type Annotator interface {
	Annotate(ctx context.Context, rule *alert.Rule, transition *alert.Transition) error
}

// Evaluator advances the state machine of every enabled rule against the
// metric source, persists transition history, and emits notifications.
type Evaluator struct {
	rules         alert.Repository
	states        alert.StateRepository
	history       alert.HistoryRepository
	metrics       *MetricSource
	notifier      alert.Notifier
	hub           DashboardBroadcaster
	annotator     Annotator
	dashboardCache *cache.Section
	logger        *slog.Logger
}

func NewEvaluator(rules alert.Repository, states alert.StateRepository, history alert.HistoryRepository, metrics *MetricSource, notifier alert.Notifier, hub DashboardBroadcaster, logger *slog.Logger) *Evaluator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Evaluator{rules: rules, states: states, history: history, metrics: metrics, notifier: notifier, hub: hub, logger: logger}
}

// SetAnnotator wires the timeline annotator for firing/resolved transitions.
func (e *Evaluator) SetAnnotator(a Annotator) {
	e.annotator = a
}

// SetDashboardCache wires the dashboard stats cache so firing/resolved
// transitions invalidate the org's dashboard stats entry.
func (e *Evaluator) SetDashboardCache(c *cache.Section) {
	e.dashboardCache = c
}

// EvaluateAll evaluates every enabled rule and returns the number of state
// transitions that occurred.
func (e *Evaluator) EvaluateAll(ctx context.Context, now time.Time) (int, error) {
	rules, err := e.rules.ListEnabled(ctx)
	if err != nil {
		return 0, err
	}

	transitions := 0
	for _, rule := range rules {
		transitioned, err := e.evaluateRule(ctx, rule, now)
		if err != nil {
			e.logger.Error("alert rule evaluation failed", "rule_id", rule.ID, "org_id", rule.OrgID, "error", err)
			continue
		}
		if transitioned {
			transitions++
		}
	}
	metrics.Get().UpdateAlertRules(len(rules))
	return transitions, nil
}

// EvaluateRule evaluates one rule immediately (manual trigger) and returns
// whether a transition occurred.
func (e *Evaluator) EvaluateRule(ctx context.Context, ruleID string, now time.Time) (bool, error) {
	rule, err := e.rules.GetByID(ctx, ruleID)
	if err != nil {
		return false, err
	}
	return e.evaluateRule(ctx, rule, now)
}

func (e *Evaluator) evaluateRule(ctx context.Context, rule *alert.Rule, now time.Time) (bool, error) {
	value, err := e.metrics.Value(ctx, rule.OrgID, rule.Metric, now)
	if err != nil {
		return false, err
	}

	inst, err := e.states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		return false, err
	}

	transition := inst.Evaluate(rule, value, now)
	if err := e.states.Upsert(ctx, inst); err != nil {
		return false, err
	}
	if !inst.NotificationDue {
		metrics.Get().RecordAlertEvaluation("unchanged")
		return false, nil
	}

	e.logger.Info("alert notification",
		"rule_id", rule.ID, "rule", rule.Name, "org_id", rule.OrgID,
		"from", transition.From, "to", transition.To, "value", transition.Value)

	if e.history != nil {
		evt := &alert.Event{
			ID:        uuid.New(),
			RuleID:    rule.ID,
			FromState: transition.From,
			ToState:   transition.To,
			Value:     transition.Value,
			CreatedAt: transition.At.UnixMilli(),
		}
		if err := e.history.Append(ctx, evt); err != nil {
			e.logger.Error("alert history append failed", "rule_id", rule.ID, "error", err)
		}
	}

	metrics.Get().RecordAlertEvaluation(string(transition.To))
	if e.annotator != nil && (transition.Firing() || transition.Resolved()) {
		if err := e.annotator.Annotate(ctx, rule, transition); err != nil {
			e.logger.Error("alert annotation failed", "rule_id", rule.ID, "error", err)
		}
	}
	if e.dashboardCache != nil && (transition.Firing() || transition.Resolved()) {
		e.dashboardCache.Delete(rule.OrgID)
	}
	notif := &alert.Notification{Rule: rule, Transition: transition}
	if e.notifier != nil && rule.WebhookURL != "" {
		if err := e.notifier.Notify(ctx, notif); err != nil {
			e.logger.Error("alert webhook delivery failed", "rule_id", rule.ID, "error", err)
		}
	}
	if e.hub != nil {
		data, err := json.Marshal(map[string]interface{}{
			"ruleId":    rule.ID,
			"ruleName":  rule.Name,
			"orgId":     rule.OrgID,
			"metric":    rule.Metric,
			"fromState": transition.From,
			"toState":   transition.To,
			"value":     transition.Value,
		})
		if err == nil {
			if err := e.hub.BroadcastEvent("alert_notification", data); err != nil {
				e.logger.Error("alert dashboard broadcast failed", "rule_id", rule.ID, "error", err)
			}
		}
	}
	return true, nil
}
