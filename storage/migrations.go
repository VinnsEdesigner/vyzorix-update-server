package storage

import "strings"

func baseMigrationSQL() string {
	return strings.Join([]string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA cache_size=-2000`,
		`PRAGMA busy_timeout=5000`,
		`PRAGMA foreign_keys=ON`,
		`CREATE TABLE IF NOT EXISTS devices (id TEXT PRIMARY KEY, firebase_install_id TEXT NOT NULL, fcm_token TEXT, app_version TEXT, device_class TEXT, command_secret TEXT NOT NULL, online INTEGER NOT NULL DEFAULT 0, registered_at INTEGER NOT NULL, last_seen INTEGER NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS telemetry (id INTEGER PRIMARY KEY AUTOINCREMENT, device_id TEXT NOT NULL, received_at INTEGER NOT NULL, payload TEXT NOT NULL, risk_score INTEGER, buffer_level INTEGER, thermal_temp REAL, FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE)`,
		`CREATE TABLE IF NOT EXISTS commands (dispatch_id TEXT PRIMARY KEY, device_id TEXT NOT NULL, command TEXT NOT NULL, args TEXT, delivery TEXT NOT NULL, created_at INTEGER NOT NULL, delivered_at INTEGER, wake_sent INTEGER NOT NULL DEFAULT 0, wake_error TEXT, FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE)`,
		`CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry(device_id, received_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_commands_device_time ON commands(device_id, created_at DESC)`,
	}, "; ") + ";"
}

func additiveMigrations() []string {
	return []string{
		`ALTER TABLE commands ADD COLUMN wake_sent INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE commands ADD COLUMN wake_error TEXT`,
	}
}
