package organization

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
	emailSvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/google/uuid"
)

// InvitationDefaultTTL is the default time-to-live for invitations (7 days).
const InvitationDefaultTTL = 7 * 24 * time.Hour

// InvitationMaxTTL is the maximum time-to-live for invitations (30 days).
const InvitationMaxTTL = 30 * 24 * time.Hour

// MaxPendingInvitationsPerOrg is the maximum pending invitations per organization.
const MaxPendingInvitationsPerOrg = 20

var (
	ErrMaxInvitationsReached = errors.New("maximum pending invitations reached")
	ErrCannotInviteSelf      = errors.New("cannot invite yourself")
	ErrInvitationNotFound    = errors.New("invitation not found")
	ErrAlreadyOrgMember      = errors.New("operator is already a member of this organization")
	ErrOrgAtCapacity         = errors.New("organization has reached its member limit")
)

// isUniqueConstraintError checks if the error is a unique constraint violation.
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite unique constraint error message contains "UNIQUE constraint failed".
	errStr := err.Error()
	return strings.Contains(errStr, "UNIQUE constraint failed") || strings.Contains(errStr, "unique constraint")
}

// EmailService defines the interface for sending emails.
type EmailService interface {
	SendInvitationEmail(ctx context.Context, to string, data emailSvc.InvitationData) error
	SendInvitationAcceptedEmail(ctx context.Context, to string, data emailSvc.InvitationData) error
	SendInvitationRejectedEmail(ctx context.Context, to string, data emailSvc.InvitationData) error
}

// InvitationService handles invitation operations.
type InvitationService struct {
	invitationRepo organization.InvitationRepository
	orgRepo        organization.OrganizationRepository
	memberRepo     organization.MemberRepository
	txManager      transaction.TxManager
	emailService   EmailService
	logger         *slog.Logger
	baseURL        string
	emailWg        sync.WaitGroup // tracks background email goroutines.
}

// NewInvitationService creates a new InvitationService.
func NewInvitationService(
	invitationRepo organization.InvitationRepository,
	orgRepo organization.OrganizationRepository,
	memberRepo organization.MemberRepository,
	txManager transaction.TxManager,
	emailService EmailService,
	logger *slog.Logger,
	baseURL string,
) *InvitationService {
	if logger == nil {
		logger = slog.Default()
	}
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	return &InvitationService{
		invitationRepo: invitationRepo,
		orgRepo:        orgRepo,
		memberRepo:     memberRepo,
		txManager:      txManager,
		emailService:   emailService,
		logger:         logger,
		baseURL:        baseURL,
	}
}

// Shutdown waits for all pending email goroutines to complete.
// Call this during graceful shutdown to ensure no emails are dropped.
func (s *InvitationService) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.emailWg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// getBaseURL returns the base URL for the application.
func (s *InvitationService) getBaseURL() string {
	return s.baseURL
}

// CreateInvitation creates a new invitation and sends an email.
func (s *InvitationService) CreateInvitation(ctx context.Context, orgID, inviterID, email string, role organization.OrganizationRole, notes string) (*organization.Invitation, error) {
	// Validate inputs and check permissions.
	member, org, inv, err := s.validateAndPrepareInvitation(ctx, orgID, inviterID, email, role)
	if err != nil {
		return nil, err
	}

	if notes != "" {
		inv.InviterNotes = notes
	}

	if err := s.invitationRepo.Create(ctx, inv); err != nil {
		return nil, err
	}

	s.sendInvitationEmailAsync(inv, email, member.OperatorName, org.Name)
	s.logInvitationCreated(orgID, email, role, inviterID)

	return inv, nil
}

