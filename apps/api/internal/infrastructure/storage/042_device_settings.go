package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
)

// DeviceSettingsRepository implements device.DeviceSettingsRepository.
type DeviceSettingsRepository struct {
	db *sql.DB
}

// NewDeviceSettingsRepository creates a new DeviceSettingsRepository.
func NewDeviceSettingsRepository(db *sql.DB) *DeviceSettingsRepository {
	return &DeviceSettingsRepository{db: db}
}

// migrateDeviceSettings creates the device_settings table.
func migrateDeviceSettings(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS device_settings (
			id TEXT PRIMARY KEY,
			device_imei TEXT NOT NULL UNIQUE,
			custom_name TEXT,
			location TEXT,
			metadata JSONB,
			thresholds JSONB,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			FOREIGN KEY (device_imei) REFERENCES devices(imei) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create index for efficient lookups.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_device_settings_imei ON device_settings(device_imei)
	`)

	return err
}

// Create creates new device settings.
func (r *DeviceSettingsRepository) Create(ctx context.Context, settings *device.DeviceSettings) error {
	var metadataJSON, thresholdsJSON []byte
	var err error

	if settings.Metadata != nil {
		metadataJSON, err = json.Marshal(settings.Metadata)
		if err != nil {
			return err
		}
	}

	if settings.Thresholds != nil {
		thresholdsJSON, err = json.Marshal(settings.Thresholds)
		if err != nil {
			return err
		}
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO device_settings (id, device_imei, custom_name, location, metadata, thresholds, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, settings.ID, settings.DeviceIMEI, settings.CustomName, settings.Location, metadataJSON, thresholdsJSON, settings.CreatedAt.Unix(), settings.UpdatedAt.Unix())

	return err
}

// FindByID retrieves settings by ID.
func (r *DeviceSettingsRepository) FindByID(ctx context.Context, id string) (*device.DeviceSettings, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, device_imei, custom_name, location, metadata, thresholds, created_at, updated_at
		FROM device_settings
		WHERE id = ?
	`, id)

	return r.scanSettings(row)
}

// FindByDeviceIMEI retrieves settings by device IMEI.
func (r *DeviceSettingsRepository) FindByDeviceIMEI(ctx context.Context, imei string) (*device.DeviceSettings, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, device_imei, custom_name, location, metadata, thresholds, created_at, updated_at
		FROM device_settings
		WHERE device_imei = ?
	`, imei)

	return r.scanSettings(row)
}

// Update updates device settings.
func (r *DeviceSettingsRepository) Update(ctx context.Context, settings *device.DeviceSettings) error {
	var metadataJSON, thresholdsJSON []byte
	var err error

	if settings.Metadata != nil {
		metadataJSON, err = json.Marshal(settings.Metadata)
		if err != nil {
			return err
		}
	}

	if settings.Thresholds != nil {
		thresholdsJSON, err = json.Marshal(settings.Thresholds)
		if err != nil {
			return err
		}
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE device_settings
		SET custom_name = ?, location = ?, metadata = ?, thresholds = ?, updated_at = ?
		WHERE id = ?
	`, settings.CustomName, settings.Location, metadataJSON, thresholdsJSON, time.Now().Unix(), settings.ID)

	return err
}

// Delete deletes device settings by ID.
func (r *DeviceSettingsRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM device_settings WHERE id = ?`, id)
	return err
}

// DeleteByDeviceIMEI deletes settings by device IMEI.
func (r *DeviceSettingsRepository) DeleteByDeviceIMEI(ctx context.Context, imei string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM device_settings WHERE device_imei = ?`, imei)
	return err
}

// scanSettings scans a row into DeviceSettings.
func (r *DeviceSettingsRepository) scanSettings(row *sql.Row) (*device.DeviceSettings, error) {
	var id, deviceIMEI, customName, location string
	var metadataJSON, thresholdsJSON []byte
	var createdAt, updatedAt int64

	err := row.Scan(&id, &deviceIMEI, &customName, &location, &metadataJSON, &thresholdsJSON, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, device.ErrSettingsNotFound
		}
		return nil, err
	}

	settings := &device.DeviceSettings{
		ID:         id,
		DeviceIMEI: deviceIMEI,
		CustomName: customName,
		Location:   location,
		CreatedAt:  time.Unix(createdAt, 0),
		UpdatedAt:  time.Unix(updatedAt, 0),
	}

	if len(metadataJSON) > 0 {
		var metadata map[string]string
		if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
			settings.Metadata = metadata
		}
	}

	if len(thresholdsJSON) > 0 {
		var thresholds device.Thresholds
		if err := json.Unmarshal(thresholdsJSON, &thresholds); err == nil {
			settings.Thresholds = &thresholds
		}
	}

	return settings, nil
}
