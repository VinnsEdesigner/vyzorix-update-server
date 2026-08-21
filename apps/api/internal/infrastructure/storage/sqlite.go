package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	// CGO SQLite driver for local-file backend. Registered as "sqlite3".
	_ "github.com/mattn/go-sqlite3"

	// Pure-Go libSQL driver for the Turso (remote) backend. Registered as "libsql".
	// Implements database/sql over the Turso HTTP/v2 pipeline, so every existing.
	// repository method works unchanged against either backend.
	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Backend identifies which storage backend a SQLite handle is bound to.
type Backend string

const (
	BackendSQLite Backend = "sqlite" // local file via mattn/go-sqlite3.
	BackendTurso  Backend = "turso"  // remote Turso libSQL over HTTP.
)

// Config holds storage configuration for either backend.
type Config struct {
	Path              string
	TursoURL          string
	TursoAuthToken    string
	Backend           Backend
	JournalMode       string
	ConnMaxLifetime   time.Duration
	HealthCheckPeriod time.Duration
	MaxOpenConns      int
	MaxIdleConns      int
	RequestTimeout    time.Duration
	ConnMaxIdleTime   time.Duration
	CacheSize         int
	BusyTimeout       int
	ForeignKeys       bool
}

// DefaultConfig returns the default SQLite configuration for a local file path.
func DefaultConfig(dbPath string) *Config {
	return &Config{
		Backend:     BackendSQLite,
		Path:        dbPath,
		JournalMode: "WAL",
		CacheSize:   -2000, // 2GB cache.
		BusyTimeout: 5000,  // 5 seconds.
		ForeignKeys: true,
	}
}

// DefaultTursoConfig returns the default Turso (remote libSQL) configuration.
func DefaultTursoConfig(url, authToken string) *Config {
	return &Config{
		Backend:           BackendTurso,
		TursoURL:          url,
		TursoAuthToken:    authToken,
		MaxOpenConns:      16,
		MaxIdleConns:      8,
		ConnMaxLifetime:   30 * time.Minute,
		ConnMaxIdleTime:   5 * time.Minute,
		RequestTimeout:    15 * time.Second,
		HealthCheckPeriod: 30 * time.Second,
	}
}

// resolvedBackend auto-detects the backend when Backend is empty.
func (c *Config) resolvedBackend() Backend {
	if c.Backend != "" {
		return c.Backend
	}
	if c.TursoURL != "" {
		return BackendTurso
	}
	return BackendSQLite
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

// buildTursoDSN builds the libSQL connection string with the auth token appended.
// as a query parameter (the pure-Go driver reads it from the URL).
func (c *Config) buildTursoDSN() string {
	url := c.TursoURL
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	return url + sep + "authToken=" + c.TursoAuthToken
}

// SQLite provides the base database connection. Despite the legacy name, it is.
// backend-agnostic: the underlying *sql.DB may be a local sqlite3 file or a.
// remote Turso libSQL endpoint. The public surface (DB/Close/Ping/WithTx/.
// TxManager/Backend/Info) is identical for both backends.
type SQLite struct {
	db         *sql.DB
	cfg        *Config
	logger     *slog.Logger
	stopHealth chan struct{}
	backend    Backend
	stopped    atomic.Bool
}

// Open opens a database connection using the backend selected in cfg.
func Open(cfg *Config) (*SQLite, error) {
	return openWithLogger(cfg, nil)
}

// OpenWithLogger is like Open but wires a structured logger for health checks.
func OpenWithLogger(cfg *Config, log *slog.Logger) (*SQLite, error) {
	return openWithLogger(cfg, log)
}

func openWithLogger(cfg *Config, log *slog.Logger) (*SQLite, error) {
	if cfg == nil {
		return nil, errors.New("storage: nil config")
	}
	backend := cfg.resolvedBackend()

	var (
		db      *sql.DB
		dsn     string
		drvName string
		err     error
	)
	switch backend {
	case BackendTurso:
		if cfg.TursoURL == "" {
			return nil, errors.New("storage: turso backend requires TursoURL")
		}
		if cfg.TursoAuthToken == "" {
			return nil, errors.New("storage: turso backend requires TursoAuthToken")
		}
		drvName = "libsql"
		dsn = cfg.buildTursoDSN()
		db, err = sql.Open(drvName, dsn)
		if err != nil {
			return nil, fmt.Errorf("storage: failed to open turso connection: %w", err)
		}
		// Turso tolerates concurrent connections; pool is configured below.
	case BackendSQLite:
		drvName = "sqlite3"
		dsn = cfg.buildDSN()
		db, err = sql.Open(drvName, dsn)
		if err != nil {
			return nil, fmt.Errorf("storage: failed to open sqlite: %w", err)
		}
		// SQLite serializes writes; keep the single-writer invariant.
		cfg.MaxOpenConns = 1
		cfg.MaxIdleConns = 1
		cfg.ConnMaxLifetime = time.Hour
	default:
		return nil, fmt.Errorf("storage: unknown backend %q", backend)
	}

	applyPool(db, cfg, backend)

	// Verify connectivity before running migrations so a misconfigured Turso.
	// endpoint fails fast at boot with a clear cause.
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("storage: %s ping failed: %w", backend, err)
	}

	// Run migrations. The DDL is SQLite-compatible and applies to libSQL too,
	// so one migration registry serves both backends (no divergent Turso path).
	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("storage: %s migrations failed: %w", backend, err)
	}

	s := &SQLite{
		db:         db,
		backend:    backend,
		cfg:        cfg,
		logger:     log,
		stopHealth: make(chan struct{}),
	}

	// A background health check keeps the Turso pool warm and surfaces.
	// connectivity loss to /health without per-request pings.
	if backend == BackendTurso && cfg.HealthCheckPeriod > 0 {
		s.startHealthCheck()
	}

	return s, nil
}

