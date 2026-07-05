package notification

import (
	"context"
	"database/sql"
	"time"
)

// AuditEntry represents a notification audit log entry.
type AuditEntry struct {
	SentAt     time.Time `json:"sentAt"`
	ID         string    `json:"id"`
	OperatorID string    `json:"operatorId"`
	EventType  string    `json:"eventType"`
	Channel    string    `json:"channel"`
	DeviceID   string    `json:"deviceId,omitempty"`
	Payload    string    `json:"payload,omitempty"`
	ErrorMsg   string    `json:"errorMsg,omitempty"`
	Success    bool      `json:"success"`
}

// Repository handles notification audit log operations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new notification audit repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// LogEntry logs a notification audit entry.
func (r *Repository) LogEntry(ctx context.Context, entry *AuditEntry) error {
	query := `
		INSERT INTO notification_audit_log (
			id, operator_id, event_type, channel, device_id, payload, success, error_msg, sent_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		entry.OperatorID,
		entry.EventType,
		entry.Channel,
		entry.DeviceID,
		entry.Payload,
		entry.Success,
		entry.ErrorMsg,
		entry.SentAt.UnixMilli(),
	)

	return err
}

// GetByOperator retrieves audit logs for an operator.
func (r *Repository) GetByOperator(ctx context.Context, operatorID string, limit, offset int) ([]*AuditEntry, error) {
	query := `
		SELECT id, operator_id, event_type, channel, device_id, payload, success, error_msg, sent_at
		FROM notification_audit_log
		WHERE operator_id = ?
		ORDER BY sent_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, operatorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var entries []*AuditEntry
	for rows.Next() {
		var entry AuditEntry
		var sentAt int64
		var deviceID, payload, errorMsg sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.OperatorID,
			&entry.EventType,
			&entry.Channel,
			&deviceID,
			&payload,
			&entry.Success,
			&errorMsg,
			&sentAt,
		)
		if err != nil {
			return nil, err
		}

		entry.DeviceID = deviceID.String
		entry.Payload = payload.String
		entry.ErrorMsg = errorMsg.String
		entry.SentAt = time.UnixMilli(sentAt)

		entries = append(entries, &entry)
	}

	return entries, rows.Err()
}

// GetByEventType retrieves audit logs by event type.
func (r *Repository) GetByEventType(ctx context.Context, eventType string, limit, offset int) ([]*AuditEntry, error) {
	query := `
		SELECT id, operator_id, event_type, channel, device_id, payload, success, error_msg, sent_at
		FROM notification_audit_log
		WHERE event_type = ?
		ORDER BY sent_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, eventType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var entries []*AuditEntry
	for rows.Next() {
		var entry AuditEntry
		var sentAt int64
		var deviceID, payload, errorMsg sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.OperatorID,
			&entry.EventType,
			&entry.Channel,
			&deviceID,
			&payload,
			&entry.Success,
			&errorMsg,
			&sentAt,
		)
		if err != nil {
			return nil, err
		}

		entry.DeviceID = deviceID.String
		entry.Payload = payload.String
		entry.ErrorMsg = errorMsg.String
		entry.SentAt = time.UnixMilli(sentAt)

		entries = append(entries, &entry)
	}

	return entries, rows.Err()
}
