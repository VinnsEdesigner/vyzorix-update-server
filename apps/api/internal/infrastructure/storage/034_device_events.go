package storage

import (
	"context"
	"database/sql"
)

// migrateDeviceEvents creates the device_events table for Diagnostics Timeline.
// This table stores chronological event audit trail for device diagnostics.
func migrateDeviceEvents(tx *sql.Tx) error {
	ctx := context.Background()

	// Create device_events table.
	// Note: SQLite uses TEXT for timestamps (stored as ISO8601 strings) and TEXT for JSON.
	// TIMESTAMPTZ and JSONB are PostgreSQL-specific types not supported by SQLite.
	query := `
	CREATE TABLE IF NOT EXISTS device_events (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		device_id TEXT NOT NULL,
		event_type TEXT NOT NULL,
		timestamp TEXT NOT NULL DEFAULT (datetime('now')),
		data TEXT,
		
		CONSTRAINT fk_device_events_device 
			FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
	)`

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// Create index on device_id and timestamp for efficient timeline queries.
	indexQuery := `
	CREATE INDEX IF NOT EXISTS idx_device_events_device_timestamp 
		ON device_events(device_id, timestamp DESC)`
	_, err = tx.ExecContext(ctx, indexQuery)
	if err != nil {
		return err
	}

	// Create composite index for cursor-based pagination.
	cursorIndex := `
	CREATE INDEX IF NOT EXISTS idx_device_events_cursor 
		ON device_events(device_id, timestamp DESC, id)`
	_, err = tx.ExecContext(ctx, cursorIndex)
	if err != nil {
		return err
	}

	// Create index for event type filtering.
	typeIndex := `
	CREATE INDEX IF NOT EXISTS idx_device_events_type 
		ON device_events(device_id, event_type, timestamp DESC)`
	_, err = tx.ExecContext(ctx, typeIndex)
	if err != nil {
		return err
	}

	return nil
}
