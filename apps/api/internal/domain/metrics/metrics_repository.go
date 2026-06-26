package metrics

import (
	"context"
	"time"
)

// Repository defines the interface for metrics data access.
type Repository interface {
	// GetTelemetryFrames retrieves raw telemetry frames for a device within a time range.
	GetTelemetryFrames(ctx context.Context, deviceID string, startTime, endTime time.Time, limit int) ([]*TelemetryFrame, error)

	// GetLatestTelemetry retrieves the most recent telemetry frame for a device.
	GetLatestTelemetry(ctx context.Context, deviceID string) (*TelemetryFrame, error)

	// GetAggregatedMetrics retrieves aggregated metrics for charting.
	GetAggregatedMetrics(ctx context.Context, deviceID string, metric string, startTime, endTime time.Time, resolution string) ([]*MetricDataPoint, error)

	// GetMetricStats calculates statistics for a metric over a time range.
	GetMetricStats(ctx context.Context, deviceID string, metric string, startTime, endTime time.Time) (*MetricStats, error)

	// GetThresholdBreachEvents retrieves events where metrics exceeded thresholds.
	GetThresholdBreachEvents(ctx context.Context, deviceID string, startTime, endTime time.Time, thresholds *ThresholdPreset) ([]*MetricThresholdEvent, error)
}
