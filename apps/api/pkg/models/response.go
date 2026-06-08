package models

// ErrorResponse is the standard error envelope returned by the API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// OKResponse is the standard success envelope returned by the API.
type OKResponse struct {
	Database   string `json:"database,omitempty"`
	ServerTime int64  `json:"serverTime,omitempty"`
	OK         bool   `json:"ok"`
}
