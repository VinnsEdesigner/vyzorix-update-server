// Package storage provides SQLite storage implementations.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/event"
)

// Ensure EventRepository implements event.Repository.
var _ event.Repository = (*EventRepository)(nil)

// EventRepository implements event.Repository using SQLite.
type EventRepository struct {
	db *sql.DB
}

// NewEventRepository creates a new EventRepository.
func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Store saves an event to the repository.
func (r *EventRepository) Store(ctx context.Context, evt *event.Event) error {
	dataBytes, err := json.Marshal(evt.Data)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO device_events (id, device_id, event_type, timestamp, data, severity, source, operator_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		evt.ID,
		evt.DeviceID,
		string(evt.Type),
		evt.Timestamp,
		dataBytes,
		string(evt.Severity),
		evt.Source,
		evt.OperatorID,
	)
	return err
}

// StoreBatch saves multiple events in a single transaction.
func (r *EventRepository) StoreBatch(ctx context.Context, events []*event.Event) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO device_events (id, device_id, event_type, timestamp, data, severity, source, operator_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, evt := range events {
		dataBytes, err := json.Marshal(evt.Data)
		if err != nil {
			return err
		}

		_, err = stmt.ExecContext(ctx,
			evt.ID,
			evt.DeviceID,
			string(evt.Type),
			evt.Timestamp,
			dataBytes,
			string(evt.Severity),
			evt.Source,
			evt.OperatorID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID retrieves an event by its ID.
func (r *EventRepository) GetByID(ctx context.Context, id string) (*event.Event, error) {
	query := `
		SELECT id, device_id, event_type, timestamp, data, severity, source, operator_id
		FROM device_events
		WHERE id = ?`

	var evt event.Event
	var dataBytes []byte
	var timestamp time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&evt.ID,
		&evt.DeviceID,
		&evt.Type,
		&timestamp,
		&dataBytes,
		&evt.Severity,
		&evt.Source,
		&evt.OperatorID,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	evt.Timestamp = timestamp

	if len(dataBytes) > 0 {
		if err := json.Unmarshal(dataBytes, &evt.Data); err != nil {
			evt.Data = nil
		}
	}

	return &evt, nil
}

// GetByDevice retrieves events for a specific device.
func (r *EventRepository) GetByDevice(ctx context.Context, deviceID string, filter *event.EventFilter) (*event.EventResult, error) {
	if filter == nil {
		filter = &event.EventFilter{Limit: 50}
	}

	query := `
		SELECT id, device_id, event_type, timestamp, data, severity, source, operator_id
		FROM device_events
		WHERE device_id = ?`
	args := make([]interface{}, 0, 1); args = append(args, deviceID)

	r.appendFilterConditions(query, filter, &args)

	query += " ORDER BY timestamp DESC"
	query += " LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	events, err := r.scanEvents(rows)
	if err != nil {
		return nil, err
	}

	hasMore := len(events) > filter.Limit
	if hasMore {
		events = events[:filter.Limit]
	}

	return &event.EventResult{
		Events:  events,
		HasMore: hasMore,
	}, nil
}

// GetByDevices retrieves events for multiple devices.
func (r *EventRepository) GetByDevices(ctx context.Context, deviceIDs []string, filter *event.EventFilter) (*event.EventResult, error) {
	if filter == nil {
		filter = &event.EventFilter{Limit: 50}
	}
	if len(deviceIDs) == 0 {
		return &event.EventResult{Events: []event.Event{}}, nil
	}

	placeholders := make([]string, len(deviceIDs))
	args := make([]interface{}, len(deviceIDs))
	for i, id := range deviceIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT id, device_id, event_type, timestamp, data, severity, source, operator_id
		FROM device_events
		WHERE device_id IN (` + strings.Join(placeholders, ",") + ")"

	r.appendFilterConditions(query, filter, &args)

	query += " ORDER BY timestamp DESC"
	query += " LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	events, err := r.scanEvents(rows)
	if err != nil {
		return nil, err
	}

	hasMore := len(events) > filter.Limit
	if hasMore {
		events = events[:filter.Limit]
	}

	return &event.EventResult{
		Events:  events,
		HasMore: hasMore,
	}, nil
}

// GetByType retrieves events of a specific type.
func (r *EventRepository) GetByType(ctx context.Context, eventType event.EventType, filter *event.EventFilter) (*event.EventResult, error) {
	if filter == nil {
		filter = &event.EventFilter{Limit: 50}
	}

	query := `
		SELECT id, device_id, event_type, timestamp, data, severity, source, operator_id
		FROM device_events
		WHERE event_type = ?`
	args := make([]interface{}, 0, 1); args = append(args, string(eventType))

	r.appendFilterConditions(query, filter, &args)

	query += " ORDER BY timestamp DESC"
	query += " LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	events, err := r.scanEvents(rows)
	if err != nil {
		return nil, err
	}

	hasMore := len(events) > filter.Limit
	if hasMore {
		events = events[:filter.Limit]
	}

	return &event.EventResult{
		Events:  events,
		HasMore: hasMore,
	}, nil
}

// GetByOperator retrieves events for devices owned by an operator.
func (r *EventRepository) GetByOperator(ctx context.Context, operatorID string, filter *event.EventFilter) (*event.EventResult, error) {
	if filter == nil {
		filter = &event.EventFilter{Limit: 50}
	}

	// First get device IDs for this operator
	devicesQuery := `SELECT id FROM devices WHERE operator_id = ?`
	rows, err := r.db.QueryContext(ctx, devicesQuery, operatorID)
	if err != nil {
		return nil, err
	}

	var deviceIDs []string
	for rows.Next() {
		var id string
		if scanErr := rows.Scan(&id); scanErr != nil {
			_ = rows.Close()
			return nil, scanErr
		}
		deviceIDs = append(deviceIDs, id)
	}
	// Check rows.Err() after iteration
	if rowsErr := rows.Err(); rowsErr != nil {
		_ = rows.Close()
		return nil, rowsErr
	}
	_ = rows.Close()

	if len(deviceIDs) == 0 {
		return &event.EventResult{Events: []event.Event{}}, nil
	}

	return r.GetByDevices(ctx, deviceIDs, filter)
}

// GetRecent retrieves the most recent events.
func (r *EventRepository) GetRecent(ctx context.Context, limit int) ([]event.Event, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, device_id, event_type, timestamp, data, severity, source, operator_id
		FROM device_events
		ORDER BY timestamp DESC
		LIMIT ?`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEvents(rows)
}

// GetRecentByDevice retrieves recent events for a specific device.
func (r *EventRepository) GetRecentByDevice(ctx context.Context, deviceID string, limit int) ([]event.Event, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, device_id, event_type, timestamp, data, severity, source, operator_id
		FROM device_events
		WHERE device_id = ?
		ORDER BY timestamp DESC
		LIMIT ?`

	rows, err := r.db.QueryContext(ctx, query, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanEvents(rows)
}

// CountByType counts events by type within a time range.
func (r *EventRepository) CountByType(ctx context.Context, eventType event.EventType, startTime, endTime time.Time) (int, error) {
	query := `
		SELECT COUNT(*) FROM device_events
		WHERE event_type = ? AND timestamp >= ? AND timestamp <= ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, string(eventType), startTime, endTime).Scan(&count)
	return count, err
}

// DeleteOld removes events older than the specified time.
func (r *EventRepository) DeleteOld(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM device_events WHERE timestamp < ?`
	result, err := r.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// appendFilterConditions appends filter conditions to the query.
func (r *EventRepository) appendFilterConditions(query string, filter *event.EventFilter, args *[]interface{}) {
	if len(filter.EventTypes) > 0 {
		placeholders := make([]string, len(filter.EventTypes))
		for i, et := range filter.EventTypes {
			placeholders[i] = "?"
			*args = append(*args, string(et))
		}
		query += " AND event_type IN (" + strings.Join(placeholders, ",") + ")"
	}

	if len(filter.Severities) > 0 {
		placeholders := make([]string, len(filter.Severities))
		for i, s := range filter.Severities {
			placeholders[i] = "?"
			*args = append(*args, string(s))
		}
		query += " AND severity IN (" + strings.Join(placeholders, ",") + ")"
	}

	if !filter.StartTime.IsZero() {
		*args = append(*args, filter.StartTime)
		query += " AND timestamp >= ?"
	}

	if !filter.EndTime.IsZero() {
		*args = append(*args, filter.EndTime)
		query += " AND timestamp <= ?"
	}
}

// scanEvents scans rows into event slices.
func (r *EventRepository) scanEvents(rows *sql.Rows) ([]event.Event, error) {
	var events []event.Event
	for rows.Next() {
		var evt event.Event
		var dataBytes []byte
		var timestamp time.Time

		err := rows.Scan(
			&evt.ID,
			&evt.DeviceID,
			&evt.Type,
			&timestamp,
			&dataBytes,
			&evt.Severity,
			&evt.Source,
			&evt.OperatorID,
		)
		if err != nil {
			return nil, err
		}

		evt.Timestamp = timestamp

		if len(dataBytes) > 0 {
			if err := json.Unmarshal(dataBytes, &evt.Data); err != nil {
				evt.Data = nil
			}
		}

		events = append(events, evt)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
