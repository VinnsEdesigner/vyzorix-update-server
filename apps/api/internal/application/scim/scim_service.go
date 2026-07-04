package scim

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/scim"
)

const (
	coreSchema  = "urn:ietf:params:scim:schemas:core:2.0:User"
	listSchema  = "urn:ietf:params:scim:api:messages:2.0:ListResponse"
	errorSchema = "urn:ietf:params:scim:api:messages:2.0:Error"
)

// Repository defines the interface for operator persistence.
type Repository interface {
	FindByID(ctx context.Context, id string) (*operator.Operator, error)
	FindByEmail(ctx context.Context, email string) (*operator.Operator, error)
	List(ctx context.Context, limit, offset int) ([]*operator.Operator, error)
	Create(ctx context.Context, op *operator.Operator) error
	Update(ctx context.Context, op *operator.Operator) error
	Delete(ctx context.Context, id string) error
}

// SCIMService handles SCIM provisioning operations.
type SCIMService struct {
	repo     Repository
	basePath string
}

// NewSCIMService creates a new SCIM service.
func NewSCIMService(repo Repository, basePath string) *SCIMService {
	return &SCIMService{
		repo:     repo,
		basePath: strings.TrimSuffix(basePath, "/"),
	}
}

// CreateUser creates a new user via SCIM.
func (s *SCIMService) CreateUser(ctx context.Context, req *scim.SCIMUser) (*scim.SCIMUser, error) {
	// Check if user with this email already exists
	if len(req.Emails) > 0 && req.Emails[0].Value != "" {
		existing, err := s.repo.FindByEmail(ctx, req.Emails[0].Value)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("user already exists")
		}
	}

	op := &operator.Operator{
		ID:    generateID(),
		Email: getPrimaryEmail(req),
		Name:  req.Name.GivenName + " " + req.Name.FamilyName,
		Role:  operator.RoleOperator,
	}

	if err := s.repo.Create(ctx, op); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.operatorToSCIM(op, req), nil
}

// GetUser retrieves a user by ID.
func (s *SCIMService) GetUser(ctx context.Context, id string) (*scim.SCIMUser, error) {
	op, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.operatorToSCIM(op, nil), nil
}

// ListUsers lists all users with pagination.
func (s *SCIMService) ListUsers(ctx context.Context, startIndex, count int) (*scim.SCIMListResponse, error) {
	if startIndex < 1 {
		startIndex = 1
	}
	if count < 1 || count > 100 {
		count = 100
	}

	offset := startIndex - 1
	ops, err := s.repo.List(ctx, count, offset)
	if err != nil {
		return nil, err
	}

	resources := make([]scim.SCIMUser, 0, len(ops))
	for _, op := range ops {
		resources = append(resources, *s.operatorToSCIM(op, nil))
	}

	return &scim.SCIMListResponse{
		TotalResults: len(resources),
		StartIndex:   startIndex,
		ItemsPerPage: len(resources),
		Schemas:     []string{listSchema},
		Resources:   resources,
	}, nil
}

// UpdateUser updates an existing user.
func (s *SCIMService) UpdateUser(ctx context.Context, id string, req *scim.SCIMUser) (*scim.SCIMUser, error) {
	op, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if email := getPrimaryEmail(req); email != "" {
		op.Email = email
	}
	if req.Name.GivenName != "" || req.Name.FamilyName != "" {
		op.Name = strings.TrimSpace(req.Name.GivenName + " " + req.Name.FamilyName)
	}
	if req.DisplayName != "" {
		op.Name = req.DisplayName
	}
	// Note: Operator doesn't have Locked field - would need to be added if lockout is required

	if err := s.repo.Update(ctx, op); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.operatorToSCIM(op, req), nil
}

// DeleteUser deletes a user by ID.
func (s *SCIMService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *SCIMService) operatorToSCIM(op *operator.Operator, req *scim.SCIMUser) *scim.SCIMUser {
	now := op.UpdatedAt
	if now.IsZero() {
		now = op.CreatedAt
	}
	if now.IsZero() {
		now = op.CreatedAt
	}

	user := &scim.SCIMUser{
		ID:         op.ID,
		UserName:   op.Email,
		DisplayName: op.Name,
		Active:     true,
		Emails: []scim.SCIMEmail{
			{Value: op.Email, Primary: true},
		},
		Name: scim.SCIMName{
			GivenName:  firstName(op.Name),
			FamilyName: lastName(op.Name),
			Formatted:  op.Name,
		},
		Meta: scim.SCIMMeta{
			ResourceType: "User",
			Created:     op.CreatedAt,
			Modified:    now,
			Location:    fmt.Sprintf("%s/Users/%s", s.basePath, op.ID),
		},
	}

	if req != nil && req.ExternalID != "" {
		user.ExternalID = req.ExternalID
	}

	return user
}

func getPrimaryEmail(user *scim.SCIMUser) string {
	for _, e := range user.Emails {
		if e.Primary || e.Value != "" {
			return e.Value
		}
	}
	return user.UserName
}

func firstName(fullName string) string {
	parts := strings.SplitN(fullName, " ", 2)
	return parts[0]
}

func lastName(fullName string) string {
	parts := strings.SplitN(fullName, " ", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func generateID() string {
	return fmt.Sprintf("scim_%d", time.Now().UnixNano())
}
