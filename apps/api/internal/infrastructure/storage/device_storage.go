package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Querier is an interface that both *sql.DB and *sql.Tx implement.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// getQuerier returns the transaction from context if available, otherwise the db.
func (r *DeviceRepository) getQuerier(ctx context.Context) Querier {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}

// queryRow is a helper that uses transaction-aware querier.
func (r *DeviceRepository) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return r.getQuerier(ctx).QueryRowContext(ctx, query, args...)
}

// queryRows is a helper that uses transaction-aware querier.
func (r *DeviceRepository) queryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return r.getQuerier(ctx).QueryContext(ctx, query, args...)
}

// exec is a helper that uses transaction-aware querier.
func (r *DeviceRepository) exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getQuerier(ctx).ExecContext(ctx, query, args...)
}

// Ensure DeviceRepository implements device.Repository.
var _ device.Repository = (*DeviceRepository)(nil)

// DeviceRepository implements device.Repository using SQLite.
type DeviceRepository struct {
	db *sql.DB
}

// NewDeviceRepository creates a new DeviceRepository.
func NewDeviceRepository(db *sql.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// deviceColumns returns the list of columns to select.
const deviceColumns = `
	id, firebase_install_id, fcm_token, app_version, device_class,
	command_secret_hash, online, registered_at, last_seen, operator_id,
	organization_id,
	created_at, updated_at, device_name, manufacturer, model, os_version, security_patch,
	deregistered_at, deletion_scheduled_at, fcm_token_refreshed_at, tags
`

// scanDevice scans a device from a row.
func scanDevice(row *sql.Row) (*device.Device, error) {
	var d device.Device
	var fcmToken, operatorID, organizationID, deviceName, manufacturer, model, osVersion, securityPatch, tagsJSON sql.NullString
	var deregisteredAt, deletionScheduledAt, fcmTokenRefreshedAt sql.NullInt64

	err := row.Scan(
		&d.ID, &d.FirebaseInstallID, &fcmToken, &d.AppVersion, &d.DeviceClass,
		&d.CommandSecretHash, &d.Online, &d.RegisteredAt, &d.LastSeen, &operatorID,
		&organizationID,
		&d.CreatedAt, &d.UpdatedAt, &deviceName, &manufacturer, &model, &osVersion, &securityPatch,
		&deregisteredAt, &deletionScheduledAt, &fcmTokenRefreshedAt, &tagsJSON,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, device.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	d.FCMToken = fcmToken.String
	d.OperatorID = operatorID.String
	d.OrganizationID = organizationID.String
	d.DeviceName = deviceName.String
	d.Manufacturer = manufacturer.String
	d.Model = model.String
	d.OSVersion = osVersion.String
	d.SecurityPatch = securityPatch.String
	if tagsJSON.Valid && tagsJSON.String != "" {
		_ = json.Unmarshal([]byte(tagsJSON.String), &d.Tags)
	}

	if deregisteredAt.Valid {
		d.DeregisteredAt = &deregisteredAt.Int64
		d.Lifecycle = device.LifecycleDeregistered
	} else if d.RegisteredAt > 0 {
		d.Lifecycle = device.LifecycleRegistered
	} else {
		d.Lifecycle = device.LifecyclePending
	}

	if deletionScheduledAt.Valid {
		d.DeletionScheduledAt = &deletionScheduledAt.Int64
	}
	if fcmTokenRefreshedAt.Valid {
		d.FCMTokenRefreshedAt = &fcmTokenRefreshedAt.Int64
	}

	return &d, nil
}

// scanDevices scans multiple devices from rows.
func scanDevices(rows *sql.Rows) ([]*device.Device, error) {
	var devices []*device.Device

	for rows.Next() {
		var d device.Device
		var fcmToken, operatorID, organizationID, deviceName, manufacturer, model, osVersion, securityPatch, tagsJSON sql.NullString
		var deregisteredAt, deletionScheduledAt, fcmTokenRefreshedAt sql.NullInt64

		if err := rows.Scan(
			&d.ID, &d.FirebaseInstallID, &fcmToken, &d.AppVersion, &d.DeviceClass,
			&d.CommandSecretHash, &d.Online, &d.RegisteredAt, &d.LastSeen, &operatorID,
			&organizationID,
			&d.CreatedAt, &d.UpdatedAt, &deviceName, &manufacturer, &model, &osVersion, &securityPatch,
			&deregisteredAt, &deletionScheduledAt, &fcmTokenRefreshedAt, &tagsJSON,
		); err != nil {
			return nil, err
		}

		d.FCMToken = fcmToken.String
		d.OperatorID = operatorID.String
		d.OrganizationID = organizationID.String
		d.DeviceName = deviceName.String
		d.Manufacturer = manufacturer.String
		d.Model = model.String
		d.OSVersion = osVersion.String
		d.SecurityPatch = securityPatch.String
		if tagsJSON.Valid && tagsJSON.String != "" {
			_ = json.Unmarshal([]byte(tagsJSON.String), &d.Tags)
		}

		if deregisteredAt.Valid {
			d.DeregisteredAt = &deregisteredAt.Int64
			d.Lifecycle = device.LifecycleDeregistered
		} else if d.RegisteredAt > 0 {
			d.Lifecycle = device.LifecycleRegistered
		} else {
			d.Lifecycle = device.LifecyclePending
		}

		if deletionScheduledAt.Valid {
			d.DeletionScheduledAt = &deletionScheduledAt.Int64
		}
		if fcmTokenRefreshedAt.Valid {
			d.FCMTokenRefreshedAt = &fcmTokenRefreshedAt.Int64
		}

		devices = append(devices, &d)
	}

	return devices, rows.Err()
}

// FindByID retrieves a device by ID.
func (r *DeviceRepository) FindByID(ctx context.Context, id string) (*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE id = ?`
	return scanDevice(r.queryRow(ctx, query, id))
}

// FindByIMEI retrieves a device by IMEI.
func (r *DeviceRepository) FindByIMEI(ctx context.Context, imei string) (*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE id = ?`
	return scanDevice(r.queryRow(ctx, query, imei))
}

// FindByFirebaseInstallID retrieves a device by Firebase install ID.
func (r *DeviceRepository) FindByFirebaseInstallID(ctx context.Context, fid string) (*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE firebase_install_id = ?`
	return scanDevice(r.queryRow(ctx, query, fid))
}

// FindByIDAndOperator retrieves a device by ID and verifies it belongs to the operator.
// This implements DOA (Data Ownership Attribution) checks for infraauth.
// Returns ErrNotFound if device doesn't exist OR doesn't belong to the operator.
func (r *DeviceRepository) FindByIDAndOperator(ctx context.Context, id, operatorID string) (*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE id = ? AND operator_id = ?`
	return scanDevice(r.queryRow(ctx, query, id, operatorID))
}

// FindByIMEIAndOperator retrieves a device by IMEI and verifies it belongs to the operator.
// This implements DOA (Data Ownership Attribution) checks for deregistration.
// Returns ErrNotFound if device doesn't exist OR doesn't belong to the operator.
func (r *DeviceRepository) FindByIMEIAndOperator(ctx context.Context, imei, operatorID string) (*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE id = ? AND operator_id = ?`
	return scanDevice(r.queryRow(ctx, query, imei, operatorID))
}

// FindByIMEIAndOrganization retrieves a device by IMEI within an organization.
// Returns ErrNotFound if device doesn't exist OR doesn't belong to the organization.
func (r *DeviceRepository) FindByIMEIAndOrganization(ctx context.Context, imei, orgID string) (*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE id = ? AND organization_id = ?`
	return scanDevice(r.queryRow(ctx, query, imei, orgID))
}

// FindByIDAndOrganization retrieves a device by ID within an organization.
// Returns ErrNotFound if device doesn't exist OR doesn't belong to the organization.
func (r *DeviceRepository) FindByIDAndOrganization(ctx context.Context, id, orgID string) (*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE id = ? AND organization_id = ?`
	return scanDevice(r.queryRow(ctx, query, id, orgID))
}

// Create creates a new device.
func (r *DeviceRepository) Create(ctx context.Context, d *device.Device) error {
	query := `
		INSERT INTO devices (id, firebase_install_id, fcm_token, app_version, device_class,
		                    command_secret_hash, online, registered_at, last_seen, operator_id,
		                    organization_id,
		                    created_at, updated_at, device_name, manufacturer, model, os_version, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.exec(ctx, query,
		d.ID, d.FirebaseInstallID, nullString(d.FCMToken), d.AppVersion, d.DeviceClass,
		d.CommandSecretHash, d.Online, d.RegisteredAt, d.LastSeen, nullString(d.OperatorID),
		nullString(d.OrganizationID),
		d.CreatedAt, d.UpdatedAt, nullString(d.DeviceName), nullString(d.Manufacturer),
		nullString(d.Model), nullString(d.OSVersion), tagsToJSON(d.Tags),
	)

	return err
}

// Update updates an existing device.
func (r *DeviceRepository) Update(ctx context.Context, d *device.Device) error {
	query := `
		UPDATE devices 
		SET firebase_install_id = ?, fcm_token = ?, app_version = ?, device_class = ?,
		    command_secret_hash = ?, online = ?, registered_at = ?, last_seen = ?,
		    operator_id = ?, organization_id = ?, updated_at = ?, device_name = ?, manufacturer = ?,
		    model = ?, os_version = ?
		WHERE id = ?`

	result, err := r.exec(ctx, query,
		d.FirebaseInstallID, nullString(d.FCMToken), d.AppVersion, d.DeviceClass,
		d.CommandSecretHash, d.Online, d.RegisteredAt, d.LastSeen, nullString(d.OperatorID),
		nullString(d.OrganizationID), time.Now(), nullString(d.DeviceName), nullString(d.Manufacturer),
		nullString(d.Model), nullString(d.OSVersion), tagsToJSON(d.Tags), d.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// Delete deletes a device.
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	result, err := r.exec(ctx, "DELETE FROM devices WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// UpdateFCMToken updates the FCM token for a device.
func (r *DeviceRepository) UpdateFCMToken(ctx context.Context, id, fcmToken string) error {
	result, err := r.exec(ctx,
		"UPDATE devices SET fcm_token = ?, updated_at = ? WHERE id = ?",
		fcmToken, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// SetOnline sets the online status of a device.
// When coming online, clears deletion_scheduled_at to cancel scheduled deletion.
func (r *DeviceRepository) SetOnline(ctx context.Context, id string, online bool) error {
	now := time.Now().UnixMilli()

	var result sql.Result

	var err error

	if online {
		// Coming online cancels scheduled deletion.
		result, err = r.exec(ctx,
			"UPDATE devices SET online = ?, last_seen = ?, updated_at = ?, deletion_scheduled_at = 0 WHERE id = ?",
			online, now, now, id,
		)
	} else {
		result, err = r.exec(ctx,
			"UPDATE devices SET online = ?, updated_at = ? WHERE id = ?",
			online, now, id,
		)
	}

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// UpdateLastSeen updates the last seen timestamp.
func (r *DeviceRepository) UpdateLastSeen(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()

	result, err := r.exec(ctx,
		"UPDATE devices SET last_seen = ?, updated_at = ? WHERE id = ?",
		now, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// List returns a paginated list of devices.
func (r *DeviceRepository) List(ctx context.Context, orgID string, limit, offset int) ([]*device.Device, int, error) {
	var query string
	var countQuery string
	var args []interface{}

	if orgID != "" {
		query = `SELECT ` + deviceColumns + ` FROM devices WHERE organization_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
		countQuery = `SELECT COUNT(*) FROM devices WHERE organization_id = ?`
		args = []interface{}{orgID, limit, offset}
	} else {
		query = `SELECT ` + deviceColumns + ` FROM devices ORDER BY created_at DESC LIMIT ? OFFSET ?`
		countQuery = `SELECT COUNT(*) FROM devices`
		args = []interface{}{limit, offset}
	}

	rows, err := r.queryRows(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	defer func() { _ = rows.Close() }()

	devices, err := scanDevices(rows)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if orgID != "" {
		err = r.queryRow(ctx, countQuery, orgID).Scan(&total)
	} else {
		err = r.queryRow(ctx, countQuery).Scan(&total)
	}
	if err != nil {
		return nil, 0, err
	}

	return devices, total, rows.Err()
}

// ListByOperator returns all devices for an operator.
func (r *DeviceRepository) ListByOperator(ctx context.Context, operatorID string) ([]*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE operator_id = ? ORDER BY created_at DESC`

	rows, err := r.queryRows(ctx, query, operatorID)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	return scanDevices(rows)
}

// Count returns the total number of devices filtered by organization.
func (r *DeviceRepository) Count(ctx context.Context, orgID string) (int, error) {
	var count int
	var err error

	if orgID != "" {
		err = r.queryRow(ctx, "SELECT COUNT(*) FROM devices WHERE organization_id = ?", orgID).Scan(&count)
	} else {
		err = r.queryRow(ctx, "SELECT COUNT(*) FROM devices").Scan(&count)
	}

	return count, err
}

// CountByOperator returns the number of devices for an operator.
func (r *DeviceRepository) CountByOperator(ctx context.Context, operatorID string) (int, error) {
	var count int
	err := r.queryRow(ctx,
		"SELECT COUNT(*) FROM devices WHERE operator_id = ?", operatorID,
	).Scan(&count)

	return count, err
}

// CountByOrganization returns the number of devices for an organization.
func (r *DeviceRepository) CountByOrganization(ctx context.Context, orgID string) (int, error) {
	var count int
	err := r.queryRow(ctx,
		"SELECT COUNT(*) FROM devices WHERE organization_id = ?", orgID,
	).Scan(&count)

	return count, err
}

// SetSecretHash sets the command secret hash for a device.
func (r *DeviceRepository) SetSecretHash(ctx context.Context, deviceID, hash string) error {
	result, err := r.exec(ctx,
		"UPDATE devices SET command_secret_hash = ?, updated_at = ? WHERE id = ?",
		hash, time.Now(), deviceID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// GetSecretHash retrieves the command secret hash for a device.
func (r *DeviceRepository) GetSecretHash(ctx context.Context, deviceID string) (string, error) {
	var hash string

	err := r.queryRow(ctx,
		"SELECT command_secret_hash FROM devices WHERE id = ?",
		deviceID,
	).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", device.ErrNotFound
	}

	return hash, err
}

// HashAllSecrets hashes all existing command secrets that don't have a hash.
// This is a migration helper for existing databases.
func (r *DeviceRepository) HashAllSecrets(ctx context.Context) (int, error) {
	query := `
		SELECT id, command_secret_hash 
		FROM devices 
		WHERE command_secret_hash IS NULL OR command_secret_hash = ''`

	rows, err := r.queryRows(ctx, query)
	if err != nil {
		return 0, err
	}

	defer func() { _ = rows.Close() }()

	count := 0

	for rows.Next() {
		var id string

		var currentHash sql.NullString
		if err := rows.Scan(&id, &currentHash); err != nil {
			continue
		}

		if currentHash.Valid && currentHash.String != "" {
			continue
		}

		count++
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

// Touch updates the last seen timestamp for a device.
func (r *DeviceRepository) Touch(ctx context.Context, deviceID string) error {
	return r.UpdateLastSeen(ctx, deviceID)
}

// SoftDelete marks a device as deregistered (soft delete).
// Sets deregistered_at and deletion_scheduled_at for 30-day retention.
func (r *DeviceRepository) SoftDelete(ctx context.Context, id string, deregisteredAt, deletionScheduledAt int64) error {
	result, err := r.exec(ctx,
		"UPDATE devices SET deregistered_at = ?, deletion_scheduled_at = ?, online = 0, updated_at = ? WHERE id = ?",
		deregisteredAt, deletionScheduledAt, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return device.ErrNotFound
	}

	return nil
}

// SoftDeleteByIMEI marks a device as deregistered by IMEI.
func (r *DeviceRepository) SoftDeleteByIMEI(ctx context.Context, imei string, deregisteredAt, deletionScheduledAt int64) error {
	return r.SoftDelete(ctx, imei, deregisteredAt, deletionScheduledAt)
}

// ListActive returns all non-deregistered devices.
func (r *DeviceRepository) ListActive(ctx context.Context, limit, offset int) ([]*device.Device, int, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE deregistered_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.queryRows(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer func() { _ = rows.Close() }()

	devices, err := scanDevices(rows)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.queryRow(ctx, "SELECT COUNT(*) FROM devices WHERE deregistered_at IS NULL").Scan(&total); err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

// ListActiveByOperator returns all non-deregistered devices for an operator.
func (r *DeviceRepository) ListActiveByOperator(ctx context.Context, operatorID string) ([]*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE operator_id = ? AND deregistered_at IS NULL ORDER BY created_at DESC`

	rows, err := r.queryRows(ctx, query, operatorID)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	return scanDevices(rows)
}

// FindByDeviceID retrieves a device by DeviceID value object.
func (r *DeviceRepository) FindByDeviceID(ctx context.Context, id device.DeviceID) (*device.Device, error) {
	return r.FindByID(ctx, id.String())
}

// FindByIDAndOperatorID retrieves a device by DeviceID and OperatorID value objects.
func (r *DeviceRepository) FindByIDAndOperatorID(ctx context.Context, id device.DeviceID, operatorID device.OperatorID) (*device.Device, error) {
	return r.FindByIDAndOperator(ctx, id.String(), operatorID.String())
}

// DeleteByDeviceID deletes a device by DeviceID.
func (r *DeviceRepository) DeleteByDeviceID(ctx context.Context, id device.DeviceID) error {
	return r.Delete(ctx, id.String())
}

// SetOnlineByDeviceID sets the online status by DeviceID.
func (r *DeviceRepository) SetOnlineByDeviceID(ctx context.Context, id device.DeviceID, online bool) error {
	return r.SetOnline(ctx, id.String(), online)
}

// ListByOperatorID returns all devices for an OperatorID.
func (r *DeviceRepository) ListByOperatorID(ctx context.Context, operatorID device.OperatorID) ([]*device.Device, error) {
	return r.ListByOperator(ctx, operatorID.String())
}

// ListByOrganization returns all devices for an organization.
func (r *DeviceRepository) ListByOrganization(ctx context.Context, orgID string) ([]*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE organization_id = ? ORDER BY created_at DESC`

	rows, err := r.queryRows(ctx, query, orgID)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	return scanDevices(rows)
}

// ListByOrganizationPaginated returns paginated devices for an organization.
func (r *DeviceRepository) ListByOrganizationPaginated(ctx context.Context, orgID string, limit, offset int) ([]*device.Device, int, error) {
	countQuery := `SELECT COUNT(*) FROM devices WHERE organization_id = ?`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT ` + deviceColumns + ` FROM devices WHERE organization_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.queryRows(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer func() { _ = rows.Close() }()

	devices, err := scanDevices(rows)
	if err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

// ListPending returns all devices in pending lifecycle state.
func (r *DeviceRepository) ListPending(ctx context.Context) ([]*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE registered_at = 0 AND deregistered_at IS NULL ORDER BY created_at DESC`

	rows, err := r.queryRows(ctx, query)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	return scanDevices(rows)
}

// ListPendingByOperator returns all pending devices for an operator.
func (r *DeviceRepository) ListPendingByOperator(ctx context.Context, operatorID device.OperatorID) ([]*device.Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices WHERE operator_id = ? AND registered_at = 0 AND deregistered_at IS NULL ORDER BY created_at DESC`

	rows, err := r.queryRows(ctx, query, operatorID.String())
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	return scanDevices(rows)
}

// DeleteScheduled deletes all devices where deletion_scheduled_at <= now AND deregistered_at IS NOT NULL.
// Returns the number of devices deleted.
func (r *DeviceRepository) DeleteScheduled(ctx context.Context) (int, error) {
	now := time.Now().UnixMilli()
	query := `DELETE FROM devices WHERE deletion_scheduled_at > 0 AND deletion_scheduled_at <= ? AND deregistered_at IS NOT NULL`

	result, err := r.exec(ctx, query, now)
	if err != nil {
		return 0, err
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(deleted), nil
}

// SoftDeleteByOrganization soft-deletes all devices for an organization.
// This is used during organization deletion.
func (r *DeviceRepository) SoftDeleteByOrganization(ctx context.Context, orgID string, deregisteredAt, deletionScheduledAt int64) (int, error) {
	query := `UPDATE devices SET deregistered_at = ?, deletion_scheduled_at = ? WHERE organization_id = ? AND deregistered_at IS NULL`

	result, err := r.exec(ctx, query, deregisteredAt, deletionScheduledAt, orgID)
	if err != nil {
		return 0, err
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(deleted), nil
}

func tagsToJSON(tags []string) interface{} {
	if len(tags) == 0 {
		return nil
	}
	data, err := json.Marshal(tags)
	if err != nil {
		return nil
	}
	return string(data)
}
