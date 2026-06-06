package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"strings"

	"github.com/VinnsEdesigner/vyzorix-update-server/models"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

var ErrHijack = errors.New("device_id already registered to a different firebaseInstallId")

type Store struct {
	mu   sync.Mutex
	db   *sql.DB
	path string
}

func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_cache_size=-2000&_busy_timeout=5000&_foreign_keys=1")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: path}
	return s, s.migrate(context.Background())
}
func (s *Store) Close() error                  { return s.db.Close() }
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `PRAGMA journal_mode=WAL`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `PRAGMA cache_size=-2000`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `PRAGMA busy_timeout=5000`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `PRAGMA foreign_keys=ON`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS devices (id TEXT PRIMARY KEY, firebase_install_id TEXT NOT NULL, fcm_token TEXT, app_version TEXT, device_class TEXT, command_secret TEXT NOT NULL, online INTEGER NOT NULL DEFAULT 0, registered_at INTEGER NOT NULL, last_seen INTEGER NOT NULL)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS telemetry (id INTEGER PRIMARY KEY AUTOINCREMENT, device_id TEXT NOT NULL, received_at INTEGER NOT NULL, payload TEXT NOT NULL, risk_score INTEGER, buffer_level INTEGER, thermal_temp REAL, FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS commands (dispatch_id TEXT PRIMARY KEY, device_id TEXT NOT NULL, command TEXT NOT NULL, args TEXT, delivery TEXT NOT NULL, created_at INTEGER NOT NULL, delivered_at INTEGER, wake_sent INTEGER NOT NULL DEFAULT 0, wake_error TEXT, FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry(device_id, received_at DESC)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_commands_device_time ON commands(device_id, created_at DESC)`); err != nil {
		return err
	}
	// Additive migrations
	s.db.ExecContext(ctx, `ALTER TABLE commands ADD COLUMN wake_sent INTEGER NOT NULL DEFAULT 0`) //nolint:errcheck
	s.db.ExecContext(ctx, `ALTER TABLE commands ADD COLUMN wake_error TEXT`)                   //nolint:errcheck
	s.db.ExecContext(ctx, `ALTER TABLE commands ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'`) //nolint:errcheck
	s.db.ExecContext(ctx, `ALTER TABLE commands ADD COLUMN completed_at INTEGER`)               //nolint:errcheck
	s.db.ExecContext(ctx, `ALTER TABLE commands ADD COLUMN result TEXT`)                      //nolint:errcheck
	// Command secret hash column for audit/compliance
	s.db.ExecContext(ctx, `ALTER TABLE devices ADD COLUMN command_secret_hash TEXT`)        //nolint:errcheck
	return s.migrateAuth(ctx)
}

func (s *Store) migrateAuth(ctx context.Context) error {
	// Operators table — stores human dashboard users
	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS operators (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			password_hash TEXT,
			role TEXT NOT NULL DEFAULT 'operator',
			google_id TEXT UNIQUE,
			email_verified INTEGER NOT NULL DEFAULT 0,
			verification_sent_at INTEGER,
			risk_warn INTEGER NOT NULL DEFAULT 70,
			risk_crit INTEGER NOT NULL DEFAULT 90,
			thermal_warn INTEGER NOT NULL DEFAULT 42,
			thermal_crit INTEGER NOT NULL DEFAULT 45,
			buffer_warn INTEGER NOT NULL DEFAULT 60,
			buffer_crit INTEGER NOT NULL DEFAULT 80,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`); err != nil {
		return err
	}
	// Active sessions — JWT token hashes for revocation support
	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS auth_sessions (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			user_agent TEXT,
			ip_address TEXT,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_auth_sessions_operator ON auth_sessions(operator_id)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires ON auth_sessions(expires_at)`); err != nil {
		return err
	}
	// Email verification tokens
	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS email_verifications (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_email_verifications_operator ON email_verifications(operator_id)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token_hash)`); err != nil {
		return err
	}
	// Password reset tokens
	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at INTEGER NOT NULL,
			used_at INTEGER,
			created_at INTEGER NOT NULL,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_password_reset_operator ON password_reset_tokens(operator_id)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_password_reset_token ON password_reset_tokens(token_hash)`); err != nil {
		return err
	}
	// Additive migrations for new columns
	s.db.ExecContext(ctx, `ALTER TABLE operators ADD COLUMN email_verified INTEGER NOT NULL DEFAULT 0`) //nolint:errcheck
	s.db.ExecContext(ctx, `ALTER TABLE operators ADD COLUMN verification_sent_at INTEGER`)             //nolint:errcheck
	return s.migrateSettings(ctx)
}

