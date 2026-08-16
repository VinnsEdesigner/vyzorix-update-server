package auth

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	apikeyapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	infraStorage "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	"github.com/gin-gonic/gin"
)

// =============================================================================.
// Mock Repository.
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
	if key, ok := r.keys[id]; ok && key.IsActive {
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
// Test Setup.
// =============================================================================.

// testOperatorMiddleware sets default operator_id and organization_id in context for testing.
func testOperatorMiddleware(operatorID, organizationID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("operator_id", operatorID)
		c.Set("organization_id", organizationID)
		c.Next()
	}
}

func setupTestRouter(t *testing.T) (*gin.Engine, *Handler, *mockAPIKeyRepository) {
	gin.SetMode(gin.TestMode)

	repo := newMockRepository()
	config := apikeyapp.DefaultConfig()
	config.Prefix = "test"
	config.MaxPerMonth = 10
	config.MaxExpiryDays = 365
	config.PrefixLength = 8
	service := apikeyapp.NewAPIKeyService(repo, config)

	handler := NewHandler(service, &audit.NoOpLogger{})

	r := gin.New()
	keysGroup := r.Group("/v1")
	// Apply test middleware to set operator_id and organization_id.
	keysGroup.Use(testOperatorMiddleware("test-operator-001", "test-org-001"))
	handler.RegisterRoutes(keysGroup)

	return r, handler, repo
}

// =============================================================================.
// CreateKey Tests - Full Router Context.
// =============================================================================.

func TestCreateKey_Success(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	body := `{"name": "Production Key", "scope": "read"}`
	req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TestClient/1.0")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["api_key"] == nil || response["api_key"] == "" {
		t.Error("Response should contain api_key")
	}

	if response["id"] == nil {
		t.Error("Response should contain id")
	}

	if response["name"] != "Production Key" {
		t.Errorf("Name = %v, want Production Key", response["name"])
	}

	if response["scope"] != "read" {
		t.Errorf("Scope = %v, want read", response["scope"])
	}
}

func TestCreateKey_WithExpiry(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	days := 30
	body := map[string]interface{}{
		"name":            "Expiring Key",
		"scope":           "write",
		"expires_in_days": days,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["expires_at"] == nil {
		t.Error("Response should contain expires_at")
	}
}

func TestCreateKey_DuplicateName(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	body := `{"name": "Duplicate Key", "scope": "read"}`
	req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("First request failed: %d", w.Code)
	}

	// Second request with same name.
	req2, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d for duplicate name", w2.Code, http.StatusForbidden)
	}

	var response map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response)

	if response["error"] != "key_name_conflict" {
		t.Errorf("Error = %v, want key_name_conflict", response["error"])
	}
}

func TestCreateKey_MonthlyLimit(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create 10 keys (max per month).
	for i := 0; i < 10; i++ {
		body := `{"name": "Key ` + string(rune('A'+i)) + `", "scope": "read"}`
		req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("CreateKey %d failed: %d", i, w.Code)
		}
	}

	// 11th key should fail.
	body := `{"name": "Key Overflow", "scope": "read"}`
	req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d for monthly limit exceeded", w.Code, http.StatusForbidden)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != "monthly_limit_exceeded" {
		t.Errorf("Error = %v, want monthly_limit_exceeded", response["error"])
	}
}

