package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	notificationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/uuid"
)

// Dispatcher routes notifications to contact points via their channel,
// records delivery attempts, and emits metrics.
type Dispatcher struct {
	repo       notificationdomain.Repository
	deliveries notificationdomain.DeliveryRepository
	channels   map[notificationdomain.ChannelType]notificationdomain.Channel
	logger     *slog.Logger
}

// NewDispatcher creates a Dispatcher.
func NewDispatcher(repo notificationdomain.Repository, deliveries notificationdomain.DeliveryRepository, channels map[notificationdomain.ChannelType]notificationdomain.Channel, logger *slog.Logger) *Dispatcher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Dispatcher{repo: repo, deliveries: deliveries, channels: channels, logger: logger}
}

// Send dispatches a message to every enabled contact point of an org.
// Returns the number of successful sends; individual failures are logged
// but do not stop other contact points from receiving the message.
func (d *Dispatcher) Send(ctx context.Context, orgID string, msg *notificationdomain.Message) (int, error) {
	if err := msg.Validate(); err != nil {
		return 0, fmt.Errorf("notification message invalid: %w", err)
	}
	points, err := d.repo.ListEnabledByOrg(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("list contact points: %w", err)
	}
	if len(points) == 0 {
		return 0, nil
	}

	succeeded := 0
	for _, cp := range points {
		if err := d.dispatchOne(ctx, cp, msg); err != nil {
			d.logger.Error("notification delivery failed",
				"contact_point_id", cp.ID, "channel", cp.Channel, "error", err)
			continue
		}
		succeeded++
	}
	return succeeded, nil
}

// SendToPoint dispatches a message to a specific contact point (test button).
func (d *Dispatcher) SendToPoint(ctx context.Context, cp *notificationdomain.ContactPoint, msg *notificationdomain.Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("notification message invalid: %w", err)
	}
	return d.dispatchOne(ctx, cp, msg)
}

func (d *Dispatcher) dispatchOne(ctx context.Context, cp *notificationdomain.ContactPoint, msg *notificationdomain.Message) error {
	channel, ok := d.channels[cp.Channel]
	if !ok {
		return fmt.Errorf("no channel registered for %q", cp.Channel)
	}
	err := channel.Send(ctx, cp, msg)
	status := "sent"
	errStr := ""
	if err != nil {
		status = "failed"
		errStr = err.Error()
		metrics.Get().RecordNotificationDelivery(string(cp.Channel), "failed")
	} else {
		metrics.Get().RecordNotificationDelivery(string(cp.Channel), "sent")
	}

	delivery := &notificationdomain.Delivery{
		ID:             uuid.New(),
		ContactPointID: cp.ID,
		Channel:        cp.Channel,
		Status:         status,
		Error:          errStr,
		Message:        msg,
		CreatedAt:      time.Now(),
	}
	if d.deliveries != nil {
		if appendErr := d.deliveries.Append(ctx, delivery); appendErr != nil {
			d.logger.Error("delivery record append failed", "error", appendErr)
		}
	}
	return err
}
