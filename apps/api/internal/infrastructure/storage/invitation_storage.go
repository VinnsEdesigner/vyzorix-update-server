package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Ensure InvitationRepository implements organization.InvitationRepository.
var _ organization.InvitationRepository = (*InvitationStorage)(nil)

// InvitationStorage implements organization.InvitationRepository using SQLite.
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
func (s *InvitationStorage) Create(ctx context.Context, invite *organization.Invitation) error {
	query := `
		INSERT INTO invitations (id, organization_id, email, role, status, token, inviter_notes, invitee_notes, invited_by, invited_at, responder_id, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.getQuerier(ctx).ExecContext(ctx, query,
		invite.ID,
		invite.OrganizationID,
		invite.Email,
		string(invite.Role),
		string(invite.Status),
		invite.Token,
		nullString(invite.InviterNotes),
		nullString(invite.InviteeNotes),
		invite.InvitedBy,
		invite.InvitedAt.UnixMilli(),
		nullString(invite.RespondedBy),
		invite.ExpiresAt.UnixMilli(),
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return organization.ErrInvitationExists
		}
		return err
	}

	return nil
}

// FindByID retrieves an invitation by ID.
func (s *InvitationStorage) FindByID(ctx context.Context, id string) (*organization.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.responder_id, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name, COALESCE(op.email, '') as inviter_email
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.id = ?`

	return s.scanInvitation(s.getQuerier(ctx).QueryRowContext(ctx, query, id))
}

// FindByToken retrieves an invitation by its secure token.
func (s *InvitationStorage) FindByToken(ctx context.Context, token string) (*organization.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.responder_id, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name, COALESCE(op.email, '') as inviter_email
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.token = ?`

	return s.scanInvitation(s.getQuerier(ctx).QueryRowContext(ctx, query, token))
}

// FindByEmail retrieves all invitations for an email across all organizations.
func (s *InvitationStorage) FindByEmail(ctx context.Context, email string) ([]*organization.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.responder_id, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name, COALESCE(op.email, '') as inviter_email
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.email = ?
		ORDER BY i.invited_at DESC`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, email)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanInvitations(rows)
}

