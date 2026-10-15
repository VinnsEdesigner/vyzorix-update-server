package middleware

import (
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
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Fatalf("expected body %q, got %q", "ok", w.Body.String())
	}
}

func TestGinTimeout_SlowRequestReturns504(t *testing.T) {
	r := newTimeoutEngine(
		TimeoutConfig{DefaultTimeout: 40 * time.Millisecond},
		func(c *gin.Context) {
			time.Sleep(300 * time.Millisecond)
			c.String(http.StatusOK, "too late")
		},
		"/slow",
	)

	w := doGet(r, "/slow")
	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d (body %q)", w.Code, w.Body.String())
	}
}

func TestGinTimeout_SkipsLongLivedPath(t *testing.T) {
	// A handler on a /stream path must not be killed by the deadline, because
	// WebSocket/SSE connections are long-lived by design.
	r := newTimeoutEngine(
		TimeoutConfig{DefaultTimeout: 30 * time.Millisecond},
		func(c *gin.Context) {
			time.Sleep(120 * time.Millisecond)
			c.String(http.StatusOK, "streamed")
		},
		"/v1/device/:imei/stream",
	)

	w := doGet(r, "/v1/device/abc/stream")
	if w.Code != http.StatusOK {
		t.Fatalf("expected long-lived path to bypass timeout (200), got %d", w.Code)
	}
}

func TestGinTimeout_SkipsStreamingPrefix(t *testing.T) {
	// Static/binary downloads must not be buffered by the timeout writer.
	r := newTimeoutEngine(
		TimeoutConfig{DefaultTimeout: 30 * time.Millisecond},
		func(c *gin.Context) {
			time.Sleep(120 * time.Millisecond)
			c.String(http.StatusOK, "asset")
		},
		"/assets/*filepath",
	)

	w := doGet(r, "/assets/app.js")
	if w.Code != http.StatusOK {
		t.Fatalf("expected streaming path to bypass timeout (200), got %d", w.Code)
	}
}

func TestGinTimeout_SkipsConfiguredPath(t *testing.T) {
	r := newTimeoutEngine(
		TimeoutConfig{
			DefaultTimeout: 30 * time.Millisecond,
			SkipPaths:      []string{"/health"},
		},
		func(c *gin.Context) {
			time.Sleep(120 * time.Millisecond)
			c.String(http.StatusOK, "healthy")
		},
		"/health",
	)

	w := doGet(r, "/health")
	if w.Code != http.StatusOK {
		t.Fatalf("expected skipped path to bypass timeout (200), got %d", w.Code)
	}
}

func TestGinTimeout_PerRouteOverride(t *testing.T) {
	r := newTimeoutEngine(
		TimeoutConfig{
			DefaultTimeout: time.Second,
			RouteTimeouts:  map[string]time.Duration{"/tight": 30 * time.Millisecond},
		},
		func(c *gin.Context) {
			time.Sleep(200 * time.Millisecond)
			c.String(http.StatusOK, "late")
		},
		"/tight",
	)

	w := doGet(r, "/tight")
	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected per-route override to time out (504), got %d", w.Code)
	}
}
