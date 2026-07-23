package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
)

// UpdateOperatorRequest represents a request to update an operator.
type UpdateOperatorRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Role  *string `json:"role,omitempty"`
}

// ListAllOperators returns all operators (admin use).
func (s *AuthService) ListAllOperators(ctx context.Context, limit, offset int) ([]dto.OperatorListResponse, int, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	operators, total, err := s.operatorRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	response := make([]dto.OperatorListResponse, len(operators))
	for i, op := range operators {
		// Note: Role is org-scoped, not per-operator. Use first membership role or "operator" as default.
		role := "operator"
		if len(op.Memberships) > 0 {
			role = string(op.Memberships[0].Role)
		}
		response[i] = dto.OperatorListResponse{
			ID:            op.ID,
			Email:         op.Email,
			Name:          op.Name,
			Role:          role,
			MFAEnabled:    op.MFASecret != "",
			EmailVerified: op.EmailVerified,
			CreatedAt:     op.CreatedAt.UnixMilli(),
		}
	}

	return response, total, nil
}

// CreateOperator creates a new operator (admin only).
func (s *AuthService) CreateOperator(ctx context.Context, req *dto.RegisterRequest) (*operator.Operator, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	existing, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if existing != nil {
		return nil, application.ErrEmailExists
	}

	passwordHash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	role := operator.RoleOperator
	if req.Role != "" {
		role = operator.OperatorRole(req.Role)
		if !role.IsValid() {
			return nil, application.ErrInvalidInput
		}
	}

	now := time.Now()
	id := shared.GenerateID()

	op := &operator.Operator{
		ID:           id,
		Email:        email,
		Name:         strings.TrimSpace(req.Name),
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.operatorRepo.Create(ctx, op); err != nil {
		return nil, err
	}

	return op, nil
}

// GetOperatorByID retrieves an operator by ID.
func (s *AuthService) GetOperatorByID(ctx context.Context, id string) (*operator.Operator, error) {
	return s.operatorRepo.FindByID(ctx, id)
}

// VerifyJWT verifies a JWT token and returns the claims.
func (s *AuthService) VerifyJWT(token string) (*infraauth.OperatorClaims, error) {
	if s.jwtManager == nil {
		return nil, infraauth.ErrInvalidToken
	}

	return s.jwtManager.Verify(token)
}

// UpdateOperator updates an existing operator (admin only).
func (s *AuthService) UpdateOperator(ctx context.Context, operatorID string, req *UpdateOperatorRequest) (*operator.Operator, error) {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if errors.Is(err, operator.ErrNotFound) {
			return nil, application.ErrOperatorNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, application.ErrInvalidInput
		}
		op.Name = name
	}

	if req.Email != nil {
		existing, err := s.operatorRepo.FindByEmail(ctx, *req.Email)
		if err != nil && err != operator.ErrNotFound {
			return nil, err
		}

		if existing != nil && existing.ID != operatorID {
			return nil, application.ErrEmailExists
		}

		op.Email = *req.Email
	}

	if req.Role != nil {
		// Role field is intentionally ignored here.
		// Role is org-scoped and updates should be done through organization membership.
		_ = req.Role // Acknowledge the field is set but not used.
	}

	op.UpdatedAt = time.Now()

	if err := s.operatorRepo.Update(ctx, op); err != nil {
		return nil, err
	}

	return op, nil
}

// DeleteOperator deletes an operator (admin only).
// Implements the operator deletion flow:.
// 1. Check if operator is last super_admin of any org - CANNOT delete if so.
// 2. Cancel pending invitations sent by user.
// 3. Remove from all org memberships.
// 4. Delete sessions.
// DeleteOperator deletes an operator (admin only).
// Implements the operator deletion flow:.
// 1. Check if operator is last super_admin of any org - CANNOT delete if so.
// 2. Cancel pending invitations sent by user.
// 3. Remove from all org memberships.
// 4. Delete sessions.
// 5. Delete operator record.
func (s *AuthService) DeleteOperator(ctx context.Context, operatorID string) error {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if errors.Is(err, operator.ErrNotFound) {
			return application.ErrOperatorNotFound
		}
		return err
	}

	// Check if operator is the last super_admin of any organization.
	memberships, err := s.memberRepo.ListByOperator(ctx, operatorID)
	if err != nil {
		return err
	}

	for _, m := range memberships {
		if m.Role.IsSuperAdmin() && m.IsActive() {
			// Check if this is the last super_admin.
			count, err := s.memberRepo.CountSuperAdminsByOrganization(ctx, m.OrganizationID)
			if err != nil {
				return err
			}
			if count <= 1 {
				return application.ErrCannotDeleteLastSuperAdmin
			}
		}
	}

	// Cancel all pending invitations sent by this operator.
	if err := s.invitationRepo.SoftDeleteByInvitedBy(ctx, operatorID); err != nil {
		return err
	}

	// Remove from all organization memberships.
	if err := s.memberRepo.SoftDeleteByOperator(ctx, operatorID); err != nil {
		return err
	}

	// Delete all sessions for this operator.
	if err := s.sessionRepo.DeleteByOperatorID(ctx, operatorID); err != nil {
		return err
	}

	// Delete the operator record.
	if err := s.operatorRepo.Delete(ctx, operatorID); err != nil {
		return err
	}

	_ = op // op loaded for validation above.
	return nil
}

// DeleteOwnAccount deletes the operator's own account.
// Requires password confirmation and follows the deletion flow.
func (s *AuthService) DeleteOwnAccount(ctx context.Context, operatorID, password string) error {
	op, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if errors.Is(err, operator.ErrNotFound) {
			return application.ErrOperatorNotFound
		}
		return err
	}

	// Verify password.
	if op.PasswordHash != "" {
		if err := s.passwordHasher.Verify(password, op.PasswordHash); err != nil {
			return application.ErrInvalidCredentials
		}
	}

	// Check if operator is the last super_admin of any organization.
	memberships, err := s.memberRepo.ListByOperator(ctx, operatorID)
	if err != nil {
		return err
	}

	for _, m := range memberships {
		if m.Role.IsSuperAdmin() && m.IsActive() {
			count, err := s.memberRepo.CountSuperAdminsByOrganization(ctx, m.OrganizationID)
			if err != nil {
				return err
			}
			if count <= 1 {
				return application.ErrCannotDeleteLastSuperAdmin
			}
		}
	}

	// Cancel all pending invitations sent by this operator.
	if err := s.invitationRepo.SoftDeleteByInvitedBy(ctx, operatorID); err != nil {
		return err
	}

	// Remove from all organization memberships.
	if err := s.memberRepo.SoftDeleteByOperator(ctx, operatorID); err != nil {
		return err
	}

	// Delete all sessions for this operator.
	if err := s.sessionRepo.DeleteByOperatorID(ctx, operatorID); err != nil {
		return err
	}

	// Delete the operator record.
	return s.operatorRepo.Delete(ctx, operatorID)
}

// GetDevicesByOperator returns all devices owned by an operator (for deletion warning).
func (s *AuthService) GetDevicesByOperator(ctx context.Context, operatorID string) (int, error) {
	return s.deviceRepo.CountByOperator(ctx, operatorID)
}
