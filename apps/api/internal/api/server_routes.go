package api

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
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
	// Note: setupDashboardRoutes is called within setupAuthenticatedRoutes.
	s.setupMethodHandlers()
}

// idempotencyMiddleware returns the idempotency middleware if repository is configured.
// This middleware is applied globally to all /v1 tenant routes.
func (s *Server) idempotencyMiddleware() gin.HandlerFunc {
	if s.idempotencyRepo == nil {
		return func(c *gin.Context) { c.Next() }
	}

	config := middleware.IdempotencyConfig{
		Repository:   s.idempotencyRepo,
		HeaderName:   "X-Idempotency-Key",
		PathPrefixes: []string{"/v1/", "/api/v1/"},
		ExcludedPaths: []string{
			// Auth endpoints - should not be idempotent.
			"/v1/auth/login",
			"/v1/auth/register",
			"/v1/auth/csrf-token",
			"/v1/auth/refresh",
			"/v1/auth/logout",
			// Health/metrics endpoints.
			"/health",
			"/healthz",
			"/metrics",
		},
		TTL:     24 * time.Hour,
		Enabled: true,
	}
	return middleware.GinIdempotency(config)
}

// deviceIdempotencyMiddleware returns the idempotency middleware for device routes.
// These are separate from tenant routes and have their own auth mechanism (HMAC).
func (s *Server) deviceIdempotencyMiddleware() gin.HandlerFunc {
	if s.idempotencyRepo == nil {
		return func(c *gin.Context) { c.Next() }
	}

	config := middleware.IdempotencyConfig{
		Repository: s.idempotencyRepo,
		HeaderName: "X-Idempotency-Key",
		Paths: []string{
			"/v1/device/inbox",
			"/v1/device/confirm",
		},
		PathPrefixes: []string{"/v1/device/", "/v1/command/"},
		ExcludedPaths: []string{
			"/v1/device/status",
			"/v1/device/count",
		},
		TTL:     24 * time.Hour,
		Enabled: true,
	}
	return middleware.GinIdempotency(config)
}

func (s *Server) setupGlobalMiddleware() {
	s.engine.Use(s.mwFactory.Tracing())
	s.engine.Use(s.mwFactory.Logger())
	s.engine.Use(s.mwFactory.CORS())
	s.engine.Use(s.mwFactory.SecurityHeaders())
	s.engine.Use(s.mwFactory.BodySizeLimit())
	s.engine.Use(s.mwFactory.DisableTrace())
	s.engine.Use(s.mwFactory.DisableConnect())
	s.engine.Use(s.mwFactory.ErrorHandler())

	// Apply API key authentication to all routes except /health and /healthz.
	// Public paths handled internally by TenantAPIKeyAuth.

	ssrConfig := infraConfig.LoadSSRConfig()
	if ssrConfig.EnableSSR {
		s.engine.Use(s.mwFactory.SSRProxy(ssrConfig))
	} else {
		s.log.Warn("SSR disabled - serving static HTML files only")
	}
}

