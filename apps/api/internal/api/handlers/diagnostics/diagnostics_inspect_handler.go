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

// Compile-time references for swaggo-annotated openapi DTO types.
var _ openapi.DeviceInspectionResult

// Handler combines all diagnostics handlers for backwards compatibility.
type Handler struct {
	InspectHandler  *InspectHandler
	TimelineHandler *TimelineHandler
}

// InspectHandler handles device inspection HTTP requests.
type InspectHandler struct {
	service     *appdiagnostics.Service
	rateLimiter func(c *gin.Context)
}

// NewInspectHandler creates a new inspect handler.
func NewInspectHandler(service *appdiagnostics.Service, rateLimiter func(c *gin.Context)) *InspectHandler {
	return &InspectHandler{
		service:     service,
		rateLimiter: rateLimiter,
	}
}

// GetDeviceInspection handles GET /v1/device/:imei/inspect.
// @Summary      Get device inspection
// @Description  Returns full device inspection data for the Diagnostics Inspector
// @Tags         diagnostics
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei  path  string  true  "device IMEI"
// @Success      200  {object}  openapi.DeviceInspectionResult  "device inspection"
// @Failure      400  {object}  openapi.ErrorResponse  "IMEI required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      404  {object}  openapi.ErrorResponse  "device not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/{imei}/inspect [get]
func (h *InspectHandler) GetDeviceInspection(c *gin.Context) {
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

	inspection, err := h.service.GetDeviceInspectionHTTP(c.Request.Context(), imei, orgID)
	if err != nil {
		switch err {
		case domaindiagnostics.ErrDeviceNotFound:
			_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Device not found"))
		default:
			_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to get device inspection"))
		}
		return
	}

	c.JSON(http.StatusOK, inspection)
}

// RateLimit returns the rate limiter middleware.
func (h *InspectHandler) RateLimit() func(c *gin.Context) {
	return h.rateLimiter
}
