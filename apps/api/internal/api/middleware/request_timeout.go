// Package middleware provides HTTP middleware.
package middleware

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"

	"github.com/gin-gonic/gin"
)

// TimeoutConfig holds configuration for the timeout middleware.
type TimeoutConfig struct {
	RouteTimeouts  map[string]time.Duration
	SkipPaths      []string
	DefaultTimeout time.Duration
}

// DefaultTimeoutConfig returns the default timeout configuration.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		DefaultTimeout: 30 * time.Second,
		RouteTimeouts:  make(map[string]time.Duration),
		SkipPaths:      []string{},
	}
}

// timeoutBufferThreshold is the response size past which the buffering writer
// switches to streaming passthrough. Buffering exists only so a timed-out
// handler cannot write to the client; once a response is this large it is
// certainly not a timeout case, and buffering it would mean holding the whole
// body (potentially a multi-megabyte APK) in memory.
const timeoutBufferThreshold = 64 << 10 // 64 KiB.

// skipPaths that are always excluded: liveness/readiness probes must never be
// buffered or delayed.
var timeoutSkipPaths = []string{"/health", "/healthz", "/metrics"}

// timeoutWriter wraps the real gin.ResponseWriter so the handler's output is
// buffered (and therefore discardable) until the request either completes or the
// deadline fires. Unlike a plain buffering writer it degrades gracefully:
//
//   - Hijack passes straight through, so WebSocket/SSE upgrades are unaffected.
//   - Flush/FlushBuffer and an over-threshold write switch to passthrough, so
//     server-sent events and large file downloads are streamed, not buffered.
//   - CloseNotify forwards the underlying channel.
//
// After the switch to passthrough the response can no longer be discarded, which
// is correct: those cases are never the ones a timeout needs to suppress.
type timeoutWriter struct {
	gin.ResponseWriter

	body      *bytes.Buffer
	headers   http.Header
	threshold int
	timedOut  bool
	passthru  bool
	wroteHdr  bool
	code      int
}

func newTimeoutWriter(w gin.ResponseWriter, threshold int) *timeoutWriter {
	return &timeoutWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		headers:        make(http.Header),
		threshold:      threshold,
	}
}

func (w *timeoutWriter) Header() http.Header { return w.headers }

func (w *timeoutWriter) WriteHeader(code int) {
	if w.timedOut || w.wroteHdr || code == -1 {
		return
	}
	w.wroteHdr = true
	w.code = code
	if w.passthru {
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *timeoutWriter) WriteHeaderNow() {
	if w.timedOut || w.wroteHdr {
		return
	}
	w.wroteHdr = true
	if w.code == 0 {
		w.code = http.StatusOK
	}
	if w.passthru {
		w.ResponseWriter.WriteHeader(w.code)
	}
}

func (w *timeoutWriter) Write(p []byte) (int, error) {
	if w.timedOut {
		return len(p), nil
	}
	if !w.passthru && w.body.Len()+len(p) > w.threshold {
		w.enterPassthrough()
	}
	if w.passthru {
		return w.ResponseWriter.Write(p)
	}
	return w.body.Write(p)
}

func (w *timeoutWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *timeoutWriter) Status() int {
	if w.code == 0 || w.timedOut {
		return w.ResponseWriter.Status()
	}
	return w.code
}

func (w *timeoutWriter) Size() int {
	if w.passthru {
		return w.ResponseWriter.Size()
	}
	return w.body.Len()
}

func (w *timeoutWriter) Written() bool {
	if w.passthru {
		return w.ResponseWriter.Written()
	}
	return w.wroteHdr || w.body.Len() > 0
}

// Flush switches to passthrough and flushes the underlying writer, so SSE and
// other streaming responses are delivered incrementally.
func (w *timeoutWriter) Flush() {
	if w.timedOut {
		return
	}
	w.enterPassthrough()
	w.ResponseWriter.Flush()
}

// Hijack passes through to the real connection for protocol upgrades
// (WebSocket), which cannot work through a buffering writer.
//
// gin's ResponseWriter always advertises Hijack but panics (rather than
// returning an error) when the underlying writer is not hijackable, so the call
// is guarded with a recover and converted to a normal error.
func (w *timeoutWriter) Hijack() (conn net.Conn, rw *bufio.ReadWriter, err error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not support hijacking")
	}

	defer func() {
		if p := recover(); p != nil {
			conn, rw = nil, nil
			err = fmt.Errorf("hijack not supported by underlying writer: %v", p)
		}
	}()

	conn, rw, err = h.Hijack()
	if err == nil {
		w.passthru = true
	}
	return conn, rw, err
}

