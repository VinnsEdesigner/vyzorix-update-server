package device

import (
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"

	"github.com/gin-gonic/gin"
)

// TransferHandler handles device transfer requests.
type TransferHandler struct {
	deviceService *device.Service
	presenter     *response.Presenter
}

// NewTransferHandler creates a new TransferHandler.
func NewTransferHandler(deviceService *device.Service, presenter *response.Presenter) *TransferHandler {
	return &TransferHandler{
		deviceService: deviceService,
		presenter:     presenter,
	}
}

// TransferRequest represents a device transfer request.
type TransferRequest struct {
	TargetOrgID string `json:"targetOrgId" binding:"required"`
	Notes       string `json:"notes"`
}

// Transfer handles POST /v1/organizations/:id/devices/:imei/transfer.
func (h *TransferHandler) Transfer(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	sourceOrgID := c.Param("id")
	if sourceOrgID == "" {
		h.presenter.BadRequest(c, "source organization ID is required")
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		h.presenter.BadRequest(c, "device IMEI is required")
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid request body")
		return
	}

	if req.TargetOrgID == "" {
		h.presenter.BadRequest(c, "target organization ID is required")
		return
	}

	if req.TargetOrgID == sourceOrgID {
		h.presenter.BadRequest(c, "cannot transfer to the same organization")
		return
	}

	err := h.deviceService.TransferDevice(c.Request.Context(), imei, sourceOrgID, req.TargetOrgID, op.ID)
	if err != nil {
		if errors.Is(err, application.ErrDeviceNotFound) {
			h.presenter.NotFound(c, "device not found in this organization")
			return
		}
		if errors.Is(err, application.ErrDeviceOnline) {
			h.presenter.Forbidden(c, "device must be offline to transfer")
			return
		}
		if errors.Is(err, application.ErrForbidden) {
			h.presenter.Forbidden(c, "access denied")
			return
		}
		h.presenter.InternalError(c, "failed to transfer device")
		return
	}

	h.presenter.OK(c, gin.H{
		"message":       "device transferred successfully",
		"device_imei":   imei,
		"source_org":    sourceOrgID,
		"target_org":    req.TargetOrgID,
	})
}

// RegisterRoutes registers the transfer routes.
func (h *TransferHandler) RegisterRoutes(r *gin.RouterGroup, orgMiddleware, membershipMiddleware gin.HandlerFunc) {
	devices := r.Group("/:id/devices")
	devices.Use(orgMiddleware)
	devices.Use(membershipMiddleware)
	{
		devices.POST("/:imei/transfer", h.Transfer)
	}
}