// migrateSettings creates the settings table for system configuration.
func (s *Store) migrateSettings(ctx context.Context) error {
	// Settings table for system configuration
	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`); err != nil {
		return err
	}

	// Set defaults if not exist (per COMMAND_SECURITY.md: 30-second HMAC window)
	defaults := map[string]string{
		"enforce_hmac":        "false",
		"hmac_window_seconds": "30",
		"rate_limit_capacity": "100",
		"rate_limit_refill":   "60",
	}

	for key, value := range defaults {
		//nolint:errcheck // INSERT OR IGNORE is best-effort for default settings
		s.db.ExecContext(ctx, `
			INSERT OR IGNORE INTO settings(key, value, updated_at) VALUES(?, ?, ?)
		`, key, value, time.Now().UTC().UnixMilli())
	}

	return nil
}

func (s *Store) Register(ctx context.Context, req models.RegisterRequest) (struct {
	ID                string
	FirebaseInstallID string
	FCMToken          string
	AppVersion        string
	DeviceClass       string
	CommandSecret     string
	Online            bool
	RegisteredAt      time.Time
	LastSeen          time.Time
}, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()

	var existingID, existingFID string
	err := s.db.QueryRowContext(ctx, `SELECT id, firebase_install_id FROM devices WHERE id = ?`, req.DeviceID).Scan(&existingID, &existingFID)
	if err == nil {
		if existingFID != req.FirebaseInstallID {
			return struct {
				ID                string
				FirebaseInstallID string
				FCMToken          string
				AppVersion        string
				DeviceClass       string
				CommandSecret     string
				Online            bool
				RegisteredAt      time.Time
				LastSeen          time.Time
			}{}, false, ErrHijack
		}
		_, err = s.db.ExecContext(ctx, `UPDATE devices SET fcm_token=?, app_version=?, device_class=?, last_seen=? WHERE id=?`,
			req.FCMToken, req.AppVersion, req.DeviceClass, now.UnixMilli(), req.DeviceID)
		if err != nil {
			return struct {
				ID                string
				FirebaseInstallID string
				FCMToken          string
				AppVersion        string
				DeviceClass       string
				CommandSecret     string
				Online            bool
				RegisteredAt      time.Time
				LastSeen          time.Time
			}{}, false, err
		}
		var cmdSecret string
		var regAt int64
		_ = s.db.QueryRowContext(ctx, `SELECT command_secret, registered_at FROM devices WHERE id = ?`, req.DeviceID).Scan(&cmdSecret, &regAt)
		return struct {
			ID                string
			FirebaseInstallID string
			FCMToken          string
			AppVersion        string
			DeviceClass       string
			CommandSecret     string
			Online            bool
			RegisteredAt      time.Time
			LastSeen          time.Time
		}{ID: req.DeviceID, FirebaseInstallID: existingFID, FCMToken: req.FCMToken, AppVersion: req.AppVersion, DeviceClass: req.DeviceClass, CommandSecret: cmdSecret, Online: true, RegisteredAt: time.UnixMilli(regAt), LastSeen: now}, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return struct {
			ID                string
			FirebaseInstallID string
			FCMToken          string
			AppVersion        string
			DeviceClass       string
			CommandSecret     string
			Online            bool
			RegisteredAt      time.Time
			LastSeen          time.Time
		}{}, false, err
	}
	secret, err := randomHex(32)
	if err != nil {
		return struct {
			ID                string
			FirebaseInstallID string
			FCMToken          string
			AppVersion        string
			DeviceClass       string
			CommandSecret     string
			Online            bool
			RegisteredAt      time.Time
			LastSeen          time.Time
		}{}, false, err
	}
	// Generate bcrypt hash of the secret for audit/compliance
	hasher := NewSecretHash()
	secretHash, _ := hasher.HashSecret(secret) // Error ignored - hash is optional

	_, err = s.db.ExecContext(ctx, `INSERT INTO devices(id,firebase_install_id,fcm_token,app_version,device_class,command_secret,command_secret_hash,online,registered_at,last_seen) VALUES(?,?,?,?,?,?,?,0,?,?)`,
		req.DeviceID, req.FirebaseInstallID, req.FCMToken, req.AppVersion, req.DeviceClass, secret, secretHash, now.UnixMilli(), now.UnixMilli())
	if err != nil {
		return struct {
			ID                string
			FirebaseInstallID string
			FCMToken          string
			AppVersion        string
			DeviceClass       string
			CommandSecret     string
			Online            bool
			RegisteredAt      time.Time
			LastSeen          time.Time
		}{}, false, err
	}
	return struct {
		ID                string
		FirebaseInstallID string
		FCMToken          string
		AppVersion        string
		DeviceClass       string
		CommandSecret     string
		Online            bool
		RegisteredAt      time.Time
		LastSeen          time.Time
	}{ID: req.DeviceID, FirebaseInstallID: req.FirebaseInstallID, FCMToken: req.FCMToken, AppVersion: req.AppVersion, DeviceClass: req.DeviceClass, CommandSecret: secret, RegisteredAt: now, LastSeen: now}, true, nil
}

type deviceRow struct {
	ID                string
	FirebaseInstallID string
	FCMToken          string
	AppVersion        string
	DeviceClass       string
	CommandSecret     string
	Online            int
	RegisteredAt      int64
	LastSeen          int64
}

func rowToDevice(r deviceRow) struct {
	ID                string
	FirebaseInstallID string
	FCMToken          string
	AppVersion        string
	DeviceClass       string
	CommandSecret     string
	Online            bool
	RegisteredAt      time.Time
	LastSeen          time.Time
} {
	return struct {
		ID                string
		FirebaseInstallID string
		FCMToken          string
		AppVersion        string
		DeviceClass       string
		CommandSecret     string
		Online            bool
		RegisteredAt      time.Time
		LastSeen          time.Time
	}{ID: r.ID, FirebaseInstallID: r.FirebaseInstallID, FCMToken: r.FCMToken, AppVersion: r.AppVersion, DeviceClass: r.DeviceClass, CommandSecret: r.CommandSecret, Online: r.Online != 0, RegisteredAt: time.UnixMilli(r.RegisteredAt).UTC(), LastSeen: time.UnixMilli(r.LastSeen).UTC()}
}

func (s *Store) Device(ctx context.Context, id string) (struct {
	ID                string
	FirebaseInstallID string
	FCMToken          string
	AppVersion        string
	DeviceClass       string
	CommandSecret     string
	Online            bool
	RegisteredAt      time.Time
	LastSeen          time.Time
}, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var r deviceRow
	err := s.db.QueryRowContext(ctx, `SELECT id, firebase_install_id, fcm_token, app_version, device_class, command_secret, online, registered_at, last_seen FROM devices WHERE id = ?`, id).Scan(&r.ID, &r.FirebaseInstallID, &r.FCMToken, &r.AppVersion, &r.DeviceClass, &r.CommandSecret, &r.Online, &r.RegisteredAt, &r.LastSeen)
	if errors.Is(err, sql.ErrNoRows) {
		return struct {
			ID                string
			FirebaseInstallID string
			FCMToken          string
			AppVersion        string
			DeviceClass       string
			CommandSecret     string
			Online            bool
			RegisteredAt      time.Time
			LastSeen          time.Time
		}{}, false, nil
	}
	if err != nil {
		return struct {
			ID                string
			FirebaseInstallID string
			FCMToken          string
			AppVersion        string
			DeviceClass       string
			CommandSecret     string
			Online            bool
			RegisteredAt      time.Time
			LastSeen          time.Time
		}{}, false, err
	}
	return rowToDevice(r), true, nil
}

func (s *Store) Secret(ctx context.Context, id string) (string, bool) {
	d, ok, err := s.Device(ctx, id)
	if err != nil || !ok {
		return "", false
	}
	return d.CommandSecret, true
}

// SecretHash provides bcrypt hash utilities for command secrets.
type SecretHash struct{}

// NewSecretHash creates a new SecretHash utility.
func NewSecretHash() *SecretHash { return &SecretHash{} }

// HashSecret generates a bcrypt hash of the given secret for storage/audit.
// The secret is used directly for HMAC verification; this hash is for compliance.
func (h *SecretHash) HashSecret(secret string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyHash checks if the secret matches the stored bcrypt hash.
// Returns true if the secret matches the hash.
func (h *SecretHash) VerifyHash(secret, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret))
	return err == nil
}

// SetSecretHash stores the bcrypt hash of a device's command secret.
func (s *Store) SetSecretHash(ctx context.Context, deviceID, hash string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE devices SET command_secret_hash = ? WHERE id = ?`,
		hash, deviceID,
	)
	return err
}