func (w *timeoutWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.CloseNotify()
}

// Pusher forwards HTTP/2 server push when available.
func (w *timeoutWriter) Pusher() http.Pusher {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p
	}
	return nil
}

// enterPassthrough flushes anything buffered to the real writer and switches to
// direct streaming for all subsequent writes.
func (w *timeoutWriter) enterPassthrough() {
	if w.passthru {
		return
	}
	w.passthru = true
	if !w.wroteHdr {
		w.wroteHdr = true
		if w.code == 0 {
			w.code = http.StatusOK
		}
	}
	dst := w.ResponseWriter.Header()
	for k, vv := range w.headers {
		dst[k] = vv
	}
	w.ResponseWriter.WriteHeader(w.code)
	if w.body.Len() > 0 {
		_, _ = w.ResponseWriter.Write(w.body.Bytes())
		w.body.Reset()
	}
}

// commit writes the buffered response to the real writer on normal completion.
func (w *timeoutWriter) commit() {
	if w.timedOut || w.passthru {
		return
	}
	dst := w.ResponseWriter.Header()
	for k, vv := range w.headers {
		dst[k] = vv
	}
	if w.code == 0 {
		w.code = http.StatusOK
	}
	w.ResponseWriter.WriteHeader(w.code)
	if w.body.Len() > 0 {
		_, _ = w.ResponseWriter.Write(w.body.Bytes())
	}
}

// discard marks the response as timed out so late writes are ignored.
func (w *timeoutWriter) discard() {
	w.timedOut = true
	w.body.Reset()
}

// GinTimeout returns a Gin middleware that aborts a request after a configured
// deadline, responding with the standard structured error envelope.
//
// It replaces github.com/gin-contrib/timeout, whose buffering writer breaks
// WebSocket upgrades and buffers large/streaming responses in memory. This
// implementation keeps the buffering (so a timed-out handler cannot write to the
// client) but its writer passes through on hijack/flush/oversize, so streaming
// and downloads are never buffered.
//
// Paths in config.SkipPaths and the built-in health/metrics paths bypass the
// deadline entirely. Per-route overrides are read from TimeoutConfig.RouteTimeouts.
func GinTimeout(config TimeoutConfig) gin.HandlerFunc {
	timeout := config.DefaultTimeout
	if timeout <= 0 {
		timeout = DefaultTimeoutConfig().DefaultTimeout
	}

	skip := make(map[string]struct{}, len(config.SkipPaths)+len(timeoutSkipPaths))
	for _, p := range config.SkipPaths {
		skip[p] = struct{}{}
	}
	for _, p := range timeoutSkipPaths {
		skip[p] = struct{}{}
	}

	return func(c *gin.Context) {
		if _, ok := skip[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		deadline := timeout
		if rt, ok := config.RouteTimeouts[c.Request.URL.Path]; ok && rt > 0 {
			deadline = rt
		}

		underlying := c.Writer
		tw := newTimeoutWriter(underlying, timeoutBufferThreshold)
		c.Writer = tw

		finish := make(chan struct{})
		panicCh := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicCh <- p
				}
			}()
			c.Next()
			close(finish)
		}()

		select {
		case <-finish:
			tw.commit()
			c.Writer = underlying

		case p := <-panicCh:
			tw.discard()
			c.Writer = underlying
			panic(p)

		case <-time.After(deadline):
			tw.discard()
			c.Writer = underlying
			c.Abort()
			if !underlying.Written() {
				responses.RespondStructured(c, http.StatusGatewayTimeout,
					"the request timed out, please try again")
			}
		}
	}
}