func (s *Server) setupStaticRoutes() {
	s.engine.Static("/assets", filepath.Join(s.config.PublicDir, "assets"))

	// /health - public, simple health check with JSON response.
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// /healthz - server access token protected (INFRASTRUCTURE-protected).
	// Requires X-Vyzorix-Token header with valid TokenSecret.
	auth := middleware.Authenticator{
		ServerAPIToken:    s.config.ServerAPIToken,
		DevelopmentBypass: s.config.Env != "production",
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
	// Also register at /v1/ without the /api prefix so requests forwarded.
	// by the proxy (which strips /api) resolve to the correct handler.
	s.engine.GET("/v1/version", s.versionHandler)
	s.engine.GET("/v1/changelog", s.changelogHandler)
	s.engine.GET("/v1/apk/*name", s.apkHandler)
	s.engine.GET("/v1/check-update", s.updaterHandler.CheckUpdate)
	s.engine.POST("/v1/download-progress", s.updaterHandler.DownloadProgress)
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

	// Public CSRF token endpoint - must be before CSRF middleware.
	// This generates a CSRF token for public endpoints (login, register).
	if s.csrfProtector != nil {
		authGroup.GET("/csrf-token", func(c *gin.Context) {
			token, err := s.csrfProtector.GetTokenForPublicEndpoint(c)
			if err != nil {
				_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to generate CSRF token"))

				return
			}
			c.JSON(http.StatusOK, gin.H{
				"csrf_token": token,
			})
		})
	}

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
	// DEPRECATED: /v1/device/register removed. Use /v1/device/inbox for new registration flow.
	public.GET("/v1/device/:imei/status", s.deviceStatusHandler.Handle)

	// Apply idempotency middleware to device routes.
	devicePublicGroup := public.Group("/v1/device")
	devicePublicGroup.Use(s.deviceIdempotencyMiddleware())

	// Public inbox endpoint - device submits registration request.
	if s.inboxHandler != nil {
		devicePublicGroup.POST("/inbox", s.inboxHandler.CreateInboxRequest)
	}
	// Public confirm endpoint - device confirms registration after receiving commandSecret.
	if s.deviceConfirmHandler != nil {
		devicePublicGroup.POST("/confirm", s.deviceConfirmHandler.Handle)
	}
}

func (s *Server) setupAuthenticatedRoutes() {
	public := s.engine.Group("")
	r := public.Group("/v1")

	// TENANT routes: Session OR API Key + Scope.
	// These are routes that operators access with their session or API keys.
	tenantGroup := r.Group("")
	tenantGroup.Use(s.authLimiter.Middleware())
	tenantGroup.Use(s.cookieAuth.Middleware())
	if s.revocationList != nil {
		tenantGroup.Use(middleware.AuthRevocationMiddleware(s.revocationList))
	}
	// Apply API key authentication middleware for TENANT paths.
	// This allows both session auth AND API key auth for tenant endpoints.
	if s.tenantAPIKeyAuth != nil {
		tenantGroup.Use(s.tenantAPIKeyAuth.Middleware())
		// Apply scope enforcement middleware - ensures API keys respect their scope.
		tenantGroup.Use(s.tenantAPIKeyAuth.ScopeEnforcement(middleware.MethodToScope))
		// Apply API key rate limiting after auth.
		if s.apiKeyRateLimiter != nil {
			tenantGroup.Use(middleware.APIKeyRateLimitMiddleware(s.apiKeyRateLimiter))
		}
	}

	// Apply idempotency middleware globally to all /v1 routes.
	// This provides automatic idempotency for all POST/PUT/PATCH requests.
	tenantGroup.Use(s.idempotencyMiddleware())

	// Verify per-session HMAC request signatures on tenant routes.
	// This runs after cookie/API-key auth so the session is available.
	tenantGroup.Use(middleware.SessionSignatureMiddleware(s.sessionSignatureVerifier))

	s.setupDashboardRoutes(tenantGroup)
	s.setupDeviceInboxRoutes(tenantGroup)
	s.setupDevicesRoutes(tenantGroup)
	s.setupCommandManagementRoutes(tenantGroup)
	s.setupTelemetryRoutes(tenantGroup)
	s.setupConnectionsRoutes(tenantGroup)
	s.setupUpdatesRoutes(tenantGroup)
	s.setupDiagnosticsRoutes(tenantGroup)
	s.setupSearchRoutes(tenantGroup)

	// API Keys management: SESSION ONLY (no API key auth for managing keys).
	// These routes manage API keys - shouldn't use API key auth.
	s.setupAPIKeysRoutes(tenantGroup)

	// Session management routes: Auth + Org required for org-scoped session management.
	s.setupSessionsRoutes(tenantGroup)

	// Client credentials routes: Auth + Org required for org-scoped API key management.
	s.setupClientCredentialsRoutes(tenantGroup)

	// Organization routes: Session required (operators, invitations, members).
	s.setupOrganizationRoutes(tenantGroup)

	// ADMIN routes: SuperAdmin session required (NO API key auth).
	// These require valid session AND SuperAdmin role.
	s.setupAdminRoutes(tenantGroup)

	// DEVICE routes: HMAC authentication only (no session/API key).
	// These are device-initiated requests that use HMAC signatures.
	// Note: /command and /fcm-token require HMAC, but /count and /:imei are operator endpoints.
	s.setupDeviceManagementRoutes(r)
}

func (s *Server) setupDashboardRoutes(r ...*gin.RouterGroup) {
	var router *gin.RouterGroup
	if len(r) > 0 {
		router = r[0]
	} else {
		router = s.engine.Group("/v1")
	}

	// Dashboard routes require organization context for multi-tenant isolation.
	dashboard := router.Group("/dashboard")
	dashboard.Use(middleware.NewOrganizationContext(nil).Middleware())
	dashboard.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())

	dashboard.GET("/devices", s.requireScope(permission.ActionDeviceRead, permission.WildcardScope(permission.ScopeDevices)), s.deviceListHandler.Handle)
	dashboard.GET("/devices/operator", s.requireScope(permission.ActionDeviceRead, permission.WildcardScope(permission.ScopeDevices)), s.deviceListHandler.ListByOperator)
	if s.commandHistoryHandler != nil {
		s.commandHistoryHandler.RegisterRoutes(dashboard, s.dashboardRateLimiter)
	}
	if s.deviceLogsHandler != nil {
		s.deviceLogsHandler.RegisterRoutes(dashboard, s.dashboardRateLimiter)
	}

	// Device events history endpoints.
	if s.deviceEventsHandler != nil {
		dashboard.GET("/device/:imei/events", s.requireScope(permission.ActionEventsRead, permission.WildcardScope(permission.ScopeDevices)), s.deviceEventsHandler.GetEvents)
		dashboard.GET("/events/recent", s.requireScope(permission.ActionEventsRead, permission.WildcardScope(permission.ScopeDevices)), s.deviceEventsHandler.GetRecentEvents)
		dashboard.GET("/events/types/:type", s.requireScope(permission.ActionEventsRead, permission.WildcardScope(permission.ScopeDevices)), s.deviceEventsHandler.GetEventsByType)
		dashboard.GET("/events/:id", s.requireScope(permission.ActionEventsRead, permission.WildcardScope(permission.ScopeDevices)), s.deviceEventsHandler.GetEventByID)
	}

	if s.deviceMetricsHandler != nil {
		dashboard.Use(s.requireScope(permission.ActionStatsRead, permission.WildcardScope(permission.ScopeOrg)))
		s.deviceMetricsHandler.RegisterMetricsRoutes(dashboard, s.dashboardRateLimiter)
	}
	if s.deviceTelemetryHandler != nil {
		dashboard.Use(s.requireScope(permission.ActionTelemetryRead, permission.WildcardScope(permission.ScopeTelemetry)))
		s.deviceTelemetryHandler.RegisterTelemetryRoutes(dashboard, s.dashboardRateLimiter)
	}
	if s.dashboardStatsHandler != nil {
		dashboard.Use(s.requireScope(permission.ActionStatsRead, permission.WildcardScope(permission.ScopeOrg)))
		s.dashboardStatsHandler.RegisterRoutes(dashboard, s.memberHandler.MembershipChecker())
	}
}

