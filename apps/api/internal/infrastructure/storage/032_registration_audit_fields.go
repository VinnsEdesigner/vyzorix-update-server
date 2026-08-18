package storage

import (
	"context"
	"database/sql"
	"strings"
)

// migrateRegistrationAuditFields adds client_ip and user_agent columns to registration_logs table.

func migrateRegistrationAuditFields(tx *sql.Tx) error {
	ctx := context.Background()

	// Add client_ip column if not exists.
	_, err := tx.ExecContext(ctx, `
		ALTER TABLE registration_logs ADD COLUMN client_ip TEXT
	`)
	if err != nil && !strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return err
	}

	// Add user_agent column if not exists (same approach).
	_, err = tx.ExecContext(ctx, `
		ALTER TABLE registration_logs ADD COLUMN user_agent TEXT
	`)
	if err != nil && !strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return err
	}

	return nil
}