// validateAndPrepareInvitation performs validation and prepares invitation data.
func (s *InvitationService) validateAndPrepareInvitation(ctx context.Context, orgID, inviterID, email string, role organization.OrganizationRole) (*organization.OrganizationMember, *organization.Organization, *organization.Invitation, error) {
	if email == "" {
		return nil, nil, nil, errors.New("email is required")
	}

	// Note: super_admin cannot be invited via invitation.
	if role != organization.RoleAdmin && role != organization.RoleOperator && role != organization.RoleViewer {
		return nil, nil, nil, errors.New("invalid invitation role")
	}

	member, memberErr := s.memberRepo.FindByOperatorAndOrg(ctx, inviterID, orgID)
	if memberErr != nil {
		if errors.Is(memberErr, organization.ErrMemberNotFound) {
			return nil, nil, nil, organization.ErrForbidden
		}
		return nil, nil, nil, memberErr
	}

	if !member.Role.CanManageMembers() {
		return nil, nil, nil, organization.ErrForbidden
	}

	if member.OperatorEmail == email {
		return nil, nil, nil, ErrCannotInviteSelf
	}

	org, orgErr := s.orgRepo.FindByID(ctx, orgID)
	if orgErr != nil {
		if errors.Is(orgErr, organization.ErrNotFound) {
			return nil, nil, nil, organization.ErrNotFound
		}
		return nil, nil, nil, orgErr
	}

	if limitErr := s.checkInvitationLimits(ctx, orgID, email); limitErr != nil {
		return nil, nil, nil, limitErr
	}

	inv, invErr := organization.NewInvitation(uuid.New().String(), orgID, email, role, inviterID)
	if invErr != nil {
		return nil, nil, nil, invErr
	}

	return member, org, inv, nil
}

// checkInvitationLimits checks pending invitation limits.
func (s *InvitationService) checkInvitationLimits(ctx context.Context, orgID, email string) error {
	pending, err := s.invitationRepo.FindPendingByOrganization(ctx, orgID)
	if err != nil {
		return err
	}
	if len(pending) >= MaxPendingInvitationsPerOrg {
		return ErrMaxInvitationsReached
	}

	existing, err := s.invitationRepo.FindPendingByOrganizationAndEmail(ctx, orgID, email)
	if err != nil && !errors.Is(err, organization.ErrInvitationNotFound) {
		return err
	}
	if len(existing) > 0 {
		return organization.ErrInvitationExists
	}
	return nil
}

// sendInvitationEmailAsync sends the invitation email asynchronously.
func (s *InvitationService) sendInvitationEmailAsync(invitation *organization.Invitation, inviteeEmail, memberName, orgName string) {
	if s.emailService == nil {
		return
	}
	s.emailWg.Add(1)
	go func() {
		defer s.emailWg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		inviteData := emailSvc.InvitationData{
			InviteeName:      inviteeEmail,
			InviterName:      memberName,
			OrganizationName: orgName,
			Role:             string(invitation.Role),
			InviterNotes:     invitation.InviterNotes,
			AcceptURL:        fmt.Sprintf("%s/invite/%s/accept", s.getBaseURL(), invitation.Token),
			ExpiryDays:       7,
			BaseURL:          s.getBaseURL(),
		}
		if err := s.emailService.SendInvitationEmail(ctx, inviteeEmail, inviteData); err != nil {
			s.logger.Error("failed to send invitation email",
				"invitation_id", invitation.ID, "email", inviteeEmail, "error", err)
		}
	}()
}

// logInvitationCreated logs the invitation creation.
func (s *InvitationService) logInvitationCreated(orgID, email string, role organization.OrganizationRole, inviterID string) {
	s.logger.Info("invitation created",
		"org_id", orgID, "invitee_email", email, "role", role, "invited_by", inviterID)
}

// GetInvitationByToken retrieves an invitation by its token (public endpoint).
func (s *InvitationService) GetInvitationByToken(ctx context.Context, token string) (*organization.Invitation, error) {
	inv, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, organization.ErrInvitationNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	// Check if expired using lifecycle method.
	if inv.IsExpired() {
		return nil, organization.ErrInvitationExpired
	}

	return inv, nil
}

// GetInvitationByID retrieves an invitation by its ID.
func (s *InvitationService) GetInvitationByID(ctx context.Context, invitationID string) (*organization.Invitation, error) {
	inv, err := s.invitationRepo.FindByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, organization.ErrInvitationNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}
	return inv, nil
}

// CancelInvitation cancels/deletes an invitation.
func (s *InvitationService) CancelInvitation(ctx context.Context, invitationID string) error {
	return s.invitationRepo.Delete(ctx, invitationID)
}

