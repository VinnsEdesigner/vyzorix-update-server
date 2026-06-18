// Package storage provides SQLite database operations.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Migration represents a single database migration.
type Migration struct {
	Version int
	Name    string
	Apply   func(*sql.DB) error
}

// migrations is the registry of all database migrations.
// Migrations must be applied in order (by Version).
var migrations = []Migration{
	{1, "create_devices_table", migrateCreateDevices},
	{2, "create_telemetry_table", migrateCreateTelemetry},
	{3, "create_commands_table", migrateCreateCommands},
	{4, "create_operators_table", migrateCreateOperators},
	{5, "create_auth_sessions_table", migrateCreateAuthSessions},
	{6, "create_email_verifications_table", migrateCreateEmailVerifications},
	{7, "create_password_reset_table", migrateCreatePasswordReset},
	{8, "create_settings_table", migrateCreateSettings},
	{9, "add_commands_columns", migrateAddCommandsColumns},
	{10, "add_device_secret_hash", migrateAddDeviceSecretHash},
	{11, "add_operators_github_id", migrateAddOperatorsGitHubID},
	{12, "create_resend_tracker_table", migrateCreateResendTracker},
	{13, "create_api_clients_table", migrateCreateAPIClients},
	{14, "create_signing_keys_table", migrateCreateSigningKeys},
	{15, "create_session_revocation_list", migrateCreateSessionRevocationList},
	{16, "create_failed_login_attempts", migrateCreateFailedLoginAttempts},
	{17, "create_account_lockouts", migrateCreateAccountLockouts},
	{18, "create_audit_logs", migrateCreateAuditLogs},
}

// RunMigrations applies all pending migrations to the database.
func RunMigrations(db *sql.DB) error {
	// Ensure migrations table exists
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Apply pending migrations
	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue // Already applied
		}

		if err := m.Apply(db); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", m.Version, m.Name, err)
		}

		if err := setVersion(db, m.Version); err != nil {
			return fmt.Errorf("failed to set version after migration %d: %w", m.Version, err)
		}

		fmt.Printf("Applied migration %d: %s\n", m.Version, m.Name)
	}

	return nil
}

const migrationsTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	applied_at INTEGER NOT NULL
)`

func createMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(migrationsTable)
	return err
}

func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow(`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return version, err
}

func setVersion(db *sql.DB, version int) error {
	_, err := db.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, sqlDateTime())
	return err
}

func sqlDateTime() int64 {
	return time.Now().UTC().UnixMilli()
}

// =============================================================================
// Migration Functions
// =============================================================================

func migrateCreateDevices(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS devices (
			id TEXT PRIMARY KEY,
			firebase_install_id TEXT NOT NULL,
			fcm_token TEXT,
			app_version TEXT,
			device_class TEXT,
			command_secret TEXT NOT NULL,
			online INTEGER NOT NULL DEFAULT 0,
			registered_at INTEGER NOT NULL,
			last_seen INTEGER NOT NULL
		)
	`)
	return err
}

func migrateCreateTelemetry(db *sql.DB) error {
	// Use TEXT PRIMARY KEY for UUIDv7 support
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS telemetry (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL,
			received_at INTEGER NOT NULL,
			payload TEXT NOT NULL,
			risk_score INTEGER,
			buffer_level INTEGER,
			thermal_temp REAL,
			FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create index for efficient queries
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry(device_id, received_at DESC)
	`)
	return err
}

func migrateCreateCommands(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS commands (
			dispatch_id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL,
			command TEXT NOT NULL,
			args TEXT,
			delivery TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			delivered_at INTEGER,
			status TEXT NOT NULL DEFAULT 'pending',
			FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_commands_device_time ON commands(device_id, created_at DESC)
	`)
	return err
}

func migrateCreateOperators(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS operators (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			password_hash TEXT,
			role TEXT NOT NULL DEFAULT 'operator',
			google_id TEXT UNIQUE,
			github_id TEXT UNIQUE,
			email_verified INTEGER NOT NULL DEFAULT 0,
			verification_sent_at INTEGER,
			risk_warn INTEGER NOT NULL DEFAULT 50,
			risk_crit INTEGER NOT NULL DEFAULT 75,
			thermal_warn INTEGER NOT NULL DEFAULT 45,
			thermal_crit INTEGER NOT NULL DEFAULT 55,
			buffer_warn INTEGER NOT NULL DEFAULT 50,
			buffer_crit INTEGER NOT NULL DEFAULT 80,
			strict_hmac INTEGER NOT NULL DEFAULT 0,
			auto_reconnect INTEGER NOT NULL DEFAULT 1,
			notifications_enabled INTEGER NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	return err
}

func migrateCreateAuthSessions(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS auth_sessions (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			user_agent TEXT,
			ip_address TEXT,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_auth_sessions_operator ON auth_sessions(operator_id)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires ON auth_sessions(expires_at)
	`)
	return err
}

