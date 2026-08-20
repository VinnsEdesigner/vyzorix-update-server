package provisioning

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device_group"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
)

type fakeOrgRepo struct {
	orgs map[string]*organization.Organization
}

func (f *fakeOrgRepo) FindByName(_ context.Context, _, name string) (*organization.Organization, error) {
	if o, ok := f.orgs[name]; ok {
		return o, nil
	}
	return nil, organization.ErrNotFound
}

type fakeOpRepo struct {
	ops map[string]*operator.Operator
}

func (f *fakeOpRepo) FindByEmail(_ context.Context, email string) (*operator.Operator, error) {
	if op, ok := f.ops[email]; ok {
		return op, nil
	}
	return nil, operator.ErrNotFound
}

type fakeGroupRepo struct {
	groups  map[string]*device_group.Group
	members map[string]map[string]bool
}

func (f *fakeGroupRepo) Save(_ context.Context, g *device_group.Group) error {
	if f.groups == nil {
		f.groups = make(map[string]*device_group.Group)
	}
	f.groups[g.ID] = g
	return nil
}

func (f *fakeGroupRepo) AddMember(_ context.Context, groupID, operatorID string) error {
	if f.members == nil {
		f.members = make(map[string]map[string]bool)
	}
	if f.members[groupID] == nil {
		f.members[groupID] = make(map[string]bool)
	}
	f.members[groupID][operatorID] = true
	return nil
}

type fakeGrantRepo struct {
	grants []*permission.ResourcePermission
}

func (f *fakeGrantRepo) Save(_ context.Context, p *permission.ResourcePermission) error {
	f.grants = append(f.grants, p)
	return nil
}

type fakeOrgCreator struct {
	counter int
	orgs    map[string]*organization.Organization
}

func (f *fakeOrgCreator) CreateOrganization(_ context.Context, _, name, desc string, _ int, _ string) (*organization.Organization, error) {
	f.counter++
	org := &organization.Organization{ID: "org-" + name, Name: name, Description: desc}
	if f.orgs == nil {
		f.orgs = make(map[string]*organization.Organization)
	}
	f.orgs[name] = org
	return org, nil
}

type fakeOpCreator struct {
	ops map[string]*operator.Operator
}

func (f *fakeOpCreator) CreateOperator(_ context.Context, email, name, _, _ string) (*operator.Operator, error) {
	op := &operator.Operator{ID: "op-" + email, Email: email, Name: name}
	if f.ops == nil {
		f.ops = make(map[string]*operator.Operator)
	}
	f.ops[email] = op
	return op, nil
}

type fakeKeyCreator struct {
	keys []string
}

func (f *fakeKeyCreator) CreateKey(_ context.Context, _, _, name, _ string) (string, error) {
	key := "vxyz-fake-" + name
	f.keys = append(f.keys, key)
	return key, nil
}

func writeProvisioningFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "provisioning.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return path
}

func TestLoadAndApply_EmptyPath_NoOp(t *testing.T) {
	p := New(slog.Default())
	err := p.LoadAndApply(context.Background(), "")
	if err != nil {
		t.Errorf("empty path should return nil, got: %v", err)
	}
}

func TestLoadAndApply_NonexistentFile_NoOp(t *testing.T) {
	p := New(slog.Default())
	err := p.LoadAndApply(context.Background(), "/nonexistent/path.json")
	if err != nil {
		t.Errorf("nonexistent file should return nil, got: %v", err)
	}
}

func TestLoadAndApply_CreatesOrg(t *testing.T) {
	json := `{
		"organizations": [{"name": "Test Org", "description": "test"}]
	}`
	path := writeProvisioningFile(t, json)
	orgRepo := &fakeOrgRepo{orgs: make(map[string]*organization.Organization)}
	orgCreator := &fakeOrgCreator{}
	p := New(slog.Default()).
		WithRepositories(orgRepo, nil, nil, nil).
		WithCreators(orgCreator, nil, nil)

	err := p.LoadAndApply(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadAndApply: %v", err)
	}
	if len(orgCreator.orgs) != 1 {
		t.Errorf("expected 1 org created, got %d", len(orgCreator.orgs))
	}
	if orgCreator.orgs["Test Org"] == nil {
		t.Error("Test Org should be created")
	}
}