// AcceptInvitation accepts an invitation (for authenticated users).
func (s *InvitationService) AcceptInvitation(ctx context.Context, token, operatorID, operatorEmail, notes string) (*organization.OrganizationMember, error) {
	var resultMember *organization.OrganizationMember
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		inv, err := s.invitationRepo.FindByToken(txCtx, token)
		if err != nil {
			if errors.Is(err, organization.ErrInvitationNotFound) {
				return ErrInvitationNotFound
			}
			return err
		}

		// Check if invitation can be accepted using lifecycle method.
		if !inv.CanBeAccepted() {
			if inv.HasResponded() {
				return organization.ErrAlreadyResponded
			}
			return organization.ErrInvitationExpired
		}

		// Verify email matches.
		if inv.Email != operatorEmail {
			return organization.ErrEmailMismatch
		}

		// Check if operator is already a member of this org.
		existingMember, err := s.memberRepo.FindByOperatorAndOrg(txCtx, operatorID, inv.OrganizationID)
		if err == nil && existingMember != nil && existingMember.IsActive() {
			return ErrAlreadyOrgMember
		}

		// Check org capacity.
		org, err := s.orgRepo.FindByID(txCtx, inv.OrganizationID)
		if err != nil {
			return err
		}
		if !org.CanAddMember() {
			return ErrOrgAtCapacity
		}

		// Use lifecycle Accept() method to properly transition invitation state.
		if err := inv.Accept(operatorID, notes); err != nil {
			return err
		}

		if err := s.invitationRepo.Update(txCtx, inv); err != nil {
			return err
		}

		// Create membership using domain constructor.
		member := organization.NewMember(
			uuid.New().String(),
			inv.OrganizationID,
			operatorID,
			inv.Role, // Role is already OrganizationRole.
		)
		if inv.InvitedBy != "" {
			member.InvitedBy = &inv.InvitedBy
		}

		if err := s.memberRepo.Create(txCtx, member); err != nil {
			// This can happen in a race condition where two concurrent AcceptInvitation.
			// calls both pass the CanBeAccepted check before either creates the membership.
			if isUniqueConstraintError(err) {
				return ErrAlreadyOrgMember
			}
			return err
		}

		s.logger.Info("invitation accepted",
			"invitation_id", inv.ID,
			"operator_id", operatorID,
			"org_id", inv.OrganizationID,
		)

		// Store member reference for return value.
		resultMember = member

		// Send notification email to inviter (async).
		if s.emailService != nil {
			inv := inv // capture loop var.
			opEmail := operatorEmail
			invNotes := notes
			invOrgID := inv.OrganizationID
			invRole := inv.Role
			invInviterEmail := inv.InviterEmail
			s.emailWg.Add(1)
			go func() {
				defer s.emailWg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				org, _ := s.orgRepo.FindByID(ctx, invOrgID)
				orgName := ""
				if org != nil {
					orgName = org.Name
				}
				inviteData := emailSvc.InvitationData{
					InviteeName:      opEmail,
					OrganizationName: orgName,
					Role:             string(invRole),
					InviteeNotes:     invNotes,
					AcceptedAt:       time.Now().Format("2006-01-02 15:04:05"),
					BaseURL:          s.getBaseURL(),
				}
				if invInviterEmail != "" {
					if err := s.emailService.SendInvitationAcceptedEmail(ctx, invInviterEmail, inviteData); err != nil {
						s.logger.Error("failed to send acceptance notification", "error", err)
					}
				}
			}()
		}

		return nil
	})

	return resultMember, err
}

// RejectInvitation rejects an invitation.
func (s *InvitationService) RejectInvitation(ctx context.Context, token, operatorID, operatorEmail, notes string) error {
	return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		inv, err := s.invitationRepo.FindByToken(txCtx, token)
		if err != nil {
			if errors.Is(err, organization.ErrInvitationNotFound) {
				return ErrInvitationNotFound
			}
			return err
		}

		// Check if invitation can be rejected (must be pending).
		if !inv.IsPending() {
			return organization.ErrAlreadyResponded
		}

		// Verify email matches.
		if inv.Email != operatorEmail {
			return organization.ErrEmailMismatch
		}

		// Use lifecycle Reject() method to properly transition invitation state.
		if err := inv.Reject(operatorID, notes); err != nil {
			return err
		}

		if err := s.invitationRepo.Update(txCtx, inv); err != nil {
			return err
		}

		s.logger.Info("invitation rejected",
			"invitation_id", inv.ID,
			"operator_id", operatorID,
		)

		// Send notification email to inviter (async).
		if s.emailService != nil {
			inv := inv // capture loop var.
			opEmail := operatorEmail
			invNotes := notes
			invOrgID := inv.OrganizationID
			invRole := inv.Role
			invInviterEmail := inv.InviterEmail
			s.emailWg.Add(1)
			go func() {
				defer s.emailWg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				org, _ := s.orgRepo.FindByID(ctx, invOrgID)
				orgName := ""
				if org != nil {
					orgName = org.Name
				}
				inviteData := emailSvc.InvitationData{
					InviteeName:      opEmail,
					OrganizationName: orgName,
					Role:             string(invRole),
					InviteeNotes:     invNotes,
					BaseURL:          s.getBaseURL(),
				}
				if invInviterEmail != "" {
					if err := s.emailService.SendInvitationRejectedEmail(ctx, invInviterEmail, inviteData); err != nil {
						s.logger.Error("failed to send rejection notification", "error", err)
					}
				}
			}()
		}

		return nil
	})
}

