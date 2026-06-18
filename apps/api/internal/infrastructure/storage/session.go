package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
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
