package storage

import (
	"context"
	"database/sql"
)

// migrateAddDevicesColumns adds new columns to the devices table.
func migrateAddDevicesColumns(db *sql.DB) error {
	cols := []string{
		`ALTER TABLE devices ADD COLUMN device_name TEXT`,
		`ALTER TABLE devices ADD COLUMN os_version TEXT`,
		`ALTER TABLE devices ADD COLUMN security_patch TEXT`,
		`ALTER TABLE devices ADD COLUMN build_id TEXT`,
		`ALTER TABLE devices ADD COLUMN deregistered_at INTEGER`,
		`ALTER TABLE devices ADD COLUMN deletion_scheduled_at INTEGER`,
		`ALTER TABLE devices ADD COLUMN fcm_token_refreshed_at INTEGER`,
	}

	for _, col := range cols {
		db.ExecContext(context.Background(), col) //nolint:errcheck
	}

	return nil
}