// ListInvitationsByOrganization lists all invitations for an organization.
func (s *InvitationService) ListInvitationsByOrganization(ctx context.Context, orgID string, status *organization.InvitationStatus) ([]*organization.Invitation, error) {
	var filter *organization.InvitationFilter
	if status != nil {
		filter = &organization.InvitationFilter{Status: status}
	}

	// If no filter, get all invitations.
	if filter == nil {
		return s.invitationRepo.FindByOrganization(ctx, orgID)
	}

	// Use paginated method with filter.
	invitations, _, err := s.invitationRepo.FindByOrganizationPaginated(ctx, orgID, 1000, 0, filter)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// ListInvitationsByOrganizationPaginated lists invitations with pagination.
func (s *InvitationService) ListInvitationsByOrganizationPaginated(ctx context.Context, orgID string, page, limit int, status *organization.InvitationStatus) (*InvitationListResponse, error) {
	// Apply defaults and limits.
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	offset := (page - 1) * limit

	var filter *organization.InvitationFilter
	if status != nil {
		filter = &organization.InvitationFilter{Status: status}
	}

	// Use repository paginated method - does LIMIT/OFFSET at DB level.
	invitations, total, err := s.invitationRepo.FindByOrganizationPaginated(ctx, orgID, limit, offset, filter)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return &InvitationListResponse{
		Items: invitations,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasMore:    page < totalPages,
		},
	}, nil
}

// ListInvitationsByInviter lists all invitations sent by an operator.
func (s *InvitationService) ListInvitationsByInviter(ctx context.Context, inviterID string) ([]*organization.Invitation, error) {
	return s.invitationRepo.ListByInviter(ctx, inviterID)
}

// ListPendingInvitationsForEmail lists all pending invitations for an email.
func (s *InvitationService) ListPendingInvitationsForEmail(ctx context.Context, email string) ([]*organization.Invitation, error) {
	return s.invitationRepo.FindPendingByEmail(ctx, email)
}

// ExpireInvitation manually expires an invitation.
func (s *InvitationService) ExpireInvitation(ctx context.Context, invitationID, actorID string) error {
	// Get invitation to verify actor has permission.
	inv, err := s.invitationRepo.FindByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, organization.ErrInvitationNotFound) {
			return ErrInvitationNotFound
		}
		return err
	}

	// Verify actor is inviter or org admin.
	member, err := s.memberRepo.FindByOperatorAndOrg(ctx, actorID, inv.OrganizationID)
	if err != nil {
		return organization.ErrForbidden
	}

	if !member.Role.CanManageMembers() && member.OperatorID != inv.InvitedBy {
		return organization.ErrForbidden
	}

	// Use lifecycle Expire() method to properly transition invitation state.
	if err := inv.Expire(); err != nil {
		return err
	}

	if err := s.invitationRepo.Update(ctx, inv); err != nil {
		return err
	}

	s.logger.Info("invitation expired",
		"invitation_id", invitationID,
		"actor_id", actorID,
	)

	return nil
}

// ExpireStaleInvitations expires all stale pending invitations.
func (s *InvitationService) ExpireStaleInvitations(ctx context.Context) error {
	// This would typically be called by a background job.
	// For now, we'll rely on the ExpireByOrganization method when org is deleted.
	s.logger.Info("expiring stale invitations")
	return nil
}
