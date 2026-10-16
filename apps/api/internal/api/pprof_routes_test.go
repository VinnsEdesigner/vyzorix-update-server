package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

// pprofTestServer builds a Server whose admin group has the same auth chain as
// production (cookie auth is skipped here; the super-admin check is what matters)
// and the pprof routes registered.
func pprofTestServer(t *testing.T, enablePprof bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	s := &Server{
		config: config.Config{EnablePprof: enablePprof},
		log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	r := gin.New()
	admin := r.Group("/v1/admin")

	if enablePprof {
		// Inject a super-admin operator + org context, then the real guard.
		admin.Use(func(c *gin.Context) {
			orgID := c.GetHeader("X-Organization-ID")
			op := &operator.Operator{
				ID: "op-1",
				Memberships: []*organization.OrganizationMember{
					organization.NewMember("m-1", orgID, "op-1", organization.RoleSuperAdmin),
				},
			}
			c.Set(middleware.ContextKeyOperator, op)
			c.Set(middleware.ContextKeyOrganizationID, orgID)
			c.Next()
		})
		admin.Use(middleware.RequireSuperAdmin())
	}

	s.registerPprofRoutes(admin)

	return r
}

// TestPprofRoutes_DisabledByDefault proves that with the flag off no pprof route
// is registered, so the paths fall through to the SPA fallback rather than
// serving process memory.
func TestPprofRoutes_DisabledByDefault(t *testing.T) {
	r := pprofTestServer(t, false)

	for _, path := range []string{"/v1/admin/pprof/", "/v1/admin/pprof/heap", "/v1/admin/pprof/profile"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("X-Organization-ID", "org-1")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("%s should not be registered when ENABLE_PPROF is off (got %d)", path, w.Code)
		}
	}
}

// TestPprofRoutes_RequireSuperAdmin proves the enabled routes are behind the
// super-admin guard: without an operator in context they must not serve a profile.
func TestPprofRoutes_RequireSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := &Server{
		config: config.Config{EnablePprof: true},
		log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	r := gin.New()
	admin := r.Group("/v1/admin")
	admin.Use(middleware.RequireSuperAdmin()) // no operator injected -> must reject
	s.registerPprofRoutes(admin)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/pprof/heap", nil)
	req.Header.Set("X-Organization-ID", "org-1")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated pprof access must be 401, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "heap") {
		t.Fatal("a profile must not be served to an unauthenticated caller")
	}
}

// TestPprofRoutes_ServeProfiles proves that when enabled and authenticated the
// endpoints return real profile data: the goroutine profile as text, and the heap
// profile in the pprof wire format (gzip).
func TestPprofRoutes_ServeProfiles(t *testing.T) {
	r := pprofTestServer(t, true)

	t.Run("goroutine text profile", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/pprof/goroutine?debug=1", nil)
		req.Header.Set("X-Organization-ID", "org-1")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("goroutine profile: got %d", w.Code)
		}
		body := w.Body.String()
		// A debug=1 goroutine profile is human-readable text mentioning goroutines.
		if !strings.Contains(body, "goroutine") {
			t.Fatalf("expected a goroutine profile, got: %.120s", body)
		}
	})

	t.Run("heap binary profile", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/pprof/heap", nil)
		req.Header.Set("X-Organization-ID", "org-1")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("heap profile: got %d", w.Code)
		}
		body := w.Body.Bytes()
		if len(body) == 0 {
			t.Fatal("heap profile was empty")
		}
		// The pprof wire format is gzip-compressed.
		if len(body) < 2 || body[0] != 0x1f || body[1] != 0x8b {
			t.Fatalf("heap profile is not in gzip/pprof format (first bytes: % x)", body[:min(4, len(body))])
		}
	})

	t.Run("cmdline", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/pprof/cmdline", nil)
		req.Header.Set("X-Organization-ID", "org-1")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("cmdline: got %d", w.Code)
		}
		if w.Body.Len() == 0 {
			t.Fatal("cmdline was empty")
		}
	})
}
