package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Ensure MemberRepository implements organization.MemberRepository.
var _ organization.MemberRepository = (*MemberStorage)(nil)

// MemberStorage implements organization.MemberRepository using SQLite.
type MemberStorage struct {
	db *sql.DB
}

// NewMemberStorage creates a new MemberStorage.
func NewMemberStorage(db *sql.DB) *MemberStorage {
	return &MemberStorage{db: db}
}

// getQuerier returns the transaction from context if available, otherwise the db.
func (s *MemberStorage) getQuerier(ctx context.Context) Querier {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return s.db
}

// Create creates a new membership.
func (s *MemberStorage) Create(ctx context.Context, member *organization.OrganizationMember) error {
	query := `
		INSERT INTO organization_members (id, organization_id, operator_id, role, invited_by, joined_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.getQuerier(ctx).ExecContext(ctx, query,
		member.ID,
		member.OrganizationID,
		member.OperatorID,
		string(member.Role),
		member.InvitedBy,
		member.JoinedAt.UnixMilli(),
		string(member.Lifecycle),
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return organization.ErrMemberExists
		}
		return err
	}

	return nil
}

// scanMember scans a member from a row.
func scanMember(row *sql.Row) (*organization.OrganizationMember, error) {
	var member organization.OrganizationMember
	var joinedAtMs int64
	var invitedBy sql.NullString
	var removedAt sql.NullInt64
	var operatorName, operatorEmail sql.NullString
	var status string

	err := row.Scan(
		&member.ID,
		&member.OrganizationID,
		&member.OperatorID,
		&member.Role,
		&invitedBy,
		&joinedAtMs,
		&removedAt,
		&status,
		&operatorName,
		&operatorEmail,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, organization.ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}

	member.JoinedAt = time.UnixMilli(joinedAtMs)
	member.Lifecycle = organization.MemberLifecycle(status)
	if invitedBy.Valid {
		member.InvitedBy = &invitedBy.String
	}
	if removedAt.Valid {
		t := time.UnixMilli(removedAt.Int64)
		member.RemovedAt = &t
	}
	member.OperatorName = operatorName.String
	member.OperatorEmail = operatorEmail.String

	return &member, nil
}

// FindByID retrieves a member by ID.
func (s *MemberStorage) FindByID(ctx context.Context, id string) (*organization.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.operator_id, om.role, om.invited_by, om.joined_at, om.removed_at, om.status,
			   COALESCE(op.name, '') as operator_name, COALESCE(op.email, '') as operator_email
		FROM organization_members om
		LEFT JOIN operators op ON om.operator_id = op.id
		WHERE om.id = ?`

	return scanMember(s.getQuerier(ctx).QueryRowContext(ctx, query, id))
}

// FindByMemberID retrieves a member by MemberID value object.
func (s *MemberStorage) FindByMemberID(ctx context.Context, id organization.MemberID) (*organization.OrganizationMember, error) {
	return s.FindByID(ctx, id.String())
}

