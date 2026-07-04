package scim

import (
	"context"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/scim"
)

type SCIMService struct {
	operatorRepo operator.Repository
}

func NewSCIMService(operatorRepo operator.Repository) *SCIMService {
	return &SCIMService{operatorRepo: operatorRepo}
}

func (s *SCIMService) ProvisionUser(ctx context.Context, user *scim.SCIMUser) (*operator.Operator, error) {
	email := ""
	if len(user.Emails) > 0 {
		email = strings.ToLower(user.Emails[0].Value)
	}

	var op *operator.Operator
	if user.ExternalID != "" {
		op, _ = s.operatorRepo.FindByID(ctx, user.ExternalID)
	}
	if op == nil && email != "" {
		op, _ = s.operatorRepo.FindByEmail(ctx, email)
	}

	now := time.Now()

	if op == nil {
		op = &operator.Operator{
			ID:            shared.GenerateID(),
			Email:         email,
			Name:          user.Name.GivenName + " " + user.Name.FamilyName,
			CreatedAt:     now,
			UpdatedAt:     now,
			EmailVerified: true,
			Role:          operator.RoleOperator,
		}
		if err := s.operatorRepo.Create(ctx, op); err != nil {
			return nil, err
		}
	} else {
		op.Name = user.Name.GivenName + " " + user.Name.FamilyName
		op.UpdatedAt = now
		op.EmailVerified = true
		if err := s.operatorRepo.Update(ctx, op); err != nil {
			return nil, err
		}
	}

	return op, nil
}

func (s *SCIMService) GetUser(ctx context.Context, id string) (*scim.SCIMUser, error) {
	op, err := s.operatorRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toSCIMUser(op), nil
}

func (s *SCIMService) ListUsers(ctx context.Context) ([]*scim.SCIMUser, error) {
	ops, _, err := s.operatorRepo.List(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}

	users := make([]*scim.SCIMUser, 0, len(ops))
	for _, op := range ops {
		users = append(users, s.toSCIMUser(op))
	}

	return users, nil
}

func (s *SCIMService) DeleteUser(ctx context.Context, id string) error {
	op, err := s.operatorRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	op.UpdatedAt = time.Now()
	return s.operatorRepo.Update(ctx, op)
}

func (s *SCIMService) toSCIMUser(op *operator.Operator) *scim.SCIMUser {
	parts := strings.SplitN(op.Name, " ", 2)
	given := ""
	family := ""
	if len(parts) >= 1 {
		given = parts[0]
	}
	if len(parts) >= 2 {
		family = parts[1]
	}

	return &scim.SCIMUser{
		ID:         op.ID,
		ExternalID: op.ID,
		UserName:   op.Email,
		Name: scim.SCIMName{
			Formatted:  op.Name,
			GivenName:  given,
			FamilyName: family,
		},
		DisplayName: op.Name,
		Emails: []scim.SCIMEmail{
			{Value: op.Email, Type: "work", Primary: true},
		},
		Active: true,
		Meta: scim.SCIMMeta{
			ResourceType: "User",
			Created:      op.CreatedAt,
			Modified:     op.UpdatedAt,
			Location:     "/scim/v2/Users/" + op.ID,
		},
	}
}
