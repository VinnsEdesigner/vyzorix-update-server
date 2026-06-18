// Package storage provides SQLite database operations.
package storage

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"
)

// HashSHA512 returns the SHA512 hash of the input string.
func HashSHA512(input string) string {
	hash := sha512.Sum512([]byte(input))
	return hex.EncodeToString(hash[:])
}

// APIClient represents an API client for request signing.
type APIClient struct {
	ID                string    `json:"id"`              // UUIDv7
	OperatorID        string    `json:"operatorId"`      // Owner operator
	Name              string    `json:"name"`            // Client name
	Platform          string    `json:"platform"`        // web, ios, android
	ClientSecretHash  string    `json:"-"`               // Argon2id hash, never exposed
	HmacKey          string    `json:"-"`               // HMAC verification key (SHA512 of secret), never exposed
	AllowedOrigins   []string  `json:"allowedOrigins"`  // JSON array
	AllowedPaths     []string  `json:"allowedPaths"`    // JSON array
	RateLimit        int       `json:"rateLimit"`       // Requests per minute
	IsActive         bool      `json:"isActive"`        // Can be deactivated
	RequestCount     int64     `json:"requestCount"`    // Total requests made
	LastRequestAt    *int64    `json:"lastRequestAt"`   // Timestamp
	CreatedAt        int64     `json:"createdAt"`       // Timestamp
	UpdatedAt        int64     `json:"updatedAt"`       // Timestamp
}

// SigningKey represents a signing key version for a client.
type SigningKey struct {
	ID        string    `json:"id"`         // UUIDv7
	ClientID  string    `json:"clientId"`   // APIClient reference
	KeyHash   string    `json:"-"`          // SHA512 of key for lookup
	Version   int       `json:"version"`    // Key version number
	IssuedAt  int64     `json:"issuedAt"`   // When key was issued
	ExpiresAt *int64    `json:"expiresAt"`  // When key expires (nil = never)
	IsActive  bool      `json:"isActive"`   // Key is currently valid
	RevokedAt *int64    `json:"revokedAt"`  // When key was revoked (nil = not revoked)
}

// DeriveHmacKey derives an HMAC verification key from the client secret.
// This is a one-way deterministic derivation: SHA512(secret).
// Both server and client can compute this key.
func DeriveHmacKey(clientSecret string) string {
	hash := sha512.Sum512([]byte(clientSecret))
	return hex.EncodeToString(hash[:])
}

// VerifyHmacKey verifies a client secret against the stored HMAC key.
func VerifyHmacKey(clientSecret, storedHmacKey string) bool {
	derivedKey := DeriveHmacKey(clientSecret)
	return hmac.Equal([]byte(derivedKey), []byte(storedHmacKey))
}

// CreateAPIClient creates a new API client and returns the client with plaintext secret.
// The plaintext secret is ONLY available at creation time and cannot be retrieved later.
func (s *Store) CreateAPIClient(ctx context.Context, req CreateAPIClientRequest) (*APIClient, string, error) {
	// Generate client ID (UUIDv7)
	clientID := NewUUIDv7()

	// Generate client secret (32 bytes = 64 hex chars)
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, "", err
	}
	clientSecret := hex.EncodeToString(secretBytes)

	// Hash the secret with Argon2id for authentication
	secretHash, err := HashPassword(clientSecret)
	if err != nil {
		return nil, "", err
	}

	// Derive HMAC key for request signing verification
	hmacKey := DeriveHmacKey(clientSecret)

	// Serialize allowed origins/paths to JSON
	originsJSON, err := json.Marshal(req.AllowedOrigins)
	if err != nil {
		return nil, "", err
	}
	pathsJSON, err := json.Marshal(req.AllowedPaths)
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC().UnixMilli()

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO api_clients (id, operator_id, name, platform, client_secret_hash, hmac_key, allowed_origins, allowed_paths, rate_limit, is_active, request_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, clientID, req.OperatorID, req.Name, req.Platform, secretHash, hmacKey, string(originsJSON), string(pathsJSON), req.RateLimit, 1, 0, now, now)
	if err != nil {
		return nil, "", err
	}

	client := &APIClient{
		ID:                clientID,
		OperatorID:        req.OperatorID,
		Name:              req.Name,
		Platform:          req.Platform,
		ClientSecretHash:  secretHash,
		HmacKey:          hmacKey,
		AllowedOrigins:   req.AllowedOrigins,
		AllowedPaths:     req.AllowedPaths,
		RateLimit:        req.RateLimit,
		IsActive:         true,
		RequestCount:     0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	return client, clientSecret, nil
}

// CreateAPIClientRequest is the request to create a new API client.
type CreateAPIClientRequest struct {
	OperatorID     string   `json:"operatorId"`
	Name           string   `json:"name"`
	Platform       string   `json:"platform"` // web, ios, android
	AllowedOrigins []string `json:"allowedOrigins"`
	AllowedPaths   []string `json:"allowedPaths"`
	RateLimit      int      `json:"rateLimit"`
}

