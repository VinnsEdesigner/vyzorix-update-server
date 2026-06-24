package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	infraConfig "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

// setupRoutes configures all API routes on the Gin engine.
func (s *Server) setupRoutes() {
	s.engine.HandleMethodNotAllowed = true

	// Global middleware
	s.engine.Use(s.mwFactory.RequestID())
	s.engine.Use(s.mwFactory.Logger())
	s.engine.Use(s.mwFactory.CORS())
	s.engine.Use(s.mwFactory.SecurityHeaders())
	s.engine.Use(s.mwFactory.BodySizeLimit())
	s.engine.Use(s.mwFactory.DisableTrace())
	s.engine.Use(s.mwFactory.DisableConnect())
	s.engine.Use(s.mwFactory.ErrorHandler())

	// Static assets
	s.engine.Static("/assets", filepath.Join(s.config.PublicDir, "assets"))

	// Health & static endpoints
	s.engine.GET("/health", s.healthHandler)
	s.engine.GET("/healthz", s.healthHandler)

	if s.metricsHandler != nil {
		s.engine.GET("/metrics", s.metricsHandler.Handle)
	}

	s.engine.GET("/api/v1/version", s.versionHandler)
	s.engine.GET("/api/v1/changelog", s.changelogHandler)
	s.engine.GET("/api/v1/apk/*name", s.apkHandler)
	s.engine.GET("/bin/*name", s.binHandler)
	s.engine.GET("/api/v1/check-update", s.updaterHandler.CheckUpdate)
	s.engine.POST("/api/v1/download-progress", s.updaterHandler.DownloadProgress)

	// SSR Proxy
	ssrConfig := infraConfig.LoadSSRConfig()
	if ssrConfig.EnableSSR {
		s.engine.Use(s.mwFactory.SSRProxy(ssrConfig))
	} else {
		s.log.Warn("SSR disabled - serving static HTML files only")
	}

	// Public endpoints
	public := s.engine.Group("")
	public.Use(s.rateLimiter.Middleware())
	public.GET("/", s.dashboardHandler)

	// Auth routes
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

	// Device registration (public)
	public.POST("/v1/device/register",
		middleware.ValidationMiddleware(&middleware.DeviceRegisterSchema{}),
		s.deviceRegisterHandler.Handle,
	)
	public.GET("/v1/device/:id/status", s.deviceStatusHandler.Handle)

	// Authenticated routes
	r := public.Group("/v1")
	r.Use(s.authLimiter.Middleware())
	r.Use(s.cookieAuth.Middleware())

	if s.revocationList != nil {
		r.Use(middleware.AuthRevocationMiddleware(s.revocationList))
	}

	// Dashboard
	r.GET("/dashboard/devices", s.deviceListHandler.Handle)
	r.GET("/dashboard/devices/operator", s.deviceListHandler.ListByOperator)

	// Admin clients
	adminClients := r.Group("/admin/clients")
	adminClients.GET("", s.adminClientsHandler.List)
	adminClients.GET("/:clientId", s.adminClientsHandler.Get)
	adminClients.PATCH("/:clientId", s.adminClientsHandler.Update)
	adminClients.DELETE("/:clientId", s.adminClientsHandler.Delete)
	adminClients.POST("/:clientId/rotate-key", s.adminClientsHandler.RotateKey)

	// Device management
	deviceMgmt := r.Group("/device")
	deviceMgmt.Use(middleware.RequestSigningMiddleware(s.signatureVerifier))
	deviceMgmt.Use(middleware.MandatoryEncryptionMiddleware(s.encryptKeyFn))
	deviceMgmt.Use(s.requireHMAC())
	deviceMgmt.GET("/count", s.deviceListHandler.Count)
	deviceMgmt.GET("/:id", s.deviceListHandler.GetDevice)
	deviceMgmt.PATCH("/:id/fcm-token",
		middleware.ValidationMiddleware(&middleware.FCMTokenUpdateSchema{}),
		s.deviceUpdaterHandler.UpdateFCMToken,
	)
	deviceMgmt.POST("/:id/command",
		middleware.ValidationMiddleware(&middleware.CommandExecuteSchema{}),
		s.commandHandler.Handle,
		s.requireStrictHMAC(),
	)
	deviceMgmt.GET("/:id/commands/pending", s.commandHandler.GetPending)
	deviceMgmt.DELETE("/:id", s.deviceUpdaterHandler.Delete)
	deviceMgmt.GET("/:id/stream", s.streamHandler.Handle)
	deviceMgmt.GET("/:id/connection-status", s.connectionStatusHandler.GetStatus)
	deviceMgmt.POST("/:id/disconnect", s.connectionStatusHandler.DisconnectDevice)

	// Command management
	commandMgmt := r.Group("/command")
	commandMgmt.Use(middleware.RequestSigningMiddleware(s.signatureVerifier))
	commandMgmt.Use(middleware.MandatoryEncryptionMiddleware(s.encryptKeyFn))
	commandMgmt.GET("/:dispatchId/status", s.commandHandler.GetStatus)
	commandMgmt.POST("/:dispatchId/retry", s.commandHandler.Retry)
	commandMgmt.DELETE("/:dispatchId", s.commandHandler.Cancel)

	// Telemetry
	telemetry := r.Group("/telemetry")
	telemetry.GET("/history", s.telemetryHistoryHandler.Query)
	telemetry.GET("/history/export", s.telemetryHistoryHandler.ExportJSON)
	telemetry.GET("/latest/:deviceId", s.telemetryHistoryHandler.GetLatest)
	telemetry.GET("/stats/:deviceId", s.telemetryHistoryHandler.GetStats)
	telemetry.DELETE("/cleanup", s.telemetryHistoryHandler.CleanupOld)

	// Connections
	connections := r.Group("/connections")
	connections.GET("", s.connectionStatusHandler.GetAllStatus)
	connections.GET("/metrics", s.connectionStatusHandler.GetMetrics)

	// Method Not Allowed handler
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

	// SPA fallback
	s.engine.NoRoute(func(c *gin.Context) {
		s.dashboardHandler(c)
	})
}
