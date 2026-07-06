package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Ensure TelemetryRepository implements telemetry.Repository.
var _ telemetry.Repository = (*TelemetryRepository)(nil)

// TelemetryRepository implements telemetry.Repository using SQLite.
type TelemetryRepository struct {
	db *sql.DB
}

// NewTelemetryRepository creates a new TelemetryRepository.
func NewTelemetryRepository(db *sql.DB) *TelemetryRepository {
	return &TelemetryRepository{db: db}
}
// getQuerier returns the transaction from context if available, otherwise the db.
func (r *TelemetryRepository) getQuerier(ctx context.Context) Querier {
if tx, ok := transaction.TxFromContext(ctx); ok {
return tx
}
return r.db
}

// queryRow is a helper that uses transaction-aware querier.
func (r *TelemetryRepository) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
return r.getQuerier(ctx).QueryRowContext(ctx, query, args...)
}

// queryRows is a helper that uses transaction-aware querier.
func (r *TelemetryRepository) queryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
return r.getQuerier(ctx).QueryContext(ctx, query, args...)
}

// exec is a helper that uses transaction-aware querier.
func (r *TelemetryRepository) exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
return r.getQuerier(ctx).ExecContext(ctx, query, args...)
}

// Save saves a telemetry frame for a device.
// Implements auto-pruning to keep only the latest 5000 entries.
func (r *TelemetryRepository) Save(ctx context.Context, deviceID string, raw []byte, frame telemetry.TelemetryFrame) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Generate UUIDv7 for the telemetry entry
	telemetryID := shared.GenerateID()

	now := time.Now().UnixMilli()

	// Use uptime from frame if available, otherwise 0
	uptime := frame.Uptime
	if uptime == 0 {
		uptime = 0 // Default to 0 if not provided
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO telemetry(id, device_id, received_at, frame_data, risk_score, buffer_level, thermal_temp, uptime) 
		 VALUES(?,?,?,?,?,?,?,?)`,
		telemetryID, deviceID, now, string(raw), frame.RiskScore, frame.BufferLevel, frame.ThermalTemp, uptime,
	)
	if err != nil {
		return err
	}

	// Prune old telemetry entries (keep latest 5000)
	_, err = tx.ExecContext(ctx,
		`DELETE FROM telemetry WHERE id NOT IN (SELECT id FROM telemetry ORDER BY received_at DESC LIMIT 5000)`,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// List retrieves telemetry entries for a device with pagination.
func (r *TelemetryRepository) List(ctx context.Context, deviceID string, limit int) ([]telemetry.TelemetryEntry, error) {
	if limit <= 0 {
		limit = 100
	}

	if limit > 5000 {
		limit = 5000
	}

	rows, err := r.queryRows(ctx,
		`SELECT id, device_id, received_at, payload, risk_score, buffer_level, thermal_temp, COALESCE(uptime, 0) 
		 FROM telemetry WHERE device_id = ? ORDER BY received_at DESC LIMIT ?`,
		deviceID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries []telemetry.TelemetryEntry

	for rows.Next() {
		var e telemetry.TelemetryEntry

		var receivedAt int64
		if err := rows.Scan(&e.ID, &e.DeviceID, &receivedAt, &e.Payload, &e.RiskScore, &e.BufferLevel, &e.ThermalTemp, &e.Uptime); err != nil {
			return nil, err
		}

		e.ReceivedAt = time.UnixMilli(receivedAt).UTC()
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// ListSince retrieves telemetry entries since a given timestamp.
func (r *TelemetryRepository) ListSince(ctx context.Context, deviceID string, sinceTimestamp int64, limit int) ([]telemetry.TelemetryEntry, error) {
	if limit <= 0 {
		limit = 100
	}

	if limit > 5000 {
		limit = 5000
	}

	rows, err := r.queryRows(ctx,
		`SELECT id, device_id, received_at, payload, risk_score, buffer_level, thermal_temp, COALESCE(uptime, 0) 
		 FROM telemetry WHERE device_id = ? AND received_at > ? ORDER BY received_at DESC LIMIT ?`,
		deviceID, sinceTimestamp, limit,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries []telemetry.TelemetryEntry

	for rows.Next() {
		var e telemetry.TelemetryEntry

		var receivedAt int64
		if err := rows.Scan(&e.ID, &e.DeviceID, &receivedAt, &e.Payload, &e.RiskScore, &e.BufferLevel, &e.ThermalTemp, &e.Uptime); err != nil {
			return nil, err
		}

		e.ReceivedAt = time.UnixMilli(receivedAt).UTC()
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// Count returns the number of telemetry entries for a device.
func (r *TelemetryRepository) Count(ctx context.Context, deviceID string) (int, error) {
	var count int
	err := r.queryRow(ctx,
		`SELECT COUNT(*) FROM telemetry WHERE device_id = ?`,
		deviceID,
	).Scan(&count)

	return count, err
}

// DeleteOlderThan removes telemetry entries older than the given timestamp.
func (r *TelemetryRepository) DeleteOlderThan(ctx context.Context, olderThanTimestamp int64) (int64, error) {
	result, err := r.exec(ctx,
		`DELETE FROM telemetry WHERE received_at < ?`,
		olderThanTimestamp,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
