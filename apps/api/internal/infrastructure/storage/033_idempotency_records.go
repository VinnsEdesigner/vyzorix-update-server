package storage

import (
	"context"
	"database/sql"
)

// This table stores idempotency keys and their associated responses for request deduplication.
func migrateIdempotencyRecords(db *sql.DB) error {
	ctx := context.Background()

	// Create idempotency_records table.
	query := `
	CREATE TABLE IF NOT EXISTS idempotency_records (
		id TEXT PRIMARY KEY,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		hash TEXT NOT NULL,
		status_code INTEGER NOT NULL,
		response_body BLOB NOT NULL,
		created_at DATETIME NOT NULL,
		expires_at DATETIME NOT NULL,
		content_type TEXT NOT NULL,
		client_ip TEXT,
		user_agent TEXT
	)`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// Create index on expires_at for cleanup queries.
	indexQuery := `
	CREATE INDEX IF NOT EXISTS idx_idempotency_expires_at ON idempotency_records(expires_at)`
	_, err = db.ExecContext(ctx, indexQuery)
	if err != nil {
		return err
	}

	// Create composite index on method+path for lookup optimization.
	compositeQuery := `
	CREATE INDEX IF NOT EXISTS idx_idempotency_method_path ON idempotency_records(method, path)`
	_, err = db.ExecContext(ctx, compositeQuery)
	if err != nil {
		return err
	}

	return nil
}