// applyPool configures connection-pool limits appropriate to the backend.
func applyPool(db *sql.DB, cfg *Config, backend Backend) {
	if backend == BackendTurso {
		if cfg.MaxOpenConns > 0 {
			db.SetMaxOpenConns(cfg.MaxOpenConns)
		}
		if cfg.MaxIdleConns > 0 {
			db.SetMaxIdleConns(cfg.MaxIdleConns)
		}
		if cfg.ConnMaxLifetime > 0 {
			db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		}
		if cfg.ConnMaxIdleTime > 0 {
			db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
		}
		return
	}
	// SQLite: single-writer serial pool.
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
}

// DB returns the underlying sql.DB.
func (s *SQLite) DB() *sql.DB {
	return s.db
}

// Backend reports which storage backend this handle is bound to.
func (s *SQLite) Backend() Backend {
	return s.backend
}

// Info returns human-readable backend metadata for /health and diagnostics.
// It deliberately omits the auth token.
func (s *SQLite) Info() map[string]any {
	info := map[string]any{
		"backend":  string(s.backend),
		"driver":   s.driverName(),
		"max_open": s.cfg.MaxOpenConns,
		"max_idle": s.cfg.MaxIdleConns,
	}
	if s.backend == BackendTurso {
		info["url"] = redactTursoURL(s.cfg.TursoURL)
		info["health_check"] = s.cfg.HealthCheckPeriod.String()
	} else {
		info["path"] = s.cfg.Path
	}
	return info
}

func (s *SQLite) driverName() string {
	if s.backend == BackendTurso {
		return "libsql"
	}
	return "sqlite3"
}

// redactTursoURL strips any authToken query param so the URL is safe to log.
func redactTursoURL(raw string) string {
	if i := strings.Index(raw, "?"); i >= 0 {
		return raw[:i] + "?(redacted)"
	}
	return raw
}

// startHealthCheck runs a periodic ping against the remote backend.
func (s *SQLite) startHealthCheck() {
	if s.logger == nil {
		s.logger = slog.Default()
	}
	go func() {
		t := time.NewTicker(s.cfg.HealthCheckPeriod)
		defer t.Stop()
		for {
			select {
			case <-s.stopHealth:
				return
			case <-t.C:
				ctx, cancel := context.WithTimeout(context.Background(), s.cfg.RequestTimeout)
				if err := s.db.PingContext(ctx); err != nil {
					s.logger.Warn("turso health check failed",
						"backend", string(s.backend),
						"error", err)
				}
				cancel()
			}
		}
	}()
}

// Close closes the database connection and stops the health checker.
func (s *SQLite) Close() error {
	if s.stopHealth != nil && !s.stopped.Swap(true) {
		close(s.stopHealth)
	}
	return s.db.Close()
}

// Ping checks if the database is reachable.
func (s *SQLite) Ping() error {
	return s.db.Ping()
}

