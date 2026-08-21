package storage

import (
	"context"
	"database/sql"
)

// migrateAlertLabelInstanceKeys rebuilds alert_instances with a composite
// (rule_id, labels_hash) primary key so one rule can track several labeled
// series, and adds the rule no-data/error handling policies. Existing
// instances keep their state with an empty labels hash.
func migrateAlertLabelInstanceKeys(tx *sql.Tx) error {
	ctx := context.Background()

	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS alert_instances_new (
			rule_id TEXT NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
			labels_hash TEXT NOT NULL DEFAULT '',
			labels TEXT NOT NULL DEFAULT '{}',
			state TEXT NOT NULL DEFAULT 'inactive',
			since INTEGER NOT NULL DEFAULT 0,
			last_evaluated INTEGER NOT NULL DEFAULT 0,
			last_value REAL NOT NULL DEFAULT 0,
			last_notified_at INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (rule_id, labels_hash)
		)
	`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO alert_instances_new (rule_id, labels_hash, labels, state, since, last_evaluated, last_value, last_notified_at)
		SELECT rule_id, '', '{}', state, since, last_evaluated, last_value,
			COALESCE(last_notified_at, 0)
		FROM alert_instances
	`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DROP TABLE alert_instances`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE alert_instances_new RENAME TO alert_instances`); err != nil {
		return err
	}
	if err := addColumnIfMissing(tx, "alert_rules", "on_no_data",
		`ALTER TABLE alert_rules ADD COLUMN on_no_data TEXT NOT NULL DEFAULT ''`); err != nil {
		return err
	}
	return addColumnIfMissing(tx, "alert_rules", "on_error",
		`ALTER TABLE alert_rules ADD COLUMN on_error TEXT NOT NULL DEFAULT ''`)
}

// addColumnIfMissing applies alterSQL only when the column is absent, so
// the migration stays idempotent on partially-applied databases.
func addColumnIfMissing(tx *sql.Tx, table, column, alterSQL string) error {
	cols, err := tx.QueryContext(context.Background(), `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return err
	}
	defer func() { _ = cols.Close() }()
	for cols.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		scanErr := cols.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk)
		if scanErr != nil {
			return scanErr
		}
		if name == column {
			return nil
		}
	}
	if scanErr := cols.Err(); scanErr != nil {
		return scanErr
	}
	_, err = tx.ExecContext(context.Background(), alterSQL)
	return err
}
