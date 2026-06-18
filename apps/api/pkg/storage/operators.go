// Package storage provides SQLite database operations.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

// OperatorCount returns the total number of operators in the system.
func (s *Store) OperatorCount(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM operators`).Scan(&n)
	return n, err
}

// ListOperators retrieves all operators in the system.
func (s *Store) ListOperators(ctx context.Context) ([]*models.Operator, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, email, name, password_hash, role, google_id, github_id, COALESCE(email_verified, 0),
		        COALESCE(risk_warn, 50), COALESCE(risk_crit, 75),
		        COALESCE(thermal_warn, 45), COALESCE(thermal_crit, 55),
		        COALESCE(buffer_warn, 50), COALESCE(buffer_crit, 80),
		        COALESCE(strict_hmac, 0), COALESCE(auto_reconnect, 1), COALESCE(notifications_enabled, 1),
		        created_at, updated_at
		 FROM operators ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var operators []*models.Operator
	for rows.Next() {
		op, err := scanOperator(rows)
		if err != nil {
			return nil, err
		}
		operators = append(operators, op)
	}
	return operators, rows.Err()
}

// scanOperator scans a single operator from a row.
func scanOperator(scanner interface{ Scan(...any) error }) (*models.Operator, error) {
	var r struct {
		ID                   string
		Email                string
		Name                 string
		PasswordHash         []byte
		Role                 string
		GoogleID             sql.NullString
		GitHubID             sql.NullString
		EmailVerified        int
		RiskWarn             int
		RiskCrit             int
		ThermalWarn          int
		ThermalCrit          int
		BufferWarn           int
		BufferCrit           int
		StrictHmac           int
		AutoReconnect        int
		NotificationsEnabled int
		CreatedAt            int64
		UpdatedAt            int64
	}
	if err := scanner.Scan(
		&r.ID, &r.Email, &r.Name, &r.PasswordHash, &r.Role, &r.GoogleID, &r.GitHubID, &r.EmailVerified,
		&r.RiskWarn, &r.RiskCrit, &r.ThermalWarn, &r.ThermalCrit, &r.BufferWarn, &r.BufferCrit,
		&r.StrictHmac, &r.AutoReconnect, &r.NotificationsEnabled,
		&r.CreatedAt, &r.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &models.Operator{
		ID:           r.ID,
		Email:        r.Email,
		Name:         r.Name,
		PasswordHash: string(r.PasswordHash),
		Role:         models.OperatorRole(r.Role),
		GoogleID:     r.GoogleID.String,
		GitHubID:     r.GitHubID.String,
		Thresholds: models.Thresholds{
			RiskWarn:    r.RiskWarn,
			RiskCrit:    r.RiskCrit,
			ThermalWarn: r.ThermalWarn,
			ThermalCrit: r.ThermalCrit,
			BufferWarn:  r.BufferWarn,
			BufferCrit:  r.BufferCrit,
		},
		Client: models.ClientSettings{
			StrictHmac:           r.StrictHmac != 0,
			AutoReconnect:        r.AutoReconnect != 0,
			NotificationsEnabled: r.NotificationsEnabled != 0,
		},
		EmailVerified: r.EmailVerified != 0,
		CreatedAt:    time.UnixMilli(r.CreatedAt).UTC(),
		UpdatedAt:    time.UnixMilli(r.UpdatedAt).UTC(),
	}, nil
}

// GetOperatorByEmail retrieves an operator by email address.
func (s *Store) GetOperatorByEmail(ctx context.Context, email string) (*models.Operator, error) {
	var r struct {
		ID            string
		Email         string
		Name          string
		PasswordHash  []byte
		Role          string
		GoogleID      sql.NullString
		EmailVerified int
		CreatedAt     int64
		UpdatedAt     int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, role, google_id, COALESCE(email_verified, 0), created_at, updated_at
		 FROM operators WHERE email = ?`,
		email,
	).Scan(&r.ID, &r.Email, &r.Name, &r.PasswordHash, &r.Role, &r.GoogleID, &r.EmailVerified, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &models.Operator{
		ID:            r.ID,
		Email:         r.Email,
		Name:          r.Name,
		PasswordHash:  string(r.PasswordHash),
		Role:          models.OperatorRole(r.Role),
		GoogleID:      r.GoogleID.String,
		EmailVerified: r.EmailVerified != 0,
		CreatedAt:     time.UnixMilli(r.CreatedAt).UTC(),
		UpdatedAt:     time.UnixMilli(r.UpdatedAt).UTC(),
	}, nil
}

