package dashboard

import (
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
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
// Returns aggregated dashboard statistics for the organization.
func (h *StatsHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract operator for auth check
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Operator context required"})
		return
	}

	// Extract organization ID from context
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	response, err := h.dashboardSvc.GetDashboardStatsByOrganization(ctx, orgID)
	if err != nil {
		h.logger.Error("Failed to get dashboard stats", "organizationID", orgID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to retrieve dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, response)
}
