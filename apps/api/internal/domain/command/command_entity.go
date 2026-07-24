package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned when a command is not found.
var ErrNotFound = errors.New("command not found")

// ErrInvalidCommandTransition is returned when an invalid command status transition is attempted.
var ErrInvalidCommandTransition = errors.New("invalid command status transition")

// Status represents the status of a command.
type Status string

const (
	// StatusPending indicates the command is waiting to be processed.
	StatusPending Status = "pending"
	// StatusDelivered indicates the command has been delivered to the device.
	StatusDelivered Status = "delivered"
	// StatusCompleted indicates the command has been completed.
	StatusCompleted Status = "completed"
	// StatusFailed indicates the command has failed.
	StatusFailed Status = "failed"
	// StatusCancelled indicates the command was cancelled.
	StatusCancelled Status = "cancelled"
)

// CommandStatusTransitions defines valid state transitions for command lifecycle.
// The map key is the current state, and the value is the set of allowed next states.
var CommandStatusTransitions = map[Status]map[Status]bool{
	StatusPending: {
		StatusDelivered: true,
		StatusFailed:    true,
		StatusCancelled: true,
	},
	StatusDelivered: {
		StatusCompleted: true,
		StatusFailed:    true,
		StatusCancelled: true,
	},
	StatusCompleted: {}, // Terminal state.
	StatusFailed:    {}, // Terminal state.
	StatusCancelled: {}, // Terminal state.
}

// CanTransitionTo returns true if the status can transition to the target status.
func (s Status) CanTransitionTo(target Status) bool {
	allowed, exists := CommandStatusTransitions[s]
	if !exists {
		return false
	}
	return allowed[target]
}

// TransitionTo transitions the status to a new state.
// Returns ErrInvalidCommandTransition if the transition is not allowed.
func (s *Status) TransitionTo(target Status) error {
	if !s.CanTransitionTo(target) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidCommandTransition, *s, target)
	}
	*s = target
	return nil
}

// IsPending checks if the command is pending.
func (s Status) IsPending() bool {
	return s == StatusPending
}

// IsDelivered checks if the command has been delivered.
func (s Status) IsDelivered() bool {
	return s == StatusDelivered
}

// IsCompleted checks if the command has been completed.
func (s Status) IsCompleted() bool {
	return s == StatusCompleted
}

// IsFailed checks if the command has failed.
func (s Status) IsFailed() bool {
	return s == StatusFailed
}

// IsCancelled checks if the command was cancelled.
func (s Status) IsCancelled() bool {
	return s == StatusCancelled
}

// IsTerminal checks if the status is a terminal state.
func (s Status) IsTerminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

// IsActive checks if the command is in an active state (pending or delivered).
func (s Status) IsActive() bool {
	return s == StatusPending || s == StatusDelivered
}

// Command type constants.
const (
	// TypeWakeUpUpdater is sent to wake a device to check for updates.
	TypeWakeUpUpdater = "WAKE_UP_UPDATER"
	// TypeCheckUpdate is sent to trigger an update check on a device.
	TypeCheckUpdate = "CHECK_UPDATE"
)

// Command represents a command to be sent to a device.
type Command struct {
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeliveredAt   *int64
	CompletedAt   *int64
	ExpiresAt     *time.Time
	NextRetryAt   *time.Time
	DispatchID    string
	Command       string
	Status        Status
	FailureReason string
	DeviceID      string
	ID            string
	Args          []byte
	RetryCount    int
	MaxRetries    int
}

// IsExpired returns true if the command has expired based on its TTL.
func (c *Command) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false // No TTL set, never expires.
	}
	return time.Now().After(*c.ExpiresAt)
}

// CommandFrame is the internal representation of a command for the WebSocket hub.
type CommandFrame struct {
	DeliveryConfirmation chan<- bool     `json:"-"`
	Type                 string          `json:"type"`
	DispatchID           string          `json:"dispatchId"`
	Command              string          `json:"command"`
	Nonce                string          `json:"nonce"`
	Signature            string          `json:"signature,omitempty"`
	Args                 json.RawMessage `json:"args,omitempty"`
	Timestamp            int64           `json:"timestamp"`
}

// IsPending returns true if the command is pending.
func (c *Command) IsPending() bool {
	return c.Status.IsPending()
}

// IsDelivered returns true if the command has been delivered.
func (c *Command) IsDelivered() bool {
	return c.Status.IsDelivered()
}

// IsCompleted returns true if the command has been completed.
func (c *Command) IsCompleted() bool {
	return c.Status.IsCompleted()
}

// IsFailed returns true if the command has failed.
func (c *Command) IsFailed() bool {
	return c.Status.IsFailed()
}

// IsFinished returns true if the command is in a terminal state.
func (c *Command) IsFinished() bool {
	return c.Status.IsTerminal()
}

// DeliveredAtTime returns the DeliveredAt as a time.Time.
func (c *Command) DeliveredAtTime() *time.Time {
	if c.DeliveredAt == nil {
		return nil
	}

	t := time.UnixMilli(*c.DeliveredAt)

	return &t
}

// CompletedAtTime returns the CompletedAt as a time.Time.
func (c *Command) CompletedAtTime() *time.Time {
	if c.CompletedAt == nil {
		return nil
	}

	t := time.UnixMilli(*c.CompletedAt)

	return &t
}

// SetArgs sets the arguments from a struct (JSON encoded).
func (c *Command) SetArgs(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	c.Args = data

	return nil
}

// GetArgs unmarshals the arguments into a struct.
func (c *Command) GetArgs(v interface{}) error {
	return json.Unmarshal(c.Args, v)
}

// CanRetry returns true if the command can be retried.
func (c *Command) CanRetry() bool {
	return c.MaxRetries == 0 || c.RetryCount < c.MaxRetries
}

// ShouldRetryNow returns true if enough time has passed since the last attempt.
func (c *Command) ShouldRetryNow() bool {
	if c.NextRetryAt == nil {
		return true // First attempt or no backoff set.
	}
	return time.Now().After(*c.NextRetryAt)
}

// IncrementRetry sets up the command for the next retry attempt.
// Returns false if max retries exceeded.
func (c *Command) IncrementRetry(baseDelay time.Duration) bool {
	c.RetryCount++
	if c.MaxRetries > 0 && c.RetryCount >= c.MaxRetries {
		return false
	}
	// Exponential backoff: baseDelay * 2^(retryCount-1).
	delay := baseDelay * time.Duration(1<<(c.RetryCount-1))
	next := time.Now().Add(delay)
	c.NextRetryAt = &next
	return true
}
