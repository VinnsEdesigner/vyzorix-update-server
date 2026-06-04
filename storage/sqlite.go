package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix-update-server/models"
)

var ErrHijack = errors.New("device_id already registered to a different firebaseInstallId")

type Store struct {
	mu   sync.Mutex
	path string
}

func Open(path string) (*Store, error) {
	if _, err := exec.LookPath("sqlite3"); err != nil {
		return nil, fmt.Errorf("sqlite3 CLI is required: %w", err)
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	s := &Store{path: path}
	return s, s.migrate(context.Background())
}
func (s *Store) Close() error                   { return nil }
func (s *Store) Ping(ctx context.Context) error { _, err := s.exec(ctx, `SELECT 1;`); return err }
func (s *Store) migrate(ctx context.Context) error {
	_, err := s.exec(ctx, `PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; CREATE TABLE IF NOT EXISTS devices (id TEXT PRIMARY KEY, firebase_install_id TEXT NOT NULL, fcm_token TEXT, app_version TEXT, device_class TEXT, command_secret TEXT NOT NULL, online INTEGER NOT NULL DEFAULT 0, registered_at INTEGER NOT NULL, last_seen INTEGER NOT NULL); CREATE TABLE IF NOT EXISTS telemetry (id INTEGER PRIMARY KEY AUTOINCREMENT, device_id TEXT NOT NULL, received_at INTEGER NOT NULL, payload TEXT NOT NULL, risk_score INTEGER, buffer_level INTEGER, thermal_temp REAL, FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE); CREATE TABLE IF NOT EXISTS commands (dispatch_id TEXT PRIMARY KEY, device_id TEXT NOT NULL, command TEXT NOT NULL, args TEXT, delivery TEXT NOT NULL, created_at INTEGER NOT NULL, delivered_at INTEGER, FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE); CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry(device_id, received_at DESC); CREATE INDEX IF NOT EXISTS idx_commands_device_time ON commands(device_id, created_at DESC);`)
	return err
}

func (s *Store) Register(ctx context.Context, req models.RegisterRequest) (models.Device, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	existing, ok, err := s.deviceLocked(ctx, req.DeviceID)
	if err != nil {
		return models.Device{}, false, err
	}
	if ok {
		if existing.FirebaseInstallID != req.FirebaseInstallID {
			return models.Device{}, false, ErrHijack
		}
		_, err = s.exec(ctx, fmt.Sprintf(`UPDATE devices SET fcm_token=%s, app_version=%s, device_class=%s, last_seen=%d WHERE id=%s;`, q(req.FCMToken), q(req.AppVersion), q(req.DeviceClass), now.UnixMilli(), q(req.DeviceID)))
		if err != nil {
			return models.Device{}, false, err
		}
		existing.FCMToken = req.FCMToken
		existing.AppVersion = req.AppVersion
		existing.DeviceClass = req.DeviceClass
		existing.LastSeen = now
		return existing, false, nil
	}
	secret, err := randomHex(32)
	if err != nil {
		return models.Device{}, false, err
	}
	d := models.Device{ID: req.DeviceID, FirebaseInstallID: req.FirebaseInstallID, FCMToken: req.FCMToken, AppVersion: req.AppVersion, DeviceClass: req.DeviceClass, CommandSecret: secret, RegisteredAt: now, LastSeen: now}
	_, err = s.exec(ctx, fmt.Sprintf(`INSERT INTO devices(id,firebase_install_id,fcm_token,app_version,device_class,command_secret,online,registered_at,last_seen) VALUES(%s,%s,%s,%s,%s,%s,0,%d,%d);`, q(d.ID), q(d.FirebaseInstallID), q(d.FCMToken), q(d.AppVersion), q(d.DeviceClass), q(d.CommandSecret), d.RegisteredAt.UnixMilli(), d.LastSeen.UnixMilli()))
	return d, true, err
}
func (s *Store) Device(ctx context.Context, id string) (models.Device, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deviceLocked(ctx, id)
}
func (s *Store) deviceLocked(ctx context.Context, id string) (models.Device, bool, error) {
	rows, err := s.queryDevices(ctx, fmt.Sprintf(`SELECT id AS id, firebase_install_id AS firebaseInstallId, fcm_token AS fcmToken, app_version AS appVersion, device_class AS deviceClass, command_secret AS commandSecret, online AS online, registered_at AS registeredAt, last_seen AS lastSeen FROM devices WHERE id=%s;`, q(id)))
	if err != nil {
		return models.Device{}, false, err
	}
	if len(rows) == 0 {
		return models.Device{}, false, nil
	}
	return rowToDevice(rows[0]), true, nil
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
	return s.execOnly(ctx, fmt.Sprintf(`UPDATE devices SET online=%d,last_seen=%d WHERE id=%s;`, v, time.Now().UnixMilli(), q(id)))
}
func (s *Store) Touch(ctx context.Context, id string) error {
	return s.execOnly(ctx, fmt.Sprintf(`UPDATE devices SET last_seen=%d WHERE id=%s;`, time.Now().UnixMilli(), q(id)))
}
func (s *Store) UpdateFCM(ctx context.Context, id, token string) error {
	return s.execOnly(ctx, fmt.Sprintf(`UPDATE devices SET fcm_token=%s,last_seen=%d WHERE id=%s;`, q(token), time.Now().UnixMilli(), q(id)))
}
func (s *Store) DeleteDevice(ctx context.Context, id string) error {
	return s.execOnly(ctx, fmt.Sprintf(`DELETE FROM devices WHERE id=%s;`, q(id)))
}
func (s *Store) SaveTelemetry(ctx context.Context, id string, raw []byte, t models.TelemetryFrame) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.exec(ctx, fmt.Sprintf(`INSERT INTO telemetry(device_id,received_at,payload,risk_score,buffer_level,thermal_temp) VALUES(%s,%d,%s,%d,%d,%f); UPDATE devices SET last_seen=%d WHERE id=%s; DELETE FROM telemetry WHERE id NOT IN (SELECT id FROM telemetry ORDER BY received_at DESC LIMIT 5000);`, q(id), time.Now().UnixMilli(), q(string(raw)), t.RiskScore, t.BufferLevel, t.ThermalTemp, time.Now().UnixMilli(), q(id)))
	return err
}
func (s *Store) SaveCommand(ctx context.Context, dispatchID, deviceID, command string, args []byte, delivery string) error {
	return s.execOnly(ctx, fmt.Sprintf(`INSERT INTO commands(dispatch_id,device_id,command,args,delivery,created_at) VALUES(%s,%s,%s,%s,%s,%d);`, q(dispatchID), q(deviceID), q(command), q(string(args)), q(delivery), time.Now().UnixMilli()))
}
func (s *Store) MarkDelivered(ctx context.Context, dispatchID string) error {
	return s.execOnly(ctx, fmt.Sprintf(`UPDATE commands SET delivery='sent',delivered_at=%d WHERE dispatch_id=%s;`, time.Now().UnixMilli(), q(dispatchID)))
}
func (s *Store) Devices(ctx context.Context) ([]models.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.queryDevices(ctx, `SELECT id AS id, firebase_install_id AS firebaseInstallId, fcm_token AS fcmToken, app_version AS appVersion, device_class AS deviceClass, command_secret AS commandSecret, online AS online, registered_at AS registeredAt, last_seen AS lastSeen FROM devices ORDER BY last_seen DESC;`)
	if err != nil {
		return nil, err
	}
	out := make([]models.Device, 0, len(rows))
	for _, row := range rows {
		out = append(out, rowToDevice(row))
	}
	return out, nil
}

