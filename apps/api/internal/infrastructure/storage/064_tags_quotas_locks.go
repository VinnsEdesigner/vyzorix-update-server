package storage

import (
	"context"
	"database/sql"
)

func migrateAddTagsAndQuotas(tx *sql.Tx) error {
	ctx := context.Background()
	_, err := tx.ExecContext(ctx, `ALTER TABLE devices ADD COLUMN tags TEXT DEFAULT '[]'`)
	if err != nil && !isDuplicateColumn(err) {
		return err
	}
	_, err = tx.ExecContext(ctx, `ALTER TABLE organizations ADD COLUMN quotas TEXT DEFAULT '{}'`)
	if err != nil && !isDuplicateColumn(err) {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS server_locks (name TEXT PRIMARY KEY, holder TEXT NOT NULL, acquired_at INTEGER NOT NULL, expires_at INTEGER NOT NULL)`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `ALTER TABLE operator_settings ADD COLUMN preferences TEXT DEFAULT '{}'`)
	if err != nil && !isDuplicateColumn(err) {
		return err
	}
	return nil
}

func isDuplicateColumn(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate column") || contains(err.Error(), "already exists"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
