package alert

import (
	"context"
	"sort"
	"strings"
	"time"
)

// State is the evaluation state of a rule instance.
type State string

const (
	// StateInactive means the condition is not currently breached.
	StateInactive State = "inactive"
	// StatePending means the condition is breached but has not held for the
	// rule's ForSeconds yet.
	StatePending State = "pending"
	// StateFiring means the condition held long enough; the alert is active.
	StateFiring State = "firing"
	// StateNoData means the metric source produced no signal this tick and
	// the rule's OnNoData policy routed the instance here.
	StateNoData State = "no_data"
	// StateError means the metric source errored on the last evaluation and
	// the rule's OnError policy routed the instance here.
	StateError State = "error"
)

// Valid reports whether the state is a known instance state.
func (s State) Valid() bool {
	switch s {
	case StateInactive, StatePending, StateFiring, StateNoData, StateError:
		return true
	}
	return false
}

// Instance tracks the runtime state of one labeled series of a rule. Labels
// disambiguate series when a metric fans out (e.g. per-device); instances of
// the same rule are keyed by the deterministic LabelsHash on read/write.
type Instance struct {
	Since           time.Time
	LastEvaluated   time.Time
	LastNotifiedAt  time.Time
	RuleID          string
	Labels          map[string]string
	State           State
	LastValue       float64
	NotificationDue bool
}

// LabelsHash flattens labels into a deterministic storage key. Empty labels
// hash to the empty string, collapsing to one instance per rule (fleet-wide
// metrics).
func LabelsHash(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(labels[k])
		b.WriteByte('|')
	}
	return b.String()
}

// LabelsHash returns the deterministic key of the instance's label set.
func (i *Instance) LabelsHash() string { return LabelsHash(i.Labels) }

// NewInstance returns the initial (inactive) instance for a rule+labels pair.
func NewInstance(ruleID string, labels map[string]string) *Instance {
	return &Instance{RuleID: ruleID, Labels: labels, State: StateInactive}
}

// Transition describes a state change produced by one evaluation.
type Transition struct {
	Labels map[string]string
	At     time.Time
	From   State
	To     State
	Value  float64
}

// Firing reports whether the transition entered the firing state.
func (t *Transition) Firing() bool { return t.To == StateFiring }

// Resolved reports whether the transition left pending/firing for inactive.
func (t *Transition) Resolved() bool {
	return t.To == StateInactive && (t.From == StatePending || t.From == StateFiring)
}

// Evaluate advances the state machine for one observation. It mutates the
// instance and returns a Transition when the state changed, nil otherwise.
func (i *Instance) Evaluate(rule *Rule, value float64, now time.Time) *Transition {
	from := i.startFrom(now)
	i.LastValue = value

	breached := rule.Condition.Breached(value, rule.Threshold)
	switch {
	case breached && i.State == StateInactive:
		if rule.ForSeconds <= 0 {
			i.State = StateFiring
		} else {
			i.State = StatePending
		}
		i.Since = now
	case breached && i.State == StatePending:
		if now.Sub(i.Since) >= rule.ForDuration() {
			i.State = StateFiring
			i.Since = now
		}
	case !breached && (i.State == StatePending || i.State == StateFiring):
		i.State = StateInactive
		i.Since = now
	}
	return i.gate(rule, from, value, now)
}

// EvaluateNoData applies the rule's OnNoData policy for a tick with no
// signal. Returns a transition only when the policy changes the state.
func (i *Instance) EvaluateNoData(rule *Rule, now time.Time) *Transition {
	from := i.startFrom(now)
	switch rule.OnNoData {
	case NoDataIgnore, "":
		// leave untouched (default).
	case NoDataNoData:
		if i.State != StateNoData {
			i.State = StateNoData
			i.Since = now
		}
	case NoDataResolve:
		if i.State == StatePending || i.State == StateFiring {
			i.State = StateInactive
			i.Since = now
		}
	}
	return i.gate(rule, from, i.LastValue, now)
}

// EvaluateError applies the rule's OnError policy for a tick where the
// metric source failed. Returns a transition only when the policy changes
// the state.
func (i *Instance) EvaluateError(rule *Rule, now time.Time) *Transition {
	from := i.startFrom(now)
	switch rule.OnError {
	case ErrorIgnore, "":
		// leave untouched (default).
	case ErrorError:
		if i.State != StateError {
			i.State = StateError
			i.Since = now
		}
	case ErrorResolve:
		if i.State == StatePending || i.State == StateFiring {
			i.State = StateInactive
			i.Since = now
		}
	}
	return i.gate(rule, from, i.LastValue, now)
}

// startFrom captures the incoming state and recovers no_data/error: any
// evaluation entry resumes from inactive so a single healthy tick heals the
// instance.
func (i *Instance) startFrom(now time.Time) State {
	from := i.State
	if from == StateNoData || from == StateError {
		i.State = StateInactive
		i.Since = now
		from = StateInactive
	}
	return from
}

// gate decides whether a notification is warranted: on a state change, or
// when a still-firing instance's re-notification interval has elapsed since
// the last notification attempt.
func (i *Instance) gate(rule *Rule, from State, value float64, now time.Time) *Transition {
	i.LastEvaluated = now
	if i.State == from {
		if from != StateFiring || rule.NotifyIntervalSeconds <= 0 {
			return nil
		}
		if !now.Add(-rule.NotifyInterval()).After(i.LastNotifiedAt) {
			return nil
		}
	}
	i.NotificationDue = true
	return &Transition{From: from, To: i.State, Value: value, At: now, Labels: i.Labels}
}

// NotifyInterval returns the re-notification period for a firing rule.
func (r *Rule) NotifyInterval() time.Duration {
	return time.Duration(r.NotifyIntervalSeconds) * time.Second
}

// StateRepository persists rule instances, keyed per rule by labels hash.
type StateRepository interface {
	// GetByRuleID returns all instances of a rule keyed by labels hash.
	GetByRuleID(ctx context.Context, ruleID string) (map[string]*Instance, error)
	// Upsert saves one instance keyed by rule ID + labels hash.
	Upsert(ctx context.Context, inst *Instance) error
	// ListByOrg returns instances joined to the org's rules, grouped by
	// rule ID then labels hash.
	ListByOrg(ctx context.Context, orgID string) (map[string]map[string]*Instance, error)
	// DeleteByRuleID removes all instances for a rule.
	DeleteByRuleID(ctx context.Context, ruleID string) error
	// DeleteStaleForRule removes instances whose labels hash is not in
	// keep — used when a metric's label set shrinks between ticks.
	DeleteStaleForRule(ctx context.Context, ruleID string, keep []string) error
}

// Notification is emitted when a rule instance fires or resolves.
type Notification struct {
	Rule       *Rule
	Transition *Transition
}

// Notifier delivers firing/resolved notifications. Rules without a webhook
// URL are never passed to the notifier.
type Notifier interface {
	Notify(ctx context.Context, n *Notification) error
}
