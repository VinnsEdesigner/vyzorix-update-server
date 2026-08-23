// Package diagnostics provides HTTP handlers for device diagnostics.
package diagnostics

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	appdiagnostics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	domaindiagnostics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/diagnostics"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// Compile-time reference for the swaggo-annotated openapi DTO type.
var _ openapi.TimelineResult

// TimelineHandler handles device timeline HTTP requests.
type TimelineHandler struct {
	service     *appdiagnostics.Service
	rateLimiter func(c *gin.Context)
}

// NewTimelineHandler creates a new timeline handler.
func NewTimelineHandler(service *appdiagnostics.Service, rateLimiter func(c *gin.Context)) *TimelineHandler {
	return &TimelineHandler{
		service:     service,
		rateLimiter: rateLimiter,
	}
}

// GetDeviceTimeline handles GET /v1/device/:imei/timeline.
// Returns chronological event timeline for the Diagnostics Timeline.
// @Summary      Get device timeline
// @Description  Returns chronological event timeline for the Diagnostics Timeline
// @Tags         diagnostics
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei       path   string  true   "device IMEI"
// @Param        eventType  query  string  false  "filter by event type"
// @Param        startTime  query  int64   false  "epoch-millis lower bound"
// @Param        endTime    query  int64   false  "epoch-millis upper bound"
// @Param        cursor     query  string  false  "pagination cursor"
// @Param        limit      query  int     false  "result limit"
// @Success      200  {object}  openapi.TimelineResult  "timeline events"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      404  {object}  openapi.ErrorResponse  "device not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/{imei}/timeline [get]
func (h *TimelineHandler) GetDeviceTimeline(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "IMEI is required"))
		return
	}

	// Require organization context for multi-tenant isolation.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	// Get operator from context for DOA check.
	op := middleware.GetOperatorFromContext(c)
	operatorID := ""
	if op != nil {
		operatorID = op.ID
	}

	// Verify operator authorization (DOA check).
	authResp := h.service.VerifyDeviceOwnership(c.Request.Context(), imei, operatorID, orgID)
	if !authResp.Authorized {
		if authResp.Forbidden {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthzInsufficientPermissions, "Access denied - device does not belong to organization"))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Authentication required"))
		return
	}

	var req appdiagnostics.TimelineRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid query parameters"))
		return
	}

	req.IMEI = imei

	result, err := h.service.GetDeviceTimeline(c.Request.Context(), imei, &req, orgID)
	if err != nil {
		switch err {
		case domaindiagnostics.ErrDeviceNotFound:
			_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Device not found"))
		default:
			_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to get device timeline"))
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// RateLimit returns the rate limiter middleware.
func (h *TimelineHandler) RateLimit() func(c *gin.Context) {
	return h.rateLimiter
}
