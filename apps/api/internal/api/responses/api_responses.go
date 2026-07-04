package responses

// MessageResponse is a simple message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// OKResponse indicates a successful operation with no data.
type OKResponse struct {
	OK bool `json:"ok"`
}

// ErrorResponse is the standard error envelope for API endpoints.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// RegisterRequest is the payload for device/client registration.
type RegisterRequest struct {
	DeviceID  string `json:"deviceId"`
	Firmware  string `json:"firmware"`
	Signature string `json:"signature,omitempty"`
}

// RegisterResponse is returned on successful device/client registration.
type RegisterResponse struct {
	ClientID     string `json:"clientId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresAt    int64  `json:"expiresAt"`
}
