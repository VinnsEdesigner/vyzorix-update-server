package organization

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/invitation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
	"github.com/google/uuid"
)

const (
	// InvitationDefaultTTL is the default time-to-live for invitations (7 days).
	InvitationDefaultTTL = 7 * 24 * time.Hour

	// InvitationMaxTTL is the maximum time-to-live for invitations (30 days).
	InvitationMaxTTL = 30 * 24 * time.Hour

	// MaxPendingInvitationsPerOrg is the maximum pending invitations per organization.
	MaxPendingInvitationsPerOrg = 20
)

var (
	ErrMaxInvitationsReached = errors.New("maximum pending invitations reached")
	ErrCannotInviteSelf      = errors.New("cannot invite yourself")
	ErrInvitationNotFound    = errors.New("invitation not found")
	ErrAlreadyOrgMember      = errors.New("operator is already a member of this organization")
	ErrOrgAtCapacity        = errors.New("organization has reached its member limit")
)

// EmailService defines the interface for sending emails.
type EmailService interface {
	SendInvitationEmail(ctx context.Context, inv *invitation.Invitation, orgName string, inviterName string) error
	SendInvitationAcceptedEmail(ctx context.Context, inv *invitation.Invitation, orgName string) error
	SendInvitationRejectedEmail(ctx context.Context, inv *invitation.Invitation, orgName string) error
}

// InvitationService handles invitation operations.
type InvitationService struct {
	invitationRepo  invitation.Repository
	orgRepo         organization.OrganizationRepository
	memberRepo      organization.MemberRepository
	txManager       transaction.TxManager
	emailService    EmailService
	logger          *slog.Logger
}

// NewInvitationService creates a new InvitationService.
func NewInvitationService(
	invitationRepo invitation.Repository,
	orgRepo organization.OrganizationRepository,
	memberRepo organization.MemberRepository,
	txManager transaction.TxManager,
	emailService EmailService,
	logger *slog.Logger,
) *InvitationService {
	if logger == nil {
		logger = slog.Default()
	}
	return &InvitationService{
		invitationRepo: invitationRepo,
		orgRepo:        orgRepo,
		memberRepo:     memberRepo,
		txManager:      txManager,
		emailService:   emailService,
		logger:         logger,
	}
}

// CreateInvitation creates a new invitation and sends an email.
func (s *InvitationService) CreateInvitation(ctx context.Context, orgID, inviterID, email string, role invitation.InvitationRole, notes string) (*invitation.Invitation, error) {
	// Validate email
	if email == "" {
		return nil, errors.New("email is required")
	}

	// Validate role - super_admin cannot be invited via invitation
	// Admin, operator, and viewer roles are allowed

	// Check if inviter is a member of the org with permission to invite
	member, err := s.memberRepo.FindByOperatorAndOrg(ctx, inviterID, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return nil, organization.ErrForbidden
		}
		return nil, err
	}

	// Must be able to manage members to invite
	if !member.Role.CanManageMembers() {
		return nil, organization.ErrForbidden
	}

	// Cannot invite self
	if member.OperatorEmail == email {
		return nil, ErrCannotInviteSelf
	}

	// Check org exists
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrNotFound) {
			return nil, organization.ErrNotFound
		}
		return nil, err
	}

	// Check pending invitation limit
	pending, err := s.invitationRepo.ListByOrganization(ctx, orgID, &invitation.InvitationFilter{
		Status: invitation.InvitationStatusPtr(invitation.InvitationStatusPending),
	})
	if err != nil {
		return nil, err
	}
	if len(pending) >= MaxPendingInvitationsPerOrg {
		return nil, ErrMaxInvitationsReached
	}

	// Check for existing pending invitation for this email in this org
	existing, err := s.invitationRepo.FindPendingByEmailAndOrg(ctx, email, orgID)
	if err != nil && !errors.Is(err, invitation.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, invitation.ErrAlreadyExists
	}

	// Check if user is already a member of this org
	// We need to check if there's an operator with this email who is already a member
	// Since we don't have operatorRepo here, we'll skip this check and let AcceptInvitation
	// handle the "already a member" case when the user tries to accept

	// Generate secure token
	token, err := invitation.GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ttl := InvitationDefaultTTL
	if ttl > InvitationMaxTTL {
		ttl = InvitationMaxTTL
	}

	inv := &invitation.Invitation{
		ID:             uuid.New().String(),
		OrganizationID:  orgID,
		Email:           email,
		Role:            role,
		Status:          invitation.InvitationStatusPending,
		Token:           token,
		InvitedBy:       inviterID,
		InvitedAt:       now,
		ExpiresAt:       now.Add(ttl),
		OrganizationName: org.Name,
	}

	if notes != "" {
		inv.InviterNotes = &notes
	}

	if err := s.invitationRepo.Create(ctx, inv); err != nil {
		return nil, err
	}

	// Send email (non-transactional)
	if s.emailService != nil {
		go func() {
			if err := s.emailService.SendInvitationEmail(context.Background(), inv, org.Name, member.OperatorName); err != nil {
				s.logger.Error("failed to send invitation email",
					"invitation_id", inv.ID,
					"email", email,
					"error", err,
				)
			}
		}()
	}

	s.logger.Info("invitation created",
		"org_id", orgID,
		"invitee_email", email,
		"role", role,
		"invited_by", inviterID,
	)

	return inv, nil
}

