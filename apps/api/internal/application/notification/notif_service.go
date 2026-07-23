// Package notification provides the notification service for sending alerts.
package notification

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	infranotification "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/notification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"
)

// EventType represents the type of event that triggered a notification.
type EventType string

const (
	EventTypeThresholdBreach     EventType = "threshold_breach"
	EventTypeDeviceOffline       EventType = "device_offline"
	EventTypeDeviceOnline        EventType = "device_online"
	EventTypeUpdateAvailable     EventType = "update_available"
	EventTypeCommandFailed       EventType = "command_failed"
	EventTypeRegistrationRequest EventType = "registration_request"
	EventTypeError               EventType = "error"
)

// EventData contains the data for a notification event.
type EventData struct {
	Timestamp     time.Time
	AlertType     string
	Threshold     string
	OperatorID    string
	OperatorEmail string
	OperatorName  string
	EventType     EventType
	CurrentValue  string
	DeviceName    string
	CommandName   string
	FailureReason string
	UpdateVersion string
	RequesterName string
	ErrorMessage  string
	DeviceID      string
}

// Service handles sending notifications based on operator preferences.
type Service struct {
	operatorRepo  operator.Repository
	emailService  *email.Service
	webhookClient *webhook.Client
	auditRepo     *infranotification.Repository
	logger        *slog.Logger
}

// NewService creates a new notification service.
func NewService(
	operatorRepo operator.Repository,
	emailSvc *email.Service,
	webhookClient *webhook.Client,
	auditRepo *infranotification.Repository,
	logger *slog.Logger,
) *Service {
	return &Service{
		operatorRepo:  operatorRepo,
		emailService:  emailSvc,
		webhookClient: webhookClient,
		auditRepo:     auditRepo,
		logger:        logger,
	}
}

// SendNotification sends a notification based on operator preferences and event type.
func (s *Service) SendNotification(ctx context.Context, data EventData) error {
	// Get operator's notification settings if not provided.
	if data.OperatorEmail == "" || data.OperatorName == "" {
		op, err := s.operatorRepo.FindByID(ctx, data.OperatorID)
		if err != nil {
			return fmt.Errorf("failed to get operator: %w", err)
		}
		if data.OperatorEmail == "" {
			data.OperatorEmail = op.Email
		}
		if data.OperatorName == "" {
			data.OperatorName = op.Name
		}
	}

	// Get operator's notification settings.
	settings, err := s.operatorRepo.GetNotifications(ctx, data.OperatorID)
	if err != nil {
		return fmt.Errorf("failed to get notification settings: %w", err)
	}

	// Check if notifications are enabled.
	if !settings.Enabled {
		s.logger.Debug("notifications disabled for operator", "operatorID", data.OperatorID)
		return nil
	}

	// Build common email data.
	emailData := email.NotificationData{
		OperatorName:  data.OperatorName,
		DeviceID:      data.DeviceID,
		DeviceName:    data.DeviceName,
		AlertType:     data.AlertType,
		CurrentValue:  data.CurrentValue,
		Threshold:     data.Threshold,
		CommandName:   data.CommandName,
		FailureReason: data.FailureReason,
		UpdateVersion: data.UpdateVersion,
		RequesterName: data.RequesterName,
		ErrorMessage:  data.ErrorMessage,
		Timestamp:     data.Timestamp.Format("2006-01-02 15:04:05 MST"),
		BaseURL:       "https://app.vyzorix.com",
	}

	// Send email if enabled for this event type.
	if s.shouldSendEmail(settings, data.EventType) {
		if err := s.sendEmail(ctx, data, emailData); err != nil {
			s.logger.Error("failed to send email notification", "error", err, "operatorID", data.OperatorID)
			// Don't return error - try webhook even if email fails.
		}
	}

	// Send webhook if configured.
	if err := s.sendWebhook(ctx, data, settings); err != nil {
		s.logger.Error("failed to send webhook notification", "error", err, "operatorID", data.OperatorID)
		// Don't return error for webhook failures.
	}

	return nil
}

// shouldSendEmail checks if email should be sent for this event type.
func (s *Service) shouldSendEmail(settings *operator.NotificationSettings, eventType EventType) bool {
	switch eventType {
	case EventTypeThresholdBreach:
		return settings.Email.ThresholdBreach
	case EventTypeDeviceOffline:
		return settings.Email.DeviceOffline
	case EventTypeDeviceOnline:
		return settings.Email.DeviceOnline
	case EventTypeUpdateAvailable:
		return settings.Email.UpdateAvailable
	case EventTypeCommandFailed:
		return settings.Email.CommandFailed
	case EventTypeRegistrationRequest:
		return settings.Email.RegistrationRequest
	case EventTypeError:
		// Error events are always sent if notifications are enabled.
		return true
	default:
		return false
	}
}

