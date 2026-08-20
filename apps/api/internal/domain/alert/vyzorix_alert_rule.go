// Package alert provides the alerting domain: org-scoped rules that evaluate
// fleet metrics against thresholds and a state machine that tracks each rule's
// instance through inactive → pending → firing → resolved transitions.
package alert

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNotFound is returned when an alert rule is not found.
var ErrNotFound = errors.New("alert rule not found")

// Metric is a fleet-level measurement a rule can evaluate.
type Metric string

const (
	// MetricDeviceOfflineCount counts devices in the org currently offline.
	MetricDeviceOfflineCount Metric = "device_offline_count"
	// MetricDeviceOfflinePercent is the share of the org's fleet that is offline (0-100).
	MetricDeviceOfflinePercent Metric = "device_offline_percent"
	// MetricCommandFailureRate is the percent of commands failed within the
	// evaluation window (0-100).
	MetricCommandFailureRate Metric = "command_failure_rate"
)

// Valid reports whether the metric is a known measurement.
func (m Metric) Valid() bool {
	switch m {
	case MetricDeviceOfflineCount, MetricDeviceOfflinePercent, MetricCommandFailureRate:
		return true
	}
	return false
}

// Condition compares an observed value against the rule threshold.
type Condition string

const (
	ConditionGt  Condition = "gt"
	ConditionGte Condition = "gte"
	ConditionLt  Condition = "lt"
	ConditionLte Condition = "lte"
)

// Valid reports whether the condition is a known comparator.
func (c Condition) Valid() bool {
	switch c {
	case ConditionGt, ConditionGte, ConditionLt, ConditionLte:
		return true
	}
	return false
}

// Breached reports whether the observed value satisfies the condition against
// the threshold.
func (c Condition) Breached(value, threshold float64) bool {
	switch c {
	case ConditionGt:
		return value > threshold
	case ConditionGte:
		return value >= threshold
	case ConditionLt:
		return value < threshold
	case ConditionLte:
		return value <= threshold
	}
	return false
}

// Rule is an org-scoped alert definition evaluated on a fixed cadence.
type Rule struct {
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ID         string
	OrgID      string
	Name       string
	WebhookURL string
	Metric     Metric
	Condition  Condition
	Threshold  float64
	// ForSeconds is how long the condition must hold before firing
	// (0 = fire immediately on breach).
	ForSeconds int
	// NotifyIntervalSeconds re-notifies while still firing after this elapsed
	// (0 = notify once at firing and once at resolve).
	NotifyIntervalSeconds int
	Enabled               bool
}

// Validate checks the rule is well-formed and internally consistent.
func (r *Rule) Validate() error {
	if strings.TrimSpace(r.OrgID) == "" {
		return errors.New("org_id is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if !r.Metric.Valid() {
		return fmt.Errorf("invalid metric %q", r.Metric)
	}
	if !r.Condition.Valid() {
		return fmt.Errorf("invalid condition %q", r.Condition)
	}
	if r.ForSeconds < 0 {
		return errors.New("for_seconds cannot be negative")
	}
	if r.NotifyIntervalSeconds < 0 {
		return errors.New("notify_interval_seconds cannot be negative")
	}
	return nil
}

// ForDuration returns the pending duration before a breached rule fires.
func (r *Rule) ForDuration() time.Duration {
	return time.Duration(r.ForSeconds) * time.Second
}

// Repository persists alert rules.
type Repository interface {
	// Save upserts a rule.
	Save(ctx context.Context, rule *Rule) error
	// GetByID returns a rule or ErrNotFound.
	GetByID(ctx context.Context, id string) (*Rule, error)
	// ListByOrg returns all rules of an org ordered by name.
	ListByOrg(ctx context.Context, orgID string) ([]*Rule, error)
	// ListEnabled returns all enabled rules across orgs (used by the evaluator).
	ListEnabled(ctx context.Context) ([]*Rule, error)
	// Delete removes a rule, returning whether it existed.
	Delete(ctx context.Context, id string) (bool, error)
}
