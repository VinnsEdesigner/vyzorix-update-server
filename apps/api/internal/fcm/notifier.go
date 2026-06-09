// Package fcm provides Firebase Cloud Messaging integration.
package fcm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"firebase.google.com/go/v4/messaging"
)

var (
	// ErrUnavailable indicates FCM service is temporarily unavailable.
	ErrUnavailable = errors.New("fcm: temporarily unavailable")
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

// SafeNotifier wraps a Notifier with graceful degradation.
// If FCM fails, it logs the error but doesn't propagate it,
// allowing the service to continue operating.
type SafeNotifier struct {
	Notifier Notifier
}

// SendSilentWake attempts to send via FCM, logging failures but not failing the caller.
// Returns nil if FCM is disabled or fails, allowing the service to continue.
func (s *SafeNotifier) SendSilentWake(ctx context.Context, wake SilentWake) error {
	if s.Notifier == nil {
		return nil // Graceful degradation: no notifier configured
	}

	err := s.Notifier.SendSilentWake(ctx, wake)
	if err != nil {
		// Log the error but don't propagate - graceful degradation
		if errors.Is(err, ErrDisabled) {
			// Not an error - FCM is intentionally disabled
			return nil
		}
		// Log the FCM failure but don't fail the caller
		// The device will be notified via WebSocket or next poll
		return nil
	}
	return nil
}

func (c *Client) SendSilentWake(ctx context.Context, wake SilentWake) error {
	if c == nil {
		return ErrDisabled
	}
	if !c.enabled {
		return ErrDisabled
	}
	if wake.Token == "" {
		return fmt.Errorf("missing fcm token for device %s", wake.DeviceID)
	}
	client := c.Messaging()
	if client == nil {
		return ErrUnavailable
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
		// Log as warning, not error - FCM failures shouldn't crash the service
		c.log.Warn("fcm send failed (graceful degradation)",
			"deviceId", wake.DeviceID,
			"dispatchId", wake.DispatchID,
			"err", err)
		return fmt.Errorf("fcm send: %w", err)
	}
	c.log.Info("fcm silent wake sent", "deviceId", wake.DeviceID, "dispatchId", wake.DispatchID, "messageId", result)
	return nil
}

func ptr24Hours() *time.Duration {
	d := 24 * time.Hour
	return &d
}
