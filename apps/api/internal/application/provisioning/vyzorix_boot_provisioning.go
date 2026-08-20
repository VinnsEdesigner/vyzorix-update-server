package provisioning

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device_group"
	notification "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/serviceaccount"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Organizations   []OrgConfig            `json:"organizations" yaml:"organizations"`
	Operators       []OperatorConfig       `json:"operators" yaml:"operators"`
	DeviceGroups    []GroupConfig          `json:"device_groups" yaml:"device_groups"`
	APIKeys         []APIKeyConfig         `json:"api_keys" yaml:"api_keys"`
	Grants          []GrantConfig          `json:"permission_grants" yaml:"permission_grants"`
	AlertRules      []AlertRuleConfig      `json:"alert_rules" yaml:"alert_rules"`
	ContactPoints   []ContactPointConfig   `json:"contact_points" yaml:"contact_points"`
	ServiceAccounts []ServiceAccountConfig `json:"service_accounts" yaml:"service_accounts"`
}

type ServiceAccountConfig struct {
	Name    string `json:"name" yaml:"name"`
	OrgName string `json:"org_name" yaml:"org_name"`
}

type ContactPointConfig struct {
	Config  map[string]string `json:"config" yaml:"config"`
	Name    string            `json:"name" yaml:"name"`
	OrgName string            `json:"org_name" yaml:"org_name"`
	Channel string            `json:"channel" yaml:"channel"`
	Secret  string            `json:"secret" yaml:"secret"`
	Enabled bool              `json:"enabled" yaml:"enabled"`
}

type AlertRuleConfig struct {
	Name                  string  `json:"name" yaml:"name"`
	OrgName               string  `json:"org_name" yaml:"org_name"`
	Metric                string  `json:"metric" yaml:"metric"`
	Condition             string  `json:"condition" yaml:"condition"`
	WebhookURL            string  `json:"webhook_url" yaml:"webhook_url"`
	Threshold             float64 `json:"threshold" yaml:"threshold"`
	ForSeconds            int     `json:"for_seconds" yaml:"for_seconds"`
	NotifyIntervalSeconds int     `json:"notify_interval_seconds" yaml:"notify_interval_seconds"`
	Enabled               bool    `json:"enabled" yaml:"enabled"`
}

type OrgConfig struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

type OperatorConfig struct {
	Email    string `json:"email" yaml:"email"`
	Name     string `json:"name" yaml:"name"`
	Password string `json:"password" yaml:"password"`
	OrgName  string `json:"org_name" yaml:"org_name"`
	Role     string `json:"role" yaml:"role"`
}

type GroupConfig struct {
	Name           string   `json:"name" yaml:"name"`
	OrgName        string   `json:"org_name" yaml:"org_name"`
	OperatorEmails []string `json:"operator_emails" yaml:"operator_emails"`
}

type APIKeyConfig struct {
	Name    string `json:"name" yaml:"name"`
	OrgName string `json:"org_name" yaml:"org_name"`
	Scope   string `json:"scope" yaml:"scope"`
}

