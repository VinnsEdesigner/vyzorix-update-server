package operator

import (
	"net/http"

	appoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	infrawebhook "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notifications, err := h.service.GetNotifications(c.Request.Context(), op.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// PatchNotifications handles PATCH /v1/auth/me/notifications.
func (h *NotificationHandler) PatchNotifications(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req operator.NotificationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	notifications, err := h.service.UpdateNotifications(c.Request.Context(), op.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// TestWebhook handles POST /v1/auth/me/notifications/webhook/test.
func (h *NotificationHandler) TestWebhook(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	// Block SSRF: reject URLs resolving to private/internal IPs.
	if err := infrawebhook.ValidateURL(req.URL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook URL: " + err.Error()})
		return
	}

	result, err := h.webhookClient.Test(c.Request.Context(), req.URL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success":      false,
			"error":       "webhook_error",
			"message":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RotateWebhookSecret handles POST /v1/auth/me/notifications/webhook/rotate.
func (h *NotificationHandler) RotateWebhookSecret(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	secret, err := h.service.RotateWebhookSecret(c.Request.Context(), op.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate webhook secret"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"secret": secret})
}
