package models

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type OKResponse struct {
	Database   string `json:"database,omitempty"`
	ServerTime int64  `json:"serverTime,omitempty"`
	OK         bool   `json:"ok"`
}
