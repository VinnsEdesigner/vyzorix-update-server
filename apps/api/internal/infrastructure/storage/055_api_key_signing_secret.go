package storage

import "database/sql"

// migrateAddAPIKeySigningSecret adds a signing_secret column to api_keys.
// API keys need their own signing secret so API-key-authenticated requests.
// can be HMAC-signed (Domain A: client↔server request signing) without a.
// session. The secret is a deterministic SHA-512 derivation of the full API.
// key value, generated at creation time and stored alongside the key hash.
// Existing keys get an empty signing_secret (signing is opt-in per key until.
// rotated). Idempotent.
func migrateAddAPIKeySigningSecret(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE api_keys ADD COLUMN signing_secret TEXT NOT NULL DEFAULT ''
	`)
	if err != nil {
		// Column may already exist (idempotent re-run).
		if err.Error() != "duplicate column name: signing_secret" {
			return err
		}
	}
	return nil
}
