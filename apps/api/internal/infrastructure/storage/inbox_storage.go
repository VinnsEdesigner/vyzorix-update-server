package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/inbox"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
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

// getQuerier returns the transaction from context if available, otherwise the db.
func (r *InboxRepository) getQuerier(ctx context.Context) Querier {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}

// queryRow is a helper that uses transaction-aware querier.
func (r *InboxRepository) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return r.getQuerier(ctx).QueryRowContext(ctx, query, args...)
}

// queryRows is a helper that uses transaction-aware querier.
func (r *InboxRepository) queryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return r.getQuerier(ctx).QueryContext(ctx, query, args...)
}

// exec is a helper that uses transaction-aware querier.
func (r *InboxRepository) exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getQuerier(ctx).ExecContext(ctx, query, args...)
}

// Create creates a new inbox entry.
func (r *InboxRepository) Create(ctx context.Context, e *inbox.InboxEntry) error {
	now := time.Now().UnixMilli()
	query := `
		INSERT INTO inbox_requests (
			id, device_imei, firebase_install_id, fcm_token, device_name,
			manufacturer, os_version, app_version, device_class, device_model, status,
			reviewed_by, reviewed_at, reviewed_reason, rejection_reason,
			command_secret_hash, organization_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.exec(ctx, query,
		e.ID, e.IMEI, e.FirebaseInstallID, e.FCMToken, e.DeviceName,
		e.Manufacturer, e.OSVersion, e.AppVersion, e.DeviceClass, e.Model, e.Status,
		"", nil, "", "",
		e.CommandSecretHash, e.OrganizationID, e.CreatedAt, now)
	return err
}

// CreateOrReplace creates a new inbox entry or replaces an existing one for the same IMEI.
// This is atomic and avoids TOCTOU races between delete and create.
func (r *InboxRepository) CreateOrReplace(ctx context.Context, e *inbox.InboxEntry) error {
	now := time.Now().UnixMilli()
	query := `
		INSERT OR REPLACE INTO inbox_requests (
			id, device_imei, firebase_install_id, fcm_token, device_name,
			manufacturer, os_version, app_version, device_class, device_model, status,
			reviewed_by, reviewed_at, reviewed_reason, rejection_reason,
			command_secret_hash, organization_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.exec(ctx, query,
		e.ID, e.IMEI, e.FirebaseInstallID, e.FCMToken, e.DeviceName,
		e.Manufacturer, e.OSVersion, e.AppVersion, e.DeviceClass, e.Model, e.Status,
		"", nil, "", "",
		e.CommandSecretHash, e.OrganizationID, e.CreatedAt, now)
	return err
}

// GetByID retrieves an inbox entry by ID.
func (r *InboxRepository) GetByID(ctx context.Context, id string) (*inbox.InboxEntry, error) {
	query := `
		SELECT id, device_imei, firebase_install_id, fcm_token, device_name,
			   manufacturer, os_version, app_version, device_class, device_model, status,
			   reviewed_by, reviewed_at, reviewed_reason, rejection_reason,
			   command_secret_hash, organization_id, created_at, updated_at,
			   confirmed_at
		FROM inbox_requests WHERE id = ?`

	return r.scanEntry(r.queryRow(ctx, query, id))
}

// GetByIMEI retrieves an inbox entry by IMEI.
func (r *InboxRepository) GetByIMEI(ctx context.Context, imei string) (*inbox.InboxEntry, error) {
	query := `
		SELECT id, device_imei, firebase_install_id, fcm_token, device_name,
			   manufacturer, os_version, app_version, device_class, device_model, status,
			   reviewed_by, reviewed_at, reviewed_reason, rejection_reason,
			   command_secret_hash, organization_id, created_at, updated_at,
			   confirmed_at
		FROM inbox_requests WHERE device_imei = ?`

	return r.scanEntry(r.queryRow(ctx, query, imei))
}

// ListByOperator retrieves paginated inbox entries for a specific operator within an organization with optional status filter.
func (r *InboxRepository) ListByOperator(ctx context.Context, operatorID, orgID, status string, limit, offset int) ([]*inbox.InboxEntry, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Build WHERE clause based on status and organization:.
	// - "pending": show ALL pending entries (no operator filter - any operator can see pending).
	// - "approved"/"rejected": show entries with that status AND reviewed_by = operatorID.
	// - "all"/"": show pending entries + entries reviewed by this operator.
	// Always filter by organization ID for multi-tenant isolation.
	var whereClause string
	var args []interface{}

	statusLower := strings.ToLower(status)
	switch statusLower {
	case "pending":
		// Show pending entries that are either unassigned (reviewed_by IS NULL).
		// or assigned to this specific operator, within the organization.
		// This ensures operators only see pending entries they can work on.
		whereClause = "WHERE organization_id = ? AND status = ? AND (reviewed_by IS NULL OR reviewed_by = ?)"
		args = append(args, orgID, inbox.StatusPending, operatorID)
	case "approved", "rejected":
		// Show entries with this status that were reviewed by this operator within the organization.
		whereClause = "WHERE organization_id = ? AND status = ? AND reviewed_by = ?"
		args = append(args, orgID, statusLower, operatorID)
	default:
		// "all" or empty: show pending entries (unassigned or mine) + entries reviewed by this operator.
		whereClause = "WHERE organization_id = ? AND ((status = ? AND (reviewed_by IS NULL OR reviewed_by = ?)) OR reviewed_by = ?)"
		args = append(args, orgID, inbox.StatusPending, operatorID, operatorID)
	}

	// Get total count.
	var total int
	countQuery := "SELECT COUNT(*) FROM inbox_requests " + whereClause
	if err := r.queryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get entries.
	listQuery := `
		SELECT id, device_imei, firebase_install_id, fcm_token, device_name,
			   manufacturer, os_version, app_version, device_class, device_model, status,
			   reviewed_by, reviewed_at, reviewed_reason, rejection_reason,
			   command_secret_hash, organization_id, created_at, updated_at,
			   confirmed_at
		FROM inbox_requests
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	args = append(args, limit, offset)
	rows, err := r.queryRows(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

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
	now := time.Now().UnixMilli()
	query := `
		UPDATE inbox_requests SET
			device_imei = ?, firebase_install_id = ?, fcm_token = ?,
			device_name = ?, manufacturer = ?, os_version = ?, app_version = ?,
			device_class = ?, device_model = ?, status = ?,
			reviewed_by = ?, reviewed_at = ?, reviewed_reason = ?,
			rejection_reason = ?, command_secret_hash = ?, organization_id = ?, updated_at = ?
		WHERE id = ?`

	// Map entity fields to schema columns:.
	// reviewed_at: store ApprovedAt if approved, RejectedAt if rejected.
	// reviewed_reason: store notes when approved.
	// rejection_reason: store notes when rejected.
	var reviewedAt *int64
	var reviewedReason, rejectionReason string

	switch e.Status {
	case inbox.StatusApproved:
		reviewedAt = e.ApprovedAt
		reviewedReason = e.Notes
		rejectionReason = ""
	case inbox.StatusRejected:
		reviewedAt = e.RejectedAt
		reviewedReason = ""
		rejectionReason = e.Notes
	default:
		// StatusPending - no review yet, leave defaults.
	}

	_, err := r.exec(ctx, query,
		e.IMEI, e.FirebaseInstallID, e.FCMToken,
		e.DeviceName, e.Manufacturer, e.OSVersion, e.AppVersion,
		e.DeviceClass, e.Model, e.Status,
		e.OperatorID, reviewedAt, reviewedReason,
		rejectionReason, e.CommandSecretHash, e.OrganizationID, now, e.ID)
	return err
}

// Delete deletes an inbox entry by ID.
func (r *InboxRepository) Delete(ctx context.Context, id string) error {
	_, err := r.exec(ctx, "DELETE FROM inbox_requests WHERE id = ?", id)
	return err
}

// DeleteByIMEI deletes all inbox entries for a given IMEI.
// Used when device re-registers to clean up stale entries.
func (r *InboxRepository) DeleteByIMEI(ctx context.Context, imei string) error {
	_, err := r.exec(ctx, "DELETE FROM inbox_requests WHERE device_imei = ?", imei)
	return err
}

// ExistsByIMEI checks if an inbox entry exists for the given IMEI.
func (r *InboxRepository) ExistsByIMEI(ctx context.Context, imei string) (bool, error) {
	var exists bool
	err := r.queryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM inbox_requests WHERE device_imei = ?)", imei).Scan(&exists)
	return exists, err
}

// ExistsByFirebaseInstallID checks if an inbox entry exists for the given Firebase install ID.
func (r *InboxRepository) ExistsByFirebaseInstallID(ctx context.Context, firebaseInstallID string) (bool, error) {
	var exists bool
	err := r.queryRow(ctx,
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

	err := r.queryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// scanEntry scans a single inbox entry from a row.
func (r *InboxRepository) scanEntry(row *sql.Row) (*inbox.InboxEntry, error) {
	var e inbox.InboxEntry
	var reviewedAt sql.NullInt64
	var reviewedReason, rejectionReason sql.NullString
	var commandSecretHash sql.NullString
	var organizationID sql.NullString
	var confirmedAt sql.NullInt64
	var updatedAt int64

	// SELECT order: id, device_imei, firebase_install_id, fcm_token, device_name,.
	//               manufacturer, os_version, app_version, device_class, device_model, status,.
	//               reviewed_by, reviewed_at, reviewed_reason, rejection_reason,.
	//               command_secret_hash, organization_id, created_at, updated_at,.
	//               confirmed_at.
	err := row.Scan(
		&e.ID, &e.IMEI, &e.FirebaseInstallID, &e.FCMToken, &e.DeviceName,
		&e.Manufacturer, &e.OSVersion, &e.AppVersion, &e.DeviceClass, &e.Model, &e.Status,
		&e.OperatorID, &reviewedAt, &reviewedReason, &rejectionReason,
		&commandSecretHash, &organizationID, &e.CreatedAt, &updatedAt,
		&confirmedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, inbox.ErrInboxNotFound
	}
	if err != nil {
		return nil, err
	}

	e.UpdatedAt = updatedAt
	e.CommandSecretHash = commandSecretHash.String
	e.OrganizationID = organizationID.String
	e.ConfirmedAt = nullInt64ToPtr(confirmedAt)

	// DB schema has single reviewed_at column - use ApprovedAt for approved status.
	// RejectedAt remains nil since there's no separate column.
	switch e.Status {
	case inbox.StatusApproved:
		e.ApprovedAt = nullInt64ToPtr(reviewedAt)
		e.Notes = reviewedReason.String
	case inbox.StatusRejected:
		e.RejectedAt = nullInt64ToPtr(reviewedAt)
		e.Notes = rejectionReason.String
	case inbox.StatusPending, inbox.StatusAcknowledged, inbox.StatusApproving, inbox.StatusExpired:
		// No action needed for non-reviewed statuses.
	}

	return &e, nil
}

// scanEntryRows scans inbox entries from rows.
func (r *InboxRepository) scanEntryRows(rows *sql.Rows) (*inbox.InboxEntry, error) {
	var e inbox.InboxEntry
	var reviewedAt sql.NullInt64
	var reviewedReason, rejectionReason sql.NullString
	var commandSecretHash sql.NullString
	var organizationID sql.NullString
	var confirmedAt sql.NullInt64
	var updatedAt int64

	// SELECT order: id, device_imei, firebase_install_id, fcm_token, device_name,.
	//               manufacturer, os_version, app_version, device_class, device_model, status,.
	//               reviewed_by, reviewed_at, reviewed_reason, rejection_reason,.
	//               command_secret_hash, organization_id, created_at, updated_at,.
	//               confirmed_at.
	err := rows.Scan(
		&e.ID, &e.IMEI, &e.FirebaseInstallID, &e.FCMToken, &e.DeviceName,
		&e.Manufacturer, &e.OSVersion, &e.AppVersion, &e.DeviceClass, &e.Model, &e.Status,
		&e.OperatorID, &reviewedAt, &reviewedReason, &rejectionReason,
		&commandSecretHash, &organizationID, &e.CreatedAt, &updatedAt,
		&confirmedAt,
	)
	if err != nil {
		return nil, err
	}

	e.UpdatedAt = updatedAt
	e.CommandSecretHash = commandSecretHash.String
	e.OrganizationID = organizationID.String
	e.ConfirmedAt = nullInt64ToPtr(confirmedAt)

	// DB schema has single reviewed_at column - use ApprovedAt for approved status.
	// RejectedAt remains nil since there's no separate column.
	switch e.Status {
	case inbox.StatusApproved:
		e.ApprovedAt = nullInt64ToPtr(reviewedAt)
		e.Notes = reviewedReason.String
	case inbox.StatusRejected:
		e.RejectedAt = nullInt64ToPtr(reviewedAt)
		e.Notes = rejectionReason.String
	case inbox.StatusPending, inbox.StatusAcknowledged, inbox.StatusApproving, inbox.StatusExpired:
		// No action needed for non-reviewed statuses.
	}

	return &e, nil
}

// nullInt64ToPtr converts sql.NullInt64 to *int64.
func nullInt64ToPtr(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}