// GetSecretHash retrieves the bcrypt hash of a device's command secret.
func (s *Store) GetSecretHash(ctx context.Context, deviceID string) (string, error) {
	var hash string
	err := s.db.QueryRowContext(ctx,
		`SELECT command_secret_hash FROM devices WHERE id = ?`,
		deviceID,
	).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return hash, err
}

// HashAllSecrets hashes all existing command secrets that don't have a hash.
// This is a migration helper for existing databases.
func (s *Store) HashAllSecrets(ctx context.Context) (int, error) {
	hasher := NewSecretHash()
	
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, command_secret FROM devices WHERE command_secret_hash IS NULL OR command_secret_hash = ''`,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, secret string
		if err := rows.Scan(&id, &secret); err != nil {
			continue
		}
		hash, err := hasher.HashSecret(secret)
		if err != nil {
			continue
		}
		if err := s.SetSecretHash(ctx, id, hash); err != nil {
			continue
		}
		count++
	}
	return count, rows.Err()
}

func (s *Store) SetOnline(ctx context.Context, id string, online bool) error {
	v := 0
	if online {
		v = 1
	}
	_, err := s.db.ExecContext(ctx, `UPDATE devices SET online=?, last_seen=? WHERE id=?`, v, time.Now().UnixMilli(), id)
	return err
}

func (s *Store) Touch(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE devices SET last_seen=? WHERE id=?`, time.Now().UnixMilli(), id)
	return err
}

