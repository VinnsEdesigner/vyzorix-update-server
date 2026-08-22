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

// StreamPublisher is the org-channel publish surface (live-channel pattern):
// alert transitions are routed to stream/<org>/alerts instead of raw broadcast.
type StreamPublisher interface {
	Publish(channel string, msg map[string]interface{})
}

// Annotator marks alert transitions on the fleet timeline so operators can
// correlate rollouts with failure spikes.
type Annotator interface {
	Annotate(ctx context.Context, rule *alert.Rule, transition *alert.Transition) error
}

// Evaluator advances the state machine of every enabled rule against the
// metric source, persists transition history, and emits notifications.
type Evaluator struct {
	rules          alert.Repository
	states         alert.StateRepository
	history        alert.HistoryRepository
	metrics        *MetricSource
	notifier       alert.Notifier
	hub            DashboardBroadcaster
	publisher      StreamPublisher
	annotator      Annotator
	dashboardCache *cache.Section
	logger         *slog.Logger
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

// SetStreamPublisher wires the org-channel publisher for alert transitions
// (replaces broadcast-to-all with scoped stream publish when set).
func (e *Evaluator) SetStreamPublisher(p StreamPublisher) {
	e.publisher = p
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
		transitions += transitioned
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
	n, err := e.evaluateRule(ctx, rule, now)
	return n > 0, err
}

// evaluateRule evaluates one rule across every labeled series the metric
// emitted this tick and returns the number of transitions emitted.
func (e *Evaluator) evaluateRule(ctx context.Context, rule *alert.Rule, now time.Time) (int, error) {
	instances, err := e.states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		return 0, err
	}

	series, metricErr := e.metrics.Series(ctx, rule.OrgID, rule.Metric, now)
	if metricErr != nil {
		// The error policy applies to every existing instance; entering the
		// error state notifies once, like a state change.
		for _, inst := range instances {
			transition := e.evaluateInstance(inst, rule, now, nil, true)
			if err := e.states.Upsert(ctx, inst); err != nil {
				return 0, err
			}
			if transition != nil {
				e.emit(ctx, rule, transition)
			}
		}
		metrics.Get().RecordAlertEvaluation(string(alert.StateError))
		return 0, metricErr
	}

	keep := make([]string, 0, len(series))
	transitions := 0
	for _, obs := range series {
		hash := alert.LabelsHash(obs.Labels)
		keep = append(keep, hash)
		inst, ok := instances[hash]
		if !ok {
			inst = alert.NewInstance(rule.ID, obs.Labels)
		}
		transition := e.evaluateInstance(inst, rule, now, obs, false)
		if err := e.states.Upsert(ctx, inst); err != nil {
			return transitions, err
		}
		if transition != nil {
			e.emit(ctx, rule, transition)
			transitions++
		}
	}

	// Drop instances whose label set vanished (device class gone) so state
	// cannot accrue silently.
	if err := e.states.DeleteStaleForRule(ctx, rule.ID, keep); err != nil {
		e.logger.Error("alert stale instance cleanup failed", "rule_id", rule.ID, "error", err)
	}
	return transitions, nil
}

// evaluateInstance advances one instance against its observation and returns
// the notification transition, if any. The attempt is stamped immediately so
// a failing channel re-notifies on the rule's interval, not every tick.
func (e *Evaluator) evaluateInstance(inst *alert.Instance, rule *alert.Rule, now time.Time, obs *LabelSeries, metricFailed bool) *alert.Transition {
	var transition *alert.Transition
	switch {
	case metricFailed:
		transition = inst.EvaluateError(rule, now)
	case obs != nil && obs.NoData:
		transition = inst.EvaluateNoData(rule, now)
	case obs != nil:
		transition = inst.Evaluate(rule, obs.Value, now)
	}
	if transition == nil {
		return nil
	}
	inst.LastNotifiedAt = now
	return transition
}

// emit persists history and notifies; firing/resolved also annotate and
// invalidate the dashboard cache.
func (e *Evaluator) emit(ctx context.Context, rule *alert.Rule, transition *alert.Transition) {
	e.logger.Info("alert notification",
		"rule_id", rule.ID, "rule", rule.Name, "org_id", rule.OrgID,
		"labels", transition.Labels, "from", transition.From, "to", transition.To,
		"value", transition.Value)

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
	if e.publisher != nil {
		e.publisher.Publish("stream/"+rule.OrgID+"/alerts", map[string]interface{}{
			"ruleId":    rule.ID,
			"ruleName":  rule.Name,
			"metric":    rule.Metric,
			"labels":    transition.Labels,
			"fromState": transition.From,
			"toState":   transition.To,
			"value":     transition.Value,
		})
	} else if e.hub != nil {
		data, err := json.Marshal(map[string]interface{}{
			"ruleId":    rule.ID,
			"ruleName":  rule.Name,
			"orgId":     rule.OrgID,
			"metric":    rule.Metric,
			"labels":    transition.Labels,
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
}
