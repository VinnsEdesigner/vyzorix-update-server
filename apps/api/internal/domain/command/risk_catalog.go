package command

import "time"

// RiskTier classifies the potential blast radius of an operation. Higher tiers.
// demand stronger authorization and an explicit confirmation step before the.
// server acts. Tiers are ordered: zero < low < medium < high < critical.
type RiskTier string

const (
	// RiskTierZero is a pure read with no side effects.
	RiskTierZero RiskTier = "zero"
	// RiskTierLow has minor, reversible side effects scoped to a single resource.
	RiskTierLow RiskTier = "low"
	// RiskTierMedium mutates user-owned data; reversible but impactful.
	RiskTierMedium RiskTier = "medium"
	// RiskTierHigh affects system configuration or is org-wide; requires.
	// confirmation and (in Phase 3) org-admin authority.
	RiskTierHigh RiskTier = "high"
	// RiskTierCritical is destructive and irreversible (e.g. factory reset).
	// Requires confirmation and (in Phase 3) MFA re-verification.
	RiskTierCritical RiskTier = "critical"
)

// AtLeast reports whether this tier is >= other in the defined ordering.
func (t RiskTier) AtLeast(other RiskTier) bool {
	return tierRank[t] >= tierRank[other]
}

var tierRank = map[RiskTier]int{
	RiskTierZero:     0,
	RiskTierLow:      1,
	RiskTierMedium:   2,
	RiskTierHigh:     3,
	RiskTierCritical: 4,
}

// CommandRiskProfile describes the risk characteristics of a single command.
// It is the static policy that the RiskEvaluator combines with runtime actor.
// context to reach an authorization decision.
type CommandRiskProfile struct {
	// Tier is the base risk tier for this command.
	Tier RiskTier
	// RequiresConfirmation is true when the command must not execute until the.
	// caller explicitly confirms (e.g. a "confirm": true flag on the request).
	RequiresConfirmation bool
	// ConfirmationTTL bounds how long a confirmation remains valid once given.
	// Zero means the confirmation must be present on the executing request.
	ConfirmationTTL time.Duration
}

// DefaultConfirmationTTL is the validity window for a carried confirmation.
const DefaultConfirmationTTL = 5 * time.Minute

// riskRegistry is the catalog of commands with known risk profiles. Commands.
// not present here fall back to DefaultRiskProfile in the evaluator.
var riskRegistry = map[string]CommandRiskProfile{
	// Existing device commands — benign, server→device triggers.
	TypeWakeUpUpdater: {Tier: RiskTierLow},
	TypeCheckUpdate:   {Tier: RiskTierLow},

	// Dotted-style device commands referenced by the hardening spec. Aliased to.
	// the legacy uppercase names so both shapes resolve to the same profile.
	"device.wake_up":      {Tier: RiskTierLow},
	"device.check_update": {Tier: RiskTierLow},
	"device.status":       {Tier: RiskTierZero},
	"device.reboot": {
		Tier:                 RiskTierHigh,
		RequiresConfirmation: true,
		ConfirmationTTL:      DefaultConfirmationTTL,
	},
	"device.factory_reset": {
		Tier:                 RiskTierCritical,
		RequiresConfirmation: true,
		ConfirmationTTL:      2 * time.Minute,
	},
}

// LookupRiskProfile returns the risk profile for the given command name. The.
// lookup tries the name as-is, then the dotted alias for legacy uppercase.
// commands, and finally falls back to DefaultRiskProfile for unknown commands.
// (fail-open at a conservative tier rather than blocking innovation, but still.
// flagging the action in the audit trail).
func LookupRiskProfile(commandName string) CommandRiskProfile {
	if p, ok := riskRegistry[commandName]; ok {
		return p
	}
	return DefaultRiskProfile
}

// DefaultRiskProfile is the conservative fallback for commands not in the.
// catalog: medium tier, no confirmation required. New commands therefore land.
// auditable-but-executable until explicitly classified.
var DefaultRiskProfile = CommandRiskProfile{
	Tier: RiskTierMedium,
}