func (s *Server) setupAdminRoutes(r *gin.RouterGroup) {
	// SuperAdmin-only client management routes (scoped under organization).
	adminClients := r.Group("/admin/clients")
	adminClients.Use(middleware.NewOrganizationContext(nil).Middleware())
	adminClients.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
	adminClients.Use(middleware.RequireSuperAdmin())
	adminClients.GET("", s.adminClientsHandler.List)
	adminClients.GET("/:clientId", s.adminClientsHandler.Get)
	adminClients.PATCH("/:clientId", s.adminClientsHandler.Update)
	adminClients.DELETE("/:clientId", s.adminClientsHandler.Delete)
	adminClients.POST("/:clientId/rotate-key", s.adminClientsHandler.RotateKey)

	// Super admin API key management routes (scoped under organization).
	if s.superAdminAPIKeys != nil {
		adminAPIKeys := r.Group("/admin/api-keys")
		adminAPIKeys.Use(middleware.NewOrganizationContext(nil).Middleware())
		adminAPIKeys.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
		adminAPIKeys.Use(middleware.RequireSuperAdmin())
		s.superAdminAPIKeys.RegisterRoutes(adminAPIKeys)
	}

	// Admin diagnostic + update routes (super_admin, no org scope needed).
	adminDiag := r.Group("/admin")
	adminDiag.Use(s.cookieAuth.Middleware())
	adminDiag.Use(middleware.RequireSuperAdmin())
	{
		adminDiag.GET("/support-bundle", s.supportBundleHandler.GetBundle)
		adminDiag.GET("/updates/check", s.updateCheckerHandler.Check)
	}
}

