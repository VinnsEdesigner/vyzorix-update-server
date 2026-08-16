package keys

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain"
	infraStorage "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

// =============================================================================.
// Mock Repository for Testing.
// =============================================================================.

type mockAPIKeyRepository struct {
	mu     sync.RWMutex
	keys   map[string]*infraStorage.APIKey
	byHash map[string]*infraStorage.APIKey
}

func newMockRepository() *mockAPIKeyRepository {
	return &mockAPIKeyRepository{
		keys:   make(map[string]*infraStorage.APIKey),
		byHash: make(map[string]*infraStorage.APIKey),
	}
}

func (r *mockAPIKeyRepository) Create(ctx context.Context, key *infraStorage.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.keys[key.ID] = key
	if key.KeyHash != "" {
		r.byHash[key.KeyHash] = key
	}
	return nil
}

func (r *mockAPIKeyRepository) GetByID(ctx context.Context, id string) (*infraStorage.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if key, ok := r.keys[id]; ok {
		return key, nil
	}
	return nil, sql.ErrNoRows
}

func (r *mockAPIKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*infraStorage.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if key, ok := r.byHash[keyHash]; ok && key.IsActive {
		return key, nil
	}
	return nil, sql.ErrNoRows
}

func (r *mockAPIKeyRepository) ListByOperator(ctx context.Context, operatorID string, limit, offset int) ([]*infraStorage.APIKey, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*infraStorage.APIKey
	for _, k := range r.keys {
		if k.OperatorID == operatorID {
			result = append(result, k)
		}
	}
	total := len(result)
	if offset >= len(result) {
		return []*infraStorage.APIKey{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (r *mockAPIKeyRepository) ListAll(ctx context.Context, limit, offset int) ([]*infraStorage.APIKey, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*infraStorage.APIKey
	for _, k := range r.keys {
		if k.IsActive {
			result = append(result, k)
		}
	}
	total := len(result)
	if offset >= len(result) {
		return []*infraStorage.APIKey{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (r *mockAPIKeyRepository) ListAllWithOperator(ctx context.Context, limit, offset int, operatorID, search string) ([]infraStorage.APIKeyWithOperator, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []infraStorage.APIKeyWithOperator
	for _, k := range r.keys {
		if k.IsActive {
			if operatorID != "" && k.OperatorID != operatorID {
				continue
			}
			if search != "" {
				if !strings.Contains(strings.ToLower(k.Name), strings.ToLower(search)) {
					continue
				}
			}
			result = append(result, infraStorage.APIKeyWithOperator{APIKey: *k})
		}
	}
	total := len(result)
	if offset >= len(result) {
		return []infraStorage.APIKeyWithOperator{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (r *mockAPIKeyRepository) Update(ctx context.Context, key *infraStorage.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.keys[key.ID] = key
	return nil
}

func (r *mockAPIKeyRepository) Revoke(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if key, ok := r.keys[id]; ok {
		key.IsActive = false
		now := time.Now().UnixMilli()
		key.RevokedAt = &now
		key.UpdatedAt = now
	}
	return nil
}

func (r *mockAPIKeyRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if key, ok := r.keys[id]; ok {
		delete(r.byHash, key.KeyHash)
	}
	delete(r.keys, id)
	return nil
}

func (r *mockAPIKeyRepository) CountByOperatorThisMonth(ctx context.Context, operatorID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	count := 0
	for _, k := range r.keys {
		if k.OperatorID == operatorID && k.CreatedAt >= monthStart {
			count++
		}
	}
	return count, nil
}

func (r *mockAPIKeyRepository) CountAll(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, k := range r.keys {
		if k.IsActive {
			count++
		}
	}
	return count, nil
}

func (r *mockAPIKeyRepository) CountAllIncludingRevoked(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.keys), nil
}

func (r *mockAPIKeyRepository) CountRevoked(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, k := range r.keys {
		if !k.IsActive {
			count++
		}
	}
	return count, nil
}

func (r *mockAPIKeyRepository) SumRequestsByScope(ctx context.Context) (map[string]int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]int64)
	for _, k := range r.keys {
		result[k.Scope] += k.RequestCount
	}
	return result, nil
}

func (r *mockAPIKeyRepository) TopOperatorsByRequests(ctx context.Context, limit int) ([]infraStorage.OperatorRequestTotal, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	byOp := make(map[string]int64)
	for _, k := range r.keys {
		byOp[k.OperatorID] += k.RequestCount
	}
	totalOperators := len(byOp)
	var results []infraStorage.OperatorRequestTotal
	for opID, total := range byOp {
		results = append(results, infraStorage.OperatorRequestTotal{OperatorID: opID, TotalRequests: total})
		if len(results) >= limit {
			break
		}
	}
	return results, totalOperators, nil
}

func (r *mockAPIKeyRepository) IncrementRequestCount(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if key, ok := r.keys[id]; ok {
		key.RequestCount++
		now := time.Now().UnixMilli()
		key.LastRequest = &now
		key.UpdatedAt = now
	}
	return nil
}

func (r *mockAPIKeyRepository) ExistsByOperatorAndName(ctx context.Context, operatorID, name string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, k := range r.keys {
		if k.OperatorID == operatorID && k.Name == name && k.IsActive {
			return true, nil
		}
	}
	return false, nil
}

func (r *mockAPIKeyRepository) ExistsByOperatorAndNameExcluding(ctx context.Context, operatorID, name, excludeKeyID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, k := range r.keys {
		if k.OperatorID == operatorID && k.Name == name && k.ID != excludeKeyID && k.IsActive {
			return true, nil
		}
	}
	return false, nil
}

// =============================================================================.
// Test Fixtures.
// =============================================================================.

func setupTestService(t *testing.T) (*APIKeyService, *mockAPIKeyRepository) {
	repo := newMockRepository()
	config := DefaultConfig()
	config.Prefix = "test"
	config.MaxPerMonth = 10
	config.MaxExpiryDays = 365
	config.PrefixLength = 8
	svc := NewAPIKeyService(repo, config)
	return svc, repo
}

// =============================================================================.

func TestAPIKeyService_CreateKey(t *testing.T) {
	svc, repo := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Test Key",
		Scope: domain.ScopeRead,
	}

	result, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if result.FullKey == "" {
		t.Error("FullKey should not be empty")
	}

	if result.Name != "Test Key" {
		t.Errorf("Name = %s, want Test Key", result.Name)
	}

	if result.Scope != domain.ScopeRead {
		t.Errorf("Scope = %s, want read", result.Scope)
	}

	if !result.IsActive {
		t.Error("IsActive should be true")
	}

	stored, err := repo.GetByID(ctx, result.ID)
	if err != nil {
		t.Fatalf("Failed to get stored key: %v", err)
	}

	if stored.KeyHash == "" {
		t.Error("KeyHash should not be empty")
	}
}

func TestAPIKeyService_CreateKey_WithExpiry(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	days := 30
	req := &domain.CreateAPIKeyRequest{
		Name:          "Expiring Key",
		Scope:         domain.ScopeRead,
		ExpiresInDays: &days,
	}

	result, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if result.ExpiresAt == nil {
		t.Fatal("ExpiresAt should not be nil")
	}

	expectedExpiry := time.Now().AddDate(0, 0, 30)
	if result.ExpiresAt.Sub(expectedExpiry) > time.Minute {
		t.Errorf("ExpiresAt = %v, want approximately %v", result.ExpiresAt, expectedExpiry)
	}
}

func TestAPIKeyService_CreateKey_DuplicateName(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Duplicate Key",
		Scope: domain.ScopeRead,
	}

	_, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("First CreateKey failed: %v", err)
	}

	_, err = svc.GenerateKey(ctx, "operator-1", req)
	if err != domain.ErrKeyNameConflict {
		t.Errorf("Second CreateKey error = %v, want ErrKeyNameConflict", err)
	}
}

func TestAPIKeyService_CreateKey_MonthlyLimit(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		req := &domain.CreateAPIKeyRequest{
			Name:  "Key " + string(rune('A'+i)),
			Scope: domain.ScopeRead,
		}
		_, err := svc.GenerateKey(ctx, "operator-1", req)
		if err != nil {
			t.Fatalf("CreateKey %d failed: %v", i, err)
		}
	}

	req := &domain.CreateAPIKeyRequest{
		Name:  "Key K",
		Scope: domain.ScopeRead,
	}
	_, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != domain.ErrMonthlyLimitExceeded {
		t.Errorf("Eleventh CreateKey error = %v, want ErrMonthlyLimitExceeded", err)
	}
}

func TestAPIKeyService_ValidateKey(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Validation Test Key",
		Scope: domain.ScopeRead,
	}

	result, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	validated, err := svc.ValidateKey(ctx, result.FullKey)
	if err != nil {
		t.Fatalf("ValidateKey failed: %v", err)
	}

	if validated.ID != result.ID {
		t.Errorf("Validated ID = %s, want %s", validated.ID, result.ID)
	}

	if validated.OperatorID != "operator-1" {
		t.Errorf("OperatorID = %s, want operator-1", validated.OperatorID)
	}
}

func TestAPIKeyService_ValidateKey_Invalid(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	_, err := svc.ValidateKey(ctx, "invalid_key_that_does_not_exist")
	if err != domain.ErrInvalidAPIKey {
		t.Errorf("ValidateKey error = %v, want ErrInvalidAPIKey", err)
	}
}

func TestAPIKeyService_ValidateKey_Expired(t *testing.T) {
	svc, repo := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Expired Key",
		Scope: domain.ScopeRead,
	}

	result, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	past := time.Now().Add(-24 * time.Hour).UnixMilli()
	repo.keys[result.ID].ExpiresAt = &past

	_, err = svc.ValidateKey(ctx, result.FullKey)
	if err != domain.ErrAPIKeyExpired {
		t.Errorf("ValidateKey error = %v, want ErrAPIKeyExpired", err)
	}
}

