package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// currentTimeMillis returns the current time in Unix milliseconds.
func currentTimeMillis() int64 {
	return time.Now().UnixMilli()
}

// getMonthStartMillis returns the start of the current month in Unix milliseconds.
func getMonthStartMillis(nowMillis int64) int64 {
	now := time.UnixMilli(nowMillis)
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return startOfMonth.UnixMilli()
}

// APIKey represents an API key in the system.
type APIKey struct {
	LastRequest   *int64
	ExpiresAt     *int64
	RevokedAt     *int64
	OperatorID    string
	Name          string
	KeyPrefix     string
	KeyHash       string
	SigningSecret string
	Scope         string
	ID            string
	RequestCount  int64
	CreatedAt     int64
	UpdatedAt     int64
	IsActive      bool
}

// APIKeyWithOperator joins an API key with its operator's display name.
// Used by ListAllWithOperator for the super-admin "all keys" view.
type APIKeyWithOperator struct {
	OperatorName string
	APIKey
}

// OperatorRequestTotal is a per-operator cumulative request-count aggregate
// used by the super-admin global stats "top operators" view.
type OperatorRequestTotal struct {
	OperatorID    string
	OperatorName  string
	TotalRequests int64
}

// APIKeyRepository defines the interface for API key operations.
type APIKeyRepository interface {
	Create(ctx context.Context, key *APIKey) error
	GetByID(ctx context.Context, id string) (*APIKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*APIKey, error)
	ListByOperator(ctx context.Context, operatorID string, limit, offset int) ([]*APIKey, int, error)
	ListAll(ctx context.Context, limit, offset int) ([]*APIKey, int, error)
	// ListAllWithOperator lists all API keys joined with the operator's display
	// name, for the super-admin "all keys" view that needs operator identity.
	// operatorID and search are optional filters (empty = unfiltered).
	ListAllWithOperator(ctx context.Context, limit, offset int, operatorID, search string) ([]APIKeyWithOperator, int, error)
	Update(ctx context.Context, key *APIKey) error
	Revoke(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	CountByOperatorThisMonth(ctx context.Context, operatorID string) (int, error)
	CountAll(ctx context.Context) (int, error)
	CountAllIncludingRevoked(ctx context.Context) (int, error)
	CountRevoked(ctx context.Context) (int, error)
	// SumRequestsByScope returns the cumulative request totals grouped by scope.
	SumRequestsByScope(ctx context.Context) (map[string]int64, error)
	// TopOperatorsByRequests returns the top operators by cumulative request
	// count, joined with the operator's display name, limited to `limit` rows.
	TopOperatorsByRequests(ctx context.Context, limit int) ([]OperatorRequestTotal, int, error)
	IncrementRequestCount(ctx context.Context, id string) error
	ExistsByOperatorAndName(ctx context.Context, operatorID, name string) (bool, error)
	ExistsByOperatorAndNameExcluding(ctx context.Context, operatorID, name, excludeKeyID string) (bool, error)
}

// APIKeyRepositoryImpl handles API key persistence.
type APIKeyRepositoryImpl struct {
	db *sql.DB
}

// NewAPIKeyRepository creates a new APIKeyRepository.
func NewAPIKeyRepository(db *sql.DB) APIKeyRepository {
	return &APIKeyRepositoryImpl{db: db}
}

// Create inserts a new API key.
func (r *APIKeyRepositoryImpl) Create(ctx context.Context, key *APIKey) error {
	query := `
		INSERT INTO api_keys (id, operator_id, name, key_prefix, key_hash, signing_secret, scope, expires_at, is_active, request_count, last_request_at, created_at, updated_at, revoked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		key.ID, key.OperatorID, key.Name, key.KeyPrefix, key.KeyHash,
		key.SigningSecret, key.Scope, key.ExpiresAt, key.IsActive, key.RequestCount,
		key.LastRequest, key.CreatedAt, key.UpdatedAt, key.RevokedAt,
	)
	return err
}

// GetByID retrieves an API key by its ID.
func (r *APIKeyRepositoryImpl) GetByID(ctx context.Context, id string) (*APIKey, error) {
	query := `
		SELECT id, operator_id, name, key_prefix, key_hash, signing_secret, scope, expires_at, is_active, request_count, last_request_at, created_at, updated_at, revoked_at
		FROM api_keys WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanAPIKey(row)
}

// GetByKeyHash retrieves an API key by its hash.
func (r *APIKeyRepositoryImpl) GetByKeyHash(ctx context.Context, keyHash string) (*APIKey, error) {
	query := `
		SELECT id, operator_id, name, key_prefix, key_hash, signing_secret, scope, expires_at, is_active, request_count, last_request_at, created_at, updated_at, revoked_at
		FROM api_keys WHERE key_hash = ? AND is_active = 1
	`
	row := r.db.QueryRowContext(ctx, query, keyHash)
	return r.scanAPIKey(row)
}

// ListByOperator retrieves all API keys for an operator with pagination.
func (r *APIKeyRepositoryImpl) ListByOperator(ctx context.Context, operatorID string, limit, offset int) ([]*APIKey, int, error) {
	// Get total count.
	var total int
	countQuery := `SELECT COUNT(*) FROM api_keys WHERE operator_id = ?`
	if err := r.db.QueryRowContext(ctx, countQuery, operatorID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results.
	query := `
		SELECT id, operator_id, name, key_prefix, key_hash, signing_secret, scope, expires_at, is_active, request_count, last_request_at, created_at, updated_at, revoked_at
		FROM api_keys WHERE operator_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, operatorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var keys []*APIKey
	for rows.Next() {
		key, err := r.scanAPIKeyFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		keys = append(keys, key)
	}

	return keys, total, rows.Err()
}

// Update updates an API key.
func (r *APIKeyRepositoryImpl) Update(ctx context.Context, key *APIKey) error {
	query := `
		UPDATE api_keys 
		SET name = ?, scope = ?, expires_at = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, key.Name, key.Scope, key.ExpiresAt, key.UpdatedAt, key.ID)
	return err
}

// Revoke marks an API key as revoked.
func (r *APIKeyRepositoryImpl) Revoke(ctx context.Context, id string) error {
	query := `
		UPDATE api_keys 
		SET is_active = 0, revoked_at = ?, updated_at = ?
		WHERE id = ?
	`
	now := currentTimeMillis()
	_, err := r.db.ExecContext(ctx, query, now, now, id)
	return err
}

// Delete removes an API key.
func (r *APIKeyRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM api_keys WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// CountByOperatorThisMonth counts all API keys created by an operator this month (including revoked).
// This enforces the "20 keys per operator per month" limit properly, preventing bypass via rotation.
func (r *APIKeyRepositoryImpl) CountByOperatorThisMonth(ctx context.Context, operatorID string) (int, error) {
	// Get start of current month.
	now := currentTimeMillis()
	monthStart := getMonthStartMillis(now)

	// Count ALL keys including revoked - this prevents circumvention via rotation.
	query := `SELECT COUNT(*) FROM api_keys WHERE operator_id = ? AND created_at >= ?`
	var count int
	if err := r.db.QueryRowContext(ctx, query, operatorID, monthStart).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// IncrementRequestCount increments the request counter for an API key.
func (r *APIKeyRepositoryImpl) IncrementRequestCount(ctx context.Context, id string) error {
	query := `
		UPDATE api_keys 
		SET request_count = request_count + 1, last_request_at = ?, updated_at = ?
		WHERE id = ?
	`
	now := currentTimeMillis()
	_, err := r.db.ExecContext(ctx, query, now, now, id)
	return err
}

// scanAPIKey scans a row into an APIKey struct.
func (r *APIKeyRepositoryImpl) scanAPIKey(row *sql.Row) (*APIKey, error) {
	var key APIKey
	var expiresAt, lastRequest, revokedAt sql.NullInt64

	err := row.Scan(
		&key.ID, &key.OperatorID, &key.Name, &key.KeyPrefix, &key.KeyHash,
		&key.SigningSecret, &key.Scope, &expiresAt, &key.IsActive, &key.RequestCount,
		&lastRequest, &key.CreatedAt, &key.UpdatedAt, &revokedAt,
	)
	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Int64
	}
	if lastRequest.Valid {
		key.LastRequest = &lastRequest.Int64
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Int64
	}

	return &key, nil
}

// scanAPIKeyFromRows scans rows into an APIKey struct.
func (r *APIKeyRepositoryImpl) scanAPIKeyFromRows(rows *sql.Rows) (*APIKey, error) {
	var key APIKey
	var expiresAt, lastRequest, revokedAt sql.NullInt64

	err := rows.Scan(
		&key.ID, &key.OperatorID, &key.Name, &key.KeyPrefix, &key.KeyHash,
		&key.SigningSecret, &key.Scope, &expiresAt, &key.IsActive, &key.RequestCount,
		&lastRequest, &key.CreatedAt, &key.UpdatedAt, &revokedAt,
	)
	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Int64
	}
	if lastRequest.Valid {
		key.LastRequest = &lastRequest.Int64
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Int64
	}

	return &key, nil
}

// ListAll returns all API keys across all operators with pagination.
func (r *APIKeyRepositoryImpl) ListAll(ctx context.Context, limit, offset int) ([]*APIKey, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Get total count.
	var total int
	countQuery := `SELECT COUNT(*) FROM api_keys WHERE is_active = 1`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count all keys: %w", err)
	}

	// Get keys.
	query := `
		SELECT id, operator_id, name, key_prefix, key_hash, signing_secret, scope, expires_at,
		       is_active, request_count, last_request_at, created_at, updated_at, revoked_at
		FROM api_keys
		WHERE is_active = 1
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all keys: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var keys []*APIKey
	for rows.Next() {
		key, err := r.scanAPIKeyFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		keys = append(keys, key)
	}

	return keys, total, nil
}


// ListAllWithOperator lists all API keys joined with the owning operator's
// display name. Used by the super-admin "all keys" view, which must show which
// operator owns each key — information the plain ListAll/ToResponse path drops.
// operatorID (non-empty) filters to a single operator; search (non-empty) does
// a case-insensitive LIKE match on key name, prefix, and operator name/email.
func (r *APIKeyRepositoryImpl) ListAllWithOperator(ctx context.Context, limit, offset int, operatorID, search string) ([]APIKeyWithOperator, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Build the WHERE clause dynamically. We always restrict to active keys for
	// the admin list view (revoked keys are excluded); filters are appended.
	where := "WHERE k.is_active = 1"
	var args []interface{}
	if operatorID != "" {
		where += " AND k.operator_id = ?"
		args = append(args, operatorID)
	}
	if search != "" {
		where += " AND (k.name LIKE ? OR k.key_prefix LIKE ? OR o.name LIKE ? OR o.email LIKE ?)"
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM api_keys k LEFT JOIN operators o ON o.id = k.operator_id " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count all keys: %w", err)
	}

	query := `
		SELECT k.id, k.operator_id, k.name, k.key_prefix, k.key_hash, k.signing_secret, k.scope,
		       k.expires_at, k.is_active, k.request_count, k.last_request_at,
		       k.created_at, k.updated_at, k.revoked_at, o.name
		FROM api_keys k
		LEFT JOIN operators o ON o.id = k.operator_id
	` + where + `
		ORDER BY k.created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all keys with operator: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []APIKeyWithOperator
	for rows.Next() {
		var item APIKeyWithOperator
		var expiresAt, lastRequest, revokedAt sql.NullInt64
		var operatorName sql.NullString
		if err := rows.Scan(
			&item.ID, &item.OperatorID, &item.Name, &item.KeyPrefix, &item.KeyHash,
			&item.SigningSecret, &item.Scope, &expiresAt, &item.IsActive, &item.RequestCount,
			&lastRequest, &item.CreatedAt, &item.UpdatedAt, &revokedAt, &operatorName,
		); err != nil {
			return nil, 0, err
		}
		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Int64
		}
		if lastRequest.Valid {
			item.LastRequest = &lastRequest.Int64
		}
		if revokedAt.Valid {
			item.RevokedAt = &revokedAt.Int64
		}
		item.OperatorName = operatorName.String
		results = append(results, item)
	}

	return results, total, rows.Err()
}