// GetOperatorByGoogleID retrieves an operator by Google OAuth subject ID.
func (s *Store) GetOperatorByGoogleID(ctx context.Context, googleID string) (*models.Operator, error) {
	var r struct {
		ID            string
		Email         string
		Name          string
		Role          string
		GoogleID      string
		PasswordHash  []byte
		EmailVerified int
		CreatedAt     int64
		UpdatedAt     int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, role, google_id, COALESCE(email_verified, 0), created_at, updated_at
		 FROM operators WHERE google_id = ?`,
		googleID,
	).Scan(&r.ID, &r.Email, &r.Name, &r.PasswordHash, &r.Role, &r.GoogleID, &r.EmailVerified, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &models.Operator{
		ID:            r.ID,
		Email:         r.Email,
		Name:          r.Name,
		PasswordHash:  string(r.PasswordHash),
		Role:          models.OperatorRole(r.Role),
		GoogleID:      r.GoogleID,
		EmailVerified: r.EmailVerified != 0,
		CreatedAt:     time.UnixMilli(r.CreatedAt).UTC(),
		UpdatedAt:     time.UnixMilli(r.UpdatedAt).UTC(),
	}, nil
}

// GetOperatorByGitHubID retrieves an operator by GitHub OAuth subject ID.
func (s *Store) GetOperatorByGitHubID(ctx context.Context, githubID string) (*models.Operator, error) {
	var r struct {
		ID                   string
		Email                string
		Name                 string
		Role                 string
		GitHubID             sql.NullString
		GoogleID             sql.NullString
		PasswordHash         []byte
		EmailVerified        int
		RiskWarn             int
		RiskCrit             int
		ThermalWarn          int
		ThermalCrit          int
		BufferWarn           int
		BufferCrit           int
		StrictHmac           int
		AutoReconnect        int
		NotificationsEnabled int
		CreatedAt            int64
		UpdatedAt            int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, role, google_id, github_id, COALESCE(email_verified, 0),
		        COALESCE(risk_warn, 50), COALESCE(risk_crit, 75),
		        COALESCE(thermal_warn, 45), COALESCE(thermal_crit, 55),
		        COALESCE(buffer_warn, 50), COALESCE(buffer_crit, 80),
		        COALESCE(strict_hmac, 0), COALESCE(auto_reconnect, 1), COALESCE(notifications_enabled, 1),
		        created_at, updated_at
		 FROM operators WHERE github_id = ?`,
		githubID,
	).Scan(
		&r.ID, &r.Email, &r.Name, &r.PasswordHash, &r.Role, &r.GoogleID, &r.GitHubID, &r.EmailVerified,
		&r.RiskWarn, &r.RiskCrit, &r.ThermalWarn, &r.ThermalCrit, &r.BufferWarn, &r.BufferCrit,
		&r.StrictHmac, &r.AutoReconnect, &r.NotificationsEnabled,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &models.Operator{
		ID:           r.ID,
		Email:        r.Email,
		Name:         r.Name,
		PasswordHash: string(r.PasswordHash),
		Role:         models.OperatorRole(r.Role),
		GoogleID:     r.GoogleID.String,
		GitHubID:     r.GitHubID.String,
		Thresholds: models.Thresholds{
			RiskWarn:    r.RiskWarn,
			RiskCrit:    r.RiskCrit,
			ThermalWarn: r.ThermalWarn,
			ThermalCrit: r.ThermalCrit,
			BufferWarn:  r.BufferWarn,
			BufferCrit:  r.BufferCrit,
		},
		Client: models.ClientSettings{
			StrictHmac:           r.StrictHmac != 0,
			AutoReconnect:        r.AutoReconnect != 0,
			NotificationsEnabled: r.NotificationsEnabled != 0,
		},
		EmailVerified: r.EmailVerified != 0,
		CreatedAt:    time.UnixMilli(r.CreatedAt).UTC(),
		UpdatedAt:    time.UnixMilli(r.UpdatedAt).UTC(),
	}, nil
}

// GetOperatorByID retrieves an operator by their ID.
func (s *Store) GetOperatorByID(ctx context.Context, id string) (*models.Operator, error) {
	var r struct {
		ID                   string
		Email                string
		Name                 string
		PasswordHash         []byte
		Role                 string
		GoogleID             sql.NullString
		GitHubID             sql.NullString
		EmailVerified        int
		RiskWarn             int
		RiskCrit             int
		ThermalWarn          int
		ThermalCrit          int
		BufferWarn           int
		BufferCrit           int
		StrictHmac           int
		AutoReconnect        int
		NotificationsEnabled int
		CreatedAt            int64
		UpdatedAt            int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, role, google_id, github_id, COALESCE(email_verified, 0),
		        COALESCE(risk_warn, 50), COALESCE(risk_crit, 75),
		        COALESCE(thermal_warn, 45), COALESCE(thermal_crit, 55),
		        COALESCE(buffer_warn, 50), COALESCE(buffer_crit, 80),
		        COALESCE(strict_hmac, 0), COALESCE(auto_reconnect, 1), COALESCE(notifications_enabled, 1),
		        created_at, updated_at
		 FROM operators WHERE id = ?`,
		id,
	).Scan(
		&r.ID, &r.Email, &r.Name, &r.PasswordHash, &r.Role, &r.GoogleID, &r.GitHubID, &r.EmailVerified,
		&r.RiskWarn, &r.RiskCrit, &r.ThermalWarn, &r.ThermalCrit, &r.BufferWarn, &r.BufferCrit,
		&r.StrictHmac, &r.AutoReconnect, &r.NotificationsEnabled,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &models.Operator{
		ID:           r.ID,
		Email:        r.Email,
		Name:         r.Name,
		PasswordHash: string(r.PasswordHash),
		Role:         models.OperatorRole(r.Role),
		GoogleID:     r.GoogleID.String,
		GitHubID:     r.GitHubID.String,
		Thresholds: models.Thresholds{
			RiskWarn:    r.RiskWarn,
			RiskCrit:    r.RiskCrit,
			ThermalWarn: r.ThermalWarn,
			ThermalCrit: r.ThermalCrit,
			BufferWarn:  r.BufferWarn,
			BufferCrit:  r.BufferCrit,
		},
		Client: models.ClientSettings{
			StrictHmac:           r.StrictHmac != 0,
			AutoReconnect:        r.AutoReconnect != 0,
			NotificationsEnabled: r.NotificationsEnabled != 0,
		},
		EmailVerified: r.EmailVerified != 0,
		CreatedAt:    time.UnixMilli(r.CreatedAt).UTC(),
		UpdatedAt:    time.UnixMilli(r.UpdatedAt).UTC(),
	}, nil
}

// CreateOperator inserts a new operator.
func (s *Store) CreateOperator(ctx context.Context, op *models.Operator) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO operators(id, email, name, password_hash, role, google_id, github_id, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		op.ID, op.Email, op.Name, op.PasswordHash, string(op.Role), op.GoogleID, op.GitHubID, now.UnixMilli(), now.UnixMilli(),
	)
	return err
}

// UpdateOperatorGoogleID sets the google_id for an operator.
func (s *Store) UpdateOperatorGoogleID(ctx context.Context, operatorID, googleID string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE operators SET google_id = ?, updated_at = ? WHERE id = ?`,
		googleID, now.UnixMilli(), operatorID,
	)
	return err
}

// UpdateOperatorGitHubID sets the github_id for an operator.
func (s *Store) UpdateOperatorGitHubID(ctx context.Context, operatorID, githubID string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE operators SET github_id = ?, updated_at = ? WHERE id = ?`,
		githubID, now.UnixMilli(), operatorID,
	)
	return err
}

// UpdateOperatorName updates the display name for an operator.
func (s *Store) UpdateOperatorName(ctx context.Context, operatorID, name string) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET name = ?, updated_at = ? WHERE id = ?`,
		strings.TrimSpace(name), now.UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// UpdateOperatorThresholds updates the alert thresholds for an operator.
func (s *Store) UpdateOperatorThresholds(ctx context.Context, operatorID string, th models.Thresholds) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET risk_warn=?, risk_crit=?, thermal_warn=?, thermal_crit=?, buffer_warn=?, buffer_crit=?, updated_at=? WHERE id=?`,
		th.RiskWarn, th.RiskCrit, th.ThermalWarn, th.ThermalCrit, th.BufferWarn, th.BufferCrit, now.UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// UpdateOperatorClientSettings updates the client preferences for an operator.
func (s *Store) UpdateOperatorClientSettings(ctx context.Context, operatorID string, cs models.ClientSettings) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET strict_hmac=?, auto_reconnect=?, notifications_enabled=?, updated_at=? WHERE id=?`,
		boolToInt(cs.StrictHmac), boolToInt(cs.AutoReconnect), boolToInt(cs.NotificationsEnabled), now.UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// ResetOperatorSettings resets all settings to defaults for an operator.
func (s *Store) ResetOperatorSettings(ctx context.Context, operatorID string) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET
			risk_warn=50, risk_crit=75, thermal_warn=45, thermal_crit=55, buffer_warn=50, buffer_crit=80,
			strict_hmac=0, auto_reconnect=1, notifications_enabled=1,
			updated_at=?
		 WHERE id=?`,
		now.UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// GetOperatorEmailVerified returns whether an operator has verified their email.
func (s *Store) GetOperatorEmailVerified(ctx context.Context, operatorID string) (bool, error) {
	var verified int
	err := s.db.QueryRowContext(ctx,
		`SELECT email_verified FROM operators WHERE id = ?`,
		operatorID,
	).Scan(&verified)
	if errors.Is(err, sql.ErrNoRows) {
		return false, errors.New("operator not found")
	}
	return verified != 0, err
}

// SetOperatorEmailVerified marks an operator's email as verified.
func (s *Store) SetOperatorEmailVerified(ctx context.Context, operatorID string, verified bool) error {
	v := 0
	if verified {
		v = 1
	}
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET email_verified = ?, updated_at = ? WHERE id = ?`,
		v, time.Now().UTC().UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// UpdateOperatorPassword updates the password hash for an operator.
func (s *Store) UpdateOperatorPassword(ctx context.Context, operatorID, passwordHash string) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET password_hash = ?, updated_at = ? WHERE id = ?`,
		passwordHash, now.UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// UpdateOperatorMFA updates the MFA secret and backup codes for an operator.
func (s *Store) UpdateOperatorMFA(ctx context.Context, operatorID, mfaSecret string, backupCodes []string) error {
	backupCodesJSON := ""
	if len(backupCodes) > 0 {
		codes, err := json.Marshal(backupCodes)
		if err != nil {
			return fmt.Errorf("failed to marshal backup codes: %w", err)
		}
		backupCodesJSON = string(codes)
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET mfa_secret = ?, mfa_enabled = 1, mfa_backup_codes = ?, updated_at = ? WHERE id = ?`,
		mfaSecret, backupCodesJSON, time.Now().UTC().UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

// DisableOperatorMFA disables MFA for an operator by clearing the MFA secret and backup codes.
func (s *Store) DisableOperatorMFA(ctx context.Context, operatorID string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE operators SET mfa_secret = '', mfa_enabled = 0, mfa_backup_codes = '', updated_at = ? WHERE id = ?`,
		time.Now().UTC().UnixMilli(), operatorID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("operator not found")
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}