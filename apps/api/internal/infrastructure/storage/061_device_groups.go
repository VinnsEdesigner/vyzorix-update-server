package storage

import (
	"context"
	"database/sql"
)

// migrateDeviceGroups creates the device-group (team) tables. Devices are
// partitioned into groups within an org; a member of a group can access the
// group's devices (Issue 5: teams / device groups).
func migrateDeviceGroups(tx *sql.Tx) error {
	ctx := context.Background()

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS device_groups (
			id         TEXT PRIMARY KEY,
			org_id     TEXT NOT NULL,
			name       TEXT NOT NULL,
			created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS device_group_members (
			group_id     TEXT NOT NULL,
			operator_id  TEXT NOT NULL,
			created_at   DATETIME NOT NULL,
			PRIMARY KEY (group_id, operator_id)
		)`,
		`CREATE TABLE IF NOT EXISTS device_group_devices (
			group_id   TEXT NOT NULL,
			device_id  TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			PRIMARY KEY (group_id, device_id)
		)`,
	}
	for _, q := range stmts {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return err
		}
	}

	ndx := []string{
		`CREATE INDEX IF NOT EXISTS idx_device_groups_org ON device_groups(org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_dgm_group ON device_group_members(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_dgm_operator ON device_group_members(operator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_dgd_group ON device_group_devices(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_dgd_device ON device_group_devices(device_id)`,
	}
	for _, q := range ndx {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	return nil
}
