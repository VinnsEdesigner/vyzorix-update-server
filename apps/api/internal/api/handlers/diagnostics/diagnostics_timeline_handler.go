// Package diagnostics provides HTTP handlers for device diagnostics.
package diagnostics

import (
	"net/http"

	appdiagnostics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	domaindiagnostics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// TimelineHandler handles device timeline HTTP requests.
type TimelineHandler struct {
	service      *appdiagnostics.Service
	rateLimiter func(c *gin.Context)
}

// NewTimelineHandler creates a new timeline handler.
func NewTimelineHandler(service *appdiagnostics.Service, rateLimiter func(c *gin.Context)) *TimelineHandler {
	return &TimelineHandler{
		service:      service,
		rateLimiter: rateLimiter,
	}
}

// GetDeviceTimeline handles GET /v1/device/:imei/timeline.
// Returns chronological event timeline for the Diagnostics Timeline.
func (h *TimelineHandler) GetDeviceTimeline(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, appdiagnostics.ErrorResponse{
			Error:   "bad_request",
			Message: "IMEI is required",
		})
		return
	}

	// Require organization context for multi-tenant isolation
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, appdiagnostics.ErrorResponse{
			Error:   "bad_request",
			Message: "organization context required",
			Code:    "ORG_REQUIRED",
		})
		return
	}

	// Get operator from context for DOA check
	op := middleware.GetOperatorFromContext(c)
	operatorID := ""
	if op != nil {
		operatorID = op.ID
	}

	// Verify device belongs to organization (org-scoped ownership check)
	authResp := h.service.VerifyDeviceOwnership(c.Request.Context(), imei, operatorID, orgID)
	if !authResp.Authorized {
		if authResp.Forbidden {
			c.JSON(http.StatusForbidden, appdiagnostics.ErrorResponse{
				Error:   "forbidden",
				Message: "Access denied - device does not belong to organization",
				Code:    "FORBIDDEN",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, appdiagnostics.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	var req appdiagnostics.TimelineRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, appdiagnostics.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid query parameters",
		})
		return
	}

	req.IMEI = imei

	result, err := h.service.GetDeviceTimeline(c.Request.Context(), imei, &req, orgID)
	if err != nil {
		switch err {
		case domaindiagnostics.ErrDeviceNotFound:
			c.JSON(http.StatusNotFound, appdiagnostics.ErrorResponse{
				Error:   "not_found",
				Message: "Device not found",
				Code:    "DEVICE_NOT_FOUND",
			})
		default:
			c.JSON(http.StatusInternalServerError, appdiagnostics.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to get device timeline",
			})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// RateLimit returns the rate limiter middleware.
func (h *TimelineHandler) RateLimit() func(c *gin.Context) {
	return h.rateLimiter
}
