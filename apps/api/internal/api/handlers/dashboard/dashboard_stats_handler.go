package dashboard

import (
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.DashboardStats
	_ openapi.ErrorResponse
)

// StatsHandler handles dashboard stats endpoints.
type StatsHandler struct {
	dashboardSvc *dashboard.Service
	logger       *slog.Logger
}

// NewStatsHandler creates a new dashboard stats handler.
func NewStatsHandler(dashboardSvc *dashboard.Service, logger *slog.Logger) *StatsHandler {
	return &StatsHandler{
		dashboardSvc: dashboardSvc,
		logger:       logger,
	}
}

// GetStats handles GET /v1/dashboard/stats.
// Returns aggregated dashboard statistics for the organization.
// @Summary      Get dashboard stats
// @Description  Returns aggregated dashboard statistics for the organization
// @Tags         dashboard
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.DashboardStats  "dashboard stats"
// @Failure      401  {object}  openapi.ErrorResponse  "operator context required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /dashboard/stats [get]
func (h *StatsHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract operator for auth check.
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Operator context required"))
		return
	}

	// Extract organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	response, err := h.dashboardSvc.GetDashboardStatsByOrganization(ctx, orgID)
	if err != nil {
		h.logger.Error("Failed to get dashboard stats", "organizationID", orgID, "error", err)
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to retrieve dashboard stats"))
		return
	}

	c.JSON(http.StatusOK, response)
}
