package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	gqlmiddleware "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/subscription"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/admin"
	authhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	cmdhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/command"
	devicehandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers"
	updaterhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/updater"
	websockethandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	infraConfig "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"

	"github.com/gin-gonic/gin"
	gql "github.com/graphql-go/graphql"
)

// ServerConfig holds the server configuration.
type ServerConfig struct {
	AuthService     *appsvc.AuthService
	DeviceService   *device.Service
	ClientService   *client.Service
	CommandService  *cmdapp.Service
	RateLimiter     *middleware.RateLimiter
	AuthLimiter     *middleware.RateLimiter
	Config          infraConfig.Config
	Log             *slog.Logger
	SessionManager  *infraauth.SessionManager
	GoogleVerifier  *infraauth.GoogleTokenVerifier
	EmailService    *emailService.Service
	Hub             *hub.Hub
	FCMNotifier     fcm.Notifier
	DB              *storage.SQLite
	Lockout         *middleware.Lockout
	OperatorRepo    operator.Repository
	Metrics         *metrics.Metrics
	AuditLogger     *audit.Logger
	IPIntelligence  *middleware.IPIntelligence
}

// Server is the main API server.
type Server struct {
	engine        *gin.Engine
	rateLimiter   *middleware.RateLimiter
	authLimiter   *middleware.RateLimiter
	config        infraConfig.Config
	log           *slog.Logger

	// Database for health checks
	db *storage.SQLite

	// Metrics
	metricsHandler *metrics.MetricsHandler

	// Auth handlers
	authHandlers *authhandlers.AllHandlers

	// Device handlers
	deviceRegisterHandler *devicehandlers.RegisterHandler
	deviceStatusHandler   *devicehandlers.StatusHandler
	deviceUpdaterHandler  *devicehandlers.UpdaterHandler
	deviceListHandler     *devicehandlers.ListHandler

	// Command handler
	commandHandler *cmdhandlers.ExecuteHandler

	// WebSocket handler
	streamHandler *websockethandlers.StreamHandler

	// Telemetry history handler
	telemetryHistoryHandler *handlers.TelemetryHistoryHandler

	// Connection status handler
	connectionStatusHandler *handlers.ConnectionStatusHandler

	// Admin handlers
	adminClientsHandler *admin.ClientsHandler

	// Updater handlers
	updaterHandler *updaterhandlers.Handler

	// Middleware
	cookieAuth         *middleware.CookieAuth
	signatureVerifier  *middleware.SignatureVerifier
	lockout           *middleware.Lockout
	csrfProtector     *middleware.CSRFProtector
	turnstileVerifier *middleware.TurnstileVerifier
	revocationList    *infraauth.RevocationList
	ipIntelligence    *middleware.IPIntelligence

	// Hub for WebSocket state
	hub           *hub.Hub

	// HMAC verifier for device command verification
	hmacVerifier cryptohmac.Verifier

	// Encryption key function for response encryption
	encryptKeyFn func(clientID string) ([]byte, bool)

	// Session manager for GraphQL auth
	sessionManager *infraauth.SessionManager
}

