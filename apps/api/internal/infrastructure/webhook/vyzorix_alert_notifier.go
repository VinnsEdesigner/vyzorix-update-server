package webhook

import (
	"context"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
)

// Alert notifier event types. Reuses the payload contract dashboards already
// understand (threshold_breach) and adds alert_resolved for recoveries.
const (
	EventTypeAlertFiring   EventType = EventTypeThresholdBreach
	EventTypeAlertResolved EventType = "alert_resolved"
)

// AlertNotifier delivers alert state transitions to the rule's webhook URL
// through the SSRF-safe, retrying Client. It implements alert.Notifier.
type AlertNotifier struct {
	client *Client
	secret string
}

// NewAlertNotifier creates a notifier; secret enables HMAC signatures when set.
func NewAlertNotifier(client *Client, secret string) *AlertNotifier {
	return &AlertNotifier{client: client, secret: secret}
}

// Notify sends a firing/resolved transition to the rule's webhook URL.
func (n *AlertNotifier) Notify(ctx context.Context, notif *alert.Notification) error {
	eventType := EventTypeAlertFiring
	if notif.Transition.Resolved() {
		eventType = EventTypeAlertResolved
	}

	rule := notif.Rule
	payload := &Payload{
		Type:      eventType,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"ruleId":    rule.ID,
			"ruleName":  rule.Name,
			"orgId":     rule.OrgID,
			"metric":    string(rule.Metric),
			"condition": string(rule.Condition),
			"threshold": rule.Threshold,
			"value":     notif.Transition.Value,
			"fromState": string(notif.Transition.From),
			"toState":   string(notif.Transition.To),
		},
	}
	return n.client.Send(ctx, rule.WebhookURL, n.secret, payload)
}
