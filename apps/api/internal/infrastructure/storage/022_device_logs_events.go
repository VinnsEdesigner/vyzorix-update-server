package storage

import (
	"context"
	"database/sql"
)

// migrateCreateDeviceLogsAndEvents creates device_logs table.
// NOTE: device_events table is created by migration 034 (migrateDeviceEvents).
// to ensure consistent schema with timestamp and data columns (TEXT in SQLite).
func migrateCreateDeviceLogsAndEvents(db *sql.DB) error {
	// Create device_logs table per SERVER_BACKEND_DASHBOARD_COMMANDS_API.md Section 4.1.
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS device_logs (
			id              TEXT PRIMARY KEY,
			device_id       TEXT NOT NULL,
			event_type      TEXT NOT NULL,
			timestamp       INTEGER NOT NULL,
			data            TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying logs by device and timestamp (cursor-based pagination).
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_device_logs_device_timestamp 
		ON device_logs(device_id, timestamp DESC, id)
	`)
	if err != nil {
		return err
	}

	// Create index for querying logs by event type.
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_device_logs_event_type 
		ON device_logs(event_type)
	`)
	if err != nil {
		return err
	}

	return err
}
