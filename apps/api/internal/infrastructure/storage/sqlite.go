package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Config holds SQLite configuration.
type Config struct {
	Path        string
	JournalMode string // WAL, DELETE, etc.
	CacheSize   int    // KB, negative means MB
	BusyTimeout int    // milliseconds
	ForeignKeys bool
}

// DefaultConfig returns the default SQLite configuration.
func DefaultConfig(dbPath string) *Config {
	return &Config{
		Path:        dbPath,
		JournalMode: "WAL",
		CacheSize:   -2000, // 2GB cache
		BusyTimeout: 5000,  // 5 seconds
		ForeignKeys: true,
	}
}

// buildDSN builds the SQLite DSN from config.
func (c *Config) buildDSN() string {
	dsn := c.Path

	// Ensure directory exists.
	dir := filepath.Dir(c.Path)
	if dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	// Add query parameters.
	params := "?_journal_mode=" + c.JournalMode
	params += "&_cache_size=" + fmt.Sprintf("%d", c.CacheSize)
	params += "&_busy_timeout=" + fmt.Sprintf("%d", c.BusyTimeout)
	params += "&_foreign_keys="

	if c.ForeignKeys {
		params += "1"
	} else {
		params += "0"
	}

	return dsn + params
}

// SQLite provides the base SQLite database connection.
type SQLite struct {
	db *sql.DB
}

