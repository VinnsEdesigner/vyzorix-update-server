package storage

import (
	"context"
	"database/sql"
)

// migrateCreateDeviceLogsAndEvents creates device_logs and device_events tables.
func migrateCreateDeviceLogsAndEvents(db *sql.DB) error {
	// Create device_logs table for device-side logging
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS device_logs (
			id              TEXT PRIMARY KEY,
			device_id       TEXT NOT NULL,
			log_level       TEXT NOT NULL,
			message         TEXT NOT NULL,
			metadata        TEXT,
			created_at      INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying logs by device
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_device_logs_device 
		ON device_logs(device_id, created_at DESC)
	`)
	if err != nil {
		return err
	}

	// Create index for querying logs by level
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_device_logs_level 
		ON device_logs(log_level, created_at DESC)
	`)
	if err != nil {
		return err
	}

	// Create device_events table for device lifecycle events
	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS device_events (
			id              TEXT PRIMARY KEY,
			device_id       TEXT NOT NULL,
			event_type      TEXT NOT NULL,
			payload         TEXT,
			created_at      INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying events by device
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_device_events_device 
		ON device_events(device_id, created_at DESC)
	`)
	if err != nil {
		return err
	}

	// Create index for querying events by type
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_device_events_type 
		ON device_events(event_type, created_at DESC)
	`)

	return err
}