func (s *Store) UpdateFCM(ctx context.Context, id, token string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE devices SET fcm_token=?, last_seen=? WHERE id=?`, token, time.Now().UnixMilli(), id)
	return err
}

func (s *Store) DeleteDevice(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM devices WHERE id=?`, id)
	return err
}

func (s *Store) SaveTelemetry(ctx context.Context, id string, raw []byte, t models.TelemetryFrame) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	_, err = tx.ExecContext(ctx, `INSERT INTO telemetry(device_id,received_at,payload,risk_score,buffer_level,thermal_temp) VALUES(?,?,?,?,?,?)`,
		id, time.Now().UnixMilli(), string(raw), t.RiskScore, t.BufferLevel, t.ThermalTemp)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE devices SET last_seen=? WHERE id=?`, time.Now().UnixMilli(), id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM telemetry WHERE id NOT IN (SELECT id FROM telemetry ORDER BY received_at DESC LIMIT 5000)`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) SaveCommand(ctx context.Context, dispatchID, deviceID, command string, args []byte, delivery string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO commands(dispatch_id,device_id,command,args,delivery,created_at) VALUES(?,?,?,?,?,?)`,
		dispatchID, deviceID, command, string(args), delivery, time.Now().UnixMilli())
	return err
}

func (s *Store) MarkWake(ctx context.Context, dispatchID string, errText string) error {
	wakeSent := 1
	if errText != "" {
		wakeSent = 0
	}
	_, err := s.db.ExecContext(ctx, `UPDATE commands SET wake_sent=?, wake_error=? WHERE dispatch_id=?`, wakeSent, errText, dispatchID)
	return err
}

func (s *Store) MarkDelivered(ctx context.Context, dispatchID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE commands SET delivery='sent', delivered_at=?, status='sent' WHERE dispatch_id=?`, time.Now().UnixMilli(), dispatchID)
	return err
}

// CommandStatus represents the status of a command dispatch.
type CommandStatus struct {
	DispatchID  string     `json:"dispatchId"`
	DeviceID    string     `json:"deviceId"`
	Command     string     `json:"command"`
	Args        string     `json:"args,omitempty"`
	Status      string     `json:"status"`
	Delivery    string     `json:"delivery"`
	CreatedAt   time.Time  `json:"createdAt"`
	DeliveredAt *time.Time `json:"deliveredAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	Result      string     `json:"result,omitempty"`
	WakeError   string     `json:"wakeError,omitempty"`
}

// GetCommandStatus retrieves the status of a command dispatch.
func (s *Store) GetCommandStatus(ctx context.Context, dispatchID string) (*CommandStatus, error) {
	var cs CommandStatus
	var deliveredAt, completedAt sql.NullInt64

	err := s.db.QueryRowContext(ctx, `
		SELECT dispatch_id, device_id, command, args, status, delivery,
		       created_at, delivered_at, completed_at, result, wake_error
		FROM commands WHERE dispatch_id = ?
	`, dispatchID).Scan(
		&cs.DispatchID, &cs.DeviceID, &cs.Command, &cs.Args,
		&cs.Status, &cs.Delivery, &cs.CreatedAt,
		&deliveredAt, &completedAt, &cs.Result, &cs.WakeError,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if deliveredAt.Valid {
		t := time.UnixMilli(deliveredAt.Int64).UTC()
		cs.DeliveredAt = &t
	}
	if completedAt.Valid {
		t := time.UnixMilli(completedAt.Int64).UTC()
		cs.CompletedAt = &t
	}

	return &cs, nil
}

// UpdateCommandStatus updates the status of a command.
func (s *Store) UpdateCommandStatus(ctx context.Context, dispatchID, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE commands SET status=? WHERE dispatch_id=?`, status, dispatchID)
	return err
}