// PingContext checks reachability with a deadline.
func (s *SQLite) PingContext(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// BeginTx starts a new transaction.
func (s *SQLite) BeginTx() (*sql.Tx, error) {
	return s.db.Begin()
}

// WithTx executes a function within a transaction.
// If the function returns an error, the transaction is rolled back.
// If the function succeeds, the transaction is committed.
func (s *SQLite) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Attach transaction to context.
	txCtx := transaction.ContextWithTx(ctx, tx)

	// Execute function with transaction context.
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	// Commit transaction.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TxManager returns a transaction.TXManager implementation.
func (s *SQLite) TxManager() transaction.TxManager {
	return s
}

// =============================================================================.
// Migrations.
// =============================================================================.

// Migration represents a database migration.
type Migration struct {
	Apply   func(*sql.Tx) error
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
	// v20-v27: New tables for enterprise features.
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
	// Multi-tenant organization model.
	{Apply: migrateOrganizations, Name: "create_organizations_tables", Version: 39},
	// Organization settings table.
	{Apply: migrateOrganizationSettings, Name: "create_organization_settings_table", Version: 40},
	// Device settings table.
	{Apply: migrateDeviceSettings, Name: "create_device_settings_table", Version: 41},
	// Combined post-V41 migrations: org context, inbox org, MFA tracking.
	{Apply: migratePostV41Combined, Name: "post_v41_combined", Version: 42},
	// Command outbox retry support.
	{Apply: migrateAddCommandRetryColumns, Name: "add_command_retry_columns", Version: 43},
	// Secure command secret storage - hash instead of plaintext.
	{Apply: migrateCommandSecretHash, Name: "add_command_secret_hash", Version: 44},
	// Single-use confirmation token tracking.
	{Apply: migrateConfirmedAt, Name: "add_confirmed_at", Version: 45},
	// Email verification delivery tracking.
	{Apply: migrateEmailVerificationTracking, Name: "add_email_verification_tracking", Version: 46},

	{Apply: migrateCreatePendingFCM, Name: "create_pending_fcm_table", Version: 47},
	// Add the model column to devices. The code referenced it (deviceColumns,
	// INSERT/UPDATE) but no prior migration created it; fresh databases failed.
	// with "no such column: model". Idempotent (skips if column exists).
	{Apply: migrateAddDeviceModelColumn, Name: "add_devices_model_column", Version: 48},
	// Relax devices.command_secret NOT NULL. The column is legacy (plaintext.
	// secret) and superseded by command_secret_hash; DeviceRepository.Create.
	// never sets it, so the NOT NULL constraint broke every device creation.
	// Rebuilds the table to make the column nullable. Idempotent.
	{Apply: migrateRelaxDeviceCommandSecretNull, Name: "relax_devices_command_secret_null", Version: 49},
	// Repair the device_settings foreign key: migration 042 referenced.
	// devices(imei) but the IMEI lives in devices.id, making the FK.
	// unresolvable and blocking device deletion. Rebuilds the table with.
	// REFERENCES devices(id). Idempotent.
	{Apply: migrateFixDeviceSettingsFK, Name: "fix_device_settings_fk", Version: 50},
	// Add organization_id to update_pushes. The org-scoped push queries.
	// (ListPushes/GetPushByID/ListPushDevices) and CreatePush reference it,
	// but the create_update_tables migration (023) never created the column.
	// and no prior migration added it. Idempotent.
	{Apply: migrateAddUpdatePushOrgColumn, Name: "add_update_pushes_organization_id", Version: 51},
	// Rebuild api_keys and api_clients to the rich schema the repository.
	// code expects (operator_id, key_prefix, scope, is_active, ...). The.
	// live tables were created by a deleted migration (038) / migration 013.
	// with an incompatible minimal schema; SetupAPIKeysTable is never wired.
	// and CREATE TABLE IF NOT EXISTS can't add columns, so the rich schema.
	// was never applied and every API-key/client endpoint 500'd. Idempotent.
	{Apply: migrateRebuildAPIKeyTables, Name: "rebuild_api_key_tables", Version: 52},
	// Fix timestamp columns - convert TEXT to INTEGER for consistency.
	// Add resource_type/resource_id/metadata/result columns to audit_logs.
	// The original migration 018 created the table without these columns, so.
	// audit log writes failed with "no column named resource_type". Idempotent.
	{Apply: migrateAuditLogsResourceColumns, Name: "add_audit_logs_resource_columns", Version: 53},
	// Add signing_key column to auth_sessions for per-session HMAC request signing.
	// The browser client receives this key on login and signs every request;
	// the server verifies the signature to confirm the request is from a.
	// verified client. Idempotent.
	{Apply: migrateAddSessionSigningKey, Name: "add_auth_sessions_signing_key", Version: 54},
	// Add signing_secret column to api_keys so API-key-authenticated requests.
	// can be HMAC-signed without a session (Domain A extends to API keys).
	// Idempotent.
	{Apply: migrateAddAPIKeySigningSecret, Name: "add_api_keys_signing_secret", Version: 55},
	// Fix refresh_tokens schema: migration 025 created the revoked flag as.
	// `revoked` and omitted `revoked_at`, but the repository queries.
	// `is_revoked`/`revoked_at`. Rebuilds the table with the correct column.
	// names so logout and token revocation stop 500'ing. Idempotent.
	{Apply: migrateFixRefreshTokensSchema, Name: "fix_refresh_tokens_schema", Version: 56},
	// Add trace_id and risk_tier columns to audit_logs so security events can.
	// be correlated with request traces and classified by command risk. Backs.
	// the Phase 2 risk/audit work. Idempotent: skips columns that already exist.
	{Apply: migrateAuditLogsRiskColumns, Name: "add_audit_logs_risk_columns", Version: 57},
	// Create command_confirmations table for short-lived, single-use.
	// confirmation tokens that gate risky device commands (Phase 3).
	{Apply: migrateCommandConfirmations, Name: "create_command_confirmations_table", Version: 58},
	// Add actor_type, actor_email, old_value, new_value columns to audit_logs.
	// for change-tracking compliance. Idempotent.
	{Apply: migrateAuditLogsChangeTrackingColumns, Name: "add_audit_logs_change_tracking_columns", Version: 59},
	// Scoped permission grants for custom per-resource scopes (Issue 4).
	{Apply: migratePermissionGrants, Name: "create_permission_grants_table", Version: 60},
	{Apply: migrateDeviceGroups, Name: "create_device_groups_tables", Version: 61},
	{Apply: migrateResourcePermissions, Name: "create_resource_permissions_table", Version: 62},
	{Apply: migrateAddTagsAndQuotas, Name: "add_tags_quotas_locks", Version: 63},
	// Alerting engine: org-scoped rules and their evaluation instances.
	{Apply: migrateCreateAlerting, Name: "create_alerting_tables", Version: 64},
	{Apply: migrateAlertHistoryColumns, Name: "add_alert_history_columns", Version: 65},
	{Apply: migrateContactPoints, Name: "create_contact_points", Version: 66},
	{Apply: migrateServiceAccounts, Name: "create_service_accounts", Version: 67},
	{Apply: migrateServiceAccountLastUsed, Name: "add_service_account_last_used", Version: 68},
	{Apply: migrateAnnotations, Name: "create_annotations", Version: 69},
	{Apply: migrateConfigVersions, Name: "create_config_versions", Version: 70},
	{Apply: migrateAlertLabelInstanceKeys, Name: "alert_label_instance_keys", Version: 71},
}

// runMigrations applies all pending migrations.
func runMigrations(db *sql.DB) error {
	// Ensure migrations table exists.
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version.
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Apply pending migrations. Each migration runs inside a transaction so.
	// that a partial failure (e.g. a table-rebuild that drops the old table but.
	// fails to rename the new one) rolls back atomically and never leaves the.
	// schema in a half-rebuilt "bricked" state. SQLite WAL mode supports.
	// transactional DDL, so CREATE/DROP/RENAME all roll back on failure.
	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("migration %d (%s) failed to begin tx: %w", m.Version, m.Name, err)
		}

		if err := m.Apply(tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %d (%s) failed: %w", m.Version, m.Name, err)
		}

		if err := setVersionTx(tx, m.Version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to set version after migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("migration %d (%s) failed to commit: %w", m.Version, m.Name, err)
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
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	return version, err
}

func setVersionTx(tx *sql.Tx, version int) error {
	_, err := tx.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, time.Now().UTC().UnixMilli())
	return err
}

