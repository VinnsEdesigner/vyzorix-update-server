package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Ensure MetricsRepository implements metrics.Repository.
var _ metrics.Repository = (*MetricsRepository)(nil)

// MetricsRepository implements metrics.Repository using SQLite.
type MetricsRepository struct {
	db *sql.DB
}

// NewMetricsRepository creates a new MetricsRepository.
func NewMetricsRepository(db *sql.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}
// getQuerier returns the transaction from context if available, otherwise the db.
func (r *MetricsRepository) getQuerier(ctx context.Context) Querier {
if tx, ok := transaction.TxFromContext(ctx); ok {
return tx
}
return r.db
}

// queryRow is a helper that uses transaction-aware querier.
func (r *MetricsRepository) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
return r.getQuerier(ctx).QueryRowContext(ctx, query, args...)
}

// queryRows is a helper that uses transaction-aware querier.
func (r *MetricsRepository) queryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
return r.getQuerier(ctx).QueryContext(ctx, query, args...)
}


// GetTelemetryFrames retrieves raw telemetry frames for a device within a time range.
func (r *MetricsRepository) GetTelemetryFrames(ctx context.Context, deviceID string, startTime, endTime time.Time, limit int) ([]*metrics.TelemetryFrame, error) {
	query := `
		SELECT device_id, risk_score, thermal_temp, buffer_level, received_at, COALESCE(uptime, 0)
		FROM telemetry
		WHERE device_id = ? AND received_at >= ? AND received_at <= ?
		ORDER BY received_at DESC
		LIMIT ?`

	rows, err := r.queryRows(ctx, query, deviceID, startTime.UnixMilli(), endTime.UnixMilli(), limit)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var frames []*metrics.TelemetryFrame

	for rows.Next() {
		frame := metrics.TelemetryFrame{}
		var timestamp int64

		err := rows.Scan(
			&frame.DeviceID, &frame.RiskScore, &frame.ThermalTemp,
			&frame.BufferLevel, &timestamp, &frame.Uptime,
		)
		if err != nil {
			return nil, err
		}

		frame.Timestamp = time.UnixMilli(timestamp)
		frames = append(frames, &frame)
	}

	return frames, rows.Err()
}