// FindByOperatorAndOrg finds a membership by operator ID and organization ID.
func (s *MemberStorage) FindByOperatorAndOrg(ctx context.Context, operatorID, orgID string) (*organization.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.operator_id, om.role, om.invited_by, om.joined_at, om.removed_at, om.status,
			   COALESCE(op.name, '') as operator_name, COALESCE(op.email, '') as operator_email
		FROM organization_members om
		LEFT JOIN operators op ON om.operator_id = op.id
		WHERE om.operator_id = ? AND om.organization_id = ? AND om.status = 'active'`

	var member organization.OrganizationMember
	var joinedAtMs int64
	var invitedBy sql.NullString
	var removedAt sql.NullInt64
	var status string
	var operatorName, operatorEmail sql.NullString

	err := s.getQuerier(ctx).QueryRowContext(ctx, query, operatorID, orgID).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.OperatorID,
		&member.Role,
		&invitedBy,
		&joinedAtMs,
		&removedAt,
		&status,
		&operatorName,
		&operatorEmail,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, organization.ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}

	member.JoinedAt = time.UnixMilli(joinedAtMs)
	// Map status string to lifecycle.
	member.Lifecycle = memberStatusToLifecycle(status)

	if invitedBy.Valid {
		member.InvitedBy = &invitedBy.String
	}
	if removedAt.Valid {
		t := time.UnixMilli(removedAt.Int64)
		member.RemovedAt = &t
	}
	member.OperatorName = operatorName.String
	member.OperatorEmail = operatorEmail.String

	return &member, nil
}

// FindByOrganization lists all members of an organization.
func (s *MemberStorage) FindByOrganization(ctx context.Context, orgID string) ([]*organization.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.operator_id, om.role, om.invited_by, om.joined_at, om.removed_at, om.status,
			   COALESCE(op.name, '') as operator_name, COALESCE(op.email, '') as operator_email
		FROM organization_members om
		LEFT JOIN operators op ON om.operator_id = op.id
		WHERE om.organization_id = ? AND om.status = 'active'
		ORDER BY om.joined_at ASC`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var members []*organization.OrganizationMember
	for rows.Next() {
		var member organization.OrganizationMember
		var joinedAtMs int64
		var invitedBy sql.NullString
		var removedAt sql.NullInt64
		var status string
		var operatorName, operatorEmail sql.NullString

		err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.OperatorID,
			&member.Role,
			&invitedBy,
			&joinedAtMs,
			&removedAt,
			&status,
			&operatorName,
			&operatorEmail,
		)
		if err != nil {
			return nil, err
		}

		member.JoinedAt = time.UnixMilli(joinedAtMs)
		// Map status string to lifecycle.
		member.Lifecycle = memberStatusToLifecycle(status)

		if invitedBy.Valid {
			member.InvitedBy = &invitedBy.String
		}
		if removedAt.Valid {
			t := time.UnixMilli(removedAt.Int64)
			member.RemovedAt = &t
		}
		member.OperatorName = operatorName.String
		member.OperatorEmail = operatorEmail.String

		members = append(members, &member)
	}

	return members, rows.Err()
}

// FindActiveByOrganizationPaginated lists active members with pagination.
func (s *MemberStorage) FindActiveByOrganizationPaginated(ctx context.Context, orgID string, limit, offset int) ([]*organization.OrganizationMember, int, error) {
	// Get total count.
	countQuery := `
		SELECT COUNT(*) FROM organization_members
		WHERE organization_id = ? AND status = 'active'`

	var total int
	err := s.getQuerier(ctx).QueryRowContext(ctx, countQuery, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results.
	query := `
		SELECT om.id, om.organization_id, om.operator_id, om.role, om.invited_by, om.joined_at, om.removed_at, om.status,
			   COALESCE(op.name, '') as operator_name, COALESCE(op.email, '') as operator_email
		FROM organization_members om
		LEFT JOIN operators op ON om.operator_id = op.id
		WHERE om.organization_id = ? AND om.status = 'active'
		ORDER BY om.joined_at ASC
		LIMIT ? OFFSET ?`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var members []*organization.OrganizationMember
	for rows.Next() {
		var member organization.OrganizationMember
		var joinedAtMs int64
		var invitedBy sql.NullString
		var removedAt sql.NullInt64
		var status string
		var operatorName, operatorEmail sql.NullString

		err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.OperatorID,
			&member.Role,
			&invitedBy,
			&joinedAtMs,
			&removedAt,
			&status,
			&operatorName,
			&operatorEmail,
		)
		if err != nil {
			return nil, 0, err
		}

		member.JoinedAt = time.UnixMilli(joinedAtMs)
		member.Lifecycle = memberStatusToLifecycle(status)

		if invitedBy.Valid {
			member.InvitedBy = &invitedBy.String
		}
		if removedAt.Valid {
			t := time.UnixMilli(removedAt.Int64)
			member.RemovedAt = &t
		}
		member.OperatorName = operatorName.String
		member.OperatorEmail = operatorEmail.String

		members = append(members, &member)
	}

	return members, total, rows.Err()
}

// Update updates an existing membership.
func (s *MemberStorage) Update(ctx context.Context, member *organization.OrganizationMember) error {
	query := `
		UPDATE organization_members
		SET role = ?, status = ?
		WHERE id = ?`

	result, err := s.getQuerier(ctx).ExecContext(ctx, query,
		string(member.Role),
		string(member.Lifecycle),
		member.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return organization.ErrMemberNotFound
	}

	return nil
}

// SoftDelete soft-deletes a membership (marks as removed).
func (s *MemberStorage) SoftDelete(ctx context.Context, id string) error {
	query := `
		UPDATE organization_members
		SET status = 'removed', removed_at = ?
		WHERE id = ? AND status = 'active'`

	now := time.Now().UnixMilli()
	result, err := s.getQuerier(ctx).ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return organization.ErrMemberNotFound
	}

	return nil
}

