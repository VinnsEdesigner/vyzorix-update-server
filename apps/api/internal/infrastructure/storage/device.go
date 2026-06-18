package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
)

// Ensure DeviceRepository implements device.Repository.
var _ device.Repository = (*DeviceRepository)(nil)

// DeviceRepository implements device.Repository using SQLite.
type DeviceRepository struct {
	db *sql.DB
}

// NewDeviceRepository creates a new DeviceRepository.
func NewDeviceRepository(db *sql.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// FindByID retrieves a device by ID.
func (r *DeviceRepository) FindByID(ctx context.Context, id string) (*device.Device, error) {
	query := `
		SELECT id, firebase_install_id, fcm_token, app_version, device_class,
		       command_secret_hash, online, registered_at, last_seen, operator_id,
		       created_at, updated_at
		FROM devices WHERE id = ?`

	var d device.Device
	var fcmToken, operatorID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.FirebaseInstallID, &fcmToken, &d.AppVersion, &d.DeviceClass,
		&d.CommandSecretHash, &d.Online, &d.RegisteredAt, &d.LastSeen, &operatorID,
		&d.CreatedAt, &d.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, device.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	d.FCMToken = fcmToken.String
	d.OperatorID = operatorID.String

	return &d, nil
}

// FindByFirebaseInstallID retrieves a device by Firebase install ID.
func (r *DeviceRepository) FindByFirebaseInstallID(ctx context.Context, fid string) (*device.Device, error) {
	query := `
		SELECT id, firebase_install_id, fcm_token, app_version, device_class,
		       command_secret_hash, online, registered_at, last_seen, operator_id,
		       created_at, updated_at
		FROM devices WHERE firebase_install_id = ?`

	var d device.Device
	var fcmToken, operatorID sql.NullString

	err := r.db.QueryRowContext(ctx, query, fid).Scan(
		&d.ID, &d.FirebaseInstallID, &fcmToken, &d.AppVersion, &d.DeviceClass,
		&d.CommandSecretHash, &d.Online, &d.RegisteredAt, &d.LastSeen, &operatorID,
		&d.CreatedAt, &d.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, device.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	d.FCMToken = fcmToken.String
	d.OperatorID = operatorID.String

	return &d, nil
}

// Create creates a new device.
func (r *DeviceRepository) Create(ctx context.Context, d *device.Device) error {
	query := `
		INSERT INTO devices (id, firebase_install_id, fcm_token, app_version, device_class,
		                    command_secret_hash, online, registered_at, last_seen, operator_id,
		                    created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.FirebaseInstallID, nullString(d.FCMToken), d.AppVersion, d.DeviceClass,
		d.CommandSecretHash, d.Online, d.RegisteredAt, d.LastSeen, nullString(d.OperatorID),
		d.CreatedAt, d.UpdatedAt,
	)

	return err
}

// Update updates an existing device.
func (r *DeviceRepository) Update(ctx context.Context, d *device.Device) error {
	query := `
		UPDATE devices 
		SET firebase_install_id = ?, fcm_token = ?, app_version = ?, device_class = ?,
		    command_secret_hash = ?, online = ?, registered_at = ?, last_seen = ?,
		    operator_id = ?, updated_at = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		d.FirebaseInstallID, nullString(d.FCMToken), d.AppVersion, d.DeviceClass,
		d.CommandSecretHash, d.Online, d.RegisteredAt, d.LastSeen, nullString(d.OperatorID),
		time.Now(), d.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// Delete deletes a device.
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM devices WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// UpdateFCMToken updates the FCM token for a device.
func (r *DeviceRepository) UpdateFCMToken(ctx context.Context, id, fcmToken string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE devices SET fcm_token = ?, updated_at = ? WHERE id = ?",
		fcmToken, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// SetOnline sets the online status of a device.
func (r *DeviceRepository) SetOnline(ctx context.Context, id string, online bool) error {
	now := time.Now().UnixMilli()
	var result sql.Result
	var err error

	if online {
		result, err = r.db.ExecContext(ctx,
			"UPDATE devices SET online = ?, last_seen = ?, updated_at = ? WHERE id = ?",
			online, now, now, id,
		)
	} else {
		result, err = r.db.ExecContext(ctx,
			"UPDATE devices SET online = ?, updated_at = ? WHERE id = ?",
			online, now, id,
		)
	}

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// UpdateLastSeen updates the last seen timestamp.
func (r *DeviceRepository) UpdateLastSeen(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	result, err := r.db.ExecContext(ctx,
		"UPDATE devices SET last_seen = ?, updated_at = ? WHERE id = ?",
		now, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// List returns a paginated list of devices.
func (r *DeviceRepository) List(ctx context.Context, limit, offset int) ([]*device.Device, int, error) {
	query := `
		SELECT id, firebase_install_id, fcm_token, app_version, device_class,
		       command_secret_hash, online, registered_at, last_seen, operator_id,
		       created_at, updated_at
		FROM devices ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var devices []*device.Device
	for rows.Next() {
		var d device.Device
		var fcmToken, operatorID sql.NullString

		if err := rows.Scan(
			&d.ID, &d.FirebaseInstallID, &fcmToken, &d.AppVersion, &d.DeviceClass,
			&d.CommandSecretHash, &d.Online, &d.RegisteredAt, &d.LastSeen, &operatorID,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		d.FCMToken = fcmToken.String
		d.OperatorID = operatorID.String
		devices = append(devices, &d)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices").Scan(&total); err != nil {
		return nil, 0, err
	}

	return devices, total, rows.Err()
}

// ListByOperator returns all devices for an operator.
func (r *DeviceRepository) ListByOperator(ctx context.Context, operatorID string) ([]*device.Device, error) {
	query := `
		SELECT id, firebase_install_id, fcm_token, app_version, device_class,
		       command_secret_hash, online, registered_at, last_seen, operator_id,
		       created_at, updated_at
		FROM devices WHERE operator_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, operatorID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var devices []*device.Device
	for rows.Next() {
		var d device.Device
		var fcmToken, opID sql.NullString

		if err := rows.Scan(
			&d.ID, &d.FirebaseInstallID, &fcmToken, &d.AppVersion, &d.DeviceClass,
			&d.CommandSecretHash, &d.Online, &d.RegisteredAt, &d.LastSeen, &opID,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}

		d.FCMToken = fcmToken.String
		d.OperatorID = opID.String
		devices = append(devices, &d)
	}

	return devices, rows.Err()
}

// Count returns the total number of devices.
func (r *DeviceRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices").Scan(&count)
	return count, err
}

// CountByOperator returns the number of devices for an operator.
func (r *DeviceRepository) CountByOperator(ctx context.Context, operatorID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM devices WHERE operator_id = ?", operatorID,
	).Scan(&count)
	return count, err
}
