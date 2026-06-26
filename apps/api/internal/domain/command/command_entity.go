package command

import (
	"encoding/json"
	"errors"
	"time"
)

// ErrNotFound is returned when a command is not found.
var ErrNotFound = errors.New("command not found")

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
)

// Command represents a command to be sent to a device.
type Command struct {
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeliveredAt   *int64
	CompletedAt   *int64
	ID            string
	DeviceID      string
	DispatchID    string
	Command       string
	Status        Status
	FailureReason string
	Args          []byte
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
	return c.Status == StatusPending
}

// IsDelivered returns true if the command has been delivered.
func (c *Command) IsDelivered() bool {
	return c.Status == StatusDelivered
}

// IsCompleted returns true if the command has been completed.
func (c *Command) IsCompleted() bool {
	return c.Status == StatusCompleted
}

// IsFailed returns true if the command has failed.
func (c *Command) IsFailed() bool {
	return c.Status == StatusFailed
}

// IsFinished returns true if the command is in a terminal state.
func (c *Command) IsFinished() bool {
	return c.Status == StatusCompleted || c.Status == StatusFailed
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
