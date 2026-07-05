package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/refresh_token"
)

// Ensure SessionRepository implements session.Repository.
var _ session.Repository = (*SessionRepository)(nil)

// SessionRepository implements session.Repository using SQLite.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// FindByID retrieves a session by ID.
func (r *SessionRepository) FindByID(ctx context.Context, id string) (*session.Session, error) {
	query := `SELECT id, operator_id, expires_at, created_at, ip_address, user_agent FROM auth_sessions WHERE id = ?`

	var s session.Session

	var ipAddress, userAgent sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.OperatorID, &s.ExpiresAt, &s.CreatedAt, &ipAddress, &userAgent,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, session.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	s.IPAddress = ipAddress.String
	s.UserAgent = userAgent.String

	return &s, nil
}

// FindByOperatorID retrieves all sessions for an operator.
func (r *SessionRepository) FindByOperatorID(ctx context.Context, operatorID string) ([]*session.Session, error) {
	query := `SELECT id, operator_id, expires_at, created_at, ip_address, user_agent FROM auth_sessions WHERE operator_id = ?`

	rows, err := r.db.QueryContext(ctx, query, operatorID)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var sessions []*session.Session

	for rows.Next() {
		var s session.Session

		var ipAddress, userAgent sql.NullString

		if err := rows.Scan(&s.ID, &s.OperatorID, &s.ExpiresAt, &s.CreatedAt, &ipAddress, &userAgent); err != nil {
			return nil, err
		}

		s.IPAddress = ipAddress.String
		s.UserAgent = userAgent.String
		sessions = append(sessions, &s)
	}

	return sessions, rows.Err()
}

// ListActiveByOperator retrieves active (non-expired) sessions for an operator, ordered by creation time.
func (r *SessionRepository) ListActiveByOperator(ctx context.Context, operatorID string) ([]*session.Session, error) {
	query := `SELECT id, operator_id, expires_at, created_at, ip_address, user_agent 
		FROM auth_sessions 
		WHERE operator_id = ? AND expires_at > ?
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, operatorID, time.Now())
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var sessions []*session.Session

	for rows.Next() {
		var s session.Session

		var ipAddress, userAgent sql.NullString

		if err := rows.Scan(&s.ID, &s.OperatorID, &s.ExpiresAt, &s.CreatedAt, &ipAddress, &userAgent); err != nil {
			return nil, err
		}

		s.IPAddress = ipAddress.String
		s.UserAgent = userAgent.String
		sessions = append(sessions, &s)
	}

	return sessions, rows.Err()
}

// Create creates a new session.
func (r *SessionRepository) Create(ctx context.Context, s *session.Session) error {
	query := `INSERT INTO auth_sessions (id, operator_id, expires_at, created_at, ip_address, user_agent) VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.OperatorID, s.ExpiresAt, s.CreatedAt,
		nullString(s.IPAddress), nullString(s.UserAgent),
	)

	return err
}

// Delete deletes a session.
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM auth_sessions WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return session.ErrNotFound
	}

	return nil
}

// DeleteByOperatorID deletes all sessions for an operator.
func (r *SessionRepository) DeleteByOperatorID(ctx context.Context, operatorID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM auth_sessions WHERE operator_id = ?", operatorID)
	return err
}

// DeleteExpired deletes all expired sessions.
func (r *SessionRepository) DeleteExpired(ctx context.Context) (int, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM auth_sessions WHERE expires_at < ?", time.Now())
	if err != nil {
		return 0, err
	}

	deleted, err := result.RowsAffected()

	return int(deleted), err
}

// Extend extends a session's expiration time.
func (r *SessionRepository) Extend(ctx context.Context, id string, newExpiry time.Time) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE auth_sessions SET expires_at = ? WHERE id = ?",
		newExpiry, id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return session.ErrNotFound
	}

	return nil
}

// AddSessionRevocation adds a session token hash to the revocation list.
func (r *SessionRepository) AddSessionRevocation(ctx context.Context, tokenHash, reason string) error {
	query := `INSERT INTO session_revocations (session_id, revoked_at, reason) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, tokenHash, time.Now().UnixMilli(), reason)

	return err
}

// IsSessionRevoked checks if a session token hash is in the revocation list.
func (r *SessionRepository) IsSessionRevoked(ctx context.Context, tokenHash string) (bool, error) {
	query := `SELECT COUNT(*) FROM session_revocations WHERE session_id = ?`

	var count int

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// RemoveSessionRevocation removes a session from the revocation list.
func (r *SessionRepository) RemoveSessionRevocation(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM session_revocations WHERE session_id = ?`
	_, err := r.db.ExecContext(ctx, query, tokenHash)

	return err
}