type deviceRow struct {
	ID                string `json:"id"`
	FirebaseInstallID string `json:"firebaseInstallId"`
	FCMToken          string `json:"fcmToken"`
	AppVersion        string `json:"appVersion"`
	DeviceClass       string `json:"deviceClass"`
	CommandSecret     string `json:"commandSecret"`
	Online            int    `json:"online"`
	RegisteredAt      int64  `json:"registeredAt"`
	LastSeen          int64  `json:"lastSeen"`
}

func rowToDevice(r deviceRow) models.Device {
	return models.Device{ID: r.ID, FirebaseInstallID: r.FirebaseInstallID, FCMToken: r.FCMToken, AppVersion: r.AppVersion, DeviceClass: r.DeviceClass, CommandSecret: r.CommandSecret, Online: r.Online != 0, RegisteredAt: time.UnixMilli(r.RegisteredAt).UTC(), LastSeen: time.UnixMilli(r.LastSeen).UTC()}
}
func (s *Store) queryDevices(ctx context.Context, sql string) ([]deviceRow, error) {
	out, err := s.execArgs(ctx, []string{"-json", s.path}, sql)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var rows []deviceRow
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}
func (s *Store) execOnly(ctx context.Context, sql string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.exec(ctx, sql)
	return err
}
func (s *Store) exec(ctx context.Context, sql string) (string, error) {
	return s.execArgs(ctx, []string{s.path}, sql)
}
func (s *Store) execArgs(ctx context.Context, args []string, sql string) (string, error) {
	cmd := exec.CommandContext(ctx, "sqlite3", args...)
	cmd.Stdin = strings.NewReader(sql)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return string(b), fmt.Errorf("sqlite3: %w: %s", err, strings.TrimSpace(string(b)))
	}
	return string(b), nil
}
func q(v string) string { return "'" + strings.ReplaceAll(v, "'", "''") + "'" }
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
