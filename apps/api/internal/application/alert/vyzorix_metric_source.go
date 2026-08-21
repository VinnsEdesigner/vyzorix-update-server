// Package alert provides the alerting application services: rule CRUD and the
// evaluator that advances rule instances against fleet metrics.
package alert

import (
	"context"
	"database/sql"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/alert"
)

// DefaultWindowSeconds is the observation window for windowed metrics
// (command_failure_rate); offline metrics are point-in-time.
const DefaultWindowSeconds = 3600

// LabelSeries is one labeled observation of a metric. NoData marks a series
// whose window carried no signal at all (e.g. zero commands); an empty label
// set is the fleet-wide aggregate.
type LabelSeries struct {
	Labels map[string]string
	Value  float64
	NoData bool
}

// MetricSource computes org-scoped metric series over the database. It
// mirrors the cross-resource query style of the search service (raw *sql.DB).
//
// Offline metrics fan out one series per device_class plus the unlabeled
// fleet aggregate. command_failure_rate reports the aggregate only and flags
// NoData when the window held no commands.
type MetricSource struct {
	db            *sql.DB
	windowSeconds int64
}

// NewMetricSource creates a metric source; windowSeconds <= 0 uses the default.
func NewMetricSource(db *sql.DB, windowSeconds int64) *MetricSource {
	if windowSeconds <= 0 {
		windowSeconds = DefaultWindowSeconds
	}
	return &MetricSource{db: db, windowSeconds: windowSeconds}
}

// Series computes the labeled series for a metric at the given time.
func (s *MetricSource) Series(ctx context.Context, orgID string, metric alert.Metric, now time.Time) ([]*LabelSeries, error) {
	switch metric {
	case alert.MetricDeviceOfflineCount:
		return s.countOfflineByClass(ctx, orgID, false)
	case alert.MetricDeviceOfflinePercent:
		return s.countOfflineByClass(ctx, orgID, true)
	case alert.MetricCommandFailureRate:
		return s.commandFailureRate(ctx, orgID, now)
	}
	return nil, nil
}

func (s *MetricSource) countOfflineByClass(ctx context.Context, orgID string, percent bool) ([]*LabelSeries, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT COALESCE(NULLIF(device_class, ''), 'unclassified'),
			COUNT(*),
			COALESCE(SUM(CASE WHEN online = 0 THEN 1 ELSE 0 END), 0)
		FROM devices
		WHERE organization_id = ?
		GROUP BY device_class
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var series []*LabelSeries
	var totalAll, offlineAll float64
	for rows.Next() {
		var class string
		var total, offline int64
		if err := rows.Scan(&class, &total, &offline); err != nil {
			return nil, err
		}
		value := float64(offline)
		if percent && total > 0 {
			value = value * 100 / float64(total)
		}
		series = append(series, &LabelSeries{
			Labels: map[string]string{"device_class": class},
			Value:  value,
		})
		totalAll += float64(total)
		offlineAll += float64(offline)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// The unlabeled aggregate complements the class breakdown so one rule can
	// still watch the fleet as a single series. Percent metrics on an empty
	// fleet mark no-data instead of silently zeroing.
	aggregate := &LabelSeries{Labels: nil, Value: offlineAll}
	if percent {
		if totalAll > 0 {
			aggregate.Value = offlineAll * 100 / totalAll
		} else {
			aggregate.NoData = true
		}
	}
	return append(series, aggregate), nil
}

func (s *MetricSource) commandFailureRate(ctx context.Context, orgID string, now time.Time) ([]*LabelSeries, error) {
	windowStart := now.Add(-time.Duration(s.windowSeconds) * time.Second).UnixMilli()
	var total, failed int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN c.status = 'failed' THEN 1 ELSE 0 END), 0)
		FROM commands c
		JOIN devices d ON c.device_id = d.id
		WHERE d.organization_id = ? AND c.created_at >= ?`, orgID, windowStart,
	).Scan(&total, &failed)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return []*LabelSeries{{Labels: nil, Value: 0, NoData: true}}, nil
	}
	return []*LabelSeries{{Labels: nil, Value: float64(failed) * 100 / float64(total)}}, nil
}
