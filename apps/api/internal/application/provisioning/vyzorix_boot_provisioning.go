package provisioning

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device_group"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
	"github.com/google/uuid"
)

type Config struct {
	Organizations []OrgConfig      `json:"organizations"`
	Operators     []OperatorConfig `json:"operators"`
	DeviceGroups  []GroupConfig    `json:"device_groups"`
	APIKeys       []APIKeyConfig   `json:"api_keys"`
	Grants        []GrantConfig    `json:"permission_grants"`
}

type OrgConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type OperatorConfig struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	OrgName  string `json:"org_name"`
	Role     string `json:"role"`
}

type GroupConfig struct {
	Name           string   `json:"name"`
	OrgName        string   `json:"org_name"`
	OperatorEmails []string `json:"operator_emails"`
}

type APIKeyConfig struct {
	Name    string `json:"name"`
	OrgName string `json:"org_name"`
	Scope   string `json:"scope"`
}

type GrantConfig struct {
	SubjectType string   `json:"subject_type"`
	SubjectID   string   `json:"subject_id"`
	OrgName     string   `json:"org_name"`
	Scope       string   `json:"scope"`
	Actions     []string `json:"actions"`
}

type OrgRepo interface {
	FindByName(ctx context.Context, operatorID, name string) (*organization.Organization, error)
}

type OperatorRepo interface {
	FindByEmail(ctx context.Context, email string) (*operator.Operator, error)
}

type GroupRepo interface {
	Save(ctx context.Context, g *device_group.Group) error
	AddMember(ctx context.Context, groupID, operatorID string) error
}

type GrantRepo interface {
	Save(ctx context.Context, p *permission.ResourcePermission) error
}

type OrgCreator interface {
	CreateOrganization(ctx context.Context, operatorID, name, description string, maxMembers int, role string) (*organization.Organization, error)
}

type OperatorCreator interface {
	CreateOperator(ctx context.Context, email, name, password, role string) (*operator.Operator, error)
}

type KeyCreator interface {
	CreateKey(ctx context.Context, operatorID, orgID, name, scope string) (string, error)
}

type Provisioner struct {
	log        *slog.Logger
	orgRepo    OrgRepo
	opRepo     OperatorRepo
	groupRepo  GroupRepo
	grantRepo  GrantRepo
	orgCreator OrgCreator
	opCreator  OperatorCreator
	keyCreator KeyCreator
}

func New(log *slog.Logger) *Provisioner {
	return &Provisioner{log: log}
}

func (p *Provisioner) WithRepositories(orgRepo OrgRepo, opRepo OperatorRepo, groupRepo GroupRepo, grantRepo GrantRepo) *Provisioner {
	p.orgRepo = orgRepo
	p.opRepo = opRepo
	p.groupRepo = groupRepo
	p.grantRepo = grantRepo
	return p
}

func (p *Provisioner) WithCreators(orgCreator OrgCreator, opCreator OperatorCreator, keyCreator KeyCreator) *Provisioner {
	p.orgCreator = orgCreator
	p.opCreator = opCreator
	p.keyCreator = keyCreator
	return p
}

func (p *Provisioner) LoadAndApply(ctx context.Context, path string) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			p.log.Info("no provisioning file found, skipping", "path", path)
			return nil
		}
		return fmt.Errorf("read provisioning file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse provisioning file: %w", err)
	}

	p.log.Info("provisioning file loaded",
		"orgs", len(cfg.Organizations),
		"operators", len(cfg.Operators),
		"groups", len(cfg.DeviceGroups),
		"keys", len(cfg.APIKeys),
		"grants", len(cfg.Grants),
	)

	orgMap := make(map[string]string)

	for _, oc := range cfg.Organizations {
		orgID, err := p.provisionOrg(ctx, oc)
		if err != nil {
			p.log.Error("failed to provision org", "name", oc.Name, "error", err)
			continue
		}
		orgMap[oc.Name] = orgID
		p.log.Info("provisioned org", "name", oc.Name, "id", orgID)
	}

	opMap := make(map[string]string)
	for _, opc := range cfg.Operators {
		opID, err := p.provisionOperator(ctx, opc)
		if err != nil {
			p.log.Error("failed to provision operator", "email", opc.Email, "error", err)
			continue
		}
		opMap[opc.Email] = opID
		p.log.Info("provisioned operator", "email", opc.Email, "id", opID)
	}

	for _, gc := range cfg.DeviceGroups {
		if err := p.provisionGroup(ctx, gc, orgMap, opMap); err != nil {
			p.log.Error("failed to provision group", "name", gc.Name, "error", err)
			continue
		}
		p.log.Info("provisioned group", "name", gc.Name)
	}

	for _, kc := range cfg.APIKeys {
		if err := p.provisionKey(ctx, kc, orgMap, opMap); err != nil {
			p.log.Error("failed to provision api key", "name", kc.Name, "error", err)
			continue
		}
		p.log.Info("provisioned api key", "name", kc.Name)
	}

	for _, gc := range cfg.Grants {
		if err := p.provisionGrant(ctx, gc, orgMap, opMap); err != nil {
			p.log.Error("failed to provision grant", "subject", gc.SubjectID, "error", err)
			continue
		}
		p.log.Info("provisioned grant", "subject", gc.SubjectID, "scope", gc.Scope)
	}

	return nil
}

