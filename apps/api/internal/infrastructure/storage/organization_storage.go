package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Ensure OrganizationRepository implements organization.Repository.
var _ organization.OrganizationRepository = (*OrganizationStorage)(nil)

// OrganizationStorage implements organization.OrganizationRepository using SQLite.
type OrganizationStorage struct {
	db *sql.DB
}

// NewOrganizationStorage creates a new OrganizationStorage.
func NewOrganizationStorage(db *sql.DB) *OrganizationStorage {
	return &OrganizationStorage{db: db}
}

// getQuerier returns the transaction from context if available, otherwise the db.
func (s *OrganizationStorage) getQuerier(ctx context.Context) Querier {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return s.db
}

// Create creates a new organization.
func (s *OrganizationStorage) Create(ctx context.Context, org *organization.Organization) error {
	query := `
		INSERT INTO organizations (id, name, created_by, created_at, updated_at, is_active, max_members)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.getQuerier(ctx).ExecContext(ctx, query,
		org.ID,
		org.Name,
		org.CreatedBy,
		org.CreatedAt.UnixMilli(),
		org.UpdatedAt.UnixMilli(),
		boolToInt(org.IsActive),
		org.MaxMembers,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return organization.ErrOrganizationExists
		}
		return err
	}

	return nil
}

// scanOrganization scans an organization from a row.
func scanOrganization(row *sql.Row) (*organization.Organization, error) {
	var org organization.Organization
	var deletedAt sql.NullInt64
	var createdBy string

	err := row.Scan(
		&org.ID,
		&org.Name,
		&createdBy,
		&org.CreatedAt,
		&org.UpdatedAt,
		&deletedAt,
		&org.IsActive,
		&org.MaxMembers,
		&org.MemberCount,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, organization.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	org.CreatedBy = createdBy
	if deletedAt.Valid {
		org.DeletedAt = ptrTime(time.UnixMilli(deletedAt.Int64))
		org.Lifecycle = organization.OrganizationLifecycleArchived
	} else if org.IsActive {
		org.Lifecycle = organization.OrganizationLifecycleActive
	} else {
		org.Lifecycle = organization.OrganizationLifecycleInactive
	}

	return &org, nil
}

// scanOrganizations scans multiple organizations from rows.
func scanOrganizations(rows *sql.Rows) ([]*organization.Organization, error) {
	var orgs []*organization.Organization

	for rows.Next() {
		var org organization.Organization
		var deletedAt sql.NullInt64
		var createdBy string

		err := rows.Scan(
			&org.ID,
			&org.Name,
			&createdBy,
			&org.CreatedAt,
			&org.UpdatedAt,
			&deletedAt,
			&org.IsActive,
			&org.MaxMembers,
			&org.MemberCount,
		)
		if err != nil {
			return nil, err
		}

		org.CreatedBy = createdBy
		if deletedAt.Valid {
			org.DeletedAt = ptrTime(time.UnixMilli(deletedAt.Int64))
			org.Lifecycle = organization.OrganizationLifecycleArchived
		} else if org.IsActive {
			org.Lifecycle = organization.OrganizationLifecycleActive
		} else {
			org.Lifecycle = organization.OrganizationLifecycleInactive
		}

		orgs = append(orgs, &org)
	}

	return orgs, rows.Err()
}

// ptrTime returns a pointer to a time.Time.
func ptrTime(t time.Time) *time.Time {
	return &t
}

// FindByID retrieves an organization by ID.
func (s *OrganizationStorage) FindByID(ctx context.Context, id string) (*organization.Organization, error) {
	query := `
		SELECT o.id, o.name, o.created_by, o.created_at, o.updated_at, o.deleted_at, o.is_active, o.max_members,
			   COUNT(DISTINCT om.id) as member_count
		FROM organizations o
		LEFT JOIN organization_members om ON o.id = om.organization_id AND om.status = 'active'
		WHERE o.id = ? AND o.deleted_at IS NULL
		GROUP BY o.id`

	return scanOrganization(s.getQuerier(ctx).QueryRowContext(ctx, query, id))
}

// FindByOrganizationID retrieves an organization by OrganizationID value object.
func (s *OrganizationStorage) FindByOrganizationID(ctx context.Context, id organization.OrganizationID) (*organization.Organization, error) {
	return s.FindByID(ctx, id.String())
}

// FindByName finds an organization by name for a specific operator.
func (s *OrganizationStorage) FindByName(ctx context.Context, operatorID, name string) (*organization.Organization, error) {
	query := `
		SELECT o.id, o.name, o.created_by, o.created_at, o.updated_at, o.deleted_at, o.is_active, o.max_members,
			   COUNT(DISTINCT om.id) as member_count
		FROM organizations o
		LEFT JOIN organization_members om ON o.id = om.organization_id AND om.status = 'active'
		WHERE o.created_by = ? AND o.name = ? AND o.deleted_at IS NULL
		GROUP BY o.id`

	var org organization.Organization
	var deletedAt sql.NullInt64
	var createdBy string

	err := s.getQuerier(ctx).QueryRowContext(ctx, query, operatorID, name).Scan(
		&org.ID,
		&org.Name,
		&createdBy,
		&org.CreatedAt,
		&org.UpdatedAt,
		&deletedAt,
		&org.IsActive,
		&org.MaxMembers,
		&org.MemberCount,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, organization.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	org.CreatedBy = createdBy
	if deletedAt.Valid {
		t := time.UnixMilli(deletedAt.Int64)
		org.DeletedAt = &t
	}

	return &org, nil
}

// Update updates an existing organization.
func (s *OrganizationStorage) Update(ctx context.Context, org *organization.Organization) error {
	query := `
		UPDATE organizations
		SET name = ?, updated_at = ?, is_active = ?, max_members = ?
		WHERE id = ? AND deleted_at IS NULL`

	result, err := s.getQuerier(ctx).ExecContext(ctx, query,
		org.Name,
		time.Now().UnixMilli(),
		boolToInt(org.IsActive),
		org.MaxMembers,
		org.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return organization.ErrNotFound
	}

	return nil
}

// SoftDelete soft-deletes an organization.
func (s *OrganizationStorage) SoftDelete(ctx context.Context, id string) error {
	query := `
		UPDATE organizations
		SET deleted_at = ?, updated_at = ?, is_active = 0
		WHERE id = ? AND deleted_at IS NULL`

	now := time.Now().UnixMilli()
	result, err := s.getQuerier(ctx).ExecContext(ctx, query, now, now, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return organization.ErrNotFound
	}

	return nil
}

// ListByOperator lists all organizations for an operator (via memberships).
func (s *OrganizationStorage) ListByOperator(ctx context.Context, operatorID string) ([]*organization.Organization, error) {
	query := `
		SELECT o.id, o.name, o.created_by, o.created_at, o.updated_at, o.deleted_at, o.is_active, o.max_members,
			   COUNT(DISTINCT om2.id) as member_count
		FROM organizations o
		INNER JOIN organization_members om ON o.id = om.organization_id AND om.status = 'active'
		LEFT JOIN organization_members om2 ON o.id = om2.organization_id AND om2.status = 'active'
		WHERE om.operator_id = ? AND o.deleted_at IS NULL
		GROUP BY o.id
		ORDER BY o.created_at DESC`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, operatorID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var orgs []*organization.Organization
	for rows.Next() {
		var org organization.Organization
		var deletedAt sql.NullInt64
		var createdBy string

		err := rows.Scan(
			&org.ID,
			&org.Name,
			&createdBy,
			&org.CreatedAt,
			&org.UpdatedAt,
			&deletedAt,
			&org.IsActive,
			&org.MaxMembers,
			&org.MemberCount,
		)
		if err != nil {
			return nil, err
		}

		org.CreatedBy = createdBy
		if deletedAt.Valid {
			t := time.UnixMilli(deletedAt.Int64)
			org.DeletedAt = &t
		}

		orgs = append(orgs, &org)
	}

	return orgs, rows.Err()
}

// CountActiveMembers counts the number of active members in an organization.
func (s *OrganizationStorage) CountActiveMembers(ctx context.Context, orgID string) (int, error) {
	query := `
		SELECT COUNT(*) FROM organization_members
		WHERE organization_id = ? AND status = 'active'`

	var count int
	err := s.getQuerier(ctx).QueryRowContext(ctx, query, orgID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// FindByNameAndOperatorID finds an organization by name for a specific operator.
func (s *OrganizationStorage) FindByNameAndOperatorID(ctx context.Context, operatorID organization.OperatorID, name string) (*organization.Organization, error) {
	return s.FindByName(ctx, operatorID.String(), name)
}

// SoftDeleteByID soft-deletes an organization by OrganizationID.
func (s *OrganizationStorage) SoftDeleteByID(ctx context.Context, id organization.OrganizationID) error {
	return s.SoftDelete(ctx, id.String())
}

// ListByOperatorID lists all organizations for an OperatorID.
func (s *OrganizationStorage) ListByOperatorID(ctx context.Context, operatorID organization.OperatorID) ([]*organization.Organization, error) {
	return s.ListByOperator(ctx, operatorID.String())
}

// ListActive lists all active (non-archived) organizations.
func (s *OrganizationStorage) ListActive(ctx context.Context) ([]*organization.Organization, error) {
	query := `
		SELECT o.id, o.name, o.created_by, o.created_at, o.updated_at, o.deleted_at, o.is_active, o.max_members,
			   COUNT(DISTINCT om.id) as member_count
		FROM organizations o
		LEFT JOIN organization_members om ON o.id = om.organization_id AND om.status = 'active'
		WHERE o.deleted_at IS NULL AND o.is_active = 1
		GROUP BY o.id
		ORDER BY o.created_at DESC`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanOrganizations(rows)
}

// CountActiveMembersByID counts the number of active members by OrganizationID.
func (s *OrganizationStorage) CountActiveMembersByID(ctx context.Context, orgID organization.OrganizationID) (int, error) {
	return s.CountActiveMembers(ctx, orgID.String())
}
