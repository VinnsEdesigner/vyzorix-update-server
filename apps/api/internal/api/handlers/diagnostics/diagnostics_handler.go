// Package diagnostics provides HTTP handlers for device diagnostics.
package diagnostics

import (
	"net/http"

	appdiagnostics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	domaindiagnostics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/diagnostics"
	"github.com/gin-gonic/gin"
)

// Handler handles diagnostics-related HTTP requests.
type Handler struct {
	service *appdiagnostics.Service
}

// NewHandler creates a new diagnostics handler.
func NewHandler(service *appdiagnostics.Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the diagnostics routes.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/:imei/inspect", h.GetDeviceInspection)
	router.GET("/:imei/timeline", h.GetDeviceTimeline)
}

// GetDeviceInspection handles GET /v1/device/:imei/inspect.
// Returns full device inspection data for the Diagnostics Inspector.
func (h *Handler) GetDeviceInspection(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, appdiagnostics.ErrorResponse{
			Error:   "bad_request",
			Message: "IMEI is required",
		})
		return
	}

	inspection, err := h.service.GetDeviceInspection(c.Request.Context(), imei)
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

// GetDeviceTimeline handles GET /v1/device/:imei/timeline.
// Returns chronological event timeline for the Diagnostics Timeline.
func (h *Handler) GetDeviceTimeline(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, appdiagnostics.ErrorResponse{
			Error:   "bad_request",
			Message: "IMEI is required",
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

	result, err := h.service.GetDeviceTimeline(c.Request.Context(), imei, &req)
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