// NewServer creates a new API server with wired-up dependencies.
func NewServer(cfg *ServerConfig) *Server {
	if cfg.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.GinPanicRecovery(cfg.Log))

	// Create HMAC verifier with proper secret function from client service
	// Wrapper to adapt context-aware method to context-unaware interface
	hmacSecretFn := func(clientID string) (string, bool) {
		return cfg.ClientService.GetHmacKey(context.Background(), clientID)
	}
	hmacVerifier := cryptohmac.Verifier{
		Secret: hmacSecretFn,
		Nonces: cryptohmac.NewNonceCache(cfg.Config.HMACWindow),
		Window: cfg.Config.HMACWindow,
	}

	// Create SignatureVerifier for request signing (ENABLED by default per PRD)
	signingConfig := middleware.LoadSigningConfig()
	// If EnforceHMAC is explicitly set in config, it overrides the default
	if !cfg.Config.EnforceHMAC && !signingConfig.Enabled {
		// Both say disabled - keep disabled
	} else if cfg.Config.EnforceHMAC {
		signingConfig.Enabled = true
	}
	signatureVerifier := middleware.NewSignatureVerifier(signingConfig, hmacSecretFn)

	// Create encryption key function for response encryption
	// Derives AES-256 key from client_secret using SHA-512
	encryptKeyFn := func(clientID string) ([]byte, bool) {
		secret, ok := cfg.ClientService.GetHmacKey(context.Background(), clientID)
		if !ok || secret == "" {
			return nil, false
		}
		// Derive 32-byte key from secret using SHA-512 (per PRD)
		return cryptohmac.DeriveKey(secret), true
	}

	// Create CookieAuth middleware
	cookieAuth := middleware.NewCookieAuth(cfg.SessionManager, cfg.AuthService)

	// Create lockout middleware for brute force protection
	lockoutConfig := middleware.LoadLockoutConfig()
	lockout := middleware.NewLockout(lockoutConfig)

	// Create CSRF protector
	csrfConfig := middleware.DefaultCSRFConfig()
	csrfProtector := middleware.NewCSRFProtector(csrfConfig)

	// Create Turnstile verifier for bot protection
	turnstileCfg := middleware.LoadTurnstileConfig()
	turnstileVerifier := middleware.NewTurnstileVerifier(turnstileCfg)

	// Create session revocation list
	revocationConfig := middleware.LoadRevocationConfig()
	var revocationList *infraauth.RevocationList
	if revocationConfig.Enabled {
		revocationList = infraauth.DefaultRevocationList()
	}

	// Create IP intelligence
	ipIntelligence := middleware.NewIPIntelligence(middleware.LoadIPIntelligenceConfig())
	if ipIntelligence != nil {
		go ipIntelligence.StartCleanupRoutine(context.Background(), 5*time.Minute)
	}

	s := &Server{
		engine:        engine,
		rateLimiter:   cfg.RateLimiter,
		authLimiter:   cfg.AuthLimiter,
		config:        cfg.Config,
		log:           cfg.Log,
		cookieAuth:    cookieAuth,
		signatureVerifier: signatureVerifier,
		lockout:       lockout,
		csrfProtector: csrfProtector,
		turnstileVerifier: turnstileVerifier,
		revocationList: revocationList,
		ipIntelligence: ipIntelligence,
		hub:           cfg.Hub,
		hmacVerifier: hmacVerifier,
		encryptKeyFn: encryptKeyFn,
		db:           cfg.DB,
		sessionManager: cfg.SessionManager, // Store for GraphQL
	}

	// Create presenter for HTTP response handling
	presenter := response.NewPresenter(cfg.AuthService, cfg.AuditLogger, cfg.IPIntelligence)

	// Create all auth handlers
	s.authHandlers = authhandlers.NewAllHandlers(&authhandlers.Dependencies{
		AuthService:    cfg.AuthService,
		SessionManager: cfg.SessionManager,
		Config:         cfg.Config,
		GoogleVerifier: cfg.GoogleVerifier,
		ClientService:  cfg.ClientService,
		EmailService:   cfg.EmailService,
		Lockout:        cfg.Lockout,
		OperatorRepo:   cfg.OperatorRepo,
		AuditLogger:    cfg.AuditLogger,
		IPIntelligence:  cfg.IPIntelligence,
		Presenter:      presenter,
	})

	// Device handlers
	s.deviceRegisterHandler = devicehandlers.NewRegisterHandler(cfg.DeviceService)
	s.deviceStatusHandler = devicehandlers.NewStatusHandler(cfg.DeviceService)
	s.deviceUpdaterHandler = devicehandlers.NewUpdaterHandler(cfg.DeviceService)
	s.deviceListHandler = devicehandlers.NewListHandler(cfg.DeviceService, cfg.Hub)

	// Command handler with FCM notifier
	s.commandHandler = cmdhandlers.NewExecuteHandler(cfg.CommandService, cfg.DeviceService, cfg.Hub, cfg.FCMNotifier)

	// WebSocket handler
	s.streamHandler = websockethandlers.NewStreamHandler(cfg.Log, cfg.Config, cfg.Hub, hmacVerifier)

	// Telemetry history handler
	s.telemetryHistoryHandler = handlers.NewTelemetryHistoryHandler(
		cfg.Log,
		storage.NewTelemetryRepository(cfg.DB.DB()),
		nil, // Use default config
	)

	// Connection status handler
	s.connectionStatusHandler = handlers.NewConnectionStatusHandler(cfg.Log, cfg.Hub)

	// Admin handlers
	s.adminClientsHandler = admin.NewClientsHandler(cfg.ClientService)

	// Updater handlers
	s.updaterHandler = updaterhandlers.NewHandler(cfg.Log, cfg.Config)

	// Start Hub if available
	if cfg.Hub != nil {
		go cfg.Hub.Run(context.Background())
	}

	// Initialize metrics handler
	if cfg.Metrics != nil {
		s.metricsHandler = metrics.NewMetricsHandler(cfg.Metrics)
		s.engine.Use(metrics.Middleware(cfg.Metrics))
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Enable method not allowed handling
	s.engine.HandleMethodNotAllowed = true

	// Global middleware
	s.engine.Use(middleware.RequestIDMiddleware())
	s.engine.Use(middleware.Logger(s.log))
	s.engine.Use(middleware.CORSHandler(s.config.AllowedOrigins))
	// Always use strict security headers (both production and development)
	s.engine.Use(middleware.SecurityHeaders())
	s.engine.Use(middleware.BodySizeLimit(middleware.DefaultBodySizeLimit))
	// Disable dangerous HTTP methods globally
	s.engine.Use(middleware.DisableTrace())
	s.engine.Use(middleware.DisableConnect())
	// Error handler must be last to catch all errors
	s.engine.Use(middleware.ErrorHandler(s.log))

	// ============================================================.
	// Static assets - serve directly from public directory.
	// This MUST be before NoRoute to prevent index.html fallback for assets.
	// ============================================================.
	s.engine.Static("/assets", filepath.Join(s.config.PublicDir, "assets"))

	// Health endpoints
	s.engine.GET("/health", s.healthHandler)
	s.engine.GET("/healthz", s.healthHandler)

	// Metrics endpoint (Prometheus format)
	if s.metricsHandler != nil {
		s.engine.GET("/metrics", s.metricsHandler.Handle)
	}

	// Static endpoints
	s.engine.GET("/api/v1/version", s.versionHandler)
	s.engine.GET("/api/v1/changelog", s.changelogHandler)
	s.engine.GET("/api/v1/apk/*name", s.apkHandler)
	s.engine.GET("/bin/*name", s.binHandler)
	s.engine.GET("/api/v1/check-update", s.updaterHandler.CheckUpdate)
	s.engine.POST("/api/v1/download-progress", s.updaterHandler.DownloadProgress)

	// SSR Proxy - if enabled, proxy HTML requests to Node.js SSR server.
	// This allows TanStack Start SSR to work with Go backend.
	ssrConfig := infraConfig.LoadSSRConfig()
	if ssrConfig.EnableSSR {
		s.engine.Use(middleware.SSRProxy(s.log, ssrConfig, s.config.PublicDir, s.config.JWTSecret))
	} else {
		s.log.Warn("SSR disabled - serving static HTML files only")
	}

	// Public endpoints with rate limiting
	public := s.engine.Group("")
	public.Use(s.rateLimiter.Middleware())

	// Root path → native HTML landing page (no React needed).
	// Explicit route so Gin doesn't need to resolve /*path wildcard for /.
	public.GET("/", s.dashboardHandler)

	// Auth routes - auth handlers register their own routes
	authGroup := public.Group("/v1/auth")
	authGroup.Use(s.authLimiter.Middleware())
	
	// Apply security middleware to auth routes
	// IP Intelligence - block known malicious IPs
	if s.ipIntelligence != nil {
		authGroup.Use(s.ipIntelligence.Middleware())
	}
	
	// Prevent user enumeration
	authGroup.Use(middleware.PreventUserEnum())
	
	// Account lockout for brute force protection
	if s.lockout != nil && s.lockout.IsEnabled() {
		authGroup.Use(middleware.LockoutMiddleware(s.lockout))
	}
	
	// CSRF protection
	if s.csrfProtector != nil && s.csrfProtector.Config.Enabled {
		authGroup.Use(s.csrfProtector.Middleware())
	}
	
	// Turnstile bot protection
	if s.turnstileVerifier != nil && s.turnstileVerifier.Config.Enabled {
		authGroup.Use(middleware.TurnstileMiddleware(s.turnstileVerifier))
	}
	
	s.authHandlers.RegisterRoutes(authGroup, s.cookieAuth)

	// Device registration (public) - with validation
	public.POST("/v1/device/register", 
		middleware.ValidationMiddleware(&middleware.DeviceRegisterSchema{}), 
		s.deviceRegisterHandler.Handle,
	)
	public.GET("/v1/device/:id/status", s.deviceStatusHandler.Handle)

	// Authenticated routes
	r := public.Group("/v1")
	r.Use(s.authLimiter.Middleware())
	r.Use(s.cookieAuth.Middleware())
	
	// Session revocation check (after auth, before handlers)
	if s.revocationList != nil {
		r.Use(middleware.AuthRevocationMiddleware(s.revocationList))
	}
	
	{
		r.GET("/dashboard/devices", s.deviceListHandler.Handle)
		r.GET("/dashboard/devices/operator", s.deviceListHandler.ListByOperator)

		// Admin clients - with validation
		adminClients := r.Group("/admin/clients")
		adminClients.GET("", s.adminClientsHandler.List)
		adminClients.GET("/:clientId", s.adminClientsHandler.Get)
		adminClients.PATCH("/:clientId", s.adminClientsHandler.Update)
		adminClients.DELETE("/:clientId", s.adminClientsHandler.Delete)
		adminClients.POST("/:clientId/rotate-key", s.adminClientsHandler.RotateKey)

		// Device management (with request signing for API clients) - with validation
		// ALL responses are MANDATORY encrypted per PRD
		deviceMgmt := r.Group("/device")
		deviceMgmt.Use(middleware.RequestSigningMiddleware(s.signatureVerifier))
		deviceMgmt.Use(middleware.MandatoryEncryptionMiddleware(s.encryptKeyFn))
		deviceMgmt.Use(s.requireHMAC())
		{
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
		}

		// Command management (requires signing + encrypted responses)
		commandMgmt := r.Group("/command")
		commandMgmt.Use(middleware.RequestSigningMiddleware(s.signatureVerifier))
		commandMgmt.Use(middleware.MandatoryEncryptionMiddleware(s.encryptKeyFn))
		{
			commandMgmt.GET("/:dispatchId/status", s.commandHandler.GetStatus)
			commandMgmt.POST("/:dispatchId/retry", s.commandHandler.Retry)
			commandMgmt.DELETE("/:dispatchId", s.commandHandler.Cancel)
		}

		// Telemetry history management
		telemetryMgmt := r.Group("/telemetry")
		{
			telemetryMgmt.GET("/history", s.telemetryHistoryHandler.Query)
			telemetryMgmt.GET("/history/export", s.telemetryHistoryHandler.ExportJSON)
			telemetryMgmt.GET("/latest/:deviceId", s.telemetryHistoryHandler.GetLatest)
			telemetryMgmt.GET("/stats/:deviceId", s.telemetryHistoryHandler.GetStats)
			telemetryMgmt.DELETE("/cleanup", s.telemetryHistoryHandler.CleanupOld)
		}

		// Connection status management
		connections := r.Group("/connections")
		{
			connections.GET("", s.connectionStatusHandler.GetAllStatus)
			connections.GET("/metrics", s.connectionStatusHandler.GetMetrics)
		}

		// Device connection status (within device group for auth consistency)
		deviceMgmt.GET("/:id/connection-status", s.connectionStatusHandler.GetStatus)
		deviceMgmt.POST("/:id/disconnect", s.connectionStatusHandler.DisconnectDevice)
	}

	// ============================================================.
	// Method Not Allowed — return 405 for matched paths with wrong methods.
	// This must be before NoRoute.
	// ============================================================.
	s.engine.NoMethod(func(c *gin.Context) {
		// Don't serve SPA for API routes - return proper 405
		if strings.HasPrefix(c.Request.URL.Path, "/v1/") || strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error":   "method_not_allowed",
				"message": "the requested method is not allowed for this endpoint",
			})
			return
		}
		// For non-API routes, fall through to dashboard (SPA handles routing)
		s.dashboardHandler(c)
	})

	// ============================================================.
	// SPA fallback — handle any non-API routes by serving the React app.
	// Use NoRoute to catch unmatched routes and serve the SPA.
	// ============================================================.
	s.engine.NoRoute(func(c *gin.Context) {
		s.dashboardHandler(c)
	})
}

