package storage

import (
	"context"
	"database/sql"
)

func migrateAnnotations(tx *sql.Tx) error {
	ctx := context.Background()
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS annotations (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			title TEXT NOT NULL,
			text TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '[]',
			source TEXT NOT NULL DEFAULT '',
			start_time INTEGER NOT NULL,
			end_time INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_annotations_org ON annotations(org_id, start_time DESC)`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_annotations_tags ON annotations(tags)`)
	return err
}
