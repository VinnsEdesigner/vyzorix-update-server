package storage

import (
	"context"
	"database/sql"
)

// migrateResourcePermissions creates the canonical ResourcePermission table and
// retires the v60 permission_grants table (pre-production; no data to migrate).
// A grant is one row per (subject, scope) with a comma-separated actions column,
// grantable to operators, teams, or built-in roles, with managed/inherited flags.
func migrateResourcePermissions(tx *sql.Tx) error {
	ctx := context.Background()

	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS resource_permissions (
		id            TEXT PRIMARY KEY,
		org_id        TEXT NOT NULL,
		subject_type  TEXT NOT NULL,
		subject_id    TEXT NOT NULL,
		actions       TEXT NOT NULL,
		scope         TEXT NOT NULL,
		is_managed    INTEGER NOT NULL DEFAULT 0,
		is_inherited  INTEGER NOT NULL DEFAULT 0,
		created_at    INTEGER NOT NULL,
		updated_at    INTEGER NOT NULL
	)`)
	if err != nil {
		return err
	}

	ndx := []string{
		`CREATE INDEX IF NOT EXISTS idx_rp_org ON resource_permissions(org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rp_subject ON resource_permissions(subject_type, subject_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_rp_unique ON resource_permissions(org_id, subject_type, subject_id, scope)`,
	}
	for _, q := range ndx {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return err
		}
	}

	// Drop the superseded v60 table (pre-production; no data to preserve).
	_, _ = tx.ExecContext(ctx, `DROP TABLE IF EXISTS permission_grants`)
	return nil
}
