package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	infraConfig "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

// setupRoutes configures all API routes on the Gin engine.
func (s *Server) setupRoutes() {
	s.engine.HandleMethodNotAllowed = true
	s.setupGlobalMiddleware()
	s.setupStaticRoutes()
	s.setupPublicRoutes()
	s.setupAuthenticatedRoutes()
	s.setupDashboardRoutes()
	s.setupMethodHandlers()
}

func (s *Server) setupGlobalMiddleware() {
	s.engine.Use(s.mwFactory.RequestID())
	s.engine.Use(s.mwFactory.Logger())
	s.engine.Use(s.mwFactory.CORS())
	s.engine.Use(s.mwFactory.SecurityHeaders())
	s.engine.Use(s.mwFactory.BodySizeLimit())
	s.engine.Use(s.mwFactory.DisableTrace())
	s.engine.Use(s.mwFactory.DisableConnect())
	s.engine.Use(s.mwFactory.ErrorHandler())

	// Apply API key authentication to all routes except /health and /healthz
	// Public paths handled internally by TenantAPIKeyAuth

	ssrConfig := infraConfig.LoadSSRConfig()
	if ssrConfig.EnableSSR {
		s.engine.Use(s.mwFactory.SSRProxy(ssrConfig))
	} else {
		s.log.Warn("SSR disabled - serving static HTML files only")
	}
}

func (s *Server) setupStaticRoutes() {
	s.engine.Static("/assets", filepath.Join(s.config.PublicDir, "assets"))

	// /health - public, simple 200 OK with no body
	s.engine.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// /healthz - server access token protected (INFRASTRUCTURE-protected)
	// Requires X-Vyzorix-Token header with valid TokenSecret
	auth := middleware.Authenticator{
		TokenSecret:       s.config.TokenSecret,
		DevelopmentBypass:  s.config.Env != "production",
	}
	s.engine.GET("/healthz", auth.Middleware(), s.healthHandler)

	if s.metricsHandler != nil {
		s.engine.GET("/metrics", s.metricsHandler.Handle)
	}
	s.engine.GET("/api/v1/version", s.versionHandler)
	s.engine.GET("/api/v1/changelog", s.changelogHandler)
	s.engine.GET("/api/v1/apk/*name", s.apkHandler)
	s.engine.GET("/bin/*name", s.binHandler)
	s.engine.GET("/api/v1/check-update", s.updaterHandler.CheckUpdate)
	s.engine.POST("/api/v1/download-progress", s.updaterHandler.DownloadProgress)
}

func (s *Server) setupPublicRoutes() {
	public := s.engine.Group("")
	public.Use(s.rateLimiter.Middleware())
	public.GET("/", s.dashboardHandler)
	s.setupAuthRoutes(public)
	s.setupDevicePublicRoutes(public)
}

func (s *Server) setupAuthRoutes(public *gin.RouterGroup) {
	authGroup := public.Group("/v1/auth")
	authGroup.Use(s.authLimiter.Middleware())
	if s.ipIntelligence != nil {
		authGroup.Use(s.ipIntelligence.Middleware())
	}
	authGroup.Use(s.mwFactory.PreventUserEnum())
	if s.lockout != nil && s.lockout.IsEnabled() {
		authGroup.Use(s.mwFactory.Lockout())
	}
	if s.csrfProtector != nil && s.csrfProtector.Config.Enabled {
		authGroup.Use(s.mwFactory.CSRF())
	}
	if s.turnstileVerifier != nil && s.turnstileVerifier.Config.Enabled {
		authGroup.Use(s.mwFactory.Turnstile())
	}
	s.authHandlers.RegisterRoutes(authGroup, s.cookieAuth)
}

func (s *Server) setupDevicePublicRoutes(public *gin.RouterGroup) {
	public.POST("/v1/device/register",
		middleware.ValidationMiddleware(&middleware.DeviceRegisterSchema{}),
		s.deviceRegisterHandler.Handle,
	)
	public.GET("/v1/device/:imei/status", s.deviceStatusHandler.Handle)
	// Public inbox endpoint - device submits registration request
	if s.inboxHandler != nil {
		public.POST("/v1/device/inbox", s.inboxHandler.CreateInboxRequest)
	}
	// Public confirm endpoint - device confirms registration after receiving commandSecret
	if s.deviceConfirmHandler != nil {
		public.POST("/v1/device/confirm", s.deviceConfirmHandler.Handle)
	}
}

