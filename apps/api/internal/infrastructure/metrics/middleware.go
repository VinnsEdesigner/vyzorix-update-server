// Package metrics provides middleware for collecting HTTP metrics.
package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware returns a Gin middleware that collects HTTP metrics.
func Middleware(m *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		m.RecordHTTPRequest(duration, statusCode)
	}
}

// MetricsHandler handles the /metrics endpoint.
type MetricsHandler struct {
	metrics *Metrics
}

// NewMetricsHandler creates a new metrics handler.
func NewMetricsHandler(m *Metrics) *MetricsHandler {
	return &MetricsHandler{metrics: m}
}

// Handle returns the metrics in Prometheus format.
func (h *MetricsHandler) Handle(c *gin.Context) {
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.String(200, h.metrics.PrometheusOutput())
}
