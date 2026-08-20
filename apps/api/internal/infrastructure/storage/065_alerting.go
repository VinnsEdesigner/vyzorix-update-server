package storage

import (
	"context"
	"database/sql"
)

func migrateCreateAlerting(tx *sql.Tx) error {
	ctx := context.Background()
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS alert_rules (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			name TEXT NOT NULL,
			metric TEXT NOT NULL,
			condition TEXT NOT NULL,
			threshold REAL NOT NULL,
			for_seconds INTEGER NOT NULL DEFAULT 0,
			enabled INTEGER NOT NULL DEFAULT 1,
			webhook_url TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_alert_rules_org ON alert_rules(org_id)`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS alert_instances (
			rule_id TEXT PRIMARY KEY REFERENCES alert_rules(id) ON DELETE CASCADE,
			state TEXT NOT NULL DEFAULT 'inactive',
			since INTEGER NOT NULL DEFAULT 0,
			last_evaluated INTEGER NOT NULL DEFAULT 0,
			last_value REAL NOT NULL DEFAULT 0
		)
	`)
	return err
}
