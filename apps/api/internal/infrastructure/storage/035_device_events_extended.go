package storage

import (
	"context"
	"database/sql"
)

// migrateDeviceEventsExtended adds extended columns to device_events for real-time events.
func migrateDeviceEventsExtended(db *sql.DB) error {
	ctx := context.Background()

	// Add severity column if not exists (with idempotent error handling for SQLite)
	_, err := db.ExecContext(ctx, `ALTER TABLE device_events ADD COLUMN severity TEXT DEFAULT 'info'`)
	if err != nil {
		if !isColumnExistsError(err) {
			return err
		}
	}

	// Add source column if not exists (with idempotent error handling for SQLite)
	_, err = db.ExecContext(ctx, `ALTER TABLE device_events ADD COLUMN source TEXT DEFAULT 'server'`)
	if err != nil {
		if !isColumnExistsError(err) {
			return err
		}
	}

	// Add operator_id column if not exists (with idempotent error handling for SQLite)
	_, err = db.ExecContext(ctx, `ALTER TABLE device_events ADD COLUMN operator_id TEXT`)
	if err != nil {
		if !isColumnExistsError(err) {
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