func (s *Server) setupDeviceManagementRoutes(r *gin.RouterGroup) {
	deviceMgmt := r.Group("/device")
	deviceMgmt.Use(middleware.RequestSigningMiddleware(s.signatureVerifier))
	deviceMgmt.Use(middleware.MandatoryEncryptionMiddleware(s.encryptKeyFn))
	deviceMgmt.Use(s.requireHMAC())
	// Apply idempotency middleware to all device management routes.
	deviceMgmt.Use(s.deviceIdempotencyMiddleware())
	deviceMgmt.GET("/count", s.deviceListHandler.Count)
	deviceMgmt.GET("/:imei", s.deviceListHandler.GetDevice)
	deviceMgmt.PATCH("/:imei/fcm-token",
		middleware.ValidationMiddleware(&middleware.FCMTokenUpdateSchema{}),
		s.deviceUpdaterHandler.UpdateFCMToken,
	)
	deviceMgmt.POST("/:imei/command",
		s.requireStrictHMAC(),
		middleware.ValidationMiddleware(&middleware.CommandExecuteSchema{}),
		s.commandHandler.Handle,
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

		// Authenticated inbox routes - require org context for multi-tenant isolation.
		authenticatedInbox := deviceInbox.Group("")
		authenticatedInbox.Use(middleware.NewOrganizationContext(nil).Middleware())
		authenticatedInbox.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
		authenticatedInbox.Use(s.cookieAuth.Middleware())
		authenticatedInbox.Use(s.requireScope(permission.ActionInboxRead, permission.WildcardScope(permission.ScopeInbox)))
		authenticatedInbox.GET("/inbox", s.deviceRegRateLimiter.InboxListLimit(), s.inboxHandler.GetInbox)
		authenticatedInbox.GET("/inbox/:imei", s.deviceRegRateLimiter.InboxGetLimit(), s.inboxHandler.GetInboxEntry)
		authenticatedInbox.PATCH("/inbox/:imei", s.requireScope(permission.ActionInboxWrite, permission.WildcardScope(permission.ScopeInbox)), s.deviceRegRateLimiter.InboxGetLimit(), s.inboxHandler.UpdateInboxEntry)
		authenticatedInbox.POST("/inbox/:imei/ack", s.requireScope(permission.ActionInboxWrite, permission.WildcardScope(permission.ScopeInbox)), s.deviceRegRateLimiter.InboxAckLimit(), s.inboxHandler.AckInbox)
		authenticatedInbox.POST("/inbox/:imei/resend", s.requireScope(permission.ActionInboxWrite, permission.WildcardScope(permission.ScopeInbox)), s.deviceRegRateLimiter.InboxAckLimit(), s.inboxHandler.ResendApproval)

		// Note: POST /v1/device/inbox is already registered as public in setupDevicePublicRoutes.
		// No need to call RegisterPublicRoutes here as it would cause duplicate registration.
	}
}

