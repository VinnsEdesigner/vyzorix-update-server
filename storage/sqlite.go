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
	return nil
}

func (s *Store) Register(ctx context.Context, req struct {
	DeviceID          string
	FirebaseInstallID string
	FCMToken          string
	AppVersion        string
	DeviceClass       string
}) (struct {
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
		}{ID: req.DeviceID, FirebaseInstallID: existingFID, FCMToken: req.FCMToken, AppVersion: req.AppVersion, DeviceClass: req.DeviceClass, LastSeen: now}, false, nil
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
	_, err = s.db.ExecContext(ctx, `INSERT INTO devices(id,firebase_install_id,fcm_token,app_version,device_class,command_secret,online,registered_at,last_seen) VALUES(?,?,?,?,?,?,0,?,?)`,
		req.DeviceID, req.FirebaseInstallID, req.FCMToken, req.AppVersion, req.DeviceClass, secret, now.UnixMilli(), now.UnixMilli())
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
	_, err := s.db.ExecContext(ctx, `UPDATE commands SET delivery='sent', delivered_at=? WHERE dispatch_id=?`, time.Now().UnixMilli(), dispatchID)
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
		ID           string
		Email        string
		Name         string
		PasswordHash []byte
		Role         string
		GoogleID     sql.NullString
		CreatedAt    int64
		UpdatedAt    int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, role, google_id, created_at, updated_at
		 FROM operators WHERE email = ?`,
		email,
	).Scan(&r.ID, &r.Email, &r.Name, &r.PasswordHash, &r.Role, &r.GoogleID, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	op := &models.Operator{
		ID:           r.ID,
		Email:        r.Email,
		Name:         r.Name,
		PasswordHash: string(r.PasswordHash),
		Role:         models.OperatorRole(r.Role),
		GoogleID:     r.GoogleID.String,
		CreatedAt:    time.UnixMilli(r.CreatedAt).UTC(),
		UpdatedAt:    time.UnixMilli(r.UpdatedAt).UTC(),
	}
	return op, nil
}

// GetOperatorByGoogleID retrieves an operator by Google OAuth subject ID.
func (s *Store) GetOperatorByGoogleID(ctx context.Context, googleID string) (*models.Operator, error) {
	var r struct {
		ID           string
		Email        string
		Name         string
		PasswordHash []byte
		Role         string
		GoogleID     string
		CreatedAt    int64
		UpdatedAt    int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, role, google_id, created_at, updated_at
		 FROM operators WHERE google_id = ?`,
		googleID,
	).Scan(&r.ID, &r.Email, &r.Name, &r.PasswordHash, &r.Role, &r.GoogleID, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &models.Operator{
		ID:           r.ID,
		Email:        r.Email,
		Name:         r.Name,
		PasswordHash: string(r.PasswordHash),
		Role:         models.OperatorRole(r.Role),
		GoogleID:     r.GoogleID,
		CreatedAt:    time.UnixMilli(r.CreatedAt).UTC(),
		UpdatedAt:    time.UnixMilli(r.UpdatedAt).UTC(),
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
