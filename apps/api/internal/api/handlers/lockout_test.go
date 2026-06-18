// Package handlers provides tests for HTTP handlers.
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

func TestLockoutHandler_GetLockoutStatus_Unauthorized(t *testing.T) {
	// Create handler with nil dependencies - will panic if called but we test auth check first.
	handler := NewLockoutHandler(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/auth/lockout/status", nil)

	// Should return unauthorized since no operator in context.
	handler.GetLockoutStatus(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLockoutHandler_UnlockAccount_Unauthorized(t *testing.T) {
	handler := NewLockoutHandler(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/admin/lockout/unlock/op-123", nil)

	// Should return unauthorized since no operator in context.
	handler.UnlockAccount(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLockoutHandler_UnlockAccount_NonAdmin(t *testing.T) {
	handler := NewLockoutHandler(nil, nil)

	operator := &models.Operator{
		ID:    "op-test-123",
		Email: "test@example.com",
		Role:  "operator", // Not admin
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("operator", operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/admin/lockout/unlock/op-456", nil)
	c.Params = []gin.Param{{Key: "operator_id", Value: "op-456"}}

	handler.UnlockAccount(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestLockoutHandler_UnlockAccount_MissingOperatorID(t *testing.T) {
	handler := NewLockoutHandler(nil, nil)

	operator := &models.Operator{
		ID:    "op-test-123",
		Email: "test@example.com",
		Role:  "admin",
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("operator", operator)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/admin/lockout/unlock/", nil)
	c.Params = []gin.Param{{Key: "operator_id", Value: ""}}

	handler.UnlockAccount(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestLockoutStatusResponse_Struct(t *testing.T) {
	resp := LockoutStatusResponse{
		Locked:   true,
		Reason:   "Too many failed attempts",
		Until:    1234567890,
		Attempts: 5,
	}

	if !resp.Locked {
		t.Error("Expected Locked to be true")
	}
	if resp.Reason != "Too many failed attempts" {
		t.Errorf("Expected reason, got %s", resp.Reason)
	}
}

func TestLockoutInfoResponse_Struct(t *testing.T) {
	resp := LockoutInfoResponse{
		Success: true,
		Message: "Account unlocked",
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Message != "Account unlocked" {
		t.Errorf("Expected message, got %s", resp.Message)
	}
}
