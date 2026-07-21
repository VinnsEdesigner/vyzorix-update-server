package inbox

import (
	"errors"
	"fmt"
)

// InboxStatus represents the status of an inbox entry.
// Implements the 5-state model from SPEC:
// PENDING -> ACKNOWLEDGED -> APPROVING -> APPROVED
//               REJECTED                        (device confirms) -> REGISTERED (external)
// StatusExpired for auto-cleanup after 30 days
type InboxStatus string

const (
	StatusPending      InboxStatus = "pending"      // Initial state after device registration
	StatusAcknowledged InboxStatus = "acknowledged" // Device has acknowledged the request
	StatusApproving    InboxStatus = "approving"   // Operator is approving, commandSecret being generated
	StatusApproved     InboxStatus = "approved"    // Fully approved, device can confirm
	StatusRejected     InboxStatus = "rejected"    // Rejected by operator
	StatusExpired      InboxStatus = "expired"      // Auto-cleanup after 30 days
)

// ErrInvalidInboxTransition is returned when an invalid inbox status transition is attempted.
var ErrInvalidInboxTransition = errors.New("invalid inbox status transition")

// InboxStatusTransitions defines valid state transitions for inbox lifecycle.
// The map key is the current state, and the value is the set of allowed next states.
var InboxStatusTransitions = map[InboxStatus]map[InboxStatus]bool{
	StatusPending: {
		StatusAcknowledged: true,
		StatusRejected:    true,
		StatusExpired:     true,
	},
	StatusAcknowledged: {
		StatusApproving: true,
		StatusRejected:  true,
		StatusExpired:   true,
	},
	StatusApproving: {
		StatusApproved: true,
		StatusRejected: true,
	},
	StatusApproved:  {}, // Terminal state for this flow
	StatusRejected:  {}, // Terminal state
	StatusExpired:   {}, // Terminal state
}

// CanTransitionTo returns true if the status can transition to the target status.
func (s InboxStatus) CanTransitionTo(target InboxStatus) bool {
	allowed, exists := InboxStatusTransitions[s]
	if !exists {
		return false
	}
	return allowed[target]
}

// TransitionTo transitions the status to a new state.
// Returns ErrInvalidInboxTransition if the transition is not allowed.
func (s *InboxStatus) TransitionTo(target InboxStatus) error {
	if !s.CanTransitionTo(target) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidInboxTransition, *s, target)
	}
	*s = target
	return nil
}

// IsValid checks if the status is a valid inbox status.
func (s InboxStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusAcknowledged, StatusApproving, StatusApproved, StatusRejected, StatusExpired:
		return true
	default:
		return false
	}
}

// IsPending checks if the entry is in pending state.
func (s InboxStatus) IsPending() bool {
	return s == StatusPending
}

// IsAcknowledged checks if the entry has been acknowledged by the device.
func (s InboxStatus) IsAcknowledged() bool {
	return s == StatusAcknowledged
}

// IsApproving checks if the entry is in the approving phase.
func (s InboxStatus) IsApproving() bool {
	return s == StatusApproving
}

// IsApproved checks if the entry has been fully approved.
func (s InboxStatus) IsApproved() bool {
	return s == StatusApproved
}

// IsRejected checks if the entry has been rejected.
func (s InboxStatus) IsRejected() bool {
	return s == StatusRejected
}

// IsExpired checks if the entry has expired.
func (s InboxStatus) IsExpired() bool {
	return s == StatusExpired
}

// IsTerminal checks if the status is a terminal state.
func (s InboxStatus) IsTerminal() bool {
	return s == StatusApproved || s == StatusRejected || s == StatusExpired
}

// CanBeAcknowledged checks if the entry can be acknowledged by device.
func (s InboxStatus) CanBeAcknowledged() bool {
	return s == StatusPending
}

// CanBeApproved checks if the entry can be approved by operator.
func (s InboxStatus) CanBeApproved() bool {
	return s == StatusAcknowledged
}

// CanBeRejected checks if the entry can be rejected.
func (s InboxStatus) CanBeRejected() bool {
	return s == StatusPending || s == StatusAcknowledged || s == StatusApproving
}

// DeviceAckAction represents actions a device can take.
type DeviceAckAction string

const (
	DeviceAckActionAcknowledge DeviceAckAction = "acknowledge" // Device acknowledges receipt
)

// OperatorAction represents actions an operator can take.
type OperatorAction string

const (
	OperatorActionApprove OperatorAction = "approve" // Operator approves registration
	OperatorActionReject OperatorAction = "reject"  // Operator rejects registration
	OperatorActionDelete OperatorAction = "delete"  // Operator deletes entry
)

// AckAction represents the action for legacy compatibility.
// Deprecated: Use DeviceAckAction or OperatorAction instead.
type AckAction string

const (
	AckActionApprove AckAction = "approve"
	AckActionReject AckAction = "reject"
)
