package storage

import (
	"context"
	"database/sql"
)

// migrateOperatorFCMToken adds fcm_token column to operators table.
// for storing operator FCM tokens for push notifications.
func migrateOperatorFCMToken(tx *sql.Tx) error {
	ctx := context.Background()

	// Add fcm_token column to operators table if it doesn't exist.
	// SQLite doesn't support ADD COLUMN IF NOT EXISTS, so we check first.
	query := `
	ALTER TABLE operators ADD COLUMN fcm_token TEXT`

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		// Check if column already exists (error message varies by SQLite version).
		// Common error: "duplicate column name: fcm_token".
		if err.Error() != "duplicate column name: fcm_token" {
			return err
		}
		// Column already exists, not an error.
	}

	return nil
}
