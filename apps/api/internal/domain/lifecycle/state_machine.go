package lifecycle

import (
	"errors"
	"fmt"
)

// ErrInvalidTransition is returned when an invalid lifecycle state transition is attempted.
var ErrInvalidTransition = errors.New("invalid lifecycle state transition")

// TransitionMap defines valid state transitions for a lifecycle.
// The map key is the current state, and the value is the set of allowed next states.
type TransitionMap map[string]map[string]bool

// StateMachine provides a generic state machine implementation for lifecycle management.
type StateMachine struct {
	transitions TransitionMap
	current     string
}

// NewStateMachine creates a new state machine with the given initial state and transition map.
func NewStateMachine(initial string, transitions TransitionMap) *StateMachine {
	return &StateMachine{
		current:     initial,
		transitions: transitions,
	}
}

// Current returns the current state.
func (sm *StateMachine) Current() string {
	return sm.current
}

// CanTransitionTo returns true if the state machine can transition to the target state.
func (sm *StateMachine) CanTransitionTo(target string) bool {
	allowed, exists := sm.transitions[sm.current]
	if !exists {
		return false
	}
	return allowed[target]
}

// TransitionTo transitions the state machine to a new state.
// Returns ErrInvalidTransition if the transition is not allowed.
func (sm *StateMachine) TransitionTo(target string) error {
	if !sm.CanTransitionTo(target) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, sm.current, target)
	}
	sm.current = target
	return nil
}

// Transitions returns the allowed transitions from the current state.
func (sm *StateMachine) Transitions() []string {
	if allowed, exists := sm.transitions[sm.current]; exists {
		result := make([]string, 0, len(allowed))
		for state := range allowed {
			result = append(result, state)
		}
		return result
	}
	return nil
}

// IsTerminal returns true if the current state is terminal (no outgoing transitions).
func (sm *StateMachine) IsTerminal() bool {
	transitions, exists := sm.transitions[sm.current]
	return !exists || len(transitions) == 0
}
