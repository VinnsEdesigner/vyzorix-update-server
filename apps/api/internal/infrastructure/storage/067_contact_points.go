package storage

import (
	"context"
	"database/sql"
)

func migrateContactPoints(tx *sql.Tx) error {
	ctx := context.Background()
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS contact_points (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			name TEXT NOT NULL,
			channel TEXT NOT NULL,
			config TEXT NOT NULL DEFAULT '{}',
			secret TEXT NOT NULL DEFAULT '',
			template_id TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_contact_points_org ON contact_points(org_id)`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS notification_deliveries (
			id TEXT PRIMARY KEY,
			contact_point_id TEXT NOT NULL,
			channel TEXT NOT NULL,
			status TEXT NOT NULL,
			error TEXT NOT NULL DEFAULT '',
			payload TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_notification_deliveries_cp ON notification_deliveries(contact_point_id, created_at DESC)`)
	return err
}
