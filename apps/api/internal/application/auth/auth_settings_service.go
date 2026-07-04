package auth

import (
	"context"
	"net/url"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// ErrValidationError is returned when validation fails.
var ErrValidationError = &ValidationError{}

// ValidationError represents a validation error.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ClientSettingsInput represents client settings input from the API.
type ClientSettingsInput struct {
	ServerURL          *string `json:"serverUrl,omitempty"`
	DeviceID           *string `json:"deviceId,omitempty"`
	RequestTimeoutMs   *int    `json:"requestTimeoutMs,omitempty"`
	AutoReconnect      *bool   `json:"autoReconnect,omitempty"`
	StrictHmac        *bool   `json:"strictHmac,omitempty"`
	LogBufferLimit    *int    `json:"logBufferLimit,omitempty"`
	SignalHistoryLimit *int    `json:"signalHistoryLimit,omitempty"`
}

// ClientSettingsService handles client settings operations.
type ClientSettingsService struct {
	operatorRepo operator.Repository
}

// NewClientSettingsService creates a new ClientSettingsService.
func NewClientSettingsService(repo operator.Repository) *ClientSettingsService {
	return &ClientSettingsService{operatorRepo: repo}
}

// GetSettings retrieves all settings for an operator.
func (s *ClientSettingsService) GetSettings(ctx context.Context, operatorID string) (*SettingsResponse, error) {
	// Get full operator settings in single query
	settings, err := s.operatorRepo.GetOperatorSettings(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	return &SettingsResponse{
		Thresholds:    settings.Thresholds,
		Notifications: settings.Notifications,
		Client:       &settings.Client,
	}, nil
}

// UpdateClientSettings updates client settings for an operator.
func (s *ClientSettingsService) UpdateClientSettings(ctx context.Context, operatorID string, input *ClientSettingsInput) (*SettingsResponse, error) {
	// Validate input
	if input.RequestTimeoutMs != nil {
		if *input.RequestTimeoutMs < 500 || *input.RequestTimeoutMs > 60000 {
			return nil, &ValidationError{Message: "requestTimeoutMs must be between 500 and 60000"}
		}
	}

	if input.LogBufferLimit != nil {
		if *input.LogBufferLimit < 50 || *input.LogBufferLimit > 5000 {
			return nil, &ValidationError{Message: "logBufferLimit must be between 50 and 5000"}
		}
	}

	if input.SignalHistoryLimit != nil {
		if *input.SignalHistoryLimit < 30 || *input.SignalHistoryLimit > 2000 {
			return nil, &ValidationError{Message: "signalHistoryLimit must be between 30 and 2000"}
		}
	}

	// Validate ServerURL format if provided
	if input.ServerURL != nil && *input.ServerURL != "" {
		parsedURL, urlErr := url.Parse(*input.ServerURL)
		if urlErr != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			return nil, &ValidationError{Message: "serverUrl must be a valid HTTP/HTTPS URL"}
		}
		if parsedURL.Host == "" {
			return nil, &ValidationError{Message: "serverUrl must have a valid host"}
		}
	}

	// Get current operator settings to merge with input
	currentSettings, err := s.operatorRepo.GetOperatorSettings(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	// Apply updates from input to client settings
	if input.ServerURL != nil {
		currentSettings.Client.ServerURL = *input.ServerURL
	}
	if input.DeviceID != nil {
		currentSettings.Client.DeviceID = *input.DeviceID
	}
	if input.RequestTimeoutMs != nil {
		currentSettings.Client.RequestTimeoutMs = *input.RequestTimeoutMs
	}
	if input.AutoReconnect != nil {
		currentSettings.Client.AutoReconnect = *input.AutoReconnect
	}
	if input.StrictHmac != nil {
		currentSettings.Client.StrictHmac = *input.StrictHmac
	}
	if input.LogBufferLimit != nil {
		currentSettings.Client.LogBufferLimit = *input.LogBufferLimit
	}
	if input.SignalHistoryLimit != nil {
		currentSettings.Client.SignalHistoryLimit = *input.SignalHistoryLimit
	}

	// Save updated client settings
	if err = s.operatorRepo.UpdateClientSettings(ctx, operatorID, currentSettings.Client); err != nil {
		return nil, err
	}

	// Get full settings to return
	thresholds, err := s.operatorRepo.GetThresholds(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	notifications, err := s.operatorRepo.GetNotifications(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	return &SettingsResponse{
		Thresholds:    thresholds,
		Notifications: notifications,
		Client:       &currentSettings.Client,
	}, nil
}

// ResetSettings resets all settings to defaults.
func (s *ClientSettingsService) ResetSettings(ctx context.Context, operatorID string) (*SettingsResponse, error) {
	err := s.operatorRepo.ResetSettings(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	return s.GetSettings(ctx, operatorID)
}

// SettingsResponse represents the complete settings response.
type SettingsResponse struct {
	Thresholds    operator.Thresholds           `json:"thresholds"`
	Notifications *operator.NotificationSettings `json:"notifications"`
	Client        *operator.ClientSettings     `json:"client,omitempty"`
}
