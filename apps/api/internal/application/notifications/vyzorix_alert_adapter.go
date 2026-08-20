package notifications

import (
	"context"
	"fmt"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
	notificationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
)

// AlertNotifierAdapter routes alert notifications through the contact-point
// dispatcher instead of a single per-rule webhook URL. It implements
// alert.Notifier.
type AlertNotifierAdapter struct {
	dispatcher *Dispatcher
}

// NewAlertNotifierAdapter creates an AlertNotifierAdapter.
func NewAlertNotifierAdapter(dispatcher *Dispatcher) *AlertNotifierAdapter {
	return &AlertNotifierAdapter{dispatcher: dispatcher}
}

// Notify builds a message from the transition and dispatches it to all
// enabled contact points of the rule's org.
func (a *AlertNotifierAdapter) Notify(ctx context.Context, n *alert.Notification) error {
	if a.dispatcher == nil {
		return nil
	}
	msg := buildAlertMessage(n)
	if _, err := a.dispatcher.Send(ctx, n.Rule.OrgID, msg); err != nil {
		return fmt.Errorf("dispatch alert notification: %w", err)
	}
	return nil
}

func buildAlertMessage(n *alert.Notification) *notificationdomain.Message {
	rule := n.Rule
	tr := n.Transition

	subject := fmt.Sprintf("[%s] %s", tr.To, rule.Name)
	body := fmt.Sprintf("Metric %s %s %v (threshold %v) → %s; value %v",
		rule.Metric, rule.Condition, rule.Threshold, rule.Threshold, tr.To, tr.Value)

	return &notificationdomain.Message{
		Subject: subject,
		Body:    body,
		Event:   string(tr.To),
		Data: map[string]string{
			"ruleId":    rule.ID,
			"ruleName":  rule.Name,
			"orgId":     rule.OrgID,
			"metric":    string(rule.Metric),
			"condition": string(rule.Condition),
			"threshold": fmt.Sprintf("%v", rule.Threshold),
			"value":     fmt.Sprintf("%v", tr.Value),
			"fromState": string(tr.From),
			"toState":   string(tr.To),
		},
	}
}
