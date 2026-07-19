package auth

import (
	"context"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// OAuthResult holds the result of OAuth operations.
type OAuthResult struct {
	Operator *operator.Operator
	IsNew    bool
}

// DefaultOAuthRole is the role assigned to new OAuth users.
// This can be made configurable via environment variables in the future.
const DefaultOAuthRole = operator.RoleOperator

// FindOrCreateGoogleOperator finds or creates an operator by Google ID.
func (s *AuthService) FindOrCreateGoogleOperator(ctx context.Context, googleID, email, name string) (*OAuthResult, error) {
	// Try to find by Google ID.
	op, err := s.operatorRepo.FindByGoogleID(ctx, googleID)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if op != nil {
		return &OAuthResult{Operator: op, IsNew: false}, nil
	}

	// Try to find by email and link Google account.
	op, err = s.operatorRepo.FindByEmail(ctx, email)
	if err != nil && err != operator.ErrNotFound {
		return nil, err
	}

	if op != nil {
		// Link existing account to Google.
		if err = s.operatorRepo.UpdateGoogleID(ctx, op.ID, googleID); err != nil {
			return nil, err
		}

		op.GoogleID = googleID

		return &OAuthResult{Operator: op, IsNew: false}, nil
	}

	// Create new operator.
	id := shared.GenerateID()

	now := time.Now()
	newOp := &operator.Operator{
		ID:            id,
		Email:         email,
		Name:          name,
		GoogleID:      googleID,
		EmailVerified: true, // Google verifies email
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.operatorRepo.Create(ctx, newOp); err != nil {
		return nil, err
	}

	return &OAuthResult{Operator: newOp, IsNew: true}, nil
}

// OAuthLinkError represents an error when linking OAuth accounts.
type OAuthLinkError struct {
	Message string
}

func (e *OAuthLinkError) Error() string {
	return e.Message
}
