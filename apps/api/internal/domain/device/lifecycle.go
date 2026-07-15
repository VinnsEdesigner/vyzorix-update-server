package device

import (
	"errors"
	"fmt"
)

// Lifecycle represents the registration lifecycle state of a device.
type Lifecycle string

const (
	// LifecyclePending represents a device waiting for operator approval.
	LifecyclePending Lifecycle = "pending"
	// LifecycleRegistered represents an approved, active device.
	LifecycleRegistered Lifecycle = "registered"
	// LifecycleDeregistered represents a soft-deleted device (30-day retention).
	LifecycleDeregistered Lifecycle = "deregistered"
)

// ErrInvalidTransition is returned when a lifecycle state transition is not allowed.
var ErrInvalidTransition = errors.New("invalid lifecycle state transition")

// LifecycleTransitions defines valid state transitions for device lifecycle.
// The map key is the current state, and the value is the set of allowed next states.
var LifecycleTransitions = map[Lifecycle]map[Lifecycle]bool{
	LifecyclePending: {
		LifecycleRegistered:   true,
		LifecycleDeregistered: true,
	},
	LifecycleRegistered: {
		LifecycleDeregistered: true,
	},
	LifecycleDeregistered: {}, // No transitions allowed from deregistered
}

// CanTransitionTo returns true if the lifecycle can transition to the target state.
func (l Lifecycle) CanTransitionTo(target Lifecycle) bool {
	allowed, exists := LifecycleTransitions[l]
	if !exists {
		return false
	}
	return allowed[target]
}

// TransitionTo transitions the lifecycle to a new state.
// Returns ErrInvalidTransition if the transition is not allowed.
func (l *Lifecycle) TransitionTo(target Lifecycle) error {
	if !l.CanTransitionTo(target) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, *l, target)
	}
	*l = target
	return nil
}

// IsPending returns true if the lifecycle is pending.
func (l Lifecycle) IsPending() bool {
	return l == LifecyclePending
}

// IsRegistered returns true if the lifecycle is registered.
func (l Lifecycle) IsRegistered() bool {
	return l == LifecycleRegistered
}

// IsDeregistered returns true if the lifecycle is deregistered.
func (l Lifecycle) IsDeregistered() bool {
	return l == LifecycleDeregistered
}

// IsActive returns true if the device is in an active state (registered).
func (l Lifecycle) IsActive() bool {
	return l == LifecycleRegistered
}

// String returns the string representation of the lifecycle.
func (l Lifecycle) String() string {
	return string(l)
}

// Format implements fmt.Formatter for pretty printing.
func (l Lifecycle) Format(s fmt.State, verb rune) {
	format := "%s"
	if verb == 'v' && s.Flag('#') {
		format = "%q"
	}
	fmt.Fprintf(s, format, string(l))
}
