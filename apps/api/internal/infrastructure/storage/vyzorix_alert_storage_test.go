package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
)

func setupAlertTestDB(t *testing.T) *SQLite {
	t.Helper()
	cfg := DefaultConfig(filepath.Join(t.TempDir(), "alert-test.db"))
	s, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestAlertRuleRepository_CRUD(t *testing.T) {
	s := setupAlertTestDB(t)
	repo := NewAlertRuleRepository(s.DB())
	ctx := context.Background()

	rule := &alert.Rule{
		ID:         "rule-1",
		OrgID:      "org-1",
		Name:       "offline devices",
		Metric:     alert.MetricDeviceOfflineCount,
		Condition:  alert.ConditionGt,
		Threshold:  5,
		ForSeconds: 30,
		Enabled:    true,
		WebhookURL: "https://hooks.example.com/xyz",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := repo.Save(ctx, rule); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != rule.Name || got.Metric != rule.Metric || got.Condition != rule.Condition {
		t.Errorf("GetByID mismatch: %+v", got)
	}
	if got.Threshold != 5 || got.ForSeconds != 30 || !got.Enabled || got.WebhookURL != rule.WebhookURL {
		t.Errorf("GetByID fields mismatch: %+v", got)
	}

	// Upsert overwrites mutable fields.
	rule.Name = "renamed"
	rule.Enabled = false
	if err := repo.Save(ctx, rule); err != nil {
		t.Fatalf("Save update: %v", err)
	}
	got, err = repo.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "renamed" || got.Enabled {
		t.Errorf("update not applied: %+v", got)
	}

	// ListEnabled excludes disabled rules.
	enabled, err := repo.ListEnabled(ctx)
	if err != nil {
		t.Fatalf("ListEnabled: %v", err)
	}
	if len(enabled) != 0 {
		t.Errorf("ListEnabled returned %d rules, want 0", len(enabled))
	}

	// ListByOrg returns org rules ordered by name.
	second := &alert.Rule{ID: "rule-2", OrgID: "org-1", Name: "aaa first", Metric: alert.MetricCommandFailureRate, Condition: alert.ConditionGt, Threshold: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := repo.Save(ctx, second); err != nil {
		t.Fatalf("Save second: %v", err)
	}
	rules, err := repo.ListByOrg(ctx, "org-1")
	if err != nil {
		t.Fatalf("ListByOrg: %v", err)
	}
	if len(rules) != 2 || rules[0].Name != "aaa first" {
		t.Errorf("ListByOrg order/count wrong: %+v", rules)
	}

	deleted, err := repo.Delete(ctx, rule.ID)
	if err != nil || !deleted {
		t.Fatalf("Delete: %v %v", deleted, err)
	}
	if _, err := repo.GetByID(ctx, rule.ID); err != alert.ErrNotFound {
		t.Errorf("GetByID after delete: %v", err)
	}
}

func TestAlertStateRepository_RoundTrip(t *testing.T) {
	s := setupAlertTestDB(t)
	rules := NewAlertRuleRepository(s.DB())
	states := NewAlertStateRepository(s.DB())
	ctx := context.Background()

	rule := &alert.Rule{
		ID:        "rule-1",
		OrgID:     "org-1",
		Name:      "offline",
		Metric:    alert.MetricDeviceOfflineCount,
		Condition: alert.ConditionGt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := rules.Save(ctx, rule); err != nil {
		t.Fatalf("Save rule: %v", err)
	}

	inst, err := states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByRuleID empty: %v", err)
	}
	if inst.State != alert.StateInactive {
		t.Fatalf("expected fresh inactive instance before evaluation, got %+v", inst)
	}

	saved := &alert.Instance{
		RuleID:        rule.ID,
		State:         alert.StateFiring,
		Since:         time.Now(),
		LastEvaluated: time.Now(),
		LastValue:     7.5,
	}
	if err := states.Upsert(ctx, saved); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	inst, err = states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByRuleID: %v", err)
	}
	if inst.State != alert.StateFiring || inst.LastValue != 7.5 {
		t.Errorf("instance mismatch: %+v", inst)
	}

	// Upsert replaces.
	saved.State = alert.StateInactive
	if err := states.Upsert(ctx, saved); err != nil {
		t.Fatalf("Upsert update: %v", err)
	}
	inst, _ = states.GetByRuleID(ctx, rule.ID)
	if inst.State != alert.StateInactive {
		t.Errorf("expected inactive after upsert, got %s", inst.State)
	}

	byOrg, err := states.ListByOrg(ctx, "org-1")
	if err != nil {
		t.Fatalf("ListByOrg: %v", err)
	}
	if len(byOrg) != 1 || byOrg[rule.ID].State != alert.StateInactive {
		t.Errorf("ListByOrg mismatch: %+v", byOrg)
	}

	if err := states.DeleteByRuleID(ctx, rule.ID); err != nil {
		t.Fatalf("DeleteByRuleID: %v", err)
	}
	inst, err = states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByRuleID after delete: %v", err)
	}
	if inst.State != alert.StateInactive {
		t.Errorf("expected fresh inactive instance after delete, got %+v", inst)
	}
}
