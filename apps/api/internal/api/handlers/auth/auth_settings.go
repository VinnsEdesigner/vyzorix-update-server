package auth

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"

	"github.com/gin-gonic/gin"
)

// SettingsAuditEvent represents an audit event for settings changes.
type SettingsAuditEvent struct {
	Changes   map[string]interface{} `json:"changes,omitempty"`
	Action    string                 `json:"action"`
	Section   string                 `json:"section"`
	IPAddress string                 `json:"ipAddress"`
	UserAgent string                 `json:"userAgent"`
	Error     string                 `json:"error,omitempty"`
	Success   bool                   `json:"success"`
}

// SettingsHandler handles /me settings endpoints.
type SettingsHandler struct {
	authService     *auth.AuthService
	operatorRepo    operator.Repository
	settingsService *auth.ClientSettingsService
	presenter       *response.Presenter
	rateLimiter     *middleware.SettingsRateLimiterMiddleware
	auditLogger     *audit.Logger
}

func NewSettingsHandler(
	authService *auth.AuthService,
	operatorRepo operator.Repository,
	presenter *response.Presenter,
	rateLimiter *middleware.SettingsRateLimiterMiddleware,
) *SettingsHandler {
	if rateLimiter == nil {
		rateLimiter = middleware.NewSettingsRateLimiterMiddleware(nil)
	}
	return &SettingsHandler{
		authService:     authService,
		operatorRepo:    operatorRepo,
		settingsService: auth.NewClientSettingsService(operatorRepo),
		presenter:       presenter,
		rateLimiter:     rateLimiter,
	}
}

// SetAuditLogger sets the audit logger for settings operations.
func (h *SettingsHandler) SetAuditLogger(logger *audit.Logger) {
	h.auditLogger = logger
}

// RateLimitMiddleware returns rate limit middleware for settings endpoints.
func (h *SettingsHandler) RateLimitMiddleware() gin.HandlersChain {
	return gin.HandlersChain{
		h.rateLimiter.SettingsGetLimit(),
	}
}

// logSettingsAudit logs an audit event for settings operations.
func (h *SettingsHandler) logSettingsAudit(c *gin.Context, operatorID string, event *SettingsAuditEvent) {
	if h.auditLogger == nil {
		return
	}

	// Extract client info from gin context.
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	ctx := c.Request.Context()

	// Build metadata from changes.
	metadata := make(map[string]string)
	if event.Changes != nil {
		for k, v := range event.Changes {
			if strVal, ok := v.(string); ok {
				metadata[k] = strVal
			}
		}
	}
	metadata["section"] = event.Section
	metadata["action"] = event.Action

	if event.Action == "webhook_test" {
		if url, ok := event.Changes["url"].(string); ok {
			var responseTime int64
			if rt, ok := event.Changes["responseTime"].(int64); ok {
				responseTime = rt
			}
			h.auditLogger.WebhookTest(ctx, operatorID, url, ipAddress, userAgent, event.Success, responseTime)
			return
		}
	}

	if event.Action == "webhook_rotate" {
		h.auditLogger.WebhookSecretRotated(ctx, operatorID, ipAddress, userAgent)
		return
	}

	h.auditLogger.SettingsChanged(ctx, operatorID, event.Section, event.Action, ipAddress, userAgent, metadata)
}

// getOperatorFromSession extracts operator ID from the authenticated session.
// The cookieAuth middleware already validated the session and decrypted the.
// cookie, populating the operator in the gin context. Re-reading the raw.
// (encrypted) vyz_session cookie and passing it to ValidateSession is wrong —.
// ValidateSession expects the decrypted session ID, not the cookie ciphertext.
// We reuse the context value the middleware set instead of re-validating.
func (h *SettingsHandler) getOperatorFromSession(c *gin.Context) (string, error) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		return "", application.ErrUnauthorized
	}

	return op.ID, nil
}

// GetSettings handles GET /v1/auth/me/settings.
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get all operator settings.
	settings, err := h.operatorRepo.GetOperatorSettings(ctx, operatorID)
	if err != nil {
		h.presenter.InternalError(c, "failed to get settings")
		return
	}

	h.presenter.OK(c, settings)
}

