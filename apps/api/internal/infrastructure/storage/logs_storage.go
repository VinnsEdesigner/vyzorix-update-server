package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/logs"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/uuid"
)

// Ensure LogsRepository implements logs.Repository.
var _ logs.Repository = (*LogsRepository)(nil)

// LogsRepository implements logs.Repository using SQLite.
type LogsRepository struct {
	db *sql.DB
}

// NewLogsRepository creates a new LogsRepository.
func NewLogsRepository(db *sql.DB) *LogsRepository {
	return &LogsRepository{db: db}
}
// getQuerier returns the transaction from context if available, otherwise the db.
func (r *LogsRepository) getQuerier(ctx context.Context) Querier {
if tx, ok := transaction.TxFromContext(ctx); ok {
return tx
}
return r.db
}

// queryRow is a helper that uses transaction-aware querier.
func (r *LogsRepository) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
return r.getQuerier(ctx).QueryRowContext(ctx, query, args...)
}

// queryRows is a helper that uses transaction-aware querier.
func (r *LogsRepository) queryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
return r.getQuerier(ctx).QueryContext(ctx, query, args...)
}

// exec is a helper that uses transaction-aware querier.
func (r *LogsRepository) exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
return r.getQuerier(ctx).ExecContext(ctx, query, args...)
}

// CreateLog creates a new device log entry.
func (r *LogsRepository) CreateLog(ctx context.Context, log *logs.DeviceLog) error {
	var dataJSON []byte
	var err error
	if log.Data != nil {
		dataJSON = log.Data
	} else {
		dataJSON = []byte("{}")
	}

	// Generate ID if not set
	id := log.ID
	if id == "" {
		id = uuid.New()
	}

	query := `
		INSERT INTO device_logs (id, device_id, event_type, timestamp, data)
		VALUES (?, ?, ?, ?, ?)`

	_, err = r.exec(ctx, query,
		id, log.DeviceID, log.EventType, log.Timestamp.UnixMilli(), dataJSON,
	)

	return err
}

// GetLogByID retrieves a log entry by ID.
func (r *LogsRepository) GetLogByID(ctx context.Context, id string) (*logs.DeviceLog, error) {
	query := `
		SELECT id, device_id, event_type, timestamp, data
		FROM device_logs WHERE id = ?`

	var log logs.DeviceLog
	var timestamp int64
	var data []byte

	err := r.queryRow(ctx, query, id).Scan(
		&log.ID, &log.DeviceID, &log.EventType, &timestamp, &data,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, logs.ErrLogNotFound
	}

	if err != nil {
		return nil, err
	}

	log.Timestamp = time.UnixMilli(timestamp)
	if data != nil {
		log.Data = json.RawMessage(data)
	}

	return &log, nil
}

// ListLogs retrieves paginated device logs with cursor-based pagination.
// Returns logs ordered by timestamp DESC, id DESC.
func (r *LogsRepository) ListLogs(ctx context.Context, deviceID string, eventType string, startTime, endTime time.Time, limit int, cursor string) ([]*logs.DeviceLog, string, error) {
	// Convert time to milliseconds for INTEGER column
	startMs := startTime.UnixMilli()
	endMs := endTime.UnixMilli()

	// Build query with time range filter
	baseQuery := `FROM device_logs WHERE device_id = ? AND timestamp >= ? AND timestamp <= ?`
	args := []interface{}{deviceID, startMs, endMs}

	if eventType != "" {
		baseQuery += ` AND event_type = ?`
		args = append(args, eventType)
	}

	// Apply cursor if provided (cursor format: timestampMs_id)
	if cursor != "" {
		// Parse cursor - format: timestampMs_id
		idx := -1
		for i := len(cursor) - 1; i >= 0; i-- {
			if cursor[i] == '_' {
				idx = i
				break
			}
		}
		if idx > 0 {
			cursorTimeMsStr := cursor[:idx]
			cursorID := cursor[idx+1:]
			cursorTimeMs, err := strconv.ParseInt(cursorTimeMsStr, 10, 64)
			if err == nil {
				baseQuery += ` AND (timestamp < ? OR (timestamp = ? AND id < ?))`
				args = append(args, cursorTimeMs, cursorTimeMs, cursorID)
			}
		}
	}

	// Get one extra to determine if there's a next page
	query := `SELECT id, device_id, event_type, timestamp, data ` + baseQuery + ` ORDER BY timestamp DESC, id DESC LIMIT ?`
	args = append(args, limit+1)

	rows, err := r.queryRows(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}

	defer func() { _ = rows.Close() }()

	var logList []*logs.DeviceLog
	var lastLog *logs.DeviceLog

	for rows.Next() {
		log := logs.DeviceLog{}
		var timestamp int64
		var data []byte

		err := rows.Scan(&log.ID, &log.DeviceID, &log.EventType, &timestamp, &data)
		if err != nil {
			return nil, "", err
		}

		log.Timestamp = time.UnixMilli(timestamp)
		if data != nil {
			log.Data = json.RawMessage(data)
		}

		logList = append(logList, &log)
	}

	// Determine next cursor
	var nextCursor string
	if len(logList) > limit {
		// There's a next page, use the last item as cursor
		logList = logList[:limit]
		lastLog = logList[len(logList)-1]
		nextCursor = strconv.FormatInt(lastLog.Timestamp.UnixMilli(), 10) + "_" + lastLog.ID
	}

	return logList, nextCursor, rows.Err()
}

// CountLogs counts logs matching the criteria.
// If deviceID is empty, counts across all devices.
func (r *LogsRepository) CountLogs(ctx context.Context, deviceID string, eventType string, startTime, endTime time.Time) (int, error) {
	// Convert time to milliseconds for INTEGER column
	startMs := startTime.UnixMilli()
	endMs := endTime.UnixMilli()

	var query string
	var args []interface{}

	if deviceID == "" {
		// Global count - no device filter
		query = `SELECT COUNT(*) FROM device_logs WHERE timestamp >= ? AND timestamp <= ?`
		args = []interface{}{startMs, endMs}
	} else {
		query = `SELECT COUNT(*) FROM device_logs WHERE device_id = ? AND timestamp >= ? AND timestamp <= ?`
		args = []interface{}{deviceID, startMs, endMs}
	}

	if eventType != "" {
		query += ` AND event_type = ?`
		args = append(args, eventType)
	}

	var count int
	err := r.queryRow(ctx, query, args...).Scan(&count)

	return count, err
}
