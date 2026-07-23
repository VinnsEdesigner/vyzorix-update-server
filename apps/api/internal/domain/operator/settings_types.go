package operator

import "errors"

// ErrValidation is returned when threshold validation fails.
var ErrValidation = errors.New("validation error")

// ThresholdsInput represents threshold update input from the API.
type ThresholdsInput struct {
	RiskWarn    *int `json:"riskWarn,omitempty"`
	RiskCrit    *int `json:"riskCrit,omitempty"`
	ThermalWarn *int `json:"thermalWarn,omitempty"`
	ThermalCrit *int `json:"thermalCrit,omitempty"`
	BufferWarn  *int `json:"bufferWarn,omitempty"`
	BufferCrit  *int `json:"bufferCrit,omitempty"`
}

// Validate validates the threshold input.
func (t *ThresholdsInput) Validate() error {
	// If both risk values are provided, validate relationship.
	if t.RiskWarn != nil && t.RiskCrit != nil {
		if *t.RiskWarn >= *t.RiskCrit {
			return errors.New("riskWarn must be less than riskCrit")
		}
	}

	// If both thermal values are provided, validate relationship.
	if t.ThermalWarn != nil && t.ThermalCrit != nil {
		if *t.ThermalWarn >= *t.ThermalCrit {
			return errors.New("thermalWarn must be less than thermalCrit")
		}
	}

	// If both buffer values are provided, validate relationship (inverted).
	if t.BufferWarn != nil && t.BufferCrit != nil {
		if *t.BufferCrit >= *t.BufferWarn {
			return errors.New("bufferCrit must be less than bufferWarn")
		}
	}

	return nil
}

// EmailNotificationInput represents email notification settings input.
type EmailNotificationInput struct {
	ThresholdBreach     *bool `json:"thresholdBreach,omitempty"`
	DeviceOffline      *bool `json:"deviceOffline,omitempty"`
	DeviceOnline       *bool `json:"deviceOnline,omitempty"`
	UpdateAvailable    *bool `json:"updateAvailable,omitempty"`
	CommandFailed      *bool `json:"commandFailed,omitempty"`
	RegistrationRequest *bool `json:"registrationRequest,omitempty"`
}

// PushNotificationInput represents push notification settings input.
type PushNotificationInput struct {
	ThresholdBreach     *bool `json:"thresholdBreach,omitempty"`
	DeviceOffline      *bool `json:"deviceOffline,omitempty"`
	DeviceOnline       *bool `json:"deviceOnline,omitempty"`
	UpdateAvailable    *bool `json:"updateAvailable,omitempty"`
	CommandFailed      *bool `json:"commandFailed,omitempty"`
	RegistrationRequest *bool `json:"registrationRequest,omitempty"`
}

// WebhookNotificationInput represents webhook notification settings input.
type WebhookNotificationInput struct {
	Enabled *bool   `json:"enabled,omitempty"`
	URL     *string `json:"url,omitempty"`
	Types   []string `json:"types,omitempty"`
}

// NotificationInput represents notification settings update input from the API.
type NotificationInput struct {
	Enabled  *bool                    `json:"enabled,omitempty"`
	Channels *[]string                `json:"channels,omitempty"`
	Email    *EmailNotificationInput  `json:"email,omitempty"`
	Push     *PushNotificationInput   `json:"push,omitempty"`
	Webhook  *WebhookNotificationInput `json:"webhook,omitempty"`
}
