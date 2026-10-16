package outcome

import (
	"errors"
	"fmt"
	"strings"
)

type Class int

const (
	ClassUnspecified Class = iota

	ClassInvalid

	ClassNotFound

	ClassConflict

	ClassDenied

	ClassUnavailable

	ClassTimeout

	ClassRateLimited

	ClassPrecondition

	ClassFault
)

func (c Class) Retryable() bool {
	switch c {
	case ClassUnavailable, ClassTimeout, ClassRateLimited:
		return true
	default:
		return false
	}
}

func (c Class) String() string {
	switch c {
	case ClassInvalid:
		return "invalid"
	case ClassNotFound:
		return "not_found"
	case ClassConflict:
		return "conflict"
	case ClassDenied:
		return "denied"
	case ClassUnavailable:
		return "unavailable"
	case ClassTimeout:
		return "timeout"
	case ClassRateLimited:
		return "rate_limited"
	case ClassPrecondition:
		return "precondition"
	case ClassFault:
		return "fault"
	default:
		return "unspecified"
	}
}

type Blame int

const (
	BlameCaller Blame = iota
	BlameEnvironment
	BlameInternal
)

func (b Blame) String() string {
	switch b {
	case BlameCaller:
		return "caller"
	case BlameEnvironment:
		return "environment"
	default:
		return "internal"
	}
}

type EffectStatus int

const (
	EffectUnknown EffectStatus = iota
	EffectNone
	EffectApplied
	EffectPartial
)

func (s EffectStatus) String() string {
	switch s {
	case EffectNone:
		return "none"
	case EffectApplied:
		return "applied"
	case EffectPartial:
		return "partial"
	default:
		return "unknown"
	}
}

func (s EffectStatus) NeedsReconciliation() bool {
	return s == EffectUnknown || s == EffectPartial
}

//nolint:errname,govet // outcome value rather than a thrown error; field order chosen for readability
type Failure struct {
	class    Class
	blame    Blame
	effect   EffectStatus
	code     string
	message  string
	cause    *Failure
	internal string
	owned    []string
	evidence Evidence
}

func New(class Class, blame Blame, effect EffectStatus, code, message string) *Failure {
	return &Failure{
		class:   class,
		blame:   blame,
		effect:  effect,
		code:    code,
		message: message,
	}
}

func Invalid(code, message string) *Failure {
	return New(ClassInvalid, BlameCaller, EffectNone, code, message)
}

func NotFound(code, message string) *Failure {
	return New(ClassNotFound, BlameCaller, EffectNone, code, message)
}

func Conflict(code, message string) *Failure {
	return New(ClassConflict, BlameEnvironment, EffectUnknown, code, message)
}

func Denied(code, message string) *Failure {
	return New(ClassDenied, BlameCaller, EffectNone, code, message)
}

func Unavailable(code, message string) *Failure {
	return New(ClassUnavailable, BlameEnvironment, EffectUnknown, code, message)
}

func Timeout(code, message string) *Failure {
	return New(ClassTimeout, BlameEnvironment, EffectUnknown, code, message)
}

func RateLimited(code, message string) *Failure {
	return New(ClassRateLimited, BlameEnvironment, EffectNone, code, message)
}

func Precondition(code, message string) *Failure {
	return New(ClassPrecondition, BlameCaller, EffectNone, code, message)
}

func Fault(err error) *Failure {
	f := New(ClassFault, BlameInternal, EffectUnknown, "internal_fault", "An internal invariant was violated.")
	if err != nil {
		f = f.WithInternal(err.Error())
	}
	return f
}

func (f *Failure) WithEffect(effect EffectStatus) *Failure {
	clone := *f
	clone.effect = effect
	return &clone
}

func (f *Failure) WithCause(cause *Failure) *Failure {
	clone := *f
	clone.cause = cause
	return &clone
}

func (f *Failure) WithInternal(detail string) *Failure {
	clone := *f
	clone.internal = detail
	return &clone
}

func (f *Failure) WithMessage(message string) *Failure {
	clone := *f
	clone.message = message
	return &clone
}

func (f *Failure) WithOwned(resources ...string) *Failure {
	clone := *f
	clone.owned = append(append([]string{}, f.owned...), resources...)
	return &clone
}

func (f *Failure) WithEvidence(e Evidence) *Failure {
	clone := *f
	clone.evidence = e
	return &clone
}

func (f *Failure) Owned() []string {
	if f == nil {
		return nil
	}
	return f.owned
}

func (f *Failure) Owns(resource string) bool {
	for _, r := range f.Owned() {
		if r == resource {
			return true
		}
	}
	return false
}

func (f *Failure) Evidence() Evidence {
	if f == nil {
		return Evidence{}
	}
	return f.evidence
}

func (f *Failure) Retryable() bool {
	if f == nil {
		return false
	}
	if len(f.owned) > 0 {
		return false
	}
	for cur := f; cur != nil; cur = cur.cause {
		if len(cur.owned) > 0 {
			return false
		}
	}
	return f.class.Retryable()
}

func (f *Failure) Class() Class { return f.class }

func (f *Failure) Blame() Blame { return f.blame }

func (f *Failure) Effect() EffectStatus { return f.effect }

func (f *Failure) Code() string { return f.code }

func (f *Failure) Message() string { return f.message }

func (f *Failure) Internal() string { return f.internal }

func (f *Failure) Cause() *Failure { return f.cause }

func (f *Failure) HasClass(class Class) bool {
	for cur := f; cur != nil; cur = cur.cause {
		if cur.class == class {
			return true
		}
	}
	return false
}

func (f *Failure) Error() string { return f.String() }

func (f *Failure) Unwrap() error {
	if f.cause == nil {
		return nil
	}
	return f.cause
}

func (f *Failure) IsTarget(target error) bool {
	if target == nil {
		return false
	}
	return errors.Is(f, target)
}

func (f *Failure) String() string {
	if f == nil {
		return "<nil failure>"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s/%s/%s: %s", f.class, f.blame, f.effect, f.code)
	for cur := f.cause; cur != nil; cur = cur.cause {
		fmt.Fprintf(&b, " <- %s(%s)", cur.code, cur.class)
	}
	return b.String()
}
