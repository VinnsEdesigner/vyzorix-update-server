package storage

import (
	"context"
	"database/sql"
)

func migrateAlertHistoryColumns(tx *sql.Tx) error {
	ctx := context.Background()
	_, err := tx.ExecContext(ctx, `ALTER TABLE alert_rules ADD COLUMN notify_interval_seconds INTEGER NOT NULL DEFAULT 0`)
	if err != nil && !isDuplicateColumn(err) {
		return err
	}
	_, err = tx.ExecContext(ctx, `ALTER TABLE alert_instances ADD COLUMN last_notified_at INTEGER NOT NULL DEFAULT 0`)
	if err != nil && !isDuplicateColumn(err) {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS alert_events (
			id TEXT PRIMARY KEY,
			rule_id TEXT NOT NULL,
			from_state TEXT NOT NULL,
			to_state TEXT NOT NULL,
			value REAL NOT NULL,
			created_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_alert_events_rule ON alert_events(rule_id, created_at DESC)`)
	return err
}
