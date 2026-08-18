package storage

import (
	"context"
	"database/sql"
)

// migrateCreateEvents creates the events table for real-time dashboard broadcasting.
func migrateCreateEvents(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS events (
			id              TEXT PRIMARY KEY,
			event_type      TEXT NOT NULL,
			device_id       TEXT,
			operator_id     TEXT,
			payload         TEXT,
			created_at      INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Index for querying events by device.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_events_device_id 
		ON events(device_id, created_at DESC)
	`)
	if err != nil {
		return err
	}

	// Index for querying events by operator.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_events_operator_id 
		ON events(operator_id, created_at DESC)
	`)
	if err != nil {
		return err
	}

	// Index for querying events by type.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_events_type 
		ON events(event_type, created_at DESC)
	`)

	return err
}
