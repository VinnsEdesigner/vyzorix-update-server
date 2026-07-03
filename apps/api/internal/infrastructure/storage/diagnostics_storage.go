package storage

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/diagnostics"
)

// Ensure DiagnosticsRepository implements diagnostics.Repository.
var _ diagnostics.Repository = (*DiagnosticsRepository)(nil)

// DiagnosticsRepository implements diagnostics.Repository using SQLite.
type DiagnosticsRepository struct {
	db *sql.DB
}

// NewDiagnosticsRepository creates a new DiagnosticsRepository.
func NewDiagnosticsRepository(db *sql.DB) *DiagnosticsRepository {
	return &DiagnosticsRepository{db: db}
}

// GetTimelineEvents retrieves paginated timeline events for a device.
func (r *DiagnosticsRepository) GetTimelineEvents(ctx context.Context, deviceID string, filter *diagnostics.TimelineFilter) (*diagnostics.TimelineResult, error) {
	// Build query based on filter
	query := `
		SELECT id, event_type, timestamp, data
		FROM device_events
		WHERE device_id = ?`
	
	args := []interface{}{deviceID}
	
	// Apply event type filter if specified (empty means no filter)
	if filter.EventType != "" {
		query += " AND event_type = ?"
		args = append(args, filter.EventType)
	}
	
	// Apply cursor-based pagination
	if filter.Cursor != "" {
		cursorTime, cursorID := r.decodeCursor(filter.Cursor)
		if !cursorTime.IsZero() {
			query += " AND (timestamp < ? OR (timestamp = ? AND id < ?))"
			args = append(args, cursorTime, cursorTime, cursorID)
		}
	}
	
	// Apply time range filters
	if !filter.StartTime.IsZero() {
		query += " AND timestamp >= ?"
		args = append(args, filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		query += " AND timestamp <= ?"
		args = append(args, filter.EndTime)
	}
	
	// Order by timestamp desc, then id desc
	query += " ORDER BY timestamp DESC, id DESC"
	
	// Fetch limit + 1 to check for more results
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	query += " LIMIT ?"
	args = append(args, limit+1)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	
	var events []diagnostics.TimelineEvent
	for rows.Next() {
		var event diagnostics.TimelineEvent
		var dataBytes []byte
		var timestamp time.Time
		
		err := rows.Scan(&event.ID, &event.Type, &timestamp, &dataBytes)
		if err != nil {
			return nil, err
		}
		
		event.Timestamp = timestamp
		
		if len(dataBytes) > 0 {
			if err := json.Unmarshal(dataBytes, &event.Data); err != nil {
				// Log error but continue
				event.Data = nil
			}
		}
		
		events = append(events, event)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	// Check if there are more results
	hasMore := len(events) > limit
	if hasMore {
		events = events[:limit]
	}
	
	// Generate next cursor
	var nextCursor string
	if hasMore && len(events) > 0 {
		last := events[len(events)-1]
		nextCursor = r.encodeCursor(last.Timestamp, last.ID)
	}
	
	return &diagnostics.TimelineResult{
		Events:     events,
		HasMore:   hasMore,
		NextCursor: nextCursor,
	}, nil
}

// RecordEvent records a new device event.
func (r *DiagnosticsRepository) RecordEvent(ctx context.Context, event *diagnostics.TimelineEvent) error {
	var dataBytes []byte
	var err error
	
	if event.Data != nil {
		dataBytes, err = json.Marshal(event.Data)
		if err != nil {
			return err
		}
	}
	
	query := `
		INSERT INTO device_events (id, device_id, event_type, timestamp, data)
		VALUES (?, ?, ?, ?, ?)`
	
	_, err = r.db.ExecContext(ctx, query, event.ID, event.DeviceID, event.Type, event.Timestamp, dataBytes)
	return err
}

// GetTelemetryStats retrieves telemetry statistics for a device for today.
func (r *DiagnosticsRepository) GetTelemetryStats(ctx context.Context, deviceID string) (*diagnostics.TelemetryInfo, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	// Count frames today
	framesQuery := `
		SELECT COUNT(*) FROM telemetry 
		WHERE device_id = ? AND timestamp >= ?`
	
	var framesToday int
	err := r.db.QueryRowContext(ctx, framesQuery, deviceID, startOfDay).Scan(&framesToday)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	
	// Get total bytes today (from telemetry size estimate)
	bytesQuery := `
		SELECT COALESCE(SUM(length(frame_data)), 0) FROM telemetry 
		WHERE device_id = ? AND timestamp >= ?`
	
	var totalBytesToday int64
	err = r.db.QueryRowContext(ctx, bytesQuery, deviceID, startOfDay).Scan(&totalBytesToday)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	
	// Count sessions today (distinct connection sessions based on telemetry)
	sessionsQuery := `
		SELECT COUNT(DISTINCT strftime('%Y-%m-%d %H', timestamp)) FROM telemetry 
		WHERE device_id = ? AND timestamp >= ?`
	
	var sessionsToday int
	err = r.db.QueryRowContext(ctx, sessionsQuery, deviceID, startOfDay).Scan(&sessionsToday)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	
	return &diagnostics.TelemetryInfo{
		FramesToday:    framesToday,
		TotalBytesToday: totalBytesToday,
		SessionsToday:  sessionsToday,
	}, nil
}

// GetLastTelemetry retrieves the most recent telemetry data for a device.
func (r *DiagnosticsRepository) GetLastTelemetry(ctx context.Context, deviceID string) (*diagnostics.TimelineEvent, error) {
	query := `
		SELECT timestamp, frame_data FROM telemetry 
		WHERE device_id = ? 
		ORDER BY timestamp DESC LIMIT 1`
	
	var timestamp time.Time
	var frameData []byte
	
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(&timestamp, &frameData)
	if err == sql.ErrNoRows {
		return nil, diagnostics.ErrNoTelemetryData
	}
	if err != nil {
		return nil, err
	}
	
	// Parse frame data to extract telemetry info
	var data map[string]any
	if len(frameData) > 0 {
		if err := json.Unmarshal(frameData, &data); err != nil {
			data = nil
		}
	}
	
	return &diagnostics.TimelineEvent{
		ID:        "",
		DeviceID:  deviceID,
		Type:      diagnostics.EventTypeTelemetry,
		Timestamp: timestamp,
		Data:      data,
	}, nil
}

// decodeCursor decodes a pagination cursor.
func (r *DiagnosticsRepository) decodeCursor(encoded string) (time.Time, string) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return time.Time{}, ""
	}
	
	var cursor struct {
		T string `json:"t"`
		I string `json:"i"`
	}
	if err := json.Unmarshal(data, &cursor); err != nil {
		return time.Time{}, ""
	}
	
	t, _ := time.Parse(time.RFC3339Nano, cursor.T)
	return t, cursor.I
}

// encodeCursor encodes a pagination cursor.
func (r *DiagnosticsRepository) encodeCursor(t time.Time, id string) string {
	cursor := struct {
		T string `json:"t"`
		I string `json:"i"`
	}{t.Format(time.RFC3339Nano), id}
	
	data, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}
