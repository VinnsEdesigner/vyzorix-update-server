package storage

import (
	"database/sql"
)

// migrateEmailVerificationTracking adds columns to track email verification delivery status.
func migrateEmailVerificationTracking(db *sql.DB) error {
	// Add email_sent_at column to track when verification email was sent
	_, err := db.Exec(`
		ALTER TABLE email_verifications ADD COLUMN email_sent_at INTEGER
	`)
	if err != nil {
		// Column might already exist in some SQLite versions
		return nil
	}

	// Add email_error column to track email delivery failures
	_, err = db.Exec(`
		ALTER TABLE email_verifications ADD COLUMN email_error TEXT DEFAULT ''
	`)
	if err != nil {
		// Column might already exist in some SQLite versions
		return nil
	}

	return nil
}
