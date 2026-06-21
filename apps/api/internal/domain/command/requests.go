package command

import "encoding/json"

// CommandRequest is the payload for sending a command to a device.
type CommandRequest struct {
	Command   string          `json:"command"`
	Nonce     string          `json:"nonce"`
	Signature string          `json:"signature,omitempty"`
	Args      json.RawMessage `json:"args,omitempty"`
	Timestamp int64           `json:"timestamp"`
}