func TestAPIKeyService_ValidateKey_Revoked(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Revoked Key",
		Scope: domain.ScopeRead,
	}

	result, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	err = svc.RevokeKey(ctx, "operator-1", result.ID)
	if err != nil {
		t.Fatalf("RevokeKey failed: %v", err)
	}

	// Revoked keys are not returned by GetByKeyHash (security: no existence leakage).
	_, err = svc.ValidateKey(ctx, result.FullKey)
	if err != domain.ErrInvalidAPIKey {
		t.Errorf("ValidateKey error = %v, want ErrInvalidAPIKey (revoked keys not returned)", err)
	}
}

func TestAPIKeyService_RevokeKey(t *testing.T) {
	svc, repo := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Key to Revoke",
		Scope: domain.ScopeRead,
	}

	result, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	err = svc.RevokeKey(ctx, "operator-1", result.ID)
	if err != nil {
		t.Fatalf("RevokeKey failed: %v", err)
	}

	repo.mu.RLock()
	key := repo.keys[result.ID]
	repo.mu.RUnlock()

	if key.IsActive {
		t.Error("Key should not be active after revocation")
	}

	if key.RevokedAt == nil {
		t.Error("RevokedAt should be set")
	}
}

func TestAPIKeyService_RevokeKey_WrongOperator(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Key",
		Scope: domain.ScopeRead,
	}

	result, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	err = svc.RevokeKey(ctx, "operator-2", result.ID)
	if err != domain.ErrAPIKeyNotFound {
		t.Errorf("RevokeKey error = %v, want ErrAPIKeyNotFound", err)
	}
}

