package notifications

import (
	"context"
	"encoding/json"
	"time"

	notificationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"
)

// WebhookChannel sends notifications via generic webhook POSTs.
type WebhookChannel struct {
	client *webhook.Client
}

// NewWebhookChannel creates a WebhookChannel.
func NewWebhookChannel(client *webhook.Client) *WebhookChannel {
	return &WebhookChannel{client: client}
}

// Send posts the message payload to the contact point's webhook URL.
func (c *WebhookChannel) Send(ctx context.Context, cp *notificationdomain.ContactPoint, msg *notificationdomain.Message) error {
	url := cp.Config["url"]
	payload := &webhook.Payload{
		Type:      webhook.EventType(msg.Event),
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"subject": msg.Subject, "body": msg.Body},
	}
	for k, v := range msg.Data {
		payload.Data[k] = v
	}
	return c.client.Send(ctx, url, cp.Secret, payload)
}

// SlackChannel formats payloads for Slack incoming webhooks.
type SlackChannel struct {
	client *webhook.Client
}

// NewSlackChannel creates a SlackChannel.
func NewSlackChannel(client *webhook.Client) *SlackChannel {
	return &SlackChannel{client: client}
}

// Send posts a Slack-compatible payload to the contact point's webhook.
func (c *SlackChannel) Send(ctx context.Context, cp *notificationdomain.ContactPoint, msg *notificationdomain.Message) error {
	url := cp.Config["webhook"]
	var body []byte
	var err error

	if msg.Event == "firing" || msg.Event == "resolved" || msg.Event == "pending" {
		ruleName := msg.Data["ruleName"]
		metric := msg.Data["metric"]
		condition := msg.Data["condition"]
		threshold := msg.Data["threshold"]
		value := msg.Data["value"]
		body, err = RenderSlackAlert(msg.Event, ruleName, metric, condition, threshold, value, nil)
	} else {
		text := msg.Subject
		if msg.Body != "" {
			text += "\n" + msg.Body
		}
		body, err = json.Marshal(map[string]interface{}{"text": text})
	}
	if err != nil {
		return err
	}

	return c.client.SendRaw(ctx, url, cp.Secret, body)
}
