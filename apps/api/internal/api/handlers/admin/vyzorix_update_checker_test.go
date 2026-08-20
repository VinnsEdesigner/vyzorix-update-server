package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUpdateChecker_Check_ReturnsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUpdateCheckerHandler("1.0.0", "VinnsEdesigner/vyzorix-update-server")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/updates/check", nil)
	h.Check(c)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["current_version"] != "1.0.0" {
		t.Errorf("expected current_version 1.0.0, got %v", resp["current_version"])
	}
}

func TestUpdateChecker_Check_HasUpdateAvailableField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUpdateCheckerHandler("1.0.0", "VinnsEdesigner/vyzorix-update-server")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/updates/check", nil)
	h.Check(c)
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if _, exists := resp["update_available"]; !exists {
		t.Error("expected update_available field in response")
	}
}
