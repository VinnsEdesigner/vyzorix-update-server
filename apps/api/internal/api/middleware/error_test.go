
package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	domainerrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/tracing"
	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// runErrorHandler wraps a handler that records a given error and runs the
// ErrorHandler middleware, returning the response body for assertions.
func runErrorHandler(t *testing.T, recordErr error) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	// Stamp a trace id so the structured response includes one.
	c.Set(tracing.ContextKeyTraceID, "trace-test")

	handler := ErrorHandler(slog.Default())
	handler(c)

	// Simulate a handler that recorded an error and left the response unwritten,
	// then re-invoke the post-processing by calling c.Next path is awkward; instead
	// drive the middleware the way gin does: register the error then call the
	// middleware's inner logic by reconstructing. The simplest faithful path is to
	// build a gin engine.
	engine := gin.New()
	engine.Use(Tracing())
	engine.Use(ErrorHandler(slog.Default()))
	engine.POST("/", func(ctx *gin.Context) {
		if recordErr != nil {
			_ = ctx.Error(recordErr)
		}
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
	engine.ServeHTTP(w, req)

	var body map[string]any
	if w.Body.Len() > 0 {
		_ = json.Unmarshal(w.Body.Bytes(), &body)
	}
	return w, body
}

func TestErrorHandler_RendersStructuredValidationError(t *testing.T) {
	verr := domainerrors.NewValidationError([]domainerrors.ValidationDetail{
		{Field: "email", Message: "invalid format"},
		{Field: "password", Message: "too short"},
	})
	w, body := runErrorHandler(t, verr)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured error object, got %v", body)
	}
	if errObj["code"] != string(domainerrors.CodeValidationFailed) {
		t.Errorf("code = %v, want %s", errObj["code"], domainerrors.CodeValidationFailed)
	}
	details, ok := errObj["details"].([]any)
	if !ok || len(details) != 2 {
		t.Fatalf("expected 2 details, got %v", errObj["details"])
	}
	if errObj["trace_id"] == "" {
		t.Error("expected non-empty trace_id")
	}
	if errObj["docs_url"] == "" {
		t.Error("expected non-empty docs_url")
	}
}

func TestErrorHandler_GenericErrorWithoutTraceIDStillWorks(t *testing.T) {
	// A plain (non-validation) recorded error falls through to the generic path.
	w, body := runErrorHandler(t, context.DeadlineExceeded)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured error object, got %v", body)
	}
	if errObj["code"] != string(domainerrors.CodeInternalServerError) {
		t.Errorf("code = %v, want %s", errObj["code"], domainerrors.CodeInternalServerError)
	}
}