func (s *Server) setupDevicesRoutes(r *gin.RouterGroup) {
	if s.devicesHandler != nil && s.deviceRegRateLimiter != nil {
		devices := r.Group("/devices")
		// All device routes require organization context for multi-tenant isolation.
		devices.Use(middleware.NewOrganizationContext(nil).Middleware())
		devices.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
		// Apply rate limiting per spec Section 11.1.
		devices.GET("", s.deviceRegRateLimiter.DevicesListLimit(), s.requireScope(permission.ActionDeviceRead, permission.WildcardScope(permission.ScopeDevices)), s.devicesHandler.GetDevices)
		devices.GET("/:imei", s.deviceRegRateLimiter.DevicesGetLimit(), s.requireScope(permission.ActionDeviceRead, permission.WildcardScope(permission.ScopeDevices)), s.devicesHandler.GetDeviceDetail)
		devices.DELETE("/:imei", s.deviceRegRateLimiter.DevicesDeleteLimit(), s.requireScope(permission.ActionDeviceDelete, permission.WildcardScope(permission.ScopeDevices)), s.devicesHandler.DeregisterDevice)

		// Device settings routes.
		if s.deviceSettingsHandler != nil {
			devices.GET("/:imei/settings", s.requireScope(permission.ActionDeviceRead, permission.WildcardScope(permission.ScopeDevices)), s.deviceSettingsHandler.GetSettings)
			devices.PATCH("/:imei/settings", s.requireScope(permission.ActionDeviceWrite, permission.WildcardScope(permission.ScopeDevices)), s.deviceSettingsHandler.UpdateSettings)
			devices.GET("/:imei/settings/thresholds", s.requireScope(permission.ActionDeviceRead, permission.WildcardScope(permission.ScopeDevices)), s.deviceSettingsHandler.GetThresholds)
			devices.PATCH("/:imei/settings/thresholds", s.requireScope(permission.ActionDeviceWrite, permission.WildcardScope(permission.ScopeDevices)), s.deviceSettingsHandler.UpdateThresholds)
		}

		// Device tag routes.
		{
			devices.GET("/:imei/tags", s.requireScope(permission.ActionDeviceRead, permission.WildcardScope(permission.ScopeDevices)), s.deviceListHandler.GetTags)
			devices.PUT("/:imei/tags", s.requireScope(permission.ActionDeviceWrite, permission.WildcardScope(permission.ScopeDevices)), s.deviceListHandler.SetTags)
			devices.POST("/:imei/tags/:tag", s.requireScope(permission.ActionDeviceWrite, permission.WildcardScope(permission.ScopeDevices)), s.deviceListHandler.AddTag)
			devices.DELETE("/:imei/tags/:tag", s.requireScope(permission.ActionDeviceWrite, permission.WildcardScope(permission.ScopeDevices)), s.deviceListHandler.RemoveTag)
		}
	}

	// Confirmation endpoint: an operator requests a single-use token to.
	// authorize a risky command before issuing it. Registered independently of.
	// the device-listing rate limiter (which may be unset on the wire path) so it.
	// is available on both server construction paths. Org-scoped + member-checked.
	// like the other tenant routes.
	if s.confirmationHandler != nil {
		confirm := r.Group("/devices")
		confirm.Use(middleware.NewOrganizationContext(nil).Middleware())
		confirm.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
		confirm.POST("/:imei/command/confirm", s.confirmationHandler.RequestConfirmation)
	}
}

func (s *Server) setupCommandManagementRoutes(r *gin.RouterGroup) {
	commandMgmt := r.Group("/command")
	// All command routes require organization context for multi-tenant isolation.
	commandMgmt.Use(middleware.NewOrganizationContext(nil).Middleware())
	commandMgmt.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
	commandMgmt.Use(middleware.RequestSigningMiddleware(s.signatureVerifier))
	commandMgmt.Use(middleware.MandatoryEncryptionMiddleware(s.encryptKeyFn))
	commandMgmt.GET("/:dispatchId/status", s.requireScope(permission.ActionCommand, permission.WildcardScope(permission.ScopeDevices)), s.commandHandler.GetStatus)
	// Note: Idempotency is handled globally by tenantGroup middleware.
	commandMgmt.POST("/:dispatchId/retry", s.requireScope(permission.ActionCommand, permission.WildcardScope(permission.ScopeDevices)), s.commandHandler.Retry)
	commandMgmt.DELETE("/:dispatchId", s.requireScope(permission.ActionCommand, permission.WildcardScope(permission.ScopeDevices)), s.commandHandler.Cancel)
}