// MarkCommandCompleted marks a command as completed with result.
func (s *Store) MarkCommandCompleted(ctx context.Context, dispatchID, result string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE commands SET status='completed', completed_at=?, result=? WHERE dispatch_id=?`,
		time.Now().UnixMilli(), result, dispatchID)
	return err
}

// MarkCommandFailed marks a command as failed with error.
func (s *Store) MarkCommandFailed(ctx context.Context, dispatchID, errMsg string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE commands SET status='failed', completed_at=?, result=? WHERE dispatch_id=?`,
		time.Now().UnixMilli(), errMsg, dispatchID)
	return err
}

func (s *Store) Devices(ctx context.Context) ([]struct {
	ID                string
	FirebaseInstallID string
	FCMToken          string
	AppVersion        string
	DeviceClass       string
	CommandSecret     string
	Online            bool
	RegisteredAt      time.Time
	LastSeen          time.Time
}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.QueryContext(ctx, `SELECT id, firebase_install_id, fcm_token, app_version, device_class, command_secret, online, registered_at, last_seen FROM devices ORDER BY last_seen DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ID                string
		FirebaseInstallID string
		FCMToken          string
		AppVersion        string
		DeviceClass       string
		CommandSecret     string
		Online            bool
		RegisteredAt      time.Time
		LastSeen          time.Time
	}
	for rows.Next() {
		var r deviceRow
		if err := rows.Scan(&r.ID, &r.FirebaseInstallID, &r.FCMToken, &r.AppVersion, &r.DeviceClass, &r.CommandSecret, &r.Online, &r.RegisteredAt, &r.LastSeen); err != nil {
			return nil, err
		}
		out = append(out, rowToDevice(r))
	}
	return out, rows.Err()
}

// DevicesPaginated returns devices with cursor-based pagination.
// The cursor is the lastSeen timestamp (in milliseconds) from the previous page.
// Results are ordered by last_seen DESC.
func (s *Store) DevicesPaginated(ctx context.Context, limit int, cursor int64) ([]struct {
	ID                string
	FirebaseInstallID string
	FCMToken          string
	AppVersion        string
	DeviceClass       string
	CommandSecret     string
	Online            bool
	RegisteredAt      time.Time
	LastSeen          time.Time
}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var rows *sql.Rows
	var err error

	if cursor > 0 {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, firebase_install_id, fcm_token, app_version, device_class, command_secret, online, registered_at, last_seen 
			FROM devices 
			WHERE last_seen < ? 
			ORDER BY last_seen DESC 
			LIMIT ?
		`, cursor, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, firebase_install_id, fcm_token, app_version, device_class, command_secret, online, registered_at, last_seen 
			FROM devices 
			ORDER BY last_seen DESC 
			LIMIT ?
		`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []struct {
		ID                string
		FirebaseInstallID string
		FCMToken          string
		AppVersion        string
		DeviceClass       string
		CommandSecret     string
		Online            bool
		RegisteredAt      time.Time
		LastSeen          time.Time
	}
	for rows.Next() {
		var r deviceRow
		if err := rows.Scan(&r.ID, &r.FirebaseInstallID, &r.FCMToken, &r.AppVersion, &r.DeviceClass, &r.CommandSecret, &r.Online, &r.RegisteredAt, &r.LastSeen); err != nil {
			return nil, err
		}
		out = append(out, rowToDevice(r))
	}
	return out, rows.Err()
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func NewDispatchID() string {
	s, _ := randomHex(16)
	if s == "" {
		return fmt.Sprintf("dispatch-%d", time.Now().UnixNano())
	}
	return s
}

// ─── Auth storage ───────────────────────────────────────────────────────────────

// OperatorCount returns the total number of operators in the system.
func (s *Store) OperatorCount(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM operators`).Scan(&n)
	return n, err
}

