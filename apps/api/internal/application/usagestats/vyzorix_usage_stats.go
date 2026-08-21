// Package usagestats computes periodic self-telemetry: build version, feature
// toggle state, and entity counts. The update-checker surfaces it to operators.
package usagestats

import (
	"context"
	"database/sql"
	"time"
)

// EntityCounts reports per-table counts.
type EntityCounts struct {
	Devices         int `json:"devices"`
	Operators       int `json:"operators"`
	Organizations   int `json:"organizations"`
	ServiceAccounts int `json:"service_accounts"`
	AlertRules      int `json:"alert_rules"`
	ContactPoints   int `json:"contact_points"`
	Annotations     int `json:"annotations"`
}

// Snapshot is one self-telemetry point.
type Snapshot struct {
	CollectedAt time.Time       `json:"collected_at"`
	Toggles     map[string]bool `json:"toggles"`
	Counts      EntityCounts    `json:"counts"`
}

// Collector queries the DB for entity counts and feature toggles.
type Collector struct {
	db *sql.DB
}

// NewCollector creates a Collector.
func NewCollector(db *sql.DB) *Collector {
	return &Collector{db: db}
}

// Query returns the current entity + toggle snapshot.
func (c *Collector) Query(ctx context.Context) *Snapshot {
	return &Snapshot{
		CollectedAt: time.Now(),
		Counts:      c.queryCounts(ctx),
		Toggles:     map[string]bool{},
	}
}

func (c *Collector) queryCounts(ctx context.Context) EntityCounts {
	return EntityCounts{
		Devices:         c.count(ctx, "devices"),
		Operators:       c.count(ctx, "operators"),
		Organizations:   c.count(ctx, "organizations"),
		ServiceAccounts: c.count(ctx, "service_accounts"),
		AlertRules:      c.count(ctx, "alert_rules"),
		ContactPoints:   c.count(ctx, "contact_points"),
		Annotations:     c.count(ctx, "annotations"),
	}
}

func (c *Collector) count(ctx context.Context, table string) int {
	var n int
	_ = c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&n)
	return n
}
