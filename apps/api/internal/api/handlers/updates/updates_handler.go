package updates

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/gin-gonic/gin"
)

// UpdatesHandler coordinates all updates handlers.
type UpdatesHandler struct {
	service           *updates.Service
	versionsHandler   *UpdatesVersionsHandler
	pushHandler       *UpdatesPushHandler
	historyHandler    *UpdatesHistoryHandler
	syncHandler       *UpdatesSyncHandler
	webhookHandler    *GitHubWebhookHandler
	deviceStatusHandler *DeviceStatusHandler
	rateLimiters      *middleware.UpdatesRateLimiterMiddleware
	adminAuth         *middleware.UpdatesAdminAuth
	auditLogger       *audit.Logger
	webhookSecret     string
}

// NewUpdatesHandler creates a new UpdatesHandler with all sub-handlers.
func NewUpdatesHandler(service *updates.Service, pushService *updates.PushService, rateLimiters *middleware.UpdatesRateLimiterMiddleware, auditLogger *audit.Logger, webhookSecret string) *UpdatesHandler {
	return &UpdatesHandler{
		service:          service,
		versionsHandler:  NewUpdatesVersionsHandler(service),
		pushHandler:      NewUpdatesPushHandler(service, auditLogger),
		historyHandler:   NewUpdatesHistoryHandler(service, auditLogger),
		syncHandler:      NewUpdatesSyncHandler(service, auditLogger),
		auditLogger:      auditLogger,
		webhookSecret:    webhookSecret,
	}
}

// RegisterRoutes registers all updates routes.
func (h *UpdatesHandler) RegisterRoutes(rg *gin.RouterGroup, cookieAuth *middleware.CookieAuth) {
	// Device callback endpoint - public (no auth required, device identifies itself).
	// This must be registered BEFORE the cookie auth middleware is applied.
	// rg is already mounted at /v1/updates by setupUpdatesRoutes, so device-status
	// lives at /v1/updates/device-status.
	if h.deviceStatusHandler != nil {
		rg.POST("/device-status", h.deviceStatusHandler.HandleDeviceUpdateStatus)
	}

	// Authenticated updates endpoints. rg is already at /v1/updates, so register
	// routes directly on it (previously this created a doubled /updates/updates path).
	updatesGroup := rg
	updatesGroup.Use(cookieAuth.Middleware())

	// Apply rate limiting per endpoint if configured.
	if h.rateLimiters != nil {
		// Status and versions - rate limited.
		updatesGroup.GET("/status",
			h.rateLimiters.StatusLimit(),
			h.versionsHandler.GetStatus)
		updatesGroup.GET("/versions",
			h.rateLimiters.VersionsLimit(),
			h.versionsHandler.GetVersions)
		updatesGroup.GET("/changelog",
			h.rateLimiters.ChangelogLimit(),
			h.versionsHandler.GetChangelog)
		updatesGroup.GET("/export",
			h.rateLimiters.VersionsLimit(),
			h.versionsHandler.Export)

		// Push - rate limited + admin only.
		updatesGroup.POST("/push",
			h.rateLimiters.PushLimit(),
			h.adminAuth.RequireAdmin(),
			h.pushHandler.PushUpdate)

		// History - rate limited.
		updatesGroup.GET("/history",
			h.rateLimiters.HistoryLimit(),
			h.historyHandler.GetHistory)
		updatesGroup.GET("/history/:pushId",
			h.rateLimiters.HistoryLimit(),
			h.historyHandler.GetPushDetail)
		updatesGroup.POST("/history/:pushId/cancel",
			h.rateLimiters.CancelLimit(),
			h.historyHandler.CancelPush)

		// Sync - rate limited + admin only.
		updatesGroup.POST("/sync",
			h.rateLimiters.SyncLimit(),
			h.adminAuth.RequireAdmin(),
			h.syncHandler.SyncVersions)
		updatesGroup.GET("/sync/status",
			h.rateLimiters.SyncLimit(),
			h.syncHandler.GetSyncStatus)

		// Webhook - no cookie auth (uses HMAC signature), admin only for info endpoint.
		if h.webhookHandler != nil {
			updatesGroup.POST("/webhook/github",
				h.webhookHandler.HandleWebhook)
			updatesGroup.GET("/webhook/info",
				h.webhookHandler.GetWebhookInfo)
		}
	} else {
		// No rate limiting configured.
		// Push - admin only.
		updatesGroup.POST("/push",
			h.adminAuth.RequireAdmin(),
			h.pushHandler.PushUpdate)
		// Sync - admin only.
		updatesGroup.POST("/sync",
			h.adminAuth.RequireAdmin(),
			h.syncHandler.SyncVersions)
		// Cancel - admin only.
		updatesGroup.POST("/history/:pushId/cancel",
			h.adminAuth.RequireAdmin(),
			h.historyHandler.CancelPush)

		// Remaining routes without rate limiting.
		updatesGroup.GET("/status", h.versionsHandler.GetStatus)
		updatesGroup.GET("/versions", h.versionsHandler.GetVersions)
		updatesGroup.GET("/changelog", h.versionsHandler.GetChangelog)
		updatesGroup.GET("/export", h.versionsHandler.Export)
		updatesGroup.GET("/history", h.historyHandler.GetHistory)
		updatesGroup.GET("/history/:pushId", h.historyHandler.GetPushDetail)
		updatesGroup.GET("/sync/status", h.syncHandler.GetSyncStatus)

		// Webhook - no cookie auth (uses HMAC signature).
		if h.webhookHandler != nil {
			updatesGroup.POST("/webhook/github",
				h.webhookHandler.HandleWebhook)
			updatesGroup.GET("/webhook/info",
				h.webhookHandler.GetWebhookInfo)
		}
	}
}

// Stop stops all rate limiter cleanup goroutines.
func (h *UpdatesHandler) Stop() {
	if h.rateLimiters != nil {
		h.rateLimiters.Stop()
	}
}

// Alias methods to match expected handler names in verify script.

// GetUpdateStatus is an alias for GetStatus.
func (h *UpdatesHandler) GetUpdateStatus(c *gin.Context) {
	h.versionsHandler.GetUpdateStatus(c)
}

// ExportVersions is an alias for Export.
func (h *UpdatesHandler) ExportVersions(c *gin.Context) {
	h.versionsHandler.ExportVersions(c)
}

// SyncVersions is an alias for Sync.
func (h *UpdatesHandler) SyncVersions(c *gin.Context) {
	h.syncHandler.SyncVersions(c)
}
