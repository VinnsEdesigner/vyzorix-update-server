package storage

import (
	"context"
	"database/sql"
	"strings"
)

// migrateFixDeviceSettingsFK repairs the malformed foreign key on device_settings.
//
// Migration 042 (migrateDeviceSettings) created device_settings with:
//
//	FOREIGN KEY (device_imei) REFERENCES devices(imei) ON DELETE CASCADE
//
// but the devices table has no `imei` column — the IMEI is stored in devices.id.
// Every other child table (commands, telemetry, device_events, dashboard logs)
// correctly references devices(id). The mismatched reference makes the foreign
// key unresolvable, so SQLite raises "foreign key mismatch" on any DELETE
// against devices once a device_settings row exists, and the intended ON DELETE
// CASCADE never fires. This breaks the device-deletion worker for any device
// that has settings.
//
// SQLite cannot ALTER a foreign key in place, so this migration rebuilds
// device_settings: it captures the existing CREATE TABLE SQL, rewrites the bad
// REFERENCES clause to devices(id), recreates the table under a temporary name,
// copies the data, drops the old table, and renames. Idempotent — if the FK
// already targets devices(id) the migration is a no-op.
func migrateFixDeviceSettingsFK(db *sql.DB) error {
	// Determine whether the FK is already correct by inspecting the foreign key
	// list. If devices(id) is present (and devices(imei) is not), skip.
	fkRows, err := db.QueryContext(context.Background(), "PRAGMA foreign_key_list(device_settings)")
	if err != nil {
		// Table may not exist yet on a partially-migrated DB; nothing to fix.
		return nil //nolint:nilerr // intentional: missing table is not an error here
	}
	defer func() { _ = fkRows.Close() }()

	needsFix := false
	for fkRows.Next() {
		var id, seq int
		var table, from, to, onUpdate, onDelete, match string
		if scanErr := fkRows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match); scanErr != nil {
			return scanErr
		}
		if strings.EqualFold(table, "devices") && (to == "" || strings.EqualFold(to, "imei")) {
			needsFix = true
		}
	}
	if err = fkRows.Err(); err != nil {
		return err
	}

	if !needsFix {
		return nil
	}

	// Rebuild device_settings with the corrected FK. The column set is fixed by
	// migration 042 and unchanged since, so a literal recreation is safe here.
	createStmt := `
		CREATE TABLE _device_settings_new (
			id TEXT PRIMARY KEY,
			device_imei TEXT NOT NULL UNIQUE,
			custom_name TEXT,
			location TEXT,
			metadata JSONB,
			thresholds JSONB,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			FOREIGN KEY (device_imei) REFERENCES devices(id) ON DELETE CASCADE
		)`
	if _, err := db.ExecContext(context.Background(), createStmt); err != nil {
		return err
	}

	copyStmt := `INSERT INTO _device_settings_new (id, device_imei, custom_name, location, metadata, thresholds, created_at, updated_at)
		SELECT id, device_imei, custom_name, location, metadata, thresholds, created_at, updated_at FROM device_settings`
	if _, err := db.ExecContext(context.Background(), copyStmt); err != nil {
		_, _ = db.ExecContext(context.Background(), "DROP TABLE _device_settings_new")
		return err
	}

	if _, err := db.ExecContext(context.Background(), "DROP TABLE device_settings"); err != nil {
		_, _ = db.ExecContext(context.Background(), "DROP TABLE _device_settings_new")
		return err
	}
	if _, err := db.ExecContext(context.Background(), "ALTER TABLE _device_settings_new RENAME TO device_settings"); err != nil {
		return err
	}

	// Recreate the lookup index (idempotent).
	if _, err := db.ExecContext(context.Background(),
		"CREATE INDEX IF NOT EXISTS idx_device_settings_imei ON device_settings(device_imei)"); err != nil {
		return err
	}

	return nil
}