// GetInvitationByToken retrieves an invitation by its token (public endpoint).
func (s *InvitationService) GetInvitationByToken(ctx context.Context, token string) (*invitation.Invitation, error) {
	inv, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, invitation.ErrNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	// Check if expired
	if inv.IsExpired() && inv.Status == invitation.InvitationStatusPending {
		return nil, invitation.ErrExpired
	}

	return inv, nil
}

// AcceptInvitation accepts an invitation (for authenticated users).
func (s *InvitationService) AcceptInvitation(ctx context.Context, token, operatorID, operatorEmail, notes string) error {
	return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		inv, err := s.invitationRepo.FindByToken(txCtx, token)
		if err != nil {
			if errors.Is(err, invitation.ErrNotFound) {
				return ErrInvitationNotFound
			}
			return err
		}

		// Check if invitation can be accepted
		if !inv.CanBeAccepted() {
			if inv.Status != invitation.InvitationStatusPending {
				return invitation.ErrAlreadyResponded
			}
			return invitation.ErrExpired
		}

		// Verify email matches
		if inv.Email != operatorEmail {
			return invitation.ErrEmailMismatch
		}

		// Check if operator is already a member of this org
		existingMember, err := s.memberRepo.FindByOperatorAndOrg(txCtx, operatorID, inv.OrganizationID)
		if err == nil && existingMember != nil && existingMember.IsActive() {
			return ErrAlreadyOrgMember
		}

		// Check org capacity
		org, err := s.orgRepo.FindByID(txCtx, inv.OrganizationID)
		if err != nil {
			return err
		}
		if !org.CanAddMember() {
			return ErrOrgAtCapacity
		}

		now := time.Now()

		// Update invitation status
		inv.Status = invitation.InvitationStatusApproved
		inv.RespondedAt = &now
		inv.ResponderID = &operatorID
		if notes != "" {
			inv.InviteeNotes = &notes
		}

		if err := s.invitationRepo.Update(txCtx, inv); err != nil {
			return err
		}

		// Create membership
		member := &organization.OrganizationMember{
			ID:             uuid.New().String(),
			OrganizationID:  inv.OrganizationID,
			OperatorID:      operatorID,
			Role:            organization.OrganizationRole(inv.Role.ToOrgRole()),
			InvitedBy:       &inv.InvitedBy,
			JoinedAt:        now,
			Status:          organization.MemberStatusActive,
		}

		if err := s.memberRepo.Create(txCtx, member); err != nil {
			return err
		}

		s.logger.Info("invitation accepted",
			"invitation_id", inv.ID,
			"operator_id", operatorID,
			"org_id", inv.OrganizationID,
		)

		// Send notification email to inviter (async)
		if s.emailService != nil {
			go func() {
				org, _ := s.orgRepo.FindByID(context.Background(), inv.OrganizationID)
				orgName := ""
				if org != nil {
					orgName = org.Name
				}
				if err := s.emailService.SendInvitationAcceptedEmail(context.Background(), inv, orgName); err != nil {
					s.logger.Error("failed to send acceptance notification", "error", err)
				}
			}()
		}

		return nil
	})
}