// CountAll returns the total number of active API keys across all operators.
func (r *APIKeyRepositoryImpl) CountAll(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM api_keys WHERE is_active = 1`
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count all keys: %w", err)
	}
	return count, nil
}

// CountAllIncludingRevoked returns the total number of API keys (active + revoked).
func (r *APIKeyRepositoryImpl) CountAllIncludingRevoked(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM api_keys`
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count all keys including revoked: %w", err)
	}
	return count, nil
}

// CountRevoked returns the number of revoked API keys.
func (r *APIKeyRepositoryImpl) CountRevoked(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM api_keys WHERE is_active = 0`
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count revoked keys: %w", err)
	}
	return count, nil
}

// SumRequestsByScope returns cumulative request totals grouped by scope.
// Returns a map keyed by scope string ("read"/"write"/"admin").
func (r *APIKeyRepositoryImpl) SumRequestsByScope(ctx context.Context) (map[string]int64, error) {
	query := `SELECT scope, COALESCE(SUM(request_count), 0) FROM api_keys GROUP BY scope`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to sum requests by scope: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]int64)
	for rows.Next() {
		var scope string
		var sum int64
		if err := rows.Scan(&scope, &sum); err != nil {
			return nil, err
		}
		result[scope] = sum
	}
	return result, rows.Err()
}

// TopOperatorsByRequests returns the top operators by cumulative request count,
// joined with their display name. limit caps the result set; the second return
// is the total number of distinct operators that have at least one key.
func (r *APIKeyRepositoryImpl) TopOperatorsByRequests(ctx context.Context, limit int) ([]OperatorRequestTotal, int, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	var totalOperators int
	countQuery := `SELECT COUNT(DISTINCT operator_id) FROM api_keys`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalOperators); err != nil {
		return nil, 0, fmt.Errorf("failed to count distinct operators: %w", err)
	}

	query := `
		SELECT k.operator_id, COALESCE(o.name, ''), COALESCE(SUM(k.request_count), 0)
		FROM api_keys k
		LEFT JOIN operators o ON o.id = k.operator_id
		GROUP BY k.operator_id
		ORDER BY COALESCE(SUM(k.request_count), 0) DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query top operators: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []OperatorRequestTotal
	for rows.Next() {
		var item OperatorRequestTotal
		if err := rows.Scan(&item.OperatorID, &item.OperatorName, &item.TotalRequests); err != nil {
			return nil, 0, err
		}
		results = append(results, item)
	}
	return results, totalOperators, rows.Err()
}

