package api_key

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/api_key"
	infraStorage "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

// Config holds the API key service configuration.
type Config struct {
	Prefix           string // e.g., "vxyz"
	MaxPerMonth      int    // Max keys per operator per month
	MaxNameLength    int    // Max key name length
	DefaultExpiryDays int   // Default expiry (0 = never)
	MaxExpiryDays    int    // Max expiry days
	PrefixLength     int    // Prefix length for display
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Prefix:           "vxyz",
		MaxPerMonth:      20,
		MaxNameLength:    64,
		DefaultExpiryDays: 0,
		MaxExpiryDays:    365,
		PrefixLength:     8,
	}
}

// Service handles API key business logic.
type Service struct {
	repo   infraStorage.APIKeyRepository
	config Config
}

// NewService creates a new API key service.
func NewService(repo infraStorage.APIKeyRepository, config Config) *Service {
	return &Service{
		repo:   repo,
		config: config,
	}
}

// GenerateKey generates a new API key and returns the full key (only time it's available).
func (s *Service) GenerateKey(ctx context.Context, operatorID string, req *api_key.CreateAPIKeyRequest) (*api_key.APIKeyWithFullKey, error) {
	// Validate name
	if len(req.Name) > s.config.MaxNameLength {
		return nil, api_key.ErrKeyNameTooLong
	}

	// Validate scope
	if !req.Scope.IsValid() {
		return nil, api_key.ErrInvalidScope
	}

	// Check monthly limit
	count, err := s.repo.CountByOperatorThisMonth(ctx, operatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to count keys: %w", err)
	}
	if count >= s.config.MaxPerMonth {
		return nil, api_key.ErrMonthlyLimitExceeded
	}

	// Generate the full key
	fullKey, err := generateRandomKey(s.config.Prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Hash the key
	keyHash, err := hashKey(fullKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	// Calculate prefix
	prefix := fullKey[:s.config.PrefixLength]

	// Calculate expiry
	var expiresAt *time.Time
	if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		if *req.ExpiresInDays > s.config.MaxExpiryDays {
			return nil, api_key.ErrInvalidExpiryDays
		}
		exp := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &exp
	}

	now := time.Now()
	key := &infraStorage.APIKey{
		ID:         uuid.Must(uuid.NewV7()).String(),
		OperatorID: operatorID,
		Name:       req.Name,
		KeyPrefix:  prefix,
		KeyHash:    keyHash,
		Scope:      string(req.Scope),
		ExpiresAt:  toMillis(expiresAt),
		IsActive:   true,
		CreatedAt:  now.UnixMilli(),
		UpdatedAt:  now.UnixMilli(),
	}

	if err := s.repo.Create(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}

	return &api_key.APIKeyWithFullKey{
		APIKey: api_key.APIKey{
			ID:         key.ID,
			OperatorID: key.OperatorID,
			Name:       key.Name,
			KeyPrefix:  key.KeyPrefix,
			Scope:      api_key.Scope(key.Scope),
			ExpiresAt:  fromMillis(key.ExpiresAt),
			IsActive:   key.IsActive,
			CreatedAt:  fromMillisVal(key.CreatedAt),
		},
		FullKey: fullKey,
	}, nil
}

// ValidateKey validates an API key and returns the key if valid.
func (s *Service) ValidateKey(ctx context.Context, fullKey string) (*api_key.APIKey, error) {
	keyHash := hashKeyValue(fullKey)

	key, err := s.repo.GetByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, api_key.ErrInvalidAPIKey
	}

	apiKey := toDomainAPIKey(key)

	if !apiKey.IsValid() {
		if apiKey.IsExpired() {
			return nil, api_key.ErrAPIKeyExpired
		}
		if !apiKey.IsActive {
			return nil, api_key.ErrAPIKeyRevoked
		}
		return nil, api_key.ErrAPIKeyInactive
	}

	return apiKey, nil
}

// IncrementUsage increments the request counter for an API key.
func (s *Service) IncrementUsage(ctx context.Context, keyID string) error {
	return s.repo.IncrementRequestCount(ctx, keyID)
}

