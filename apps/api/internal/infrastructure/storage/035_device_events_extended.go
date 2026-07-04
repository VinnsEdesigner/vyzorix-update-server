package storage

import (
	"context"
	"database/sql"
)

// migrateDeviceEventsExtended adds extended columns to device_events for real-time events.
func migrateDeviceEventsExtended(db *sql.DB) error {
	ctx := context.Background()

	// Add severity column if not exists
	_, err := db.ExecContext(ctx, `
		ALTER TABLE device_events ADD COLUMN IF NOT EXISTS severity TEXT DEFAULT 'info'`)
	if err != nil {
		// Ignore error if column already exists (SQLite doesn't support IF NOT EXISTS for columns)
		// Check if column exists first
		var count int
		checkQuery := `SELECT COUNT(*) FROM pragma_table_info('device_events') WHERE name = 'severity'`
		if scanErr := db.QueryRowContext(ctx, checkQuery).Scan(&count); scanErr == nil && count == 0 {
			return err
		}
	}

	// Add source column if not exists
	_, err = db.ExecContext(ctx, `
		ALTER TABLE device_events ADD COLUMN IF NOT EXISTS source TEXT DEFAULT 'server'`)
	if err != nil {
		var count int
		checkQuery := `SELECT COUNT(*) FROM pragma_table_info('device_events') WHERE name = 'source'`
		if scanErr := db.QueryRowContext(ctx, checkQuery).Scan(&count); scanErr == nil && count == 0 {
			return err
		}
	}

	// Add operator_id column if not exists
	_, err = db.ExecContext(ctx, `
		ALTER TABLE device_events ADD COLUMN IF NOT EXISTS operator_id TEXT`)
	if err != nil {
		var count int
		checkQuery := `SELECT COUNT(*) FROM pragma_table_info('device_events') WHERE name = 'operator_id'`
		if scanErr := db.QueryRowContext(ctx, checkQuery).Scan(&count); scanErr == nil && count == 0 {
			return err
		}
	}

	// Create index on severity for filtering
	severityIndex := `
	CREATE INDEX IF NOT EXISTS idx_device_events_severity 
		ON device_events(device_id, severity, timestamp DESC)`
	_, err = db.ExecContext(ctx, severityIndex)
	if err != nil {
		return err
	}

	// Create index on source for filtering
	sourceIndex := `
	CREATE INDEX IF NOT EXISTS idx_device_events_source 
		ON device_events(device_id, source, timestamp DESC)`
	_, err = db.ExecContext(ctx, sourceIndex)
	if err != nil {
		return err
	}

	return nil
}