// =============================================================================.
// Individual Migration Functions.
// =============================================================================.

func migrateCreateDevices(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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

func migrateCreateTelemetry(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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

	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry(device_id, received_at DESC)
	`)

	return err
}

func migrateCreateCommands(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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

func migrateCreateOperators(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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
			last_organization_id TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)

	return err
}

func migrateCreateAuthSessions(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS auth_sessions (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			organization_id TEXT,
			expires_at INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			user_agent TEXT,
			ip_address TEXT,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)
	`)

	return err
}

func migrateCreateEmailVerifications(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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

func migrateCreatePasswordReset(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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

func migrateCreateSettings(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)

	return err
}

func migrateAddCommandsColumns(tx *sql.Tx) error {
	// Add columns if they don't exist (with idempotent error handling for SQLite).
	cols := []struct {
		sql  string
		name string
	}{
		{`ALTER TABLE commands ADD COLUMN wake_sent INTEGER NOT NULL DEFAULT 0`, "wake_sent"},
		{`ALTER TABLE commands ADD COLUMN failure_reason TEXT`, "failure_reason"},
	}
	for _, col := range cols {
		_, err := tx.ExecContext(context.Background(), col.sql)
		if err != nil {
			// Column may already exist (SQLite ignores duplicate column additions).
			if !isColumnExistsError(err) {
				return fmt.Errorf("failed to add %s column: %w", col.name, err)
			}
		}
	}

	return nil
}

func migrateAddDeviceSecretHash(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `ALTER TABLE devices ADD COLUMN command_secret_hash TEXT`)
	if err != nil {
		// Column may already exist (SQLite ignores duplicate column additions).
		if !isColumnExistsError(err) {
			return fmt.Errorf("failed to add command_secret_hash column: %w", err)
		}
	}
	return nil
}

