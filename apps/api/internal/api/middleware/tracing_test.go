<<<<<<< HEAD

=======
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/tracing"
	"github.com/gin-gonic/gin"
)

func TestTracing_GeneratesAndEchoesTraceID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	Tracing()(c)

	if id := GetTraceID(c); id == "" {
		t.Fatal("trace id should be set on context")
	} else if !tracing.ValidateTraceID(id) {
		t.Fatalf("trace id %q is not well-formed", id)
	}
	if got := w.Header().Get(tracing.TraceIDHeader); got == "" {
		t.Error("X-Trace-ID should be echoed on the response")
	}
}

func TestTracing_AcceptsProvidedTraceID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	c.Request.Header.Set(tracing.TraceIDHeader, "aabbccddeeff00112233445566778899")

	Tracing()(c)

	if got := GetTraceID(c); got != "aabbccddeeff00112233445566778899" {
		t.Errorf("trace id = %q, want the provided id", got)
	}
}

func TestTracing_AcceptsRequestIDAlias(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	c.Request.Header.Set(RequestIDHeader, "112233445566778899aabbccddeeff00")

	Tracing()(c)

	// The legacy X-Request-ID header should be accepted as an inbound alias and
	// surfaced as the trace id; only X-Trace-ID is echoed on the response.
	if got := GetTraceID(c); got != "112233445566778899aabbccddeeff00" {
		t.Errorf("trace id = %q, want the X-Request-ID value", got)
	}
	if got := w.Header().Get(tracing.TraceIDHeader); got == "" {
		t.Error("X-Trace-ID should be echoed even when sourced from X-Request-ID")
	}
	if got := w.Header().Get(RequestIDHeader); got != "" {
		t.Errorf("X-Request-ID should not be echoed, got %q", got)
	}
<<<<<<< HEAD
}
=======
}
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
