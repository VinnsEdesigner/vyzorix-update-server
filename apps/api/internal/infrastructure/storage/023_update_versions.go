package storage

import (
	"context"
	"database/sql"
)

// migrateCreateUpdateTables creates update_versions, update_pushes, and update_push_devices tables.
func migrateCreateUpdateTables(tx *sql.Tx) error {
	// Create update_versions table.
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS update_versions (
			id              TEXT PRIMARY KEY,
			version         TEXT NOT NULL UNIQUE,
			apk_filename    TEXT NOT NULL,
			apk_size        INTEGER NOT NULL,
			sha256          TEXT NOT NULL,
			release_date    INTEGER NOT NULL,
			release_notes   TEXT,
			release_type    TEXT NOT NULL DEFAULT 'minor',
			is_latest       INTEGER NOT NULL DEFAULT 0,
			created_at      INTEGER NOT NULL,
			updated_at      INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying versions by date.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_versions_date 
		ON update_versions(release_date DESC)
	`)
	if err != nil {
		return err
	}

	// Create index for querying latest version.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_versions_latest 
		ON update_versions(is_latest)
	`)
	if err != nil {
		return err
	}

	// Create update_pushes table.
	_, err = tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS update_pushes (
			id              TEXT PRIMARY KEY,
			version_id      TEXT NOT NULL,
			install_type    TEXT NOT NULL DEFAULT 'immediate',
			scheduled_at    INTEGER,
			status          TEXT NOT NULL DEFAULT 'pending',
			initiated_by   TEXT NOT NULL,
			initiated_at    INTEGER NOT NULL,
			completed_at    INTEGER,
			cancelled_at    INTEGER,
			cancelled_by    TEXT,
			FOREIGN KEY(version_id) REFERENCES update_versions(id),
			FOREIGN KEY(initiated_by) REFERENCES operators(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying pushes by status.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_update_pushes_status 
		ON update_pushes(status)
	`)
	if err != nil {
		return err
	}

	// Create index for querying pushes by initiation time.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_update_pushes_initiated_at 
		ON update_pushes(initiated_at DESC)
	`)
	if err != nil {
		return err
	}

	// Create update_push_devices table.
	_, err = tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS update_push_devices (
			id              TEXT PRIMARY KEY,
			push_id         TEXT NOT NULL,
			device_id       TEXT NOT NULL,
			status          TEXT NOT NULL DEFAULT 'pending',
			sent_at         INTEGER,
			acknowledged_at INTEGER,
			error           TEXT,
			retry_count     INTEGER NOT NULL DEFAULT 0,
			created_at      INTEGER NOT NULL,
			updated_at      INTEGER NOT NULL,
			FOREIGN KEY(push_id) REFERENCES update_pushes(id) ON DELETE CASCADE,
			FOREIGN KEY(device_id) REFERENCES devices(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create unique index for push + device combination.
	_, err = tx.ExecContext(context.Background(), `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_push_device 
		ON update_push_devices(push_id, device_id)
	`)
	if err != nil {
		return err
	}

	// Create index for querying push devices by status.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_push_devices_status 
		ON update_push_devices(status)
	`)

	return err
}
