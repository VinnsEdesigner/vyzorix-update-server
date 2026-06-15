// Package storage provides SQLite database operations.
package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

// deviceRow represents a device row from the database.
type deviceRow struct {
	ID                string
	FirebaseInstallID  string
	FCMToken          string
	AppVersion        string
	DeviceClass       string
	CommandSecret     string
	Online            int
	RegisteredAt      int64
	LastSeen          int64
}

// Device represents a device's data in a clean struct.
type Device struct {
	RegisteredAt       time.Time
	LastSeen           time.Time
	ID                 string
	FirebaseInstallID   string
	FCMToken           string
	AppVersion         string
	DeviceClass        string
	CommandSecret      string
	Online             bool
}

// rowToDevice converts a deviceRow to a Device struct.
func rowToDevice(r deviceRow) Device {
	return Device{
		ID:                r.ID,
		FirebaseInstallID: r.FirebaseInstallID,
		FCMToken:         r.FCMToken,
		AppVersion:       r.AppVersion,
		DeviceClass:      r.DeviceClass,
		CommandSecret:    r.CommandSecret,
		Online:           r.Online != 0,
		RegisteredAt:      time.UnixMilli(r.RegisteredAt).UTC(),
		LastSeen:         time.UnixMilli(r.LastSeen).UTC(),
	}
}

// Register registers a new device or returns existing device info if already registered.
// Returns (device, isExisting, error).
func (s *Store) Register(ctx context.Context, req models.RegisterRequest) (Device, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()

	var existingID, existingFID string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, firebase_install_id FROM devices WHERE id = ?`,
		req.DeviceID,
	).Scan(&existingID, &existingFID)

	if err == nil {
		// Device exists
		if existingFID != req.FirebaseInstallID {
			return Device{}, false, ErrHijack
		}
		_, err = s.db.ExecContext(ctx,
			`UPDATE devices SET fcm_token=?, app_version=?, device_class=?, last_seen=? WHERE id=?`,
			req.FCMToken, req.AppVersion, req.DeviceClass, now.UnixMilli(), req.DeviceID,
		)
		if err != nil {
			return Device{}, false, err
		}
		var cmdSecret string
		var regAt int64
		if err := s.db.QueryRowContext(ctx,
			`SELECT command_secret, registered_at FROM devices WHERE id = ?`,
			req.DeviceID,
		).Scan(&cmdSecret, &regAt); err != nil {
			cmdSecret, regAt = "", 0
		}
		return Device{
			ID:                 req.DeviceID,
			FirebaseInstallID:   existingFID,
			FCMToken:          req.FCMToken,
			AppVersion:        req.AppVersion,
			DeviceClass:       req.DeviceClass,
			CommandSecret:     cmdSecret,
			Online:            true,
			RegisteredAt:      time.UnixMilli(regAt),
			LastSeen:          now,
		}, false, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return Device{}, false, err
	}

	// New device - generate command secret
	secret, err := randomHex(32)
	if err != nil {
		return Device{}, false, err
	}

	// Hash the secret for audit/compliance (using Argon2id)
	secretHash, hashErr := HashSecret(secret)
	if hashErr != nil {
		secretHash = ""
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO devices(id,firebase_install_id,fcm_token,app_version,device_class,command_secret,command_secret_hash,online,registered_at,last_seen) VALUES(?,?,?,?,?,?,?,0,?,?)`,
		req.DeviceID, req.FirebaseInstallID, req.FCMToken, req.AppVersion, req.DeviceClass, secret, secretHash, now.UnixMilli(), now.UnixMilli(),
	)
	if err != nil {
		return Device{}, false, err
	}

	return Device{
		ID:                 req.DeviceID,
		FirebaseInstallID:   req.FirebaseInstallID,
		FCMToken:          req.FCMToken,
		AppVersion:        req.AppVersion,
		DeviceClass:       req.DeviceClass,
		CommandSecret:     secret,
		Online:            false,
		RegisteredAt:      now,
		LastSeen:          now,
	}, true, nil
}

// GetDevice retrieves a device by ID.
func (s *Store) GetDevice(ctx context.Context, id string) (Device, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var r deviceRow
	err := s.db.QueryRowContext(ctx,
		`SELECT id, firebase_install_id, fcm_token, app_version, device_class, command_secret, online, registered_at, last_seen FROM devices WHERE id = ?`,
		id,
	).Scan(&r.ID, &r.FirebaseInstallID, &r.FCMToken, &r.AppVersion, &r.DeviceClass, &r.CommandSecret, &r.Online, &r.RegisteredAt, &r.LastSeen)

	if errors.Is(err, sql.ErrNoRows) {
		return Device{}, false, nil
	}
	if err != nil {
		return Device{}, false, err
	}
	return rowToDevice(r), true, nil
}

