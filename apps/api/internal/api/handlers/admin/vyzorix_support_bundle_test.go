package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSupportBundle_GetBundle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSupportBundleHandler(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/support-bundle", nil)
	h.GetBundle(c)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if body == "" {
		t.Error("expected non-empty body")
	}
	if !contains(body, "generated_at") {
		t.Error("expected generated_at in bundle")
	}
	if !contains(body, "go_version") {
		t.Error("expected go_version in bundle")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
