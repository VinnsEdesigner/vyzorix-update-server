package middleware

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
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
	if strings.Contains(w.Body.String(), "too late") {
		t.Fatal("timed-out handler body must not reach the client")
	}
}

func TestGinTimeout_SkipsConfiguredPath(t *testing.T) {
	r := newTimeoutEngine(
		TimeoutConfig{DefaultTimeout: 30 * time.Millisecond, SkipPaths: []string{"/health"}},
		func(c *gin.Context) {
			time.Sleep(120 * time.Millisecond)
			c.String(http.StatusOK, "healthy")
		},
		"/health",
	)

	if w := doGet(r, "/health"); w.Code != http.StatusOK {
		t.Fatalf("expected skipped path to bypass timeout (200), got %d", w.Code)
	}
}

func TestGinTimeout_SkipsHealthWithoutConfig(t *testing.T) {
	// /metrics bypasses by default, no SkipPaths required.
	r := newTimeoutEngine(
		TimeoutConfig{DefaultTimeout: 30 * time.Millisecond},
		func(c *gin.Context) {
			time.Sleep(120 * time.Millisecond)
			c.String(http.StatusOK, "ok")
		},
		"/metrics",
	)

	if w := doGet(r, "/metrics"); w.Code != http.StatusOK {
		t.Fatalf("expected /metrics to bypass timeout (200), got %d", w.Code)
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

	if w := doGet(r, "/tight"); w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected per-route override to time out (504), got %d", w.Code)
	}
}

// TestTimeoutWriter_OversizeSwitchesToPassthrough proves a large response is not
// buffered in memory - the defect the previous library implementation had.
func TestTimeoutWriter_OversizeSwitchesToPassthrough(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	tw := newTimeoutWriter(c.Writer, 16) // tiny threshold to force the switch
	big := strings.Repeat("x", 64)
	if _, err := tw.Write([]byte(big)); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !tw.passthru {
		t.Fatal("writer should have switched to passthrough past the threshold")
	}
	if tw.body.Len() != 0 {
		t.Fatalf("buffer should be empty after passthrough, has %d bytes", tw.body.Len())
	}
	if rec.Body.String() != big {
		t.Fatal("oversize body should have been streamed to the real writer")
	}
}

// TestTimeoutWriter_FlushSwitchesToPassthrough covers SSE-style streaming.
func TestTimeoutWriter_FlushSwitchesToPassthrough(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	tw := newTimeoutWriter(c.Writer, timeoutBufferThreshold)

	if _, err := tw.Write([]byte("chunk")); err != nil {
		t.Fatalf("write: %v", err)
	}
	tw.Flush()

	if !tw.passthru {
		t.Fatal("Flush should switch the writer to passthrough")
	}
	if rec.Body.String() != "chunk" {
		t.Fatalf("flushed content should be on the real writer, got %q", rec.Body.String())
	}
}

// hijackableWriter implements gin.ResponseWriter with hijack support so the
// passthrough path can be exercised.
type hijackableWriter struct {
	gin.ResponseWriter
	hijacked bool
}

func (h *hijackableWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.hijacked = true
	return nil, nil, nil
}

func TestTimeoutWriter_HijackDelegates(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	inner := &hijackableWriter{ResponseWriter: c.Writer}

	tw := newTimeoutWriter(inner, timeoutBufferThreshold)
	if _, _, err := tw.Hijack(); err != nil {
		t.Fatalf("hijack should delegate: %v", err)
	}
	if !inner.hijacked {
		t.Fatal("hijack was not delegated to the underlying writer")
	}
	if !tw.passthru {
		t.Fatal("hijack should switch the writer to passthrough")
	}
}

// TestTimeoutWriter_NonHijackableReturnsError proves a clean error (not a panic)
// when the underlying writer cannot hijack.
func TestTimeoutWriter_NonHijackableReturnsError(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	tw := newTimeoutWriter(c.Writer, timeoutBufferThreshold)

	if _, _, err := tw.Hijack(); err == nil {
		t.Fatal("expected an error hijacking a non-hijackable writer")
	}
}
