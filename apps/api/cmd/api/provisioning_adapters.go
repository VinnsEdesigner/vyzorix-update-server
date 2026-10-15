package main

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	orgapplication "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	apikeydomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device_group"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

// Boot-provisioning adapters: bind the provisioner's small interfaces to the
// concrete application services and storage repositories so api_main can wire
// them. Each adapter is a one-line delegation; there is no logic here.

type provOrgRepo struct{ s *storage.OrganizationStorage }

func (a provOrgRepo) FindByName(ctx context.Context, operatorID, name string) (*organization.Organization, error) {
	return a.s.FindByName(ctx, operatorID, name)
}

type provOperatorRepo struct{ s *storage.OperatorRepository }

func (a provOperatorRepo) FindByEmail(ctx context.Context, email string) (*operator.Operator, error) {
	return a.s.FindByEmail(ctx, email)
}

type provGroupRepo struct {
	s *storage.DeviceGroupRepository
}

func (a provGroupRepo) Save(ctx context.Context, g *device_group.Group) error {
	return a.s.Save(ctx, g)
}

func (a provGroupRepo) AddMember(ctx context.Context, groupID, operatorID string) error {
	return a.s.AddMember(ctx, groupID, operatorID)
}

func (a provGroupRepo) ListByOrg(ctx context.Context, orgID string) ([]*device_group.Group, error) {
	return a.s.ListByOrg(ctx, orgID)
}

type provGrantRepo struct{ s *storage.GrantRepository }

func (a provGrantRepo) Save(ctx context.Context, p *permission.ResourcePermission) error {
	return a.s.Save(ctx, p)
}

type provOrgCreator struct {
	s *orgapplication.OrganizationService
}

func (a provOrgCreator) CreateOrganization(ctx context.Context, operatorID, name, description string, maxMembers int, role string) (*organization.Organization, error) {
	return a.s.CreateOrganization(ctx, operatorID, name, description, maxMembers, role)
}

type provOperatorCreator struct{ s *auth.AuthService }

func (a provOperatorCreator) CreateOperator(ctx context.Context, email, name, password, role string) (*operator.Operator, error) {
	return a.s.CreateOperator(ctx, &dto.RegisterRequest{Email: email, Name: name, Password: password, Role: role})
}

type provKeyCreator struct{ s *keys.APIKeyService }

func (a provKeyCreator) CreateKey(ctx context.Context, operatorID, orgID, name, scope string) (string, error) {
	res, err := a.s.GenerateKey(ctx, operatorID, orgID, &apikeydomain.CreateAPIKeyRequest{
		Name:  name,
		Scope: apikeydomain.Scope(scope),
	})
	if err != nil {
		return "", err
	}
	return res.FullKey, nil
}

func (a provKeyCreator) HasKey(ctx context.Context, operatorID, name string) (bool, error) {
	res, err := a.s.ListKeys(ctx, operatorID, 1, 100)
	if err != nil {
		return false, err
	}
	for _, k := range res.Keys {
		if k.Name == name {
			return true, nil
		}
	}
	return false, nil
}