// Open opens a SQLite database connection.
func Open(cfg *Config) (*SQLite, error) {
	db, err := sql.Open("sqlite3", cfg.buildDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool.
	db.SetMaxOpenConns(1) // SQLite doesn't handle concurrent writes well
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Verify connection.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &SQLite{db: db}, nil
}

// DB returns the underlying sql.DB.
func (s *SQLite) DB() *sql.DB {
	return s.db
}

// Close closes the database connection.
func (s *SQLite) Close() error {
	return s.db.Close()
}

// Ping checks if the database is reachable.
func (s *SQLite) Ping() error {
	return s.db.Ping()
}

// BeginTx starts a new transaction.
func (s *SQLite) BeginTx() (*sql.Tx, error) {
	return s.db.Begin()
}

// WithTx executes a function within a transaction.
// If the function returns an error, the transaction is rolled back.
// If the function succeeds, the transaction is committed.
func (s *SQLite) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Attach transaction to context
	txCtx := transaction.ContextWithTx(ctx, tx)

	// Execute function with transaction context
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TxManager returns a transaction.TXManager implementation.
func (s *SQLite) TxManager() transaction.TxManager {
	return s
}

// =============================================================================
// Migrations
// =============================================================================

// Migration represents a database migration.
type Migration struct {
	Apply   func(*sql.DB) error
	Name    string
	Version int
}

// migrations is the registry of all database migrations.
var migrations = []Migration{
	{Apply: migrateCreateDevices, Name: "create_devices_table", Version: 1},
	{Apply: migrateCreateTelemetry, Name: "create_telemetry_table", Version: 2},
	{Apply: migrateCreateCommands, Name: "create_commands_table", Version: 3},
	{Apply: migrateCreateOperators, Name: "create_operators_table", Version: 4},
	{Apply: migrateCreateAuthSessions, Name: "create_auth_sessions_table", Version: 5},
	{Apply: migrateCreateEmailVerifications, Name: "create_email_verifications_table", Version: 6},
	{Apply: migrateCreatePasswordReset, Name: "create_password_reset_table", Version: 7},
	{Apply: migrateCreateSettings, Name: "create_settings_table", Version: 8},
	{Apply: migrateAddCommandsColumns, Name: "add_commands_columns", Version: 9},
	{Apply: migrateAddDeviceSecretHash, Name: "add_device_secret_hash", Version: 10},
	{Apply: migrateAddOperatorsGitHubID, Name: "add_operators_github_id", Version: 11},
	{Apply: migrateCreateResendTracker, Name: "create_resend_tracker_table", Version: 12},
	{Apply: migrateCreateAPIClients, Name: "create_api_clients_table", Version: 13},
	{Apply: migrateCreateSigningKeys, Name: "create_signing_keys_table", Version: 14},
	{Apply: migrateCreateSessionRevocationList, Name: "create_session_revocation_list", Version: 15},
	{Apply: migrateCreateFailedLoginAttempts, Name: "create_failed_login_attempts", Version: 16},
	{Apply: migrateCreateAccountLockouts, Name: "create_account_lockouts", Version: 17},
	{Apply: migrateCreateAuditLogs, Name: "create_audit_logs", Version: 18},
	{Apply: migrateCreateMessageQueue, Name: "create_message_queue_table", Version: 19},
	// v20-v27: New tables for enterprise features
	{Apply: migrateCreateEvents, Name: "create_events_table", Version: 20},
	{Apply: migrateCreateInboxAndRegistration, Name: "create_inbox_and_registration_tables", Version: 21},
	{Apply: migrateCreateDeviceLogsAndEvents, Name: "create_device_logs_and_events_tables", Version: 22},
	{Apply: migrateCreateUpdateTables, Name: "create_update_tables", Version: 23},
	{Apply: migrateCreateOperatorSettings, Name: "create_operator_settings_table", Version: 24},
	{Apply: migrateCreateRefreshTokens, Name: "create_refresh_tokens_table", Version: 25},
	{Apply: migrateCreateNotificationAuditLog, Name: "create_notification_audit_log_table", Version: 26},
	{Apply: migrateAddDevicesColumns, Name: "add_devices_columns", Version: 27},
	{Apply: migrateCreateUpdateSyncStatus, Name: "create_update_sync_status_table", Version: 28},
	{Apply: migrateDashboardDeviceLogs, Name: "create_dashboard_device_logs_table", Version: 29},
	{Apply: migrateTelemetryUptime, Name: "add_telemetry_uptime_column", Version: 30},
	{Apply: migrateInboxIMEIUnique, Name: "add_inbox_imei_unique_constraint", Version: 31},
	{Apply: migrateRegistrationAuditFields, Name: "add_registration_audit_fields", Version: 32},
	{Apply: migrateIdempotencyRecords, Name: "create_idempotency_records_table", Version: 33},
	{Apply: migrateDeviceEvents, Name: "create_device_events_table", Version: 34},
	{Apply: migrateDeviceEventsExtended, Name: "add_device_events_extended_columns", Version: 35},
	{Apply: migrateOperatorFCMToken, Name: "add_operator_fcm_token_column", Version: 36},
	{Apply: migrateCreateOAuthStates, Name: "create_oauth_states_table", Version: 37},
	{Apply: migrateAddMFASecretMAC, Name: "add_mfa_secret_mac_column", Version: 38},
}

// runMigrations applies all pending migrations.
func runMigrations(db *sql.DB) error {
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
			continue
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

func createMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at INTEGER NOT NULL)`)
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
	_, err := db.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, time.Now().UTC().UnixMilli())
	return err
}

// =============================================================================
// Individual Migration Functions
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
			last_seen INTEGER NOT NULL,
			command_secret_hash TEXT,
			operator_id TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)

	return err
}

func migrateCreateTelemetry(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS telemetry (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL,
			received_at INTEGER NOT NULL,
			frame_data TEXT,
			risk_score INTEGER,
			buffer_level INTEGER,
			thermal_temp REAL,
			uptime INTEGER DEFAULT 0,
			FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry(device_id, received_at DESC)
	`)

	return err
}

func migrateCreateCommands(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS commands (
			id TEXT PRIMARY KEY,
			dispatch_id TEXT NOT NULL UNIQUE,
			device_id TEXT NOT NULL,
			command TEXT NOT NULL,
			args TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			delivered_at INTEGER,
			completed_at INTEGER,
			status TEXT NOT NULL DEFAULT 'pending',
			failure_reason TEXT,
			wake_sent INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
		)
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
			mfa_secret TEXT,
			mfa_secret_mac TEXT,
			mfa_enabled INTEGER NOT NULL DEFAULT 0,
			backup_codes TEXT,
			email_verified INTEGER NOT NULL DEFAULT 0,
			google_id TEXT,
			github_id TEXT,
			thresholds TEXT,
			client_settings TEXT,
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

	return err
}

