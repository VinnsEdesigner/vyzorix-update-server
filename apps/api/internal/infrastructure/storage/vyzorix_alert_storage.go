package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
)

// AlertRuleRepository is the SQL persistence for alert rules.
type AlertRuleRepository struct {
	db *sql.DB
}

// NewAlertRuleRepository creates a new AlertRuleRepository.
func NewAlertRuleRepository(db *sql.DB) *AlertRuleRepository {
	return &AlertRuleRepository{db: db}
}

// Save upserts a rule.
func (r *AlertRuleRepository) Save(ctx context.Context, rule *alert.Rule) error {
	query := `
		INSERT INTO alert_rules (id, org_id, name, metric, condition, threshold, for_seconds, notify_interval_seconds, enabled, webhook_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			metric = excluded.metric,
			condition = excluded.condition,
			threshold = excluded.threshold,
			for_seconds = excluded.for_seconds,
			notify_interval_seconds = excluded.notify_interval_seconds,
			enabled = excluded.enabled,
			webhook_url = excluded.webhook_url,
			updated_at = excluded.updated_at
	`
	enabled := 0
	if rule.Enabled {
		enabled = 1
	}
	_, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.OrgID, rule.Name, string(rule.Metric), string(rule.Condition),
		rule.Threshold, rule.ForSeconds, rule.NotifyIntervalSeconds, enabled, rule.WebhookURL,
		rule.CreatedAt.UnixMilli(), rule.UpdatedAt.UnixMilli(),
	)
	return err
}

const alertRuleColumns = `id, org_id, name, metric, condition, threshold, for_seconds, notify_interval_seconds, enabled, webhook_url, created_at, updated_at`