// GetOperatorByEmail retrieves an operator by email address.
func (s *Store) GetOperatorByEmail(ctx context.Context, email string) (*models.Operator, error) {
	var r struct {
		ID            string
		Email         string
		Name          string
		PasswordHash  []byte
		Role          string
		GoogleID      sql.NullString
		EmailVerified int
		CreatedAt     int64
		UpdatedAt     int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, role, google_id, COALESCE(email_verified, 0), created_at, updated_at
		 FROM operators WHERE email = ?`,
		email,
	).Scan(&r.ID, &r.Email, &r.Name, &r.PasswordHash, &r.Role, &r.GoogleID, &r.EmailVerified, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	op := &models.Operator{
		ID:            r.ID,
		Email:         r.Email,
		Name:          r.Name,
		PasswordHash:  string(r.PasswordHash),
		Role:          models.OperatorRole(r.Role),
		GoogleID:      r.GoogleID.String,
		EmailVerified: r.EmailVerified != 0,
		CreatedAt:     time.UnixMilli(r.CreatedAt).UTC(),
		UpdatedAt:     time.UnixMilli(r.UpdatedAt).UTC(),
	}
	return op, nil
}

// GetOperatorByGoogleID retrieves an operator by Google OAuth subject ID.
func (s *Store) GetOperatorByGoogleID(ctx context.Context, googleID string) (*models.Operator, error) {
	var r struct {
		ID            string
		Email         string
		Name          string
		PasswordHash  []byte
		Role          string
		GoogleID      string
		EmailVerified int
		CreatedAt     int64
		UpdatedAt     int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, role, google_id, COALESCE(email_verified, 0), created_at, updated_at
		 FROM operators WHERE google_id = ?`,
		googleID,
	).Scan(&r.ID, &r.Email, &r.Name, &r.PasswordHash, &r.Role, &r.GoogleID, &r.EmailVerified, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &models.Operator{
		ID:            r.ID,
		Email:         r.Email,
		Name:          r.Name,
		PasswordHash:  string(r.PasswordHash),
		Role:          models.OperatorRole(r.Role),
		GoogleID:      r.GoogleID,
		EmailVerified: r.EmailVerified != 0,
		CreatedAt:     time.UnixMilli(r.CreatedAt).UTC(),
		UpdatedAt:     time.UnixMilli(r.UpdatedAt).UTC(),
	}, nil
}

// GetOperatorByID retrieves an operator by their ID.
func (s *Store) GetOperatorByID(ctx context.Context, id string) (*models.Operator, error) {
	var r struct {
		ID            string
		Email         string
		Name          string
		PasswordHash  []byte
		Role          string
		GoogleID      sql.NullString
		EmailVerified int
		RiskWarn      int
		RiskCrit      int
		ThermalWarn   int
		ThermalCrit   int
		BufferWarn    int
		BufferCrit    int
		CreatedAt     int64
		UpdatedAt     int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, role, google_id, COALESCE(email_verified, 0),
		        COALESCE(risk_warn, 70), COALESCE(risk_crit, 90),
		        COALESCE(thermal_warn, 42), COALESCE(thermal_crit, 45),
		        COALESCE(buffer_warn, 60), COALESCE(buffer_crit, 80),
		        created_at, updated_at
		 FROM operators WHERE id = ?`,
		id,
	).Scan(&r.ID, &r.Email, &r.Name, &r.PasswordHash, &r.Role, &r.GoogleID, &r.EmailVerified,
		&r.RiskWarn, &r.RiskCrit, &r.ThermalWarn, &r.ThermalCrit, &r.BufferWarn, &r.BufferCrit,
		&r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &models.Operator{
		ID:            r.ID,
		Email:         r.Email,
		Name:          r.Name,
		PasswordHash:  string(r.PasswordHash),
		Role:          models.OperatorRole(r.Role),
		GoogleID:      r.GoogleID.String,
		EmailVerified: r.EmailVerified != 0,
		Thresholds: models.Thresholds{
			RiskWarn:    r.RiskWarn,
			RiskCrit:    r.RiskCrit,
			ThermalWarn: r.ThermalWarn,
			ThermalCrit: r.ThermalCrit,
			BufferWarn:  r.BufferWarn,
			BufferCrit:  r.BufferCrit,
		},
		CreatedAt: time.UnixMilli(r.CreatedAt).UTC(),
		UpdatedAt: time.UnixMilli(r.UpdatedAt).UTC(),
	}, nil
}

// CreateOperator inserts a new operator. role should be set to "super_admin" for the first operator.
func (s *Store) CreateOperator(ctx context.Context, op *models.Operator) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO operators(id, email, name, password_hash, role, google_id, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		op.ID, op.Email, op.Name, op.PasswordHash, string(op.Role), op.GoogleID, now.UnixMilli(), now.UnixMilli(),
	)
	return err
}

// UpdateOperatorGoogleID sets the google_id for an operator (after successful OAuth callback).
func (s *Store) UpdateOperatorGoogleID(ctx context.Context, operatorID, googleID string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE operators SET google_id = ?, updated_at = ? WHERE id = ?`,
		googleID, now.UnixMilli(), operatorID,
	)
	return err
}

// UpdateOperatorName updates the display name for an operator.
func (s *Store) UpdateOperatorName(ctx context.Context, operatorID, name string) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET name = ?, updated_at = ? WHERE id = ?`,
		strings.TrimSpace(name), now.UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// UpdateOperatorThresholds updates the alert thresholds for an operator.
func (s *Store) UpdateOperatorThresholds(ctx context.Context, operatorID string, th models.Thresholds) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET risk_warn=?, risk_crit=?, thermal_warn=?, thermal_crit=?, buffer_warn=?, buffer_crit=?, updated_at=? WHERE id=?`,
		th.RiskWarn, th.RiskCrit, th.ThermalWarn, th.ThermalCrit, th.BufferWarn, th.BufferCrit, now.UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// CreateSession inserts a new auth session.
func (s *Store) CreateSession(ctx context.Context, sess *models.Session) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO auth_sessions(id, operator_id, token_hash, expires_at, created_at, user_agent, ip_address)
		 VALUES(?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.OperatorID, sess.TokenHash, sess.ExpiresAt.UnixMilli(), sess.CreatedAt.UnixMilli(),
		sess.UserAgent, sess.IPAddress,
	)
	return err
}

