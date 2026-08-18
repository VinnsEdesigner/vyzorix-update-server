package storage

import (
	"context"
	"database/sql"
)

// migrateFixRefreshTokensSchema repairs the refresh_tokens table so its column
// names match what RefreshTokenRepository expects.
//
// Migration 025 originally created the revoked flag as `revoked` and omitted the
// `revoked_at` timestamp entirely, but the repository code (session_storage.go)
// inserts/updates using `is_revoked` and `revoked_at`. This caused every logout
// (and any refresh-token revocation) to 500 with "no such column: is_revoked".
//
// Because SQLite cannot rename columns before 3.25 (and Turso may not support
// ALTER TABLE RENAME COLUMN reliably across replicas), we rebuild the table:
// create a new table with the correct schema, copy data, and swap. Idempotent.
func migrateFixRefreshTokensSchema(tx *sql.Tx) error {
	_, _ = tx.ExecContext(context.Background(), "DROP TABLE IF EXISTS refresh_tokens_new")

	ctx := context.Background()

	// If the table doesn't exist at all, migration 025 (now fixed) will have
	// created it with the correct schema. Nothing to do here.
	var exists int
	row := tx.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='refresh_tokens'")
	if err := row.Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return nil
	}

	// Check whether the old `revoked` column is present (i.e. the schema hasn't
	// been fixed yet). If it's already correct, this migration is a no-op.
	cols, err := tx.QueryContext(ctx, "PRAGMA table_info(refresh_tokens)")
	if err != nil {
		return err
	}
	defer cols.Close()

	hasRevoked := false
	hasIsRevoked := false
	hasRevokedAt := false
	for cols.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue *string
		if err := cols.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		switch name {
		case "revoked":
			hasRevoked = true
		case "is_revoked":
			hasIsRevoked = true
		case "revoked_at":
			hasRevokedAt = true
		}
	}
	cols.Close()

	// Already migrated — nothing to do.
	if !hasRevoked && hasIsRevoked && hasRevokedAt {
		return nil
	}

	// Build a fresh table with the correct schema and copy data over.
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS refresh_tokens_new (
			id              TEXT PRIMARY KEY,
			token_hash      TEXT NOT NULL UNIQUE,
			operator_id     TEXT NOT NULL,
			session_id      TEXT NOT NULL,
			expires_at      INTEGER NOT NULL,
			created_at      INTEGER NOT NULL,
			replaced_by_id  TEXT,
			is_revoked      INTEGER NOT NULL DEFAULT 0,
			revoked_at      INTEGER,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE,
			FOREIGN KEY(session_id) REFERENCES auth_sessions(id) ON DELETE CASCADE,
			FOREIGN KEY(replaced_by_id) REFERENCES refresh_tokens_new(id)
		)
	`)
	if err != nil {
		return err
	}

	// Copy rows, mapping the old `revoked` column to `is_revoked` if present.
	if hasRevoked {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO refresh_tokens_new (id, token_hash, operator_id, session_id, expires_at, created_at, replaced_by_id, is_revoked, revoked_at)
			SELECT id, token_hash, operator_id, session_id, expires_at, created_at, replaced_by_id, revoked, NULL
			FROM refresh_tokens
		`)
	} else {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO refresh_tokens_new (id, token_hash, operator_id, session_id, expires_at, created_at, replaced_by_id, is_revoked, revoked_at)
			SELECT id, token_hash, operator_id, session_id, expires_at, created_at, replaced_by_id, 0, NULL
			FROM refresh_tokens
		`)
	}
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DROP TABLE refresh_tokens`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `ALTER TABLE refresh_tokens_new RENAME TO refresh_tokens`)
	if err != nil {
		return err
	}

	// Recreate indexes (dropped with the old table).
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash)`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_refresh_tokens_operator_id ON refresh_tokens(operator_id)`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session_id ON refresh_tokens(session_id)`)
	if err != nil {
		return err
	}

	return nil
}
