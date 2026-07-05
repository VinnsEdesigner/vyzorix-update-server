package keys

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain"
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
func (s *Service) GenerateKey(ctx context.Context, operatorID string, req *domain.CreateAPIKeyRequest) (*domain.APIKeyWithFullKey, error) {
	// Validate name
	if len(req.Name) > s.config.MaxNameLength {
		return nil, domain.ErrKeyNameTooLong
	}

	// Validate scope
	if !req.Scope.IsValid() {
		return nil, domain.ErrInvalidScope
	}

	// Check monthly limit
	count, err := s.repo.CountByOperatorThisMonth(ctx, operatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to count keys: %w", err)
	}
	if count >= s.config.MaxPerMonth {
		return nil, domain.ErrMonthlyLimitExceeded
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
			return nil, domain.ErrInvalidExpiryDays
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

	return &domain.APIKeyWithFullKey{
		APIKey: domain.APIKey{
			ID:         key.ID,
			OperatorID: key.OperatorID,
			Name:       key.Name,
			KeyPrefix:  key.KeyPrefix,
			Scope:      domain.Scope(key.Scope),
			ExpiresAt:  fromMillis(key.ExpiresAt),
			IsActive:   key.IsActive,
			CreatedAt:  fromMillisVal(key.CreatedAt),
		},
		FullKey: fullKey,
	}, nil
}

// ValidateKey validates an API key and returns the key if valid.
func (s *Service) ValidateKey(ctx context.Context, fullKey string) (*domain.APIKey, error) {
	keyHash := hashKeyValue(fullKey)

	key, err := s.repo.GetByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, domain.ErrInvalidAPIKey
	}

	apiKey := toDomainAPIKey(key)

	if !apiKey.IsValid() {
		if apiKey.IsExpired() {
			return nil, domain.ErrAPIKeyExpired
		}
		if !apiKey.IsActive {
			return nil, domain.ErrAPIKeyRevoked
		}
		return nil, domain.ErrAPIKeyInactive
	}

	return apiKey, nil
}

// IncrementUsage increments the request counter for an API key.
// VerifyKey verifies a key against a stored hash using constant-time comparison.
func (s *Service) VerifyKey(fullKey, keyHash string) bool {
	hashedKey := hashKeyValue(fullKey)
	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(hashedKey), []byte(keyHash)) == 1
}

func (s *Service) IncrementUsage(ctx context.Context, keyID string) error {
	return s.repo.IncrementRequestCount(ctx, keyID)
}

// ListKeys lists all API keys for an operator.
func (s *Service) ListKeys(ctx context.Context, operatorID string, page, limit int) (*domain.ListAPIKeysResponse, error) {
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

	responses := make([]domain.APIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = toDomainAPIKey(key).ToResponse()
	}

	totalPages := (total + limit - 1) / limit

	return &domain.ListAPIKeysResponse{
		Keys: responses,
		Pagination: domain.Pagination{
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
func (s *Service) GetKey(ctx context.Context, operatorID, keyID string) (*domain.APIKey, error) {
	key, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return nil, domain.ErrAPIKeyNotFound
	}

	// Ensure the key belongs to the operator
	if key.OperatorID != operatorID {
		return nil, domain.ErrAPIKeyNotFound
	}

	return toDomainAPIKey(key), nil
}

// UpdateKey updates an API key (name and/or scope).
func (s *Service) UpdateKey(ctx context.Context, operatorID, keyID string, req *domain.UpdateAPIKeyRequest) (*domain.APIKey, error) {
	key, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return nil, domain.ErrAPIKeyNotFound
	}

	// Ensure the key belongs to the operator
	if key.OperatorID != operatorID {
		return nil, domain.ErrAPIKeyNotFound
	}

	// Validate name if provided
	if req.Name != nil {
		if len(*req.Name) > s.config.MaxNameLength {
			return nil, domain.ErrKeyNameTooLong
		}
		key.Name = *req.Name
	}

	// Validate scope if provided
	if req.Scope != nil {
		if !req.Scope.IsValid() {
			return nil, domain.ErrInvalidScope
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
		return domain.ErrAPIKeyNotFound
	}

	// Ensure the key belongs to the operator
	if key.OperatorID != operatorID {
		return domain.ErrAPIKeyNotFound
	}

	return s.repo.Revoke(ctx, keyID)
}

// ForceRevokeKey force revokes an API key (super admin only).
func (s *Service) ForceRevokeKey(ctx context.Context, keyID string) error {
	return s.repo.Revoke(ctx, keyID)
}

// RotateKey rotates an API key, generating a new key and invalidating the old one.
func (s *Service) RotateKey(ctx context.Context, operatorID, keyID string) (*domain.APIKeyWithFullKey, error) {
	key, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return nil, domain.ErrAPIKeyNotFound
	}

	// Ensure the key belongs to the operator
	if key.OperatorID != operatorID {
		return nil, domain.ErrAPIKeyNotFound
	}

	// Revoke the old key
	if revokeErr := s.repo.Revoke(ctx, keyID); revokeErr != nil {
		return nil, fmt.Errorf("failed to revoke old key: %w", revokeErr)
	}

	// Generate a new key with the same settings
	req := &domain.CreateAPIKeyRequest{
		Name:          key.Name,
		Scope:         domain.Scope(key.Scope),
		
	}

	// Re-check monthly limit (we're creating a new key)
	count, err := s.repo.CountByOperatorThisMonth(ctx, operatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to count keys: %w", err)
	}
	if count >= s.config.MaxPerMonth {
		return nil, domain.ErrMonthlyLimitExceeded
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

	return &domain.APIKeyWithFullKey{
		APIKey: domain.APIKey{
			ID:         newKey.ID,
			OperatorID: newKey.OperatorID,
			Name:       newKey.Name,
			KeyPrefix:  newKey.KeyPrefix,
			Scope:      domain.Scope(newKey.Scope),
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
func hashKey(key string) (string, error) { //nolint:unparam
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
func toDomainAPIKey(key *infraStorage.APIKey) *domain.APIKey {
	domain := &domain.APIKey{
		ID:           key.ID,
		OperatorID:   key.OperatorID,
		Name:         key.Name,
		KeyPrefix:    key.KeyPrefix,
		Scope:        domain.Scope(key.Scope),
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
	return domain.Scope(strings.ToLower(s)).IsValid()
}

// ListAllKeys lists all API keys across all operators (super admin only).
func (s *Service) ListAllKeys(ctx context.Context, page, limit int) (*domain.ListAllAPIKeysResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	keys, total, err := s.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list all keys: %w", err)
	}

	responses := make([]domain.APIKeyResponse, len(keys))
	for i, key := range keys {
		responses[i] = toDomainAPIKey(key).ToResponse()
	}

	totalPages := (total + limit - 1) / limit

	return &domain.ListAllAPIKeysResponse{
		Keys: responses,
		Pagination: domain.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetGlobalStats returns global API key statistics (super admin only).
func (s *Service) GetGlobalStats(ctx context.Context) (*domain.GlobalAPIKeyStats, error) {
	totalActive, err := s.repo.CountAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count all keys: %w", err)
	}

	return &domain.GlobalAPIKeyStats{
		TotalActiveKeys: totalActive,
		MaxPerMonth:    s.config.MaxPerMonth,
	}, nil
}