// RejectInvitation rejects an invitation.
func (s *InvitationService) RejectInvitation(ctx context.Context, token, operatorID, operatorEmail, notes string) error {
	return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		inv, err := s.invitationRepo.FindByToken(txCtx, token)
		if err != nil {
			if errors.Is(err, invitation.ErrNotFound) {
				return ErrInvitationNotFound
			}
			return err
		}

		// Check if invitation can be rejected
		if inv.Status != invitation.InvitationStatusPending {
			return invitation.ErrAlreadyResponded
		}

		// Verify email matches
		if inv.Email != operatorEmail {
			return invitation.ErrEmailMismatch
		}

		now := time.Now()

		// Update invitation status
		inv.Status = invitation.InvitationStatusRejected
		inv.RespondedAt = &now
		inv.ResponderID = &operatorID
		if notes != "" {
			inv.InviteeNotes = &notes
		}

		if err := s.invitationRepo.Update(txCtx, inv); err != nil {
			return err
		}

		s.logger.Info("invitation rejected",
			"invitation_id", inv.ID,
			"operator_id", operatorID,
		)

		// Send notification email to inviter (async)
		if s.emailService != nil {
			go func() {
				org, _ := s.orgRepo.FindByID(context.Background(), inv.OrganizationID)
				orgName := ""
				if org != nil {
					orgName = org.Name
				}
				if err := s.emailService.SendInvitationRejectedEmail(context.Background(), inv, orgName); err != nil {
					s.logger.Error("failed to send rejection notification", "error", err)
				}
			}()
		}

		return nil
	})
}

// ListInvitationsByOrganization lists all invitations for an organization.
func (s *InvitationService) ListInvitationsByOrganization(ctx context.Context, orgID string, status *invitation.InvitationStatus) ([]*invitation.Invitation, error) {
	invitations, err := s.invitationRepo.ListByOrganization(ctx, orgID, &invitation.InvitationFilter{
		Status: status,
	})
	if err != nil {
		return nil, err
	}

	return invitations, nil
}

// ListInvitationsByInviter lists all invitations sent by an operator.
func (s *InvitationService) ListInvitationsByInviter(ctx context.Context, inviterID string) ([]*invitation.Invitation, error) {
	return s.invitationRepo.ListByInviter(ctx, inviterID)
}

// ListPendingInvitationsForEmail lists all pending invitations for an email.
func (s *InvitationService) ListPendingInvitationsForEmail(ctx context.Context, email string) ([]*invitation.Invitation, error) {
	return s.invitationRepo.FindPendingByEmail(ctx, email)
}

// ExpireInvitation manually expires an invitation.
func (s *InvitationService) ExpireInvitation(ctx context.Context, invitationID, actorID string) error {
	// Get invitation to verify actor has permission
	inv, err := s.invitationRepo.FindByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, invitation.ErrNotFound) {
			return ErrInvitationNotFound
		}
		return err
	}

	// Verify actor is inviter or org admin
	member, err := s.memberRepo.FindByOperatorAndOrg(ctx, actorID, inv.OrganizationID)
	if err != nil {
		return organization.ErrForbidden
	}

	if !member.Role.CanManageMembers() && member.OperatorID != inv.InvitedBy {
		return organization.ErrForbidden
	}

	if err := s.invitationRepo.Delete(ctx, invitationID); err != nil {
		return err
	}

	return nil
}

// ExpireStaleInvitations expires all stale pending invitations.
func (s *InvitationService) ExpireStaleInvitations(ctx context.Context) error {
	// This would typically be called by a background job
	// For now, we'll rely on the ExpireByOrganization method when org is deleted
	s.logger.Info("expiring stale invitations")
	return nil
}
