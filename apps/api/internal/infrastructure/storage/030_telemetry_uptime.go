package storage

import (
	"context"
	"database/sql"
)

// migrateTelemetryUptime adds uptime column to telemetry table.
// This enables the metrics API to return uptime statistics per spec.
func migrateTelemetryUptime(db *sql.DB) error {
	// Check if uptime column already exists.
	var count int
	err := db.QueryRowContext(context.Background(), 
		"SELECT COUNT(*) FROM pragma_table_info('telemetry') WHERE name = 'uptime'",
	).Scan(&count)
	if err != nil {
		return err
	}

	// Column already exists, skip migration.
	if count > 0 {
		return nil
	}

	// Add uptime column as INTEGER (seconds).
	_, err = db.ExecContext(context.Background(), `
		ALTER TABLE telemetry ADD COLUMN uptime INTEGER DEFAULT 0
	`)

	return err
}