func migrateAddCommandsColumns(db *sql.DB) error {
	// Add columns if they don't exist (ALTER TABLE is idempotent-safe in SQLite)
	cols := []string{
		`ALTER TABLE commands ADD COLUMN wake_sent INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE commands ADD COLUMN failure_reason TEXT`,
	}
	for _, col := range cols {
		if _, err := db.ExecContext(context.Background(), col); err != nil {
			// Column may already exist (SQLite ignores duplicate column additions)
			if !isColumnExistsError(err) {
				return err
			}
		}
	}

	return nil
}

func migrateAddDeviceSecretHash(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `ALTER TABLE devices ADD COLUMN command_secret_hash TEXT`)
	if err != nil {
		// Column may already exist (SQLite ignores duplicate column additions)
		if !isColumnExistsError(err) {
			return fmt.Errorf("failed to add command_secret_hash column: %w", err)
		}
	}
	return nil
}

func migrateAddOperatorsGitHubID(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `ALTER TABLE operators ADD COLUMN github_id TEXT`)
	if err != nil {
		// Column may already exist
		if !isColumnExistsError(err) {
			return fmt.Errorf("failed to add github_id column: %w", err)
		}
	}
	return nil
}

func migrateAddMFASecretMAC(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `ALTER TABLE operators ADD COLUMN mfa_secret_mac TEXT`)
	if err != nil {
		// Column may already exist
		if !isColumnExistsError(err) {
			return fmt.Errorf("failed to add mfa_secret_mac column: %w", err)
		}
	}
	return nil
}

// isColumnExistsError checks if the error is because the column already exists.
func isColumnExistsError(err error) bool {
	if err == nil {
		return false // No error means column was added successfully
	}
	// Check for SQLite's duplicate column error
	return strings.Contains(err.Error(), "duplicate column")
}

func migrateCreateResendTracker(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS password_reset_resend_tracker (
			id TEXT PRIMARY KEY,
			email_hash TEXT NOT NULL UNIQUE,
			resend_count INTEGER NOT NULL DEFAULT 1,
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
			name TEXT NOT NULL,
			api_key_hash TEXT NOT NULL,
			hmac_key_hash TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)

	return err
}

func migrateCreateSigningKeys(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS signing_keys (
			id TEXT PRIMARY KEY,
			client_id TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			expires_at INTEGER,
			FOREIGN KEY(client_id) REFERENCES api_clients(id) ON DELETE CASCADE
		)
	`)

	return err
}

func migrateCreateSessionRevocationList(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS session_revocations (
			session_id TEXT PRIMARY KEY,
			revoked_at INTEGER NOT NULL,
			reason TEXT
		)
	`)

	return err
}

func migrateCreateFailedLoginAttempts(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS failed_login_attempts (
			email TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			attempted_at INTEGER NOT NULL,
			success INTEGER NOT NULL DEFAULT 0
		)
	`)

	return err
}

func migrateCreateAccountLockouts(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS account_lockouts (
			email TEXT PRIMARY KEY,
			locked_until INTEGER,
			attempt_count INTEGER NOT NULL DEFAULT 0,
			first_attempt_at INTEGER NOT NULL
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
			details TEXT,
			ip_address TEXT,
			user_agent TEXT,
			created_at INTEGER NOT NULL
		)
	`)

	return err
}

func migrateCreateMessageQueue(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS message_queue (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL,
			frame_json TEXT NOT NULL,
			enqueued_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index for efficient queries
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_message_queue_device_expires 
		ON message_queue(device_id, expires_at)
	`)

	return err
}

func migrateCreateOAuthStates(db *sql.DB) error {
	// 8: Persist OAuth state to database to prevent CSRF attacks
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS oauth_states (
			id TEXT PRIMARY KEY,
			state TEXT NOT NULL,
			redirect_url TEXT NOT NULL,
			provider TEXT NOT NULL,
			expires_at INTEGER NOT NULL,
			created_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index for state lookups and cleanup
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_oauth_states_expires 
		ON oauth_states(expires_at)
	`)

	return err
}
