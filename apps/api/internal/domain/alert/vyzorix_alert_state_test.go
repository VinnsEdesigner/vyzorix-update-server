package alert

import (
	"testing"
	"time"
)

func testRule(forSeconds int) *Rule {
	return &Rule{
		ID:         "rule-1",
		OrgID:      "org-1",
		Name:       "offline devices",
		Metric:     MetricDeviceOfflineCount,
		Condition:  ConditionGt,
		Threshold:  5,
		ForSeconds: forSeconds,
		Enabled:    true,
	}
}

func TestCondition_Breached(t *testing.T) {
	cases := []struct {
		cond      Condition
		value     float64
		threshold float64
		want      bool
	}{
		{ConditionGt, 6, 5, true},
		{ConditionGt, 5, 5, false},
		{ConditionGte, 5, 5, true},
		{ConditionGte, 4.9, 5, false},
		{ConditionLt, 4, 5, true},
		{ConditionLt, 5, 5, false},
		{ConditionLte, 5, 5, true},
		{ConditionLte, 5.1, 5, false},
		{Condition("bogus"), 10, 5, false},
	}
	for _, tc := range cases {
		if got := tc.cond.Breached(tc.value, tc.threshold); got != tc.want {
			t.Errorf("%s.Breached(%v, %v) = %v, want %v", tc.cond, tc.value, tc.threshold, got, tc.want)
		}
	}
}

func TestEvaluate_ImmediateFiring(t *testing.T) {
	rule := testRule(0)
	inst := NewInstance(rule.ID, nil)
	now := time.Now()

	tr := inst.Evaluate(rule, 10, now)
	if tr == nil {
		t.Fatal("expected transition")
	}
	if tr.From != StateInactive || tr.To != StateFiring {
		t.Errorf("transition %s -> %s, want inactive -> firing", tr.From, tr.To)
	}
	if !tr.Firing() {
		t.Error("expected Firing() to be true")
	}
	if inst.Since != now {
		t.Error("expected Since to be set to evaluation time")
	}
}

func TestEvaluate_PendingThenFiring(t *testing.T) {
	rule := testRule(60)
	inst := NewInstance(rule.ID, nil)
	t0 := time.Now()

	tr := inst.Evaluate(rule, 10, t0)
	if tr == nil || tr.To != StatePending {
		t.Fatalf("expected inactive -> pending, got %+v", tr)
	}

	// Still breached but within the pending window: no transition.
	tr = inst.Evaluate(rule, 10, t0.Add(30*time.Second))
	if tr != nil {
		t.Fatalf("expected no transition within pending window, got %+v", tr)
	}
	if inst.State != StatePending {
		t.Errorf("state = %s, want pending", inst.State)
	}

	// Past the pending window: fires.
	tr = inst.Evaluate(rule, 10, t0.Add(61*time.Second))
	if tr == nil || tr.To != StateFiring {
		t.Fatalf("expected pending -> firing, got %+v", tr)
	}
}

func TestEvaluate_ResolveFromFiring(t *testing.T) {
	rule := testRule(0)
	inst := NewInstance(rule.ID, nil)
	t0 := time.Now()

	inst.Evaluate(rule, 10, t0)
	if inst.State != StateFiring {
		t.Fatalf("state = %s, want firing", inst.State)
	}

	tr := inst.Evaluate(rule, 2, t0.Add(time.Minute))
	if tr == nil || tr.To != StateInactive {
		t.Fatalf("expected firing -> inactive, got %+v", tr)
	}
	if !tr.Resolved() {
		t.Error("expected Resolved() to be true")
	}
}

func TestEvaluate_ResolveFromPending(t *testing.T) {
	rule := testRule(60)
	inst := NewInstance(rule.ID, nil)
	t0 := time.Now()

	inst.Evaluate(rule, 10, t0)
	tr := inst.Evaluate(rule, 2, t0.Add(10*time.Second))
	if tr == nil || tr.To != StateInactive {
		t.Fatalf("expected pending -> inactive, got %+v", tr)
	}
	if !tr.Resolved() {
		t.Error("expected Resolved() to be true for pending -> inactive")
	}
}

func TestEvaluate_SteadyStateNoTransition(t *testing.T) {
	rule := testRule(0)
	inst := NewInstance(rule.ID, nil)
	t0 := time.Now()

	// Not breached while inactive: no transition.
	if tr := inst.Evaluate(rule, 2, t0); tr != nil {
		t.Fatalf("expected no transition, got %+v", tr)
	}
	// Still breached while firing: no transition.
	inst.Evaluate(rule, 10, t0)
	if tr := inst.Evaluate(rule, 10, t0.Add(time.Minute)); tr != nil {
		t.Fatalf("expected no transition while continuously firing, got %+v", tr)
	}
	if inst.LastValue != 10 {
		t.Errorf("LastValue = %v, want 10", inst.LastValue)
	}
}