// Routes returns the Gin engine for serving.
func (s *Server) Routes() http.Handler {
	return s.engine
}

// Health & static handlers.
func (s *Server) healthHandler(c *gin.Context) {
	// Check database connectivity with a simple query.
	dbOk := false
	var dbErr error
	if s.db != nil {
		if err := s.db.Ping(); err == nil {
			dbOk = true
		} else {
			dbErr = err
		}
	}

	connectedDevices := 0
	if s.hub != nil {
		connectedDevices = s.hub.ClientCount()
	}

	version := ""
	if v, err := s.readVersion(); err == nil {
		version = v
	}

	status := http.StatusOK
	if !dbOk {
		status = http.StatusServiceUnavailable
	}

	response := map[string]any{
		"ok":               dbOk,
		"database":         map[bool]string{true: "ok", false: "down"}[dbOk],
		"dbOk":             dbOk,
		"serverTime":       time.Now().UnixMilli(),
		"connectedDevices": connectedDevices,
		"version":          version,
	}
	if dbErr != nil {
		response["dbError"] = dbErr.Error()
	}

	c.JSON(status, response)
}

// readVersion reads the version from version.json file.
func (s *Server) readVersion() (string, error) {
	body, err := os.ReadFile(filepath.Join(s.config.DataDir, "version.json"))
	if err != nil {
		return "", err
	}
	var v struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &v); err != nil {
		return "", err
	}
	return v.Version, nil
}

