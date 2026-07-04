package operator

import (
	"context"
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// NotificationService handles notification settings operations.
type NotificationService struct {
	operatorRepo operator.Repository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(repo operator.Repository) *NotificationService {
	return &NotificationService{operatorRepo: repo}
}

// GetNotifications retrieves notification settings for an operator.
func (s *NotificationService) GetNotifications(ctx context.Context, operatorID string) (*operator.NotificationSettings, error) {
	return s.operatorRepo.GetNotifications(ctx, operatorID)
}

// UpdateNotifications updates notification settings for an operator.
func (s *NotificationService) UpdateNotifications(ctx context.Context, operatorID string, input *operator.NotificationInput) (*operator.NotificationSettings, error) {
	// Get current notifications
	notifications, err := s.operatorRepo.GetNotifications(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if input.Enabled != nil {
		notifications.Enabled = *input.Enabled
	}

	if input.Channels != nil {
		notifications.Channels = *input.Channels
	}

	if input.Email != nil {
		if input.Email.ThresholdBreach != nil {
			notifications.Email.ThresholdBreach = *input.Email.ThresholdBreach
		}
		if input.Email.DeviceOffline != nil {
			notifications.Email.DeviceOffline = *input.Email.DeviceOffline
		}
		if input.Email.DeviceOnline != nil {
			notifications.Email.DeviceOnline = *input.Email.DeviceOnline
		}
		if input.Email.UpdateAvailable != nil {
			notifications.Email.UpdateAvailable = *input.Email.UpdateAvailable
		}
		if input.Email.CommandFailed != nil {
			notifications.Email.CommandFailed = *input.Email.CommandFailed
		}
		if input.Email.RegistrationRequest != nil {
			notifications.Email.RegistrationRequest = *input.Email.RegistrationRequest
		}
	}

	if input.Push != nil {
		if input.Push.ThresholdBreach != nil {
			notifications.Push.ThresholdBreach = *input.Push.ThresholdBreach
		}
		if input.Push.DeviceOffline != nil {
			notifications.Push.DeviceOffline = *input.Push.DeviceOffline
		}
		if input.Push.DeviceOnline != nil {
			notifications.Push.DeviceOnline = *input.Push.DeviceOnline
		}
		if input.Push.UpdateAvailable != nil {
			notifications.Push.UpdateAvailable = *input.Push.UpdateAvailable
		}
		if input.Push.CommandFailed != nil {
			notifications.Push.CommandFailed = *input.Push.CommandFailed
		}
		if input.Push.RegistrationRequest != nil {
			notifications.Push.RegistrationRequest = *input.Push.RegistrationRequest
		}
	}

	if input.Webhook != nil {
		if input.Webhook.Enabled != nil {
			notifications.Webhook.Enabled = *input.Webhook.Enabled
		}
		if input.Webhook.URL != nil {
			notifications.Webhook.URL = *input.Webhook.URL
		}
		if input.Webhook.Types != nil {
			notifications.Webhook.Types = input.Webhook.Types
		}
	}

	// Validate webhook URL if enabled
	if notifications.Webhook.Enabled && notifications.Webhook.URL == "" {
		return nil, errors.New("webhook URL is required when webhook is enabled")
	}

	// Save
	err = s.operatorRepo.UpdateNotifications(ctx, operatorID, notifications)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

// RotateWebhookSecret rotates the webhook secret for an operator.
func (s *NotificationService) RotateWebhookSecret(ctx context.Context, operatorID string) (string, error) {
	return s.operatorRepo.RotateWebhookSecret(ctx, operatorID)
}
