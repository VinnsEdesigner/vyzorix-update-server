package alert

import (
	"context"
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
)

// Valid reports whether the state is a known instance state.
func (s State) Valid() bool {
	switch s {
	case StateInactive, StatePending, StateFiring:
		return true
	}
	return false
}

// Instance tracks the runtime state of one rule. There is exactly one
// instance per rule (fleet-level metrics have no label dimensions).
type Instance struct {
	Since           time.Time
	LastEvaluated   time.Time
	LastNotifiedAt  time.Time
	RuleID          string
	State           State
	LastValue       float64
	NotificationDue bool
}

// NewInstance returns the initial (inactive) instance for a rule.
func NewInstance(ruleID string) *Instance {
	return &Instance{RuleID: ruleID, State: StateInactive}
}

// Transition describes a state change produced by one evaluation.
type Transition struct {
	At    time.Time
	From  State
	To    State
	Value float64
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
	i.LastEvaluated = now
	i.LastValue = value

	breached := rule.Condition.Breached(value, rule.Threshold)
	from := i.State

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

	if i.State == from && rule.NotifyIntervalSeconds <= 0 {
		return nil
	}
	if i.State == from && rule.NotifyIntervalSeconds > 0 && from == StateFiring && !now.Add(-rule.NotifyInterval()).After(i.LastNotifiedAt) {
		return nil
	}
	i.NotificationDue = true
	i.LastNotifiedAt = now
	return &Transition{From: from, To: i.State, Value: value, At: now}
}

// NotifyInterval returns the re-notification period for a firing rule.
func (r *Rule) NotifyInterval() time.Duration {
	return time.Duration(r.NotifyIntervalSeconds) * time.Second
}

// StateRepository persists rule instances.
type StateRepository interface {
	// GetByRuleID returns the instance for a rule, or a fresh inactive
	// instance when the rule was never evaluated.
	GetByRuleID(ctx context.Context, ruleID string) (*Instance, error)
	// Upsert saves the instance keyed by rule ID.
	Upsert(ctx context.Context, inst *Instance) error
	// ListByOrg returns instances joined to the org's rules, keyed by rule ID.
	ListByOrg(ctx context.Context, orgID string) (map[string]*Instance, error)
	// DeleteByRuleID removes the instance for a rule.
	DeleteByRuleID(ctx context.Context, ruleID string) error
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
