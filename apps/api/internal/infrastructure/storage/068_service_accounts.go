package storage

import (
	"context"
	"database/sql"
)

func migrateServiceAccounts(tx *sql.Tx) error {
	ctx := context.Background()
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS service_accounts (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			name TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_service_accounts_org ON service_accounts(org_id)`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS service_account_tokens (
			id TEXT PRIMARY KEY,
			service_id TEXT NOT NULL,
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL UNIQUE,
			key_prefix TEXT NOT NULL,
			scopes TEXT NOT NULL DEFAULT '[]',
			valid INTEGER NOT NULL DEFAULT 1,
			expires_at INTEGER,
			revoked_at INTEGER,
			created_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_service_account_tokens_service ON service_account_tokens(service_id)`)
	return err
}
