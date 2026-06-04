package fcm

import (
	"context"
	"fmt"
	"log/slog"
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
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	c.log.Info("fcm silent wake prepared", slog.String("projectId", c.projectID), slog.String("deviceId", wake.DeviceID), slog.String("command", wake.Command), slog.String("dispatchId", wake.DispatchID))
	return nil
}
