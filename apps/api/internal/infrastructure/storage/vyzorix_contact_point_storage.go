package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
)

// ContactPointRepository is the SQL persistence for contact points.
type ContactPointRepository struct {
	db *sql.DB
}

// NewContactPointRepository creates a new ContactPointRepository.
func NewContactPointRepository(db *sql.DB) *ContactPointRepository {
	return &ContactPointRepository{db: db}
}

// Save upserts a contact point.
func (r *ContactPointRepository) Save(ctx context.Context, cp *notification.ContactPoint) error {
	configJSON, err := json.Marshal(cp.Config)
	if err != nil {
		return err
	}
	enabled := 0
	if cp.Enabled {
		enabled = 1
	}
	query := `
		INSERT INTO contact_points (id, org_id, name, channel, config, secret, template_id, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			channel = excluded.channel,
			config = excluded.config,
			secret = excluded.secret,
			template_id = excluded.template_id,
			enabled = excluded.enabled,
			updated_at = excluded.updated_at
	`
	_, err = r.db.ExecContext(ctx, query,
		cp.ID, cp.OrgID, cp.Name, string(cp.Channel), string(configJSON),
		cp.Secret, cp.TemplateID, enabled,
		cp.CreatedAt.UnixMilli(), cp.UpdatedAt.UnixMilli(),
	)
	return err
}

const contactPointColumns = `id, org_id, name, channel, config, secret, template_id, enabled, created_at, updated_at`

func scanContactPoint(scanner interface{ Scan(...any) error }) (*notification.ContactPoint, error) {
	var cp notification.ContactPoint
	var channel, configJSON string
	var enabled int
	var createdAt, updatedAt int64
	err := scanner.Scan(
		&cp.ID, &cp.OrgID, &cp.Name, &channel, &configJSON,
		&cp.Secret, &cp.TemplateID, &enabled, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	cp.Channel = notification.ChannelType(channel)
	cp.Enabled = enabled == 1
	cp.CreatedAt = time.UnixMilli(createdAt)
	cp.UpdatedAt = time.UnixMilli(updatedAt)
	if configJSON != "" && configJSON != "{}" {
		if err := json.Unmarshal([]byte(configJSON), &cp.Config); err != nil {
			return nil, err
		}
	}
	if cp.Config == nil {
		cp.Config = make(map[string]string)
	}
	return &cp, nil
}

// GetByID returns a contact point or notification.ErrNotFound.
func (r *ContactPointRepository) GetByID(ctx context.Context, id string) (*notification.ContactPoint, error) {
	cp, err := scanContactPoint(r.db.QueryRowContext(ctx,
		`SELECT `+contactPointColumns+` FROM contact_points WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return nil, notification.ErrNotFound
	}
	return cp, err
}

// ListByOrg returns all contact points of an org.
func (r *ContactPointRepository) ListByOrg(ctx context.Context, orgID string) ([]*notification.ContactPoint, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+contactPointColumns+` FROM contact_points WHERE org_id = ? ORDER BY name`, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var points []*notification.ContactPoint
	for rows.Next() {
		cp, err := scanContactPoint(rows)
		if err != nil {
			return nil, err
		}
		points = append(points, cp)
	}
	return points, rows.Err()
}

// ListEnabledByOrg returns enabled contact points of an org.
func (r *ContactPointRepository) ListEnabledByOrg(ctx context.Context, orgID string) ([]*notification.ContactPoint, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+contactPointColumns+` FROM contact_points WHERE org_id = ? AND enabled = 1 ORDER BY name`, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var points []*notification.ContactPoint
	for rows.Next() {
		cp, err := scanContactPoint(rows)
		if err != nil {
			return nil, err
		}
		points = append(points, cp)
	}
	return points, rows.Err()
}

// Delete removes a contact point.
func (r *ContactPointRepository) Delete(ctx context.Context, id string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM contact_points WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// DeliveryRepository is the SQL persistence for delivery records.
type DeliveryRepository struct {
	db *sql.DB
}

// NewDeliveryRepository creates a new DeliveryRepository.
func NewDeliveryRepository(db *sql.DB) *DeliveryRepository {
	return &DeliveryRepository{db: db}
}

// Append records a delivery attempt.
func (r *DeliveryRepository) Append(ctx context.Context, d *notification.Delivery) error {
	payload, err := json.Marshal(d.Message)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO notification_deliveries (id, contact_point_id, channel, status, error, payload, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.ContactPointID, string(d.Channel), d.Status, d.Error, string(payload), d.CreatedAt.UnixMilli())
	return err
}

// ListByContactPoint returns recent deliveries for a contact point.
func (r *DeliveryRepository) ListByContactPoint(ctx context.Context, contactPointID string, limit int) ([]*notification.Delivery, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, contact_point_id, channel, status, error, payload, created_at
		FROM notification_deliveries
		WHERE contact_point_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, contactPointID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var deliveries []*notification.Delivery
	for rows.Next() {
		var d notification.Delivery
		var channel, payload string
		var createdAt int64
		if err := rows.Scan(&d.ID, &d.ContactPointID, &channel, &d.Status, &d.Error, &payload, &createdAt); err != nil {
			return nil, err
		}
		d.Channel = notification.ChannelType(channel)
		d.CreatedAt = time.UnixMilli(createdAt)
		if payload != "" {
			_ = json.Unmarshal([]byte(payload), &d.Message)
		}
		deliveries = append(deliveries, &d)
	}
	return deliveries, rows.Err()
}
