package device

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/apierrors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.DeviceTransferRequest
	_ openapi.DeviceTransferResult
	_ openapi.ErrorResponse
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
// @Summary      Transfer device
// @Description  Transfers a device to another organization. Requires super_admin in source org and membership in target org
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id    path  string  true  "source organization ID"
// @Param        imei  path  string  true  "device IMEI"
// @Param        body  body  openapi.DeviceTransferRequest  true  "target organization"
// @Success      200  {object}  openapi.DeviceTransferResult  "transfer result"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input / same org"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied / not member of target"
// @Failure      404  {object}  openapi.ErrorResponse  "device not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/devices/{imei}/transfer [post]
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

	// Authorization: only a super_admin of the source org may move its devices,
	// and only into an org the operator is an active member of. The route group
	// only enforces cookie auth, so the check must happen here.
	if !op.IsSuperAdminIn(sourceOrgID) {
		h.presenter.Forbidden(c, "super_admin role required in the source organization")
		return
	}
	if m := op.GetMembership(req.TargetOrgID); m == nil || !m.IsActive() {
		h.presenter.Forbidden(c, "not a member of the target organization")
		return
	}

	err := h.deviceService.TransferDevice(c.Request.Context(), imei, sourceOrgID, req.TargetOrgID, op.ID)
	if err != nil {
		// The domain→HTTP classification (device not found → 404, device online →
		// 400, forbidden → 403, anything else → 500) lives in one place now.
		apierrors.RespondError(c, err)
		return
	}

	h.presenter.OK(c, openapi.DeviceTransferResult{
		Success:    true,
		Message:    "device transferred successfully",
		DeviceIMEI: imei,
		SourceOrg:  sourceOrgID,
		TargetOrg:  req.TargetOrgID,
	})
}
