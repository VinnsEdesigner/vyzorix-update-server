package storage

import (
	"context"
	"database/sql"
	"strings"
)

// migrateAddDeviceModelColumn adds the model column to the devices table.
//
// The domain Device entity, deviceColumns, and the INSERT/UPDATE statements in
// device_storage.go all reference a `model` column, but no prior migration
// (v1–v47) ever created it. On any freshly-migrated database the device SELECT
// failed with "no such column: model". This migration reconciles the live
// schema with the code for both the local SQLite and Turso libSQL backends.
//
// The migration is idempotent: it inspects PRAGMA table_info(devices) and only
// issues the ALTER TABLE when the column is absent. This keeps it safe to
// re-run and avoids "duplicate column name" errors on databases that already
// carry the column (e.g. legacy local DBs created before the migration ledger
// existed).
func migrateAddDeviceModelColumn(db *sql.DB) error {
	rows, err := db.QueryContext(context.Background(), "PRAGMA table_info(devices)")
	if err != nil {
		return err
	}

	hasModel := false
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			_ = rows.Close()
			return err
		}
		if strings.EqualFold(name, "model") {
			hasModel = true
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()

	if hasModel {
		return nil
	}

	_, err = db.ExecContext(context.Background(), "ALTER TABLE devices ADD COLUMN model TEXT")
	return err
}
