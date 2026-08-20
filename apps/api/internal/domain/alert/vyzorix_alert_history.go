package alert

import "context"

// Event records one state transition for audit/history queries. The
// alert_instances table holds only current state; history lives here.
type Event struct {
	ID        string
	RuleID    string
	FromState State
	ToState   State
	Value     float64
	CreatedAt int64
}

// HistoryRepository persists transition events.
type HistoryRepository interface {
	Append(ctx context.Context, evt *Event) error
	ListByOrg(ctx context.Context, orgID, ruleID string, limit int) ([]*Event, error)
}
