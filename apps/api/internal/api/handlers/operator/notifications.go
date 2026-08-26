package operator

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	appoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/operator"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	infrawebhook "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/schema"
)

// NotificationHandler handles notification endpoints.
type NotificationHandler struct {
	service       *appoperator.NotificationService
	webhookClient *infrawebhook.Client
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(svc *appoperator.NotificationService, wh *infrawebhook.Client) *NotificationHandler {
	return &NotificationHandler{
		service:       svc,
		webhookClient: wh,
	}
}

// GetNotifications handles GET /v1/auth/me/notifications.
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "unauthorized"))
		return
	}

	notifications, err := h.service.GetNotifications(c.Request.Context(), op.ID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to get notifications"))
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// PatchNotifications handles PATCH /v1/auth/me/notifications.
func (h *NotificationHandler) PatchNotifications(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "unauthorized"))
		return
	}

	var req operator.NotificationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request"))
		return
	}

	notifications, err := h.service.UpdateNotifications(c.Request.Context(), op.ID, &req)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to update notifications"))
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// TestWebhook handles POST /v1/auth/me/notifications/webhook/test.
func (h *NotificationHandler) TestWebhook(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "unauthorized"))
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "url is required"))
		return
	}

	// Block SSRF: reject URLs resolving to private/internal IPs.
	if err := infrawebhook.ValidateURL(req.URL); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid webhook URL: "+err.Error()))
		return
	}

	result, err := h.webhookClient.Test(c.Request.Context(), req.URL)
	if err != nil {
		_ = c.Error(apperrors.NewServerErrorFromStatus(http.StatusOK,

			err.Error()))

		return
	}

	c.JSON(http.StatusOK, result)
}

// RotateWebhookSecret handles POST /v1/auth/me/notifications/webhook/rotate.
func (h *NotificationHandler) RotateWebhookSecret(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "unauthorized"))
		return
	}

	secret, err := h.service.RotateWebhookSecret(c.Request.Context(), op.ID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to rotate webhook secret"))
		return
	}

	c.JSON(http.StatusOK, schema.WebhookSecretResult{Secret: secret})
}