// shouldSendWebhook checks if webhook should be sent for this event type.
func (s *Service) shouldSendWebhook(settings *operator.NotificationSettings, eventType EventType) bool {
	if !settings.Webhook.Enabled || settings.Webhook.URL == "" {
		return false
	}

	// If no types specified, send all.
	if len(settings.Webhook.Types) == 0 {
		return true
	}

	// Check if this event type is in the allowed types.
	for _, t := range settings.Webhook.Types {
		if t == string(eventType) {
			return true
		}
	}

	return false
}

// sendEmail sends an email notification.
func (s *Service) sendEmail(ctx context.Context, data EventData, emailData email.NotificationData) error {
	if s.emailService == nil {
		s.logger.Warn("email service not configured")
		return nil
	}

	var err error

	switch data.EventType {
	case EventTypeThresholdBreach:
		err = s.emailService.SendThresholdBreachEmail(ctx, data.OperatorEmail, emailData)
	case EventTypeDeviceOffline:
		err = s.emailService.SendDeviceOfflineEmail(ctx, data.OperatorEmail, emailData)
	case EventTypeDeviceOnline:
		err = s.emailService.SendDeviceOnlineEmail(ctx, data.OperatorEmail, emailData)
	case EventTypeCommandFailed:
		err = s.emailService.SendCommandFailedEmail(ctx, data.OperatorEmail, emailData)
	case EventTypeUpdateAvailable:
		err = s.emailService.SendUpdateAvailableEmail(ctx, data.OperatorEmail, emailData)
	case EventTypeRegistrationRequest:
		err = s.emailService.SendRegistrationRequestEmail(ctx, data.OperatorEmail, emailData)
	case EventTypeError:
		err = s.emailService.SendErrorAlertEmail(ctx, data.OperatorEmail, emailData)
	default:
		s.logger.Warn("unknown event type for email", "eventType", data.EventType)
		return nil
	}

	if err != nil {
		s.auditNotification(ctx, data, "email", false, err.Error())
		return err
	}

	s.auditNotification(ctx, data, "email", true, "")
	s.logger.Info("email notification sent",
		"eventType", data.EventType,
		"operatorID", data.OperatorID,
		"deviceID", data.DeviceID,
	)

	return nil
}

// sendWebhook sends a webhook notification.
func (s *Service) sendWebhook(ctx context.Context, data EventData, settings *operator.NotificationSettings) error {
	if s.webhookClient == nil {
		s.logger.Warn("webhook client not configured")
		return nil
	}

	if !s.shouldSendWebhook(settings, data.EventType) {
		return nil
	}

	payload := &webhook.Payload{
		Type:       webhook.EventType(data.EventType),
		Timestamp:  data.Timestamp,
		DeviceID:   data.DeviceID,
		OperatorID: data.OperatorID,
		Data: map[string]interface{}{
			"deviceName":    data.DeviceName,
			"alertType":     data.AlertType,
			"currentValue":  data.CurrentValue,
			"threshold":     data.Threshold,
			"commandName":   data.CommandName,
			"failureReason": data.FailureReason,
			"updateVersion": data.UpdateVersion,
			"requesterName": data.RequesterName,
			"errorMessage":  data.ErrorMessage,
		},
	}

	err := s.webhookClient.Send(ctx, settings.Webhook.URL, settings.Webhook.Secret, payload)
	if err != nil {
		s.auditNotification(ctx, data, "webhook", false, err.Error())
		return err
	}

	s.auditNotification(ctx, data, "webhook", true, "")
	s.logger.Info("webhook notification sent",
		"eventType", data.EventType,
		"operatorID", data.OperatorID,
		"deviceID", data.DeviceID,
		"webhookURL", settings.Webhook.URL,
	)

	return nil
}

// auditNotification logs a notification to the audit trail.
func (s *Service) auditNotification(ctx context.Context, data EventData, channel string, success bool, errorMsg string) {
	if s.auditRepo == nil {
		return
	}

	entry := &infranotification.AuditEntry{
		ID:         generateID(),
		OperatorID: data.OperatorID,
		EventType:  string(data.EventType),
		Channel:    channel,
		DeviceID:   data.DeviceID,
		Success:    success,
		ErrorMsg:   errorMsg,
		SentAt:     time.Now(),
	}

	if err := s.auditRepo.LogEntry(ctx, entry); err != nil {
		s.logger.Error("failed to audit notification", "error", err)
	}
}

// generateID generates a unique ID for audit entries.
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "audit_" + hex.EncodeToString(b)
}