// GetAPIClient retrieves an API client by ID.
func (s *Store) GetAPIClient(ctx context.Context, clientID string) (*APIClient, error) {
	var client APIClient
	var allowedOrigins, allowedPaths string
	var lastRequestAt sql.NullInt64

	err := s.db.QueryRowContext(ctx, `
		SELECT id, operator_id, name, platform, client_secret_hash, hmac_key, allowed_origins, allowed_paths, rate_limit, is_active, request_count, last_request_at, created_at, updated_at
		FROM api_clients WHERE id = ?
	`, clientID).Scan(
		&client.ID, &client.OperatorID, &client.Name, &client.Platform, &client.ClientSecretHash, &client.HmacKey,
		&allowedOrigins, &allowedPaths, &client.RateLimit, &client.IsActive, &client.RequestCount,
		&lastRequestAt, &client.CreatedAt, &client.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(allowedOrigins), &client.AllowedOrigins); err != nil {
		client.AllowedOrigins = []string{}
	}
	if err := json.Unmarshal([]byte(allowedPaths), &client.AllowedPaths); err != nil {
		client.AllowedPaths = []string{}
	}
	if lastRequestAt.Valid {
		client.LastRequestAt = &lastRequestAt.Int64
	}

	return &client, nil
}

// GetAPIClientByOperator retrieves all API clients for an operator.
func (s *Store) GetAPIClientByOperator(ctx context.Context, operatorID string) ([]*APIClient, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, operator_id, name, platform, allowed_origins, allowed_paths, rate_limit, is_active, request_count, last_request_at, created_at, updated_at
		FROM api_clients WHERE operator_id = ?
	`, operatorID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var clients []*APIClient
	for rows.Next() {
		var client APIClient
		var allowedOrigins, allowedPaths string
		var lastRequestAt sql.NullInt64

		if err := rows.Scan(
			&client.ID, &client.OperatorID, &client.Name, &client.Platform,
			&allowedOrigins, &allowedPaths, &client.RateLimit, &client.IsActive,
			&client.RequestCount, &lastRequestAt, &client.CreatedAt, &client.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(allowedOrigins), &client.AllowedOrigins); err != nil {
			client.AllowedOrigins = []string{}
		}
		if err := json.Unmarshal([]byte(allowedPaths), &client.AllowedPaths); err != nil {
			client.AllowedPaths = []string{}
		}
		if lastRequestAt.Valid {
			client.LastRequestAt = &lastRequestAt.Int64
		}

		clients = append(clients, &client)
	}

	return clients, rows.Err()
}

// ListAllAPIClients returns all API clients (admin use).
func (s *Store) ListAllAPIClients(ctx context.Context) ([]*APIClient, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, operator_id, name, platform, allowed_origins, allowed_paths, rate_limit, is_active, request_count, last_request_at, created_at, updated_at
		FROM api_clients ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var clients []*APIClient
	for rows.Next() {
		var client APIClient
		var allowedOrigins, allowedPaths string
		var lastRequestAt sql.NullInt64

		if err := rows.Scan(
			&client.ID, &client.OperatorID, &client.Name, &client.Platform,
			&allowedOrigins, &allowedPaths, &client.RateLimit, &client.IsActive,
			&client.RequestCount, &lastRequestAt, &client.CreatedAt, &client.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(allowedOrigins), &client.AllowedOrigins); err != nil {
			client.AllowedOrigins = []string{}
		}
		if err := json.Unmarshal([]byte(allowedPaths), &client.AllowedPaths); err != nil {
			client.AllowedPaths = []string{}
		}
		if lastRequestAt.Valid {
			client.LastRequestAt = &lastRequestAt.Int64
		}

		clients = append(clients, &client)
	}

	return clients, rows.Err()
}

// UpdateAPIClient updates an API client.
func (s *Store) UpdateAPIClient(ctx context.Context, client *APIClient) error {
	originsJSON, err0 := json.Marshal(client.AllowedOrigins)
	pathsJSON, err1 := json.Marshal(client.AllowedPaths)
	if err0 != nil || err1 != nil {
		// Best effort - continue with empty JSON on error
		originsJSON = []byte("[]")
		pathsJSON = []byte("[]")
	}
	now := time.Now().UTC().UnixMilli()

	_, err := s.db.ExecContext(ctx, `
		UPDATE api_clients SET 
			name = ?, allowed_origins = ?, allowed_paths = ?, rate_limit = ?, 
			is_active = ?, updated_at = ?
		WHERE id = ?
	`, client.Name, string(originsJSON), string(pathsJSON), client.RateLimit, 
		client.IsActive, now, client.ID)
	return err
}

// VerifyAPIClientSecret verifies a client secret against the stored hash.
// Returns the client if valid, nil if invalid.
func (s *Store) VerifyAPIClientSecret(ctx context.Context, clientID, clientSecret string) (*APIClient, error) {
	client, err := s.GetAPIClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if !client.IsActive {
		return nil, ErrNotFound
	}

	if err := VerifyPassword(clientSecret, client.ClientSecretHash); err != nil {
		return nil, ErrNotFound
	}

	return client, nil
}

// GetAPIClientHmacKey retrieves the HMAC signing key for a client.
// Returns the derived HMAC key (SHA512 of client secret) and whether the client is active.
// This is used for request signing verification.
func (s *Store) GetAPIClientHmacKey(ctx context.Context, clientID string) (string, bool) {
	client, err := s.GetAPIClient(ctx, clientID)
	if err != nil || !client.IsActive {
		return "", false
	}
	return client.HmacKey, true
}

// UpdateAPIClientActive updates whether an API client is active.
func (s *Store) UpdateAPIClientActive(ctx context.Context, clientID string, isActive bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE api_clients SET is_active = ?, updated_at = ? WHERE id = ?
	`, isActive, time.Now().UTC().UnixMilli(), clientID)
	return err
}

