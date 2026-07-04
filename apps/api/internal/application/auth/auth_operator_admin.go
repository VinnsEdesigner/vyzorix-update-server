package auth

import (
	"context"
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
		response[i] = dto.OperatorListResponse{
			ID:            op.ID,
			Email:         op.Email,
			Name:          op.Name,
			Role:          string(op.Role),
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
		Role:         role,
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
		if err == operator.ErrNotFound {
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
		role := operator.OperatorRole(*req.Role)
		if !role.IsValid() {
			return nil, application.ErrInvalidInput
		}
		op.Role = role
	}

	op.UpdatedAt = time.Now()

	if err := s.operatorRepo.Update(ctx, op); err != nil {
		return nil, err
	}

	return op, nil
}

// DeleteOperator deletes an operator (admin only).
func (s *AuthService) DeleteOperator(ctx context.Context, operatorID string) error {
	_, err := s.operatorRepo.FindByID(ctx, operatorID)
	if err != nil {
		if err == operator.ErrNotFound {
			return application.ErrOperatorNotFound
		}
		return err
	}

	return s.operatorRepo.Delete(ctx, operatorID)
}