func TestAPIKeyService_RotateKey(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Key to Rotate",
		Scope: domain.ScopeRead,
	}

	original, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	newKey, err := svc.RotateKey(ctx, "operator-1", original.ID)
	if err != nil {
		t.Fatalf("RotateKey failed: %v", err)
	}

	if newKey.FullKey == original.FullKey {
		t.Error("Rotated key should be different from original")
	}

	// Original key should be revoked (not returned by GetByKeyHash).
	_, err = svc.ValidateKey(ctx, original.FullKey)
	if err != domain.ErrInvalidAPIKey {
		t.Errorf("Original key validation error = %v, want ErrInvalidAPIKey (revoked)", err)
	}

	// New key should be valid.
	_, err = svc.ValidateKey(ctx, newKey.FullKey)
	if err != nil {
		t.Errorf("New key validation failed: %v", err)
	}

	if newKey.Name != original.Name {
		t.Errorf("Name = %s, want %s", newKey.Name, original.Name)
	}

	if newKey.Scope != original.Scope {
		t.Errorf("Scope = %s, want %s", newKey.Scope, original.Scope)
	}
}

func TestAPIKeyService_RotateKey_MonthlyLimit(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		req := &domain.CreateAPIKeyRequest{
			Name:  string(rune('A' + i)),
			Scope: domain.ScopeRead,
		}
		_, err := svc.GenerateKey(ctx, "operator-1", req)
		if err != nil {
			t.Fatalf("CreateKey %d failed: %v", i, err)
		}
	}

	result, _ := svc.ListKeys(ctx, "operator-1", 1, 1)
	if len(result.Keys) == 0 {
		t.Fatal("No keys found")
	}

	_, err := svc.RotateKey(ctx, "operator-1", result.Keys[0].ID)
	if err != domain.ErrMonthlyLimitExceeded {
		t.Errorf("RotateKey error = %v, want ErrMonthlyLimitExceeded", err)
	}
}