func (s *Server) setupTelemetryRoutes(r *gin.RouterGroup) {
	telemetry := r.Group("/telemetry")
	telemetry.Use(middleware.NewOrganizationContext(nil).Middleware())
	telemetry.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
	telemetry.GET("/history", s.requireScope(permission.ActionTelemetryRead, permission.WildcardScope(permission.ScopeTelemetry)), s.telemetryHistoryHandler.Query)
	telemetry.GET("/history/export", s.requireScope(permission.ActionTelemetryRead, permission.WildcardScope(permission.ScopeTelemetry)), s.telemetryHistoryHandler.ExportJSON)
	telemetry.GET("/latest/:deviceId", s.requireScope(permission.ActionTelemetryRead, permission.WildcardScope(permission.ScopeTelemetry)), s.telemetryHistoryHandler.GetLatest)
	telemetry.GET("/stats/:deviceId", s.requireScope(permission.ActionTelemetryRead, permission.WildcardScope(permission.ScopeTelemetry)), s.telemetryHistoryHandler.GetStats)
	telemetry.DELETE("/cleanup", s.requireScope(permission.ActionTelemetryWrite, permission.WildcardScope(permission.ScopeTelemetry)), s.telemetryHistoryHandler.CleanupOld)
}

func (s *Server) setupConnectionsRoutes(r *gin.RouterGroup) {
	connections := r.Group("/connections")
	connections.Use(middleware.NewOrganizationContext(nil).Middleware())
	connections.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
	connections.GET("", s.requireScope(permission.ActionConnectionsRead, permission.WildcardScope(permission.ScopeDevices)), s.connectionStatusHandler.GetAllStatus)
	connections.GET("/metrics", s.requireScope(permission.ActionConnectionsRead, permission.WildcardScope(permission.ScopeDevices)), s.connectionStatusHandler.GetMetrics)
}

func (s *Server) setupUpdatesRoutes(r *gin.RouterGroup) {
	if s.updatesHandler != nil {
		updatesGroup := r.Group("/updates")
		// All updates routes require organization context for multi-tenant isolation.
		updatesGroup.Use(middleware.NewOrganizationContext(nil).Middleware())
		updatesGroup.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
		s.updatesHandler.RegisterRoutes(updatesGroup, s.cookieAuth)
	}
}

func (s *Server) setupDiagnosticsRoutes(r *gin.RouterGroup) {
	if s.diagnosticsInspectHandler != nil || s.diagnosticsTimelineHandler != nil {
		diagnosticsGroup := r.Group("/device")
		diagnosticsGroup.Use(middleware.NewOrganizationContext(nil).Middleware())
		diagnosticsGroup.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
		diagnosticsGroup.Use(s.requireScope(permission.ActionDiagnosticsRead, permission.WildcardScope(permission.ScopeDiagnostics)))
		diagnostics.RegisterRoutes(diagnosticsGroup, s.diagnosticsInspectHandler, s.diagnosticsTimelineHandler)
	}
}

func (s *Server) setupSearchRoutes(r *gin.RouterGroup) {
	searchGroup := r.Group("/search")
	searchGroup.Use(middleware.NewOrganizationContext(nil).Middleware())
	searchGroup.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
	searchGroup.Use(s.requireScope(permission.ActionSearchRead, permission.WildcardScope(permission.ScopeOrg)))
	{
		searchGroup.GET("", func(c *gin.Context) {
			orgID := middleware.GetOrganizationID(c)
			q := c.Query("q")
			resultType := c.Query("type")
			if s.searchService != nil {
				results, err := s.searchService.Search(c.Request.Context(), q, orgID, resultType, 20, c.Query("tag"), c.Query("sort"))
				if err != nil {
					c.JSON(500, gin.H{"error": "search failed"})
					return
				}
				c.JSON(200, gin.H{"results": results, "query": q})
			} else {
				c.JSON(200, gin.H{"results": []any{}, "query": q})
			}
		})
	}
}

func (s *Server) setupAPIKeysRoutes(r *gin.RouterGroup) {
	if s.apiKeysHandler != nil {
		authGroup := r.Group("/auth")
		// All API keys routes require organization context for multi-tenant isolation.
		authGroup.Use(middleware.NewOrganizationContext(nil).Middleware())
		authGroup.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
		authGroup.Use(s.requireScope(permission.ActionKeysManage, permission.WildcardScope(permission.ScopeOrg)))
		s.apiKeysHandler.RegisterRoutes(authGroup)
	}
}

