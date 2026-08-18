
package command

// Decision is the outcome of evaluating a command against its risk profile and
// the runtime actor context. It tells the caller whether to proceed, demand a
// confirmation, or refuse outright.
type Decision string

const (
	// DecisionAllow means the command may execute immediately.
	DecisionAllow Decision = "allow"
	// DecisionRequireConfirmation means the command is authorized only after
	// the caller explicitly confirms the risky operation.
	DecisionRequireConfirmation Decision = "require_confirmation"
	// DecisionDeny means the command is refused regardless of confirmation.
	DecisionDeny Decision = "deny"
)

// ActorContext carries the runtime attributes of the caller that influence a
// risk decision. It is intentionally a value type so handlers can build it
// cheaply from gin context without allocations.
type ActorContext struct {
	// OperatorID is the authenticated operator making the request. Empty for
	// system-originated commands.
	OperatorID string
	// OrgID scopes the command to an organization.
	OrgID string
	// IsSuperAdmin is true for platform-level operators who bypass org checks.
	IsSuperAdmin bool
	// MFAVerified indicates the current session has cleared MFA. Phase 3 will
	// require this for critical-tier commands; Phase 2 records it only.
	MFAVerified bool
	// Confirmed is true when the caller sent an explicit confirmation flag on
	// this request. Used to satisfy RequiresConfirmation profiles.
	Confirmed bool
}

// RiskEvaluator combines a command's static risk profile with runtime actor
// context to produce a Decision. It is a stateless value; the zero value is a
// valid evaluator that uses the default risk registry.
type RiskEvaluator struct{}

// NewRiskEvaluator returns a ready-to-use evaluator.
func NewRiskEvaluator() *RiskEvaluator { return &RiskEvaluator{} }

// Evaluate returns the authorization decision for a command given the actor.
// The accompanying CommandRiskProfile is returned so callers can include the
// tier in audit records without a second lookup.
func (RiskEvaluator) Evaluate(cmdName string, actor ActorContext) (Decision, CommandRiskProfile) {
	profile := LookupRiskProfile(cmdName)

	// Critical-tier commands require MFA. Without it the command is refused
	// pending the Phase 3 confirmation store; an explicit confirm flag alone
	// is insufficient for irreversible operations.
	if profile.Tier == RiskTierCritical && !actor.MFAVerified {
		return DecisionRequireConfirmation, profile
	}

	// High-tier and explicitly confirmation-gated commands need a confirmation
	// flag on the executing request.
	if profile.RequiresConfirmation && !actor.Confirmed {
		return DecisionRequireConfirmation, profile
	}

	return DecisionAllow, profile
}