func (s *Server) setupAuthenticatedRoutes() {
	public := s.engine.Group("")
	r := public.Group("/v1")

	// TENANT routes: Session OR API Key + Scope
	// These are routes that operators access with their session or API keys
	tenantGroup := r.Group("")
	tenantGroup.Use(s.authLimiter.Middleware())
	tenantGroup.Use(s.cookieAuth.Middleware())
	if s.revocationList != nil {
		tenantGroup.Use(middleware.AuthRevocationMiddleware(s.revocationList))
	}
	// Apply API key authentication middleware for TENANT paths
	// This allows both session auth AND API key auth for tenant endpoints
	if s.apiKeyAuth != nil {
		tenantGroup.Use(s.apiKeyAuth.Middleware())
		// Apply API key rate limiting after auth
		if s.apiKeyRateLimiter != nil {
			tenantGroup.Use(middleware.APIKeyRateLimitMiddleware(s.apiKeyRateLimiter))
		}
	}

	s.setupDashboardRoutes(tenantGroup)
	s.setupDeviceInboxRoutes(tenantGroup)
	s.setupDevicesRoutes(tenantGroup)
	s.setupCommandManagementRoutes(tenantGroup)
	s.setupTelemetryRoutes(tenantGroup)
	s.setupConnectionsRoutes(tenantGroup)
	s.setupUpdatesRoutes(tenantGroup)
	s.setupDiagnosticsRoutes(tenantGroup)

	// API Keys management: SESSION ONLY (no API key auth for managing keys)
	// These routes manage API keys - shouldn't use API key auth
	s.setupAPIKeysRoutes(tenantGroup)

	// ADMIN routes: SuperAdmin session required (NO API key auth)
	// These require valid session AND SuperAdmin role
	s.setupAdminRoutes(tenantGroup)

	// DEVICE routes: HMAC authentication only (no session/API key)
	// These are device-initiated requests that use HMAC signatures
	// Note: /command and /fcm-token require HMAC, but /count and /:imei are operator endpoints
	s.setupDeviceManagementRoutes(r)
}

func (s *Server) setupDashboardRoutes(r ...*gin.RouterGroup) {
	var router *gin.RouterGroup
	if len(r) > 0 {
		router = r[0]
	} else {
		router = s.engine.Group("/v1")
	}
	router.GET("/dashboard/devices", s.deviceListHandler.Handle)
	router.GET("/dashboard/devices/operator", s.deviceListHandler.ListByOperator)
	if s.commandHistoryHandler != nil {
		s.commandHistoryHandler.RegisterRoutes(router, s.dashboardRateLimiter)
	}
	if s.deviceLogsHandler != nil {
		s.deviceLogsHandler.RegisterRoutes(router, s.dashboardRateLimiter)
	}
	if s.deviceMetricsHandler != nil {
		s.deviceMetricsHandler.RegisterMetricsRoutes(router, s.dashboardRateLimiter)
	}
	if s.deviceTelemetryHandler != nil {
		s.deviceTelemetryHandler.RegisterTelemetryRoutes(router, s.dashboardRateLimiter)
	}
	if s.dashboardStatsHandler != nil {
		s.dashboardStatsHandler.RegisterRoutes(router)
	}
}

func (s *Server) setupAdminRoutes(r *gin.RouterGroup) {
	// SuperAdmin-only client management routes
	adminClients := r.Group("/admin/clients")
	adminClients.Use(middleware.RequireSuperAdmin())
	adminClients.GET("", s.adminClientsHandler.List)
	adminClients.GET("/:clientId", s.adminClientsHandler.Get)
	adminClients.PATCH("/:clientId", s.adminClientsHandler.Update)
	adminClients.DELETE("/:clientId", s.adminClientsHandler.Delete)
	adminClients.POST("/:clientId/rotate-key", s.adminClientsHandler.RotateKey)

	// Super admin API key management routes
	if s.superAdminAPIKeys != nil {
		adminAPIKeys := r.Group("/admin/api-keys")
		adminAPIKeys.Use(middleware.RequireSuperAdmin())
		s.superAdminAPIKeys.RegisterRoutes(adminAPIKeys)
	}
}

func (s *Server) setupDeviceManagementRoutes(r *gin.RouterGroup) {
	deviceMgmt := r.Group("/device")
	deviceMgmt.Use(middleware.RequestSigningMiddleware(s.signatureVerifier))
	deviceMgmt.Use(middleware.MandatoryEncryptionMiddleware(s.encryptKeyFn))
	deviceMgmt.Use(s.requireHMAC())
	deviceMgmt.GET("/count", s.deviceListHandler.Count)
	deviceMgmt.GET("/:imei", s.deviceListHandler.GetDevice)
	deviceMgmt.PATCH("/:imei/fcm-token",
		middleware.ValidationMiddleware(&middleware.FCMTokenUpdateSchema{}),
		s.deviceUpdaterHandler.UpdateFCMToken,
	)
	deviceMgmt.POST("/:imei/command",
		middleware.ValidationMiddleware(&middleware.CommandExecuteSchema{}),
		s.commandHandler.Handle,
		s.requireStrictHMAC(),
	)
	deviceMgmt.GET("/:imei/commands/pending", s.commandHandler.GetPending)
	deviceMgmt.DELETE("/:imei", s.deviceUpdaterHandler.Delete)
	deviceMgmt.GET("/:imei/stream", s.streamHandler.Handle)
	deviceMgmt.GET("/:imei/connection-status", s.connectionStatusHandler.GetStatus)
	deviceMgmt.POST("/:imei/disconnect", s.connectionStatusHandler.DisconnectDevice)
}

