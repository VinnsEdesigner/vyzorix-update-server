// Package storage provides SQLite database operations.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

//  System Settings 

// GetSetting retrieves a setting value by key.
func (s *Store) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM settings WHERE key = ?`, key,
	).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value, err
}

// SetSetting updates or inserts a setting value.
func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO settings(key, value, updated_at) VALUES(?, ?, ?)`,
		key, value, time.Now().UTC().UnixMilli(),
	)
	return err
}

// GetEnforceHMAC returns whether HMAC enforcement is enabled.
func (s *Store) GetEnforceHMAC(ctx context.Context) (bool, error) {
	val, err := s.GetSetting(ctx, "enforce_hmac")
	if err != nil || val == "" {
		return false, err
	}
	return val == "true" || val == "1", nil
}

// SetEnforceHMAC updates the HMAC enforcement setting.
func (s *Store) SetEnforceHMAC(ctx context.Context, enforce bool) error {
	val := "false"
	if enforce {
		val = "true"
	}
	return s.SetSetting(ctx, "enforce_hmac", val)
}

// GetHMACWindowSeconds returns the HMAC timestamp window in seconds.
func (s *Store) GetHMACWindowSeconds(ctx context.Context) (int, error) {
	val, err := s.GetSetting(ctx, "hmac_window_seconds")
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
func (s *Store) SetHMACWindowSeconds(ctx context.Context, seconds int) error {
	return s.SetSetting(ctx, "hmac_window_seconds", strconv.Itoa(seconds))
}

//  Password Reset Resend Tracker 

// GetPasswordResetResendTracker retrieves the resend tracker for an email hash.
func (s *Store) GetPasswordResetResendTracker(ctx context.Context, emailHash string) (*models.PasswordResetResendTracker, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email_hash, resend_count, last_resend_at, lockout_until, created_at, updated_at
		FROM password_reset_resend_tracker
		WHERE email_hash = ?`, emailHash)

	var tracker models.PasswordResetResendTracker
	var lastResendAt, createdAt, updatedAt int64
	var lockoutUntil *int64
	err := row.Scan(
		&tracker.ID,
		&tracker.EmailHash,
		&tracker.ResendCount,
		&lastResendAt,
		&lockoutUntil,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get resend tracker: %w", err)
	}
	tracker.LastResendAt = time.UnixMilli(lastResendAt)
	tracker.CreatedAt = time.UnixMilli(createdAt)
	tracker.UpdatedAt = time.UnixMilli(updatedAt)
	if lockoutUntil != nil {
		lt := time.UnixMilli(*lockoutUntil)
		tracker.LockoutUntil = &lt
	}
	return &tracker, nil
}

// UpsertPasswordResetResendTracker creates or updates a resend tracker.
func (s *Store) UpsertPasswordResetResendTracker(ctx context.Context, tracker *models.PasswordResetResendTracker) error {
	var lockoutUntil *int64
	if tracker.LockoutUntil != nil {
		lt := tracker.LockoutUntil.UnixMilli()
		lockoutUntil = &lt
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO password_reset_resend_tracker (id, email_hash, resend_count, last_resend_at, lockout_until, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(email_hash) DO UPDATE SET
			resend_count = excluded.resend_count,
			last_resend_at = excluded.last_resend_at,
			lockout_until = excluded.lockout_until,
			updated_at = excluded.updated_at`,
		tracker.ID,
		tracker.EmailHash,
		tracker.ResendCount,
		tracker.LastResendAt.UnixMilli(),
		lockoutUntil,
		tracker.CreatedAt.UnixMilli(),
		tracker.UpdatedAt.UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert resend tracker: %w", err)
	}
	return nil
}

// DeletePasswordResetResendTracker removes a resend tracker by email hash.
func (s *Store) DeletePasswordResetResendTracker(ctx context.Context, emailHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM password_reset_resend_tracker WHERE email_hash = ?`, emailHash)
	if err != nil {
		return fmt.Errorf("failed to delete resend tracker: %w", err)
	}
	return nil
}

// CleanupPasswordResetResendTrackers removes old resend trackers.
// Trackers older than maxAge are deleted.
func (s *Store) CleanupPasswordResetResendTrackers(ctx context.Context, maxAgeHours int) (int64, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(maxAgeHours) * time.Hour).UnixMilli()
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM password_reset_resend_tracker
		WHERE updated_at < ? OR (lockout_until IS NOT NULL AND lockout_until < ?)`,
		cutoff, time.Now().UTC().UnixMilli())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup resend trackers: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}
	return rows, nil
}

//  Email Verification 

// EmailVerification represents a pending email verification token.
type EmailVerification struct {
	ExpiresAt  time.Time
	CreatedAt  time.Time
	ID         string
	OperatorID string
	TokenHash  string
}

