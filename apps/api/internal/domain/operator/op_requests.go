package operator

// LoginRequest is the payload for operator login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the payload for operator self-registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// OperatorRegisterRequest is the payload for operator self-registration.
// Alias for RegisterRequest for backwards compatibility.
type OperatorRegisterRequest = RegisterRequest

// UpdateNameRequest is the payload for updating the operator's display name.
type UpdateNameRequest struct {
	Name *string `json:"name,omitempty"`
}

// UpdateSettingsRequest is the payload for updating operator settings.
type UpdateSettingsRequest struct {
	Name       *string         `json:"name,omitempty"`
	Thresholds *Thresholds     `json:"thresholds,omitempty"`
	Client     *ClientSettings `json:"client,omitempty"`
	Reset      bool            `json:"reset,omitempty"` // Reset all settings to defaults.
}
