
package storage

import (
	"context"
	"database/sql"
)

// migrateCommandConfirmations creates the table that stores short-lived,
// single-use confirmation tokens for risky device commands (Phase 3).
func migrateCommandConfirmations(tx *sql.Tx) error {
	ctx := context.Background()

	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS command_confirmations (
		token       TEXT PRIMARY KEY,
		operator_id TEXT NOT NULL,
		org_id      TEXT,
		command     TEXT NOT NULL,
		device_id   TEXT,
		risk_tier   TEXT,
		created_at  DATETIME NOT NULL,
		expires_at  DATETIME NOT NULL,
		consumed_at DATETIME
	)`)
	if err != nil {
		return err
	}

	// Lookup by operator + command (the consume path filters on these).
	if _, err := tx.ExecContext(ctx, `
	CREATE INDEX IF NOT EXISTS idx_command_confirmations_operator
	ON command_confirmations(operator_id, command)`); err != nil {
		return err
	}

	// Cleanup scans on expires_at.
	if _, err := tx.ExecContext(ctx, `
	CREATE INDEX IF NOT EXISTS idx_command_confirmations_expires_at
	ON command_confirmations(expires_at)`); err != nil {
		return err
	}

	return nil
}