func migrateAddOperatorsGitHubID(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `ALTER TABLE operators ADD COLUMN github_id TEXT`)
	if err != nil {
		// Column may already exist.
		if !isColumnExistsError(err) {
			return fmt.Errorf("failed to add github_id column: %w", err)
		}
	}
	return nil
}

func migrateAddMFASecretMAC(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `ALTER TABLE operators ADD COLUMN mfa_secret_mac TEXT`)
	if err != nil {
		// Column may already exist.
		if !isColumnExistsError(err) {
			return fmt.Errorf("failed to add mfa_secret_mac column: %w", err)
		}
	}
	return nil
}

// isColumnExistsError checks if the error is because the column already exists.
// SQLite silently ignores duplicate column additions.
func isColumnExistsError(err error) bool {
	if err == nil {
		return true // No error means success (column added or already exists).
	}
	// SQLite doesn't error on duplicate column - it just ignores it.
	return strings.Contains(err.Error(), "duplicate column")
}

func migrateCreateResendTracker(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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

func migrateCreateAPIClients(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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

func migrateCreateSigningKeys(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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

func migrateCreateSessionRevocationList(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS session_revocations (
			session_id TEXT PRIMARY KEY,
			revoked_at INTEGER NOT NULL,
			reason TEXT
		)
	`)

	return err
}

func migrateCreateFailedLoginAttempts(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS failed_login_attempts (
			email TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			attempted_at INTEGER NOT NULL,
			success INTEGER NOT NULL DEFAULT 0
		)
	`)

	return err
}

func migrateCreateAccountLockouts(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS account_lockouts (
			email TEXT PRIMARY KEY,
			locked_until INTEGER,
			attempt_count INTEGER NOT NULL DEFAULT 0,
			first_attempt_at INTEGER NOT NULL
		)
	`)

	return err
}

func migrateCreateAuditLogs(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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

// migrateAuditLogsResourceColumns adds the resource_type, resource_id, metadata,
// and result columns that the audit repository writes but migration 018 never.
// created. Idempotent: silently skips columns that already exist.
func migrateAuditLogsResourceColumns(tx *sql.Tx) error {
	for _, col := range []string{"resource_type", "resource_id", "metadata", "result"} {
		if _, err := tx.ExecContext(context.Background(),
			`ALTER TABLE audit_logs ADD COLUMN `+col+` TEXT`); err != nil {
			if !isDuplicateColumnErr(err) {
				return err
			}
		}
	}
	return nil
}

// isDuplicateColumnErr returns true for SQLite's "duplicate column name" error,
// which is raised when ALTER TABLE ADD COLUMN targets an existing column.
func isDuplicateColumnErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate column")
}

// migrateAuditLogsRiskColumns adds the trace_id and risk_tier columns to.
// audit_logs, enabling audit entries to be correlated with request traces and.
// classified by the risk tier of the audited operation. Idempotent.
func migrateAuditLogsRiskColumns(tx *sql.Tx) error {
	for _, col := range []string{"trace_id", "risk_tier"} {
		if _, err := tx.ExecContext(context.Background(),
			`ALTER TABLE audit_logs ADD COLUMN `+col+` TEXT`); err != nil {
			if !isDuplicateColumnErr(err) {
				return err
			}
		}
	}
	return nil
}

// migrateAuditLogsChangeTrackingColumns adds actor_type, actor_email, old_value,
// and new_value columns to audit_logs for change-tracking compliance. Idempotent.
func migrateAuditLogsChangeTrackingColumns(tx *sql.Tx) error {
	for _, col := range []string{"actor_type", "actor_email", "old_value", "new_value"} {
		if _, err := tx.ExecContext(context.Background(),
			`ALTER TABLE audit_logs ADD COLUMN `+col+` TEXT`); err != nil {
			if !isDuplicateColumnErr(err) {
				return err
			}
		}
	}
	return nil
}

// migrateAddSessionSigningKey adds the signing_key column to auth_sessions.
// This column stores the per-session HMAC secret used by the browser client.
// to sign every request. Idempotent — safe to run on databases that already.
// have the column.
func migrateAddSessionSigningKey(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(),
		`ALTER TABLE auth_sessions ADD COLUMN signing_key TEXT`)
	if err != nil && !isDuplicateColumnErr(err) {
		return err
	}
	return nil
}

func migrateCreateMessageQueue(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
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

	// Create index for efficient queries.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_message_queue_device_expires 
		ON message_queue(device_id, expires_at)
	`)

	return err
}

func migrateCreateOAuthStates(tx *sql.Tx) error {
	// 8: Persist OAuth state to database to prevent CSRF attacks.
	_, err := tx.ExecContext(context.Background(), `
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

	// Create index for state lookups and cleanup.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_oauth_states_expires 
		ON oauth_states(expires_at)
	`)

	return err
}

// migratePostV41Combined combines all post-V41 migrations into a single migration.
// This includes: org context columns, inbox org column, and MFA tracking.
func migratePostV41Combined(tx *sql.Tx) error {
	// Add last_organization_id to operators table for auto-select on login.
	_, err := tx.ExecContext(context.Background(), `
		ALTER TABLE operators ADD COLUMN last_organization_id TEXT
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}

	// Add organization_id to auth_sessions table for session-scoped org context.
	_, err = tx.ExecContext(context.Background(), `
		ALTER TABLE auth_sessions ADD COLUMN organization_id TEXT
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}

	// Add organization_id to inbox_requests table for multi-tenant isolation.
	_, err = tx.ExecContext(context.Background(), `
		ALTER TABLE inbox_requests ADD COLUMN organization_id TEXT
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}

	// Create index for org-scoped queries.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_inbox_organization 
		ON inbox_requests(organization_id, status, created_at DESC)
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}

	// Add mfa_enabled_at column to operators table for MFA tracking.
	_, err = tx.ExecContext(context.Background(), `
		ALTER TABLE operators ADD COLUMN mfa_enabled_at INTEGER
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}

	// Add mfa_verified_at column to auth_sessions table for MFA session tracking.
	_, err = tx.ExecContext(context.Background(), `
		ALTER TABLE auth_sessions ADD COLUMN mfa_verified_at INTEGER
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}

	return nil
}

// migrateAddCommandRetryColumns adds retry tracking columns for the outbox pattern.
// This enables the background worker to track retry attempts with exponential backoff.
func migrateAddCommandRetryColumns(tx *sql.Tx) error {
	// Add retry_count column to track number of delivery attempts.
	_, err := tx.ExecContext(context.Background(), `
		ALTER TABLE commands ADD COLUMN retry_count INTEGER NOT NULL DEFAULT 0
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}

	// Add max_retries column to set maximum delivery attempts before marking failed.
	_, err = tx.ExecContext(context.Background(), `
		ALTER TABLE commands ADD COLUMN max_retries INTEGER NOT NULL DEFAULT 5
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}

	// Add next_retry_at column for exponential backoff scheduling.
	_, err = tx.ExecContext(context.Background(), `
		ALTER TABLE commands ADD COLUMN next_retry_at INTEGER
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}

	// Create index for efficient pending retry queries.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_commands_pending_retry
		ON commands(status, next_retry_at)
		WHERE status = 'pending'
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}

	return nil
}
