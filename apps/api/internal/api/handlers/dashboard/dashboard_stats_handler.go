package dashboard

import (
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	"github.com/gin-gonic/gin"
)

// StatsHandler handles dashboard stats endpoints.
type StatsHandler struct {
	dashboardSvc *dashboard.Service
	logger      *slog.Logger
}

// NewStatsHandler creates a new dashboard stats handler.
func NewStatsHandler(dashboardSvc *dashboard.Service, logger *slog.Logger) *StatsHandler {
	return &StatsHandler{
		dashboardSvc: dashboardSvc,
		logger:      logger,
	}
}

// GetStats handles GET /v1/dashboard/stats.
// Returns aggregated dashboard statistics.
func (h *StatsHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	response, err := h.dashboardSvc.GetDashboardStats(ctx)
	if err != nil {
		h.logger.Error("Failed to get dashboard stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to retrieve dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, response)
}