// UpdateName handles PATCH /v1/auth/me.
func (h *SettingsHandler) UpdateName(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "not authenticated")
			return
		}

		h.presenter.InternalError(c, "an error occurred")

		return
	}

	var req struct {
		Name *string `json:"name"`
	}

	if err = c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid JSON body")
		return
	}

	if req.Name == nil || *req.Name == "" {
		h.presenter.BadRequest(c, "name is required")
		return
	}

	op, err := h.authService.UpdateOperatorName(c.Request.Context(), operatorID, *req.Name)
	if err != nil {
		h.presenter.InternalError(c, "update failed")
		return
	}

	h.presenter.OK(c, gin.H{
		"id":             op.ID,
		"email":          op.Email,
		"name":           op.Name,
		"role":           h.getOperatorRole(c, op),
		"mfa_enabled":    op.MFAEnabled,
		"email_verified": op.EmailVerified,
		"client":         op.ClientSettings,
	})
}

// getOperatorRole returns the operator's role in the current organization context.
func (h *SettingsHandler) getOperatorRole(c *gin.Context, op *operator.Operator) string {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		return ""
	}
	if m := op.GetMembership(orgID); m != nil {
		return string(m.Role)
	}
	return ""
}

// UpdateSettings handles PATCH /v1/auth/me/settings.
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "not authenticated")
			return
		}

		h.presenter.InternalError(c, "an error occurred")

		return
	}

	var req auth.UpdateSettingsRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid JSON body")
		return
	}

	if req.Reset {
		if _, err = h.authService.ResetSettings(c.Request.Context(), operatorID); err != nil {
			h.presenter.InternalError(c, "update failed")
			return
		}
	} else {
		if _, err = h.authService.UpdateSettings(c.Request.Context(), operatorID, &req); err != nil {
			h.presenter.InternalError(c, "update failed")
			return
		}
	}

	// Return the full OperatorSettings (same shape as GET /me/settings) so the.
	// client receives thresholds + notifications alongside client settings.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	settings, err := h.operatorRepo.GetOperatorSettings(ctx, operatorID)
	if err != nil {
		h.presenter.InternalError(c, "failed to get settings")
		return
	}

	h.presenter.OK(c, settings)
}

// GetThresholds handles GET /v1/auth/me/thresholds.
func (h *SettingsHandler) GetThresholds(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	thresholds, err := h.operatorRepo.GetThresholds(ctx, operatorID)
	if err != nil {
		h.presenter.InternalError(c, "failed to get thresholds")
		return
	}

	h.presenter.OK(c, thresholds)
}

// ThresholdUpdateRequest represents thresholds update request.
type ThresholdUpdateRequest struct {
	RiskWarn    *int `json:"riskWarn,omitempty"`
	RiskCrit    *int `json:"riskCrit,omitempty"`
	ThermalWarn *int `json:"thermalWarn,omitempty"`
	ThermalCrit *int `json:"thermalCrit,omitempty"`
	BufferWarn  *int `json:"bufferWarn,omitempty"`
	BufferCrit  *int `json:"bufferCrit,omitempty"`
}

