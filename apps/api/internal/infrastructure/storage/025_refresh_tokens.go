package storage

import (
	"context"
	"database/sql"
)

// migrateCreateRefreshTokens creates the refresh_tokens table.
func migrateCreateRefreshTokens(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS refresh_tokens (
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
			FOREIGN KEY(replaced_by_id) REFERENCES refresh_tokens(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create index for token hash lookups.
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash 
		ON refresh_tokens(token_hash)
	`)
	if err != nil {
		return err
	}

	// Create index for querying tokens by operator.
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_operator_id 
		ON refresh_tokens(operator_id)
	`)
	if err != nil {
		return err
	}

	// Create index for querying tokens by session.
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session_id 
		ON refresh_tokens(session_id)
	`)

	return err
}
