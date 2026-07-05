package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// UpdatesStorage implements the updates repository interface.
type UpdatesStorage struct {
	db *sql.DB
}

// NewUpdatesStorage creates a new UpdatesStorage.
func NewUpdatesStorage(db *sql.DB) *UpdatesStorage {
	return &UpdatesStorage{db: db}
}

// Ensure UpdatesStorage implements updates.Repository.
var _ updates.Repository = (*UpdatesStorage)(nil)

// Version operations

func (s *UpdatesStorage) CreateVersion(ctx context.Context, v *updates.UpdateVersion) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO update_versions (id, version, apk_filename, apk_size, sha256, release_date, release_notes, release_type, is_latest, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, v.ID, v.Version, v.APKFilename, v.APKSize, v.SHA256, v.ReleaseDate, v.ReleaseNotes, v.ReleaseType, v.IsLatest, v.CreatedAt, v.UpdatedAt)
	return err
}

func (s *UpdatesStorage) GetVersionByID(ctx context.Context, id string) (*updates.UpdateVersion, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, version, apk_filename, apk_size, sha256, release_date, release_notes, release_type, is_latest, created_at, updated_at
		FROM update_versions WHERE id = ?
	`, id)
	return s.scanVersion(row)
}

func (s *UpdatesStorage) GetVersionByVersion(ctx context.Context, version string) (*updates.UpdateVersion, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, version, apk_filename, apk_size, sha256, release_date, release_notes, release_type, is_latest, created_at, updated_at
		FROM update_versions WHERE version = ?
	`, version)
	return s.scanVersion(row)
}

func (s *UpdatesStorage) GetLatestVersion(ctx context.Context) (*updates.UpdateVersion, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, version, apk_filename, apk_size, sha256, release_date, release_notes, release_type, is_latest, created_at, updated_at
		FROM update_versions WHERE is_latest = 1 LIMIT 1
	`)
	v, err := s.scanVersion(row)
	if err == sql.ErrNoRows {
		return nil, updates.ErrVersionNotFound
	}
	return v, err
}

func (s *UpdatesStorage) ListVersions(ctx context.Context, status string, limit, offset int) ([]*updates.UpdateVersion, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var whereClause string
	args := make([]interface{}, 0, 2)

	switch status {
	case "latest":
		whereClause = "WHERE is_latest = 1"
	case "previous":
		whereClause = "WHERE is_latest = 0"
	default:
		whereClause = ""
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM update_versions %s", whereClause)
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get versions
	query := fmt.Sprintf(`
		SELECT id, version, apk_filename, apk_size, sha256, release_date, release_notes, release_type, is_latest, created_at, updated_at
		FROM update_versions
	%s
		ORDER BY release_date DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var versions []*updates.UpdateVersion
	for rows.Next() {
		v, err := s.scanVersionRows(rows)
		if err != nil {
			return nil, 0, err
		}
		versions = append(versions, v)
	}

	return versions, total, rows.Err()
}

func (s *UpdatesStorage) UpdateLatestFlag(ctx context.Context, versionID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Clear all latest flags
	if _, err := tx.ExecContext(ctx, "UPDATE update_versions SET is_latest = 0"); err != nil {
		return err
	}

	// Set new latest
	if _, err := tx.ExecContext(ctx, "UPDATE update_versions SET is_latest = 1 WHERE id = ?", versionID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *UpdatesStorage) DeleteVersion(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM update_versions WHERE id = ?", id)
	return err
}

func (s *UpdatesStorage) UpdateVersion(ctx context.Context, v *updates.UpdateVersion) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE update_versions SET
			version = ?,
			apk_filename = ?,
			apk_size = ?,
			sha256 = ?,
			release_date = ?,
			release_notes = ?,
			release_type = ?,
			is_latest = ?,
			updated_at = ?
		WHERE id = ?
	`, v.Version, v.APKFilename, v.APKSize, v.SHA256, v.ReleaseDate, v.ReleaseNotes, v.ReleaseType, v.IsLatest, time.Now().UnixMilli(), v.ID)
	return err
}

func (s *UpdatesStorage) scanVersion(row *sql.Row) (*updates.UpdateVersion, error) {
	var v updates.UpdateVersion
	var releaseNotes sql.NullString
	err := row.Scan(
		&v.ID, &v.Version, &v.APKFilename, &v.APKSize, &v.SHA256,
		&v.ReleaseDate, &releaseNotes, &v.ReleaseType, &v.IsLatest,
		&v.CreatedAt, &v.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, updates.ErrVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	if releaseNotes.Valid {
		v.ReleaseNotes = releaseNotes.String
	}
	return &v, nil
}

func (s *UpdatesStorage) scanVersionRows(rows *sql.Rows) (*updates.UpdateVersion, error) {
	var v updates.UpdateVersion
	var releaseNotes sql.NullString
	err := rows.Scan(
		&v.ID, &v.Version, &v.APKFilename, &v.APKSize, &v.SHA256,
		&v.ReleaseDate, &releaseNotes, &v.ReleaseType, &v.IsLatest,
		&v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if releaseNotes.Valid {
		v.ReleaseNotes = releaseNotes.String
	}
	return &v, nil
}

// Push operations

func (s *UpdatesStorage) CreatePush(ctx context.Context, p *updates.UpdatePush) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO update_pushes (id, version_id, install_type, scheduled_at, status, initiated_by, initiated_at, completed_at, cancelled_at, cancelled_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, p.ID, p.VersionID, p.InstallType, p.ScheduledAt, p.Status, p.InitiatedBy, p.InitiatedAt, p.CompletedAt, p.CancelledAt, p.CancelledBy)
	return err
}

func (s *UpdatesStorage) GetPushByID(ctx context.Context, id string) (*updates.UpdatePush, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, version_id, install_type, scheduled_at, status, initiated_by, initiated_at, completed_at, cancelled_at, cancelled_by
		FROM update_pushes WHERE id = ?
	`, id)
	return s.scanPush(row)
}

