package serverlock

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrLockHeld = errors.New("lock is held by another holder")

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Acquire(ctx context.Context, name, holder string, ttl time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	expires := now + ttl.Milliseconds()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO server_locks (name, holder, acquired_at, expires_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(name) DO UPDATE SET holder = ?, acquired_at = ?, expires_at = ?
		 WHERE expires_at < ?`,
		name, holder, now, expires, holder, now, expires, now)
	if err != nil {
		return false, err
	}
	var actualHolder string
	err = s.db.QueryRowContext(ctx, `SELECT holder FROM server_locks WHERE name = ?`, name).Scan(&actualHolder)
	if err != nil {
		return false, err
	}
	return actualHolder == holder, nil
}

func (s *Service) Release(ctx context.Context, name, holder string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM server_locks WHERE name = ? AND holder = ?`, name, holder)
	return err
}

func (s *Service) IsHeld(ctx context.Context, name string) (bool, error) {
	var expiresAt int64
	err := s.db.QueryRowContext(ctx, `SELECT expires_at FROM server_locks WHERE name = ?`, name).Scan(&expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return time.Now().UnixMilli() < expiresAt, nil
}
