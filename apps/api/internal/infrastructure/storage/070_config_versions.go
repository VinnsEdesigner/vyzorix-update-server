package storage

import (
	"context"
	"database/sql"
)

func migrateConfigVersions(tx *sql.Tx) error {
	ctx := context.Background()
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS config_versions (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			resource_type TEXT NOT NULL,
			version INTEGER NOT NULL,
			snapshot TEXT NOT NULL,
			changed_by TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL,
			UNIQUE(org_id, resource_type, version)
		)
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_config_versions_lookup ON config_versions(org_id, resource_type, version DESC)`)
	return err
}