func TestAPIKeyService_ListKeys(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		req := &domain.CreateAPIKeyRequest{
			Name:  "List Key " + string(rune('0'+i)),
			Scope: domain.ScopeRead,
		}
		_, err := svc.GenerateKey(ctx, "operator-1", req)
		if err != nil {
			t.Fatalf("CreateKey %d failed: %v", i, err)
		}
	}

	result, err := svc.ListKeys(ctx, "operator-1", 1, 10)
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}

	if len(result.Keys) != 5 {
		t.Errorf("Keys count = %d, want 5", len(result.Keys))
	}
}

func TestAPIKeyService_UpdateKey(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Original Name",
		Scope: domain.ScopeRead,
	}

	key, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	newName := "Updated Name"
	updateReq := &domain.UpdateAPIKeyRequest{
		Name: &newName,
	}

	updated, err := svc.UpdateKey(ctx, "operator-1", key.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateKey failed: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Name = %s, want Updated Name", updated.Name)
	}

	newScope := domain.ScopeWrite
	updateReq = &domain.UpdateAPIKeyRequest{
		Scope: &newScope,
	}

	updated, err = svc.UpdateKey(ctx, "operator-1", key.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateKey failed: %v", err)
	}

	if updated.Scope != domain.ScopeWrite {
		t.Errorf("Scope = %s, want write", updated.Scope)
	}
}

func TestAPIKeyService_UpdateKey_DuplicateName(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req1 := &domain.CreateAPIKeyRequest{Name: "Key One", Scope: domain.ScopeRead}
	key1, _ := svc.GenerateKey(ctx, "operator-1", req1)

	req2 := &domain.CreateAPIKeyRequest{Name: "Key Two", Scope: domain.ScopeRead}
	key2, _ := svc.GenerateKey(ctx, "operator-1", req2)

	newName := "Key One"
	updateReq := &domain.UpdateAPIKeyRequest{Name: &newName}

	_, err := svc.UpdateKey(ctx, "operator-1", key2.ID, updateReq)
	if err != domain.ErrKeyNameConflict {
		t.Errorf("UpdateKey error = %v, want ErrKeyNameConflict", err)
	}

	_ = key1
}

// =============================================================================.
// Tenant Isolation Tests.
// =============================================================================.

func TestAPIKeyService_TenantIsolation_ValidateKey(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Tenant Key",
		Scope: domain.ScopeRead,
	}

	key1, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	validated, err := svc.ValidateKey(ctx, key1.FullKey)
	if err != nil {
		t.Fatalf("ValidateKey failed: %v", err)
	}

	if validated.OperatorID != "operator-1" {
		t.Errorf("OperatorID = %s, want operator-1", validated.OperatorID)
	}
}

func TestAPIKeyService_TenantIsolation_GetKey(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Private Key",
		Scope: domain.ScopeRead,
	}

	key, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	retrieved, err := svc.GetKey(ctx, "operator-1", key.ID)
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}

	if retrieved.ID != key.ID {
		t.Errorf("Retrieved ID = %s, want %s", retrieved.ID, key.ID)
	}

	_, err = svc.GetKey(ctx, "operator-2", key.ID)
	if err != domain.ErrAPIKeyNotFound {
		t.Errorf("GetKey error = %v, want ErrAPIKeyNotFound", err)
	}
}

func TestAPIKeyService_TenantIsolation_UpdateKey(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Protected Key",
		Scope: domain.ScopeRead,
	}

	key, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	newName := "Hacked Name"
	updateReq := &domain.UpdateAPIKeyRequest{Name: &newName}

	_, err = svc.UpdateKey(ctx, "operator-2", key.ID, updateReq)
	if err != domain.ErrAPIKeyNotFound {
		t.Errorf("UpdateKey error = %v, want ErrAPIKeyNotFound", err)
	}
}