// ListKeys lists all API keys for an operator.
func (s *Service) ListKeys(ctx context.Context, operatorID string, page, limit int) (*api_key.ListAPIKeysResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	keys, total, err := s.repo.ListByOperator(ctx, operatorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	// Count keys created this month
	monthlyCount, err := s.repo.CountByOperatorThisMonth(ctx, operatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to count monthly keys: %w", err)
	}

	responses := make([]api_key.APIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = toDomainAPIKey(key).ToResponse()
	}

	totalPages := (total + limit - 1) / limit

	return &api_key.ListAPIKeysResponse{
		Keys: responses,
		Pagination: api_key.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
		MonthlyLimit:        s.config.MaxPerMonth,
		KeysCreatedThisMonth: monthlyCount,
	}, nil
}

// GetKey gets a single API key by ID.
func (s *Service) GetKey(ctx context.Context, operatorID, keyID string) (*api_key.APIKey, error) {
	key, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return nil, api_key.ErrAPIKeyNotFound
	}

	// Ensure the key belongs to the operator
	if key.OperatorID != operatorID {
		return nil, api_key.ErrAPIKeyNotFound
	}

	return toDomainAPIKey(key), nil
}

// UpdateKey updates an API key (name and/or scope).
func (s *Service) UpdateKey(ctx context.Context, operatorID, keyID string, req *api_key.UpdateAPIKeyRequest) (*api_key.APIKey, error) {
	key, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return nil, api_key.ErrAPIKeyNotFound
	}

	// Ensure the key belongs to the operator
	if key.OperatorID != operatorID {
		return nil, api_key.ErrAPIKeyNotFound
	}

	// Validate name if provided
	if req.Name != nil {
		if len(*req.Name) > s.config.MaxNameLength {
			return nil, api_key.ErrKeyNameTooLong
		}
		key.Name = *req.Name
	}

	// Validate scope if provided
	if req.Scope != nil {
		if !req.Scope.IsValid() {
			return nil, api_key.ErrInvalidScope
		}
		key.Scope = string(*req.Scope)
	}

	key.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.Update(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to update key: %w", err)
	}

	return toDomainAPIKey(key), nil
}

// RevokeKey revokes an API key.
func (s *Service) RevokeKey(ctx context.Context, operatorID, keyID string) error {
	key, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return api_key.ErrAPIKeyNotFound
	}

	// Ensure the key belongs to the operator
	if key.OperatorID != operatorID {
		return api_key.ErrAPIKeyNotFound
	}

	return s.repo.Revoke(ctx, keyID)
}

// ForceRevokeKey force revokes an API key (super admin only).
func (s *Service) ForceRevokeKey(ctx context.Context, keyID string) error {
	return s.repo.Revoke(ctx, keyID)
}

