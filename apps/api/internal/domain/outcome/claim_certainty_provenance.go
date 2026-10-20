package outcome

import "fmt"

type Certainty int

const (
	Unknown Certainty = iota
	Inferred
	Reported
	Observed
)

func (c Certainty) String() string {
	switch c {
	case Inferred:
		return "inferred"
	case Reported:
		return "reported"
	case Observed:
		return "observed"
	default:
		return "unknown"
	}
}

func (c Certainty) AtLeast(floor Certainty) bool { return c >= floor }

type Provenance struct {
	Source         string
	Method         string
	ObservedAtUnix int64
}

type Claim[T any] struct {
	value      T
	provenance Provenance
	certainty  Certainty
}

func Observable[T any](value T, prov Provenance) Claim[T] {
	return Claim[T]{value: value, certainty: Observed, provenance: prov}
}

func ReportedClaim[T any](value T, prov Provenance) Claim[T] {
	return Claim[T]{value: value, certainty: Reported, provenance: prov}
}

func Inference[T any](value T, prov Provenance) Claim[T] {
	return Claim[T]{value: value, certainty: Inferred, provenance: prov}
}

func UnknownClaim[T any](value T, prov Provenance) Claim[T] {
	return Claim[T]{value: value, certainty: Unknown, provenance: prov}
}

func (c Claim[T]) Value() (T, Certainty) { return c.value, c.certainty }

func (c Claim[T]) Raw() T { return c.value }

func (c Claim[T]) Certainty() Certainty { return c.certainty }

func (c Claim[T]) Provenance() Provenance { return c.provenance }

func (c Claim[T]) Established(floor Certainty) bool { return c.certainty.AtLeast(floor) }

func (c Claim[T]) TryValue(floor Certainty) (T, bool) {
	if !c.certainty.AtLeast(floor) {
		var zero T
		return zero, false
	}
	return c.value, true
}

func MapClaim[T, U any](c Claim[T], f func(T) U) Claim[U] {
	return Claim[U]{
		value:      f(c.value),
		certainty:  c.certainty,
		provenance: c.provenance,
	}
}

func Weaken[T any](c Claim[T], to Certainty) Claim[T] {
	if to >= c.certainty {
		return c
	}
	clone := c
	clone.certainty = to
	return clone
}

func Combine[T any](claims ...Claim[T]) Claim[T] {
	if len(claims) == 0 {
		var zero T
		return Claim[T]{value: zero, certainty: Unknown}
	}
	out := claims[0]
	for _, c := range claims[1:] {
		if c.certainty < out.certainty {
			out.certainty = c.certainty
		}
		if c.provenance.ObservedAtUnix != 0 &&
			(out.provenance.ObservedAtUnix == 0 ||
				c.provenance.ObservedAtUnix < out.provenance.ObservedAtUnix) {
			out.provenance.ObservedAtUnix = c.provenance.ObservedAtUnix
		}
	}
	return out
}

func (c Claim[T]) String() string {
	return fmt.Sprintf("claim(%s from %s/%s)", c.certainty, c.provenance.Source, c.provenance.Method)
}
