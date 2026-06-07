package models

import "encoding/json"

type CommandRequest struct {
	Command   string          `json:"command"`
	Nonce     string          `json:"nonce"`
	Signature string          `json:"signature,omitempty"`
	Args      json.RawMessage `json:"args,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

type CommandFrame struct {
	Type       string          `json:"type"`
	DispatchID string          `json:"dispatchId"`
	Command    string          `json:"command"`
	Nonce      string          `json:"nonce"`
	Signature  string          `json:"signature,omitempty"`
	Args       json.RawMessage `json:"args,omitempty"`
	Timestamp  int64           `json:"timestamp"`
}

type CommandResponse struct {
	DispatchID string `json:"dispatchId"`
	Delivery   string `json:"delivery"`
	ServerTime int64  `json:"serverTime"`
}