func TestAPIKeyService_TenantIsolation_ListKeys(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		req := &domain.CreateAPIKeyRequest{
			Name:  "Op1 Key " + string(rune('A'+i)),
			Scope: domain.ScopeRead,
		}
		_, _ = svc.GenerateKey(ctx, "operator-1", req)
	}

	for i := 0; i < 2; i++ {
		req := &domain.CreateAPIKeyRequest{
			Name:  "Op2 Key " + string(rune('A'+i)),
			Scope: domain.ScopeRead,
		}
		_, _ = svc.GenerateKey(ctx, "operator-2", req)
	}

	result1, err := svc.ListKeys(ctx, "operator-1", 1, 100)
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}

	if len(result1.Keys) != 3 {
		t.Errorf("operator-1 keys count = %d, want 3", len(result1.Keys))
	}

	result2, err := svc.ListKeys(ctx, "operator-2", 1, 100)
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}

	if len(result2.Keys) != 2 {
		t.Errorf("operator-2 keys count = %d, want 2", len(result2.Keys))
	}
}

// =============================================================================.
// Security Tests.
// =============================================================================.

func TestAPIKeyService_KeyGeneration_Uniqueness(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		req := &domain.CreateAPIKeyRequest{
			Name:  "Unique Key " + string(rune('A'+i)),
			Scope: domain.ScopeRead,
		}
		result, err := svc.GenerateKey(ctx, "op-"+string(rune('0'+i)), req)
		if err != nil {
			t.Fatalf("GenerateKey failed: %v", err)
		}

		if keys[result.FullKey] {
			t.Fatal("Duplicate key generated")
		}
		keys[result.FullKey] = true
	}
}

func TestAPIKeyService_KeyGeneration_Format(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Format Test",
		Scope: domain.ScopeRead,
	}

	result, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if len(result.KeyPrefix) != 8 {
		t.Errorf("KeyPrefix length = %d, want 8", len(result.KeyPrefix))
	}

	fullKey := result.FullKey
	if len(fullKey) < 50 {
		t.Errorf("FullKey length = %d, want at least 50", len(fullKey))
	}
}

func TestAPIKeyService_KeyGeneration_CryptographicRandom(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{Name: "A", Scope: domain.ScopeRead}
	key1, _ := svc.GenerateKey(ctx, "op1", req)

	req.Name = "B"
	key2, _ := svc.GenerateKey(ctx, "op2", req)

	if key1.FullKey == key2.FullKey {
		t.Error("Two keys generated at the same time should not be identical")
	}
}

func TestAPIKeyService_VerifyKey_ConstantTime(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Verify Test",
		Scope: domain.ScopeRead,
	}

	key, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	result := svc.VerifyKey(key.FullKey, hashKeyValue(key.FullKey))
	if !result {
		t.Error("VerifyKey should return true for valid key")
	}

	result = svc.VerifyKey("wrong_key", hashKeyValue(key.FullKey))
	if result {
		t.Error("VerifyKey should return false for invalid key")
	}
}

func TestAPIKeyService_IncrementUsage(t *testing.T) {
	svc, repo := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Usage Test",
		Scope: domain.ScopeRead,
	}

	key, _ := svc.GenerateKey(ctx, "operator-1", req)

	for i := 0; i < 5; i++ {
		err := svc.IncrementUsage(ctx, key.ID)
		if err != nil {
			t.Fatalf("IncrementUsage failed: %v", err)
		}
	}

	repo.mu.RLock()
	stored := repo.keys[key.ID]
	repo.mu.RUnlock()

	if stored.RequestCount != 5 {
		t.Errorf("RequestCount = %d, want 5", stored.RequestCount)
	}

	if stored.LastRequest == nil {
		t.Error("LastRequest should be set")
	}
}

// =============================================================================.
// Scope Tests.
// =============================================================================.

func TestAPIKeyService_ScopeEnforcement(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	tests := []struct {
		scope     domain.Scope
		canRead   bool
		canWrite  bool
		canDelete bool
	}{
		{domain.ScopeRead, true, false, false},
		{domain.ScopeWrite, true, true, false},
		{domain.ScopeAdmin, true, true, true},
	}

	for _, tt := range tests {
		req := &domain.CreateAPIKeyRequest{
			Name:  "Scope Test " + string(tt.scope),
			Scope: tt.scope,
		}

		key, err := svc.GenerateKey(ctx, "operator-1", req)
		if err != nil {
			t.Fatalf("GenerateKey failed: %v", err)
		}

		if key.Scope.CanRead() != tt.canRead {
			t.Errorf("Scope %s CanRead = %v, want %v", tt.scope, key.Scope.CanRead(), tt.canRead)
		}

		if key.Scope.CanWrite() != tt.canWrite {
			t.Errorf("Scope %s CanWrite = %v, want %v", tt.scope, key.Scope.CanWrite(), tt.canWrite)
		}

		if key.Scope.CanDelete() != tt.canDelete {
			t.Errorf("Scope %s CanDelete = %v, want %v", tt.scope, key.Scope.CanDelete(), tt.canDelete)
		}
	}
}