// FindPendingByEmail finds all pending invitations for an email.
func (s *InvitationStorage) FindPendingByEmail(ctx context.Context, email string) ([]*organization.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.responder_id, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name, COALESCE(op.email, '') as inviter_email
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

// FindByOrganization retrieves all invitations for an organization.
func (s *InvitationStorage) FindByOrganization(ctx context.Context, orgID string) ([]*organization.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.responder_id, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name, COALESCE(op.email, '') as inviter_email
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.organization_id = ?
		ORDER BY i.invited_at DESC`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanInvitations(rows)
}

// FindPendingByOrganization retrieves all pending invitations for an organization.
func (s *InvitationStorage) FindPendingByOrganization(ctx context.Context, orgID string) ([]*organization.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.responder_id, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name, COALESCE(op.email, '') as inviter_email
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.organization_id = ? AND i.status = 'pending' AND i.expires_at > ?
		ORDER BY i.invited_at DESC`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, orgID, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanInvitations(rows)
}

// FindByOrganizationAndEmail retrieves invitations by organization and email.
func (s *InvitationStorage) FindByOrganizationAndEmail(ctx context.Context, orgID, email string) ([]*organization.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.responder_id, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name, COALESCE(op.email, '') as inviter_email
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.organization_id = ? AND i.email = ?
		ORDER BY i.invited_at DESC`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, orgID, email)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanInvitations(rows)
}

// FindPendingByOrganizationAndEmail retrieves pending invitations by org and email.
func (s *InvitationStorage) FindPendingByOrganizationAndEmail(ctx context.Context, orgID, email string) ([]*organization.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.responder_id, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name, COALESCE(op.email, '') as inviter_email
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.organization_id = ? AND i.email = ? AND i.status = 'pending' AND i.expires_at > ?`

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, orgID, email, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanInvitations(rows)
}

// FindByOrganizationPaginated retrieves invitations with pagination.
func (s *InvitationStorage) FindByOrganizationPaginated(ctx context.Context, orgID string, limit, offset int, filter *organization.InvitationFilter) ([]*organization.Invitation, int, error) {
	// Build count query.
	countQuery := `SELECT COUNT(*) FROM invitations WHERE organization_id = ?`
	args := []interface{}{orgID}

	if filter != nil && filter.Status != nil {
		countQuery += " AND status = ?"
		args = append(args, string(*filter.Status))
	}

	var total int
	err := s.getQuerier(ctx).QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Build paginated query.
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.responder_id, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name, COALESCE(op.email, '') as inviter_email
		FROM invitations i
		LEFT JOIN organizations o ON i.organization_id = o.id
		LEFT JOIN operators op ON i.invited_by = op.id
		WHERE i.organization_id = ?`

	scanArgs := []interface{}{orgID}

	if filter != nil && filter.Status != nil {
		query += " AND i.status = ?"
		scanArgs = append(scanArgs, string(*filter.Status))
	}

	query += " ORDER BY i.invited_at DESC LIMIT ? OFFSET ?"
	scanArgs = append(scanArgs, limit, offset)

	rows, err := s.getQuerier(ctx).QueryContext(ctx, query, scanArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	invitations, err := s.scanInvitations(rows)
	if err != nil {
		return nil, 0, err
	}

	return invitations, total, nil
}

// ListByInviter lists all invitations sent by an operator.
func (s *InvitationStorage) ListByInviter(ctx context.Context, inviterID string) ([]*organization.Invitation, error) {
	query := `
		SELECT i.id, i.organization_id, i.email, i.role, i.status, i.token, i.inviter_notes, i.invitee_notes, 
			   i.invited_by, i.invited_at, i.responded_at, i.responder_id, i.expires_at,
			   COALESCE(o.name, '') as organization_name, COALESCE(op.name, '') as inviter_name, COALESCE(op.email, '') as inviter_email
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

// Update updates an invitation.
func (s *InvitationStorage) Update(ctx context.Context, invite *organization.Invitation) error {
	query := `
		UPDATE invitations
		SET status = ?, invitee_notes = ?, responded_at = ?, responder_id = ?, expires_at = ?
		WHERE id = ?`

	result, err := s.getQuerier(ctx).ExecContext(ctx, query,
		string(invite.Status),
		nullString(invite.InviteeNotes),
		nullTimePtr(invite.RespondedAt),
		nullString(invite.RespondedBy),
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
		return organization.ErrInvitationNotFound
	}

	return nil
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
		return organization.ErrInvitationNotFound
	}

	return nil
}

// SoftDelete soft-deletes an invitation.
func (s *InvitationStorage) SoftDelete(ctx context.Context, id string) error {
	// For now, just hard delete. Can be enhanced later with soft delete column.
	return s.Delete(ctx, id)
}

// SoftDeleteByToken soft-deletes an invitation by token.
func (s *InvitationStorage) SoftDeleteByToken(ctx context.Context, token string) error {
	query := `DELETE FROM invitations WHERE token = ?`

	result, err := s.getQuerier(ctx).ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return organization.ErrInvitationNotFound
	}

	return nil
}

// SoftDeleteExpired soft-deletes all expired invitations.
func (s *InvitationStorage) SoftDeleteExpired(ctx context.Context) error {
	query := `DELETE FROM invitations WHERE status = 'pending' AND expires_at <= ?`

	_, err := s.getQuerier(ctx).ExecContext(ctx, query, time.Now().UnixMilli())
	return err
}

// SoftDeleteByOrganization soft-deletes all invitations for an organization.
func (s *InvitationStorage) SoftDeleteByOrganization(ctx context.Context, orgID string) error {
	query := `DELETE FROM invitations WHERE organization_id = ?`

	_, err := s.getQuerier(ctx).ExecContext(ctx, query, orgID)
	return err
}

// SoftDeleteByInvitedBy soft-deletes all invitations sent by an operator.
func (s *InvitationStorage) SoftDeleteByInvitedBy(ctx context.Context, operatorID string) error {
	query := `DELETE FROM invitations WHERE invited_by = ?`

	_, err := s.getQuerier(ctx).ExecContext(ctx, query, operatorID)
	return err
}

// CountByOrganization counts invitations for an organization.
func (s *InvitationStorage) CountByOrganization(ctx context.Context, orgID string) (int, error) {
	query := `SELECT COUNT(*) FROM invitations WHERE organization_id = ?`

	var count int
	err := s.getQuerier(ctx).QueryRowContext(ctx, query, orgID).Scan(&count)
	return count, err
}

