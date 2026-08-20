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

// MetricSource computes org-scoped metric values over the database. It mirrors
// the cross-resource query style of the search service (raw *sql.DB).
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

// Value computes the metric for an org at the given time.
func (s *MetricSource) Value(ctx context.Context, orgID string, metric alert.Metric, now time.Time) (float64, error) {
	switch metric {
	case alert.MetricDeviceOfflineCount:
		return s.countOffline(ctx, orgID)
	case alert.MetricDeviceOfflinePercent:
		return s.offlinePercent(ctx, orgID)
	case alert.MetricCommandFailureRate:
		return s.commandFailureRate(ctx, orgID, now)
	}
	return 0, nil
}

func (s *MetricSource) countOffline(ctx context.Context, orgID string) (float64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM devices WHERE organization_id = ? AND online = 0`, orgID,
	).Scan(&n)
	return float64(n), err
}

func (s *MetricSource) offlinePercent(ctx context.Context, orgID string) (float64, error) {
	var total, offline int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN online = 0 THEN 1 ELSE 0 END), 0)
		FROM devices WHERE organization_id = ?`, orgID,
	).Scan(&total, &offline)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	return float64(offline) * 100 / float64(total), nil
}

func (s *MetricSource) commandFailureRate(ctx context.Context, orgID string, now time.Time) (float64, error) {
	windowStart := now.Add(-time.Duration(s.windowSeconds) * time.Second).UnixMilli()
	var total, failed int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN c.status = 'failed' THEN 1 ELSE 0 END), 0)
		FROM commands c
		JOIN devices d ON c.device_id = d.id
		WHERE d.organization_id = ? AND c.created_at >= ?`, orgID, windowStart,
	).Scan(&total, &failed)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	return float64(failed) * 100 / float64(total), nil
}
