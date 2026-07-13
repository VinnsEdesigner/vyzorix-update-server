package auth

import (
	"context"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// FindOrCreateGitHubOperator finds or creates an operator by GitHub ID.
func (s *AuthService) FindOrCreateGitHubOperator(ctx context.Context, githubID, email, name string) (*OAuthResult, error) {
	// Try to find by GitHub ID.
	op, err := s.operatorRepo.FindByGitHubID(ctx, githubID)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if op != nil {
		return &OAuthResult{Operator: op, IsNew: false}, nil
	}

	// Try to find by email and link GitHub account.
	op, err = s.operatorRepo.FindByEmail(ctx, email)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if op != nil {
		// Link existing account to GitHub.
		if err = s.operatorRepo.UpdateGitHubID(ctx, op.ID, githubID); err != nil {
			return nil, err
		}

		op.GitHubID = githubID

		return &OAuthResult{Operator: op, IsNew: false}, nil
	}

	// Create new operator.
	count, err := s.operatorRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	// First user is super_admin (system bootstrap), others get default OAuth role
	role := DefaultOAuthRole
	if count == 0 {
		role = operator.RoleSuperAdmin
	}

	id := shared.GenerateID()

	now := time.Now()
	newOp := &operator.Operator{
		ID:            id,
		Email:         email,
		Name:          name,
		GitHubID:      githubID,
		Role:          role,
		EmailVerified: true, // GitHub verifies email
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.operatorRepo.Create(ctx, newOp); err != nil {
		return nil, err
	}

	return &OAuthResult{Operator: newOp, IsNew: true}, nil
}
