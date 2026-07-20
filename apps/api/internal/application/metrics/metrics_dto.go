package metrics

// TimeRange constants for metrics queries.
const (
	TimeRange1h  = "1h"
	TimeRange6h  = "6h"
	TimeRange24h = "24h"
	TimeRange7d  = "7d"

	Resolution1m   = "1m"
	Resolution5m   = "5m"
	Resolution15m  = "15m"
	Resolution1h   = "1h"
	ResolutionAuto = "auto"
)

// GetMetricsRequest represents the request for GET /v1/device/:imei/metrics.
type GetMetricsRequest struct {
	DeviceID       string `param:"imei" validate:"required"`
	OrganizationID string // Required for fetching operator thresholds (organization-scoped)
	Range          string `query:"range"`
	Resolution     string `query:"resolution"`
	StartTime      int64  `query:"startTime"`
	EndTime        int64  `query:"endTime"`
}

// GetMetricsResponse represents the response for GET /v1/device/:imei/metrics.
type GetMetricsResponse struct {
	TimeRange TimeRangeResponse    `json:"timeRange"`
	Device    DeviceInfoResponse   `json:"device"`
	Events    []ThresholdEventDTO  `json:"events,omitempty"`
	Metrics   MetricsCollectionDTO `json:"metrics"`
}

// DeviceInfoResponse represents device identification in responses.
type DeviceInfoResponse struct {
	IMEI       string `json:"imei"`
	DeviceName string `json:"deviceName,omitempty"`
}

// TimeRangeResponse represents the time range in responses.
type TimeRangeResponse struct {
	Range      string `json:"range"`
	Resolution string `json:"resolution"`
	Start      int64  `json:"start"`
	End        int64  `json:"end"`
}

// MetricsCollectionDTO contains all metric types.
type MetricsCollectionDTO struct {
	RiskScore   MetricDataDTO `json:"riskScore"`
	ThermalTemp MetricDataDTO `json:"thermalTemp"`
	BufferLevel MetricDataDTO `json:"bufferLevel"`
	Uptime      MetricDataDTO `json:"uptime"`
}

// MetricDataDTO represents a complete metric with stats, chart data, and thresholds.
type MetricDataDTO struct {
	Unit      string           `json:"unit"`
	Chart     []MetricPointDTO `json:"chart"`
	Threshold ThresholdDTO     `json:"threshold"`
	Current   float64          `json:"current"`
	Avg       float64          `json:"avg"`
	Min       float64          `json:"min"`
	Max       float64          `json:"max"`
}

// MetricPointDTO represents a single data point in a chart.
type MetricPointDTO struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// ThresholdDTO represents warning and critical thresholds.
type ThresholdDTO struct {
	Warning  float64 `json:"warning"`
	Critical float64 `json:"critical"`
}

// ThresholdEventDTO represents an event when a threshold is breached.
type ThresholdEventDTO struct {
	Type      string  `json:"type"`
	Metric    string  `json:"metric"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
}

// GetTelemetryRequest represents the request for GET /v1/device/:imei/telemetry.
type GetTelemetryRequest struct {
	DeviceID  string `param:"imei" validate:"required"`
	StartTime int64  `query:"startTime"`
	EndTime   int64  `query:"endTime"`
	Limit     int    `query:"limit"`
}

// GetTelemetryResponse represents the response for GET /v1/device/:imei/telemetry.
type GetTelemetryResponse struct {
	Frames []TelemetryFrameDTO `json:"frames"`
	Stats  TelemetryStatsDTO   `json:"stats"`
}

// TelemetryFrameDTO represents a raw telemetry data point.
type TelemetryFrameDTO struct {
	Timestamp   int64   `json:"timestamp"`
	Uptime      int64   `json:"uptime"`
	RiskScore   float64 `json:"riskScore"`
	ThermalTemp float64 `json:"thermalTemp"`
	BufferLevel float64 `json:"bufferLevel"`
}

// TelemetryStatsDTO represents statistics for telemetry data.
type TelemetryStatsDTO struct {
	RiskScore   MetricStatsDTO `json:"riskScore"`
	ThermalTemp MetricStatsDTO `json:"thermalTemp"`
	BufferLevel MetricStatsDTO `json:"bufferLevel"`
}

// MetricStatsDTO represents statistical summary of a metric.
type MetricStatsDTO struct {
	Current float64 `json:"current"`
	Avg     float64 `json:"avg"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
}

// ExportMetricsRequest represents the request for GET /v1/device/:imei/metrics/export.
type ExportMetricsRequest struct {
	DeviceID string `param:"imei" validate:"required"`
	Format   string `query:"format"`
	Range    string `query:"range"`
	Metrics  string `query:"metrics"`
}

// ExportMetricsResponse represents the exported metrics data.
type ExportMetricsResponse struct {
	Format     string              `json:"format"`
	Filename   string              `json:"filename"`
	DeviceID   string              `json:"deviceId"`
	TimeRange  string              `json:"timeRange"`
	Data       []TelemetryFrameDTO `json:"data"`
	ExportedAt int64               `json:"exportedAt"`
}