func TestRule_Validate(t *testing.T) {
	valid := testRule(0)
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid rule failed validation: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*Rule)
	}{
		{"missing org", func(r *Rule) { r.OrgID = "" }},
		{"missing name", func(r *Rule) { r.Name = " " }},
		{"bad metric", func(r *Rule) { r.Metric = "bogus" }},
		{"bad condition", func(r *Rule) { r.Condition = "bogus" }},
		{"negative for_seconds", func(r *Rule) { r.ForSeconds = -1 }},
	}
	for _, tc := range cases {
		rule := testRule(0)
		tc.mutate(rule)
		if err := rule.Validate(); err == nil {
			t.Errorf("%s: expected validation error", tc.name)
		}
	}
}

func TestEvaluate_NoDataPolicy(t *testing.T) {
	// Default (ignore): no state change, no transition.
	rule := testRule(0)
	inst := NewInstance(rule.ID, nil)
	t0 := time.Now()
	if tr := inst.EvaluateNoData(rule, t0); tr != nil {
		t.Fatalf("expected no transition on default policy, got %+v", tr)
	}
	if inst.State != StateInactive {
		t.Errorf("state = %s, want inactive", inst.State)
	}

	// no_data policy routes the instance and notifies.
	rule.OnNoData = NoDataNoData
	if tr := inst.EvaluateNoData(rule, t0); tr == nil || tr.To != StateNoData {
		t.Fatalf("expected transition to no_data, got %+v", tr)
	}
	if inst.State != StateNoData {
		t.Errorf("state = %s, want no_data", inst.State)
	}

	// The next real observation recovers from no_data.
	if tr := inst.Evaluate(rule, 2, t0.Add(time.Minute)); tr != nil {
		t.Fatalf("expected clean resume without transition, got %+v", tr)
	}
	if inst.State != StateInactive {
		t.Errorf("state = %s, want inactive after recovery", inst.State)
	}

	// resolve policy closes a firing instance.
	rule2 := testRule(0)
	rule2.OnNoData = NoDataResolve
	inst2 := NewInstance(rule2.ID, nil)
	inst2.Evaluate(rule2, 10, t0)
	if inst2.State != StateFiring {
		t.Fatalf("state = %s, want firing", inst2.State)
	}
	tr := inst2.EvaluateNoData(rule2, t0.Add(time.Minute))
	if tr == nil || !tr.Resolved() {
		t.Fatalf("expected resolved transition, got %+v", tr)
	}
}

func TestEvaluate_ErrorPolicy(t *testing.T) {
	rule := testRule(0)
	rule.OnError = ErrorError
	inst := NewInstance(rule.ID, nil)
	t0 := time.Now()

	tr := inst.EvaluateError(rule, t0)
	if tr == nil || tr.To != StateError {
		t.Fatalf("expected transition to error, got %+v", tr)
	}
	// Second consecutive error: no further transition.
	if tr := inst.EvaluateError(rule, t0.Add(time.Minute)); tr != nil {
		t.Fatalf("expected single error notification, got %+v", tr)
	}
	// Healthy tick recovers to the value-driven path.
	if tr := inst.Evaluate(rule, 2, t0.Add(2*time.Minute)); tr != nil {
		t.Fatalf("expected clean resume, got %+v", tr)
	}
	if inst.State != StateInactive {
		t.Errorf("state = %s, want inactive after recovery", inst.State)
	}

	// resolve policy closes a pending instance.
	ruleResolve := testRule(60)
	ruleResolve.OnError = ErrorResolve
	inst2 := NewInstance(ruleResolve.ID, nil)
	inst2.Evaluate(ruleResolve, 10, t0)
	tr = inst2.EvaluateError(ruleResolve, t0.Add(10*time.Second))
	if tr == nil || !tr.Resolved() {
		t.Fatalf("expected resolved transition from pending, got %+v", tr)
	}
}

func TestLabelsHash_Deterministic(t *testing.T) {
	a := LabelsHash(map[string]string{"device_class": "tablet", "group": "a"})
	b := LabelsHash(map[string]string{"group": "a", "device_class": "tablet"})
	if a != b {
		t.Errorf("hash order-sensitive: %q vs %q", a, b)
	}
	if LabelsHash(nil) != "" {
		t.Error("empty labels must hash to empty string")
	}
	if LabelsHash(map[string]string{"device_class": "tablet"}) == LabelsHash(map[string]string{"device_class": "phone"}) {
		t.Error("distinct label values collide")
	}
}
