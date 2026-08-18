package storage

import (
	"database/sql"
)

// migrateEmailVerificationTracking adds columns to track email verification delivery status.
func migrateEmailVerificationTracking(tx *sql.Tx) error {
	// Add email_sent_at column to track when verification email was sent.
	_, err := tx.Exec(`
		ALTER TABLE email_verifications ADD COLUMN email_sent_at INTEGER
	`)
	if err != nil && !isColumnExistsError(err) {
		return err
	}

	// Add email_error column to track email delivery failures.
	_, err = tx.Exec(`
		ALTER TABLE email_verifications ADD COLUMN email_error TEXT DEFAULT ''
	`)
	if err != nil && !isColumnExistsError(err) {
		return err
	}

	return nil
}