// CreateEmailVerification inserts a new email verification token.
func (s *Store) CreateEmailVerification(ctx context.Context, ev *EmailVerification) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO email_verifications(id, operator_id, token_hash, expires_at, created_at)
		 VALUES(?, ?, ?, ?, ?)`,
		ev.ID, ev.OperatorID, ev.TokenHash, ev.ExpiresAt.UnixMilli(), ev.CreatedAt.UnixMilli(),
	)
	return err
}

// GetEmailVerificationByTokenHash retrieves an email verification by its token hash.
func (s *Store) GetEmailVerificationByTokenHash(ctx context.Context, tokenHash string) (*EmailVerification, error) {
	var r struct {
		ID         string
		OperatorID string
		TokenHash  string
		ExpiresAt  int64
		CreatedAt  int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, operator_id, token_hash, expires_at, created_at
		 FROM email_verifications WHERE token_hash = ?`,
		tokenHash,
	).Scan(&r.ID, &r.OperatorID, &r.TokenHash, &r.ExpiresAt, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &EmailVerification{
		ID:         r.ID,
		OperatorID: r.OperatorID,
		TokenHash:  r.TokenHash,
		ExpiresAt:  time.UnixMilli(r.ExpiresAt).UTC(),
		CreatedAt:  time.UnixMilli(r.CreatedAt).UTC(),
	}, nil
}

// DeleteEmailVerification removes an email verification by ID.
func (s *Store) DeleteEmailVerification(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM email_verifications WHERE id = ?`, id)
	return err
}

// DeleteEmailVerificationsByOperator removes all email verifications for an operator.
func (s *Store) DeleteEmailVerificationsByOperator(ctx context.Context, operatorID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM email_verifications WHERE operator_id = ?`, operatorID)
	return err
}

//  Password Reset Tokens 

// PasswordResetToken represents a pending password reset token.
type PasswordResetToken struct {
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UsedAt     *time.Time
	ID         string
	OperatorID string
	TokenHash  string
}

// CreatePasswordResetToken inserts a new password reset token.
func (s *Store) CreatePasswordResetToken(ctx context.Context, prt *PasswordResetToken) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens(id, operator_id, token_hash, expires_at, used_at, created_at)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		prt.ID, prt.OperatorID, prt.TokenHash, prt.ExpiresAt.UnixMilli(), nil, prt.CreatedAt.UnixMilli(),
	)
	return err
}

// GetPasswordResetTokenByHash retrieves a password reset token by its hash.
func (s *Store) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	var r struct {
		ID         string
		OperatorID string
		TokenHash  string
		ExpiresAt  int64
		UsedAt     *int64
		CreatedAt  int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, operator_id, token_hash, expires_at, used_at, created_at
		 FROM password_reset_tokens WHERE token_hash = ?`,
		tokenHash,
	).Scan(&r.ID, &r.OperatorID, &r.TokenHash, &r.ExpiresAt, &r.UsedAt, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var usedAt *time.Time
	if r.UsedAt != nil {
		t := time.UnixMilli(*r.UsedAt).UTC()
		usedAt = &t
	}
	return &PasswordResetToken{
		ID:         r.ID,
		OperatorID: r.OperatorID,
		TokenHash:  r.TokenHash,
		ExpiresAt:  time.UnixMilli(r.ExpiresAt).UTC(),
		UsedAt:     usedAt,
		CreatedAt:  time.UnixMilli(r.CreatedAt).UTC(),
	}, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used.
func (s *Store) MarkPasswordResetTokenUsed(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = ? WHERE id = ?`,
		time.Now().UTC().UnixMilli(), id,
	)
	return err
}

// DeletePasswordResetToken removes a password reset token by ID.
func (s *Store) DeletePasswordResetToken(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE id = ?`, id)
	return err
}

// DeletePasswordResetTokensByOperator removes all password reset tokens for an operator.
func (s *Store) DeletePasswordResetTokensByOperator(ctx context.Context, operatorID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE operator_id = ?`, operatorID)
	return err
}

//  Auth Sessions 

// CreateSession inserts a new auth session.
func (s *Store) CreateSession(ctx context.Context, sess *models.Session) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO auth_sessions(id, operator_id, token_hash, expires_at, created_at, user_agent, ip_address)
		 VALUES(?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.OperatorID, sess.TokenHash, sess.ExpiresAt.UnixMilli(), sess.CreatedAt.UnixMilli(),
		sess.UserAgent, sess.IPAddress,
	)
	return err
}

// GetSessionByTokenHash retrieves a session by its token hash.
func (s *Store) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	var r struct {
		ID         string
		OperatorID string
		UserAgent  *string
		IPAddress  *string
		ExpiresAt  int64
		CreatedAt  int64
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, operator_id, expires_at, created_at, user_agent, ip_address
		 FROM auth_sessions WHERE token_hash = ?`,
		tokenHash,
	).Scan(&r.ID, &r.OperatorID, &r.ExpiresAt, &r.CreatedAt, &r.UserAgent, &r.IPAddress)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &models.Session{
		ID:         r.ID,
		OperatorID: r.OperatorID,
		ExpiresAt:  time.UnixMilli(r.ExpiresAt).UTC(),
		CreatedAt:  time.UnixMilli(r.CreatedAt).UTC(),
		UserAgent:  stringPtrToString(r.UserAgent),
		IPAddress:  stringPtrToString(r.IPAddress),
	}, nil
}

func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DeleteSession removes a session by its token hash (logout).
func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash = ?`, tokenHash)
	return err
}

// DeleteExpiredSessions removes all sessions past their expiry time.
func (s *Store) DeleteExpiredSessions(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM auth_sessions WHERE expires_at < ?`,
		time.Now().UTC().UnixMilli(),
	)
	return err
}

// DeleteAllSessionsForOperator removes all sessions for a given operator (used on password change).
func (s *Store) DeleteAllSessionsForOperator(ctx context.Context, operatorID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE operator_id = ?`, operatorID)
	return err
}