// GetSessionByTokenHash retrieves a session by its token hash.
func (s *Store) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	var r struct {
		ID         string
		OperatorID string
		ExpiresAt  int64
		CreatedAt  int64
		UserAgent  sql.NullString
		IPAddress  sql.NullString
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, operator_id, expires_at, created_at, user_agent, ip_address
		 FROM auth_sessions WHERE token_hash = ?`,
		tokenHash,
	).Scan(&r.ID, &r.OperatorID, &r.ExpiresAt, &r.CreatedAt, &r.UserAgent, &r.IPAddress)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &models.Session{
		ID:         r.ID,
		OperatorID: r.OperatorID,
		ExpiresAt: time.UnixMilli(r.ExpiresAt).UTC(),
		CreatedAt: time.UnixMilli(r.CreatedAt).UTC(),
		UserAgent: r.UserAgent.String,
		IPAddress: r.IPAddress.String,
	}, nil
}

// DeleteSession removes a session by its token hash (logout).
func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash = ?`, tokenHash)
	return err
}

// DeleteExpiredSessions removes all sessions past their expiry time.
func (s *Store) DeleteExpiredSessions(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM auth_sessions WHERE expires_at < ?`,
		time.Now().UTC().UnixMilli(),
	)
	return err
}

// DeleteAllSessionsForOperator removes all sessions for a given operator (used on password change).
func (s *Store) DeleteAllSessionsForOperator(ctx context.Context, operatorID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE operator_id = ?`, operatorID)
	return err
}

// ─── Email Verification Tokens ─────────────────────────────────────────────────

// EmailVerification represents a pending email verification token.
type EmailVerification struct {
	ID         string
	OperatorID string
	TokenHash  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

// CreateEmailVerification inserts a new email verification token.
func (s *Store) CreateEmailVerification(ctx context.Context, ev *EmailVerification) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO email_verifications(id, operator_id, token_hash, expires_at, created_at)
		 VALUES(?, ?, ?, ?, ?)`,
		ev.ID, ev.OperatorID, ev.TokenHash, ev.ExpiresAt.UnixMilli(), ev.CreatedAt.UnixMilli(),
	)
	return err
}

// GetEmailVerificationByTokenHash retrieves an email verification by its token hash.
func (s *Store) GetEmailVerificationByTokenHash(ctx context.Context, tokenHash string) (*EmailVerification, error) {
	var r struct {
		ID         string
		OperatorID string
		TokenHash  string
		ExpiresAt  int64
		CreatedAt  int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, operator_id, token_hash, expires_at, created_at
		 FROM email_verifications WHERE token_hash = ?`,
		tokenHash,
	).Scan(&r.ID, &r.OperatorID, &r.TokenHash, &r.ExpiresAt, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &EmailVerification{
		ID:         r.ID,
		OperatorID: r.OperatorID,
		TokenHash:  r.TokenHash,
		ExpiresAt:  time.UnixMilli(r.ExpiresAt).UTC(),
		CreatedAt:  time.UnixMilli(r.CreatedAt).UTC(),
	}, nil
}

// DeleteEmailVerification removes an email verification by ID.
func (s *Store) DeleteEmailVerification(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM email_verifications WHERE id = ?`, id)
	return err
}

// DeleteEmailVerificationsByOperator removes all email verifications for an operator.
func (s *Store) DeleteEmailVerificationsByOperator(ctx context.Context, operatorID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM email_verifications WHERE operator_id = ?`, operatorID)
	return err
}

// ─── Password Reset Tokens ─────────────────────────────────────────────────────

// PasswordResetToken represents a pending password reset token.
type PasswordResetToken struct {
	ID         string
	OperatorID string
	TokenHash  string
	ExpiresAt  time.Time
	UsedAt     *time.Time
	CreatedAt  time.Time
}

// CreatePasswordResetToken inserts a new password reset token.
func (s *Store) CreatePasswordResetToken(ctx context.Context, prt *PasswordResetToken) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens(id, operator_id, token_hash, expires_at, used_at, created_at)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		prt.ID, prt.OperatorID, prt.TokenHash, prt.ExpiresAt.UnixMilli(), nil, prt.CreatedAt.UnixMilli(),
	)
	return err
}