func (s *Server) setupDeviceInboxRoutes(r *gin.RouterGroup) {
	if s.inboxHandler != nil && s.deviceRegRateLimiter != nil {
		deviceInbox := r.Group("/device")
		// Apply rate limiting per spec Section 11.1
		deviceInbox.GET("/inbox", s.deviceRegRateLimiter.InboxListLimit(), s.inboxHandler.GetInbox)
		deviceInbox.GET("/inbox/:imei", s.deviceRegRateLimiter.InboxGetLimit(), s.inboxHandler.GetInboxEntry)
		deviceInbox.POST("/inbox/:imei/ack", s.deviceRegRateLimiter.InboxAckLimit(), s.inboxHandler.AckInbox)
		// Note: POST /v1/device/inbox is public (used by devices for registration)
		s.inboxHandler.RegisterPublicRoutes(deviceInbox)
	}
}

func (s *Server) setupDevicesRoutes(r *gin.RouterGroup) {
	if s.devicesHandler != nil && s.deviceRegRateLimiter != nil {
		devices := r.Group("/devices")
		// Apply rate limiting per spec Section 11.1
		devices.GET("", s.deviceRegRateLimiter.DevicesListLimit(), s.devicesHandler.GetDevices)
		devices.GET("/:imei", s.deviceRegRateLimiter.DevicesGetLimit(), s.devicesHandler.GetDeviceDetail)
		devices.DELETE("/:imei", s.deviceRegRateLimiter.DevicesDeleteLimit(), s.devicesHandler.DeregisterDevice)
	}
}

func (s *Server) setupCommandManagementRoutes(r *gin.RouterGroup) {
	commandMgmt := r.Group("/command")
	commandMgmt.Use(middleware.RequestSigningMiddleware(s.signatureVerifier))
	commandMgmt.Use(middleware.MandatoryEncryptionMiddleware(s.encryptKeyFn))
	commandMgmt.GET("/:dispatchId/status", s.commandHandler.GetStatus)
	commandMgmt.POST("/:dispatchId/retry", s.commandHandler.Retry)
	commandMgmt.DELETE("/:dispatchId", s.commandHandler.Cancel)
}

func (s *Server) setupTelemetryRoutes(r *gin.RouterGroup) {
	telemetry := r.Group("/telemetry")
	telemetry.GET("/history", s.telemetryHistoryHandler.Query)
	telemetry.GET("/history/export", s.telemetryHistoryHandler.ExportJSON)
	telemetry.GET("/latest/:deviceId", s.telemetryHistoryHandler.GetLatest)
	telemetry.GET("/stats/:deviceId", s.telemetryHistoryHandler.GetStats)
	telemetry.DELETE("/cleanup", s.telemetryHistoryHandler.CleanupOld)
}

func (s *Server) setupConnectionsRoutes(r *gin.RouterGroup) {
	connections := r.Group("/connections")
	connections.GET("", s.connectionStatusHandler.GetAllStatus)
	connections.GET("/metrics", s.connectionStatusHandler.GetMetrics)
}

func (s *Server) setupUpdatesRoutes(r *gin.RouterGroup) {
	if s.updatesHandler != nil {
		updatesGroup := r.Group("/updates")
		s.updatesHandler.RegisterRoutes(updatesGroup, s.cookieAuth)
	}
}

func (s *Server) setupDiagnosticsRoutes(r *gin.RouterGroup) {
	if s.diagnosticsInspectHandler != nil || s.diagnosticsTimelineHandler != nil {
		diagnosticsGroup := r.Group("/device")
		diagnostics.RegisterRoutes(diagnosticsGroup, s.diagnosticsInspectHandler, s.diagnosticsTimelineHandler)
	}
}

func (s *Server) setupAPIKeysRoutes(r *gin.RouterGroup) {
	if s.apiKeysHandler != nil {
		authGroup := r.Group("/auth")
		s.apiKeysHandler.RegisterRoutes(authGroup)
	}
}

func (s *Server) setupMethodHandlers() {
	s.engine.NoMethod(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/v1/") || strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error":   "method_not_allowed",
				"message": "the requested method is not allowed for this endpoint",
			})
			return
		}
		s.dashboardHandler(c)
	})
	s.engine.NoRoute(func(c *gin.Context) {
		s.dashboardHandler(c)
	})
}
