package storage

import (
	"context"
	"database/sql"
	"errors"
)

// migrateCommandSecretHash adds command_secret_hash column to inbox_requests.
// Stores only the hash of the command secret for security.
func migrateCommandSecretHash(tx *sql.Tx) error {
	// Check if column already exists.
	var exists int
	err := tx.QueryRowContext(context.Background(),
		"SELECT 1 FROM pragma_table_info('inbox_requests') WHERE name = 'command_secret_hash'").Scan(&exists)
	if err == nil {
		// Column already exists.
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Add command_secret_hash column.
	_, err = tx.ExecContext(context.Background(), `
		ALTER TABLE inbox_requests ADD COLUMN command_secret_hash TEXT
	`)
	if err != nil {
		return err
	}

	return nil
}