func migrateCreateEmailVerifications(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS email_verifications (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_email_verifications_operator ON email_verifications(operator_id)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token_hash)
	`)
	return err
}

func migrateCreatePasswordReset(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at INTEGER NOT NULL,
			used_at INTEGER,
			created_at INTEGER NOT NULL,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_password_reset_operator ON password_reset_tokens(operator_id)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_password_reset_token ON password_reset_tokens(token_hash)
	`)
	return err
}

func migrateCreateSettings(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Set default HMAC window (30 seconds per COMMAND_SECURITY.md)
	_, err = db.ExecContext(context.Background(), `
		INSERT OR IGNORE INTO settings (key, value, updated_at) VALUES ('hmac_window_seconds', '30', ?)
	`, sqlDateTime())
	return err
}

func migrateAddCommandsColumns(db *sql.DB) error {
	// These are additive migrations - best-effort
	queries := []string{
		`ALTER TABLE commands ADD COLUMN wake_sent INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE commands ADD COLUMN wake_error TEXT`,
		`ALTER TABLE commands ADD COLUMN completed_at INTEGER`,
		`ALTER TABLE commands ADD COLUMN result TEXT`,
	}

	for _, q := range queries {
		db.ExecContext(context.Background(), q) //nolint:errcheck
	}
	return nil
}

func migrateAddDeviceSecretHash(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		ALTER TABLE devices ADD COLUMN command_secret_hash TEXT
	`)
	return err
}

func migrateAddOperatorsGitHubID(db *sql.DB) error {
	// Use PRAGMA table_info to check if column already exists
	var count int
	err := db.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM pragma_table_info('operators') WHERE name = 'github_id'
	`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Column already exists, skip
	}

	_, err = db.ExecContext(context.Background(), `
		ALTER TABLE operators ADD COLUMN github_id TEXT
	`)
	return err
}

func migrateCreateResendTracker(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS password_reset_resend_tracker (
			id TEXT PRIMARY KEY,
			email_hash TEXT NOT NULL UNIQUE,
			resend_count INTEGER NOT NULL DEFAULT 0,
			last_resend_at INTEGER NOT NULL,
			lockout_until INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	return err
}

func migrateCreateAPIClients(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS api_clients (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			name TEXT NOT NULL,
			platform TEXT NOT NULL,
			client_secret_hash TEXT NOT NULL,
			allowed_origins TEXT,
			allowed_paths TEXT,
			rate_limit INTEGER NOT NULL DEFAULT 100,
			is_active INTEGER NOT NULL DEFAULT 1,
			request_count INTEGER NOT NULL DEFAULT 0,
			last_request_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_api_clients_operator ON api_clients(operator_id)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_api_clients_active ON api_clients(is_active)
	`)
	return err
}

func migrateCreateSigningKeys(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS signing_keys (
			id TEXT PRIMARY KEY,
			client_id TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			version INTEGER NOT NULL,
			issued_at INTEGER NOT NULL,
			expires_at INTEGER,
			is_active INTEGER NOT NULL DEFAULT 1,
			revoked_at INTEGER,
			FOREIGN KEY (client_id) REFERENCES api_clients(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_signing_keys_client ON signing_keys(client_id)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_signing_keys_active ON signing_keys(client_id, is_active)
	`)
	return err
}

func migrateCreateSessionRevocationList(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS session_revocation_list (
			token_hash TEXT PRIMARY KEY,
			revoked_at INTEGER NOT NULL,
			reason TEXT
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_revocation_token ON session_revocation_list(token_hash)
	`)
	return err
}

func migrateCreateFailedLoginAttempts(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS failed_login_attempts (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			attempted_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_failed_attempts_operator ON failed_login_attempts(operator_id)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_failed_attempts_ip ON failed_login_attempts(ip_address)
	`)
	return err
}

func migrateCreateAccountLockouts(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS account_lockouts (
			operator_id TEXT PRIMARY KEY,
			locked_until INTEGER NOT NULL,
			reason TEXT,
			created_at INTEGER NOT NULL
		)
	`)
	return err
}

func migrateCreateAuditLogs(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			operator_id TEXT,
			action TEXT NOT NULL,
			resource_type TEXT,
			resource_id TEXT,
			ip_address TEXT,
			user_agent TEXT,
			metadata TEXT,
			result TEXT NOT NULL,
			created_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_audit_operator ON audit_logs(operator_id)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at DESC)
	`)
	return err
}