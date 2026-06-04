package models

import "encoding/json"

type CommandRequest struct {
	Command   string          `json:"command"`
	Args      json.RawMessage `json:"args,omitempty"`
	Nonce     string          `json:"nonce"`
	Timestamp int64           `json:"timestamp"`
	Signature string          `json:"signature,omitempty"`
}

type CommandFrame struct {
	Type       string          `json:"type"`
	DispatchID string          `json:"dispatchId"`
	Command    string          `json:"command"`
	Args       json.RawMessage `json:"args,omitempty"`
	Nonce      string          `json:"nonce"`
	Timestamp  int64           `json:"timestamp"`
	Signature  string          `json:"signature,omitempty"`
}

type CommandResponse struct {
	DispatchID string `json:"dispatchId"`
	Delivery   string `json:"delivery"`
	ServerTime int64  `json:"serverTime"`
}
