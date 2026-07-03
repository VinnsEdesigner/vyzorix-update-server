package storage

import (
	"context"
	"database/sql"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/inbox"
)

// Ensure InboxRepository implements inbox.Repository.
var _ inbox.Repository = (*InboxRepository)(nil)

// InboxRepository implements inbox.Repository using SQLite.
type InboxRepository struct {
	db *sql.DB
}

// NewInboxRepository creates a new InboxRepository.
func NewInboxRepository(db *sql.DB) *InboxRepository {
	return &InboxRepository{db: db}
}

// Create creates a new inbox entry.
func (r *InboxRepository) Create(ctx context.Context, e *inbox.InboxEntry) error {
	query := `
		INSERT INTO inbox_requests (
			id, device_imei, firebase_install_id, fcm_token, device_name,
			os_version, app_version, device_class, device_model, status,
			reviewed_by, reviewed_at, reviewed_reason, rejection_reason,
			command_secret, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		e.ID, e.IMEI, e.FirebaseInstallID, e.FCMToken, "",
		e.OSVersion, e.AppVersion, "", e.Model, e.Status,
		e.OperatorID, e.ApprovedAt, "", "",
		e.CommandSecret, e.CreatedAt, e.CreatedAt)
	return err
}

// GetByID retrieves an inbox entry by ID.
func (r *InboxRepository) GetByID(ctx context.Context, id string) (*inbox.InboxEntry, error) {
	query := `
		SELECT id, device_imei, firebase_install_id, fcm_token, device_model,
			   os_version, app_version, device_class, manufacturer, status,
			   reviewed_by, reviewed_at, reviewed_reason, rejection_reason,
			   command_secret, created_at, updated_at
		FROM inbox_requests WHERE id = ?`

	return r.scanEntry(r.db.QueryRowContext(ctx, query, id))
}

// GetByIMEI retrieves an inbox entry by IMEI.
func (r *InboxRepository) GetByIMEI(ctx context.Context, imei string) (*inbox.InboxEntry, error) {
	query := `
		SELECT id, device_imei, firebase_install_id, fcm_token, device_model,
			   os_version, app_version, device_class, manufacturer, status,
			   reviewed_by, reviewed_at, reviewed_reason, rejection_reason,
			   command_secret, created_at, updated_at
		FROM inbox_requests WHERE device_imei = ?`

	return r.scanEntry(r.db.QueryRowContext(ctx, query, imei))
}

// List retrieves paginated inbox entries with optional status filter.
func (r *InboxRepository) List(ctx context.Context, status string, limit, offset int) ([]*inbox.InboxEntry, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var whereClause string
	args := []interface{}{}

	switch strings.ToLower(status) {
	case "pending", "approved", "rejected":
		whereClause = "WHERE status = ?"
		args = append(args, status)
	case "all":
		whereClause = ""
	default:
		whereClause = "WHERE status = 'pending'"
		args = append(args, "pending")
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM inbox_requests " + whereClause
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get entries
	listQuery := `
		SELECT id, device_imei, firebase_install_id, fcm_token, device_model,
			   os_version, app_version, device_class, manufacturer, status,
			   reviewed_by, reviewed_at, reviewed_reason, rejection_reason,
			   command_secret, created_at, updated_at
		FROM inbox_requests
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []*inbox.InboxEntry
	for rows.Next() {
		entry, err := r.scanEntryRows(rows)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, entry)
	}

	return entries, total, rows.Err()
}

// Update updates an existing inbox entry.
func (r *InboxRepository) Update(ctx context.Context, e *inbox.InboxEntry) error {
	query := `
		UPDATE inbox_requests SET
			device_imei = ?, firebase_install_id = ?, fcm_token = ?,
			device_name = ?, os_version = ?, app_version = ?,
			device_class = ?, device_model = ?, status = ?,
			reviewed_by = ?, reviewed_at = ?, reviewed_reason = ?,
			rejection_reason = ?, command_secret = ?, updated_at = ?
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		e.IMEI, e.FirebaseInstallID, e.FCMToken,
		"", e.OSVersion, e.AppVersion,
		"", e.Model, e.Status,
		e.OperatorID, e.ApprovedAt, "",
		e.Notes, e.CommandSecret, e.CreatedAt, e.ID)
	return err
}

// Delete deletes an inbox entry by ID.
func (r *InboxRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM inbox_requests WHERE id = ?", id)
	return err
}

// ExistsByIMEI checks if an inbox entry exists for the given IMEI.
func (r *InboxRepository) ExistsByIMEI(ctx context.Context, imei string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM inbox_requests WHERE device_imei = ?)", imei).Scan(&exists)
	return exists, err
}

// ExistsByFirebaseInstallID checks if an inbox entry exists for the given Firebase install ID.
func (r *InboxRepository) ExistsByFirebaseInstallID(ctx context.Context, firebaseInstallID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM inbox_requests WHERE firebase_install_id = ?)", firebaseInstallID).Scan(&exists)
	return exists, err
}

// Count returns the total number of inbox entries with optional status filter.
func (r *InboxRepository) Count(ctx context.Context, status string) (int, error) {
	var count int
	var query string
	var args []interface{}

	if status == "" || status == "all" {
		query = "SELECT COUNT(*) FROM inbox_requests"
	} else {
		query = "SELECT COUNT(*) FROM inbox_requests WHERE status = ?"
		args = append(args, status)
	}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// scanEntry scans a single inbox entry from a row.
func (r *InboxRepository) scanEntry(row *sql.Row) (*inbox.InboxEntry, error) {
	var e inbox.InboxEntry
	var reviewedAt sql.NullInt64
	var commandSecret, notes, model, osVersion, appVersion, manufacturer sql.NullString

	err := row.Scan(
		&e.ID, &e.IMEI, &e.FirebaseInstallID, &e.FCMToken, &model,
		&osVersion, &appVersion, &manufacturer, &e.Status,
		&e.OperatorID, &reviewedAt, &notes, &e.Notes,
		&commandSecret, &e.CreatedAt, &e.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, inbox.ErrInboxNotFound
	}
	if err != nil {
		return nil, err
	}

	e.Model = model.String
	e.Manufacturer = manufacturer.String
	e.OSVersion = osVersion.String
	e.AppVersion = appVersion.String
	e.Notes = notes.String
	e.CommandSecret = commandSecret.String

	if reviewedAt.Valid {
		e.ApprovedAt = &reviewedAt.Int64
	}

	return &e, nil
}

// scanEntryRows scans inbox entries from rows.
func (r *InboxRepository) scanEntryRows(rows *sql.Rows) (*inbox.InboxEntry, error) {
	var e inbox.InboxEntry
	var reviewedAt sql.NullInt64
	var commandSecret, notes, model, osVersion, appVersion, manufacturer sql.NullString

	err := rows.Scan(
		&e.ID, &e.IMEI, &e.FirebaseInstallID, &e.FCMToken, &model,
		&osVersion, &appVersion, &manufacturer, &e.Status,
		&e.OperatorID, &reviewedAt, &notes, &e.Notes,
		&commandSecret, &e.CreatedAt, &e.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	e.Model = model.String
	e.Manufacturer = manufacturer.String
	e.OSVersion = osVersion.String
	e.AppVersion = appVersion.String
	e.Notes = notes.String
	e.CommandSecret = commandSecret.String

	if reviewedAt.Valid {
		e.ApprovedAt = &reviewedAt.Int64
	}

	return &e, nil
}
