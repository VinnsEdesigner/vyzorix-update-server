package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestIdempotencyScopeKey_TenantIsolation proves the stored idempotency key is
// scoped to the authenticated tenant, so two tenants that choose the same
// client-supplied idempotency key do not collide (which would let the second
// caller replay the first tenant's cached response).
func TestIdempotencyScopeKey_TenantIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	build := func(orgID, clientID string) string {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/device/inbox", nil)
		if orgID != "" {
			c.Set(ContextKeyOrganizationID, orgID)
		}
		if clientID != "" {
			c.Request.Header.Set("X-Client-ID", clientID)
		}
		return idempotencyScopeKey(c, "/v1/device/inbox", "same-user-key")
	}

	if a, b := build("org-1", ""), build("org-2", ""); a == b {
		t.Fatal("same idempotency key under different orgs must not produce the same scoped key")
	}
	if a, b := build("", "client-1"), build("", "client-2"); a == b {
		t.Fatal("same idempotency key under different clients must not produce the same scoped key")
	}
}

// TestIdempotencyScopeKey_StableForSamePrincipal proves the scoped key is
// deterministic for the same tenant + path + key, so legitimate retries still
// replay correctly.
func TestIdempotencyScopeKey_StableForSamePrincipal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	build := func(path string) string {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, path, nil)
		c.Set(ContextKeyOrganizationID, "org-1")
		return idempotencyScopeKey(c, path, "key-1")
	}

	if build("/v1/device/inbox") != build("/v1/device/inbox") {
		t.Fatal("scoped key must be stable for the same principal, path and key")
	}
}

// TestIdempotencyScopeKey_PathScoped proves the same key on different endpoints
// does not replay an unrelated response.
func TestIdempotencyScopeKey_PathScoped(t *testing.T) {
	gin.SetMode(gin.TestMode)

	build := func(path string) string {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, path, nil)
		c.Set(ContextKeyOrganizationID, "org-1")
		return idempotencyScopeKey(c, path, "key-1")
	}

	if build("/v1/device/inbox") == build("/v1/device/confirm") {
		t.Fatal("same key on different paths must not collide")
	}
}
