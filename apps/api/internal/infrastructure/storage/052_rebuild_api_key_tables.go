package storage

import (
	"context"
	"database/sql"
	"strings"
)

// migrateRebuildAPIKeyTables rebuilds the api_keys and api_clients tables to the.
// schema the current repository code actually expects.
//
// Background: an earlier multi-tenant API key implementation (commit b3fbf3f,
// file 038_api_keys.go, since deleted) created api_keys with columns.
// rate_limit_per_minute / last_used_at, and api_clients was created by migration.
// 013 with a minimal schema (id, name, api_key_hash, hmac_key_hash, ...). The.
// current code (039_api_key_storage.go APIKeyRepositoryImpl and.
// client_storage.go ClientRepository) queries a richer schema.
// (operator_id, key_prefix, scope, is_active, request_count, last_request_at,
// revoked_at for api_keys; operator_id, platform, client_secret_hash, hmac_key,
// allowed_origins, allowed_paths, rate_limit, is_active, request_count,
// last_request_at for api_clients). SetupAPIKeysTable is never wired into the.
// migration list, and CREATE TABLE IF NOT EXISTS cannot add columns to an.
// existing table, so the rich schema was never applied. The result: every.
// /v1/auth/api-keys, /v1/admin/api-keys, /v1/auth/client-credentials, and.
// /v1/admin/clients endpoint failed at runtime with "no such column".
//
// Both tables are empty in production (the feature is new), so we drop and.
// recreate them with the intended schema. Idempotent: if a table already has.
// the expected columns, it is left untouched.
func migrateRebuildAPIKeyTables(tx *sql.Tx) error {
	if err := rebuildAPIKeysTable(tx); err != nil {
		return err
	}
	if err := rebuildAPIClientsTable(tx); err != nil {
		return err
	}
	return nil
}

func rebuildAPIKeysTable(tx *sql.Tx) error {
	// Expected rich schema columns for api_keys.
	wantAPIKeys := []string{
		"id", "operator_id", "name", "key_prefix", "key_hash", "scope",
		"expires_at", "is_active", "request_count", "last_request_at",
		"created_at", "updated_at", "revoked_at", "organization_id",
	}
	has, missing, err := tableHasColumns(tx, "api_keys", wantAPIKeys)
	if err != nil {
		return err
	}
	if has {
		return nil // already the rich schema (or a superset).
	}

	// If the table exists but is missing expected columns, rebuild it.
	exists, err := tableExists(tx, "api_keys")
	if err != nil {
		return err
	}
	if exists {
		// Only rebuild when there is no data to preserve; the feature is new.
		count, errCount := rowCount(tx, "api_keys")
		if errCount != nil {
			return errCount
		}
		if count > 0 {
			// Data present: add only the missing columns defensively rather than drop.
			return addMissingColumns(tx, "api_keys", missing, map[string]string{
				"operator_id":     "TEXT",
				"key_prefix":      "TEXT",
				"scope":           "TEXT DEFAULT 'read'",
				"is_active":       "INTEGER NOT NULL DEFAULT 1",
				"request_count":   "INTEGER NOT NULL DEFAULT 0",
				"last_request_at": "INTEGER",
				"revoked_at":      "INTEGER",
			})
		}
		if _, err = tx.ExecContext(context.Background(), `DROP TABLE IF EXISTS api_keys`); err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			name TEXT NOT NULL,
			key_prefix TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			signing_secret TEXT NOT NULL DEFAULT '',
			scope TEXT NOT NULL DEFAULT 'read',
			expires_at INTEGER,
			is_active INTEGER NOT NULL DEFAULT 1,
			request_count INTEGER NOT NULL DEFAULT 0,
			last_request_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			revoked_at INTEGER,
			organization_id TEXT,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS idx_api_keys_operator_id ON api_keys(operator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_created_at ON api_keys(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_org ON api_keys(organization_id)`,
	} {
		if _, err := tx.ExecContext(context.Background(), idx); err != nil {
			return err
		}
	}
	return nil
}

func rebuildAPIClientsTable(tx *sql.Tx) error {
	wantAPIClients := []string{
		"id", "operator_id", "name", "platform", "client_secret_hash", "hmac_key",
		"allowed_origins", "allowed_paths", "rate_limit", "is_active",
		"request_count", "last_request_at", "created_at", "updated_at",
	}
	has, missing, err := tableHasColumns(tx, "api_clients", wantAPIClients)
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	exists, err := tableExists(tx, "api_clients")
	if err != nil {
		return err
	}
	if exists {
		count, errCount := rowCount(tx, "api_clients")
		if errCount != nil {
			return errCount
		}
		if count > 0 {
			return addMissingColumns(tx, "api_clients", missing, map[string]string{
				"operator_id":        "TEXT",
				"platform":           "TEXT",
				"client_secret_hash": "TEXT",
				"hmac_key":           "TEXT",
				"allowed_origins":    "TEXT",
				"allowed_paths":      "TEXT",
				"rate_limit":         "INTEGER NOT NULL DEFAULT 0",
				"is_active":          "INTEGER NOT NULL DEFAULT 1",
				"request_count":      "INTEGER NOT NULL DEFAULT 0",
				"last_request_at":    "INTEGER",
			})
		}
		if _, err = tx.ExecContext(context.Background(), `DROP TABLE IF EXISTS api_clients`); err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS api_clients (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			name TEXT NOT NULL,
			platform TEXT NOT NULL DEFAULT 'web',
			client_secret_hash TEXT NOT NULL,
			hmac_key TEXT,
			allowed_origins TEXT,
			allowed_paths TEXT,
			rate_limit INTEGER NOT NULL DEFAULT 0,
			is_active INTEGER NOT NULL DEFAULT 1,
			request_count INTEGER NOT NULL DEFAULT 0,
			last_request_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS idx_api_clients_operator_id ON api_clients(operator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_clients_is_active ON api_clients(is_active)`,
	} {
		if _, err := tx.ExecContext(context.Background(), idx); err != nil {
			return err
		}
	}
	return nil
}

// tableHasColumns reports whether the given table has all want columns.
// Returns (true, nil, nil) when all present; (false, missing, nil) otherwise.
func tableHasColumns(tx *sql.Tx, table string, want []string) (bool, []string, error) {
	existing, err := columnSet(tx, table)
	if err != nil {
		return false, nil, err
	}
	var missing []string
	for _, c := range want {
		if !existing[strings.ToLower(c)] {
			missing = append(missing, c)
		}
	}
	return len(missing) == 0, missing, nil
}

func columnSet(tx *sql.Tx, table string) (map[string]bool, error) {
	set := make(map[string]bool)
	rows, err := tx.QueryContext(context.Background(), "PRAGMA table_info("+table+")")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		set[strings.ToLower(name)] = true
	}
	return set, rows.Err()
}

func tableExists(tx *sql.Tx, table string) (bool, error) {
	var name string
	err := tx.QueryRowContext(context.Background(),
		`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func rowCount(tx *sql.Tx, table string) (int, error) {
	var n int
	err := tx.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM "+table).Scan(&n)
	return n, err
}

func addMissingColumns(tx *sql.Tx, table string, missing []string, types map[string]string) error {
	for _, col := range missing {
		decl := types[col]
		if decl == "" {
			decl = "TEXT"
		}
		if _, err := tx.ExecContext(context.Background(),
			"ALTER TABLE "+table+" ADD COLUMN "+col+" "+decl); err != nil {
			// Column may already exist concurrently; ignore that specific error.
			if !isColumnExistsError(err) {
				return err
			}
		}
	}
	return nil
}