type GrantConfig struct {
	SubjectType string   `json:"subject_type" yaml:"subject_type"`
	SubjectID   string   `json:"subject_id" yaml:"subject_id"`
	OrgName     string   `json:"org_name" yaml:"org_name"`
	Scope       string   `json:"scope" yaml:"scope"`
	Actions     []string `json:"actions" yaml:"actions"`
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

type AlertRuleRepo interface {
	Save(ctx context.Context, rule *alert.Rule) error
	ListByOrg(ctx context.Context, orgID string) ([]*alert.Rule, error)
}

type ContactPointRepo interface {
	Save(ctx context.Context, cp *notification.ContactPoint) error
	ListByOrg(ctx context.Context, orgID string) ([]*notification.ContactPoint, error)
}

type ServiceAccountRepo interface {
	Save(ctx context.Context, sa *serviceaccount.ServiceAccount) error
	ListByOrg(ctx context.Context, orgID string) ([]*serviceaccount.ServiceAccount, error)
}

type Provisioner struct {
	log         *slog.Logger
	orgRepo     OrgRepo
	opRepo      OperatorRepo
	groupRepo   GroupRepo
	grantRepo   GrantRepo
	alertRepo   AlertRuleRepo
	contactRepo ContactPointRepo
	serviceRepo ServiceAccountRepo
	orgCreator  OrgCreator
	opCreator   OperatorCreator
	keyCreator  KeyCreator
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

func (p *Provisioner) WithAlertRepository(alertRepo AlertRuleRepo) *Provisioner {
	p.alertRepo = alertRepo
	return p
}

func (p *Provisioner) WithContactPointRepository(contactRepo ContactPointRepo) *Provisioner {
	p.contactRepo = contactRepo
	return p
}

func (p *Provisioner) WithServiceAccountRepository(serviceRepo ServiceAccountRepo) *Provisioner {
	p.serviceRepo = serviceRepo
	return p
}

func (p *Provisioner) WithCreators(orgCreator OrgCreator, opCreator OperatorCreator, keyCreator KeyCreator) *Provisioner {
	p.orgCreator = orgCreator
	p.opCreator = opCreator
	p.keyCreator = keyCreator
	return p
}

func parseConfig(data []byte, cfg *Config) error {
	if strings.HasSuffix(strings.TrimSpace(string(data)), "}") || strings.HasPrefix(strings.TrimSpace(string(data)), "{") {
		return json.Unmarshal(data, cfg)
	}
	return yaml.Unmarshal(data, cfg)
}

//nolint:gocyclo // LoadAndApply pipelines all resource types.
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
	if err := parseConfig(data, &cfg); err != nil {
		return fmt.Errorf("parse provisioning file: %w", err)
	}

	p.log.Info("provisioning file loaded",
		"orgs", len(cfg.Organizations),
		"operators", len(cfg.Operators),
		"groups", len(cfg.DeviceGroups),
		"keys", len(cfg.APIKeys),
		"grants", len(cfg.Grants),
		"alerts", len(cfg.AlertRules),
		"contact_points", len(cfg.ContactPoints),
		"service_accounts", len(cfg.ServiceAccounts),
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

	for _, ac := range cfg.AlertRules {
		if err := p.provisionAlertRule(ctx, ac, orgMap); err != nil {
			p.log.Error("failed to provision alert rule", "name", ac.Name, "error", err)
			continue
		}
		p.log.Info("provisioned alert rule", "name", ac.Name)
	}

	for _, cc := range cfg.ContactPoints {
		if err := p.provisionContactPoint(ctx, cc, orgMap); err != nil {
			p.log.Error("failed to provision contact point", "name", cc.Name, "error", err)
			continue
		}
		p.log.Info("provisioned contact point", "name", cc.Name)
	}

	for _, sc := range cfg.ServiceAccounts {
		if err := p.provisionServiceAccount(ctx, sc, orgMap); err != nil {
			p.log.Error("failed to provision service account", "name", sc.Name, "error", err)
			continue
		}
		p.log.Info("provisioned service account", "name", sc.Name)
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

func (p *Provisioner) provisionServiceAccount(ctx context.Context, sc ServiceAccountConfig, orgMap map[string]string) error {
	if p.serviceRepo == nil {
		return fmt.Errorf("no service account repository configured")
	}
	orgID, ok := orgMap[sc.OrgName]
	if !ok {
		return fmt.Errorf("org %q not found", sc.OrgName)
	}

	existing, _ := p.serviceRepo.ListByOrg(ctx, orgID)
	for _, sa := range existing {
		if sa.Name == sc.Name {
			p.log.Info("service account already exists, skipping", "name", sc.Name, "org", sc.OrgName)
			return nil
		}
	}

	sa := &serviceaccount.ServiceAccount{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		Name:      sc.Name,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := sa.Validate(); err != nil {
		return fmt.Errorf("invalid service account %q: %w", sc.Name, err)
	}
	return p.serviceRepo.Save(ctx, sa)
}

func (p *Provisioner) provisionContactPoint(ctx context.Context, cc ContactPointConfig, orgMap map[string]string) error {
	if p.contactRepo == nil {
		return fmt.Errorf("no contact point repository configured")
	}
	orgID, ok := orgMap[cc.OrgName]
	if !ok {
		return fmt.Errorf("org %q not found", cc.OrgName)
	}

	existing, _ := p.contactRepo.ListByOrg(ctx, orgID)
	for _, cp := range existing {
		if cp.Name == cc.Name {
			p.log.Info("contact point already exists, skipping", "name", cc.Name, "org", cc.OrgName)
			return nil
		}
	}

	cp := &notification.ContactPoint{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		Name:      cc.Name,
		Channel:   notification.ChannelType(cc.Channel),
		Secret:    cc.Secret,
		Config:    cc.Config,
		Enabled:   cc.Enabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if cp.Config == nil {
		cp.Config = make(map[string]string)
	}
	if err := cp.Validate(); err != nil {
		return fmt.Errorf("invalid contact point %q: %w", cc.Name, err)
	}
	return p.contactRepo.Save(ctx, cp)
}

func (p *Provisioner) provisionAlertRule(ctx context.Context, ac AlertRuleConfig, orgMap map[string]string) error {
	if p.alertRepo == nil {
		return fmt.Errorf("no alert rule repository configured")
	}
	orgID, ok := orgMap[ac.OrgName]
	if !ok {
		return fmt.Errorf("org %q not found", ac.OrgName)
	}

	// Idempotent: save via upsert if a rule with the same name already exists in the org.
	existing, _ := p.alertRepo.ListByOrg(ctx, orgID)
	for _, rule := range existing {
		if rule.Name == ac.Name {
			p.log.Info("alert rule already exists, skipping", "name", ac.Name, "org", ac.OrgName)
			return nil
		}
	}

	rule := &alert.Rule{
		ID:                    uuid.New().String(),
		OrgID:                 orgID,
		Name:                  ac.Name,
		WebhookURL:            ac.WebhookURL,
		Metric:                alert.Metric(ac.Metric),
		Condition:             alert.Condition(ac.Condition),
		Threshold:             ac.Threshold,
		ForSeconds:            ac.ForSeconds,
		NotifyIntervalSeconds: ac.NotifyIntervalSeconds,
		Enabled:               ac.Enabled,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("invalid alert rule %q: %w", ac.Name, err)
	}
	return p.alertRepo.Save(ctx, rule)
}
