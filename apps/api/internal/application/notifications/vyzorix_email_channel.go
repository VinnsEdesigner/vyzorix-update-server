// Package notifications provides contact point channels and the dispatcher
// that routes notifications to them.
package notifications

import (
	"context"
	"errors"

	notificationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
)

// EmailChannel sends notifications through the Resend-backed email service.
type EmailChannel struct {
	service *email.Service
}

// NewEmailChannel creates an EmailChannel.
func NewEmailChannel(service *email.Service) *EmailChannel {
	return &EmailChannel{service: service}
}

// Send delivers the message via the email service.
func (c *EmailChannel) Send(ctx context.Context, cp *notificationdomain.ContactPoint, msg *notificationdomain.Message) error {
	to := cp.Config["to"]
	if to == "" {
		return errors.New("email contact point missing 'to'")
	}
	if msg.HTML != "" {
		return c.service.SendNotificationEmail(ctx, to, msg.Subject, msg.HTML)
	}
	return c.service.SendNotificationEmail(ctx, to, msg.Subject, msg.Body)
}
