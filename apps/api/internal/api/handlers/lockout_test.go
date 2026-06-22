package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

func TestLockoutHandler_GetLockoutStatus_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := auth.NewLockoutHandler(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/auth/lockout/status", nil)

	handler.GetLockoutStatus(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLockoutHandler_UnlockAccount_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := auth.NewLockoutHandler(nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/admin/lockout/unlock/op-123", nil)

	handler.UnlockAccount(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLockoutMiddleware_NewLockout(t *testing.T) {
	config := middleware.LockoutConfig{
		Enabled:           true,
		MaxAttempts:       5,
		LockoutDuration:   5 * time.Minute,
		MaxLockoutDuration: 30 * time.Minute,
	}
	lockout := middleware.NewLockout(config)
	if lockout == nil {
		t.Error("NewLockout should not return nil")
	}
}