// =============================================================================.
// Validation Tests.
// =============================================================================.

func TestAPIKeyService_Validation_InvalidScope(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	req := &domain.CreateAPIKeyRequest{
		Name:  "Invalid Scope Key",
		Scope: "invalid",
	}

	_, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != domain.ErrInvalidScope {
		t.Errorf("CreateKey error = %v, want ErrInvalidScope", err)
	}
}

func TestAPIKeyService_Validation_NameTooLong(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	longName := ""
	for i := 0; i < 100; i++ {
		longName += "x"
	}

	req := &domain.CreateAPIKeyRequest{
		Name:  longName,
		Scope: domain.ScopeRead,
	}

	_, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != domain.ErrKeyNameTooLong {
		t.Errorf("CreateKey error = %v, want ErrKeyNameTooLong", err)
	}
}

func TestAPIKeyService_Validation_ExpiryTooLong(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	days := 500
	req := &domain.CreateAPIKeyRequest{
		Name:          "Long Expiry",
		Scope:         domain.ScopeRead,
		ExpiresInDays: &days,
	}

	_, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != domain.ErrInvalidExpiryDays {
		t.Errorf("CreateKey error = %v, want ErrInvalidExpiryDays", err)
	}
}

func TestAPIKeyService_Validation_ExpiryZero(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	days := 0
	req := &domain.CreateAPIKeyRequest{
		Name:          "No Expiry",
		Scope:         domain.ScopeRead,
		ExpiresInDays: &days,
	}

	result, err := svc.GenerateKey(ctx, "operator-1", req)
	if err != nil {
		t.Fatalf("CreateKey failed: %v", err)
	}

	if result.ExpiresAt != nil {
		t.Error("ExpiresAt should be nil for zero days")
	}
}

func TestAPIKeyService_ListAllKeys(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := svc.GenerateKey(ctx, "operator-1", &domain.CreateAPIKeyRequest{
			Name: fmt.Sprintf("Op1 Key %d", i), Scope: domain.ScopeRead,
		})
		if err != nil {
			t.Fatalf("GenerateKey op1-%d failed: %v", i, err)
		}
	}
	for i := 0; i < 2; i++ {
		_, err := svc.GenerateKey(ctx, "operator-2", &domain.CreateAPIKeyRequest{
			Name: fmt.Sprintf("Op2 Key %d", i), Scope: domain.ScopeWrite,
		})
		if err != nil {
			t.Fatalf("GenerateKey op2-%d failed: %v", i, err)
		}
	}

	result, err := svc.ListAllKeys(ctx, 1, 50, "", "")
	if err != nil {
		t.Fatalf("ListAllKeys failed: %v", err)
	}
	if len(result.Keys) != 5 {
		t.Errorf("total keys = %d, want 5", len(result.Keys))
	}

	result, err = svc.ListAllKeys(ctx, 1, 50, "operator-1", "")
	if err != nil {
		t.Fatalf("ListAllKeys operator filter failed: %v", err)
	}
	if len(result.Keys) != 3 {
		t.Errorf("operator-1 keys = %d, want 3", len(result.Keys))
	}
	for _, k := range result.Keys {
		if k.OperatorID != "operator-1" {
			t.Errorf("expected operator-1, got %s", k.OperatorID)
		}
	}

	result, err = svc.ListAllKeys(ctx, 1, 50, "", "Op2")
	if err != nil {
		t.Fatalf("ListAllKeys search failed: %v", err)
	}
	if len(result.Keys) != 2 {
		t.Errorf("search Op2 keys = %d, want 2", len(result.Keys))
	}
}

