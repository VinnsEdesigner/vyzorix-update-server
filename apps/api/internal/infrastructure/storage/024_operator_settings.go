package storage

import (
	"context"
	"database/sql"
)

// migrateCreateOperatorSettings creates the operator_settings table.
func migrateCreateOperatorSettings(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS operator_settings (
			operator_id              TEXT PRIMARY KEY,
			server_url               TEXT,
			device_id                TEXT,
			request_timeout_ms        INTEGER NOT NULL DEFAULT 8000,
			auto_reconnect           INTEGER NOT NULL DEFAULT 1,
			strict_hmac              INTEGER NOT NULL DEFAULT 0,
			log_buffer_limit         INTEGER NOT NULL DEFAULT 500,
			signal_history_limit     INTEGER NOT NULL DEFAULT 240,
			risk_warn                INTEGER NOT NULL DEFAULT 70,
			risk_crit                INTEGER NOT NULL DEFAULT 85,
			thermal_warn             INTEGER NOT NULL DEFAULT 45,
			thermal_crit             INTEGER NOT NULL DEFAULT 50,
			buffer_warn              INTEGER NOT NULL DEFAULT 30,
			buffer_crit              INTEGER NOT NULL DEFAULT 15,
			notifications_enabled    INTEGER NOT NULL DEFAULT 1,
			notify_email             TEXT,
			notify_push              INTEGER NOT NULL DEFAULT 1,
			notify_webhook           INTEGER NOT NULL DEFAULT 0,
			webhook_url              TEXT,
			webhook_secret           TEXT,
			webhook_types            TEXT,
			notify_threshold_breach  INTEGER NOT NULL DEFAULT 1,
			notify_device_offline    INTEGER NOT NULL DEFAULT 1,
			notify_device_online     INTEGER NOT NULL DEFAULT 1,
			notify_update_available   INTEGER NOT NULL DEFAULT 1,
			notify_command_failed     INTEGER NOT NULL DEFAULT 1,
			notify_registration_request INTEGER NOT NULL DEFAULT 1,
			created_at               INTEGER NOT NULL,
			updated_at               INTEGER NOT NULL,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)
	`)

	return err
}
