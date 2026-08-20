package auth

import (
	"context"
	"encoding/json"
	"fmt"
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
	StrictHmac         *bool   `json:"strictHmac,omitempty"`
	LogBufferLimit     *int    `json:"logBufferLimit,omitempty"`
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
	// Get full operator settings in single query.
	settings, err := s.operatorRepo.GetOperatorSettings(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	return &SettingsResponse{
		Notifications: settings.Notifications,
		Client:        &settings.Client,
	}, nil
}

// UpdateClientSettings updates client settings for an operator.
func (s *ClientSettingsService) UpdateClientSettings(ctx context.Context, operatorID string, input *ClientSettingsInput) (*SettingsResponse, error) {
	if err := s.validateClientSettings(input); err != nil {
		return nil, err
	}

	currentSettings, err := s.operatorRepo.GetOperatorSettings(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	s.applyClientSettingsUpdates(&currentSettings.Client, input)

	if err = s.operatorRepo.UpdateClientSettings(ctx, operatorID, currentSettings.Client); err != nil {
		return nil, err
	}

	notifications, err := s.operatorRepo.GetNotifications(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	return &SettingsResponse{
		Notifications: notifications,
		Client:        &currentSettings.Client,
	}, nil
}

// validateClientSettings validates the client settings input.
func (s *ClientSettingsService) validateClientSettings(input *ClientSettingsInput) error {
	if input.RequestTimeoutMs != nil {
		if *input.RequestTimeoutMs < 500 || *input.RequestTimeoutMs > 60000 {
			return &ValidationError{Message: "requestTimeoutMs must be between 500 and 60000"}
		}
	}
	if input.LogBufferLimit != nil {
		if *input.LogBufferLimit < 50 || *input.LogBufferLimit > 5000 {
			return &ValidationError{Message: "logBufferLimit must be between 50 and 5000"}
		}
	}
	if input.SignalHistoryLimit != nil {
		if *input.SignalHistoryLimit < 30 || *input.SignalHistoryLimit > 2000 {
			return &ValidationError{Message: "signalHistoryLimit must be between 30 and 2000"}
		}
	}
	return s.validateServerURL(input.ServerURL)
}

// validateServerURL validates the server URL format.
func (s *ClientSettingsService) validateServerURL(serverURL *string) error {
	if serverURL == nil || *serverURL == "" {
		return nil
	}
	parsedURL, urlErr := url.Parse(*serverURL)
	if urlErr != nil {
		return &ValidationError{Message: "serverUrl must be a valid HTTP/HTTPS URL"}
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return &ValidationError{Message: "serverUrl must be a valid HTTP/HTTPS URL"}
	}
	if parsedURL.Host == "" {
		return &ValidationError{Message: "serverUrl must have a valid host"}
	}
	return nil
}

// applyClientSettingsUpdates applies input updates to client settings.
func (s *ClientSettingsService) applyClientSettingsUpdates(client *operator.ClientSettings, input *ClientSettingsInput) {
	if input.ServerURL != nil {
		client.ServerURL = *input.ServerURL
	}
	if input.DeviceID != nil {
		client.DeviceID = *input.DeviceID
	}
	if input.RequestTimeoutMs != nil {
		client.RequestTimeoutMs = *input.RequestTimeoutMs
	}
	if input.AutoReconnect != nil {
		client.AutoReconnect = *input.AutoReconnect
	}
	if input.StrictHmac != nil {
		client.StrictHmac = *input.StrictHmac
	}
	if input.LogBufferLimit != nil {
		client.LogBufferLimit = *input.LogBufferLimit
	}
	if input.SignalHistoryLimit != nil {
		client.SignalHistoryLimit = *input.SignalHistoryLimit
	}
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
	Notifications *operator.NotificationSettings `json:"notifications"`
	Client        *operator.ClientSettings       `json:"client,omitempty"`
	Preferences   map[string]any                 `json:"preferences,omitempty"`
}

var defaultPreferences = map[string]any{
	"theme":               "dark",
	"refresh_interval":    30,
	"notification_filter": "all",
	"language":            "en",
}

func (s *ClientSettingsService) GetPreferences(ctx context.Context, operatorID string) (map[string]any, error) {
	raw, err := s.operatorRepo.GetPreferencesRaw(ctx, operatorID)
	if err != nil {
		return nil, fmt.Errorf("get preferences: %w", err)
	}
	if raw == "" {
		return defaultPreferences, nil
	}
	prefs := make(map[string]any)
	if err := json.Unmarshal([]byte(raw), &prefs); err != nil {
		return defaultPreferences, fmt.Errorf("invalid preferences JSON: %w", err)
	}
	for k, v := range defaultPreferences {
		if _, exists := prefs[k]; !exists {
			prefs[k] = v
		}
	}
	return prefs, nil
}

func (s *ClientSettingsService) UpdatePreferences(ctx context.Context, operatorID string, updates map[string]any) (map[string]any, error) {
	current, err := s.GetPreferences(ctx, operatorID)
	if err != nil {
		current = make(map[string]any)
	}
	for k, v := range updates {
		if v == nil {
			delete(current, k)
		} else {
			current[k] = v
		}
	}
	data, err := json.Marshal(current)
	if err != nil {
		return nil, ErrValidationError
	}
	if err := s.operatorRepo.SavePreferencesRaw(ctx, operatorID, string(data)); err != nil {
		return nil, err
	}
	return current, nil
}
