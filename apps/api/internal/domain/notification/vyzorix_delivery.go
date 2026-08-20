package notification

import (
	"context"
	"errors"
	"time"
)

// Message is a notification payload delivered through a channel.
type Message struct {
	Subject string
	Body    string
	HTML    string
	Data    map[string]string
	Event   string
}

// Channel is a transport for a Message. Implementations must be safe for
// concurrent use; the dispatcher may invoke them from multiple goroutines.
type Channel interface {
	Send(ctx context.Context, cp *ContactPoint, msg *Message) error
}

// Delivery records one attempt to deliver a Message through a contact point.
type Delivery struct {
	CreatedAt      time.Time
	Message        *Message
	ID             string
	ContactPointID string
	Channel        ChannelType
	Error          string
	Status string
}

// DeliveryRepository persists delivery records for audit/replay.
type DeliveryRepository interface {
	Append(ctx context.Context, d *Delivery) error
	ListByContactPoint(ctx context.Context, contactPointID string, limit int) ([]*Delivery, error)
}

// ErrInvalidMessage is returned when a Message fails validation.
var ErrInvalidMessage = errors.New("invalid notification message")

// Validate checks the Message has minimum content.
func (m *Message) Validate() error {
	if m.Subject == "" && m.Body == "" && m.HTML == "" && m.Event == "" {
		return ErrInvalidMessage
	}
	return nil
}
