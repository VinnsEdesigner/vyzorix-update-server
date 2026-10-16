package storage

import (
	"context"
	"database/sql"
)

// migrateKnownDevices creates the tables backing the persistent device store used
// for new-device login notifications.
//
// known_devices holds one row per (operator, fingerprint). A fingerprint is
// derived from request signals (IP, User-Agent, Accept-Language, Accept-Encoding)
// so it is client-influenced; the per-operator row count is capped in the store.
//
// known_device_signals holds the distinct IPs / User-Agents observed for a device.
// Keeping them in their own table (rather than a JSON blob) lets the store evict
// individual signals and query "has this device been seen from this IP?" without
// rewriting a whole column.
func migrateKnownDevices(tx *sql.Tx) error {
	ctx := context.Background()

	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS known_devices (
			operator_id   TEXT NOT NULL,
			fingerprint   TEXT NOT NULL,
			first_seen_at INTEGER NOT NULL,
			last_seen_at  INTEGER NOT NULL,
			is_trusted    INTEGER NOT NULL DEFAULT 1,
			PRIMARY KEY (operator_id, fingerprint)
		)
	`); err != nil {
		return err
	}

	// Supports eviction of the least-recently-seen device per operator.
	if _, err := tx.ExecContext(ctx,
		`CREATE INDEX IF NOT EXISTS idx_known_devices_operator_last_seen
			ON known_devices(operator_id, last_seen_at)`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS known_device_signals (
			operator_id TEXT NOT NULL,
			fingerprint TEXT NOT NULL,
			kind        TEXT NOT NULL,
			value       TEXT NOT NULL,
			seen_at     INTEGER NOT NULL,
			PRIMARY KEY (operator_id, fingerprint, kind, value)
		)
	`); err != nil {
		return err
	}

	// Supports eviction of the oldest signal per (device, kind).
	if _, err := tx.ExecContext(ctx,
		`CREATE INDEX IF NOT EXISTS idx_known_device_signals_operator_kind
			ON known_device_signals(operator_id, fingerprint, kind, seen_at)`); err != nil {
		return err
	}

	// Cascading cleanup: dropping a device drops its signals. SQLite needs
	// foreign_keys enabled for this to fire, so the store also deletes explicitly.
	return nil
}
