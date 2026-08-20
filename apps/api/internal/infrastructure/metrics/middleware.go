// Package metrics provides middleware for collecting HTTP metrics.
package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Middleware returns a Gin middleware that collects HTTP metrics with
// route-template labels (c.FullPath() avoids high-cardinality literal paths).
func Middleware(m *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		m.RecordHTTPRequest(route, c.Request.Method, c.Writer.Status(), time.Since(start))
	}
}

// MetricsHandler handles the /metrics endpoint.
type MetricsHandler struct {
	http.Handler
}

// NewMetricsHandler creates a new metrics handler.
func NewMetricsHandler(m *Metrics) *MetricsHandler {
	return &MetricsHandler{Handler: promhttp.HandlerFor(m.Registry(), promhttp.HandlerOpts{})}
}

// Handle serves the prometheus-formatted metrics.
func (h *MetricsHandler) Handle(c *gin.Context) {
	h.ServeHTTP(c.Writer, c.Request)
}
