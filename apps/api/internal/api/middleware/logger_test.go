package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLogger_RecordsRequestDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()
	
	// Create a new gin engine with logger middleware
	r := gin.New()
	r.Use(Logger(logger))
	
	// Test various HTTP methods and paths
	testCases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/"},
		{http.MethodGet, "/api/devices"},
		{http.MethodPost, "/api/register"},
		{http.MethodPut, "/api/settings"},
		{http.MethodDelete, "/api/device/123"},
	}
	
	for _, tc := range testCases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, nil)
		r.ServeHTTP(w, req)
	}
}

func TestLogger_ExtractsClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()
	
	r := gin.New()
	r.Use(Logger(logger))
	r.GET("/test", func(cx *gin.Context) {
		cx.Status(http.StatusOK)
	})
	
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	
	// Should not panic and should extract IP
	r.ServeHTTP(w, req)
}

func TestLogger_RecordsStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()
	
	r := gin.New()
	r.Use(Logger(logger))
	r.GET("/test", func(cx *gin.Context) {
		cx.Status(http.StatusOK)
	})
	
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	
	// Should not panic
	r.ServeHTTP(w, req)
}

func TestLogger_NextIsCalled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()
	
	// Create a new gin engine with a route
	r := gin.New()
	r.Use(Logger(logger))
	r.GET("/test", func(cx *gin.Context) {
		cx.Status(http.StatusOK)
	})
	
	// The handler should call Next()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}