// GetPasswordResetTokenByHash retrieves a password reset token by its hash.
func (s *Store) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	var r struct {
		ID         string
		OperatorID string
		TokenHash  string
		ExpiresAt  int64
		UsedAt     sql.NullInt64
		CreatedAt  int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, operator_id, token_hash, expires_at, used_at, created_at
		 FROM password_reset_tokens WHERE token_hash = ?`,
		tokenHash,
	).Scan(&r.ID, &r.OperatorID, &r.TokenHash, &r.ExpiresAt, &r.UsedAt, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var usedAt *time.Time
	if r.UsedAt.Valid {
		t := time.UnixMilli(r.UsedAt.Int64).UTC()
		usedAt = &t
	}
	return &PasswordResetToken{
		ID:         r.ID,
		OperatorID: r.OperatorID,
		TokenHash:  r.TokenHash,
		ExpiresAt:  time.UnixMilli(r.ExpiresAt).UTC(),
		UsedAt:     usedAt,
		CreatedAt:  time.UnixMilli(r.CreatedAt).UTC(),
	}, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used.
func (s *Store) MarkPasswordResetTokenUsed(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = ? WHERE id = ?`,
		time.Now().UTC().UnixMilli(), id,
	)
	return err
}

// DeletePasswordResetToken removes a password reset token by ID.
func (s *Store) DeletePasswordResetToken(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE id = ?`, id)
	return err
}

// DeletePasswordResetTokensByOperator removes all password reset tokens for an operator.
func (s *Store) DeletePasswordResetTokensByOperator(ctx context.Context, operatorID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE operator_id = ?`, operatorID)
	return err
}

// GetOperatorEmailVerified returns whether an operator has verified their email.
func (s *Store) GetOperatorEmailVerified(ctx context.Context, operatorID string) (bool, error) {
	var verified int
	err := s.db.QueryRowContext(ctx,
		`SELECT email_verified FROM operators WHERE id = ?`,
		operatorID,
	).Scan(&verified)
	if errors.Is(err, sql.ErrNoRows) {
		return false, errors.New("operator not found")
	}
	return verified != 0, err
}

// SetOperatorEmailVerified marks an operator's email as verified.
func (s *Store) SetOperatorEmailVerified(ctx context.Context, operatorID string, verified bool) error {
	v := 0
	if verified {
		v = 1
	}
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET email_verified = ?, updated_at = ? WHERE id = ?`,
		v, time.Now().UTC().UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// UpdateOperatorPassword updates the password hash for an operator.
func (s *Store) UpdateOperatorPassword(ctx context.Context, operatorID, passwordHash string) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET password_hash = ?, updated_at = ? WHERE id = ?`,
		passwordHash, now.UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// ─── System Settings ─────────────────────────────────────────────────────────────

// GetSetting retrieves a setting value by key.
func (s *Store) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM settings WHERE key = ?`, key,
	).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value, err
}

// SetSetting updates or inserts a setting value.
func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO settings(key, value, updated_at) VALUES(?, ?, ?)`,
		key, value, time.Now().UTC().UnixMilli(),
	)
	return err
}

// GetEnforceHMAC returns whether HMAC enforcement is enabled.
func (s *Store) GetEnforceHMAC(ctx context.Context) (bool, error) {
	val, err := s.GetSetting(ctx, "enforce_hmac")
	if err != nil || val == "" {
		return false, err
	}
	return val == "true" || val == "1", nil
}

// SetEnforceHMAC updates the HMAC enforcement setting.
func (s *Store) SetEnforceHMAC(ctx context.Context, enforce bool) error {
	val := "false"
	if enforce {
		val = "true"
	}
	return s.SetSetting(ctx, "enforce_hmac", val)
}

// GetHMACWindowSeconds returns the HMAC timestamp window in seconds.
func (s *Store) GetHMACWindowSeconds(ctx context.Context) (int, error) {
	val, err := s.GetSetting(ctx, "hmac_window_seconds")
	if err != nil || val == "" {
		return 30, nil // default 30 seconds per COMMAND_SECURITY.md
	}
	var seconds int
	_, err = fmt.Sscanf(val, "%d", &seconds)
	if err != nil {
		return 30, nil
	}
	return seconds, nil
}

// SetHMACWindowSeconds updates the HMAC timestamp window.
func (s *Store) SetHMACWindowSeconds(ctx context.Context, seconds int) error {
	return s.SetSetting(ctx, "hmac_window_seconds", fmt.Sprintf("%d", seconds))
}
