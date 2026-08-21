package alert

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

// recordingNotifier is a test double for the domain Notifier port.
type recordingNotifier struct {
	notifications []*alert.Notification
}

func (n *recordingNotifier) Notify(_ context.Context, notif *alert.Notification) error {
	n.notifications = append(n.notifications, notif)
	return nil
}

// recordingHub records broadcast calls.
type recordingHub struct {
	eventTypes []string
}

func (h *recordingHub) BroadcastEvent(eventType string, _ []byte) error {
	h.eventTypes = append(h.eventTypes, eventType)
	return nil
}

func setupAlertTestDB(t *testing.T) *storage.SQLite {
	t.Helper()
	cfg := storage.DefaultConfig(filepath.Join(t.TempDir(), "alert-app-test.db"))
	s, err := storage.Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func insertDevice(t *testing.T, s *storage.SQLite, id, orgID, deviceClass string, online int) {
	t.Helper()
	now := time.Now().UnixMilli()
	_, err := s.DB().Exec(
		`INSERT INTO devices (id, firebase_install_id, command_secret, device_class, online, registered_at, last_seen, created_at, updated_at, organization_id)
		 VALUES (?, ?, 'secret', ?, ?, ?, ?, ?, ?, ?)`,
		id, "fcm-"+id, deviceClass, online, now, now, now, now, orgID,
	)
	if err != nil {
		t.Fatalf("insert device %s: %v", id, err)
	}
}

func insertCommand(t *testing.T, s *storage.SQLite, id, deviceID, status string, createdAt time.Time) {
	t.Helper()
	_, err := s.DB().Exec(
		`INSERT INTO commands (id, dispatch_id, device_id, command, created_at, updated_at, status) VALUES (?, ?, ?, 'reboot', ?, ?, ?)`,
		id, "dispatch-"+id, deviceID, createdAt.UnixMilli(), createdAt.UnixMilli(), status,
	)
	if err != nil {
		t.Fatalf("insert command %s: %v", id, err)
	}
}

func aggregateInstance(t *testing.T, states *storage.AlertStateRepository, ctx context.Context, ruleID string, hash string) *alert.Instance {
	instances, err := states.GetByRuleID(ctx, ruleID)
	if err != nil {
		t.Fatalf("GetByRuleID: %v", err)
	}
	inst, ok := instances[hash]
	if !ok {
		t.Fatalf("expected instance for hash %q, got %+v", hash, instances)
	}
	return inst
}

func TestService_CRUDAndScoping(t *testing.T) {
	s := setupAlertTestDB(t)
	service := NewService(
		storage.NewAlertRuleRepository(s.DB()),
		storage.NewAlertStateRepository(s.DB()),
		storage.NewAlertHistoryRepository(s.DB()),
	)
	ctx := context.Background()

	in := &RuleInput{
		OrgID:      "org-1",
		Name:       "offline devices",
		Metric:     alert.MetricDeviceOfflineCount,
		Condition:  alert.ConditionGt,
		Threshold:  1,
		ForSeconds: 0,
		OnNoData:   alert.NoDataNoData,
		OnError:    alert.ErrorResolve,
		Enabled:    true,
	}
	rule, err := service.CreateRule(ctx, in)
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}
	if rule.ID == "" {
		t.Fatal("expected generated rule ID")
	}
	if rule.OnNoData != alert.NoDataNoData || rule.OnError != alert.ErrorResolve {
		t.Errorf("policies not carried: %+v", rule)
	}

	// Invalid input rejected.
	bad := &RuleInput{OrgID: "org-1", Name: "", Metric: alert.MetricDeviceOfflineCount, Condition: alert.ConditionGt}
	if _, err := service.CreateRule(ctx, bad); err == nil {
		t.Error("expected validation error for empty name")
	}

	// Cross-org reads are hidden.
	if _, err := service.GetRule(ctx, "org-2", rule.ID); err != ErrRuleNotFound {
		t.Errorf("cross-org GetRule: %v", err)
	}

	// Update + disable clears the instances.
	in.Enabled = false
	in.Name = "renamed"
	if _, err := service.UpdateRule(ctx, "org-1", rule.ID, in); err != nil {
		t.Fatalf("UpdateRule: %v", err)
	}

	views, err := service.ListRules(ctx, "org-1")
	if err != nil {
		t.Fatalf("ListRules: %v", err)
	}
	if len(views) != 1 || views[0].Rule.Name != "renamed" {
		t.Errorf("ListRules mismatch: %+v", views)
	}

	if err := service.DeleteRule(ctx, "org-1", rule.ID); err != nil {
		t.Fatalf("DeleteRule: %v", err)
	}
	if err := service.DeleteRule(ctx, "org-1", rule.ID); err != ErrRuleNotFound {
		t.Errorf("second delete: %v", err)
	}
}

