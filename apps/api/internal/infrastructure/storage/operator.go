package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// Ensure OperatorRepository implements operator.Repository.
var _ operator.Repository = (*OperatorRepository)(nil)

// OperatorRepository implements operator.Repository using SQLite.
type OperatorRepository struct {
	db *sql.DB
}

// NewOperatorRepository creates a new OperatorRepository.
func NewOperatorRepository(db *sql.DB) *OperatorRepository {
	return &OperatorRepository{db: db}
}

// FindByID retrieves an operator by ID.
func (r *OperatorRepository) FindByID(ctx context.Context, id string) (*operator.Operator, error) {
	query := `
		SELECT id, email, name, password_hash, role, google_id, github_id, 
		       mfa_secret, mfa_enabled, mfa_backup_codes, email_verified, created_at, updated_at 
		FROM operators WHERE id = ?`

	var op operator.Operator
	var googleID, githubID, mfaSecret, mfaBackupCodes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&op.ID, &op.Email, &op.Name, &op.PasswordHash, &op.Role,
		&googleID, &githubID, &mfaSecret, &op.MFAEnabled, &mfaBackupCodes,
		&op.EmailVerified, &op.CreatedAt, &op.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, operator.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	op.GoogleID = googleID.String
	op.GitHubID = githubID.String
	op.MFASecret = mfaSecret.String

	// Parse backup codes from JSON
	if mfaBackupCodes.Valid && mfaBackupCodes.String != "" {
		if err := json.Unmarshal([]byte(mfaBackupCodes.String), &op.BackupCodes); err != nil {
			op.BackupCodes = []string{}
		}
	} else {
		op.BackupCodes = []string{}
	}

	return &op, nil
}

