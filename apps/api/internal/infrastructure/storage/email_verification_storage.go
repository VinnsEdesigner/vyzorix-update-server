package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/email_verification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// Ensure EmailVerificationRepository implements email_verification.Repository.
var _ email_verification.Repository = (*EmailVerificationRepository)(nil)

// EmailVerificationRepository implements email_verification.Repository using SQLite.
type EmailVerificationRepository struct {
	db *sql.DB
}

// NewEmailVerificationRepository creates a new EmailVerificationRepository.
func NewEmailVerificationRepository(db *sql.DB) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

// getQuerier returns the transaction from context if available, otherwise the db.
func (r *EmailVerificationRepository) getQuerier(ctx context.Context) Querier {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}

// queryRow is a helper that uses transaction-aware querier.
func (r *EmailVerificationRepository) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return r.getQuerier(ctx).QueryRowContext(ctx, query, args...)
}

// queryRows is a helper that uses transaction-aware querier.
func (r *EmailVerificationRepository) queryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return r.getQuerier(ctx).QueryContext(ctx, query, args...)
}

// exec is a helper that uses transaction-aware querier.
func (r *EmailVerificationRepository) exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getQuerier(ctx).ExecContext(ctx, query, args...)
}

// Create creates a new email verification.
func (r *EmailVerificationRepository) Create(ctx context.Context, ev *email_verification.EmailVerification) error {
	_, err := r.exec(ctx,
		`INSERT INTO email_verifications(id, operator_id, token_hash, expires_at, created_at)
		 VALUES(?, ?, ?, ?, ?)`,
		ev.ID, ev.OperatorID, ev.TokenHash, ev.ExpiresAt.UnixMilli(), ev.CreatedAt.UnixMilli(),
	)

	return err
}

// FindByTokenHash retrieves an email verification by token hash.
func (r *EmailVerificationRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*email_verification.EmailVerification, error) {
	var ev email_verification.EmailVerification

	var expiresAt, createdAt int64

	err := r.queryRow(ctx,
		`SELECT id, operator_id, token_hash, expires_at, created_at
		 FROM email_verifications WHERE token_hash = ?`,
		tokenHash,
	).Scan(&ev.ID, &ev.OperatorID, &ev.TokenHash, &expiresAt, &createdAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, email_verification.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	ev.ExpiresAt = time.UnixMilli(expiresAt).UTC()
	ev.CreatedAt = time.UnixMilli(createdAt).UTC()

	return &ev, nil
}

// Delete removes an email verification by ID.
func (r *EmailVerificationRepository) Delete(ctx context.Context, id string) error {
	_, err := r.exec(ctx, `DELETE FROM email_verifications WHERE id = ?`, id)
	return err
}

// DeleteByOperator removes all verifications for an operator.
func (r *EmailVerificationRepository) DeleteByOperator(ctx context.Context, operatorID string) error {
	_, err := r.exec(ctx, `DELETE FROM email_verifications WHERE operator_id = ?`, operatorID)
	return err
}
