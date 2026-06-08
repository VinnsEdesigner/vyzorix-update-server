package models

import "encoding/json"

// CommandRequest is the payload for sending a command to a device.
type CommandRequest struct {
	Command   string          `json:"command"`
	Nonce     string          `json:"nonce"`
	Signature string          `json:"signature,omitempty"`
	Args      json.RawMessage `json:"args,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

// CommandFrame is the internal representation of a command for the WebSocket hub.
type CommandFrame struct {
	Type       string          `json:"type"`
	DispatchID string          `json:"dispatchId"`
	Command    string          `json:"command"`
	Nonce      string          `json:"nonce"`
	Signature  string          `json:"signature,omitempty"`
	Args       json.RawMessage `json:"args,omitempty"`
	Timestamp  int64           `json:"timestamp"`
}

// CommandResponse is the server's response to a command dispatch.
type CommandResponse struct {
	DispatchID string `json:"dispatchId"`
	Delivery   string `json:"delivery"`
	ServerTime int64  `json:"serverTime"`
}
