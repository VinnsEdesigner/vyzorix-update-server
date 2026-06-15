// Package storage provides SQLite database operations.
package storage

import (
	"context"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

// SaveTelemetry saves a telemetry frame for a device.
// Uses UUIDv7 for the telemetry ID.
func (s *Store) SaveTelemetry(ctx context.Context, deviceID string, raw []byte, t models.TelemetryFrame) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	// Generate UUIDv7 for the telemetry entry
	telemetryID := NewUUIDv7()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO telemetry(id, device_id, received_at, payload, risk_score, buffer_level, thermal_temp) VALUES(?,?,?,?,?,?,?)`,
		telemetryID, deviceID, time.Now().UnixMilli(), string(raw), t.RiskScore, t.BufferLevel, t.ThermalTemp,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE devices SET last_seen=? WHERE id=?`,
		time.Now().UnixMilli(), deviceID,
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

// TelemetryEntry represents a telemetry record.
type TelemetryEntry struct {
	ID           string
	DeviceID     string
	ReceivedAt   time.Time
	Payload      string
	RiskScore    int
	BufferLevel  int
	ThermalTemp  float64
}

// GetTelemetry retrieves telemetry entries for a device.
func (s *Store) GetTelemetry(ctx context.Context, deviceID string, limit int) ([]TelemetryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, device_id, received_at, payload, risk_score, buffer_level, thermal_temp 
		 FROM telemetry WHERE device_id = ? ORDER BY received_at DESC LIMIT ?`,
		deviceID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var entries []TelemetryEntry
	for rows.Next() {
		var e TelemetryEntry
		var receivedAt int64
		if err := rows.Scan(&e.ID, &e.DeviceID, &receivedAt, &e.Payload, &e.RiskScore, &e.BufferLevel, &e.ThermalTemp); err != nil {
			return nil, err
		}
		e.ReceivedAt = time.UnixMilli(receivedAt).UTC()
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetTelemetrySince retrieves telemetry entries since a given timestamp.
func (s *Store) GetTelemetrySince(ctx context.Context, deviceID string, since int64, limit int) ([]TelemetryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, device_id, received_at, payload, risk_score, buffer_level, thermal_temp 
		 FROM telemetry WHERE device_id = ? AND received_at > ? ORDER BY received_at DESC LIMIT ?`,
		deviceID, since, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var entries []TelemetryEntry
	for rows.Next() {
		var e TelemetryEntry
		var receivedAt int64
		if err := rows.Scan(&e.ID, &e.DeviceID, &receivedAt, &e.Payload, &e.RiskScore, &e.BufferLevel, &e.ThermalTemp); err != nil {
			return nil, err
		}
		e.ReceivedAt = time.UnixMilli(receivedAt).UTC()
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// CountTelemetry returns the number of telemetry entries for a device.
func (s *Store) CountTelemetry(ctx context.Context, deviceID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM telemetry WHERE device_id = ?`,
		deviceID,
	).Scan(&count)
	return count, err
}

// DeleteOldTelemetry removes telemetry entries older than the given timestamp.
func (s *Store) DeleteOldTelemetry(ctx context.Context, olderThan int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.db.ExecContext(ctx,
		`DELETE FROM telemetry WHERE received_at < ?`,
		olderThan,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}