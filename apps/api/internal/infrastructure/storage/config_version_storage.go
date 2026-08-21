package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/configversion"
)

// ConfigVersionRepository is the SQL persistence for config versions.
type ConfigVersionRepository struct {
	db *sql.DB
}

// NewConfigVersionRepository creates a new ConfigVersionRepository.
func NewConfigVersionRepository(db *sql.DB) *ConfigVersionRepository {
	return &ConfigVersionRepository{db: db}
}

const configVersionColumns = `id, org_id, resource_type, version, snapshot, changed_by, created_at`

func scanConfigVersion(scanner interface{ Scan(...any) error }) (*configversion.ConfigVersion, error) {
	var v configversion.ConfigVersion
	var resourceType string
	var createdAt int64
	err := scanner.Scan(
		&v.ID, &v.OrgID, &resourceType, &v.Version, &v.Snapshot,
		&v.ChangedBy, &createdAt,
	)
	if err != nil {
		return nil, err
	}
	v.ResourceType = configversion.ResourceType(resourceType)
	v.CreatedAt = time.UnixMilli(createdAt)
	return &v, nil
}

// Append inserts a new version and returns its assigned version number.
func (r *ConfigVersionRepository) Append(ctx context.Context, v *configversion.ConfigVersion) (int64, error) {
	next, err := r.Latest(ctx, v.OrgID, v.ResourceType)
	if err != nil {
		return 0, err
	}
	v.Version = next + 1

	query := `
		INSERT INTO config_versions (id, org_id, resource_type, version, snapshot, changed_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, query,
		v.ID, v.OrgID, string(v.ResourceType), v.Version, v.Snapshot,
		v.ChangedBy, v.CreatedAt.UnixMilli(),
	)
	return v.Version, err
}

// List returns versions of one resource, newest first.
func (r *ConfigVersionRepository) List(ctx context.Context, orgID string, resourceType configversion.ResourceType, limit int) ([]*configversion.ConfigVersion, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+configVersionColumns+` FROM config_versions
		 WHERE org_id = ? AND resource_type = ? ORDER BY version DESC LIMIT ?`,
		orgID, string(resourceType), limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var versions []*configversion.ConfigVersion
	for rows.Next() {
		v, err := scanConfigVersion(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// Get returns one version or configversion.ErrNotFound.
func (r *ConfigVersionRepository) Get(ctx context.Context, orgID string, resourceType configversion.ResourceType, version int64) (*configversion.ConfigVersion, error) {
	v, err := scanConfigVersion(r.db.QueryRowContext(ctx,
		`SELECT `+configVersionColumns+` FROM config_versions
		 WHERE org_id = ? AND resource_type = ? AND version = ?`,
		orgID, string(resourceType), version))
	if err == sql.ErrNoRows {
		return nil, configversion.ErrNotFound
	}
	return v, err
}

// Latest returns the highest version number for a resource, or 0 when none.
func (r *ConfigVersionRepository) Latest(ctx context.Context, orgID string, resourceType configversion.ResourceType) (int64, error) {
	var latest sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT MAX(version) FROM config_versions WHERE org_id = ? AND resource_type = ?`,
		orgID, string(resourceType)).Scan(&latest)
	if err != nil {
		return 0, err
	}
	if !latest.Valid {
		return 0, nil
	}
	return latest.Int64, nil
}
