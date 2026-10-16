package outcome

import "fmt"

type Result[T any] struct {
	value   T
	failure *Failure
	ok      bool
}

func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, ok: true}
}

func Fail[T any](failure *Failure) Result[T] {
	if failure == nil {
		failure = Fault(fmt.Errorf("outcome.Fail called with a nil failure"))
	}
	return Result[T]{failure: failure, ok: false}
}

func FromError[T any](err error) Result[T] {
	if err == nil {
		var zero T
		return Ok(zero)
	}
	return Fail[T](Fault(err))
}

func (r Result[T]) IsOk() bool { return r.ok }

func (r Result[T]) IsFailed() bool { return !r.ok }

func (r Result[T]) Value() (T, bool) {
	if !r.ok {
		var zero T
		return zero, false
	}
	return r.value, true
}

func (r Result[T]) Failure() *Failure {
	if r.ok {
		return nil
	}
	return r.failure
}

func Map[T, U any](r Result[T], f func(T) U) Result[U] {
	if !r.ok {
		return Fail[U](r.failure)
	}
	return Ok(f(r.value))
}

func FlatMap[T, U any](r Result[T], f func(T) Result[U]) Result[U] {
	if !r.ok {
		return Fail[U](r.failure)
	}
	return f(r.value)
}

func MapFailure[T any](r Result[T], f func(*Failure) *Failure) Result[T] {
	if r.ok {
		return r
	}
	return Fail[T](f(r.failure))
}

func Recover[T any](r Result[T], fallback func(*Failure) T) Result[T] {
	if r.ok {
		return r
	}
	return Ok(fallback(r.failure))
}

func Collect[T any](items []Result[T]) Result[[]T] {
	out := make([]T, 0, len(items))
	for _, item := range items {
		if !item.ok {
			return Fail[[]T](item.failure)
		}
		out = append(out, item.value)
	}
	return Ok(out)
}

func Partition[T any](items []Result[T]) (values []T, failures []*Failure) {
	for _, item := range items {
		if item.ok {
			values = append(values, item.value)
		} else {
			failures = append(failures, item.failure)
		}
	}
	return values, failures
}

func (r Result[T]) String() string {
	if r.ok {
		return "ok"
	}
	return "failed(" + r.failure.String() + ")"
}
