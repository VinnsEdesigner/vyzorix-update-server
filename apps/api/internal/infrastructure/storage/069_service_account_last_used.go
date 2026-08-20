package storage

import (
	"context"
	"database/sql"
)

func migrateServiceAccountLastUsed(tx *sql.Tx) error {
	ctx := context.Background()
	_, err := tx.ExecContext(ctx, `ALTER TABLE service_account_tokens ADD COLUMN last_used_at INTEGER`)
	if err != nil && !isDuplicateColumn(err) {
		return err
	}
	return nil
}
