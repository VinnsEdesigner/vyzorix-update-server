
package command

import "testing"

func TestRiskTierOrdering(t *testing.T) {
	tests := []struct {
		a, b    RiskTier
		atLeast bool
	}{
		{RiskTierCritical, RiskTierZero, true},
		{RiskTierCritical, RiskTierCritical, true},
		{RiskTierHigh, RiskTierCritical, false},
		{RiskTierZero, RiskTierLow, false},
		{RiskTierMedium, RiskTierLow, true},
	}
	for _, tt := range tests {
		if got := tt.a.AtLeast(tt.b); got != tt.atLeast {
			t.Errorf("%s.AtLeast(%s) = %v, want %v", tt.a, tt.b, got, tt.atLeast)
		}
	}
}

func TestLookupRiskProfileKnownCommands(t *testing.T) {
	tests := []struct {
		name        string
		wantTier    RiskTier
		wantConfirm bool
	}{
		{TypeWakeUpUpdater, RiskTierLow, false},
		{TypeCheckUpdate, RiskTierLow, false},
		{"device.status", RiskTierZero, false},
		{"device.reboot", RiskTierHigh, true},
		{"device.factory_reset", RiskTierCritical, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := LookupRiskProfile(tt.name)
			if p.Tier != tt.wantTier {
				t.Errorf("tier = %s, want %s", p.Tier, tt.wantTier)
			}
			if p.RequiresConfirmation != tt.wantConfirm {
				t.Errorf("RequiresConfirmation = %v, want %v", p.RequiresConfirmation, tt.wantConfirm)
			}
		})
	}
}

func TestLookupRiskProfileUnknownFallsBackToDefault(t *testing.T) {
	p := LookupRiskProfile("some.unlisted.command")
	if p.Tier != DefaultRiskProfile.Tier {
		t.Errorf("unknown command tier = %s, want default %s", p.Tier, DefaultRiskProfile.Tier)
	}
	if p.RequiresConfirmation {
		t.Error("unknown commands should not require confirmation by default")
	}
}

func TestFactoryResetRequiresShorterConfirmationTTL(t *testing.T) {
	reboot := LookupRiskProfile("device.reboot")
	reset := LookupRiskProfile("device.factory_reset")
	if reset.ConfirmationTTL >= reboot.ConfirmationTTL {
		t.Errorf("factory_reset TTL (%s) should be shorter than reboot TTL (%s)",
			reset.ConfirmationTTL, reboot.ConfirmationTTL)
	}
}