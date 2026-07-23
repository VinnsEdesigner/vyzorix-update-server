package storage

import (
	"context"
	"database/sql"
	"encoding/json"
)

// PendingFCMNotification represents a pending FCM notification for retry.
type PendingFCMNotification struct {
	ID           int64  `json:"id"`
	DispatchID   string `json:"dispatchId"`
	DeviceID     string `json:"deviceId"`
	Token        string `json:"token"`
	Command      string `json:"command"`
	Priority     string `json:"priority"`
	RetryCount   int    `json:"retryCount"`
	NextRetryAt  int64  `json:"nextRetryAt"`  // Unix timestamp
	LastError    string `json:"lastError"`
	CreatedAt    int64  `json:"createdAt"`   // Unix timestamp
	UpdatedAt    int64  `json:"updatedAt"`   // Unix timestamp
}

// PendingFCMRepository handles pending FCM notification persistence.
type PendingFCMRepository struct {
	db *sql.DB
}

// NewPendingFCMRepository creates a new pending FCM repository.
func NewPendingFCMRepository(db *sql.DB) *PendingFCMRepository {
	return &PendingFCMRepository{db: db}
}

// Create inserts a new pending notification.
func (r *PendingFCMRepository) Create(ctx context.Context, notification *PendingFCMNotification) error {
	query := `
		INSERT INTO pending_fcm (dispatch_id, device_id, token, command, priority, retry_count, next_retry_at, last_error, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		notification.DispatchID,
		notification.DeviceID,
		notification.Token,
		notification.Command,
		notification.Priority,
		notification.RetryCount,
		notification.NextRetryAt,
		notification.LastError,
		notification.CreatedAt,
		notification.UpdatedAt,
	)
	return err
}

// GetPending retrieves pending notifications due for retry.
func (r *PendingFCMRepository) GetPending(ctx context.Context, limit int) ([]PendingFCMNotification, error) {
	query := `
		SELECT id, dispatch_id, device_id, token, command, priority, retry_count, next_retry_at, last_error, created_at, updated_at
		FROM pending_fcm
		WHERE next_retry_at <= ?
		ORDER BY next_retry_at ASC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, context.Background().Value("now"), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []PendingFCMNotification
	for rows.Next() {
		var n PendingFCMNotification
		err := rows.Scan(
			&n.ID, &n.DispatchID, &n.DeviceID, &n.Token, &n.Command, &n.Priority,
			&n.RetryCount, &n.NextRetryAt, &n.LastError, &n.CreatedAt, &n.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

// Update updates a pending notification after a retry attempt.
func (r *PendingFCMRepository) Update(ctx context.Context, notification *PendingFCMNotification) error {
	query := `
		UPDATE pending_fcm
		SET retry_count = ?, next_retry_at = ?, last_error = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		notification.RetryCount,
		notification.NextRetryAt,
		notification.LastError,
		notification.UpdatedAt,
		notification.ID,
	)
	return err
}

// Delete removes a notification after successful delivery.
func (r *PendingFCMRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM pending_fcm WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DeleteByDispatchID removes all pending notifications for a dispatch.
func (r *PendingFCMRepository) DeleteByDispatchID(ctx context.Context, dispatchID string) error {
	query := `DELETE FROM pending_fcm WHERE dispatch_id = ?`
	_, err := r.db.ExecContext(ctx, query, dispatchID)
	return err
}

// migrateCreatePendingFCM creates the pending_fcm table for FCM retry persistence.
func migrateCreatePendingFCM(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS pending_fcm (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			dispatch_id     TEXT NOT NULL,
			device_id       TEXT NOT NULL,
			token           TEXT NOT NULL,
			command         TEXT NOT NULL,
			priority        TEXT NOT NULL DEFAULT 'high',
			retry_count     INTEGER NOT NULL DEFAULT 0,
			next_retry_at   INTEGER NOT NULL,
			last_error      TEXT,
			created_at     INTEGER NOT NULL,
			updated_at     INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying pending notifications
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_pending_fcm_next_retry
		ON pending_fcm(next_retry_at)
	`)
	if err != nil {
		return err
	}

	// Create index for dispatch lookup
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_pending_fcm_dispatch
		ON pending_fcm(dispatch_id)
	`)
	return err
}

// FCMNotificationData is used to serialize the notification payload for storage.
type FCMNotificationData struct {
	Token         string `json:"token"`
	Command       string `json:"command"`
	DispatchID    string `json:"dispatchId"`
	DeviceID      string `json:"deviceId"`
	Priority      string `json:"priority"`
	CommandSecret string `json:"commandSecret,omitempty"`
	APKFilename   string `json:"apkFilename,omitempty"`
	SHA256        string `json:"sha256,omitempty"`
	DownloadURL   string `json:"downloadUrl,omitempty"`
	APKSize       int64  `json:"apkSize,omitempty"`
}

// Serialize serializes notification data to JSON.
func (n *FCMNotificationData) Serialize() (string, error) {
	data, err := json.Marshal(n)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeserializePendingFCMNotification creates a PendingFCMNotification from stored data.
func DeserializePendingFCMNotification(id int64, dispatchID, deviceID, token, command, priority string, retryCount int, nextRetryAt int64, lastError string, createdAt, updatedAt int64) *PendingFCMNotification {
	return &PendingFCMNotification{
		ID:          id,
		DispatchID:  dispatchID,
		DeviceID:    deviceID,
		Token:       token,
		Command:     command,
		Priority:    priority,
		RetryCount:  retryCount,
		NextRetryAt: nextRetryAt,
		LastError:   lastError,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
