package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health handles GET /health.
func (h *HealthHandler) Health(c *gin.Context) {
	// Check database connectivity for a thorough health check
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbHealthy := true
	dbStatus := "connected"
	if err := h.db.PingContext(ctx); err != nil {
		dbHealthy = false
		dbStatus = "disconnected"
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if !dbHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": gin.H{
			"database": dbStatus,
		},
	})
}

// Ready handles GET /ready.
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check database connectivity.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready":  false,
			"reason": "database unreachable",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ready": true,
	})
}
