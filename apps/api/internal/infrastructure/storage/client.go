package storage

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/client"
)

// Helper functions for client repository

// NewUUIDv7 generates a UUIDv7 string.
func NewUUIDv7() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	// Set version 7 (UUIDv7)
	b[6] = (b[6] & 0x0f) | 0x70
	// Set variant 10
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b)
}

// readRandom reads random bytes.
func readRandom(b []byte) (int, error) {
	return rand.Read(b)
}

// hexEncode encodes bytes to hex string.
func hexEncode(b []byte) string {
	return hex.EncodeToString(b)
}

// hashSHA512 computes SHA512 hash.
func hashSHA512(input string) string {
	h := sha512.Sum512([]byte(input))
	return hex.EncodeToString(h[:])
}

// hmacSHA256 computes HMAC-SHA256.
func hmacSHA256(key, data string) string {
	h := hmac.New(sha512.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Ensure ClientRepository implements client.Repository.
var _ client.Repository = (*ClientRepository)(nil)

// ClientRepository implements client.Repository using SQLite.
type ClientRepository struct {
	db *sql.DB
}

// NewClientRepository creates a new ClientRepository.
func NewClientRepository(db *sql.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

// FindByID retrieves a client by ID.
func (r *ClientRepository) FindByID(ctx context.Context, id string) (*client.Client, error) {
	query := `SELECT id, operator_id, name, platform, client_secret_hash, hmac_key, 
		allowed_origins, allowed_paths, rate_limit, is_active, request_count, 
		last_request_at, created_at, updated_at 
		FROM api_clients WHERE id = ?`

	var c client.Client
	var allowedOrigins, allowedPaths sql.NullString
	var lastRequestAt sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.OperatorID, &c.Name, &c.Platform, &c.ClientSecretHash, &c.HmacKey,
		&allowedOrigins, &allowedPaths, &c.RateLimit, &c.IsActive, &c.RequestCount,
		&lastRequestAt, &c.CreatedAt, &c.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, client.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if allowedOrigins.Valid {
		_ = json.Unmarshal([]byte(allowedOrigins.String), &c.AllowedOrigins)
	}
	if allowedPaths.Valid {
		_ = json.Unmarshal([]byte(allowedPaths.String), &c.AllowedPaths)
	}
	if lastRequestAt.Valid {
		c.LastRequestAt = &lastRequestAt.Int64
	}

	return &c, nil
}

// FindByOperatorID retrieves all clients for an operator.
func (r *ClientRepository) FindByOperatorID(ctx context.Context, operatorID string) ([]*client.Client, error) {
	query := `SELECT id, operator_id, name, platform, client_secret_hash, hmac_key, 
		allowed_origins, allowed_paths, rate_limit, is_active, request_count, 
		last_request_at, created_at, updated_at 
		FROM api_clients WHERE operator_id = ?`

	rows, err := r.db.QueryContext(ctx, query, operatorID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var clients []*client.Client
	for rows.Next() {
		var c client.Client
		var allowedOrigins, allowedPaths sql.NullString
		var lastRequestAt sql.NullInt64

		if err := rows.Scan(
			&c.ID, &c.OperatorID, &c.Name, &c.Platform, &c.ClientSecretHash, &c.HmacKey,
			&allowedOrigins, &allowedPaths, &c.RateLimit, &c.IsActive, &c.RequestCount,
			&lastRequestAt, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if allowedOrigins.Valid {
			_ = json.Unmarshal([]byte(allowedOrigins.String), &c.AllowedOrigins)
		}
		if allowedPaths.Valid {
			_ = json.Unmarshal([]byte(allowedPaths.String), &c.AllowedPaths)
		}
		if lastRequestAt.Valid {
			c.LastRequestAt = &lastRequestAt.Int64
		}

		clients = append(clients, &c)
	}

	return clients, rows.Err()
}

// FindAll retrieves all clients with pagination (admin use).
func (r *ClientRepository) FindAll(ctx context.Context, limit, offset int) ([]*client.Client, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM api_clients`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, operator_id, name, platform, client_secret_hash, hmac_key, 
		allowed_origins, allowed_paths, rate_limit, is_active, request_count, 
		last_request_at, created_at, updated_at 
		FROM api_clients ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var clients []*client.Client
	for rows.Next() {
		var c client.Client
		var allowedOrigins, allowedPaths sql.NullString
		var lastRequestAt sql.NullInt64

		if err := rows.Scan(
			&c.ID, &c.OperatorID, &c.Name, &c.Platform, &c.ClientSecretHash, &c.HmacKey,
			&allowedOrigins, &allowedPaths, &c.RateLimit, &c.IsActive, &c.RequestCount,
			&lastRequestAt, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if allowedOrigins.Valid {
			_ = json.Unmarshal([]byte(allowedOrigins.String), &c.AllowedOrigins)
		}
		if allowedPaths.Valid {
			_ = json.Unmarshal([]byte(allowedPaths.String), &c.AllowedPaths)
		}
		if lastRequestAt.Valid {
			c.LastRequestAt = &lastRequestAt.Int64
		}

		clients = append(clients, &c)
	}

	return clients, total, rows.Err()
}

// Create creates a new client.
func (r *ClientRepository) Create(ctx context.Context, c *client.Client, secret string) (*client.Client, string, error) {
	originsJSON, err := json.Marshal(c.AllowedOrigins)
	if err != nil {
		return nil, "", err
	}
	pathsJSON, err := json.Marshal(c.AllowedPaths)
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now

	query := `INSERT INTO api_clients 
		(id, operator_id, name, platform, client_secret_hash, hmac_key, 
		allowed_origins, allowed_paths, rate_limit, is_active, request_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		c.ID, c.OperatorID, c.Name, c.Platform, c.ClientSecretHash, c.HmacKey,
		string(originsJSON), string(pathsJSON), c.RateLimit, c.IsActive, 0,
		c.CreatedAt.UnixMilli(), c.UpdatedAt.UnixMilli(),
	)
	if err != nil {
		return nil, "", err
	}

	return c, secret, nil
}

// Update updates a client.
func (r *ClientRepository) Update(ctx context.Context, c *client.Client) error {
	originsJSON, err := json.Marshal(c.AllowedOrigins)
	if err != nil {
		return err
	}
	pathsJSON, err := json.Marshal(c.AllowedPaths)
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()

	query := `UPDATE api_clients SET 
		name = ?, allowed_origins = ?, allowed_paths = ?, rate_limit = ?, 
		is_active = ?, updated_at = ?
		WHERE id = ?`

	_, err = r.db.ExecContext(ctx, query,
		c.Name, string(originsJSON), string(pathsJSON), c.RateLimit,
		c.IsActive, now, c.ID,
	)
	return err
}

// Delete deletes a client.
func (r *ClientRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM api_clients WHERE id = ?`, id)
	return err
}

// RotateSigningKey rotates the signing key with a grace period.
func (r *ClientRepository) RotateSigningKey(ctx context.Context, clientID string, gracePeriod time.Duration) (*client.SigningKey, string, error) {
	// Generate key ID
	keyID := NewUUIDv7()

	// Generate the actual key
	keyBytes := make([]byte, 32)
	if _, err := readRandom(keyBytes); err != nil {
		return nil, "", err
	}
	key := hexEncode(keyBytes)

	// Hash the key for storage
	keyHash := hashSHA512(key)

	// Get current max version
	var maxVersion int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM signing_keys WHERE client_id = ?
	`, clientID).Scan(&maxVersion)
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UnixMilli()
	gracePeriodMs := gracePeriod.Milliseconds()
	var expiresAt *int64
	if gracePeriodMs > 0 {
		expiresAtVal := now + gracePeriodMs
		expiresAt = &expiresAtVal
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO signing_keys (id, client_id, key_hash, version, issued_at, expires_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, keyID, clientID, keyHash, maxVersion+1, now, expiresAt, true)
	if err != nil {
		return nil, "", err
	}

	// Deactivate old keys
	_, _ = r.db.ExecContext(ctx, `
		UPDATE signing_keys SET is_active = 0 WHERE client_id = ? AND id != ?
	`, clientID, keyID)

	return &client.SigningKey{
		ID:        keyID,
		ClientID:  clientID,
		KeyHash:   keyHash,
		Version:   maxVersion + 1,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
		IsActive:  true,
	}, key, nil
}

// ValidateSigningKey validates a signing key.
func (r *ClientRepository) ValidateSigningKey(ctx context.Context, clientID, signature, payload, timestamp string) error {
	now := time.Now().UnixMilli()

	var key client.SigningKey
	var expiresAt sql.NullInt64

	err := r.db.QueryRowContext(ctx, `
		SELECT id, client_id, key_hash, version, issued_at, expires_at, is_active
		FROM signing_keys
		WHERE client_id = ? AND is_active = 1 AND (expires_at IS NULL OR expires_at > ?)
		ORDER BY version DESC LIMIT 1
	`, clientID, now).Scan(
		&key.ID, &key.ClientID, &key.KeyHash, &key.Version,
		&key.IssuedAt, &expiresAt, &key.IsActive,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return client.ErrSigningKeyNotFound
	}
	if err != nil {
		return err
	}

	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Int64
	}

	// Verify signature using HMAC-SHA256
	expectedSig := hmacSHA256(key.KeyHash, timestamp+payload)
	if expectedSig != signature {
		return errors.New("invalid signature")
	}

	return nil
}

// GetHmacKey retrieves the HMAC signing key for a client.
func (r *ClientRepository) GetHmacKey(ctx context.Context, clientID string) (string, bool) {
	query := `SELECT hmac_key, is_active FROM api_clients WHERE id = ?`
	
	var hmacKey string
	var isActive bool
	
	err := r.db.QueryRowContext(ctx, query, clientID).Scan(&hmacKey, &isActive)
	if err != nil || !isActive {
		return "", false
	}
	
	return hmacKey, true
}
