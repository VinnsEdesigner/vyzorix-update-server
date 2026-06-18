// Package storage provides SQLite database operations.
package storage

import (
	"context"
	"database/sql"
	"time"
)

// SessionRevocation represents a revoked session entry in the database.
type SessionRevocation struct {
	TokenHash string    `json:"token_hash"`
	RevokedAt time.Time `json:"revoked_at"`
	Reason    string    `json:"reason,omitempty"`
}

// AddSessionRevocation adds a session token hash to the revocation list.
func (s *Store) AddSessionRevocation(ctx context.Context, tokenHash, reason string) error {
	query := `INSERT INTO session_revocation_list (token_hash, revoked_at, reason) VALUES (?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, tokenHash, time.Now().UnixMilli(), reason)
	return err
}

// IsSessionRevoked checks if a session token hash is in the revocation list.
func (s *Store) IsSessionRevoked(ctx context.Context, tokenHash string) (bool, error) {
	query := `SELECT COUNT(*) FROM session_revocation_list WHERE token_hash = ?`
	var count int
	err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RemoveSessionRevocation removes a session from the revocation list.
func (s *Store) RemoveSessionRevocation(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM session_revocation_list WHERE token_hash = ?`
	_, err := s.db.ExecContext(ctx, query, tokenHash)
	return err
}

// ListSessionRevocations retrieves all revoked sessions, optionally filtered by reason.
func (s *Store) ListSessionRevocations(ctx context.Context, reason string, limit int) ([]*SessionRevocation, error) {
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

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var revocations []*SessionRevocation
	for rows.Next() {
		var r SessionRevocation
		var revokedAt int64
		var reason sql.NullString
		if err := rows.Scan(&r.TokenHash, &revokedAt, &reason); err != nil {
			return nil, err
		}
		r.RevokedAt = time.UnixMilli(revokedAt)
		if reason.Valid {
			r.Reason = reason.String
		}
		revocations = append(revocations, &r)
	}
	return revocations, rows.Err()
}

// CleanupSessionRevocations removes revocation entries older than the specified duration.
func (s *Store) CleanupSessionRevocations(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan).UnixMilli()
	query := `DELETE FROM session_revocation_list WHERE revoked_at < ?`
	result, err := s.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

// RevokeAllOperatorSessions revokes all sessions for a specific operator.
func (s *Store) RevokeAllOperatorSessions(ctx context.Context, operatorID string) error {
	// Get all session hashes for this operator and add them to revocation list
	query := `SELECT id FROM auth_sessions WHERE operator_id = ?`
	rows, err := s.db.QueryContext(ctx, query, operatorID)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			continue
		}
		// Add to revocation list with reason
		if err := s.AddSessionRevocation(ctx, sessionID, "operator_logout"); err != nil {
			return err
		}
	}
	return rows.Err()
}