func (s *Server) versionHandler(c *gin.Context) {
	s.serveStaticFile(c, filepath.Join(s.config.DataDir, "version.json"), "application/json; charset=utf-8")
}

func (s *Server) changelogHandler(c *gin.Context) {
	s.serveStaticFile(c, filepath.Join(s.config.DataDir, "changelog.json"), "application/json; charset=utf-8")
}

func (s *Server) apkHandler(c *gin.Context) {
	name := c.Param("name")
	s.serveDownload(c, strings.TrimPrefix(name, "/"))
}

func (s *Server) binHandler(c *gin.Context) {
	name := c.Param("name")
	s.serveDownload(c, strings.TrimPrefix(name, "/"))
}

func (s *Server) serveDownload(c *gin.Context, name string) {
	if name == "" || strings.ContainsAny(name, "/\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid filename"})
		return
	}
	c.Header("Content-Type", "application/vnd.android.package-archive")
	c.Header("Cache-Control", "no-store")
	c.File(filepath.Join(s.config.BinDir, name))
}

func (s *Server) serveStaticFile(c *gin.Context, path, ct string) {
	c.Header("Content-Type", ct)
	c.Header("Cache-Control", "no-store")
	c.File(path)
}

func (s *Server) dashboardHandler(c *gin.Context) {
	path := c.Request.URL.Path

	// / → serve the native static landing page (pure HTML, no React).
	if path == "/" {
		c.File(filepath.Join(s.config.PublicDir, "landing.html"))
		return
	}

	// All other non-API paths → serve the React SPA (index.html).
	// TanStack Router inside the SPA handles client-side routing.
	clean := strings.TrimPrefix(filepath.Clean(path), "/")
	if clean == "." || clean == "" {
		clean = "index.html"
	}
	candidate := filepath.Join(s.config.PublicDir, clean)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		c.File(candidate)
		return
	}
	c.File(filepath.Join(s.config.PublicDir, "index.html"))
}