// UpdateThresholds handles PATCH /v1/auth/me/thresholds.
func (h *SettingsHandler) UpdateThresholds(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	var req ThresholdUpdateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid JSON body")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	thresholds, err := h.operatorRepo.GetThresholds(ctx, operatorID)
	if err != nil {
		h.presenter.InternalError(c, "failed to get current thresholds")
		return
	}

	if req.RiskWarn != nil {
		thresholds.RiskWarn = *req.RiskWarn
	}
	if req.RiskCrit != nil {
		thresholds.RiskCrit = *req.RiskCrit
	}
	if req.ThermalWarn != nil {
		thresholds.ThermalWarn = *req.ThermalWarn
	}
	if req.ThermalCrit != nil {
		thresholds.ThermalCrit = *req.ThermalCrit
	}
	if req.BufferWarn != nil {
		thresholds.BufferWarn = *req.BufferWarn
	}
	if req.BufferCrit != nil {
		thresholds.BufferCrit = *req.BufferCrit
	}

	if err = h.operatorRepo.UpdateThresholds(ctx, operatorID, thresholds); err != nil {
		h.logSettingsAudit(c, operatorID, &SettingsAuditEvent{
			Action:  "update",
			Section: "thresholds",
			Changes: map[string]interface{}{
				"riskWarn":    thresholds.RiskWarn,
				"riskCrit":    thresholds.RiskCrit,
				"thermalWarn": thresholds.ThermalWarn,
				"thermalCrit": thresholds.ThermalCrit,
				"bufferWarn":  thresholds.BufferWarn,
				"bufferCrit":  thresholds.BufferCrit,
			},
			Success: false,
			Error:   err.Error(),
		})
		h.presenter.InternalError(c, "failed to update thresholds")
		return
	}

	h.logSettingsAudit(c, operatorID, &SettingsAuditEvent{
		Action:  "update",
		Section: "thresholds",
		Changes: map[string]interface{}{
			"riskWarn":    thresholds.RiskWarn,
			"riskCrit":    thresholds.RiskCrit,
			"thermalWarn": thresholds.ThermalWarn,
			"thermalCrit": thresholds.ThermalCrit,
			"bufferWarn":  thresholds.BufferWarn,
			"bufferCrit":  thresholds.BufferCrit,
		},
		Success: true,
	})

	h.presenter.OK(c, thresholds)
}

// GetNotifications handles GET /v1/auth/me/notifications.
func (h *SettingsHandler) GetNotifications(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	notifications, err := h.operatorRepo.GetNotifications(ctx, operatorID)
	if err != nil {
		h.presenter.InternalError(c, "failed to get notifications")
		return
	}

	h.presenter.OK(c, notifications)
}

// NotificationUpdateRequest represents notification settings update request.
type NotificationUpdateRequest struct {
	Enabled  *bool                          `json:"enabled,omitempty"`
	Channels *[]string                      `json:"channels,omitempty"`
	Email    *operator.EmailNotifications   `json:"email,omitempty"`
	Push     *operator.PushNotifications    `json:"push,omitempty"`
	Webhook  *operator.WebhookNotifications `json:"webhook,omitempty"`
}

// UpdateNotifications handles PATCH /v1/auth/me/notifications.
func (h *SettingsHandler) UpdateNotifications(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	var req NotificationUpdateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid JSON body")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get current notifications.
	notifications, err := h.operatorRepo.GetNotifications(ctx, operatorID)
	if err != nil {
		h.presenter.InternalError(c, "failed to get current notifications")
		return
	}

	// Apply updates.
	if req.Enabled != nil {
		notifications.Enabled = *req.Enabled
	}
	if req.Channels != nil {
		notifications.Channels = *req.Channels
	}
	if req.Email != nil {
		notifications.Email = *req.Email
	}
	if req.Push != nil {
		notifications.Push = *req.Push
	}
	if req.Webhook != nil {
		notifications.Webhook = *req.Webhook
	}

	// Validate webhook URL if enabled.
	if notifications.Webhook.Enabled && notifications.Webhook.URL == "" {
		h.presenter.BadRequest(c, "webhook URL is required when webhook is enabled")
		return
	}

	// Track changes for audit.
	changes := map[string]interface{}{
		"enabled":  notifications.Enabled,
		"channels": notifications.Channels,
	}

	// Save updated notifications.
	if err = h.operatorRepo.UpdateNotifications(ctx, operatorID, notifications); err != nil {
		h.logSettingsAudit(c, operatorID, &SettingsAuditEvent{
			Action:  "update",
			Section: "notifications",
			Changes: changes,
			Success: false,
			Error:   err.Error(),
		})
		h.presenter.InternalError(c, "failed to update notifications")
		return
	}

	// Log successful update.
	h.logSettingsAudit(c, operatorID, &SettingsAuditEvent{
		Action:  "update",
		Section: "notifications",
		Changes: changes,
		Success: true,
	})

	h.presenter.OK(c, notifications)
}

