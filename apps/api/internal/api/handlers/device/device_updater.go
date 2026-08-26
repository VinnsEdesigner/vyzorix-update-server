package device

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.DeviceFCMTokenRequest
	_ openapi.SuccessResult
	_ openapi.ErrorResponse
)

// UpdaterHandler handles device update endpoints.
type UpdaterHandler struct {
	deviceService *device.Service
}

// NewUpdaterHandler creates a new UpdaterHandler.
func NewUpdaterHandler(deviceService *device.Service) *UpdaterHandler {
	return &UpdaterHandler{deviceService: deviceService}
}

// UpdateFCMToken handles PATCH /v1/device/:imei/fcm-token.
// @Summary      Update FCM token
// @Description  Updates a device's FCM push notification token
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        imei  path  string  true  "device IMEI"
// @Param        body  body  openapi.DeviceFCMTokenRequest  true  "FCM token"
// @Success      200  {object}  openapi.SuccessResult  "token updated"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid IMEI / body"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/{imei}/fcm-token [patch]
func (h *UpdaterHandler) UpdateFCMToken(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "device ID is required"))
		return
	}

	var req struct {
		FCMToken string `json:"fcmToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid request"))
		return
	}

	err := h.deviceService.UpdateFCMToken(c.Request.Context(), imei, req.FCMToken)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to update FCM token"))
		return
	}

	c.JSON(http.StatusOK, schema.SuccessResult{Success: true})
}

// Delete handles DELETE /v1/device/:imei.
// @Summary      Delete device
// @Description  Deletes a device
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        imei  path  string  true  "device IMEI"
// @Success      200  {object}  openapi.SuccessResult  "device deleted"
// @Failure      400  {object}  openapi.ErrorResponse  "device ID required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/{imei} [delete]
func (h *UpdaterHandler) Delete(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "device ID is required"))
		return
	}

	err := h.deviceService.Delete(c.Request.Context(), imei)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to delete device"))
		return
	}

	c.JSON(http.StatusOK, schema.SuccessResult{Success: true})
}