// ExistsByOperatorAndName checks if an active API key with the given name exists for the operator.
func (r *APIKeyRepositoryImpl) ExistsByOperatorAndName(ctx context.Context, operatorID, name string) (bool, error) {
	query := `SELECT COUNT(*) FROM api_keys WHERE operator_id = ? AND name = ? AND is_active = 1`
	var count int
	if err := r.db.QueryRowContext(ctx, query, operatorID, name).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// ExistsByOperatorAndNameExcluding checks if an active API key with the given name exists for the operator, excluding a specific key ID.
func (r *APIKeyRepositoryImpl) ExistsByOperatorAndNameExcluding(ctx context.Context, operatorID, name, excludeKeyID string) (bool, error) {
	query := `SELECT COUNT(*) FROM api_keys WHERE operator_id = ? AND name = ? AND is_active = 1 AND id != ?`
	var count int
	if err := r.db.QueryRowContext(ctx, query, operatorID, name, excludeKeyID).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// SetupAPIKeysTable creates the api_keys table if it doesn't exist.
func SetupAPIKeysTable(db *sql.DB) error {
	// Create api_keys table.
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			operator_id TEXT NOT NULL,
			name TEXT NOT NULL,
			key_prefix TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			signing_secret TEXT NOT NULL DEFAULT '',
			scope TEXT NOT NULL DEFAULT 'read',
			expires_at INTEGER,
			is_active INTEGER NOT NULL DEFAULT 1,
			request_count INTEGER NOT NULL DEFAULT 0,
			last_request_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			revoked_at INTEGER,
			FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create indexes.
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_api_keys_operator_id ON api_keys(operator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_created_at ON api_keys(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active)`,
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
