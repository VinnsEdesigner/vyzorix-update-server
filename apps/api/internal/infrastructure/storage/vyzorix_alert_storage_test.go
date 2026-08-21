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
		ID:        "rule-1",
		OrgID:     "org-1",
		Name:      "offline",
		Metric:    alert.MetricDeviceOfflineCount,
		Condition: alert.ConditionGt,
		Threshold: 3,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Enabled:   true,
		OnNoData:  alert.NoDataNoData,
		OnError:   alert.ErrorResolve,
	}
	if err := repo.Save(ctx, rule); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.OrgID != "org-1" || got.Metric != alert.MetricDeviceOfflineCount || !got.Enabled {
		t.Errorf("rule mismatch: %+v", got)
	}
	if got.OnNoData != alert.NoDataNoData || got.OnError != alert.ErrorResolve {
		t.Errorf("policies not persisted: %+v", got)
	}

	// Idempotent save replaces row.
	rule.Threshold = 5
	if err := repo.Save(ctx, rule); err != nil {
		t.Fatalf("Save update: %v", err)
	}
	got, _ = repo.GetByID(ctx, rule.ID)
	if got.Threshold != 5 {
		t.Errorf("expected threshold 5, got %v", got.Threshold)
	}

	// ListByOrg scopes by org and orders by name.
	other := &alert.Rule{
		ID:        "rule-2",
		OrgID:     "org-1",
		Name:      "failure-rate",
		Metric:    alert.MetricCommandFailureRate,
		Condition: alert.ConditionGte,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Save(ctx, other); err != nil {
		t.Fatalf("Save second: %v", err)
	}
	foreign := &alert.Rule{
		ID:        "rule-x",
		OrgID:     "org-2",
		Name:      "aaa",
		Metric:    alert.MetricDeviceOfflineCount,
		Condition: alert.ConditionGt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Save(ctx, foreign); err != nil {
		t.Fatalf("Save foreign: %v", err)
	}

	list, err := repo.ListByOrg(ctx, "org-1")
	if err != nil {
		t.Fatalf("ListByOrg: %v", err)
	}
	if len(list) != 2 || list[0].Name != "failure-rate" {
		t.Errorf("ListByOrg mismatch: %+v", list)
	}
	enabled, err := repo.ListEnabled(ctx)
	if err != nil {
		t.Fatalf("ListEnabled: %v", err)
	}
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled rules, got %d", len(enabled))
	}

	deleted, err := repo.Delete(ctx, rule.ID)
	if !deleted || err != nil {
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

	instances, err := states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByRuleID empty: %v", err)
	}
	if len(instances) != 0 {
		t.Fatalf("expected no instances before evaluation, got %+v", instances)
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

	instances, err = states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByRuleID: %v", err)
	}
	inst, ok := instances[""]
	if !ok || inst.State != alert.StateFiring || inst.LastValue != 7.5 {
		t.Errorf("instance mismatch: %+v", instances)
	}

	// Upsert replaces within the same labels hash.
	saved.State = alert.StateInactive
	if err := states.Upsert(ctx, saved); err != nil {
		t.Fatalf("Upsert update: %v", err)
	}
	instances, _ = states.GetByRuleID(ctx, rule.ID)
	if instances[""].State != alert.StateInactive {
		t.Errorf("expected inactive after upsert, got %s", instances[""].State)
	}

	// Label-keyed series coexist per rule.
	labeled := &alert.Instance{
		RuleID:        rule.ID,
		Labels:        map[string]string{"device_class": "tablet"},
		State:         alert.StatePending,
		Since:         time.Now(),
		LastEvaluated: time.Now(),
		LastValue:     2,
	}
	if err := states.Upsert(ctx, labeled); err != nil {
		t.Fatalf("Upsert labeled: %v", err)
	}
	instances, err = states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByRuleID labeled: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	labeledHash := alert.LabelsHash(map[string]string{"device_class": "tablet"})
	if instances[labeledHash].State != alert.StatePending {
		t.Errorf("labeled instance mismatch: %+v", instances[labeledHash])
	}

	byOrg, err := states.ListByOrg(ctx, "org-1")
	if err != nil {
		t.Fatalf("ListByOrg: %v", err)
	}
	if len(byOrg[rule.ID]) != 2 || byOrg[rule.ID][labeledHash].State != alert.StatePending {
		t.Errorf("ListByOrg mismatch: %+v", byOrg)
	}

	// Stale cleanup keeps only the given label hashes.
	if err := states.DeleteStaleForRule(ctx, rule.ID, []string{labeledHash}); err != nil {
		t.Fatalf("DeleteStaleForRule: %v", err)
	}
	instances, err = states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByRuleID after stale cleanup: %v", err)
	}
	if len(instances) != 1 || instances[labeledHash] == nil {
		t.Errorf("expected only labeled instance to remain: %+v", instances)
	}

	if err := states.DeleteByRuleID(ctx, rule.ID); err != nil {
		t.Fatalf("DeleteByRuleID: %v", err)
	}
	instances, err = states.GetByRuleID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetByRuleID after delete: %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("expected no instances after delete, got %+v", instances)
	}
}
