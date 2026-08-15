package storage

import (
	"context"
	"database/sql"
	"strings"
)

// migrateAddUpdatePushOrgColumn adds the organization_id column to update_pushes.
//
// The multi-tenant refactor (migration 039/040) made update pushes org-scoped —
// UpdatesStorage.ListPushes / GetPushByID / ListPushDevices all filter with
// "WHERE organization_id = ?" and CreatePush inserts it — but the original
// create_update_tables migration (023) never included the column, and no later
// migration added it. As a result every push query failed at runtime with
// "no such column: organization_id" (GET /v1/updates/history returned 500).
//
// Idempotent: skips if the column already exists.
func migrateAddUpdatePushOrgColumn(db *sql.DB) error {
	rows, err := db.QueryContext(context.Background(), "PRAGMA table_info(update_pushes)")
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	hasOrg := false
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt sql.NullString
		if scanErr := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); scanErr != nil {
			return scanErr
		}
		if strings.EqualFold(name, "organization_id") {
			hasOrg = true
		}
	}
	if err = rows.Err(); err != nil {
		return err
	}

	if hasOrg {
		return nil
	}

	_, err = db.ExecContext(context.Background(), "ALTER TABLE update_pushes ADD COLUMN organization_id TEXT")
	return err
}