func scanAlertRule(scanner interface{ Scan(...any) error }) (*alert.Rule, error) {
	var rule alert.Rule
	var metric, condition string
	var enabled int
	var createdAt, updatedAt int64
	err := scanner.Scan(
		&rule.ID, &rule.OrgID, &rule.Name, &metric, &condition,
		&rule.Threshold, &rule.ForSeconds, &rule.NotifyIntervalSeconds, &enabled, &rule.WebhookURL,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	rule.Metric = alert.Metric(metric)
	rule.Condition = alert.Condition(condition)
	rule.Enabled = enabled == 1
	rule.CreatedAt = time.UnixMilli(createdAt)
	rule.UpdatedAt = time.UnixMilli(updatedAt)
	return &rule, nil
}

// GetByID returns a rule or alert.ErrNotFound.
func (r *AlertRuleRepository) GetByID(ctx context.Context, id string) (*alert.Rule, error) {
	rule, err := scanAlertRule(r.db.QueryRowContext(ctx,
		`SELECT `+alertRuleColumns+` FROM alert_rules WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return nil, alert.ErrNotFound
	}
	return rule, err
}

// ListByOrg returns all rules of an org ordered by name.
func (r *AlertRuleRepository) ListByOrg(ctx context.Context, orgID string) ([]*alert.Rule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+alertRuleColumns+` FROM alert_rules WHERE org_id = ? ORDER BY name`, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var rules []*alert.Rule
	for rows.Next() {
		rule, err := scanAlertRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// ListEnabled returns all enabled rules across orgs.
func (r *AlertRuleRepository) ListEnabled(ctx context.Context) ([]*alert.Rule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+alertRuleColumns+` FROM alert_rules WHERE enabled = 1 ORDER BY org_id, name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var rules []*alert.Rule
	for rows.Next() {
		rule, err := scanAlertRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// Delete removes a rule and its instance, returning whether the rule existed.
func (r *AlertRuleRepository) Delete(ctx context.Context, id string) (bool, error) {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM alert_instances WHERE rule_id = ?`, id); err != nil {
		return false, err
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM alert_rules WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// AlertStateRepository is the SQL persistence for alert rule instances.
type AlertStateRepository struct {
	db *sql.DB
}

// NewAlertStateRepository creates a new AlertStateRepository.
func NewAlertStateRepository(db *sql.DB) *AlertStateRepository {
	return &AlertStateRepository{db: db}
}

// GetByRuleID returns the instance for a rule. When the rule was never
// evaluated a fresh inactive instance comes back — evaluation order between
// "no row yet" and "stored inactive" is not observable by callers.
func (r *AlertStateRepository) GetByRuleID(ctx context.Context, ruleID string) (*alert.Instance, error) {
	inst, err := scanAlertInstance(r.db.QueryRowContext(ctx,
		`SELECT rule_id, state, since, last_evaluated, last_value, last_notified_at FROM alert_instances WHERE rule_id = ?`, ruleID))
	if err == sql.ErrNoRows {
		return alert.NewInstance(ruleID), nil
	}
	return inst, err
}

// Upsert saves the instance keyed by rule ID.
func (r *AlertStateRepository) Upsert(ctx context.Context, inst *alert.Instance) error {
	query := `
		INSERT INTO alert_instances (rule_id, state, since, last_evaluated, last_value, last_notified_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(rule_id) DO UPDATE SET
			state = excluded.state,
			since = excluded.since,
			last_evaluated = excluded.last_evaluated,
			last_value = excluded.last_value,
			last_notified_at = excluded.last_notified_at
	`
	_, err := r.db.ExecContext(ctx, query,
		inst.RuleID, string(inst.State), inst.Since.UnixMilli(),
		inst.LastEvaluated.UnixMilli(), inst.LastValue, inst.LastNotifiedAt.UnixMilli(),
	)
	return err
}

// ListByOrg returns instances joined to the org's rules, keyed by rule ID.
func (r *AlertStateRepository) ListByOrg(ctx context.Context, orgID string) (map[string]*alert.Instance, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT i.rule_id, i.state, i.since, i.last_evaluated, i.last_value
		FROM alert_instances i
		JOIN alert_rules ar ON ar.id = i.rule_id
		WHERE ar.org_id = ?
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	instances := make(map[string]*alert.Instance)
	for rows.Next() {
		inst, err := scanAlertInstance(rows)
		if err != nil {
			return nil, err
		}
		instances[inst.RuleID] = inst
	}
	return instances, rows.Err()
}

// DeleteByRuleID removes the instance for a rule.
func (r *AlertStateRepository) DeleteByRuleID(ctx context.Context, ruleID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM alert_instances WHERE rule_id = ?`, ruleID)
	return err
}

func scanAlertInstance(scanner interface{ Scan(...any) error }) (*alert.Instance, error) {
	var inst alert.Instance
	var state string
	var since, lastEvaluated, lastNotified int64
	err := scanner.Scan(&inst.RuleID, &state, &since, &lastEvaluated, &inst.LastValue, &lastNotified)
	if err != nil {
		return nil, err
	}
	inst.State = alert.State(state)
	inst.Since = time.UnixMilli(since)
	inst.LastEvaluated = time.UnixMilli(lastEvaluated)
	inst.LastNotifiedAt = time.UnixMilli(lastNotified)
	return &inst, nil
}

// AlertHistoryRepository is the SQL persistence for transition events.
type AlertHistoryRepository struct {
	db *sql.DB
}

func NewAlertHistoryRepository(db *sql.DB) *AlertHistoryRepository {
	return &AlertHistoryRepository{db: db}
}

func (r *AlertHistoryRepository) Append(ctx context.Context, evt *alert.Event) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO alert_events (id, rule_id, from_state, to_state, value, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		evt.ID, evt.RuleID, string(evt.FromState), string(evt.ToState), evt.Value, evt.CreatedAt,
	)
	return err
}

func (r *AlertHistoryRepository) ListByOrg(ctx context.Context, orgID, ruleID string, limit int) ([]*alert.Event, error) {
	if limit <= 0 {
		limit = 200
	}
	query := `
SELECT e.id, e.rule_id, e.from_state, e.to_state, e.value, e.created_at
FROM alert_events e
JOIN alert_rules ar ON ar.id = e.rule_id
WHERE ar.org_id = ?`
	args := []interface{}{orgID}
	if ruleID != "" {
		query += " AND e.rule_id = ?"
		args = append(args, ruleID)
	}
	query += " ORDER BY e.created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []*alert.Event
	for rows.Next() {
		var evt alert.Event
		var fromState, toState string
		if err := rows.Scan(&evt.ID, &evt.RuleID, &fromState, &toState, &evt.Value, &evt.CreatedAt); err != nil {
			return nil, err
		}
		evt.FromState = alert.State(fromState)
		evt.ToState = alert.State(toState)
		events = append(events, &evt)
	}
	return events, rows.Err()
}
