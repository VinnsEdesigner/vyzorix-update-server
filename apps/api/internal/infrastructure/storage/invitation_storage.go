package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/invitation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Ensure InvitationRepository implements invitation.Repository.
var _ invitation.Repository = (*InvitationStorage)(nil)

// InvitationStorage implements invitation.Repository using SQLite.
type InvitationStorage struct {
	db *sql.DB
}

// NewInvitationStorage creates a new InvitationStorage.
func NewInvitationStorage(db *sql.DB) *InvitationStorage {
	return &InvitationStorage{db: db}
}

// getQuerier returns the transaction from context if available, otherwise the db.
func (s *InvitationStorage) getQuerier(ctx context.Context) Querier {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return s.db
}

// Create creates a new invitation.
func (s *InvitationStorage) Create(ctx context.Context, invite *invitation.Invitation) error {
	query := `
		INSERT INTO invitations (id, organization_id, email, role, status, token, inviter_notes, invitee_notes, invited_by, invited_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.getQuerier(ctx).ExecContext(ctx, query,
		invite.ID,
		invite.OrganizationID,
		invite.Email,
		string(invite.Role),
		string(invite.Status),
		invite.Token,
		nullStringPtr(invite.InviterNotes),
		nullStringPtr(invite.InviteeNotes),
		invite.InvitedBy,
		invite.InvitedAt.UnixMilli(),
		invite.ExpiresAt.UnixMilli(),
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return invitation.ErrAlreadyExists
		}
		return err
	}

	return nil
}

// FindByID retrieves an invitation by ID.
func (s *InvitationStorage) FindByID(ctx context.Context, id string) (*invitation.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.id = ?`

	return s.scanInvitation(s.getQuerier(ctx).QueryRowContext(ctx, query, id))
}

// FindByToken retrieves an invitation by its secure token.
func (s *InvitationStorage) FindByToken(ctx context.Context, token string) (*invitation.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.token = ?`

	return s.scanInvitation(s.getQuerier(ctx).QueryRowContext(ctx, query, token))
}

// FindPendingByEmail finds all pending invitations for an email.
func (s *InvitationStorage) FindPendingByEmail(ctx context.Context, email string) ([]*invitation.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.email = ? AND i.status = 'pending' AND i.expires_at > ?
		ORDER BY i.invited_at DESC`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, email, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanInvitations(rows)
}

// FindPendingByEmailAndOrg finds a pending invitation for an email and organization.
func (s *InvitationStorage) FindPendingByEmailAndOrg(ctx context.Context, email, orgID string) (*invitation.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.email = ? AND i.organization_id = ? AND i.status = 'pending' AND i.expires_at > ?`

	return s.scanInvitation(s.getQuerier(ctx).QueryRowContext(ctx, query, email, orgID, time.Now().UnixMilli()))
}

// Update updates an invitation.
func (s *InvitationStorage) Update(ctx context.Context, invite *invitation.Invitation) error {
	query := `
		UPDATE invitations
		SET status = ?, invitee_notes = ?, responded_at = ?, expires_at = ?
		WHERE id = ?`

	result, err := s.getQuerier(ctx).ExecContext(ctx, query,
		string(invite.Status),
		nullStringPtr(invite.InviteeNotes),
		nullTimePtr(invite.RespondedAt),
		invite.ExpiresAt.UnixMilli(),
		invite.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return invitation.ErrNotFound
	}

	return nil
}

// ListByOrganization lists all invitations for an organization.
func (s *InvitationStorage) ListByOrganization(ctx context.Context, orgID string, filter *invitation.InvitationFilter) ([]*invitation.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.organization_id = ?`
	args := []interface{}{orgID}

	if filter != nil && filter.Status != nil {
		query += " AND i.status = ?"
		args = append(args, string(*filter.Status))
	}

	query += " ORDER BY i.invited_at DESC"

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanInvitations(rows)
}

// ListByInviter lists all invitations sent by an operator.
func (s *InvitationStorage) ListByInviter(ctx context.Context, inviterID string) ([]*invitation.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.invited_by = ?
		ORDER BY i.invited_at DESC`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, inviterID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanInvitations(rows)
}

// Delete deletes an invitation.
func (s *InvitationStorage) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM invitations WHERE id = ?`

	result, err := s.getQuerier(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return invitation.ErrNotFound
	}

	return nil
}

// ExpireByOrganization expires all pending invitations for an organization.
func (s *InvitationStorage) ExpireByOrganization(ctx context.Context, orgID string) error {
	query := `
		UPDATE invitations
		SET status = 'expired'
		WHERE organization_id = ? AND status = 'pending' AND expires_at <= ?`

	_, err := s.getQuerier(ctx).ExecContext(ctx, query, orgID, time.Now().UnixMilli())
	return err
}

// scanInvitation scans a single invitation from a row.
func (s *InvitationStorage) scanInvitation(row *sql.Row) (*invitation.Invitation, error) {
	var invite invitation.Invitation
	var inviterNotes, inviteeNotes sql.NullString
	var respondedAt sql.NullInt64
	var organizationName, inviterName sql.NullString

	err := row.Scan(
		&invite.ID,
		&invite.OrganizationID,
		&invite.Email,
		&invite.Role,
		&invite.Status,
		&invite.Token,
		&inviterNotes,
		&inviteeNotes,
		&invite.InvitedBy,
		&invite.InvitedAt,
		&respondedAt,
		&invite.ExpiresAt,
		&organizationName,
		&inviterName,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, invitation.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if inviterNotes.Valid {
		invite.InviterNotes = &inviterNotes.String
	}
	if inviteeNotes.Valid {
		invite.InviteeNotes = &inviteeNotes.String
	}
	if respondedAt.Valid {
		t := time.UnixMilli(respondedAt.Int64)
		invite.RespondedAt = &t
	}
	invite.OrganizationName = organizationName.String
	invite.InviterName = inviterName.String

	return &invite, nil
}

// scanInvitations scans multiple invitations from rows.
func (s *InvitationStorage) scanInvitations(rows *sql.Rows) ([]*invitation.Invitation, error) {
	var invitations []*invitation.Invitation

	for rows.Next() {
		var invite invitation.Invitation
		var inviterNotes, inviteeNotes sql.NullString
		var respondedAt sql.NullInt64
		var organizationName, inviterName sql.NullString

		err := rows.Scan(
			&invite.ID,
			&invite.OrganizationID,
			&invite.Email,
			&invite.Role,
			&invite.Status,
			&invite.Token,
			&inviterNotes,
			&inviteeNotes,
			&invite.InvitedBy,
			&invite.InvitedAt,
			&respondedAt,
			&invite.ExpiresAt,
			&organizationName,
			&inviterName,
		)
		if err != nil {
			return nil, err
		}

		if inviterNotes.Valid {
			invite.InviterNotes = &inviterNotes.String
		}
		if inviteeNotes.Valid {
			invite.InviteeNotes = &inviteeNotes.String
		}
		if respondedAt.Valid {
			t := time.UnixMilli(respondedAt.Int64)
			invite.RespondedAt = &t
		}
		invite.OrganizationName = organizationName.String
		invite.InviterName = inviterName.String

		invitations = append(invitations, &invite)
	}

	return invitations, rows.Err()
}
