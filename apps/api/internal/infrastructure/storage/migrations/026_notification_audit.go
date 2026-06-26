package storage

import (
	"context"
	"database/sql"
)

// migrateCreateNotificationAuditLog creates the notification_audit_log table.
func migrateCreateNotificationAuditLog(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS notification_audit_log (
			id              TEXT PRIMARY KEY,
			operator_id     TEXT NOT NULL,
			event_type      TEXT NOT NULL,
			channel         TEXT NOT NULL,
			payload         TEXT,
			sent_at         INTEGER NOT NULL,
			FOREIGN KEY(operator_id) REFERENCES operators(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying by operator
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_notification_audit_operator 
		ON notification_audit_log(operator_id, sent_at DESC)
	`)
	if err != nil {
		return err
	}

	// Create index for querying by event type
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_notification_audit_type 
		ON notification_audit_log(event_type, sent_at DESC)
	`)
	if err != nil {
		return err
	}

	// Create index for querying by channel
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_notification_audit_channel 
		ON notification_audit_log(channel, sent_at DESC)
	`)

	return err
}