// requireHMAC is middleware that validates HMAC signatures for device API requests.
// It checks the EnforceHMAC config flag and skips HMAC validation if disabled.
func (s *Server) requireHMAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := s.hmacVerifier.ReadAndVerifyHTTP(c.Request)
		if err != nil {
			if !s.config.EnforceHMAC {
				c.Next()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid request"})
			c.Abort()
			return
		}
		c.Set("hmac_body", body)
		c.Next()
	}
}

// requireStrictHMAC checks the operator's strictHmac setting and enforces HMAC.
// signature validation when enabled for that operator.
func (s *Server) requireStrictHMAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		op := middleware.GetOperatorFromContext(c)
		if op == nil {
			// No operator in context means JWT auth didn't run or failed.
			c.Next()
			return
		}
		// If operator has strictHmac disabled, skip HMAC validation.
		if !op.ClientSettings.StrictHmac {
			c.Next()
			return
		}
		// Operator has strictHmac enabled — validate the signature.
		_, err := s.hmacVerifier.ReadAndVerifyHTTP(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "strictHmac is enabled: HMAC signature verification failed"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RegisterGraphQL initializes and registers the GraphQL server with the API server.
func (s *Server) RegisterGraphQL(
	deviceService *device.Service,
	commandService *cmdapp.Service,
	telemetryRepo *storage.TelemetryRepository,
	wsHub *hub.Hub,
) error {
	// Get auth services from server config
	authService := s.getAuthService()
	sessionManager := s.getSessionManager()

	// Create auth middleware for GraphQL
	authMw := gqlmiddleware.NewAuthMiddleware(sessionManager, authService, s.log)

	// Create resolver
	res := resolver.NewResolver(
		deviceService,
		commandService,
		wsHub,
		telemetryRepo,
		nil, // FCM notifier
		authMw,
		s.log,
	)

	// Build schema
	gqlSchema, err := schema.BuildSchema(res)
	if err != nil {
		return err
	}

	// Create handler
	h := &gqlHandler{
		schema:         gqlSchema,
		authMiddleware: authMw,
	}

	// Register routes
	s.engine.POST("/graphql", h.Handle)
	s.engine.GET("/graphql", h.Handle)
	s.engine.GET("/playground", h.Playground)

	// Create subscription handler
	subsHandler := subscription.NewHandler(wsHub, res, authMw, s.log)
	s.engine.GET("/graphql/ws", subsHandler.HandleWebSocket)

	s.log.Info("GraphQL server registered", "path", "/graphql", "playground", "/playground", "subscriptions", "/graphql/ws")
	return nil
}

// gqlHandler is the GraphQL HTTP handler.
type gqlHandler struct {
	schema         gql.Schema
	authMiddleware *gqlmiddleware.AuthMiddleware
}

// gqlRequest represents a GraphQL request.
type gqlRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

// Handle processes GraphQL requests.
func (h *gqlHandler) Handle(c *gin.Context) {
	var req gqlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []gin.H{{"message": "invalid request body"}},
		})
		return
	}

	// Authenticate
	headers := map[string]string{
		"Cookie":        c.GetHeader("Cookie"),
		"Authorization": c.GetHeader("Authorization"),
	}

	op, err := h.authMiddleware.Authenticate(c.Request.Context(), headers)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errors": []gin.H{{"message": "authentication required"}},
		})
		return
	}

	// Add operator to context
	ctx := gqlcontext.WithOperator(c.Request.Context(), op)

	// Execute query
	result := gql.Do(gql.Params{
		Schema:         h.schema,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
		Context:        ctx,
	})

	// Convert errors
	if len(result.Errors) > 0 {
		gqlErrs := make([]map[string]interface{}, 0, len(result.Errors))
		for _, err := range result.Errors {
			ext := make(map[string]interface{})
			if err.Extensions != nil {
				if code, ok := err.Extensions["code"].(string); ok {
					ext["code"] = code
				}
			}
			gqlErrs = append(gqlErrs, map[string]interface{}{
				"message":    err.Message,
				"extensions": ext,
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"data":   result.Data,
			"errors": gqlErrs,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result.Data,
	})
}

// Playground serves the GraphQL playground.
func (h *gqlHandler) Playground(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, playgroundHTML)
}

const playgroundHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Vyzorix GraphQL Playground</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
  <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
  <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
  <div id="root">
    <style>
      body { background-color: rgb(23, 42, 58); font-family: Open Sans, sans-serif; height: 90vh; }
      #root { height: 100%; width: 100%; display: flex; align-items: center; justify-content: center; }
      .loading { font-size: 32px; font-weight: 200; color: rgba(255, 255, 255, .6); margin-left: 28px; }
      img { width: 78px; height: 78px; }
      .title { font-weight: 400; }
    </style>
    <img src='https://cdn.jsdelivr.net/npm/graphql-playground-react/build/logo.png' alt=''>
    <div class="loading">Loading <span class="title">Vyzorix GraphQL Playground</span></div>
  </div>
  <script>
    window.addEventListener('load', function (event) {
      GraphQLPlayground.init(document.getElementById('root'), {
        endpoint: '/graphql',
        settings: { 'request.credentials': 'include' },
        headers: { 'Authorization': 'Bearer YOUR_TOKEN_HERE' }
      })
    })
  </script>
</body>
</html>`

// getAuthService returns the AuthService from auth handlers.
func (s *Server) getAuthService() *appsvc.AuthService {
	if s.authHandlers != nil {
		return s.authHandlers.AuthService
	}
	return nil
}

// getSessionManager returns the SessionManager from the server.
func (s *Server) getSessionManager() *infraauth.SessionManager {
	return s.sessionManager
}
