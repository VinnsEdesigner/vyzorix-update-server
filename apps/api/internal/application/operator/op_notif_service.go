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
	notifications, err := s.operatorRepo.GetNotifications(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	s.applyNotificationUpdates(input, notifications)

	err = s.validateNotificationSettings(notifications)
	if err != nil {
		return nil, err
	}

	err = s.operatorRepo.UpdateNotifications(ctx, operatorID, notifications)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

// applyNotificationUpdates applies input updates to notification settings.
func (s *NotificationService) applyNotificationUpdates(input *operator.NotificationInput, notifications *operator.NotificationSettings) {
	if input.Enabled != nil {
		notifications.Enabled = *input.Enabled
	}
	if input.Channels != nil {
		notifications.Channels = *input.Channels
	}

	s.applyEmailUpdates(input.Email, &notifications.Email)
	s.applyPushUpdates(input.Push, &notifications.Push)
	s.applyWebhookUpdates(input.Webhook, &notifications.Webhook)
}

// applyEmailUpdates applies email notification updates.
func (s *NotificationService) applyEmailUpdates(input *operator.EmailNotificationInput, email *operator.EmailNotifications) {
	if input == nil {
		return
	}
	if input.ThresholdBreach != nil {
		email.ThresholdBreach = *input.ThresholdBreach
	}
	if input.DeviceOffline != nil {
		email.DeviceOffline = *input.DeviceOffline
	}
	if input.DeviceOnline != nil {
		email.DeviceOnline = *input.DeviceOnline
	}
	if input.UpdateAvailable != nil {
		email.UpdateAvailable = *input.UpdateAvailable
	}
	if input.CommandFailed != nil {
		email.CommandFailed = *input.CommandFailed
	}
	if input.RegistrationRequest != nil {
		email.RegistrationRequest = *input.RegistrationRequest
	}
}

// applyPushUpdates applies push notification updates.
func (s *NotificationService) applyPushUpdates(input *operator.PushNotificationInput, push *operator.PushNotifications) {
	if input == nil {
		return
	}
	if input.ThresholdBreach != nil {
		push.ThresholdBreach = *input.ThresholdBreach
	}
	if input.DeviceOffline != nil {
		push.DeviceOffline = *input.DeviceOffline
	}
	if input.DeviceOnline != nil {
		push.DeviceOnline = *input.DeviceOnline
	}
	if input.UpdateAvailable != nil {
		push.UpdateAvailable = *input.UpdateAvailable
	}
	if input.CommandFailed != nil {
		push.CommandFailed = *input.CommandFailed
	}
	if input.RegistrationRequest != nil {
		push.RegistrationRequest = *input.RegistrationRequest
	}
}

// applyWebhookUpdates applies webhook notification updates.
func (s *NotificationService) applyWebhookUpdates(input *operator.WebhookNotificationInput, webhook *operator.WebhookNotifications) {
	if input == nil {
		return
	}
	if input.Enabled != nil {
		webhook.Enabled = *input.Enabled
	}
	if input.URL != nil {
		webhook.URL = *input.URL
	}
	if input.Types != nil {
		webhook.Types = input.Types
	}
}

// validateNotificationSettings validates the notification settings.
func (s *NotificationService) validateNotificationSettings(notifications *operator.NotificationSettings) error {
	if notifications.Webhook.Enabled && notifications.Webhook.URL == "" {
		return errors.New("webhook URL is required when webhook is enabled")
	}
	return nil
}

// RotateWebhookSecret rotates the webhook secret for an operator.
func (s *NotificationService) RotateWebhookSecret(ctx context.Context, operatorID string) (string, error) {
	return s.operatorRepo.RotateWebhookSecret(ctx, operatorID)
}