func TestAPIKeyService_GetGlobalStats(t *testing.T) {
	svc, repo := setupTestService(t)
	ctx := context.Background()

	for _, tc := range []struct {
		op, name string
		scope    domain.Scope
	}{
		{"operator-1", "Read Key 1", domain.ScopeRead},
		{"operator-1", "Write Key 1", domain.ScopeWrite},
		{"operator-2", "Read Key 2", domain.ScopeRead},
		{"operator-2", "Admin Key 1", domain.ScopeAdmin},
	} {
		_, err := svc.GenerateKey(ctx, tc.op, &domain.CreateAPIKeyRequest{
			Name: tc.name, Scope: tc.scope,
		})
		if err != nil {
			t.Fatalf("GenerateKey %s failed: %v", tc.name, err)
		}
	}

	for _, k := range repo.keys {
		switch k.Name {
		case "Read Key 1":
			k.RequestCount = 10
		case "Write Key 1":
			k.RequestCount = 5
		case "Read Key 2":
			k.RequestCount = 20
		case "Admin Key 1":
			k.IsActive = false
		}
	}

	stats, err := svc.GetGlobalStats(ctx)
	if err != nil {
		t.Fatalf("GetGlobalStats failed: %v", err)
	}

	if stats.TotalKeys != 4 {
		t.Errorf("TotalKeys = %d, want 4", stats.TotalKeys)
	}
	if stats.ActiveKeys != 3 {
		t.Errorf("ActiveKeys = %d, want 3", stats.ActiveKeys)
	}
	if stats.RevokedKeys != 1 {
		t.Errorf("RevokedKeys = %d, want 1", stats.RevokedKeys)
	}
	if stats.TotalRequests != 35 {
		t.Errorf("TotalRequests = %d, want 35", stats.TotalRequests)
	}
	if stats.RequestsByScope["read"] != 30 {
		t.Errorf("RequestsByScope read = %d, want 30", stats.RequestsByScope["read"])
	}
	if stats.RequestsByScope["write"] != 5 {
		t.Errorf("RequestsByScope write = %d, want 5", stats.RequestsByScope["write"])
	}
	if len(stats.TopOperators) == 0 {
		t.Error("TopOperators should not be empty")
	}
	if stats.TotalOperators != 2 {
		t.Errorf("TotalOperators = %d, want 2", stats.TotalOperators)
	}
}

// =============================================================================
// Signing Secret (Domain A: API-key-authenticated request signing)
// =============================================================================

func TestAPIKeyService_GenerateKey_SigningSecretDerivedFromFullKey(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	result, err := svc.GenerateKey(ctx, "operator-1", &domain.CreateAPIKeyRequest{
		Name:  "Signing Key",
		Scope: domain.ScopeRead,
	})
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if result.SigningKey == "" {
		t.Fatal("SigningKey should not be empty on key creation")
	}

	sum := sha512.Sum512([]byte(result.FullKey))
	expected := hex.EncodeToString(sum[:])
	if result.SigningKey != expected {
		t.Errorf("SigningKey = %s, want %s", result.SigningKey, expected)
	}
}

func TestAPIKeyService_ValidateKey_ReturnsSigningSecret(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	result, err := svc.GenerateKey(ctx, "operator-1", &domain.CreateAPIKeyRequest{
		Name:  "Validate Signing",
		Scope: domain.ScopeRead,
	})
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	validated, err := svc.ValidateKey(ctx, result.FullKey)
	if err != nil {
		t.Fatalf("ValidateKey failed: %v", err)
	}

	if validated.SigningSecret != result.SigningKey {
		t.Errorf("ValidateKey SigningSecret = %s, want %s", validated.SigningSecret, result.SigningKey)
	}
}

func TestAPIKeyService_RotateKey_RegeneratesSigningSecret(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	original, err := svc.GenerateKey(ctx, "operator-1", &domain.CreateAPIKeyRequest{
		Name:  "Rotate Signing",
		Scope: domain.ScopeRead,
	})
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	rotated, err := svc.RotateKey(ctx, "operator-1", original.ID)
	if err != nil {
		t.Fatalf("RotateKey failed: %v", err)
	}

	if rotated.SigningKey == "" {
		t.Fatal("Rotated SigningKey should not be empty")
	}
	if rotated.SigningKey == original.SigningKey {
		t.Error("Rotated SigningKey should differ from original")
	}

	sum := sha512.Sum512([]byte(rotated.FullKey))
	expected := hex.EncodeToString(sum[:])
	if rotated.SigningKey != expected {
		t.Errorf("Rotated SigningKey = %s, want %s", rotated.SigningKey, expected)
	}
}
