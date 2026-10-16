package outcome

import "strings"

//nolint:govet // fieldalignment: field order chosen for readability, not packing
type Remediation struct {
	Action            RemediationAction
	Guidance          string
	RetryAfterSeconds int
}

type RemediationAction int

const (
	ActionRetry RemediationAction = iota
	ActionWait
	ActionReconcile
	ActionFixRequest
	ActionEscalate
)

func (a RemediationAction) String() string {
	switch a {
	case ActionRetry:
		return "retry"
	case ActionWait:
		return "wait"
	case ActionReconcile:
		return "reconcile"
	case ActionFixRequest:
		return "fix_request"
	default:
		return "escalate"
	}
}

func Remediate(f *Failure) Remediation {
	if f == nil {
		return Remediation{Action: ActionRetry, Guidance: "No failure to remediate."}
	}

	if owned := f.Owned(); len(owned) > 0 {
		return Remediation{
			Action:   ActionEscalate,
			Guidance: "Release the resources this operation still holds (" + joinShort(owned) + ") before retrying.",
		}
	}

	switch f.Effect() {
	case EffectPartial:
		return Remediation{
			Action:   ActionReconcile,
			Guidance: "Part of this operation applied. Verify the target's state before continuing.",
		}
	case EffectUnknown:
		return Remediation{
			Action:   ActionReconcile,
			Guidance: "The outcome could not be confirmed. Inspect the target before retrying.",
		}
	case EffectNone, EffectApplied:
		// Nothing changed, or the change is confirmed and complete. Fall through
		// to the class-based action below.
	}

	switch f.Class() {
	case ClassRateLimited:
		return Remediation{Action: ActionWait, RetryAfterSeconds: 5, Guidance: "Wait for the limit to reset, then retry."}
	case ClassUnavailable, ClassTimeout:
		return Remediation{Action: ActionRetry, Guidance: "Retry the request."}
	case ClassInvalid, ClassPrecondition:
		return Remediation{Action: ActionFixRequest, Guidance: "Correct the request and try again."}
	case ClassNotFound:
		return Remediation{Action: ActionFixRequest, Guidance: "The referenced entity does not exist; check the identifier."}
	case ClassConflict:
		return Remediation{Action: ActionReconcile, Guidance: "Re-read the current state, then retry against the new version."}
	case ClassDenied:
		return Remediation{Action: ActionEscalate, Guidance: "The caller lacks permission; request access before retrying."}
	default:
		return Remediation{Action: ActionEscalate, Guidance: "An internal invariant broke; report this failure."}
	}
}

func joinShort(items []string) string {
	if len(items) <= 3 {
		return strings.Join(items, ", ")
	}
	return strings.Join(items[:3], ", ") + "…"
}
