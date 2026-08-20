package provisioning

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

func writeYAMLProvisioningFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "provisioning.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return path
}

func TestParseConfig_YAML(t *testing.T) {
	yamlData := []byte(`
organizations:
  - name: YAML Org
    description: from yaml
operators:
  - email: yaml@test.local
    name: YAML Admin
    password: pass123
    org_name: YAML Org
    role: admin
`)
	var cfg Config
	if err := parseConfig(yamlData, &cfg); err != nil {
		t.Fatalf("parseConfig YAML: %v", err)
	}
	if len(cfg.Organizations) != 1 {
		t.Errorf("expected 1 org, got %d", len(cfg.Organizations))
	}
	if cfg.Organizations[0].Name != "YAML Org" {
		t.Errorf("expected org name 'YAML Org', got %s", cfg.Organizations[0].Name)
	}
	if len(cfg.Operators) != 1 {
		t.Errorf("expected 1 operator, got %d", len(cfg.Operators))
	}
	if cfg.Operators[0].Email != "yaml@test.local" {
		t.Errorf("expected email 'yaml@test.local', got %s", cfg.Operators[0].Email)
	}
}

func TestParseConfig_JSON(t *testing.T) {
	jsonData := []byte(`{"organizations":[{"name":"JSON Org"}]}`)
	var cfg Config
	if err := parseConfig(jsonData, &cfg); err != nil {
		t.Fatalf("parseConfig JSON: %v", err)
	}
	if len(cfg.Organizations) != 1 {
		t.Errorf("expected 1 org, got %d", len(cfg.Organizations))
	}
	if cfg.Organizations[0].Name != "JSON Org" {
		t.Errorf("expected org name 'JSON Org', got %s", cfg.Organizations[0].Name)
	}
}

func TestLoadAndApply_YAMLFile_CreatesOrg(t *testing.T) {
	yamlContent := `
organizations:
  - name: YAML Test Org
    description: provisioned from yaml
`
	path := writeYAMLProvisioningFile(t, yamlContent)
	orgRepo := &fakeOrgRepo{orgs: make(map[string]*organization.Organization)}
	orgCreator := &fakeOrgCreator{}
	p := New(slog.Default()).
		WithRepositories(orgRepo, nil, nil, nil).
		WithCreators(orgCreator, nil, nil)

	err := p.LoadAndApply(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadAndApply YAML: %v", err)
	}
	if len(orgCreator.orgs) != 1 {
		t.Errorf("expected 1 org from YAML, got %d", len(orgCreator.orgs))
	}
	if orgCreator.orgs["YAML Test Org"] == nil {
		t.Error("YAML Test Org should be created")
	}
}

func TestLoadAndApply_YAMLFile_CreatesAllResources(t *testing.T) {
	yamlContent := `
organizations:
  - name: Full YAML Org
    description: complete provisioning test

operators:
  - email: yamladmin@test.local
    name: YAML Admin
    password: SecurePass!
    org_name: Full YAML Org
    role: super_admin

device_groups:
  - name: staging
    org_name: Full YAML Org
    operator_emails:
      - yamladmin@test.local

api_keys:
  - name: yaml-ci-key
    org_name: Full YAML Org
    scope: admin

permission_grants:
  - subject_type: operator
    subject_id: yamladmin@test.local
    org_name: Full YAML Org
    scope: "devices:*"
    actions:
      - device.manage
      - device.read
`
	path := writeYAMLProvisioningFile(t, yamlContent)
	orgRepo := &fakeOrgRepo{orgs: make(map[string]*organization.Organization)}
	opRepo := &fakeOpRepo{ops: make(map[string]*operator.Operator)}
	groupRepo := &fakeGroupRepo{}
	grantRepo := &fakeGrantRepo{}
	orgCreator := &fakeOrgCreator{}
	opCreator := &fakeOpCreator{}
	keyCreator := &fakeKeyCreator{}
	p := New(slog.Default()).
		WithRepositories(orgRepo, opRepo, groupRepo, grantRepo).
		WithCreators(orgCreator, opCreator, keyCreator)

	err := p.LoadAndApply(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadAndApply YAML full: %v", err)
	}

	if len(orgCreator.orgs) != 1 {
		t.Errorf("expected 1 org, got %d", len(orgCreator.orgs))
	}
	if len(opCreator.ops) != 1 {
		t.Errorf("expected 1 operator, got %d", len(opCreator.ops))
	}
	if len(groupRepo.groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groupRepo.groups))
	}
	if len(keyCreator.keys) != 1 {
		t.Errorf("expected 1 api key, got %d", len(keyCreator.keys))
	}
	if len(grantRepo.grants) != 1 {
		t.Errorf("expected 1 grant, got %d", len(grantRepo.grants))
	}
	if len(grantRepo.grants[0].Actions) != 2 {
		t.Errorf("expected 2 actions in grant, got %d", len(grantRepo.grants[0].Actions))
	}
}