// setupSessionsRoutes configures session management routes under /v1/auth/sessions.
// These routes require authentication (session cookie) but NOT organization context.
func (s *Server) setupSessionsRoutes(r *gin.RouterGroup) {
	sessionsGroup := r.Group("/auth/sessions")
	// Note: Auth middleware is applied at the tenantGroup level, not here.
	// Sessions are per-operator, not per-organization.
	sessionsGroup.Use(middleware.NoCache())
	{
		sessionsGroup.GET("", s.authHandlers.Sessions.ListSessions)
		sessionsGroup.GET("/concurrent", s.authHandlers.Sessions.CheckConcurrent)
		sessionsGroup.DELETE("/:id", s.authHandlers.Sessions.RevokeSession)
		sessionsGroup.DELETE("", s.authHandlers.Sessions.RevokeAllExceptCurrent)
		sessionsGroup.POST("/revoke-all", s.authHandlers.Sessions.RevokeAllDevices)
		sessionsGroup.GET("/:id", s.authHandlers.Sessions.GetSession)
	}
}

// setupClientCredentialsRoutes configures client credentials (API key) routes under /v1/auth/client-credentials.
// These routes require both authentication AND organization context for org-scoped API key management.
func (s *Server) setupClientCredentialsRoutes(r *gin.RouterGroup) {
	clientCredsGroup := r.Group("/auth/client-credentials")
	clientCredsGroup.Use(middleware.NewOrganizationContext(nil).Middleware())
	clientCredsGroup.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
	clientCredsGroup.Use(s.requireScope(permission.ActionKeysManage, permission.WildcardScope(permission.ScopeOrg)))
	clientCredsGroup.Use(middleware.NoCache())
	{
		clientCredsGroup.POST("", s.authHandlers.ClientCreds.Create)
		clientCredsGroup.GET("", s.authHandlers.ClientCreds.List)
		clientCredsGroup.GET("/:clientId", s.authHandlers.ClientCreds.Get)
		clientCredsGroup.DELETE("/:clientId", s.authHandlers.ClientCreds.Delete)
		clientCredsGroup.PATCH("/:clientId", s.authHandlers.ClientCreds.Update)
		clientCredsGroup.POST("/:clientId/rotate-secret", s.authHandlers.ClientCreds.RotateSecret)
	}
}

