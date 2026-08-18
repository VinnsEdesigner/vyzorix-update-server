package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// OrganizationSettingsRepository implements organization.OrganizationSettingsRepository.
type OrganizationSettingsRepository struct {
	db *sql.DB
}

// NewOrganizationSettingsRepository creates a new OrganizationSettingsRepository.
func NewOrganizationSettingsRepository(db *sql.DB) *OrganizationSettingsRepository {
	return &OrganizationSettingsRepository{db: db}
}

// migrateOrganizationSettings creates the organization_settings table.
func migrateOrganizationSettings(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS organization_settings (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL UNIQUE,
			timezone TEXT NOT NULL DEFAULT 'UTC',
			date_format TEXT NOT NULL DEFAULT 'YYYY-MM-DD',
			alert_cooldown_minutes INTEGER NOT NULL DEFAULT 15,
			default_thresholds JSONB,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create index for efficient lookups.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_org_settings_org ON organization_settings(organization_id)
	`)

	return err
}

// Create creates new organization settings.
func (r *OrganizationSettingsRepository) Create(ctx context.Context, settings *organization.OrganizationSettings) error {
	var thresholdsJSON []byte
	var err error
	if settings.DefaultThresholds != nil {
		thresholdsJSON, err = json.Marshal(settings.DefaultThresholds)
		if err != nil {
			return err
		}
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO organization_settings (id, organization_id, timezone, date_format, alert_cooldown_minutes, default_thresholds, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, settings.ID, settings.OrganizationID, settings.Timezone, settings.DateFormat, settings.AlertCooldownMinutes, thresholdsJSON, settings.CreatedAt.Unix(), settings.UpdatedAt.Unix())

	return err
}

// FindByID retrieves settings by ID.
func (r *OrganizationSettingsRepository) FindByID(ctx context.Context, id string) (*organization.OrganizationSettings, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, organization_id, timezone, date_format, alert_cooldown_minutes, default_thresholds, created_at, updated_at
		FROM organization_settings
		WHERE id = ?
	`, id)

	return r.scanSettings(row)
}

// FindByOrganizationID retrieves settings by organization ID.
func (r *OrganizationSettingsRepository) FindByOrganizationID(ctx context.Context, orgID string) (*organization.OrganizationSettings, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, organization_id, timezone, date_format, alert_cooldown_minutes, default_thresholds, created_at, updated_at
		FROM organization_settings
		WHERE organization_id = ?
	`, orgID)

	return r.scanSettings(row)
}

// Update updates organization settings.
func (r *OrganizationSettingsRepository) Update(ctx context.Context, settings *organization.OrganizationSettings) error {
	var thresholdsJSON []byte
	var err error
	if settings.DefaultThresholds != nil {
		thresholdsJSON, err = json.Marshal(settings.DefaultThresholds)
		if err != nil {
			return err
		}
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE organization_settings
		SET timezone = ?, date_format = ?, alert_cooldown_minutes = ?, default_thresholds = ?, updated_at = ?
		WHERE id = ?
	`, settings.Timezone, settings.DateFormat, settings.AlertCooldownMinutes, thresholdsJSON, time.Now().Unix(), settings.ID)

	return err
}

// Delete deletes organization settings by ID.
func (r *OrganizationSettingsRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM organization_settings WHERE id = ?`, id)
	return err
}

// DeleteByOrganizationID deletes settings by organization ID.
func (r *OrganizationSettingsRepository) DeleteByOrganizationID(ctx context.Context, orgID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM organization_settings WHERE organization_id = ?`, orgID)
	return err
}

// scanSettings scans a row into OrganizationSettings.
func (r *OrganizationSettingsRepository) scanSettings(row *sql.Row) (*organization.OrganizationSettings, error) {
	var id, orgID, timezone, dateFormat string
	var alertCooldownMinutes int
	var thresholdsJSON []byte
	var createdAt, updatedAt int64

	err := row.Scan(&id, &orgID, &timezone, &dateFormat, &alertCooldownMinutes, &thresholdsJSON, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, organization.ErrSettingsNotFound
		}
		return nil, err
	}

	settings := &organization.OrganizationSettings{
		ID:                   id,
		OrganizationID:       orgID,
		Timezone:             timezone,
		DateFormat:           dateFormat,
		AlertCooldownMinutes: alertCooldownMinutes,
		CreatedAt:            time.Unix(createdAt, 0),
		UpdatedAt:            time.Unix(updatedAt, 0),
	}

	if len(thresholdsJSON) > 0 {
		var thresholds organization.Thresholds
		if err := json.Unmarshal(thresholdsJSON, &thresholds); err == nil {
			settings.DefaultThresholds = &thresholds
		}
	}

	return settings, nil
}