// CountByOrganization counts members in an organization (including removed).
func (s *MemberStorage) CountByOrganization(ctx context.Context, orgID string) (int, error) {
	query := `SELECT COUNT(*) FROM organization_members WHERE organization_id = ?`

	var count int
	err := s.getQuerier(ctx).QueryRowContext(ctx, query, orgID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CountActiveByOrganization counts active (non-removed) members in an organization.
func (s *MemberStorage) CountActiveByOrganization(ctx context.Context, orgID string) (int, error) {
	query := `SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND status = 'active'`

	var count int
	err := s.getQuerier(ctx).QueryRowContext(ctx, query, orgID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CountSuperAdminsByOrganization counts super_admin members in an organization.
func (s *MemberStorage) CountSuperAdminsByOrganization(ctx context.Context, orgID string) (int, error) {
	query := `SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND role = 'super_admin' AND status = 'active'`

	var count int
	err := s.getQuerier(ctx).QueryRowContext(ctx, query, orgID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ListByOperator lists all memberships for an operator.
func (s *MemberStorage) ListByOperator(ctx context.Context, operatorID string) ([]*organization.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.operator_id, om.role, om.invited_by, om.joined_at, om.removed_at, om.status,
			   COALESCE(op.name, '') as operator_name, COALESCE(op.email, '') as operator_email
		FROM organization_members om
		LEFT JOIN operators op ON om.operator_id = op.id
		WHERE om.operator_id = ? AND om.status = 'active'
		ORDER BY om.joined_at DESC`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, operatorID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var members []*organization.OrganizationMember
	for rows.Next() {
		var member organization.OrganizationMember
		var joinedAtMs int64
		var invitedBy sql.NullString
		var removedAt sql.NullInt64
		var status string
		var operatorName, operatorEmail sql.NullString

		err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.OperatorID,
			&member.Role,
			&invitedBy,
			&joinedAtMs,
			&removedAt,
			&status,
			&operatorName,
			&operatorEmail,
		)
		if err != nil {
			return nil, err
		}

		member.JoinedAt = time.UnixMilli(joinedAtMs)
		// Map status string to lifecycle.
		member.Lifecycle = memberStatusToLifecycle(status)

		if invitedBy.Valid {
			member.InvitedBy = &invitedBy.String
		}
		if removedAt.Valid {
			t := time.UnixMilli(removedAt.Int64)
			member.RemovedAt = &t
		}
		member.OperatorName = operatorName.String
		member.OperatorEmail = operatorEmail.String

		members = append(members, &member)
	}

	return members, rows.Err()
}

// ListByOperatorPaginated lists memberships for an operator with pagination.
func (s *MemberStorage) ListByOperatorPaginated(ctx context.Context, operatorID string, limit, offset int) ([]*organization.OrganizationMember, int, error) {
	// Get total count.
	countQuery := `
		SELECT COUNT(*) FROM organization_members
		WHERE operator_id = ? AND status = 'active'`

	var total int
	err := s.getQuerier(ctx).QueryRowContext(ctx, countQuery, operatorID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results.
	query := `
		SELECT om.id, om.organization_id, om.operator_id, om.role, om.invited_by, om.joined_at, om.removed_at, om.status,
			   COALESCE(op.name, '') as operator_name, COALESCE(op.email, '') as operator_email
		FROM organization_members om
		LEFT JOIN operators op ON om.operator_id = op.id
		WHERE om.operator_id = ? AND om.status = 'active'
		ORDER BY om.joined_at DESC
		LIMIT ? OFFSET ?`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, operatorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var members []*organization.OrganizationMember
	for rows.Next() {
		var member organization.OrganizationMember
		var joinedAtMs int64
		var invitedBy sql.NullString
		var removedAt sql.NullInt64
		var status string
		var operatorName, operatorEmail sql.NullString

		err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.OperatorID,
			&member.Role,
			&invitedBy,
			&joinedAtMs,
			&removedAt,
			&status,
			&operatorName,
			&operatorEmail,
		)
		if err != nil {
			return nil, 0, err
		}

		member.JoinedAt = time.UnixMilli(joinedAtMs)
		member.Lifecycle = memberStatusToLifecycle(status)

		if invitedBy.Valid {
			member.InvitedBy = &invitedBy.String
		}
		if removedAt.Valid {
			t := time.UnixMilli(removedAt.Int64)
			member.RemovedAt = &t
		}
		member.OperatorName = operatorName.String
		member.OperatorEmail = operatorEmail.String

		members = append(members, &member)
	}

	return members, total, rows.Err()
}


// SoftDeleteByOperator soft-deletes all memberships for an operator (used during operator deletion).
func (s *MemberStorage) SoftDeleteByOperator(ctx context.Context, operatorID string) error {
	query := `
		UPDATE organization_members
		SET status = 'removed', removed_at = ?
		WHERE operator_id = ? AND status = 'active'`

	now := time.Now().UnixMilli()
	_, err := s.getQuerier(ctx).ExecContext(ctx, query, now, operatorID)
	return err
}

// FindByOperatorAndOrgID finds a membership by OperatorID and OrganizationID value objects.
func (s *MemberStorage) FindByOperatorAndOrgID(ctx context.Context, operatorID organization.OperatorID, orgID organization.OrganizationID) (*organization.OrganizationMember, error) {
	return s.FindByOperatorAndOrg(ctx, operatorID.String(), orgID.String())
}

// FindByOrganizationID lists all members by OrganizationID value object.
func (s *MemberStorage) FindByOrganizationID(ctx context.Context, orgID organization.OrganizationID) ([]*organization.OrganizationMember, error) {
	return s.FindByOrganization(ctx, orgID.String())
}

// FindActiveByOrganization lists all active members of an organization.
func (s *MemberStorage) FindActiveByOrganization(ctx context.Context, orgID string) ([]*organization.OrganizationMember, error) {
	return s.FindByOrganization(ctx, orgID)
}

// FindActiveByOrganizationID lists all active members by OrganizationID.
func (s *MemberStorage) FindActiveByOrganizationID(ctx context.Context, orgID organization.OrganizationID) ([]*organization.OrganizationMember, error) {
	return s.FindByOrganization(ctx, orgID.String())
}

// SoftDeleteByMemberID soft-deletes a membership by MemberID.
func (s *MemberStorage) SoftDeleteByMemberID(ctx context.Context, id organization.MemberID) error {
	return s.SoftDelete(ctx, id.String())
}

// SoftDeleteByOperatorID soft-deletes all memberships for an OperatorID.
func (s *MemberStorage) SoftDeleteByOperatorID(ctx context.Context, operatorID organization.OperatorID) error {
	return s.SoftDeleteByOperator(ctx, operatorID.String())
}

// SoftDeleteByOrganization soft-deletes all memberships for an organization.
func (s *MemberStorage) SoftDeleteByOrganization(ctx context.Context, orgID string) error {
	query := `
		UPDATE organization_members
		SET status = 'removed', removed_at = ?
		WHERE organization_id = ? AND status = 'active'`

	now := time.Now().UnixMilli()
	_, err := s.getQuerier(ctx).ExecContext(ctx, query, now, orgID)
	return err
}

// CountByOrganizationID counts members by OrganizationID.
func (s *MemberStorage) CountByOrganizationID(ctx context.Context, orgID organization.OrganizationID) (int, error) {
	return s.CountByOrganization(ctx, orgID.String())
}

// CountActiveByOrganizationID counts active members by OrganizationID.
func (s *MemberStorage) CountActiveByOrganizationID(ctx context.Context, orgID organization.OrganizationID) (int, error) {
	return s.CountActiveByOrganization(ctx, orgID.String())
}

// CountSuperAdminsByOrganizationID counts super_admin members by OrganizationID.
func (s *MemberStorage) CountSuperAdminsByOrganizationID(ctx context.Context, orgID organization.OrganizationID) (int, error) {
	return s.CountSuperAdminsByOrganization(ctx, orgID.String())
}

// ListByOperatorID lists all memberships by OperatorID.
func (s *MemberStorage) ListByOperatorID(ctx context.Context, operatorID organization.OperatorID) ([]*organization.OrganizationMember, error) {
	return s.ListByOperator(ctx, operatorID.String())
}

// memberStatusToLifecycle maps a status string from the database to MemberLifecycle.
func memberStatusToLifecycle(status string) organization.MemberLifecycle {
	switch status {
	case "active":
		return organization.MemberLifecycleActive
	case "suspended":
		return organization.MemberLifecycleSuspended
	case "removed":
		return organization.MemberLifecycleRemoved
	default:
		return organization.MemberLifecycleActive
	}
}
