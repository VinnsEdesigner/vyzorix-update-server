package outcome

type Disposition int

const (
	DispositionCompleted Disposition = iota
	DispositionFailedClean
	DispositionUncertain
	DispositionPartial
)

func (d Disposition) String() string {
	switch d {
	case DispositionCompleted:
		return "completed"
	case DispositionFailedClean:
		return "failed_clean"
	case DispositionPartial:
		return "partial"
	default:
		return "uncertain"
	}
}

func (d Disposition) SafeToRetry() bool { return d == DispositionFailedClean }

func (d Disposition) NeedsReconciliation() bool {
	return d == DispositionUncertain || d == DispositionPartial
}

func DispositionFor(f *Failure) Disposition {
	if f == nil {
		return DispositionCompleted
	}
	switch f.effect {
	case EffectNone:
		return DispositionFailedClean
	case EffectApplied:
		return DispositionCompleted
	case EffectPartial:
		return DispositionPartial
	default:
		return DispositionUncertain
	}
}

type Attempt[T any] struct {
	result      Result[T]
	disposition Disposition
}

func Completed[T any](value T) Attempt[T] {
	return Attempt[T]{result: Ok(value), disposition: DispositionCompleted}
}

func CleanFailure[T any](failure *Failure) Attempt[T] {
	normalised := failure.WithEffect(EffectNone)
	return Attempt[T]{result: Fail[T](normalised), disposition: DispositionFailedClean}
}

func UncertainFailure[T any](failure *Failure) Attempt[T] {
	normalised := failure.WithEffect(EffectUnknown)
	return Attempt[T]{result: Fail[T](normalised), disposition: DispositionUncertain}
}

func PartialFailure[T any](failure *Failure, value T) Attempt[T] {
	normalised := failure.WithEffect(EffectPartial)
	return Attempt[T]{result: Fail[T](normalised), disposition: DispositionPartial}
}

func (a Attempt[T]) Result() Result[T] { return a.result }

func (a Attempt[T]) Disposition() Disposition { return a.disposition }

func (a Attempt[T]) IsOk() bool { return a.result.IsOk() }

func (a Attempt[T]) Value() (T, bool) { return a.result.Value() }

func (a Attempt[T]) Failure() *Failure { return a.result.Failure() }

func (a Attempt[T]) String() string {
	return "attempt(" + a.disposition.String() + ", " + a.result.String() + ")"
}
