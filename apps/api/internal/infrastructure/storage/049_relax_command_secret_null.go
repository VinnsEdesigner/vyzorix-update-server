package storage

import (
	"context"
	"database/sql"
	"strings"
)

// migrateRelaxDeviceCommandSecretNull relaxes the NOT NULL constraint on
// devices.command_secret.
//
// The devices table was originally created (migrateCreateDevices) with
// `command_secret TEXT NOT NULL`, carrying the plaintext command secret. The
// codebase has since migrated to storing only the hashed secret in
// `command_secret_hash` (migration v44); DeviceRepository.Create never writes
// `command_secret`. As a result every device creation failed with
// "NOT NULL constraint failed: devices.command_secret" on any backend that
// enforces the constraint.
//
// SQLite cannot ALTER a column's constraint in place, so this migration rebuilds
// the table: it creates a shadow copy with `command_secret TEXT` (nullable),
// copies all rows, drops the original, and renames the copy back. It preserves
// every other column and is idempotent — it detects whether the constraint is
// already relaxed (column nullable) and skips the rebuild.
//
// The rebuild copies data with an explicit column list derived from
// PRAGMA table_info so it is resilient to the exact column set present on the
// live database (which may carry columns added by later migrations).
func migrateRelaxDeviceCommandSecretNull(db *sql.DB) error {
	// Inspect the current devices schema.
	rows, err := db.QueryContext(context.Background(), "PRAGMA table_info(devices)")
	if err != nil {
		return err
	}

	type colInfo struct {
		name    string
		typ     string
		notnull int
		dflt    sql.NullString
		pk      int
	}
	var cols []colInfo
	commandSecretNotNull := false
	for rows.Next() {
		var ci colInfo
		var cid int
		if err := rows.Scan(&cid, &ci.name, &ci.typ, &ci.notnull, &ci.dflt, &ci.pk); err != nil {
			_ = rows.Close()
			return err
		}
		cols = append(cols, ci)
		if strings.EqualFold(ci.name, "command_secret") && ci.notnull == 1 {
			commandSecretNotNull = true
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()

	// Nothing to do if the column is already nullable (or absent).
	if !commandSecretNotNull {
		return nil
	}

	// Build the shadow table definition, mirroring every column but relaxing
	// command_secret to nullable. Preserve types, defaults, and PK.
	var def strings.Builder
	def.WriteString("CREATE TABLE _devices_new (")
	for i, ci := range cols {
		if i > 0 {
			def.WriteString(", ")
		}
		def.WriteString(quoteIdent(ci.name))
		def.WriteString(" ")
		def.WriteString(ci.typ)
		if strings.EqualFold(ci.name, "command_secret") {
			// Drop the NOT NULL constraint for this column only.
		} else if ci.notnull == 1 {
			def.WriteString(" NOT NULL")
		}
		if ci.dflt.Valid {
			def.WriteString(" DEFAULT ")
			def.WriteString(ci.dflt.String)
		}
		if ci.pk == 1 {
			def.WriteString(" PRIMARY KEY")
		}
	}
	def.WriteString(")")

	if _, err := db.ExecContext(context.Background(), def.String()); err != nil {
		return err
	}

	// Copy all data across, using the live column set.
	names := make([]string, 0, len(cols))
	for _, ci := range cols {
		names = append(names, quoteIdent(ci.name))
	}
	colList := strings.Join(names, ", ")
	placeholders := strings.Repeat("?, ", len(names))
	placeholders = strings.TrimSuffix(placeholders, ", ")

	copyQuery := "INSERT INTO _devices_new (" + colList + ") SELECT " + colList + " FROM devices"
	if _, err := db.ExecContext(context.Background(), copyQuery); err != nil {
		_, _ = db.ExecContext(context.Background(), "DROP TABLE _devices_new")
		return err
	}

	// Swap tables. Re-create indexes that referenced the old table afterward is
	// the caller's responsibility via the standard migration ledger; the devices
	// indexes are created idempotently by earlier migrations with IF NOT EXISTS,
	// so we drop and let them remain absent only if they were table-bound. To be
	// safe, drop the old table and rename.
	if _, err := db.ExecContext(context.Background(), "DROP TABLE devices"); err != nil {
		_, _ = db.ExecContext(context.Background(), "DROP TABLE _devices_new")
		return err
	}
	if _, err := db.ExecContext(context.Background(), "ALTER TABLE _devices_new RENAME TO devices"); err != nil {
		return err
	}

	// Re-create the standard devices indexes (idempotent). These mirror the
	// index definitions from earlier migrations so query plans remain optimal.
	indexStmts := []string{
		"CREATE INDEX IF NOT EXISTS idx_devices_operator_id ON devices(operator_id)",
		"CREATE INDEX IF NOT EXISTS idx_devices_organization_id ON devices(organization_id)",
		"CREATE INDEX IF NOT EXISTS idx_devices_firebase_install_id ON devices(firebase_install_id)",
	}
	for _, s := range indexStmts {
		if _, err := db.ExecContext(context.Background(), s); err != nil {
			// Non-fatal: indexes are not required for correctness.
			continue
		}
	}

	return nil
}

// quoteIdent wraps an SQLite identifier in double quotes for safe use in DDL.
func quoteIdent(name string) string {
	return "\"" + strings.ReplaceAll(name, "\"", "\"\"") + "\""
}