// WebhookTestRequest represents webhook test request.
type WebhookTestRequest struct {
	URL string `json:"url" binding:"required"`
}

// TestWebhook handles POST /v1/auth/me/notifications/webhook/test.
func (h *SettingsHandler) TestWebhook(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	var req WebhookTestRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid JSON body")
		return
	}

	// Validate URL.
	if req.URL == "" {
		h.presenter.BadRequest(c, "url is required")
		return
	}

	// Validate URL format and block SSRF (private/internal IPs).
	if validateErr := webhook.ValidateURL(req.URL); validateErr != nil {
		h.presenter.BadRequest(c, "invalid webhook URL: "+validateErr.Error())
		return
	}

	// Test webhook with a ping.
	start := time.Now()
	client := &http.Client{Timeout: 10 * time.Second}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	reqCtx, cancelReq := context.WithTimeout(ctx, 10*time.Second)
	defer cancelReq()

	httpReq, reqErr := http.NewRequestWithContext(reqCtx, http.MethodPost, req.URL, bytes.NewBufferString(`{"type":"test","message":"Vyzorix webhook test"}`))
	if reqErr != nil {
		h.presenter.BadRequest(c, "failed to create request")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Vyzorix-WebhookTest/1.0")

	resp, err := client.Do(httpReq)
	responseTime := time.Since(start).Milliseconds()

	// Log webhook test attempt.
	h.logSettingsAudit(c, operatorID, &SettingsAuditEvent{
		Action:  "webhook_test",
		Section: "notifications",
		Changes: map[string]interface{}{
			"url":          req.URL,
			"responseTime": responseTime,
		},
		Success: err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300,
	})

	if err != nil {
		h.presenter.OK(c, gin.H{
			"success":      false,
			"error":        "webhook_timeout",
			"message":      "Webhook did not respond within 10 seconds",
			"responseTime": responseTime,
		})
		return
	}

	// Close body and check status.
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	if resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		h.presenter.OK(c, gin.H{
			"success":      true,
			"statusCode":   resp.StatusCode,
			"responseTime": responseTime,
		})
	} else {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		h.presenter.OK(c, gin.H{
			"success":      false,
			"error":        "webhook_failed",
			"statusCode":   statusCode,
			"responseTime": responseTime,
			"message":      "Webhook returned non-2xx status",
		})
	}
}

// RotateWebhookSecret handles POST /v1/auth/me/notifications/webhook/rotate.
func (h *SettingsHandler) RotateWebhookSecret(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	secret, err := h.operatorRepo.RotateWebhookSecret(ctx, operatorID)
	if err != nil {
		h.logSettingsAudit(c, operatorID, &SettingsAuditEvent{
			Action:  "webhook_rotate",
			Section: "notifications",
			Success: false,
			Error:   err.Error(),
		})
		h.presenter.InternalError(c, "failed to rotate webhook secret")
		return
	}

	// Log successful rotation.
	h.logSettingsAudit(c, operatorID, &SettingsAuditEvent{
		Action:  "webhook_rotate",
		Section: "notifications",
		Success: true,
	})

	h.presenter.OK(c, gin.H{
		"secret": secret,
	})
}

func (h *SettingsHandler) GetPreferences(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}
	prefs, err := h.settingsService.GetPreferences(c.Request.Context(), op.ID)
	if err != nil {
		h.presenter.InternalError(c, "failed to get preferences")
		return
	}
	h.presenter.OK(c, gin.H{"preferences": prefs})
}

func (h *SettingsHandler) UpdatePreferences(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.presenter.BadRequest(c, "invalid JSON body")
		return
	}
	prefs, err := h.settingsService.UpdatePreferences(c.Request.Context(), op.ID, updates)
	if err != nil {
		h.presenter.InternalError(c, "failed to update preferences")
		return
	}
	h.presenter.OK(c, gin.H{"preferences": prefs})
}
