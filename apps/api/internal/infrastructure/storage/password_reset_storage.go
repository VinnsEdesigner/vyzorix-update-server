package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/password_reset"
)

// Ensure PasswordResetRepository implements password_reset.Repository.
var _ password_reset.Repository = (*PasswordResetRepository)(nil)

// PasswordResetRepository implements password_reset.Repository using SQLite.
type PasswordResetRepository struct {
	db *sql.DB
}

// NewPasswordResetRepository creates a new PasswordResetRepository.
func NewPasswordResetRepository(db *sql.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

// Create creates a new password reset token.
func (r *PasswordResetRepository) Create(ctx context.Context, token *password_reset.PasswordResetToken) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens(id, operator_id, token_hash, expires_at, used_at, created_at)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		token.ID, token.OperatorID, token.TokenHash, token.ExpiresAt.UnixMilli(), nil, token.CreatedAt.UnixMilli(),
	)

	return err
}

// FindByTokenHash retrieves a token by its hash.
func (r *PasswordResetRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*password_reset.PasswordResetToken, error) {
	var token password_reset.PasswordResetToken

	var expiresAt, createdAt int64

	var usedAt *int64

	err := r.db.QueryRowContext(ctx,
		`SELECT id, operator_id, token_hash, expires_at, used_at, created_at
		 FROM password_reset_tokens WHERE token_hash = ?`,
		tokenHash,
	).Scan(&token.ID, &token.OperatorID, &token.TokenHash, &expiresAt, &usedAt, &createdAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, password_reset.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	token.ExpiresAt = time.UnixMilli(expiresAt).UTC()
	token.CreatedAt = time.UnixMilli(createdAt).UTC()

	if usedAt != nil {
		t := time.UnixMilli(*usedAt).UTC()
		token.UsedAt = &t
	}

	return &token, nil
}

// MarkUsed marks a token as used.
func (r *PasswordResetRepository) MarkUsed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = ? WHERE id = ?`,
		time.Now().UTC().UnixMilli(), id,
	)

	return err
}

// DeleteByOperator removes all tokens for an operator.
func (r *PasswordResetRepository) DeleteByOperator(ctx context.Context, operatorID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE operator_id = ?`, operatorID)
	return err
}

// GetResendTracker retrieves the resend tracker for an email hash.
func (r *PasswordResetRepository) GetResendTracker(ctx context.Context, emailHash string) (*password_reset.ResendTracker, error) {
	var tracker password_reset.ResendTracker

	var lastResendAt, createdAt, updatedAt int64

	var lockoutUntil *int64

	err := r.db.QueryRowContext(ctx,
		`SELECT id, email_hash, resend_count, last_resend_at, lockout_until, created_at, updated_at
		 FROM password_reset_resend_tracker WHERE email_hash = ?`,
		emailHash,
	).Scan(&tracker.ID, &tracker.EmailHash, &tracker.ResendCount, &lastResendAt, &lockoutUntil, &createdAt, &updatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, password_reset.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	tracker.LastResendAt = time.UnixMilli(lastResendAt).UTC()
	tracker.CreatedAt = time.UnixMilli(createdAt).UTC()
	tracker.UpdatedAt = time.UnixMilli(updatedAt).UTC()

	if lockoutUntil != nil {
		t := time.UnixMilli(*lockoutUntil).UTC()
		tracker.LockoutUntil = &t
	}

	return &tracker, nil
}

// UpsertResendTracker creates or updates a resend tracker.
func (r *PasswordResetRepository) UpsertResendTracker(ctx context.Context, tracker *password_reset.ResendTracker) error {
	var lockoutUntil *int64

	if tracker.LockoutUntil != nil {
		lt := tracker.LockoutUntil.UnixMilli()
		lockoutUntil = &lt
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO password_reset_resend_tracker(id, email_hash, resend_count, last_resend_at, lockout_until, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(email_hash) DO UPDATE SET
			resend_count = excluded.resend_count,
			last_resend_at = excluded.last_resend_at,
			lockout_until = excluded.lockout_until,
			updated_at = excluded.updated_at`,
		tracker.ID, tracker.EmailHash, tracker.ResendCount, tracker.LastResendAt.UnixMilli(),
		lockoutUntil, tracker.CreatedAt.UnixMilli(), tracker.UpdatedAt.UnixMilli(),
	)

	return err
}

// DeleteResendTracker removes a resend tracker by email hash.
func (r *PasswordResetRepository) DeleteResendTracker(ctx context.Context, emailHash string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM password_reset_resend_tracker WHERE email_hash = ?`, emailHash)
	return err
}

// CleanupResendTrackers removes old resend trackers.
func (r *PasswordResetRepository) CleanupResendTrackers(ctx context.Context, maxAgeHours int) (int64, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(maxAgeHours) * time.Hour).UnixMilli()

	result, err := r.db.ExecContext(ctx,
		`DELETE FROM password_reset_resend_tracker WHERE updated_at < ? AND lockout_until IS NULL`,
		cutoff,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
