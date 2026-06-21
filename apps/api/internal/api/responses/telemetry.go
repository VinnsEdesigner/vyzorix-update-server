package responses

// TelemetryAck acknowledges receipt of telemetry data.
type TelemetryAck struct {
	Received int  `json:"received"`
	Queued   int  `json:"queued,omitempty"`
}