// ListSessionRevocations retrieves all revoked sessions, optionally filtered by reason.
func (r *SessionRepository) ListSessionRevocations(ctx context.Context, reason string, limit int) ([]*session.SessionRevocation, error) {
	if limit <= 0 {
		limit = 100
	}

	if limit > 1000 {
		limit = 1000
	}

	var query string

	var args []interface{}

	if reason != "" {
		query = `SELECT token_hash, revoked_at, reason FROM session_revocation_list WHERE reason = ? ORDER BY revoked_at DESC LIMIT ?`
		args = []interface{}{reason, limit}
	} else {
		query = `SELECT token_hash, revoked_at, reason FROM session_revocation_list ORDER BY revoked_at DESC LIMIT ?`
		args = []interface{}{limit}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var revocations []*session.SessionRevocation

	for rows.Next() {
		var rev session.SessionRevocation

		var revokedAt int64

		var reasonVal sql.NullString
		if err := rows.Scan(&rev.TokenHash, &revokedAt, &reasonVal); err != nil {
			return nil, err
		}

		rev.RevokedAt = time.UnixMilli(revokedAt)
		if reasonVal.Valid {
			rev.Reason = reasonVal.String
		}

		revocations = append(revocations, &rev)
	}

	return revocations, rows.Err()
}

// CleanupSessionRevocations removes revocation entries older than the specified duration.
func (r *SessionRepository) CleanupSessionRevocations(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan).UnixMilli()
	query := `DELETE FROM session_revocation_list WHERE revoked_at < ?`

	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, err
	}

	n, _ := result.RowsAffected()

	return int(n), nil
}

// RevokeAllOperatorSessions revokes all sessions for a specific operator.
func (r *SessionRepository) RevokeAllOperatorSessions(ctx context.Context, operatorID string) error {
	query := `SELECT id FROM auth_sessions WHERE operator_id = ?`

	rows, err := r.db.QueryContext(ctx, query, operatorID)
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			continue
		}

		if err := r.AddSessionRevocation(ctx, sessionID, "operator_logout"); err != nil {
			return err
		}
	}

	return rows.Err()
}

// =============================================================================
// Refresh Token Repository (Implements refresh_token.Repository)
// =============================================================================

// Ensure RefreshTokenRepository implements refresh_token.Repository.
var _ refresh_token.Repository = (*RefreshTokenRepository)(nil)

// RefreshTokenRepository implements refresh_token.Repository using SQLite.
type RefreshTokenRepository struct {
db *sql.DB
}


// NewRefreshTokenRepository creates a new RefreshTokenRepository.
func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token.
func (r *RefreshTokenRepository) Create(ctx context.Context, rt *refresh_token.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, token_hash, operator_id, session_id, expires_at, created_at, revoked_at, replaced_by_id, is_revoked)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		rt.ID, rt.TokenHash, rt.OperatorID, rt.SessionID, rt.ExpiresAt, rt.CreatedAt, rt.RevokedAt, rt.ReplacedByID, rt.IsRevoked,
	)

	return err
}

// FindByID retrieves a refresh token by ID.
func (r *RefreshTokenRepository) FindByID(ctx context.Context, id string) (*refresh_token.RefreshToken, error) {
	query := `
		SELECT id, token_hash, operator_id, session_id, expires_at, created_at, revoked_at, replaced_by_id, is_revoked
		FROM refresh_tokens WHERE id = ?`

	var rt refresh_token.RefreshToken

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rt.ID, &rt.TokenHash, &rt.OperatorID, &rt.SessionID, &rt.ExpiresAt, &rt.CreatedAt, &rt.RevokedAt, &rt.ReplacedByID, &rt.IsRevoked,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, refresh_token.ErrNotFound
	}

	return &rt, err
}

// FindByTokenHash retrieves a refresh token by its hash.
func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*refresh_token.RefreshToken, error) {
	query := `
		SELECT id, token_hash, operator_id, session_id, expires_at, created_at, revoked_at, replaced_by_id, is_revoked
		FROM refresh_tokens WHERE token_hash = ?`

	var rt refresh_token.RefreshToken

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&rt.ID, &rt.TokenHash, &rt.OperatorID, &rt.SessionID, &rt.ExpiresAt, &rt.CreatedAt, &rt.RevokedAt, &rt.ReplacedByID, &rt.IsRevoked,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, refresh_token.ErrNotFound
	}

	return &rt, err
}

// Revoke revokes a refresh token by ID.
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true, revoked_at = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return refresh_token.ErrNotFound
	}

	return nil
}

// RevokeByTokenHash revokes a refresh token by its hash.
func (r *RefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true, revoked_at = ? WHERE token_hash = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), tokenHash)

	return err
}

// RevokeAllForOperator revokes all refresh tokens for an operator.
func (r *RefreshTokenRepository) RevokeAllForOperator(ctx context.Context, operatorID string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true, revoked_at = ? WHERE operator_id = ? AND is_revoked = false`
	_, err := r.db.ExecContext(ctx, query, time.Now(), operatorID)

	return err
}

// CleanupExpired removes expired refresh tokens older than the specified duration.
func (r *RefreshTokenRepository) CleanupExpired(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	query := `DELETE FROM refresh_tokens WHERE expires_at < ? AND is_revoked = true AND revoked_at < ?`

	result, err := r.db.ExecContext(ctx, query, cutoff, cutoff)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()

	return int(rows), err
}
