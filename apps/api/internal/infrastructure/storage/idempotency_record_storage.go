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
	if err != nil {
		// Check for duplicate key error.
		if isUniqueConstraintError(err) {
			return idempotency.ErrConflict
		}
		return err
	}
	return nil
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

// Count returns the total number of idempotency records.
func (r *IdempotencyRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM idempotency_records WHERE expires_at > ?`
	var count int64
	err := r.db.QueryRowContext(ctx, query, time.Now()).Scan(&count)
	return count, err
}

// CountByPath returns the number of records for a given path prefix.
func (r *IdempotencyRepository) CountByPath(ctx context.Context, pathPrefix string) (int64, error) {
	query := `SELECT COUNT(*) FROM idempotency_records WHERE path LIKE ? AND expires_at > ?`
	var count int64
	err := r.db.QueryRowContext(ctx, query, pathPrefix+"%", time.Now()).Scan(&count)
	return count, err
}

// GetByPath retrieves records by path prefix with pagination.
func (r *IdempotencyRepository) GetByPath(ctx context.Context, pathPrefix string, limit, offset int) ([]*idempotency.IdempotencyRecord, error) {
	query := `
		SELECT id, method, path, hash, status_code, response_body, 
		       created_at, expires_at, content_type, client_ip, user_agent
		FROM idempotency_records
		WHERE path LIKE ? AND expires_at > ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, pathPrefix+"%", time.Now(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var records []*idempotency.IdempotencyRecord
	for rows.Next() {
		var record idempotency.IdempotencyRecord
		var responseBody []byte
		var createdAt, expiresAt time.Time

		scanErr := rows.Scan(
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
		if scanErr != nil {
			return nil, scanErr
		}

		record.ResponseBody = responseBody
		record.CreatedAt = createdAt
		record.ExpiresAt = expiresAt
		records = append(records, &record)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	return records, nil
}
