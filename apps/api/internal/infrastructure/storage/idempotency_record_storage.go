// Package storage provides SQLite storage implementations.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/idempotency"
)

// Ensure IdempotencyRepository implements idempotency.Repository.
var _ idempotency.Repository = (*IdempotencyRepository)(nil)

// IdempotencyRepository implements idempotency.Repository using SQLite.
type IdempotencyRepository struct {
	db *sql.DB
}

// NewIdempotencyRepository creates a new IdempotencyRepository.
func NewIdempotencyRepository(db *sql.DB) *IdempotencyRepository {
	return &IdempotencyRepository{db: db}
}

// Get retrieves an idempotency record by key.
func (r *IdempotencyRepository) Get(ctx context.Context, key string) (*idempotency.IdempotencyRecord, error) {
	query := `
		SELECT id, method, path, hash, status_code, response_body, 
		       created_at, expires_at, content_type, client_ip, user_agent
		FROM idempotency_records
		WHERE id = ? AND expires_at > ?`

	var record idempotency.IdempotencyRecord
	var responseBody []byte
	var createdAt, expiresAt time.Time

	err := r.db.QueryRowContext(ctx, query, key, time.Now()).Scan(
		&record.ID,
		&record.Method,
		&record.Path,
		&record.Hash,
		&record.StatusCode,
		&responseBody,
		&createdAt,
		&expiresAt,
		&record.ContentType,
		&record.ClientIP,
		&record.UserAgent,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, idempotency.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	record.ResponseBody = responseBody
	record.CreatedAt = createdAt
	record.ExpiresAt = expiresAt

	return &record, nil
}

// Create stores a new idempotency record.
func (r *IdempotencyRepository) Create(ctx context.Context, record *idempotency.IdempotencyRecord) error {
	query := `
		INSERT INTO idempotency_records (
			id, method, path, hash, status_code, response_body,
			created_at, expires_at, content_type, client_ip, user_agent
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		record.ID,
		record.Method,
		record.Path,
		record.Hash,
		record.StatusCode,
		record.ResponseBody,
		record.CreatedAt,
		record.ExpiresAt,
		record.ContentType,
		record.ClientIP,
		record.UserAgent,
	)
	return err
}

// Delete removes an idempotency record by key.
func (r *IdempotencyRepository) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM idempotency_records WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, key)
	return err
}

// DeleteExpired removes all expired records.
func (r *IdempotencyRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM idempotency_records WHERE expires_at <= ?`
	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
