package storage

import (
	"context"
	"database/sql"
)

// migrateCreateInboxAndRegistration creates inbox_requests and registration_logs tables.
func migrateCreateInboxAndRegistration(db *sql.DB) error {
	// Create inbox_requests table
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS inbox_requests (
			id                      TEXT PRIMARY KEY,
			device_imei             TEXT NOT NULL,
			firebase_install_id      TEXT NOT NULL,
			fcm_token               TEXT,
			device_name             TEXT,
			manufacturer            TEXT,
			os_version              TEXT,
			app_version             TEXT,
			device_class            TEXT,
			device_model            TEXT,
			status                  TEXT NOT NULL DEFAULT 'pending',
			reviewed_by             TEXT,
			reviewed_at             INTEGER,
			reviewed_reason         TEXT,
			rejection_reason        TEXT,
			command_secret          TEXT,
			created_at              INTEGER NOT NULL,
			updated_at              INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying pending inbox requests
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_inbox_status 
		ON inbox_requests(status, created_at DESC)
	`)
	if err != nil {
		return err
	}

	// Create index for querying by device IMEI
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_inbox_imei 
		ON inbox_requests(device_imei)
	`)
	if err != nil {
		return err
	}

	// Create registration_logs table
	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS registration_logs (
			id                  TEXT PRIMARY KEY,
			inbox_request_id    TEXT,
			device_id           TEXT,
			action              TEXT NOT NULL,
			old_status          TEXT,
			new_status          TEXT,
			performed_by        TEXT,
			reason              TEXT,
			created_at          INTEGER NOT NULL,
			FOREIGN KEY(inbox_request_id) REFERENCES inbox_requests(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying logs by inbox request
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_registration_logs_request 
		ON registration_logs(inbox_request_id, created_at DESC)
	`)
	if err != nil {
		return err
	}

	// Create index for querying logs by device
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_registration_logs_device 
		ON registration_logs(device_id, created_at DESC)
	`)

	return err
}