func (s *UpdatesStorage) GetPushByIDWithVersion(ctx context.Context, id string) (*updates.UpdatePush, *updates.UpdateVersion, error) {
	push, err := s.GetPushByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	version, err := s.GetVersionByID(ctx, push.VersionID)
	if err != nil {
		return push, nil, err
	}
	return push, version, nil
}

func (s *UpdatesStorage) UpdatePushStatus(ctx context.Context, id string, status updates.UpdateStatus) error {
	_, err := s.db.ExecContext(ctx, "UPDATE update_pushes SET status = ? WHERE id = ?", status, id)
	return err
}

func (s *UpdatesStorage) CompletePush(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	_, err := s.db.ExecContext(ctx, "UPDATE update_pushes SET status = ?, completed_at = ? WHERE id = ?",
		updates.UpdateStatusCompleted, now, id)
	return err
}

func (s *UpdatesStorage) CancelPush(ctx context.Context, id, cancelledBy string) error {
	now := time.Now().UnixMilli()
	_, err := s.db.ExecContext(ctx, "UPDATE update_pushes SET status = ?, cancelled_at = ?, cancelled_by = ? WHERE id = ?",
		updates.UpdateStatusCancelled, now, cancelledBy, id)
	return err
}

func (s *UpdatesStorage) ListPushes(ctx context.Context, status string, limit, offset int) ([]*updates.UpdatePush, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var whereClause string
	args := make([]interface{}, 0, 2)

	if status != "" && status != "all" {
		whereClause = "WHERE status = ?"
		args = append(args, status)
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM update_pushes %s", whereClause)
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get pushes
	query := fmt.Sprintf(`
		SELECT id, version_id, install_type, scheduled_at, status, initiated_by, initiated_at, completed_at, cancelled_at, cancelled_by
		FROM update_pushes
	%s
		ORDER BY initiated_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var pushes []*updates.UpdatePush
	for rows.Next() {
		p, err := s.scanPushRows(rows)
		if err != nil {
			return nil, 0, err
		}
		pushes = append(pushes, p)
	}

	return pushes, total, rows.Err()
}

func (s *UpdatesStorage) scanPush(row *sql.Row) (*updates.UpdatePush, error) {
	var p updates.UpdatePush
	var scheduledAt, completedAt, cancelledAt sql.NullInt64
	var cancelledBy sql.NullString
	err := row.Scan(
		&p.ID, &p.VersionID, &p.InstallType, &scheduledAt, &p.Status,
		&p.InitiatedBy, &p.InitiatedAt, &completedAt, &cancelledAt, &cancelledBy,
	)
	if err == sql.ErrNoRows {
		return nil, updates.ErrPushNotFound
	}
	if err != nil {
		return nil, err
	}
	if scheduledAt.Valid {
		p.ScheduledAt = &scheduledAt.Int64
	}
	if completedAt.Valid {
		p.CompletedAt = &completedAt.Int64
	}
	if cancelledAt.Valid {
		p.CancelledAt = &cancelledAt.Int64
	}
	if cancelledBy.Valid {
		p.CancelledBy = cancelledBy.String
	}
	return &p, nil
}

func (s *UpdatesStorage) scanPushRows(rows *sql.Rows) (*updates.UpdatePush, error) {
	var p updates.UpdatePush
	var scheduledAt, completedAt, cancelledAt sql.NullInt64
	var cancelledBy sql.NullString
	err := rows.Scan(
		&p.ID, &p.VersionID, &p.InstallType, &scheduledAt, &p.Status,
		&p.InitiatedBy, &p.InitiatedAt, &completedAt, &cancelledAt, &cancelledBy,
	)
	if err != nil {
		return nil, err
	}
	if scheduledAt.Valid {
		p.ScheduledAt = &scheduledAt.Int64
	}
	if completedAt.Valid {
		p.CompletedAt = &completedAt.Int64
	}
	if cancelledAt.Valid {
		p.CancelledAt = &cancelledAt.Int64
	}
	if cancelledBy.Valid {
		p.CancelledBy = cancelledBy.String
	}
	return &p, nil
}

// Push device operations

func (s *UpdatesStorage) CreatePushDevice(ctx context.Context, d *updates.UpdatePushDevice) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO update_push_devices (id, push_id, device_id, status, sent_at, acknowledged_at, error, retry_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.PushID, d.DeviceID, d.Status, d.SentAt, d.AcknowledgedAt, d.Error, d.RetryCount, d.CreatedAt, d.UpdatedAt)
	return err
}

func (s *UpdatesStorage) GetPushDevices(ctx context.Context, pushID string) ([]*updates.UpdatePushDevice, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, push_id, device_id, status, sent_at, acknowledged_at, error, retry_count, created_at, updated_at
		FROM update_push_devices WHERE push_id = ?
		ORDER BY created_at ASC
	`, pushID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var devices []*updates.UpdatePushDevice
	for rows.Next() {
		d, err := s.scanPushDeviceRows(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (s *UpdatesStorage) UpdatePushDeviceStatus(ctx context.Context, id string, status updates.DevicePushStatus, errorMsg string) error {
	now := time.Now().UnixMilli()
	var sentAt, ackAt *int64
	if status == updates.DevicePushStatusSent {
		sentAt = &now
	}
	if status == updates.DevicePushStatusAcknowledged || status == updates.DevicePushStatusInProgress {
		ackAt = &now
	}
	if status == updates.DevicePushStatusCompleted {
		ackAt = &now
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE update_push_devices SET status = ?, sent_at = COALESCE(?, sent_at), acknowledged_at = COALESCE(?, acknowledged_at), error = ?, updated_at = ?
		WHERE id = ?
	`, status, sentAt, ackAt, errorMsg, now, id)
	return err
}

func (s *UpdatesStorage) CountPushDevicesByStatus(ctx context.Context, pushID string, status updates.DevicePushStatus) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM update_push_devices WHERE push_id = ? AND status = ?
	`, pushID, status).Scan(&count)
	return count, err
}

func (s *UpdatesStorage) GetPushDeviceByPushAndDevice(ctx context.Context, pushID, deviceID string) (*updates.UpdatePushDevice, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, push_id, device_id, status, sent_at, acknowledged_at, error, retry_count, created_at, updated_at
		FROM update_push_devices WHERE push_id = ? AND device_id = ?
	`, pushID, deviceID)

	var d updates.UpdatePushDevice
	var sentAt, ackAt sql.NullInt64
	var errMsg sql.NullString
	err := row.Scan(&d.ID, &d.PushID, &d.DeviceID, &d.Status, &sentAt, &ackAt, &errMsg, &d.RetryCount, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, updates.ErrPushNotFound
		}
		return nil, err
	}
	if sentAt.Valid {
		d.SentAt = &sentAt.Int64
	}
	if ackAt.Valid {
		d.AcknowledgedAt = &ackAt.Int64
	}
	if errMsg.Valid {
		d.Error = errMsg.String
	}
	return &d, nil
}

func (s *UpdatesStorage) UpdatePushDeviceStatusByDispatch(ctx context.Context, dispatchID, deviceID string, status updates.DevicePushStatus, errorMsg string) error {
	// Find the push device by looking up the command with the given dispatch_id
	// The dispatch_id for update push is the push_id itself
	// We need to find the device's push record by device_id and then update by id
	var devicePushID string
	err := s.db.QueryRowContext(ctx, `
		SELECT upd.id FROM update_push_devices upd
		JOIN update_pushes up ON upd.push_id = up.id
		WHERE up.id = ? AND upd.device_id = ?
	`, dispatchID, deviceID).Scan(&devicePushID)
	if err != nil {
		if err == sql.ErrNoRows {
			return updates.ErrPushNotFound
		}
		return err
	}
	return s.UpdatePushDeviceStatus(ctx, devicePushID, status, errorMsg)
}

func (s *UpdatesStorage) scanPushDeviceRows(rows *sql.Rows) (*updates.UpdatePushDevice, error) {
	var d updates.UpdatePushDevice
	var sentAt, ackAt sql.NullInt64
	var errMsg sql.NullString
	err := rows.Scan(
		&d.ID, &d.PushID, &d.DeviceID, &d.Status, &sentAt, &ackAt, &errMsg, &d.RetryCount, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if sentAt.Valid {
		d.SentAt = &sentAt.Int64
	}
	if ackAt.Valid {
		d.AcknowledgedAt = &ackAt.Int64
	}
	if errMsg.Valid {
		d.Error = errMsg.String
	}
	return &d, nil
}

// Sync state operations

func (s *UpdatesStorage) GetSyncState(ctx context.Context) (*updates.SyncState, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT status, last_sync_at, next_sync_at, last_error, versions_found
		FROM update_sync_status WHERE id = 1
	`)
	var state updates.SyncState
	var lastSyncAt, nextSyncAt sql.NullInt64
	var lastError sql.NullString
	var versionsFound int
	err := row.Scan(
		&state.Status, &lastSyncAt, &nextSyncAt, &lastError, &versionsFound,
	)
	if err == sql.ErrNoRows {
		return &updates.SyncState{
			Status: updates.SyncStatusIdle,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	if lastSyncAt.Valid {
		state.LastSyncAt = &lastSyncAt.Int64
	}
	if nextSyncAt.Valid {
		state.NextSyncAt = &nextSyncAt.Int64
	}
	if lastError.Valid {
		state.Error = lastError.String
	}
	state.VersionsFound = versionsFound
	return &state, nil
}

func (s *UpdatesStorage) UpdateSyncState(ctx context.Context, state *updates.SyncState) error {
	now := time.Now().UnixMilli()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO update_sync_status (id, status, last_sync_at, next_sync_at, last_error, versions_found, created_at, updated_at)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status = excluded.status,
			last_sync_at = excluded.last_sync_at,
			next_sync_at = excluded.next_sync_at,
			last_error = excluded.last_error,
			versions_found = excluded.versions_found,
			updated_at = excluded.updated_at
	`, state.Status, state.LastSyncAt, state.NextSyncAt, state.Error, state.VersionsFound, now, now)
	return err
}

// TryAcquireSyncLock attempts to atomically acquire the sync lock by updating status to syncing.
// Returns (acquired, currentState, error). If acquired is false, caller should not proceed.
func (s *UpdatesStorage) TryAcquireSyncLock(ctx context.Context) (bool, *updates.SyncState, error) {
	now := time.Now().UnixMilli()

	// Use a transaction to atomically check and update
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Get current state with row lock
	row := tx.QueryRowContext(ctx, `
		SELECT status, last_sync_at, next_sync_at, last_error, versions_found
		FROM update_sync_status WHERE id = 1
	`)
	var state updates.SyncState
	var lastSyncAt, nextSyncAt sql.NullInt64
	var lastError sql.NullString
	var versionsFound int
	err = row.Scan(&state.Status, &lastSyncAt, &nextSyncAt, &lastError, &versionsFound)
	if err == sql.ErrNoRows {
		// No state exists - try to create with syncing status
		_, insertErr := tx.ExecContext(ctx, `
			INSERT INTO update_sync_status (id, status, created_at, updated_at)
			VALUES (1, ?, ?, ?)
		`, updates.SyncStatusSyncing, now, now)
		if insertErr != nil {
			return false, nil, insertErr
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return false, nil, commitErr
		}
		return true, &updates.SyncState{Status: updates.SyncStatusSyncing}, nil
	}
	if err != nil {
		return false, nil, err
	}

	// Check if already syncing
	if state.Status == updates.SyncStatusSyncing {
		// Still syncing, don't acquire
		if lastSyncAt.Valid {
			state.LastSyncAt = &lastSyncAt.Int64
		}
		if nextSyncAt.Valid {
			state.NextSyncAt = &nextSyncAt.Int64
		}
		if lastError.Valid {
			state.Error = lastError.String
		}
		state.VersionsFound = versionsFound
		return false, &state, nil
	}

	// Try to update to syncing
	_, updateErr := tx.ExecContext(ctx, `
		UPDATE update_sync_status SET status = ?, last_sync_at = ?, last_error = '', updated_at = ?
		WHERE id = 1 AND status != ?
	`, updates.SyncStatusSyncing, now, now, updates.SyncStatusSyncing)
	if updateErr != nil {
		return false, nil, updateErr
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return false, nil, commitErr
	}

	state.Status = updates.SyncStatusSyncing
	state.LastSyncAt = &now
	return true, &state, nil
}

// ExtractVersionStatus returns "latest", "previous", or "all" based on is_latest flag.
func ExtractVersionStatus(versions []*updates.UpdateVersion) string {
	if len(versions) == 0 {
		return "all"
	}
	for _, v := range versions {
		if v.IsLatest {
			return "latest"
		}
	}
	return "previous"
}

// ParseReleaseType parses a release type string.
func ParseReleaseType(s string) updates.ReleaseType {
	switch strings.ToLower(s) {
	case "major":
		return updates.ReleaseTypeMajor
	case "minor":
		return updates.ReleaseTypeMinor
	case "patch":
		return updates.ReleaseTypePatch
	default:
		return updates.ReleaseTypePatch
	}
}