// FindByEmail retrieves an operator by email.
func (r *OperatorRepository) FindByEmail(ctx context.Context, email string) (*operator.Operator, error) {
	query := `
		SELECT id, email, name, password_hash, role, google_id, github_id, 
		       mfa_secret, mfa_enabled, email_verified, created_at, updated_at 
		FROM operators WHERE email = ?`

	var op operator.Operator
	var googleID, githubID, mfaSecret sql.NullString

	err := r.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(
		&op.ID, &op.Email, &op.Name, &op.PasswordHash, &op.Role,
		&googleID, &githubID, &mfaSecret, &op.MFAEnabled, &op.EmailVerified,
		&op.CreatedAt, &op.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, operator.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	op.GoogleID = googleID.String
	op.GitHubID = githubID.String
	op.MFASecret = mfaSecret.String

	return &op, nil
}

// FindByGoogleID retrieves an operator by Google ID.
func (r *OperatorRepository) FindByGoogleID(ctx context.Context, googleID string) (*operator.Operator, error) {
	query := `
		SELECT id, email, name, password_hash, role, google_id, github_id, 
		       mfa_secret, mfa_enabled, mfa_backup_codes, email_verified, created_at, updated_at 
		FROM operators WHERE google_id = ?`

	var op operator.Operator
	var googleIDVal, githubID, mfaSecret, mfaBackupCodes sql.NullString

	err := r.db.QueryRowContext(ctx, query, googleID).Scan(
		&op.ID, &op.Email, &op.Name, &op.PasswordHash, &op.Role,
		&googleIDVal, &githubID, &mfaSecret, &op.MFAEnabled, &mfaBackupCodes,
		&op.EmailVerified, &op.CreatedAt, &op.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, operator.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	op.GoogleID = googleIDVal.String
	op.GitHubID = githubID.String
	op.MFASecret = mfaSecret.String
	if mfaBackupCodes.Valid && mfaBackupCodes.String != "" {
		_ = json.Unmarshal([]byte(mfaBackupCodes.String), &op.BackupCodes)
	}

	return &op, nil
}

// FindByGitHubID retrieves an operator by GitHub ID.
func (r *OperatorRepository) FindByGitHubID(ctx context.Context, githubID string) (*operator.Operator, error) {
	query := `
		SELECT id, email, name, password_hash, role, google_id, github_id, 
		       mfa_secret, mfa_enabled, mfa_backup_codes, email_verified, created_at, updated_at 
		FROM operators WHERE github_id = ?`

	var op operator.Operator
	var googleID, githubIDVal, mfaSecret, mfaBackupCodes sql.NullString

	err := r.db.QueryRowContext(ctx, query, githubID).Scan(
		&op.ID, &op.Email, &op.Name, &op.PasswordHash, &op.Role,
		&googleID, &githubIDVal, &mfaSecret, &op.MFAEnabled, &mfaBackupCodes,
		&op.EmailVerified, &op.CreatedAt, &op.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, operator.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	op.GoogleID = googleID.String
	op.GitHubID = githubIDVal.String
	op.MFASecret = mfaSecret.String
	if mfaBackupCodes.Valid && mfaBackupCodes.String != "" {
		_ = json.Unmarshal([]byte(mfaBackupCodes.String), &op.BackupCodes)
	}

	return &op, nil
}

// Create creates a new operator.
func (r *OperatorRepository) Create(ctx context.Context, op *operator.Operator) error {
	query := `
		INSERT INTO operators (id, email, name, password_hash, role, google_id, github_id, 
		                       mfa_secret, mfa_enabled, email_verified, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		op.ID, strings.ToLower(op.Email), op.Name, op.PasswordHash, op.Role,
		nullString(op.GoogleID), nullString(op.GitHubID), nullString(op.MFASecret),
		op.MFAEnabled, op.EmailVerified, op.CreatedAt, op.UpdatedAt,
	)

	return err
}

// Update updates an existing operator.
func (r *OperatorRepository) Update(ctx context.Context, op *operator.Operator) error {
	query := `
		UPDATE operators 
		SET email = ?, name = ?, password_hash = ?, role = ?, google_id = ?, github_id = ?,
		    mfa_secret = ?, mfa_enabled = ?, email_verified = ?, updated_at = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		strings.ToLower(op.Email), op.Name, op.PasswordHash, op.Role,
		nullString(op.GoogleID), nullString(op.GitHubID), nullString(op.MFASecret),
		op.MFAEnabled, op.EmailVerified, time.Now(), op.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// Delete deletes an operator.
func (r *OperatorRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM operators WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// Count returns the total number of operators.
func (r *OperatorRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM operators").Scan(&count)
	return count, err
}

// List returns a paginated list of operators.
func (r *OperatorRepository) List(ctx context.Context, limit, offset int) ([]*operator.Operator, int, error) {
	query := `
		SELECT id, email, name, password_hash, role, google_id, github_id, 
		       mfa_secret, mfa_enabled, email_verified, created_at, updated_at 
		FROM operators ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var operators []*operator.Operator
	for rows.Next() {
		var op operator.Operator
		var googleID, githubID, mfaSecret sql.NullString

		if err := rows.Scan(
			&op.ID, &op.Email, &op.Name, &op.PasswordHash, &op.Role,
			&googleID, &githubID, &mfaSecret, &op.MFAEnabled, &op.EmailVerified,
			&op.CreatedAt, &op.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		op.GoogleID = googleID.String
		op.GitHubID = githubID.String
		op.MFASecret = mfaSecret.String
		operators = append(operators, &op)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM operators").Scan(&total); err != nil {
		return nil, 0, err
	}

	return operators, total, rows.Err()
}

// UpdatePassword updates the password hash for an operator.
func (r *OperatorRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE operators SET password_hash = ?, updated_at = ? WHERE id = ?",
		passwordHash, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// UpdateMFA updates MFA settings for an operator.
func (r *OperatorRepository) UpdateMFA(ctx context.Context, id, secret string, enabled bool) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE operators SET mfa_secret = ?, mfa_enabled = ?, updated_at = ? WHERE id = ?",
		secret, enabled, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// UpdateOperatorMFA updates the MFA secret and backup codes for an operator.
func (r *OperatorRepository) UpdateOperatorMFA(ctx context.Context, operatorID, mfaSecret string, backupCodes []string) error {
	backupCodesJSON := "[]"
	if len(backupCodes) > 0 {
		data, err := json.Marshal(backupCodes)
		if err != nil {
			return err
		}
		backupCodesJSON = string(data)
	}

	result, err := r.db.ExecContext(ctx,
		"UPDATE operators SET mfa_secret = ?, mfa_enabled = 1, mfa_backup_codes = ?, updated_at = ? WHERE id = ?",
		mfaSecret, backupCodesJSON, time.Now(), operatorID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// VerifyEmail marks an operator's email as verified.
func (r *OperatorRepository) VerifyEmail(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE operators SET email_verified = 1, updated_at = ? WHERE id = ?",
		time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// UpdateEmailVerified updates the email verified status for an operator.
func (r *OperatorRepository) UpdateEmailVerified(ctx context.Context, id string, verified bool) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE operators SET email_verified = ?, updated_at = ? WHERE id = ?",
		verified, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// UpdateGoogleID updates the Google ID for an operator.
func (r *OperatorRepository) UpdateGoogleID(ctx context.Context, id, googleID string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE operators SET google_id = ?, updated_at = ? WHERE id = ?",
		googleID, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// UpdateGitHubID updates the GitHub ID for an operator.
func (r *OperatorRepository) UpdateGitHubID(ctx context.Context, id, githubID string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE operators SET github_id = ?, updated_at = ? WHERE id = ?",
		githubID, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// UpdateName updates the display name for an operator.
func (r *OperatorRepository) UpdateName(ctx context.Context, id, name string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE operators SET name = ?, updated_at = ? WHERE id = ?",
		strings.TrimSpace(name), time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// UpdateThresholds updates the alert thresholds for an operator.
func (r *OperatorRepository) UpdateThresholds(ctx context.Context, id string, th operator.Thresholds) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE operators SET risk_warn = ?, risk_crit = ?, thermal_warn = ?, thermal_crit = ?, 
		 buffer_warn = ?, buffer_crit = ?, updated_at = ? WHERE id = ?`,
		th.RiskWarn, th.RiskCrit, th.ThermalWarn, th.ThermalCrit,
		th.BufferWarn, th.BufferCrit, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// UpdateClientSettings updates the client preferences for an operator.
func (r *OperatorRepository) UpdateClientSettings(ctx context.Context, id string, cs operator.ClientSettings) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE operators SET strict_hmac = ?, auto_reconnect = ?, notifications_enabled = ?, 
		 updated_at = ? WHERE id = ?`,
		cs.StrictHmac, cs.AutoReconnect, cs.NotificationsEnabled, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// ResetSettings resets all settings to defaults for an operator.
func (r *OperatorRepository) ResetSettings(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE operators SET 
		 risk_warn = 50, risk_crit = 75, thermal_warn = 45, thermal_crit = 55, 
		 buffer_warn = 50, buffer_crit = 80,
		 strict_hmac = 0, auto_reconnect = 1, notifications_enabled = 1,
		 updated_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// GetEmailVerified returns whether an operator has verified their email.
func (r *OperatorRepository) GetEmailVerified(ctx context.Context, id string) (bool, error) {
	var verified int
	err := r.db.QueryRowContext(ctx,
		"SELECT email_verified FROM operators WHERE id = ?",
		id,
	).Scan(&verified)
	if errors.Is(err, sql.ErrNoRows) {
		return false, operator.ErrNotFound
	}
	if err != nil {
		return false, err
	}
	return verified != 0, nil
}

// DisableMFA disables MFA for an operator by clearing the MFA secret and backup codes.
func (r *OperatorRepository) DisableMFA(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE operators SET mfa_secret = '', mfa_enabled = 0, mfa_backup_codes = '', updated_at = ? WHERE id = ?",
		time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return operator.ErrNotFound
	}

	return nil
}

// GetSetting retrieves a setting value by key.
func (r *OperatorRepository) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx,
		`SELECT value FROM settings WHERE key = ?`, key,
	).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value, err
}

// SetSetting updates or inserts a setting value.
func (r *OperatorRepository) SetSetting(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO settings(key, value, updated_at) VALUES(?, ?, ?)`,
		key, value, time.Now().UTC().UnixMilli(),
	)
	return err
}

// GetEnforceHMAC returns whether HMAC enforcement is enabled.
func (r *OperatorRepository) GetEnforceHMAC(ctx context.Context) (bool, error) {
	val, err := r.GetSetting(ctx, "enforce_hmac")
	if err != nil || val == "" {
		return false, err
	}
	return val == "true" || val == "1", nil
}

// SetEnforceHMAC updates the HMAC enforcement setting.
func (r *OperatorRepository) SetEnforceHMAC(ctx context.Context, enforce bool) error {
	val := "false"
	if enforce {
		val = "true"
	}
	return r.SetSetting(ctx, "enforce_hmac", val)
}

// GetHMACWindowSeconds returns the HMAC timestamp window in seconds.
func (r *OperatorRepository) GetHMACWindowSeconds(ctx context.Context) (int, error) {
	val, err := r.GetSetting(ctx, "hmac_window_seconds")
	if err != nil {
		return 30, err
	}
	if val == "" {
		return 30, nil // default 30 seconds per COMMAND_SECURITY.md
	}
	var seconds int
	_, err = fmt.Sscanf(val, "%d", &seconds)
	if err != nil {
		return 30, err
	}
	return seconds, nil
}

// SetHMACWindowSeconds updates the HMAC timestamp window.
func (r *OperatorRepository) SetHMACWindowSeconds(ctx context.Context, seconds int) error {
	return r.SetSetting(ctx, "hmac_window_seconds", strconv.Itoa(seconds))
}

// nullString returns a sql.NullString for optional string fields.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
