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
		string(member.Status),
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return organization.ErrMemberExists
		}
		return err
	}

	return nil
}

// FindByID retrieves a member by ID.
func (s *MemberStorage) FindByID(ctx context.Context, id string) (*organization.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.operator_id, om.role, om.invited_by, om.joined_at, om.removed_at, om.status,
			   COALESCE(op.name, '') as operator_name, COALESCE(op.email, '') as operator_email
		FROM organization_members om
		LEFT JOIN operators op ON om.operator_id = op.id
		WHERE om.id = ?`

	var member organization.OrganizationMember
	var invitedBy sql.NullString
	var removedAt sql.NullInt64
	var operatorName, operatorEmail sql.NullString

	err := s.getQuerier(ctx).QueryRowContext(ctx, query, id).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.OperatorID,
		&member.Role,
		&invitedBy,
		&member.JoinedAt,
		&removedAt,
		&member.Status,
		&operatorName,
		&operatorEmail,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, organization.ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}

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

// FindByOperatorAndOrg finds a membership by operator ID and organization ID.
func (s *MemberStorage) FindByOperatorAndOrg(ctx context.Context, operatorID, orgID string) (*organization.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.operator_id, om.role, om.invited_by, om.joined_at, om.removed_at, om.status,
			   COALESCE(op.name, '') as operator_name, COALESCE(op.email, '') as operator_email
		FROM organization_members om
		LEFT JOIN operators op ON om.operator_id = op.id
		WHERE om.operator_id = ? AND om.organization_id = ? AND om.status = 'active'`

	var member organization.OrganizationMember
	var invitedBy sql.NullString
	var removedAt sql.NullInt64
	var operatorName, operatorEmail sql.NullString

	err := s.getQuerier(ctx).QueryRowContext(ctx, query, operatorID, orgID).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.OperatorID,
		&member.Role,
		&invitedBy,
		&member.JoinedAt,
		&removedAt,
		&member.Status,
		&operatorName,
		&operatorEmail,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, organization.ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}

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
		var invitedBy sql.NullString
		var removedAt sql.NullInt64
		var operatorName, operatorEmail sql.NullString

		err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.OperatorID,
			&member.Role,
			&invitedBy,
			&member.JoinedAt,
			&removedAt,
			&member.Status,
			&operatorName,
			&operatorEmail,
		)
		if err != nil {
			return nil, err
		}

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

// Update updates an existing membership.
func (s *MemberStorage) Update(ctx context.Context, member *organization.OrganizationMember) error {
	query := `
		UPDATE organization_members
		SET role = ?, status = ?
		WHERE id = ?`

	result, err := s.getQuerier(ctx).ExecContext(ctx, query,
		string(member.Role),
		string(member.Status),
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
		var invitedBy sql.NullString
		var removedAt sql.NullInt64
		var operatorName, operatorEmail sql.NullString

		err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.OperatorID,
			&member.Role,
			&invitedBy,
			&member.JoinedAt,
			&removedAt,
			&member.Status,
			&operatorName,
			&operatorEmail,
		)
		if err != nil {
			return nil, err
		}

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