func TestLoadAndApply_OrgAlreadyExists_SkipsCreation(t *testing.T) {
	json := `{
		"organizations": [{"name": "Existing Org"}]
	}`
	path := writeProvisioningFile(t, json)
	existing := &organization.Organization{ID: "org-existing", Name: "Existing Org"}
	orgRepo := &fakeOrgRepo{orgs: map[string]*organization.Organization{"Existing Org": existing}}
	orgCreator := &fakeOrgCreator{}
	p := New(slog.Default()).
		WithRepositories(orgRepo, nil, nil, nil).
		WithCreators(orgCreator, nil, nil)

	_ = p.LoadAndApply(context.Background(), path)
	if len(orgCreator.orgs) != 0 {
		t.Errorf("should not create org that already exists, created %d", len(orgCreator.orgs))
	}
}

func TestLoadAndApply_CreatesOperator(t *testing.T) {
	json := `{
		"organizations": [{"name": "Test Org"}],
		"operators": [{"email": "admin@test.local", "name": "Admin", "password": "pass", "org_name": "Test Org", "role": "super_admin"}]
	}`
	path := writeProvisioningFile(t, json)
	opRepo := &fakeOpRepo{ops: make(map[string]*operator.Operator)}
	opCreator := &fakeOpCreator{}
	orgCreator := &fakeOrgCreator{}
	p := New(slog.Default()).
		WithRepositories(nil, opRepo, nil, nil).
		WithCreators(orgCreator, opCreator, nil)

	_ = p.LoadAndApply(context.Background(), path)
	if len(opCreator.ops) != 1 {
		t.Errorf("expected 1 operator created, got %d", len(opCreator.ops))
	}
	if opCreator.ops["admin@test.local"] == nil {
		t.Error("admin@test.local should be created")
	}
}

func TestLoadAndApply_CreatesGroupWithMembers(t *testing.T) {
	json := `{
		"organizations": [{"name": "Test Org"}],
		"operators": [{"email": "op1@test.local", "name": "Op1", "password": "p", "org_name": "Test Org", "role": "admin"}],
		"device_groups": [{"name": "production", "org_name": "Test Org", "operator_emails": ["op1@test.local"]}]
	}`
	path := writeProvisioningFile(t, json)
	groupRepo := &fakeGroupRepo{}
	opCreator := &fakeOpCreator{}
	orgCreator := &fakeOrgCreator{}
	p := New(slog.Default()).
		WithRepositories(nil, nil, groupRepo, nil).
		WithCreators(orgCreator, opCreator, nil)

	_ = p.LoadAndApply(context.Background(), path)
	if len(groupRepo.groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groupRepo.groups))
	}
	for _, g := range groupRepo.groups {
		if g.Name != "production" {
			t.Errorf("expected group name 'production', got %s", g.Name)
		}
	}
	if len(groupRepo.members) == 0 {
		t.Error("expected group to have members")
	}
}

func TestLoadAndApply_CreatesGrant(t *testing.T) {
	json := `{
		"organizations": [{"name": "Test Org"}],
		"operators": [{"email": "op1@test.local", "name": "Op1", "password": "p", "org_name": "Test Org", "role": "admin"}],
		"permission_grants": [{"subject_type": "operator", "subject_id": "op1@test.local", "org_name": "Test Org", "scope": "devices:*", "actions": ["device.read"]}]
	}`
	path := writeProvisioningFile(t, json)
	grantRepo := &fakeGrantRepo{}
	opCreator := &fakeOpCreator{}
	orgCreator := &fakeOrgCreator{}
	opRepo := &fakeOpRepo{ops: make(map[string]*operator.Operator)}
	p := New(slog.Default()).
		WithRepositories(nil, opRepo, nil, grantRepo).
		WithCreators(orgCreator, opCreator, nil)

	_ = p.LoadAndApply(context.Background(), path)
	if len(grantRepo.grants) != 1 {
		t.Fatalf("expected 1 grant, got %d", len(grantRepo.grants))
	}
	g := grantRepo.grants[0]
	if g.Scope != "devices:*" {
		t.Errorf("expected scope 'devices:*', got %s", g.Scope)
	}
	if len(g.Actions) != 1 || g.Actions[0] != permission.ActionDeviceRead {
		t.Errorf("expected 1 action device.read, got %v", g.Actions)
	}
}
