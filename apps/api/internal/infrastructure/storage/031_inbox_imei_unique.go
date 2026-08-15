package storage

import (
	"context"
	"database/sql"
)

// migrateInboxIMEIUnique adds a unique constraint on device_imei in inbox_requests.
// This migration handles existing duplicates by keeping the oldest entry per IMEI.
func migrateInboxIMEIUnique(db *sql.DB) error {
	ctx := context.Background()

	// Step 1: Find and remove duplicate IMEIs, keeping the oldest entry.
	// This query finds all MIN(id) per device_imei and deletes everything else.
	cleanupQuery := `
		DELETE FROM inbox_requests 
		WHERE id NOT IN (
			SELECT MIN(id) 
			FROM inbox_requests 
			GROUP BY device_imei
		)
	`
	_, err := db.ExecContext(ctx, cleanupQuery)
	if err != nil {
		return err
	}

	// Step 2: Add unique constraint on device_imei.
	// SQLite doesn't support DROP COLUMN or ADD CONSTRAINT in the same way as PostgreSQL,.
	// so we recreate the table with the unique constraint.
	// However, SQLite does support UNIQUE constraints on columns directly.

	// First, check if the unique index already exists.
	var count int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_inbox_imei_unique'").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Create unique index on device_imei.
		_, err = db.ExecContext(ctx, `
			CREATE UNIQUE INDEX IF NOT EXISTS idx_inbox_imei_unique 
			ON inbox_requests(device_imei)
		`)
		if err != nil {
			return err
		}
	}

	return nil
}
