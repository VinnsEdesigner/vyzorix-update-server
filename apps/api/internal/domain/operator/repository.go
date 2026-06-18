package operator

import (
	"context"
)

// Repository defines the interface for operator data access.
type Repository interface {
	// FindByID retrieves an operator by ID.
	FindByID(ctx context.Context, id string) (*Operator, error)
	
	// FindByEmail retrieves an operator by email.
	FindByEmail(ctx context.Context, email string) (*Operator, error)
	
	// FindByGoogleID retrieves an operator by Google ID.
	FindByGoogleID(ctx context.Context, googleID string) (*Operator, error)
	
	// FindByGitHubID retrieves an operator by GitHub ID.
	FindByGitHubID(ctx context.Context, githubID string) (*Operator, error)
	
	// Create creates a new operator.
	Create(ctx context.Context, op *Operator) error
	
	// Update updates an existing operator.
	Update(ctx context.Context, op *Operator) error
	
	// Delete deletes an operator.
	Delete(ctx context.Context, id string) error
	
	// Count returns the total number of operators.
	Count(ctx context.Context) (int, error)
	
	// List returns a paginated list of operators.
	List(ctx context.Context, limit, offset int) ([]*Operator, int, error)
	
	// UpdatePassword updates the password hash for an operator.
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	
	// UpdateMFA updates MFA settings for an operator.
	UpdateMFA(ctx context.Context, id, secret string, enabled bool) error
	
	// UpdateOperatorMFA updates the MFA secret and backup codes for an operator.
	UpdateOperatorMFA(ctx context.Context, operatorID, mfaSecret string, backupCodes []string) error
	
	// VerifyEmail marks an operator's email as verified.
	VerifyEmail(ctx context.Context, id string) error
	
	// UpdateEmailVerified updates the email verified status for an operator.
	UpdateEmailVerified(ctx context.Context, id string, verified bool) error
	
	// UpdateGoogleID updates the Google ID for an operator.
	UpdateGoogleID(ctx context.Context, id, googleID string) error
	
	// UpdateGitHubID updates the GitHub ID for an operator.
	UpdateGitHubID(ctx context.Context, id, githubID string) error
}