func (s *Server) setupOrganizationRoutes(r *gin.RouterGroup) {
	if s.organizationHandler == nil {
		return
	}

	// Organization routes (authenticated + org-scoped).
	orgs := r.Group("/organizations")
	orgs.Use(s.cookieAuth.Middleware())
	orgs.Use(middleware.NewOrganizationContext(nil).Middleware())
	orgs.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
	{
		// Org CRUD: admins manage settings, only super_admin deletes.
		orgs.POST("", s.requireScope(permission.ActionSettingsWrite, permission.WildcardScope(permission.ScopeOrg)), s.organizationHandler.Create)
		orgs.GET("", s.requireScope(permission.ActionSettingsRead, permission.WildcardScope(permission.ScopeOrg)), s.organizationHandler.List)
		orgs.GET("/:id", s.requireScope(permission.ActionSettingsRead, permission.WildcardScope(permission.ScopeOrg)), s.organizationHandler.Get)
		orgs.PATCH("/:id", s.requireScope(permission.ActionSettingsWrite, permission.WildcardScope(permission.ScopeOrg)), s.organizationHandler.Update)
		orgs.DELETE("/:id", s.requireScope(permission.ActionSettingsWrite, permission.WildcardScope(permission.ScopeOrg)), s.organizationHandler.Delete)

		// Member routes: member management (suspend/reinstate/remove/role) requires members:write.
		orgs.GET("/:id/members", s.requireScope(permission.ActionMembersRead, permission.WildcardScope(permission.ScopeOrg)), s.memberHandler.List)
		orgs.DELETE("/:id/members/:memberId", s.requireScope(permission.ActionMembersDelete, permission.WildcardScope(permission.ScopeOrg)), s.memberHandler.Remove)
		orgs.PATCH("/:id/members/:memberId", s.requireScope(permission.ActionMembersWrite, permission.WildcardScope(permission.ScopeOrg)), s.memberHandler.UpdateRole)
		orgs.POST("/:id/members/:memberId/transfer", s.requireScope(permission.ActionMembersWrite, permission.WildcardScope(permission.ScopeOrg)), s.memberHandler.TransferOwnership)
		orgs.POST("/:id/members/:memberId/suspend", s.requireScope(permission.ActionMembersWrite, permission.WildcardScope(permission.ScopeOrg)), s.memberHandler.Suspend)
		orgs.POST("/:id/members/:memberId/reinstate", s.requireScope(permission.ActionMembersWrite, permission.WildcardScope(permission.ScopeOrg)), s.memberHandler.Reinstate)

		// Invitation routes under organizations.
		orgs.GET("/:id/invitations", s.requireScope(permission.ActionMembersRead, permission.WildcardScope(permission.ScopeOrg)), s.invitationHandler.ListByOrganization)

		// Organization settings routes.
		if s.organizationSettingsHandler != nil {
			orgs.GET("/:id/settings", s.requireScope(permission.ActionSettingsRead, permission.WildcardScope(permission.ScopeOrg)), s.organizationSettingsHandler.GetSettings)
			orgs.PATCH("/:id/settings", s.requireScope(permission.ActionSettingsWrite, permission.WildcardScope(permission.ScopeOrg)), s.organizationSettingsHandler.UpdateSettings)
			orgs.GET("/:id/settings/thresholds", s.requireScope(permission.ActionSettingsRead, permission.WildcardScope(permission.ScopeOrg)), s.organizationSettingsHandler.GetThresholds)
			orgs.PATCH("/:id/settings/thresholds", s.requireScope(permission.ActionSettingsWrite, permission.WildcardScope(permission.ScopeOrg)), s.organizationSettingsHandler.UpdateThresholds)
		}

		// Device transfer route (super_admin source + target membership enforced in handler).
		if s.transferHandler != nil {
			orgs.POST("/:id/devices/:imei/transfer", s.requireScope(permission.ActionDeviceAssign, permission.WildcardScope(permission.ScopeDevices)), s.transferHandler.Transfer)
		}
	}

	// Invitation routes (public + authenticated).
	invitations := r.Group("/invitations")
	invitations.Use(s.cookieAuth.Middleware())
	{
		// Note: Idempotency is handled globally by tenantGroup middleware.
		invitations.POST("", s.invitationHandler.Create)
		invitations.GET("", s.invitationHandler.ListByInviter)
		invitations.DELETE("/:id", s.invitationHandler.Delete)
	}

	// Public invitation acceptance routes.
	invite := r.Group("/invite")
	{
		invite.GET("/:token", s.invitationHandler.GetByToken)
		// Note: Idempotency is handled globally by tenantGroup middleware.
		invite.POST("/:token/accept", s.invitationHandler.Accept)
		invite.POST("/:token/reject", s.invitationHandler.Reject)
	}

	// My invitations (pending for current user).
	me := r.Group("/me")
	me.Use(s.cookieAuth.Middleware())
	{
		me.GET("/invitations", s.invitationHandler.ListPendingForEmail)
	}

	// Operator settings routes (under /v1/me/settings for simpler UX).
	meSettings := r.Group("/me/settings")
	meSettings.Use(s.cookieAuth.Middleware())
	meSettings.Use(middleware.NoCache())
	{
		meSettings.GET("", s.authHandlers.Settings.GetSettings)
		meSettings.PATCH("", s.authHandlers.Settings.UpdateSettings)
	}

	mePrefs := r.Group("/me/preferences")
	mePrefs.Use(s.cookieAuth.Middleware())
	mePrefs.Use(middleware.NoCache())
	{
		mePrefs.GET("", s.authHandlers.Settings.GetPreferences)
		mePrefs.PATCH("", s.authHandlers.Settings.UpdatePreferences)
	}
}

func (s *Server) setupMethodHandlers() {
	s.engine.NoMethod(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/v1/") || strings.HasPrefix(c.Request.URL.Path, "/api/") {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "the requested method is not allowed for this endpoint"))

			return
		}
		s.dashboardHandler(c)
	})
	s.engine.NoRoute(func(c *gin.Context) {
		s.dashboardHandler(c)
	})
}
