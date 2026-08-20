package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device_group"
)

// DeviceGroupRepository is the persistence implementation for device groups.
type DeviceGroupRepository struct {
	db *sql.DB
}

// NewDeviceGroupRepository creates a new DeviceGroupRepository.
func NewDeviceGroupRepository(db *sql.DB) *DeviceGroupRepository {
	return &DeviceGroupRepository{db: db}
}

// Save upserts a group.
func (r *DeviceGroupRepository) Save(ctx context.Context, g *device_group.Group) error {
	query := `
		INSERT INTO device_groups (id, org_id, name, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name
	`
	_, err := r.db.ExecContext(ctx, query, g.ID, g.OrgID, g.Name, g.CreatedAt.UnixMilli())
	return err
}

// GetByID returns a group or ErrNotFound.
func (r *DeviceGroupRepository) GetByID(ctx context.Context, id string) (*device_group.Group, error) {
	var g device_group.Group
	var ts int64
	err := r.db.QueryRowContext(ctx,
		`SELECT id, org_id, name, created_at FROM device_groups WHERE id = ?`, id,
	).Scan(&g.ID, &g.OrgID, &g.Name, &ts)
	if err == sql.ErrNoRows {
		return nil, device_group.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	g.CreatedAt = time.UnixMilli(ts)
	return &g, nil
}

// AddMember adds an operator to a group (idempotent).
func (r *DeviceGroupRepository) AddMember(ctx context.Context, groupID, operatorID string) error {
	query := `INSERT INTO device_group_members (group_id, operator_id, created_at) VALUES (?, ?, ?) ON CONFLICT(group_id, operator_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, groupID, operatorID, time.Now().UnixMilli())
	return err
}

// RemoveMember removes an operator from a group, returning whether it existed.
func (r *DeviceGroupRepository) RemoveMember(ctx context.Context, groupID, operatorID string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM device_group_members WHERE group_id = ? AND operator_id = ?`, groupID, operatorID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// IsMember reports whether the operator belongs to the group.
func (r *DeviceGroupRepository) IsMember(ctx context.Context, groupID, operatorID string) (bool, error) {
	var one int
	err := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM device_group_members WHERE group_id = ? AND operator_id = ? LIMIT 1`,
		groupID, operatorID,
	).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// AddDevice assigns a device to a group (idempotent).
func (r *DeviceGroupRepository) AddDevice(ctx context.Context, groupID, deviceID string) error {
	query := `INSERT INTO device_group_devices (group_id, device_id, created_at) VALUES (?, ?, ?) ON CONFLICT(group_id, device_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, groupID, deviceID, time.Now().UnixMilli())
	return err
}

// RemoveDevice unassigns a device from a group, returning whether it existed.
func (r *DeviceGroupRepository) RemoveDevice(ctx context.Context, groupID, deviceID string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM device_group_devices WHERE group_id = ? AND device_id = ?`, groupID, deviceID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// GroupIDsForDevice returns the group IDs a device belongs to.
func (r *DeviceGroupRepository) GroupIDsForDevice(ctx context.Context, deviceID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT group_id FROM device_group_devices WHERE device_id = ?`, deviceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