func (p *Provisioner) provisionOrg(ctx context.Context, oc OrgConfig) (string, error) {
	if p.orgRepo != nil {
		if existing, err := p.orgRepo.FindByName(ctx, "", oc.Name); err == nil && existing != nil {
			p.log.Info("org already exists, skipping", "name", oc.Name, "id", existing.ID)
			return existing.ID, nil
		}
	}
	if p.orgCreator == nil {
		return "", fmt.Errorf("no org creator configured")
	}
	org, err := p.orgCreator.CreateOrganization(ctx, "", oc.Name, oc.Description, 0, "super_admin")
	if err != nil {
		return "", err
	}
	return org.ID, nil
}

func (p *Provisioner) provisionOperator(ctx context.Context, opc OperatorConfig) (string, error) {
	if p.opRepo != nil {
		if existing, err := p.opRepo.FindByEmail(ctx, opc.Email); err == nil && existing != nil {
			p.log.Info("operator already exists, skipping", "email", opc.Email, "id", existing.ID)
			return existing.ID, nil
		}
	}
	if p.opCreator == nil {
		return "", fmt.Errorf("no operator creator configured")
	}
	op, err := p.opCreator.CreateOperator(ctx, opc.Email, opc.Name, opc.Password, opc.Role)
	if err != nil {
		return "", err
	}
	return op.ID, nil
}

func (p *Provisioner) provisionGroup(ctx context.Context, gc GroupConfig, orgMap, opMap map[string]string) error {
	if p.groupRepo == nil {
		return fmt.Errorf("no group repo configured")
	}
	orgID := orgMap[gc.OrgName]
	if orgID == "" {
		return fmt.Errorf("org not found: %s", gc.OrgName)
	}
	groupID := uuid.New().String()
	g := &device_group.Group{
		ID:    groupID,
		OrgID: orgID,
		Name:  gc.Name,
	}
	if err := p.groupRepo.Save(ctx, g); err != nil {
		return err
	}
	for _, email := range gc.OperatorEmails {
		opID := opMap[email]
		if opID == "" {
			p.log.Warn("operator not found for group membership", "email", email, "group", gc.Name)
			continue
		}
		_ = p.groupRepo.AddMember(ctx, groupID, opID)
	}
	return nil
}

func (p *Provisioner) provisionKey(ctx context.Context, kc APIKeyConfig, orgMap, opMap map[string]string) error {
	if p.keyCreator == nil {
		return fmt.Errorf("no key creator configured")
	}
	orgID := orgMap[kc.OrgName]
	if orgID == "" {
		return fmt.Errorf("org not found: %s", kc.OrgName)
	}
	var opID string
	for _, id := range opMap {
		opID = id
		break
	}
	if opID == "" {
		return fmt.Errorf("no operator available for key creation")
	}
	_, err := p.keyCreator.CreateKey(ctx, opID, orgID, kc.Name, kc.Scope)
	return err
}

func (p *Provisioner) provisionGrant(ctx context.Context, gc GrantConfig, orgMap, opMap map[string]string) error {
	if p.grantRepo == nil {
		return fmt.Errorf("no grant repo configured")
	}
	orgID := orgMap[gc.OrgName]
	if orgID == "" {
		return fmt.Errorf("org not found: %s", gc.OrgName)
	}
	subjectID := gc.SubjectID
	if gc.SubjectType == "operator" {
		if id, ok := opMap[gc.SubjectID]; ok {
			subjectID = id
		}
	}
	actions := make([]permission.Action, len(gc.Actions))
	for i, a := range gc.Actions {
		actions[i] = permission.Action(a)
	}
	grant := &permission.ResourcePermission{
		ID:          uuid.New().String(),
		OrgID:       orgID,
		SubjectType: permission.SubjectType(gc.SubjectType),
		SubjectID:   subjectID,
		Actions:     actions,
		Scope:       gc.Scope,
	}
	return p.grantRepo.Save(ctx, grant)
}
