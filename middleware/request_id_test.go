package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	handler := RequestIDMiddleware()
	handler(c)

	// Check that request ID is set in context
	id := GetRequestID(c)
	if id == "" {
		t.Error("Request ID should be generated")
	}

	// Check that response header is set
	if w.Header().Get(RequestIDHeader) == "" {
		t.Error("X-Request-ID header should be set")
	}
}

func TestRequestIDMiddleware_UsesProvidedID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set(RequestIDHeader, "custom-request-id-123")

	handler := RequestIDMiddleware()
	handler(c)

	// Check that provided request ID is used
	id := GetRequestID(c)
	if id != "custom-request-id-123" {
		t.Errorf("Request ID = %s, want custom-request-id-123", id)
	}
}

func TestRequestIDMiddleware_SetsResponseHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	handler := RequestIDMiddleware()
	handler(c)

	// Check response header
	header := w.Header().Get(RequestIDHeader)
	if header == "" {
		t.Error("Response X-Request-ID header should be set")
	}

	// Verify it's the same as context
	if header != GetRequestID(c) {
		t.Error("Response header should match context request ID")
	}
}

func TestGetRequestID_NoID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	id := GetRequestID(c)
	if id != "" {
		t.Errorf("GetRequestID on empty context = %s, want empty", id)
	}
}

func TestRequestIDFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	handler := RequestIDMiddleware()
	handler(c)

	id := GetRequestID(c)

	// Request ID should be a 32-character hex string (16 bytes)
	if len(id) != 32 {
		t.Errorf("Request ID length = %d, want 32", len(id))
	}

	// Should be valid hex
	for _, ch := range id {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			t.Error("Request ID should be lowercase hex")
		}
	}
}