// Secret retrieves the command secret for a device.
func (s *Store) Secret(ctx context.Context, id string) (string, bool) {
	d, ok, err := s.GetDevice(ctx, id)
	if err != nil || !ok {
		return "", false
	}
	return d.CommandSecret, true
}

// SetOnline updates the online status of a device.
func (s *Store) SetOnline(ctx context.Context, id string, online bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v := 0
	if online {
		v = 1
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE devices SET online=?, last_seen=? WHERE id=?`,
		v, time.Now().UnixMilli(), id,
	)
	return err
}

// Touch updates the last_seen timestamp of a device.
func (s *Store) Touch(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx,
		`UPDATE devices SET last_seen=? WHERE id=?`,
		time.Now().UnixMilli(), id,
	)
	return err
}

// UpdateFCM updates the FCM token for a device.
func (s *Store) UpdateFCM(ctx context.Context, id, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx,
		`UPDATE devices SET fcm_token=?, last_seen=? WHERE id=?`,
		token, time.Now().UnixMilli(), id,
	)
	return err
}

// DeleteDevice removes a device from the database.
func (s *Store) DeleteDevice(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM devices WHERE id=?`, id)
	return err
}

// ListDevices retrieves all devices ordered by last_seen.
func (s *Store) ListDevices(ctx context.Context) ([]Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, firebase_install_id, fcm_token, app_version, device_class, command_secret, online, registered_at, last_seen FROM devices ORDER BY last_seen DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var out []Device
	for rows.Next() {
		var r deviceRow
		if err := rows.Scan(&r.ID, &r.FirebaseInstallID, &r.FCMToken, &r.AppVersion, &r.DeviceClass, &r.CommandSecret, &r.Online, &r.RegisteredAt, &r.LastSeen); err != nil {
			return nil, err
		}
		out = append(out, rowToDevice(r))
	}
	return out, rows.Err()
}

// ListDevicesPaginated returns devices with cursor-based pagination.
// The cursor is the lastSeen timestamp (in milliseconds) from the previous page.
func (s *Store) ListDevicesPaginated(ctx context.Context, limit int, cursor int64) ([]Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var rows *sql.Rows
	var err error

	if cursor > 0 {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, firebase_install_id, fcm_token, app_version, device_class, command_secret, online, registered_at, last_seen 
			 FROM devices WHERE last_seen < ? ORDER BY last_seen DESC LIMIT ?`,
			cursor, limit,
		)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, firebase_install_id, fcm_token, app_version, device_class, command_secret, online, registered_at, last_seen 
			 FROM devices ORDER BY last_seen DESC LIMIT ?`,
			limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var out []Device
	for rows.Next() {
		var r deviceRow
		if err := rows.Scan(&r.ID, &r.FirebaseInstallID, &r.FCMToken, &r.AppVersion, &r.DeviceClass, &r.CommandSecret, &r.Online, &r.RegisteredAt, &r.LastSeen); err != nil {
			return nil, err
		}
		out = append(out, rowToDevice(r))
	}
	return out, rows.Err()
}

// SetSecretHash stores the hash of a device's command secret.
func (s *Store) SetSecretHash(ctx context.Context, deviceID, hash string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE devices SET command_secret_hash = ? WHERE id = ?`,
		hash, deviceID,
	)
	return err
}

// GetSecretHash retrieves the hash of a device's command secret.
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
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, command_secret FROM devices WHERE command_secret_hash IS NULL OR command_secret_hash = ''`,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close() //nolint:errcheck

	count := 0
	for rows.Next() {
		var id, secret string
		if err := rows.Scan(&id, &secret); err != nil {
			continue
		}
		hash, err := HashSecret(secret)
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

// randomHex generates a random hex string of n bytes.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// NewDispatchID generates a new dispatch ID.
func NewDispatchID() string {
	s, err := randomHex(16)
	if err != nil || s == "" {
		return fmt.Sprintf("dispatch-%d", time.Now().UnixNano())
	}
	return s
}

// ─── Backward Compatibility Aliases ─────────────────────────────────────────────
// These methods provide backward-compatible interfaces for existing code.

// Device is a backward-compatible alias for GetDevice.
func (s *Store) Device(ctx context.Context, id string) (Device, bool, error) {
	return s.GetDevice(ctx, id)
}

// Devices is a backward-compatible alias for ListDevices.
func (s *Store) Devices(ctx context.Context) ([]Device, error) {
	return s.ListDevices(ctx)
}

// DevicesPaginated is a backward-compatible alias for ListDevicesPaginated.
func (s *Store) DevicesPaginated(ctx context.Context, limit int, cursor int64) ([]Device, error) {
	return s.ListDevicesPaginated(ctx, limit, cursor)
}