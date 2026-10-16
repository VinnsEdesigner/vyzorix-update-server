package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newTimeoutEngine(cfg TimeoutConfig, handler gin.HandlerFunc, path string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(GinTimeout(cfg))
	r.GET(path, handler)
	return r
}

func doGet(r *gin.Engine, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestGinTimeout_FastRequestPassesThrough(t *testing.T) {
	r := newTimeoutEngine(
		TimeoutConfig{DefaultTimeout: time.Second},
		func(c *gin.Context) { c.String(http.StatusOK, "ok") },
		"/fast",
	)

	w := doGet(r, "/fast")
	if w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Fatalf("expected 200 ok, got %d %q", w.Code, w.Body.String())
	}
}

// TestGinTimeout_ContextDeadlineExposed proves the middleware puts the deadline on
// the request context, so a cooperative handler can observe cancellation.
func TestGinTimeout_ContextDeadlineExposed(t *testing.T) {
	var hadDeadline bool

	r := newTimeoutEngine(
		TimeoutConfig{DefaultTimeout: 50 * time.Millisecond},
		func(c *gin.Context) {
			_, hadDeadline = c.Request.Context().Deadline()
			c.String(http.StatusOK, "ok")
		},
		"/ctx",
	)

	doGet(r, "/ctx")
	if !hadDeadline {
		t.Fatal("the request context should carry a deadline")
	}
}

// TestGinTimeout_HandlerSeesCancellation proves a handler that waits on the context
// is unblocked when the deadline fires, and the middleware then reports 504.
func TestGinTimeout_HandlerSeesCancellation(t *testing.T) {
	r := newTimeoutEngine(
		TimeoutConfig{DefaultTimeout: 40 * time.Millisecond},
		func(c *gin.Context) {
			<-c.Request.Context().Done() // cooperative: stop when the deadline fires
		},
		"/slow",
	)

	w := doGet(r, "/slow")
	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504 once the context deadline fires, got %d", w.Code)
	}
}

// TestGinTimeout_SkipsHealthPaths proves probes bypass the deadline.
func TestGinTimeout_SkipsHealthPaths(t *testing.T) {
	r := newTimeoutEngine(
		TimeoutConfig{DefaultTimeout: time.Millisecond},
		func(c *gin.Context) {
			time.Sleep(30 * time.Millisecond)
			c.String(http.StatusOK, "healthy")
		},
		"/metrics",
	)

	if w := doGet(r, "/metrics"); w.Code != http.StatusOK {
		t.Fatalf("skipped path should bypass the deadline, got %d", w.Code)
	}
}

// TestGinTimeout_ConcurrentRequests is the regression guard for the data race that
// the previous goroutine-based implementation had: many concurrent requests
// through the middleware must be race-free (run with -race).
func TestGinTimeout_ConcurrentRequests(t *testing.T) {
	r := newTimeoutEngine(
		TimeoutConfig{DefaultTimeout: 50 * time.Millisecond},
		func(c *gin.Context) {
			if c.Query("wait") == "1" {
				<-c.Request.Context().Done()
				return
			}
			c.String(http.StatusOK, "ok")
		},
		"/work",
	)

	done := make(chan struct{})

	for i := 0; i < 20; i++ {
		go func(i int) {
			defer func() { done <- struct{}{} }()
			target := "/work"
			if i%2 == 0 {
				target = "/work?wait=1"
			}
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, target, nil).WithContext(context.Background())
			r.ServeHTTP(w, req)
		}(i)
	}
	for i := 0; i < 20; i++ {
		<-done
	}
}
