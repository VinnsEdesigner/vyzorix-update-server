<<<<<<< HEAD

=======
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/confirmation"
)

// Ensure ConfirmationRepository implements confirmation.Repository.
var _ confirmation.Repository = (*ConfirmationRepository)(nil)

// ConfirmationRepository implements confirmation.Repository using SQLite.
type ConfirmationRepository struct {
	db *sql.DB
}

// NewConfirmationRepository creates a new ConfirmationRepository.
func NewConfirmationRepository(db *sql.DB) *ConfirmationRepository {
	return &ConfirmationRepository{db: db}
}

// Create stores a new pending confirmation.
func (r *ConfirmationRepository) Create(ctx context.Context, c *confirmation.PendingConfirmation) error {
	var consumedAt any
	if c.ConsumedAt != nil {
		consumedAt = *c.ConsumedAt
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO command_confirmations
			(token, operator_id, org_id, command, device_id, risk_tier, created_at, expires_at, consumed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Token, c.OperatorID, c.OrgID, c.Command, c.DeviceID, c.RiskTier,
		c.CreatedAt, c.ExpiresAt, consumedAt,
	)
	return err
}

// Get retrieves a pending confirmation by its token.
func (r *ConfirmationRepository) Get(ctx context.Context, token string) (*confirmation.PendingConfirmation, error) {
	var c confirmation.PendingConfirmation
	var consumedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT token, operator_id, org_id, command, device_id, risk_tier, created_at, expires_at, consumed_at
		FROM command_confirmations WHERE token = ?`, token).Scan(
		&c.Token, &c.OperatorID, &c.OrgID, &c.Command, &c.DeviceID, &c.RiskTier,
		&c.CreatedAt, &c.ExpiresAt, &consumedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, confirmation.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if consumedAt.Valid {
		t := consumedAt.Time
		c.ConsumedAt = &t
	}
	return &c, nil
}

// Consume atomically marks a confirmation as consumed at the given time. It
// only updates rows that are not yet consumed and not yet expired, returning
// confirmation.ErrAlreadyConsumed or confirmation.ErrExpired when the
// preconditions fail.
func (r *ConfirmationRepository) Consume(ctx context.Context, token string, at time.Time) (*confirmation.PendingConfirmation, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE command_confirmations
		SET consumed_at = ?
		WHERE token = ? AND consumed_at IS NULL AND expires_at > ?`,
		at, token, at)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		// Distinguish already-consumed from expired/not-found via a lookup.
		c, getErr := r.Get(ctx, token)
		if errors.Is(getErr, confirmation.ErrNotFound) {
			return nil, confirmation.ErrNotFound
		}
		if getErr != nil {
			return nil, getErr
		}
		if c.IsConsumed() {
			return c, confirmation.ErrAlreadyConsumed
		}
		return c, confirmation.ErrExpired
	}
	return r.Get(ctx, token)
}

// DeleteExpired removes expired confirmations.
func (r *ConfirmationRepository) DeleteExpired(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM command_confirmations WHERE expires_at <= ?`, time.Now())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
<<<<<<< HEAD
}
=======
}
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
