package storage

import (
	"context"
	"database/sql"
)

// migrateDashboardDeviceLogs creates the device_logs table for dashboard logs API.
// This table stores device event logs with cursor-based pagination support.
func migrateDashboardDeviceLogs(db *sql.DB) error {
	// Create device_logs table for device event logs.
	// NOTE: Using INTEGER for timestamp (Unix milliseconds) for SQLite compatibility.
	// NOTE: Using TEXT for data (JSON) since SQLite doesn't have native JSONB.
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS device_logs (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			data TEXT,
			FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying logs by device with timestamp ordering.
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_device_logs_device_timestamp
		ON device_logs(device_id, timestamp DESC)
	`)
	if err != nil {
		return err
	}

	// Create composite index for cursor-based pagination (device_id, timestamp, id).
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_device_logs_cursor
		ON device_logs(device_id, timestamp DESC, id)
	`)
	if err != nil {
		return err
	}

	// Create index for filtering by event type.
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_device_logs_event_type
		ON device_logs(event_type)
	`)

	return err
}
