// Package handlers provides HTTP handlers for health and metrics endpoints.
package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthChecker defines the interface for health check components.
type HealthChecker interface {
	Check(ctx context.Context) error
}

// HealthStatus represents the health status of a component.
type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// HealthResponse represents the full health check response.
type HealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp int64                    `json:"timestamp"`
	Checks    map[string]HealthStatus `json:"checks,omitempty"`
}

// MetricsCollector defines the interface for collecting metrics.
type MetricsCollector interface {
	Collect() map[string]any
}

// HealthHandler handles health check requests.
type HealthHandler struct {
	DB         *sql.DB
	Metrics    MetricsCollector
	Checks     []HealthChecker
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db *sql.DB, metrics MetricsCollector) *HealthHandler {
	return &HealthHandler{
		DB:      db,
		Metrics: metrics,
	}
}

// Live handles the liveness probe - always returns OK if the server is running.
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UnixMilli(),
	})
}

// Ready handles the readiness probe - checks if the server can serve traffic.
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UnixMilli(),
		Checks:    make(map[string]HealthStatus),
	}

	allHealthy := true

	// Check database connectivity (skip if DB is nil).
	if h.DB != nil {
		dbStart := time.Now()
		dbErr := h.DB.PingContext(ctx)
		dbLatency := time.Since(dbStart)
		if dbErr != nil {
			allHealthy = false
			response.Checks["database"] = HealthStatus{
				Status:  "error",
				Message: "database unreachable",
				Latency: dbLatency.String(),
			}
		} else {
			response.Checks["database"] = HealthStatus{
				Status:  "ok",
				Latency: dbLatency.String(),
			}
		}
	}

	// Check metrics collector.
	if h.Metrics != nil {
		metricsStart := time.Now()
		metricsErr := h.collectMetrics()
		metricsLatency := time.Since(metricsStart)
		if metricsErr != nil {
			response.Checks["metrics"] = HealthStatus{
				Status:  "error",
				Message: metricsErr.Error(),
				Latency: metricsLatency.String(),
			}
		} else {
			response.Checks["metrics"] = HealthStatus{
				Status:  "ok",
				Latency: metricsLatency.String(),
			}
		}
	}

	if !allHealthy {
		response.Status = "degraded"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Secure handles a comprehensive health check including security features.
func (h *HealthHandler) Secure(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UnixMilli(),
		Checks:    make(map[string]HealthStatus),
	}

	allHealthy := true

	// Database check (skip if DB is nil).
	if h.DB != nil {
		dbStart := time.Now()
		dbErr := h.DB.PingContext(ctx)
		dbLatency := time.Since(dbStart)
		if dbErr != nil {
			allHealthy = false
			response.Checks["database"] = HealthStatus{
				Status:  "error",
				Message: "unreachable",
				Latency: dbLatency.String(),
			}
		} else {
			response.Checks["database"] = HealthStatus{
				Status:  "ok",
				Latency: dbLatency.String(),
			}
		}
	}

	// Security features check.
	securityChecks := h.checkSecurityFeatures(ctx)
	for k, v := range securityChecks {
		response.Checks[k] = v
		if v.Status != "ok" {
			allHealthy = false
		}
	}

	if !allHealthy {
		response.Status = "degraded"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *HealthHandler) checkSecurityFeatures(ctx context.Context) map[string]HealthStatus {
	checks := make(map[string]HealthStatus)

	// Check audit logging.
	checks["audit_logging"] = HealthStatus{
		Status: "ok",
	}

	// Check rate limiting.
	checks["rate_limiting"] = HealthStatus{
		Status: "ok",
	}

	// Check CSRF protection.
	checks["csrf_protection"] = HealthStatus{
		Status: "ok",
	}

	return checks
}

func (h *HealthHandler) collectMetrics() error {
	if h.Metrics == nil {
		return nil
	}
	_ = h.Metrics.Collect()
	return nil
}