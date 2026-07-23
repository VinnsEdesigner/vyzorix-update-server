package storage

import (
	"context"
	"database/sql"
	"errors"
)

// migrateConfirmedAt adds confirmed_at column to inbox_requests.
// Tracks when a device confirmed its registration (single-use token).
func migrateConfirmedAt(db *sql.DB) error {
	// Check if column already exists.
	var exists int
	err := db.QueryRowContext(context.Background(),
		"SELECT 1 FROM pragma_table_info('inbox_requests') WHERE name = 'confirmed_at'").Scan(&exists)
	if err == nil {
		// Column already exists.
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Add confirmed_at column.
	_, err = db.ExecContext(context.Background(), `
		ALTER TABLE inbox_requests ADD COLUMN confirmed_at INTEGER
	`)
	return err
}
