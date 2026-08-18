
package command

import "testing"

func TestRiskEvaluator_AllowLowRisk(t *testing.T) {
	ev := NewRiskEvaluator()
	actor := ActorContext{OperatorID: "op-1", OrgID: "org-1"}

	decision, profile := ev.Evaluate(TypeCheckUpdate, actor)
	if decision != DecisionAllow {
		t.Errorf("decision = %s, want %s", decision, DecisionAllow)
	}
	if profile.Tier != RiskTierLow {
		t.Errorf("profile tier = %s, want low", profile.Tier)
	}
}

func TestRiskEvaluator_RequireConfirmationForHighRisk(t *testing.T) {
	ev := NewRiskEvaluator()
	actor := ActorContext{OperatorID: "op-1", OrgID: "org-1", Confirmed: false}

	decision, profile := ev.Evaluate("device.reboot", actor)
	if decision != DecisionRequireConfirmation {
		t.Errorf("decision = %s, want %s", decision, DecisionRequireConfirmation)
	}
	if profile.Tier != RiskTierHigh {
		t.Errorf("profile tier = %s, want high", profile.Tier)
	}
}

func TestRiskEvaluator_HighRiskAllowedWhenConfirmed(t *testing.T) {
	ev := NewRiskEvaluator()
	actor := ActorContext{OperatorID: "op-1", OrgID: "org-1", Confirmed: true}

	decision, _ := ev.Evaluate("device.reboot", actor)
	if decision != DecisionAllow {
		t.Errorf("decision = %s, want %s (confirmed high-risk)", decision, DecisionAllow)
	}
}

func TestRiskEvaluator_CriticalRequiresMFAEvenWhenConfirmed(t *testing.T) {
	ev := NewRiskEvaluator()

	// Confirmed but no MFA → still gated until Phase 3 confirmation store.
	actor := ActorContext{OperatorID: "op-1", Confirmed: true, MFAVerified: false}
	decision, _ := ev.Evaluate("device.factory_reset", actor)
	if decision != DecisionRequireConfirmation {
		t.Errorf("decision = %s, want %s (critical needs MFA)", decision, DecisionRequireConfirmation)
	}

	// Confirmed AND MFA-verified → allowed (Phase 3 will tighten to a confirmation token).
	actor.MFAVerified = true
	decision, _ = ev.Evaluate("device.factory_reset", actor)
	if decision != DecisionAllow {
		t.Errorf("decision = %s, want %s (MFA + confirmed)", decision, DecisionAllow)
	}
}

func TestRiskEvaluator_UnknownCommandAllowedAtDefaultTier(t *testing.T) {
	ev := NewRiskEvaluator()
	actor := ActorContext{OperatorID: "op-1"}

	decision, profile := ev.Evaluate("some.new.command", actor)
	if decision != DecisionAllow {
		t.Errorf("decision = %s, want %s for default-tier command", decision, DecisionAllow)
	}
	if profile.Tier != RiskTierMedium {
		t.Errorf("profile tier = %s, want medium (default)", profile.Tier)
	}
}

func TestRiskEvaluator_ZeroValueIsUsable(t *testing.T) {
	// The zero-value RiskEvaluator must work without construction so callers
	// can embed it in structs.
	var ev RiskEvaluator
	decision, _ := ev.Evaluate(TypeWakeUpUpdater, ActorContext{})
	if decision != DecisionAllow {
		t.Errorf("zero-value evaluator decision = %s, want %s", decision, DecisionAllow)
	}
}