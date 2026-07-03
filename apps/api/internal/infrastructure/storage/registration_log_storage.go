package storage

import (
	"context"
	"database/sql"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/inbox"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Ensure RegistrationLogRepository implements inbox.RegistrationLogRepository.
var _ inbox.RegistrationLogRepository = (*RegistrationLogRepository)(nil)

// RegistrationLogRepository implements inbox.RegistrationLogRepository using SQLite.
type RegistrationLogRepository struct {
	db *sql.DB
}

// NewRegistrationLogRepository creates a new RegistrationLogRepository.
func NewRegistrationLogRepository(db *sql.DB) *RegistrationLogRepository {
	return &RegistrationLogRepository{db: db}
}
// getQuerier returns the transaction from context if available, otherwise the db.
func (r *RegistrationLogRepository) getQuerier(ctx context.Context) Querier {
if tx, ok := transaction.TxFromContext(ctx); ok {
return tx
}
return r.db
}

// queryRow is a helper that uses transaction-aware querier.
func (r *RegistrationLogRepository) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
return r.getQuerier(ctx).QueryRowContext(ctx, query, args...)
}

// queryRows is a helper that uses transaction-aware querier.
func (r *RegistrationLogRepository) queryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
return r.getQuerier(ctx).QueryContext(ctx, query, args...)
}

// exec is a helper that uses transaction-aware querier.
func (r *RegistrationLogRepository) exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
return r.getQuerier(ctx).ExecContext(ctx, query, args...)
}

// Create creates a new registration log entry.
func (r *RegistrationLogRepository) Create(ctx context.Context, log *inbox.RegistrationLog) error {
	query := `
		INSERT INTO registration_logs (
			id, inbox_request_id, device_id, action, old_status,
			new_status, performed_by, reason, created_at,
			client_ip, user_agent
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.exec(ctx, query,
		log.ID, "", log.DeviceID, log.Action, "",
		"", log.OperatorID, log.Details, log.Timestamp,
		nullString(log.ClientIP), nullString(log.UserAgent))
	return err
}

// ListByDeviceID retrieves all registration logs for a device.
func (r *RegistrationLogRepository) ListByDeviceID(ctx context.Context, deviceID string, limit, offset int) ([]*inbox.RegistrationLog, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM registration_logs WHERE device_id = ?"
	if err := r.queryRow(ctx, countQuery, deviceID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get logs
	query := `
		SELECT id, inbox_request_id, device_id, action, old_status,
			   new_status, performed_by, reason, created_at
		FROM registration_logs
		WHERE device_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.queryRows(ctx, query, deviceID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*inbox.RegistrationLog
	for rows.Next() {
		log, err := r.scanLog(rows)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, rows.Err()
}

// ListByIMEI retrieves all registration logs for an IMEI.
func (r *RegistrationLogRepository) ListByIMEI(ctx context.Context, imei string, limit, offset int) ([]*inbox.RegistrationLog, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// First get all inbox request IDs for this IMEI
	var inboxIDs []string
	idQuery := "SELECT id FROM inbox_requests WHERE device_imei = ?"
	rows, err := r.queryRows(ctx, idQuery, imei)
	if err != nil {
		return nil, 0, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, 0, err
		}
		inboxIDs = append(inboxIDs, id)
	}
	rows.Close()

	if len(inboxIDs) == 0 {
		return []*inbox.RegistrationLog{}, 0, nil
	}

	// Build query with IN clause
	placeholders := ""
	args := []interface{}{}
	for i, id := range inboxIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM registration_logs WHERE inbox_request_id IN (" + placeholders + ")"
	if err := r.queryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get logs
	listQuery := `
		SELECT id, inbox_request_id, device_id, action, old_status,
			   new_status, performed_by, reason, created_at
		FROM registration_logs
		WHERE inbox_request_id IN (` + placeholders + `)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	args = append(args, limit, offset)
	logRows, err := r.queryRows(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer logRows.Close()

	var logs []*inbox.RegistrationLog
	for logRows.Next() {
		log, err := r.scanLog(logRows)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, logRows.Err()
}

// ListByOperator retrieves all registration logs for an operator.
func (r *RegistrationLogRepository) ListByOperator(ctx context.Context, operatorID string, limit, offset int) ([]*inbox.RegistrationLog, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM registration_logs WHERE performed_by = ?"
	if err := r.queryRow(ctx, countQuery, operatorID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get logs
	query := `
		SELECT id, inbox_request_id, device_id, action, old_status,
			   new_status, performed_by, reason, created_at
		FROM registration_logs
		WHERE performed_by = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.queryRows(ctx, query, operatorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*inbox.RegistrationLog
	for rows.Next() {
		log, err := r.scanLog(rows)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, rows.Err()
}

// CountByOperator returns the number of registration logs for an operator.
func (r *RegistrationLogRepository) CountByOperator(ctx context.Context, operatorID string) (int, error) {
	var count int
	err := r.queryRow(ctx,
		"SELECT COUNT(*) FROM registration_logs WHERE performed_by = ?", operatorID).Scan(&count)
	return count, err
}

// scanLog scans a registration log from rows.
func (r *RegistrationLogRepository) scanLog(rows *sql.Rows) (*inbox.RegistrationLog, error) {
	var log inbox.RegistrationLog
	var inboxRequestID, oldStatus, newStatus, reason, clientIP, userAgent sql.NullString

	err := rows.Scan(
		&log.ID, &inboxRequestID, &log.DeviceID, &log.Action, &oldStatus,
		&newStatus, &log.OperatorID, &reason, &log.Timestamp,
		&clientIP, &userAgent,
	)
	if err != nil {
		return nil, err
	}

	log.Details = reason.String
	log.ClientIP = clientIP.String
	log.UserAgent = userAgent.String
	return &log, nil
}