// // GetLatestTelemetry retrieves the most recent telemetry frame for a device.
func (r *MetricsRepository) GetLatestTelemetry(ctx context.Context, deviceID string) (*metrics.TelemetryFrame, error) {
	query := `
		SELECT device_id, risk_score, thermal_temp, buffer_level, received_at, COALESCE(uptime, 0)
		FROM telemetry
		WHERE device_id = ?
		ORDER BY received_at DESC
		LIMIT 1`

	var frame metrics.TelemetryFrame
	var timestamp int64

	err := r.queryRow(ctx, query, deviceID).Scan(
		&frame.DeviceID, &frame.RiskScore, &frame.ThermalTemp,
		&frame.BufferLevel, &timestamp, &frame.Uptime,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, metrics.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	frame.Timestamp = time.UnixMilli(timestamp)
	return &frame, nil
}

// GetAggregatedMetrics retrieves aggregated metrics for charting with specified resolution.

func (r *MetricsRepository) GetAggregatedMetrics(ctx context.Context, deviceID string, metric string, startTime, endTime time.Time, resolution string) ([]*metrics.MetricDataPoint, error) {
	// Map metric name to column
	column := metricToColumn(metric)
	if column == "" {
		return nil, nil
	}

	// SQLite uses strftime for time bucketing - group by minute buckets
	format := resolutionToStrftime(resolution)

	
	maxTimeWindow := 30 * 24 * time.Hour
	if endTime.Sub(startTime) > maxTimeWindow {
		startTime = endTime.Add(-maxTimeWindow)
	}
	maxRows := 1000

	query := `
		SELECT 
			strftime('` + format + `', received_at / 1000, 'unixepoch') as bucket,
			AVG(` + column + `) as value,
			MIN(received_at) as min_ts
		FROM telemetry
		WHERE device_id = ? AND received_at >= ? AND received_at <= ?
		GROUP BY bucket
		ORDER BY bucket ASC
		LIMIT ?`

	rows, err := r.queryRows(ctx, query, deviceID, startTime.UnixMilli(), endTime.UnixMilli(), maxRows)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var points []*metrics.MetricDataPoint

	for rows.Next() {
		var point metrics.MetricDataPoint
		var bucketStr string
		var minTs int64

		err := rows.Scan(&bucketStr, &point.Value, &minTs)
		if err != nil {
			return nil, err
		}

		point.Timestamp = time.UnixMilli(minTs)
		points = append(points, &point)
	}

	return points, rows.Err()
}

// GetMetricStats calculates statistics for a metric over a time range.
func (r *MetricsRepository) GetMetricStats(ctx context.Context, deviceID string, metric string, startTime, endTime time.Time) (*metrics.MetricStats, error) {
	column := metricToColumn(metric)
	if column == "" {
		return nil, metrics.ErrNotFound
	}

	query := `
		SELECT 
			AVG(` + column + `) as avg,
			MIN(` + column + `) as min,
			MAX(` + column + `) as max
		FROM telemetry
		WHERE device_id = ? AND received_at >= ? AND received_at <= ?`

	var stats metrics.MetricStats

	err := r.queryRow(ctx, query, deviceID, startTime.UnixMilli(), endTime.UnixMilli()).Scan(
		&stats.Avg, &stats.Min, &stats.Max,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return &metrics.MetricStats{}, nil
	}

	if err != nil {
		return nil, err
	}

	// Get the current (latest) value
	var current float64
	currentQuery := `SELECT ` + column + ` FROM telemetry WHERE device_id = ? AND received_at >= ? AND received_at <= ? ORDER BY received_at DESC LIMIT 1`
	err = r.queryRow(ctx, currentQuery, deviceID, startTime.UnixMilli(), endTime.UnixMilli()).Scan(&current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	stats.Current = current

	return &stats, nil
}

// GetThresholdBreachEvents retrieves events where metrics exceeded thresholds.
func (r *MetricsRepository) GetThresholdBreachEvents(ctx context.Context, deviceID string, startTime, endTime time.Time, thresholds *metrics.ThresholdPreset) ([]*metrics.MetricThresholdEvent, error) {
	var events []*metrics.MetricThresholdEvent

	// Query for risk score threshold breaches (high = bad)
	riskQuery := `
		SELECT received_at, 'riskScore' as metric, risk_score, ?
		FROM telemetry
		WHERE device_id = ? AND received_at >= ? AND received_at <= ?
		AND risk_score >= ?
		ORDER BY received_at ASC`

	riskRows, scanErr := r.queryRows(ctx, riskQuery,
		thresholds.RiskScoreWarning,
		deviceID, startTime.UnixMilli(), endTime.UnixMilli(),
		thresholds.RiskScoreWarning,
	)
	if scanErr != nil {
		return nil, scanErr
	}
	defer func() { _ = riskRows.Close() }()

	for riskRows.Next() {
		var event metrics.MetricThresholdEvent
		var timestamp int64

		if scanErr = riskRows.Scan(&timestamp, &event.Metric, &event.Value, &event.Threshold); scanErr != nil {
			return nil, scanErr
		}

		event.Timestamp = time.UnixMilli(timestamp)
		event.Type = "threshold_breach"
		events = append(events, &event)
	}
	if scanErr = riskRows.Err(); scanErr != nil {
		return nil, scanErr
	}

	// Query for thermal temp threshold breaches (high = bad)
	thermalQuery := `
		SELECT received_at, 'thermalTemp' as metric, thermal_temp, ?
		FROM telemetry
		WHERE device_id = ? AND received_at >= ? AND received_at <= ?
		AND thermal_temp >= ?
		ORDER BY received_at ASC`

	thermalRows, scanErr := r.queryRows(ctx, thermalQuery,
		thresholds.ThermalWarning,
		deviceID, startTime.UnixMilli(), endTime.UnixMilli(),
		thresholds.ThermalWarning,
	)
	if scanErr != nil {
		return nil, scanErr
	}
	defer func() { _ = thermalRows.Close() }()

	for thermalRows.Next() {
		var event metrics.MetricThresholdEvent
		var timestamp int64

		if scanErr = thermalRows.Scan(&timestamp, &event.Metric, &event.Value, &event.Threshold); scanErr != nil {
			return nil, scanErr
		}

		event.Timestamp = time.UnixMilli(timestamp)
		event.Type = "threshold_breach"
		events = append(events, &event)
	}
	if scanErr = thermalRows.Err(); scanErr != nil {
		return nil, scanErr
	}

	// Query for buffer level threshold breaches (low = bad for buffer, high = bad for risk)
	// Buffer level: warning when <= 30%, critical when <= 10%
	bufferQuery := `
		SELECT received_at, 'bufferLevel' as metric, buffer_level, ?
		FROM telemetry
		WHERE device_id = ? AND received_at >= ? AND received_at <= ?
		AND buffer_level <= ?
		ORDER BY received_at ASC`

	bufferRows, scanErr := r.queryRows(ctx, bufferQuery,
		thresholds.BufferWarning,
		deviceID, startTime.UnixMilli(), endTime.UnixMilli(),
		thresholds.BufferWarning,
	)
	if scanErr != nil {
		return nil, scanErr
	}
	defer func() { _ = bufferRows.Close() }()

	for bufferRows.Next() {
		var event metrics.MetricThresholdEvent
		var timestamp int64

		if scanErr = bufferRows.Scan(&timestamp, &event.Metric, &event.Value, &event.Threshold); scanErr != nil {
			return nil, scanErr
		}

		event.Timestamp = time.UnixMilli(timestamp)
		event.Type = "threshold_breach"
		events = append(events, &event)
	}
	if scanErr = bufferRows.Err(); scanErr != nil {
		return nil, scanErr
	}

	return events, nil
}

// metricToColumn maps metric name to database column.
func metricToColumn(metric string) string {
	switch metric {
	case "riskScore":
		return "risk_score"
	case "thermalTemp":
		return "thermal_temp"
	case "bufferLevel":
		return "buffer_level"
	case "uptime":
		return "uptime"
	default:
		return ""
	}
}

// resolutionToStrftime converts resolution to SQLite strftime format.
func resolutionToStrftime(resolution string) string {
	switch resolution {
	case "1m":
		return "%Y-%m-%d %H:%M:00"
	case "5m":
		return "%Y-%m-%d %H:%M:00"
	case "15m":
		return "%Y-%m-%d %H:%M:00"
	case "1h":
		return "%Y-%m-%d %H:00:00"
	default:
		return "%Y-%m-%d %H:%M:00"
	}
}
