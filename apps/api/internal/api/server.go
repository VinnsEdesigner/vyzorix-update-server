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

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/admin"
	authhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
	cmdhandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/command"
	devicehandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/device"
	websockethandlers "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/websocket"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/fcm"
	emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/crypto"

	"github.com/gin-gonic/gin"
)

// ServerConfig holds the server configuration.
type ServerConfig struct {
	AuthService     *appsvc.AuthService
	DeviceService   *device.Service
	ClientService   *client.Service
	CommandService  *cmdapp.Service
	RateLimiter     *middleware.RateLimiter
	AuthLimiter     *middleware.RateLimiter
	Config          config.Config
	Log             *slog.Logger
	SessionManager  *infraauth.SessionManager
	GoogleVerifier  *infraauth.GoogleTokenVerifier
	EmailService    *emailService.Service
	Hub             *hub.Hub
	FCMNotifier     fcm.Notifier
	DB              *storage.SQLite
	Lockout         *middleware.Lockout
	OperatorRepo    operator.Repository
}

// Server is the main API server.
type Server struct {
	engine        *gin.Engine
	rateLimiter   *middleware.RateLimiter
	authLimiter   *middleware.RateLimiter
	config        config.Config
	log           *slog.Logger

	// Database for health checks
	db *storage.SQLite

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

	// Admin handlers
	adminClientsHandler *admin.ClientsHandler

	// Middleware
	cookieAuth         *middleware.CookieAuth
	signatureVerifier  *middleware.SignatureVerifier
	lockout           *middleware.Lockout
	csrfProtector     *middleware.CSRFProtector
	turnstileVerifier *middleware.TurnstileVerifier
	revocationList    *infraauth.RevocationList

	// Hub for WebSocket state
	hub           *hub.Hub

	// HMAC verifier for device command verification
	hmacVerifier cryptohmac.Verifier
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

	// Create SignatureVerifier for request signing
	signingConfig := middleware.SigningConfig{
		Enabled:           cfg.Config.EnforceHMAC,
		TimestampWindow:   cfg.Config.HMACWindow,
		MaxCacheSize:      100000,
		GracePeriod:       24 * time.Hour,
		AllowUnsignedFallback: false,
	}
	signatureVerifier := middleware.NewSignatureVerifier(signingConfig, hmacSecretFn)

	// Create CookieAuth middleware
	cookieAuth := middleware.NewCookieAuth(cfg.SessionManager, cfg.AuthService)

	// Create lockout middleware for brute force protection
	lockoutConfig := middleware.LoadLockoutConfig()
	lockout := middleware.NewLockout(lockoutConfig)

	// Create CSRF protector
	csrfConfig := middleware.DefaultCSRFConfig()
	csrfProtector := middleware.NewCSRFProtector(csrfConfig)

	// Create Turnstile verifier for bot protection
	turnstileConfig := config.LoadTurnstileConfig()
	turnstileVerifier := middleware.NewTurnstileVerifier(middleware.TurnstileConfig{
		Secret:   turnstileConfig.Secret,
		Enabled:  turnstileConfig.Enabled,
		CacheTTL: 5 * time.Minute,
	})

	// Create session revocation list
	revocationConfig := middleware.LoadRevocationConfig()
	var revocationList *infraauth.RevocationList
	if revocationConfig.Enabled {
		revocationList = infraauth.DefaultRevocationList()
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
		hub:           cfg.Hub,
		hmacVerifier: hmacVerifier,
		db:           cfg.DB,

		// Create all auth handlers at once
		authHandlers: authhandlers.NewAllHandlers(&authhandlers.Dependencies{
			AuthService:    cfg.AuthService,
			SessionManager: cfg.SessionManager,
			Config:         cfg.Config,
			GoogleVerifier: cfg.GoogleVerifier,
			ClientService:  cfg.ClientService,
			EmailService:   cfg.EmailService,
			Lockout:        cfg.Lockout,
			OperatorRepo:   cfg.OperatorRepo,
		}),

		// Device handlers
		deviceRegisterHandler: devicehandlers.NewRegisterHandler(cfg.DeviceService),
		deviceStatusHandler:   devicehandlers.NewStatusHandler(cfg.DeviceService),
		deviceUpdaterHandler:  devicehandlers.NewUpdaterHandler(cfg.DeviceService),
		deviceListHandler:     devicehandlers.NewListHandler(cfg.DeviceService, cfg.Hub),

		// Command handler with FCM notifier
		commandHandler: cmdhandlers.NewExecuteHandler(cfg.CommandService, cfg.DeviceService, cfg.Hub, cfg.FCMNotifier),

		// WebSocket handler
		streamHandler: websockethandlers.NewStreamHandler(cfg.Log, cfg.Config, cfg.Hub, hmacVerifier),

		// Admin handlers
		adminClientsHandler: admin.NewClientsHandler(cfg.ClientService),
	}

	// Start Hub if available
	if cfg.Hub != nil {
		go cfg.Hub.Run(context.Background())
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Global middleware
	s.engine.Use(middleware.RequestIDMiddleware())
	s.engine.Use(middleware.Logger(s.log))
	s.engine.Use(middleware.CORSHandler(s.config.AllowedOrigins))
	if s.config.Env == "production" {
		s.engine.Use(middleware.SecurityHeaders())
	} else {
		s.engine.Use(middleware.SecurityHeadersRelaxed())
	}
	s.engine.Use(middleware.BodySizeLimit(middleware.DefaultBodySizeLimit))

	// ============================================================.
	// Static assets - serve directly from public directory.
	// This MUST be before NoRoute to prevent index.html fallback for assets.
	// ============================================================.
	s.engine.Static("/assets", filepath.Join(s.config.PublicDir, "assets"))

	// Health endpoints
	s.engine.GET("/health", s.healthHandler)
	s.engine.GET("/healthz", s.healthHandler)

	// Static endpoints
	s.engine.GET("/api/v1/version", s.versionHandler)
	s.engine.GET("/api/v1/changelog", s.changelogHandler)
	s.engine.GET("/api/v1/apk/*name", s.apkHandler)
	s.engine.GET("/bin/*name", s.binHandler)

	// SSR Proxy - if enabled, proxy HTML requests to Node.js SSR server.
	// This allows TanStack Start SSR to work with Go backend.
	ssrConfig := config.LoadSSRConfig()
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

	// Device registration (public)
	public.POST("/v1/device/register", s.deviceRegisterHandler.Handle)
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

		// Admin clients
		adminClients := r.Group("/admin/clients")
		adminClients.GET("", s.adminClientsHandler.List)
		adminClients.GET("/:clientId", s.adminClientsHandler.Get)
		adminClients.PATCH("/:clientId", s.adminClientsHandler.Update)
		adminClients.DELETE("/:clientId", s.adminClientsHandler.Delete)
		adminClients.POST("/:clientId/rotate-key", s.adminClientsHandler.RotateKey)

		// Device management (with request signing for API clients)
		deviceMgmt := r.Group("/device")
		deviceMgmt.Use(middleware.RequestSigningMiddleware(s.signatureVerifier))
		deviceMgmt.Use(s.requireHMAC())
		{
			deviceMgmt.GET("/count", s.deviceListHandler.Count)
			deviceMgmt.GET("/:id", s.deviceListHandler.GetDevice)
			deviceMgmt.PATCH("/:id/fcm-token", s.deviceUpdaterHandler.UpdateFCMToken)
			deviceMgmt.POST("/:id/command", s.commandHandler.Handle, s.requireStrictHMAC())
			deviceMgmt.GET("/:id/commands/pending", s.commandHandler.GetPending)
			deviceMgmt.DELETE("/:id", s.deviceUpdaterHandler.Delete)
			deviceMgmt.GET("/:id/stream", s.streamHandler.Handle)
		}

		// Command management
		commandMgmt := r.Group("/command")
		{
			commandMgmt.GET("/:dispatchId/status", s.commandHandler.GetStatus)
			commandMgmt.POST("/:dispatchId/retry", s.commandHandler.Retry)
			commandMgmt.DELETE("/:dispatchId", s.commandHandler.Cancel)
		}
	}

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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "bad_hmac", "message": err.Error()})
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "bad_hmac", "message": "strictHmac is enabled: " + err.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}