// RotateKey rotates an API key, generating a new key and invalidating the old one.
func (s *Service) RotateKey(ctx context.Context, operatorID, keyID string) (*api_key.APIKeyWithFullKey, error) {
	key, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return nil, api_key.ErrAPIKeyNotFound
	}

	// Ensure the key belongs to the operator
	if key.OperatorID != operatorID {
		return nil, api_key.ErrAPIKeyNotFound
	}

	// Revoke the old key
	if err := s.repo.Revoke(ctx, keyID); err != nil {
		return nil, fmt.Errorf("failed to revoke old key: %w", err)
	}

	// Generate a new key with the same settings
	req := &api_key.CreateAPIKeyRequest{
		Name:          key.Name,
		Scope:         api_key.Scope(key.Scope),
		ExpiresInDays: fromMillisToDays(key.ExpiresAt),
	}

	// Re-check monthly limit (we're creating a new key)
	count, err := s.repo.CountByOperatorThisMonth(ctx, operatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to count keys: %w", err)
	}
	if count >= s.config.MaxPerMonth {
		return nil, api_key.ErrMonthlyLimitExceeded
	}

	// Generate the new key
	fullKey, err := generateRandomKey(s.config.Prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	keyHash, err := hashKey(fullKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	prefix := fullKey[:s.config.PrefixLength]
	now := time.Now()

	newKey := &infraStorage.APIKey{
		ID:         uuid.Must(uuid.NewV7()).String(),
		OperatorID: operatorID,
		Name:       req.Name,
		KeyPrefix:  prefix,
		KeyHash:    keyHash,
		Scope:      string(req.Scope),
		ExpiresAt:  key.ExpiresAt, // Keep the same expiry
		IsActive:   true,
		CreatedAt:  now.UnixMilli(),
		UpdatedAt:  now.UnixMilli(),
	}

	if err := s.repo.Create(ctx, newKey); err != nil {
		return nil, fmt.Errorf("failed to create new key: %w", err)
	}

	return &api_key.APIKeyWithFullKey{
		APIKey: api_key.APIKey{
			ID:         newKey.ID,
			OperatorID: newKey.OperatorID,
			Name:       newKey.Name,
			KeyPrefix:  newKey.KeyPrefix,
			Scope:      api_key.Scope(newKey.Scope),
			ExpiresAt:  fromMillis(newKey.ExpiresAt),
			IsActive:   newKey.IsActive,
			CreatedAt:  fromMillisVal(newKey.CreatedAt),
		},
		FullKey: fullKey,
	}, nil
}

// generateRandomKey generates a random API key with the given prefix.
func generateRandomKey(prefix string) (string, error) {
	const keyLength = 32 // 32 bytes = 64 hex characters
	bytes := make([]byte, keyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(bytes)), nil
}

// hashKey hashes a key using Argon2id.
func hashKey(key string) (string, error) {
	// Use a fixed salt for deterministic hashing (the key itself is the secret)
	salt := []byte("vyzorix-api-key-v1")
	hash := argon2.IDKey([]byte(key), salt, 1, 64*1024, 4, 32)
	return hex.EncodeToString(hash), nil
}

// hashKeyValue is a convenience function to hash a key value.
func hashKeyValue(key string) string {
	hash, _ := hashKey(key)
	return hash
}

// toDomainAPIKey converts an infrastructure API key to a domain API key.
func toDomainAPIKey(key *infraStorage.APIKey) *api_key.APIKey {
	domain := &api_key.APIKey{
		ID:           key.ID,
		OperatorID:   key.OperatorID,
		Name:         key.Name,
		KeyPrefix:    key.KeyPrefix,
		Scope:        api_key.Scope(key.Scope),
		IsActive:     key.IsActive,
		RequestCount: key.RequestCount,
		CreatedAt:    fromMillisVal(key.CreatedAt),
		UpdatedAt:    fromMillisVal(key.UpdatedAt),
	}

	if key.ExpiresAt != nil {
		t := time.UnixMilli(*key.ExpiresAt)
		domain.ExpiresAt = &t
	}

	if key.LastRequest != nil {
		t := time.UnixMilli(*key.LastRequest)
		domain.LastRequest = &t
	}

	if key.RevokedAt != nil {
		t := time.UnixMilli(*key.RevokedAt)
		domain.RevokedAt = &t
	}

	return domain
}

// toMillis converts a time to Unix milliseconds.
func toMillis(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	ms := t.UnixMilli()
	return &ms
}

// fromMillis converts Unix milliseconds to a time.
func fromMillis(ms *int64) *time.Time {
	if ms == nil {
		return nil
	}
	t := time.UnixMilli(*ms)
	return &t
}

// fromMillisVal converts Unix milliseconds to a time (non-pointer).
func fromMillisVal(ms int64) time.Time {
	return time.UnixMilli(ms)
}

// fromMillisToDays converts milliseconds to days for expiry.
func fromMillisToDays(ms *int64) *int {
	if ms == nil {
		return nil
	}
	days := int(time.Until(time.UnixMilli(*ms)).Hours() / 24)
	if days < 0 {
		return nil
	}
	return &days
}

// IsValidScope checks if a string is a valid scope.
func IsValidScope(s string) bool {
	return api_key.Scope(strings.ToLower(s)).IsValid()
}