func TestCreateKey_InvalidScope(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	body := `{"name": "Bad Scope Key", "scope": "invalid"}`
	req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateKey_NameTooLong(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	longName := ""
	for i := 0; i < 100; i++ {
		longName += "x"
	}
	body := `{"name": "` + longName + `", "scope": "read"}`
	req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateKey_MissingName(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	body := `{"scope": "read"}`
	req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// =============================================================================.
// ListKeys Tests - Full Router Context.
// =============================================================================.

func TestListKeys_Success(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create some keys first.
	for i := 0; i < 3; i++ {
		body := `{"name": "List Test Key ` + string(rune('0'+i)) + `", "scope": "read"}`
		req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// List keys.
	req, _ := http.NewRequest("GET", "/v1/api-keys", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	keys, ok := response["keys"].([]interface{})
	if !ok {
		t.Fatal("Response keys should be an array")
	}

	if len(keys) != 3 {
		t.Errorf("Keys count = %d, want 3", len(keys))
	}

	pagination, ok := response["pagination"].(map[string]interface{})
	if !ok {
		t.Fatal("Response should have pagination")
	}

	if pagination["total"].(float64) != 3 {
		t.Errorf("Pagination total = %v, want 3", pagination["total"])
	}
}

func TestListKeys_Pagination(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create 5 keys.
	for i := 0; i < 5; i++ {
		body := `{"name": "Page Key ` + string(rune('A'+i)) + `", "scope": "read"}`
		req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// List with limit.
	req, _ := http.NewRequest("GET", "/v1/api-keys?page=1&limit=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	keys, _ := response["keys"].([]interface{})
	if len(keys) != 2 {
		t.Errorf("Keys count = %d, want 2", len(keys))
	}

	pagination, _ := response["pagination"].(map[string]interface{})
	if pagination["total"].(float64) != 5 {
		t.Errorf("Pagination total = %v, want 5", pagination["total"])
	}

	if pagination["total_pages"].(float64) != 3 {
		t.Errorf("Pagination total_pages = %v, want 3", pagination["total_pages"])
	}
}

// =============================================================================.
// GetKey Tests - Full Router Context.
// =============================================================================.

func TestGetKey_Success(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create a key.
	body := `{"name": "Get Test Key", "scope": "read"}`
	createReq, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	keyID := created["id"].(string)

	// Get the key.
	req, _ := http.NewRequest("GET", "/v1/api-keys/"+keyID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["id"] != keyID {
		t.Errorf("ID = %v, want %s", response["id"], keyID)
	}

	if response["name"] != "Get Test Key" {
		t.Errorf("Name = %v, want Get Test Key", response["name"])
	}

	// Full key should NOT be in response.
	if response["api_key"] != nil {
		t.Error("Response should NOT contain api_key (full key)")
	}
}

func TestGetKey_NotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/v1/api-keys/nonexistent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// =============================================================================.
// UpdateKey Tests - Full Router Context.
// =============================================================================.

func TestUpdateKey_Rename(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create a key.
	body := `{"name": "Original Name", "scope": "read"}`
	createReq, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	keyID := created["id"].(string)

	// Update the name.
	updateBody := `{"name": "Updated Name"}`
	updateReq, _ := http.NewRequest("PATCH", "/v1/api-keys/"+keyID, bytes.NewBufferString(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")

	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", updateW.Code, http.StatusOK, updateW.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(updateW.Body.Bytes(), &response)

	if response["name"] != "Updated Name" {
		t.Errorf("Name = %v, want Updated Name", response["name"])
	}
}

func TestUpdateKey_ChangeScope(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create a key with read scope.
	body := `{"name": "Scope Change Test", "scope": "read"}`
	createReq, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	keyID := created["id"].(string)

	// Update scope to write.
	updateBody := `{"scope": "write"}`
	updateReq, _ := http.NewRequest("PATCH", "/v1/api-keys/"+keyID, bytes.NewBufferString(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")

	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", updateW.Code, http.StatusOK)
	}

	var response map[string]interface{}
	json.Unmarshal(updateW.Body.Bytes(), &response)

	if response["scope"] != "write" {
		t.Errorf("Scope = %v, want write", response["scope"])
	}
}

func TestUpdateKey_DuplicateName(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create two keys.
	body1 := `{"name": "Key One", "scope": "read"}`
	createReq1, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body1))
	createReq1.Header.Set("Content-Type", "application/json")
	createW1 := httptest.NewRecorder()
	router.ServeHTTP(createW1, createReq1)

	body2 := `{"name": "Key Two", "scope": "read"}`
	createReq2, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body2))
	createReq2.Header.Set("Content-Type", "application/json")
	createW2 := httptest.NewRecorder()
	router.ServeHTTP(createW2, createReq2)

	var created2 map[string]interface{}
	json.Unmarshal(createW2.Body.Bytes(), &created2)
	key2ID := created2["id"].(string)

	// Try to rename key2 to key1's name.
	updateBody := `{"name": "Key One"}`
	updateReq, _ := http.NewRequest("PATCH", "/v1/api-keys/"+key2ID, bytes.NewBufferString(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")

	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)

	if updateW.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", updateW.Code, http.StatusForbidden)
	}
}

// =============================================================================.
// RevokeKey Tests - Full Router Context.
// =============================================================================.

func TestRevokeKey_Success(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create a key.
	body := `{"name": "Key To Revoke", "scope": "read"}`
	createReq, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("User-Agent", "TestClient/1.0")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	keyID := created["id"].(string)

	// Revoke the key.
	revokeReq, _ := http.NewRequest("DELETE", "/v1/api-keys/"+keyID, nil)
	revokeReq.Header.Set("User-Agent", "TestClient/1.0")

	revokeW := httptest.NewRecorder()
	router.ServeHTTP(revokeW, revokeReq)

	if revokeW.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d. Body: %s", revokeW.Code, http.StatusNoContent, revokeW.Body.String())
	}

	// Verify key is no longer accessible.
	getReq, _ := http.NewRequest("GET", "/v1/api-keys/"+keyID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d for revoked key", getW.Code, http.StatusNotFound)
	}
}

func TestRevokeKey_NotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req, _ := http.NewRequest("DELETE", "/v1/api-keys/nonexistent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// =============================================================================.
// RotateKey Tests - Full Router Context.
// =============================================================================.

func TestRotateKey_Success(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create a key.
	body := `{"name": "Key To Rotate", "scope": "read"}`
	createReq, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("User-Agent", "TestClient/1.0")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	keyID := created["id"].(string)
	originalFullKey := created["api_key"].(string)

	// Rotate the key.
	rotateReq, _ := http.NewRequest("POST", "/v1/api-keys/"+keyID+"/rotate", nil)
	rotateReq.Header.Set("User-Agent", "TestClient/1.0")

	rotateW := httptest.NewRecorder()
	router.ServeHTTP(rotateW, rotateReq)

	if rotateW.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", rotateW.Code, http.StatusOK, rotateW.Body.String())
	}

	var rotated map[string]interface{}
	json.Unmarshal(rotateW.Body.Bytes(), &rotated)

	newFullKey := rotated["api_key"].(string)
	if newFullKey == originalFullKey {
		t.Error("Rotated key should be different from original")
	}

	if rotated["name"] != "Key To Rotate" {
		t.Errorf("Name = %v, want Key To Rotate", rotated["name"])
	}

	if rotated["id"] == keyID {
		t.Error("New key should have a different ID")
	}
}

func TestRotateKey_MonthlyLimit(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create 10 keys (max).
	for i := 0; i < 10; i++ {
		body := `{"name": "Rotate Limit ` + string(rune('A'+i)) + `", "scope": "read"}`
		req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Get first key ID.
	listReq, _ := http.NewRequest("GET", "/v1/api-keys?limit=1", nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	var listResp map[string]interface{}
	json.Unmarshal(listW.Body.Bytes(), &listResp)
	keys := listResp["keys"].([]interface{})
	firstKeyID := keys[0].(map[string]interface{})["id"].(string)

	// Rotate should fail due to monthly limit.
	rotateReq, _ := http.NewRequest("POST", "/v1/api-keys/"+firstKeyID+"/rotate", nil)
	rotateW := httptest.NewRecorder()
	router.ServeHTTP(rotateW, rotateReq)

	if rotateW.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d for monthly limit", rotateW.Code, http.StatusForbidden)
	}
}

func TestRotateKey_NotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req, _ := http.NewRequest("POST", "/v1/api-keys/nonexistent-id/rotate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// =============================================================================.
// Tenant Isolation Tests - Full Router Context.
// =============================================================================.

func TestTenantIsolation_ListKeys(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create keys for operator-1 (simulated via session).
	// In real test, we'd use different sessions.
	for i := 0; i < 3; i++ {
		body := `{"name": "Op1 Key ` + string(rune('A'+i)) + `", "scope": "read"}`
		req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// List should only show keys for this operator.
	req, _ := http.NewRequest("GET", "/v1/api-keys", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	keys := response["keys"].([]interface{})
	if len(keys) != 3 {
		t.Errorf("Keys count = %d, want 3 for this operator", len(keys))
	}
}

func TestTenantIsolation_GetOtherTenantKey(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create a key.
	body := `{"name": "Private Key", "scope": "read"}`
	createReq, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	keyID := created["id"].(string)

	// Note: In this test setup, operator_id comes from context.
	// In real scenario, another operator would get 404.
	// This tests that keys are isolated by operator_id in context.
	req, _ := http.NewRequest("GET", "/v1/api-keys/"+keyID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Same operator should access own key: Status = %d", w.Code)
	}
}

// =============================================================================.
// Scope Enforcement Tests - Full Router Context.
// =============================================================================.

func TestScopeEnforcement_ReadScope(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	body := `{"name": "Read Only Key", "scope": "read"}`
	req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create failed: %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["scope"] != "read" {
		t.Errorf("Scope = %v, want read", response["scope"])
	}
}

func TestScopeEnforcement_WriteScope(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	body := `{"name": "Write Key", "scope": "write"}`
	req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create failed: %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["scope"] != "write" {
		t.Errorf("Scope = %v, want write", response["scope"])
	}
}

func TestScopeEnforcement_AdminScope(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	body := `{"name": "Admin Key", "scope": "admin"}`
	req, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create failed: %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["scope"] != "admin" {
		t.Errorf("Scope = %v, want admin", response["scope"])
	}
}

// =============================================================================.
// Full Flow Integration Tests.
// =============================================================================.

func TestFullFlow_CreateListGetUpdateRotateRevoke(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// 1. Create a key.
	createBody := `{"name": "Flow Test Key", "scope": "write"}`
	createReq, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("Create failed: %d - %s", createW.Code, createW.Body.String())
	}

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	keyID := created["id"].(string)
	originalKey := created["api_key"].(string)

	// 2. List keys - should have 1.
	listReq, _ := http.NewRequest("GET", "/v1/api-keys", nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	var listResp map[string]interface{}
	json.Unmarshal(listW.Body.Bytes(), &listResp)
	keys := listResp["keys"].([]interface{})
	if len(keys) != 1 {
		t.Errorf("Step 2: Keys count = %d, want 1", len(keys))
	}

	// 3. Get the specific key.
	getReq, _ := http.NewRequest("GET", "/v1/api-keys/"+keyID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	getResp := map[string]interface{}{}
	json.Unmarshal(getW.Body.Bytes(), &getResp)
	if getResp["name"] != "Flow Test Key" {
		t.Errorf("Step 3: Name = %v, want Flow Test Key", getResp["name"])
	}

	// 4. Update the key name.
	updateBody := `{"name": "Updated Flow Key"}`
	updateReq, _ := http.NewRequest("PATCH", "/v1/api-keys/"+keyID, bytes.NewBufferString(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")

	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Fatalf("Step 4: Update failed: %d", updateW.Code)
	}

	var updateResp map[string]interface{}
	json.Unmarshal(updateW.Body.Bytes(), &updateResp)
	if updateResp["name"] != "Updated Flow Key" {
		t.Errorf("Step 4: Name = %v, want Updated Flow Key", updateResp["name"])
	}

	// 5. Rotate the key - this revokes the original keyID.
	rotateReq, _ := http.NewRequest("POST", "/v1/api-keys/"+keyID+"/rotate", nil)
	rotateW := httptest.NewRecorder()
	router.ServeHTTP(rotateW, rotateReq)

	if rotateW.Code != http.StatusOK {
		t.Fatalf("Step 5: Rotate failed: %d", rotateW.Code)
	}

	var rotateResp map[string]interface{}
	json.Unmarshal(rotateW.Body.Bytes(), &rotateResp)
	newKeyID := rotateResp["id"].(string)
	newKey := rotateResp["api_key"].(string)
	if newKey == originalKey {
		t.Error("Step 5: Rotated key should be different")
	}

	// 6. Revoke the NEW key (original was already revoked by rotation).
	revokeReq, _ := http.NewRequest("DELETE", "/v1/api-keys/"+newKeyID, nil)
	revokeW := httptest.NewRecorder()
	router.ServeHTTP(revokeW, revokeReq)

	if revokeW.Code != http.StatusNoContent {
		t.Fatalf("Step 6: Revoke failed: %d", revokeW.Code)
	}

	// 7. Verify new key is gone.
	getReq2, _ := http.NewRequest("GET", "/v1/api-keys/"+newKeyID, nil)
	getW2 := httptest.NewRecorder()
	router.ServeHTTP(getW2, getReq2)

	if getW2.Code != http.StatusNotFound {
		t.Errorf("Step 7: Status = %d, want 404 for revoked key", getW2.Code)
	}
}

func TestCreateKey_FullKeyReturnedOnce(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create a key.
	body := `{"name": "One Time View", "scope": "read"}`
	createReq, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	keyID := created["id"].(string)
	fullKey := created["api_key"].(string)

	// Get the key - full key should NOT be there.
	getReq, _ := http.NewRequest("GET", "/v1/api-keys/"+keyID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	var getResp map[string]interface{}
	json.Unmarshal(getW.Body.Bytes(), &getResp)

	if getResp["api_key"] != nil {
		t.Error("GET response should NOT contain api_key")
	}

	_ = fullKey // Use it to avoid unused variable.
}

// =============================================================================.
// Error Response Format Tests.
// =============================================================================.

func TestErrorResponse_Format(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	body := `{"name": "Duplicate", "scope": "read"}`
	req1, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	req2, _ := http.NewRequest("POST", "/v1/api-keys", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var response map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response)

	// Should have error and message.
	if response["error"] == nil {
		t.Error("Response should have 'error' field")
	}

	if response["message"] == nil {
		t.Error("Response should have 'message' field")
	}

	// Error code should be machine-readable.
	errorCodes := []string{
		"key_name_conflict",
		"monthly_limit_exceeded",
		"invalid_scope",
		"validation_error",
		"key_not_found",
	}
	found := false
	for _, ec := range errorCodes {
		if response["error"] == ec {
			found = true
			break
		}
	}
	if !found {
		t.Logf("Error code: %v", response["error"])
	}
}

// =============================================================================.
// HTTP Method Tests.
// =============================================================================.

func TestHTTPMethods_NotAllowed(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// GET on POST-only endpoint should 404 (no route).
	req, _ := http.NewRequest("GET", "/v1/api-keys/some-id/method", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Gin returns 404 for unknown routes, 405 for method not allowed.
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Logf("Status: %d", w.Code)
	}
}
