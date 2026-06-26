package metrics

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a metric is not found.
var ErrNotFound = errors.New("metric not found")

// TelemetryFrame represents a raw telemetry data point from a device.
type TelemetryFrame struct {
	Timestamp   time.Time `json:"timestamp"`
	DeviceID    string    `json:"deviceId"`
	RiskScore   float64   `json:"riskScore"`
	ThermalTemp float64   `json:"thermalTemp"`
	BufferLevel float64   `json:"bufferLevel"`
	Uptime      int64     `json:"uptime"`
}

// MetricDataPoint represents an aggregated metric data point for charts.
type MetricDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// MetricStats represents statistical summary of a metric.
type MetricStats struct {
	Current float64 `json:"current"`
	Avg     float64 `json:"avg"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
}

// MetricThreshold represents warning and critical thresholds.
type MetricThreshold struct {
	Warning  float64 `json:"warning"`
	Critical float64 `json:"critical"`
}

// MetricData represents a complete metric with stats, chart data, and thresholds.
type MetricData struct {
	Unit      string            `json:"unit"`
	Chart     []MetricDataPoint `json:"chart"`
	Stats     MetricStats       `json:"stats"`
	Threshold MetricThreshold   `json:"threshold"`
}

// DeviceMetrics represents all metrics for a device.
type DeviceMetrics struct {
	TimeRange TimeRangeInfo          `json:"timeRange"`
	Device    DeviceInfo             `json:"device"`
	Events    []MetricThresholdEvent `json:"events,omitempty"`
	Metrics   MetricsCollection      `json:"metrics"`
}

// DeviceInfo contains device identification.
type DeviceInfo struct {
	IMEI       string `json:"imei"`
	DeviceName string `json:"deviceName,omitempty"`
}

// TimeRangeInfo contains the time range parameters.
type TimeRangeInfo struct {
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	Range      string    `json:"range"`
	Resolution string    `json:"resolution"`
}

// MetricsCollection contains all metric types.
type MetricsCollection struct {
	RiskScore   MetricData `json:"riskScore"`
	ThermalTemp MetricData `json:"thermalTemp"`
	BufferLevel MetricData `json:"bufferLevel"`
	Uptime      MetricData `json:"uptime"`
}

// MetricThresholdEvent represents an event when a threshold is breached.
type MetricThresholdEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
}

// ThresholdPreset represents operator-specific threshold settings.
type ThresholdPreset struct {
	RiskScoreWarning  float64 `json:"riskScoreWarning"`
	RiskScoreCritical float64 `json:"riskScoreCritical"`
	ThermalWarning    float64 `json:"thermalWarning"`
	ThermalCritical   float64 `json:"thermalCritical"`
	BufferWarning     float64 `json:"bufferWarning"`
	BufferCritical    float64 `json:"bufferCritical"`
}
