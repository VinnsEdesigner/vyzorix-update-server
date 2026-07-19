// Package diagnostics provides HTTP handlers for device diagnostics.
package diagnostics

import (
	"net/http"

	appdiagnostics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	domaindiagnostics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// Handler combines all diagnostics handlers for backwards compatibility.
type Handler struct {
	InspectHandler   *InspectHandler
	TimelineHandler *TimelineHandler
}

// InspectHandler handles device inspection HTTP requests.
type InspectHandler struct {
	service       *appdiagnostics.Service
	rateLimiter  func(c *gin.Context)
}

// NewInspectHandler creates a new inspect handler.
func NewInspectHandler(service *appdiagnostics.Service, rateLimiter func(c *gin.Context)) *InspectHandler {
	return &InspectHandler{
		service:      service,
		rateLimiter: rateLimiter,
	}
}

// GetDeviceInspection handles GET /v1/device/:imei/inspect.
// Returns full device inspection data for the Diagnostics Inspector.
func (h *InspectHandler) GetDeviceInspection(c *gin.Context) {
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

	// Verify operator authorization (DOA check)
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

	inspection, err := h.service.GetDeviceInspectionHTTP(c.Request.Context(), imei, orgID)
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
				Message: "Failed to get device inspection",
			})
		}
		return
	}

	c.JSON(http.StatusOK, inspection)
}

// RateLimit returns the rate limiter middleware.
func (h *InspectHandler) RateLimit() func(c *gin.Context) {
	return h.rateLimiter
}