func TestEvaluator_FiresOnOfflineCount(t *testing.T) {
	s := setupAlertTestDB(t)
	rules := storage.NewAlertRuleRepository(s.DB())
	states := storage.NewAlertStateRepository(s.DB())
	ctx := context.Background()

	insertDevice(t, s, "dev-1", "org-1", "tablet", 0)
	insertDevice(t, s, "dev-2", "org-1", "tablet", 0)
	insertDevice(t, s, "dev-3", "org-1", "phone", 1)

	rule := &alert.Rule{
		ID:         "rule-1",
		OrgID:      "org-1",
		Name:       "offline devices",
		Metric:     alert.MetricDeviceOfflineCount,
		Condition:  alert.ConditionGt,
		Threshold:  1,
		Enabled:    true,
		WebhookURL: "https://hooks.example.com/x",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := rules.Save(ctx, rule); err != nil {
		t.Fatalf("Save: %v", err)
	}

	notifier := &recordingNotifier{}
	hub := &recordingHub{}
	evaluator := NewEvaluator(rules, states, storage.NewAlertHistoryRepository(s.DB()), NewMetricSource(s.DB(), 0), notifier, hub, nil)

	// Two breaching series: tablet (2 offline) + fleet aggregate (2 offline).
	transitions, err := evaluator.EvaluateAll(ctx, time.Now())
	if err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	if transitions != 2 {
		t.Errorf("transitions = %d, want 2 (tablet + aggregate)", transitions)
	}
	if len(notifier.notifications) != 2 {
		t.Fatalf("expected two firing notifications, got %d", len(notifier.notifications))
	}

	inst := aggregateInstance(t, states, ctx, rule.ID, "")
	if inst.State != alert.StateFiring {
		t.Errorf("aggregate state = %s, want firing", inst.State)
	}

	// Devices recover: both series resolve.
	if _, err := s.DB().Exec(`UPDATE devices SET online = 1`); err != nil {
		t.Fatalf("update devices: %v", err)
	}
	if _, err := evaluator.EvaluateAll(ctx, time.Now()); err != nil {
		t.Fatalf("EvaluateAll resolve: %v", err)
	}
	inst = aggregateInstance(t, states, ctx, rule.ID, "")
	if inst.State != alert.StateInactive {
		t.Errorf("aggregate state = %s, want inactive", inst.State)
	}
	if len(notifier.notifications) != 4 {
		t.Errorf("expected firing+resolved notifications per series, got %d", len(notifier.notifications))
	}
}

func TestEvaluator_CommandFailureRatePercent(t *testing.T) {
	s := setupAlertTestDB(t)
	rules := storage.NewAlertRuleRepository(s.DB())
	states := storage.NewAlertStateRepository(s.DB())
	ctx := context.Background()

	insertDevice(t, s, "dev-1", "org-1", "tablet", 1)
	now := time.Now()
	insertCommand(t, s, "cmd-1", "dev-1", "completed", now.Add(-time.Minute))
	insertCommand(t, s, "cmd-2", "dev-1", "failed", now.Add(-time.Minute))
	insertCommand(t, s, "cmd-3", "dev-1", "failed", now.Add(-time.Minute))
	insertCommand(t, s, "cmd-4", "dev-1", "completed", now.Add(-time.Minute))

	rule := &alert.Rule{
		ID:        "rule-1",
		OrgID:     "org-1",
		Name:      "failure rate",
		Metric:    alert.MetricCommandFailureRate,
		Condition: alert.ConditionGte,
		Threshold: 50,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := rules.Save(ctx, rule); err != nil {
		t.Fatalf("Save: %v", err)
	}

	notifier := &recordingNotifier{}
	hub := &recordingHub{}
	evaluator := NewEvaluator(rules, states, storage.NewAlertHistoryRepository(s.DB()), NewMetricSource(s.DB(), 0), notifier, hub, nil)
	if _, err := evaluator.EvaluateAll(ctx, now); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	inst := aggregateInstance(t, states, ctx, rule.ID, "")
	if inst.State != alert.StateFiring || inst.LastValue != 50 {
		t.Errorf("state = %s value = %v, want firing at 50", inst.State, inst.LastValue)
	}

	if len(hub.eventTypes) != 1 || hub.eventTypes[0] != "alert_notification" {
		t.Errorf("expected one hub broadcast, got %+v", hub.eventTypes)
	}

	// Strictly greater than 50 does not breach.
	rule.Condition = alert.ConditionGt
	if err := rules.Save(ctx, rule); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := evaluator.EvaluateAll(ctx, now); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	inst = aggregateInstance(t, states, ctx, rule.ID, "")
	if inst.State != alert.StateInactive {
		t.Errorf("state = %s, want inactive after condition change", inst.State)
	}
}

func TestEvaluator_PendingHoldsUntilForDuration(t *testing.T) {
	s := setupAlertTestDB(t)
	rules := storage.NewAlertRuleRepository(s.DB())
	states := storage.NewAlertStateRepository(s.DB())
	ctx := context.Background()

	insertDevice(t, s, "dev-1", "org-1", "tablet", 0)

	rule := &alert.Rule{
		ID:         "rule-1",
		OrgID:      "org-1",
		Name:       "offline pending",
		Metric:     alert.MetricDeviceOfflineCount,
		Condition:  alert.ConditionGt,
		Threshold:  0,
		ForSeconds: 60,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := rules.Save(ctx, rule); err != nil {
		t.Fatalf("Save: %v", err)
	}

	evaluator := NewEvaluator(rules, states, nil, NewMetricSource(s.DB(), 0), nil, nil, nil)
	t0 := time.Now()

	if _, err := evaluator.EvaluateAll(ctx, t0); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	inst := aggregateInstance(t, states, ctx, rule.ID, "")
	if inst.State != alert.StatePending {
		t.Errorf("state = %s, want pending during For window", inst.State)
	}

	if _, err := evaluator.EvaluateAll(ctx, t0.Add(90*time.Second)); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	inst = aggregateInstance(t, states, ctx, rule.ID, "")
	if inst.State != alert.StateFiring {
		t.Errorf("state = %s, want firing after For elapsed", inst.State)
	}
}

func TestMetricSource_LabelFanOut(t *testing.T) {
	s := setupAlertTestDB(t)
	ctx := context.Background()

	insertDevice(t, s, "dev-1", "org-1", "tablet", 0)
	insertDevice(t, s, "dev-2", "org-1", "tablet", 1)
	insertDevice(t, s, "dev-3", "org-1", "phone", 1)
	insertDevice(t, s, "dev-4", "org-1", "phone", 1)

	source := NewMetricSource(s.DB(), 0)
	series, err := source.Series(ctx, "org-1", alert.MetricDeviceOfflinePercent, time.Now())
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	if len(series) != 3 {
		t.Fatalf("expected 2 class series + aggregate, got %d", len(series))
	}
	byClass := make(map[string]float64)
	var aggregate *LabelSeries
	for _, obs := range series {
		if obs.Labels == nil {
			aggregate = obs
			continue
		}
		byClass[obs.Labels["device_class"]] = obs.Value
	}
	if aggregate == nil || aggregate.Value != 25 {
		t.Errorf("aggregate offline percent = %v, want 25", series)
	}
	if byClass["tablet"] != 50 || byClass["phone"] != 0 {
		t.Errorf("class breakdown wrong: %+v", byClass)
	}

	// Empty fleet: percent reports no data.
	series, err = source.Series(ctx, "org-empty", alert.MetricDeviceOfflinePercent, time.Now())
	if err != nil {
		t.Fatalf("Series empty: %v", err)
	}
	if len(series) != 1 || !series[0].NoData {
		t.Errorf("empty fleet should report no-data, got %+v", series)
	}
}

func TestMetricSource_NoDataFailureRate(t *testing.T) {
	s := setupAlertTestDB(t)
	ctx := context.Background()

	insertDevice(t, s, "dev-1", "org-1", "tablet", 1)

	source := NewMetricSource(s.DB(), 0)
	series, err := source.Series(ctx, "org-1", alert.MetricCommandFailureRate, time.Now())
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	if len(series) != 1 || !series[0].NoData {
		t.Errorf("zero-command window should flag no-data: %+v", series)
	}
}

func TestEvaluator_NoDataPolicyBridges(t *testing.T) {
	s := setupAlertTestDB(t)
	rules := storage.NewAlertRuleRepository(s.DB())
	states := storage.NewAlertStateRepository(s.DB())
	ctx := context.Background()

	insertDevice(t, s, "dev-1", "org-1", "tablet", 1)

	rule := &alert.Rule{
		ID:        "rule-1",
		OrgID:     "org-1",
		Name:      "failure rate on no-data",
		Metric:    alert.MetricCommandFailureRate,
		Condition: alert.ConditionGte,
		Threshold: 50,
		Enabled:   true,
		OnNoData:   alert.NoDataNoData,
		WebhookURL: "https://hooks.example.com/x",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := rules.Save(ctx, rule); err != nil {
		t.Fatalf("Save: %v", err)
	}

	notifier := &recordingNotifier{}
	evaluator := NewEvaluator(rules, states, nil, NewMetricSource(s.DB(), 0), notifier, nil, nil)
	if _, err := evaluator.EvaluateAll(ctx, time.Now()); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	inst := aggregateInstance(t, states, ctx, rule.ID, "")
	if inst.State != alert.StateNoData {
		t.Errorf("state = %s, want no_data with on_no_data policy", inst.State)
	}
	if len(notifier.notifications) != 1 {
		t.Errorf("expected no-data entry notification, got %d", len(notifier.notifications))
	}

	// A rule resolving on no-data recovers a firing instance.
	ruleMapping := &alert.Rule{
		ID:        "rule-2",
		OrgID:     "org-1",
		Name:      "failure rate resolve on no-data",
		Metric:    alert.MetricCommandFailureRate,
		Condition: alert.ConditionGte,
		Threshold: 50,
		Enabled:   true,
		OnNoData:   alert.NoDataResolve,
		WebhookURL: "https://hooks.example.com/x",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := rules.Save(ctx, ruleMapping); err != nil {
		t.Fatalf("Save second: %v", err)
	}
	now := time.Now()
	insertCommand(t, s, "cmd-1", "dev-1", "failed", now.Add(-time.Minute))
	if _, err := evaluator.EvaluateAll(ctx, now); err != nil {
		t.Fatalf("EvaluateAll second: %v", err)
	}
	inst = aggregateInstance(t, states, ctx, ruleMapping.ID, "")
	if inst.State != alert.StateFiring {
		t.Fatalf("expected firing with commands present, got %s", inst.State)
	}
	// Window expires without new commands: no-data resolve heals.
	if _, err := s.DB().Exec(`DELETE FROM commands`); err != nil {
		t.Fatalf("prune commands: %v", err)
	}
	if _, err := evaluator.EvaluateAll(ctx, now); err != nil {
		t.Fatalf("EvaluateAll no-data: %v", err)
	}
	inst = aggregateInstance(t, states, ctx, ruleMapping.ID, "")
	if inst.State != alert.StateInactive {
		t.Errorf("state = %s, want resolved on no-data policy", inst.State)
	}
}

func TestHistory_AppendAndQuery(t *testing.T) {
	s := setupAlertTestDB(t)
	rules := storage.NewAlertRuleRepository(s.DB())
	states := storage.NewAlertStateRepository(s.DB())
	history := storage.NewAlertHistoryRepository(s.DB())
	ctx := context.Background()

	insertDevice(t, s, "dev-1", "org-1", "tablet", 0)

	rule := &alert.Rule{
		ID:        "rule-1",
		OrgID:     "org-1",
		Name:      "history rule",
		Metric:    alert.MetricDeviceOfflineCount,
		Condition: alert.ConditionGt,
		Threshold: 0,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := rules.Save(ctx, rule); err != nil {
		t.Fatalf("Save: %v", err)
	}

	evaluator := NewEvaluator(rules, states, history, NewMetricSource(s.DB(), 0), nil, nil, nil)
	if _, err := evaluator.EvaluateAll(ctx, time.Now()); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}

	events, err := history.ListByOrg(ctx, "org-1", "", 10)
	if err != nil {
		t.Fatalf("ListByOrg: %v", err)
	}
	// Class series + aggregate both fired.
	if len(events) != 2 {
		t.Fatalf("expected class+aggregate events, got %d", len(events))
	}
	if events[0].FromState != alert.StateInactive || events[0].ToState != alert.StateFiring {
		t.Errorf("event states wrong: %+v", events[0])
	}

	if _, err := s.DB().Exec(`UPDATE devices SET online = 1`); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := evaluator.EvaluateAll(ctx, time.Now()); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	events, _ = history.ListByOrg(ctx, "org-1", "", 10)
	if len(events) != 4 {
		t.Errorf("expected firing+resolved for both series, got %d", len(events))
	}
}

func TestEvaluator_RepeatNotifications(t *testing.T) {
	s := setupAlertTestDB(t)
	rules := storage.NewAlertRuleRepository(s.DB())
	states := storage.NewAlertStateRepository(s.DB())
	ctx := context.Background()

	insertDevice(t, s, "dev-1", "org-1", "tablet", 0)

	rule := &alert.Rule{
		ID:                    "rule-1",
		OrgID:                 "org-1",
		Name:                  "repeat notify",
		Metric:                alert.MetricDeviceOfflineCount,
		Condition:             alert.ConditionGt,
		Threshold:             0,
		NotifyIntervalSeconds: 60,
		Enabled:               true,
		WebhookURL:            "https://hooks.example.com/x",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	if err := rules.Save(ctx, rule); err != nil {
		t.Fatalf("Save: %v", err)
	}

	notifier := &recordingNotifier{}
	evaluator := NewEvaluator(rules, states, nil, NewMetricSource(s.DB(), 0), notifier, nil, nil)
	t0 := time.Now()

	if _, err := evaluator.EvaluateAll(ctx, t0); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	if len(notifier.notifications) != 2 {
		t.Fatalf("expected firing notifications (class+aggregate), got %d", len(notifier.notifications))
	}

	// Within re-notify window: no new notification.
	if _, err := evaluator.EvaluateAll(ctx, t0.Add(30*time.Second)); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	if len(notifier.notifications) != 2 {
		t.Errorf("window should suppress, got %d notifications", len(notifier.notifications))
	}

	// Past interval: fires again per series.
	if _, err := evaluator.EvaluateAll(ctx, t0.Add(61*time.Second)); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	if len(notifier.notifications) != 4 {
		t.Errorf("expected re-notification per series, got %d", len(notifier.notifications))
	}

	// Resolves once, firing stops re-notifying.
	if _, err := s.DB().Exec(`UPDATE devices SET online = 1`); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := evaluator.EvaluateAll(ctx, t0.Add(2*time.Minute)); err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	last := notifier.notifications[len(notifier.notifications)-1]
	if !last.Transition.Resolved() {
		t.Errorf("expected resolved notify as last event, got %d events", len(notifier.notifications))
	}
}
