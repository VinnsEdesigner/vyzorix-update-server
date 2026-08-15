package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

// to avoid O(N) DELETE on every INSERT and to make retention per-device.
func (r *TelemetryRepository) Save(ctx context.Context, deviceID string, raw []byte, frame telemetry.TelemetryFrame) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer SafeRollbackNoLog(tx)

	// Generate UUIDv7 for the telemetry entry.
	telemetryID := shared.GenerateID()

	now := time.Now().UnixMilli()

	// Use uptime from frame if available, otherwise 0.
	uptime := frame.Uptime
	if uptime == 0 {
		uptime = 0 // Default to 0 if not provided.
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO telemetry(id, device_id, received_at, frame_data, risk_score, buffer_level, thermal_temp, uptime) 
		 VALUES(?,?,?,?,?,?,?,?)`,
		telemetryID, deviceID, now, string(raw), frame.RiskScore, frame.BufferLevel, frame.ThermalTemp, uptime,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// PruneDeviceTelemetry removes old telemetry entries for a specific device,.
// keeping only the latest maxEntries. This is called by a background job.

func (r *TelemetryRepository) PruneDeviceTelemetry(ctx context.Context, deviceID string, maxEntries int) error {
	if maxEntries <= 0 {
		maxEntries = 500 // Default per-device limit.
	}

	_, err := r.exec(ctx,
		`DELETE FROM telemetry WHERE device_id = ? AND id NOT IN (
			SELECT id FROM telemetry WHERE device_id = ? ORDER BY received_at DESC LIMIT ?
		)`,
		deviceID, deviceID, maxEntries,
	)
	return err
}

// PruneAllDevices iterates over all devices and prunes their telemetry.
// This should be called by a background job to avoid blocking INSERTs.
func (r *TelemetryRepository) PruneAllDevices(ctx context.Context, maxEntriesPerDevice int) error {
	if maxEntriesPerDevice <= 0 {
		maxEntriesPerDevice = 500
	}

	// Get all unique device IDs.
	rows, err := r.queryRows(ctx, `SELECT DISTINCT device_id FROM telemetry`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			continue
		}
		// Prune each device (ignore errors for individual devices).
		_ = r.PruneDeviceTelemetry(ctx, deviceID, maxEntriesPerDevice)
	}

	return rows.Err()
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
		`SELECT id, device_id, received_at, frame_data, risk_score, buffer_level, thermal_temp, COALESCE(uptime, 0) 
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
		`SELECT id, device_id, received_at, frame_data, risk_score, buffer_level, thermal_temp, COALESCE(uptime, 0) 
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

// DeleteByDeviceIDs deletes all telemetry entries for the given device IDs.
// This is used during organization deletion.
func (r *TelemetryRepository) DeleteByDeviceIDs(ctx context.Context, deviceIDs []string) (int64, error) {
	if len(deviceIDs) == 0 {
		return 0, nil
	}

	// Build query with placeholders for each device ID.
	placeholders := make([]string, len(deviceIDs))
	args := make([]interface{}, len(deviceIDs))
	for i, id := range deviceIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`DELETE FROM telemetry WHERE device_id IN (%s)`, strings.Join(placeholders, ","))

	result, err := r.exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
