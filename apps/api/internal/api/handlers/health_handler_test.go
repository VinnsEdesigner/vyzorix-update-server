// Package handlers provides tests for HTTP handlers.
package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Live(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewHealthHandler(nil, nil)
	r.GET("/health/live", handler.Live)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

func TestHealthHandler_Ready_NoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewHealthHandler(nil, nil)
	r.GET("/health/ready", handler.Ready)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without DB, should return degraded or error.
	assert.True(t, w.Code == http.StatusServiceUnavailable || w.Code == http.StatusOK)
}

func TestHealthHandler_Secure_NoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewHealthHandler(nil, nil)
	r.GET("/health/secure", handler.Secure)

	req := httptest.NewRequest(http.MethodGet, "/health/secure", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without DB, should return degraded.
	assert.True(t, w.Code == http.StatusServiceUnavailable || w.Code == http.StatusOK)
}
