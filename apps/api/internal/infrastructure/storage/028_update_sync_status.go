package storage

import (
	"context"
	"database/sql"
)

// migrateCreateUpdateSyncStatus creates the update_sync_status table.
func migrateCreateUpdateSyncStatus(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS update_sync_status (
			id                INTEGER PRIMARY KEY CHECK (id = 1),
			status            TEXT NOT NULL DEFAULT 'idle',
			last_sync_at      INTEGER,
			next_sync_at      INTEGER,
			last_error        TEXT,
			versions_found    INTEGER DEFAULT 0,
			syncing_since     INTEGER,
			created_at        INTEGER NOT NULL,
			updated_at        INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Insert default row if not exists.
	_, err = tx.ExecContext(context.Background(), `
		INSERT OR IGNORE INTO update_sync_status (id, status, created_at, updated_at)
		VALUES (1, 'idle', ?, ?)
	`, sql.NullInt64{}, sql.NullInt64{})

	return err
}
