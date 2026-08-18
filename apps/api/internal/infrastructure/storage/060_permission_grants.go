package storage

import (
	"context"
	"database/sql"
)

// migratePermissionGrants creates the table for custom scoped permission
// grants. Role defaults come from the permission engine; this table stores
// extra per-resource scopes an admin/super_admin assigns to a member, which are
// unioned on top of role defaults at evaluation time (Issue 4: scoped RBAC).
func migratePermissionGrants(tx *sql.Tx) error {
	ctx := context.Background()

	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS permission_grants (
		id          TEXT PRIMARY KEY,
		operator_id TEXT NOT NULL,
		org_id      TEXT NOT NULL,
		action      TEXT NOT NULL,
		scope       TEXT NOT NULL,
		created_at  DATETIME NOT NULL
	)`)
	if err != nil {
		return err
	}

	ndx := []string{
		`CREATE INDEX IF NOT EXISTS idx_permission_grants_operator_org ON permission_grants(operator_id, org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_permission_grants_scope ON permission_grants(scope)`,
	}
	for _, q := range ndx {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return err
		}
	}

	// Prevent duplicate (operator, org, action, scope) grants; idempotent insert.
	if _, err := tx.ExecContext(ctx, `
	CREATE UNIQUE INDEX IF NOT EXISTS idx_permission_grants_unique
	ON permission_grants(operator_id, org_id, action, scope)`); err != nil {
		return err
	}
	return nil
}