// CountPendingByOrganization counts pending invitations for an organization.
func (s *InvitationStorage) CountPendingByOrganization(ctx context.Context, orgID string) (int, error) {
	query := `SELECT COUNT(*) FROM invitations WHERE organization_id = ? AND status = 'pending' AND expires_at > ?`

	var count int
	err := s.getQuerier(ctx).QueryRowContext(ctx, query, orgID, time.Now().UnixMilli()).Scan(&count)
	return count, err
}

// ExpireByOrganization expires all pending invitations for an organization.
func (s *InvitationStorage) ExpireByOrganization(ctx context.Context, orgID string) error {
	query := `UPDATE invitations SET status = 'expired' WHERE organization_id = ? AND status = 'pending'`

	_, err := s.getQuerier(ctx).ExecContext(ctx, query, orgID)
	return err
}

// ExpireOldThan soft-deletes invitations older than the given duration.
func (s *InvitationStorage) ExpireOldThan(ctx context.Context, duration string) error {
	query := `DELETE FROM invitations WHERE status = 'pending' AND expires_at < ?`

	// Parse duration and calculate cutoff time.
	d, err := time.ParseDuration(duration)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-d)
	_, err = s.getQuerier(ctx).ExecContext(ctx, query, cutoff.UnixMilli())
	return err
}

// statusToLifecycle converts status string to InvitationLifecycle.
// This is needed because the database only stores status, but the domain model uses lifecycle.
func statusToLifecycle(status organization.InvitationStatus) organization.InvitationLifecycle {
	switch status {
	case organization.InvitationStatusApproved:
		return organization.InvitationLifecycleAccepted
	case organization.InvitationStatusRejected:
		return organization.InvitationLifecycleRejected
	case organization.InvitationStatusExpired:
		return organization.InvitationLifecycleExpired
	default:
		return organization.InvitationLifecyclePending
	}
}

// scanInvitation scans a single invitation from a row.
func (s *InvitationStorage) scanInvitation(row *sql.Row) (*organization.Invitation, error) {
	var invite organization.Invitation
	var inviterNotes, inviteeNotes, responderID sql.NullString
	var respondedAt sql.NullInt64
	var organizationName, inviterName, inviterEmail sql.NullString

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
		&responderID,
		&invite.ExpiresAt,
		&organizationName,
		&inviterName,
		&inviterEmail,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, organization.ErrInvitationNotFound
	}
	if err != nil {
		return nil, err
	}

	if inviterNotes.Valid && inviterNotes.String != "" {
		invite.InviterNotes = inviterNotes.String
	}
	if inviteeNotes.Valid && inviteeNotes.String != "" {
		invite.InviteeNotes = inviteeNotes.String
	}
	if respondedAt.Valid {
		t := time.UnixMilli(respondedAt.Int64)
		invite.RespondedAt = &t
	}
	if responderID.Valid {
		invite.RespondedBy = responderID.String
	}
	invite.OrganizationName = organizationName.String
	invite.InviterName = inviterName.String
	invite.InviterEmail = inviterEmail.String

	// Derive lifecycle from status since database only stores status.
	invite.Lifecycle = statusToLifecycle(invite.Status)

	return &invite, nil
}

// scanInvitations scans multiple invitations from rows.
func (s *InvitationStorage) scanInvitations(rows *sql.Rows) ([]*organization.Invitation, error) {
	var invitations []*organization.Invitation

	for rows.Next() {
		var invite organization.Invitation
		var inviterNotes, inviteeNotes, responderID sql.NullString
		var respondedAt sql.NullInt64
		var organizationName, inviterName, inviterEmail sql.NullString

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
			&responderID,
			&invite.ExpiresAt,
			&organizationName,
			&inviterName,
			&inviterEmail,
		)
		if err != nil {
			return nil, err
		}

		if inviterNotes.Valid && inviterNotes.String != "" {
			invite.InviterNotes = inviterNotes.String
		}
		if inviteeNotes.Valid && inviteeNotes.String != "" {
			invite.InviteeNotes = inviteeNotes.String
		}
		if respondedAt.Valid {
			t := time.UnixMilli(respondedAt.Int64)
			invite.RespondedAt = &t
		}
		if responderID.Valid {
			invite.RespondedBy = responderID.String
		}
		invite.OrganizationName = organizationName.String
		invite.InviterName = inviterName.String
		invite.InviterEmail = inviterEmail.String

		// Derive lifecycle from status since database only stores status.
		invite.Lifecycle = statusToLifecycle(invite.Status)

		invitations = append(invitations, &invite)
	}

	return invitations, rows.Err()
}