// DeleteAPIClient deletes an API client.
func (s *Store) DeleteAPIClient(ctx context.Context, clientID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM api_clients WHERE id = ?`, clientID)
	return err
}

// IncrementAPIClientRequestCount increments the request count and updates last_request_at.
func (s *Store) IncrementAPIClientRequestCount(ctx context.Context, clientID string) error {
	now := time.Now().UTC().UnixMilli()
	_, err := s.db.ExecContext(ctx, `
		UPDATE api_clients SET request_count = request_count + 1, last_request_at = ?, updated_at = ? WHERE id = ?
	`, now, now, clientID)
	return err
}

// CreateSigningKey creates a new signing key for a client.
func (s *Store) CreateSigningKey(ctx context.Context, clientID string, expiresAt *int64) (*SigningKey, string, error) {
	// Generate key ID (UUIDv7)
	keyID := NewUUIDv7()

	// Generate the actual key (32 bytes = 64 hex chars)
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", err
	}
	key := hex.EncodeToString(keyBytes)

	// Hash the key for storage/lookup
	keyHash := HashSHA512(key)

	// Get current max version for this client
	var maxVersion int
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM signing_keys WHERE client_id = ?
	`, clientID).Scan(&maxVersion)
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC().UnixMilli()

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO signing_keys (id, client_id, key_hash, version, issued_at, expires_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, keyID, clientID, keyHash, maxVersion+1, now, expiresAt, true)
	if err != nil {
		return nil, "", err
	}

	// Deactivate old keys
	_, err = s.db.ExecContext(ctx, `
		UPDATE signing_keys SET is_active = 0 WHERE client_id = ? AND id != ?
	`, clientID, keyID)
	if err != nil {
		return nil, "", err
	}

	return &SigningKey{
		ID:        keyID,
		ClientID:  clientID,
		KeyHash:   keyHash,
		Version:   maxVersion + 1,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
		IsActive:  true,
	}, key, nil
}

// GetSigningKey retrieves an active signing key by client ID.
// Returns the key with the highest version number that is active and not expired.
func (s *Store) GetSigningKey(ctx context.Context, clientID string) (*SigningKey, error) {
	now := time.Now().UTC().UnixMilli()

	var key SigningKey
	var expiresAt, revokedAt sql.NullInt64

	err := s.db.QueryRowContext(ctx, `
		SELECT id, client_id, key_hash, version, issued_at, expires_at, is_active, revoked_at
		FROM signing_keys
		WHERE client_id = ? AND is_active = 1 AND (expires_at IS NULL OR expires_at > ?)
		ORDER BY version DESC LIMIT 1
	`, clientID, now).Scan(
		&key.ID, &key.ClientID, &key.KeyHash, &key.Version,
		&key.IssuedAt, &expiresAt, &key.IsActive, &revokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Int64
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Int64
	}

	return &key, nil
}

// GetSigningKeyByHash retrieves a signing key by its hash.
func (s *Store) GetSigningKeyByHash(ctx context.Context, keyHash string) (*SigningKey, error) {
	now := time.Now().UTC().UnixMilli()

	var key SigningKey
	var expiresAt, revokedAt sql.NullInt64

	err := s.db.QueryRowContext(ctx, `
		SELECT id, client_id, key_hash, version, issued_at, expires_at, is_active, revoked_at
		FROM signing_keys
		WHERE key_hash = ? AND is_active = 1 AND (expires_at IS NULL OR expires_at > ?)
		ORDER BY version DESC LIMIT 1
	`, keyHash, now).Scan(
		&key.ID, &key.ClientID, &key.KeyHash, &key.Version,
		&key.IssuedAt, &expiresAt, &key.IsActive, &revokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Int64
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Int64
	}

	return &key, nil
}

// RevokeSigningKey revokes a signing key.
func (s *Store) RevokeSigningKey(ctx context.Context, keyID string) error {
	now := time.Now().UTC().UnixMilli()
	_, err := s.db.ExecContext(ctx, `
		UPDATE signing_keys SET is_active = 0, revoked_at = ? WHERE id = ?
	`, now, keyID)
	return err
}

// RotateSigningKey rotates the signing key for a client.
// Creates a new key and deactivates the old one after gracePeriod.
func (s *Store) RotateSigningKey(ctx context.Context, clientID string, gracePeriod time.Duration) (*SigningKey, string, error) {
	expiresAt := time.Now().UTC().Add(gracePeriod).UnixMilli()
	return s.CreateSigningKey(ctx, clientID, &expiresAt)
}