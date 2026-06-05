package fcm

import (
	"context"
	"fmt"
	"time"

	"firebase.google.com/go/v4/messaging"
)

type SilentWake struct {
	Token      string
	Command    string
	DispatchID string
	DeviceID   string
}

type Notifier interface {
	SendSilentWake(ctx context.Context, wake SilentWake) error
}

func (c *Client) SendSilentWake(ctx context.Context, wake SilentWake) error {
	if c == nil || !c.enabled {
		return ErrDisabled
	}
	if wake.Token == "" {
		return fmt.Errorf("missing fcm token for device %s", wake.DeviceID)
	}
	client := c.Messaging()
	if client == nil {
		return fmt.Errorf("fcm client unavailable")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	msg := &messaging.Message{
		Token: wake.Token,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			TTL:      ptr24Hours(),
			Data: map[string]string{
				"action":      "WAKE_DAEMON",
				"command":     wake.Command,
				"dispatch_id": wake.DispatchID,
			},
		},
		Data: map[string]string{
			"action":      "WAKE_DAEMON",
			"command":     wake.Command,
			"dispatch_id": wake.DispatchID,
		},
	}

	result, err := client.Send(ctx, msg)
	if err != nil {
		c.log.Warn("fcm send failed", "deviceId", wake.DeviceID, "dispatchId", wake.DispatchID, "err", err)
		return fmt.Errorf("fcm send: %w", err)
	}
	c.log.Info("fcm silent wake sent", "deviceId", wake.DeviceID, "dispatchId", wake.DispatchID, "messageId", result)
	return nil
}

func ptr24Hours() *time.Duration {
	d := 24 * time.Hour
	return &d
}
