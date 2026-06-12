// Package controllers provides HTTP handlers.
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	security "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/fcm"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
	hmac "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/crypto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
)

type Server struct {
	Notifier        fcm.Notifier
	Log             *slog.Logger
	Store           *storage.Store
	Hub             *hub.Hub
	Limiter         *middleware.RateLimiter
	AuthLimiter     *middleware.RateLimiter
	jwtCtrl         *AuthController
	originValidator *security.OriginValidator
	upgrader        websocket.Upgrader
	HMAC            hmac.Verifier
	Config          config.Config
}

func New(log *slog.Logger, cfg config.Config, st *storage.Store, h *hub.Hub, notifier fcm.Notifier) *Server {
	s := &Server{Log: log, Config: cfg, Store: st, Hub: h, Notifier: notifier}
	s.HMAC = hmac.Verifier{
		Window: cfg.HMACWindow,
		Nonces: hmac.NewNonceCache(cfg.HMACWindow),
		Secret: func(id string) (string, bool) { return st.Secret(context.Background(), id) },
	}
	s.Limiter = middleware.NewRateLimiter(100, time.Minute)
	// Stricter rate limiter for auth endpoints: 5 requests per minute to prevent brute force
	s.AuthLimiter = middleware.NewRateLimiter(5, time.Minute)
	s.jwtCtrl = NewAuthController(log, cfg, st)

	// Initialize origin validator
	s.originValidator = security.NewOriginValidator(cfg.AllowedOrigins)
	s.originValidator.SetLogger(log)

	// Initialize WebSocket upgrader with proper origin checking
	s.upgrader = websocket.Upgrader{
		CheckOrigin:      s.originValidator.CheckOrigin(),
		HandshakeTimeout: 10 * time.Second,
	}

	return s
}

func (s *Server) Engine() *gin.Engine {
	if s.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.Logger(s.Log))
	r.Use(middleware.CORSHandler(s.Config.AllowedOrigins))

	// Security: limit request body size to prevent large payload attacks
	r.Use(middleware.BodySizeLimit(middleware.DefaultBodySizeLimit))

	// Multipart form limit (8MB for APK uploads)
	r.MaxMultipartMemory = middleware.LargeBodySizeLimit

	// Rate limit public endpoints to prevent abuse
	public := r.Group("")
	public.Use(s.Limiter.Middleware())

	// Auth routes (no JWT required for login/register; JWT required for /me and logout)
	auth := public.Group("/v1/auth")
	auth.GET("/google", s.jwtCtrl.GoogleLoginRedirect)     // triggers OAuth redirect
	auth.GET("/google/callback", s.jwtCtrl.GoogleCallback) // OAuth callback from Google
	// Email verification and password reset (no JWT required)
	auth.POST("/verify-email", s.jwtCtrl.VerifyEmail)
	auth.POST("/resend-verification", s.jwtCtrl.ResendVerification)
	auth.POST("/reset-password", s.jwtCtrl.ResetPassword)
	// /me and /logout require JWT — middleware applied inline
	auth.GET("/me", JWTAuth(s.jwtCtrl.jwt, s.Store), s.jwtCtrl.Me)
	auth.PATCH("/me", JWTAuth(s.jwtCtrl.jwt, s.Store), s.jwtCtrl.UpdateName)
	auth.PATCH("/me/settings", JWTAuth(s.jwtCtrl.jwt, s.Store), s.jwtCtrl.UpdateSettings)
	auth.POST("/logout", JWTAuth(s.jwtCtrl.jwt, s.Store), s.jwtCtrl.Logout)

	// Stricter rate limiting for sensitive auth endpoints (5 req/min to prevent brute force)
	// Applied inline - both the general Limiter (100/min) AND AuthLimiter (5/min)
	auth.POST("/login", s.AuthLimiter.Middleware(), s.jwtCtrl.Login)
	auth.POST("/register", s.AuthLimiter.Middleware(), s.jwtCtrl.Register)
	auth.POST("/forgot-password", s.AuthLimiter.Middleware(), s.jwtCtrl.ForgotPassword)

	public.GET("/health", s.health)
	public.GET("/healthz", s.health)
	public.GET("/api/v1/version", s.version)
	public.GET("/api/v1/changelog", s.changelog)
	public.GET("/api/v1/apk/*name", s.apk)
	public.GET("/bin/*name", s.bin)

	// Static assets - serve directly from public directory
	// This MUST be before NoRoute to prevent index.html fallback for assets
	r.Static("/assets", filepath.Join(s.Config.PublicDir, "assets"))

	// Root path → native HTML landing page (no React needed)
	// Explicit route so Gin doesn't need to resolve /*path wildcard for /
	public.GET("/", s.dashboard)

	// Device routes - rate limited for public endpoints
	public.POST("/v1/device/register", s.register)
	public.GET("/v1/device/:id/status", s.status)
	r.PATCH("/v1/device/:id/fcm-token", s.requireHMAC(), s.fcmToken)
	r.POST("/v1/device/:id/command", JWTAuth(s.jwtCtrl.jwt, s.Store), s.requireStrictHMAC(), s.command)
	r.DELETE("/v1/device/:id", s.requireHMAC(), s.deleteDevice)

	// Dashboard routes — protected by JWT
	r.GET("/v1/dashboard/devices", JWTAuth(s.jwtCtrl.jwt, s.Store), s.dashboardDevices)

	// WebSocket — HMAC or JWT
	r.GET("/v1/device/:id/stream", s.stream)

	// SSR Proxy - if enabled, proxy HTML requests to Node.js SSR server
	// This allows TanStack Start SSR to work with Go backend
	ssrConfig := config.LoadSSRConfig()
	if ssrConfig.EnableSSR {
		r.Use(middleware.SSRProxy(s.Log, ssrConfig, s.Config.PublicDir, s.Config.JWTSecret))
	} else {
		// Fallback: serve static HTML files (no SSR)
		s.Log.Warn("SSR disabled - serving static HTML files only")

		// SPA fallback — handle any non-API routes by serving the React app
		// Use NoRoute to catch unmatched routes and serve the SPA
		r.NoRoute(func(c *gin.Context) {
			s.dashboard(c)
		})
	}

	return r
}

func (s *Server) Routes() http.Handler { return s.Engine() }

func (s *Server) health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// Check database connectivity with a simple query
	dbOk := false
	var dbErr error
	if err := s.Store.Ping(ctx); err == nil {
		// Additional check: verify we can execute a query
		if _, err := s.Store.Devices(ctx); err == nil {
			dbOk = true
		} else {
			dbErr = err
		}
	} else {
		dbErr = err
	}

	connectedDevices := 0
	if s.Hub != nil {
		connectedDevices = s.Hub.ClientCount()
	}

	version := ""
	if v, err := s.readVersion(); err == nil {
		version = v
	}

	status := 200
	if !dbOk {
		status = 503
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

func (s *Server) readVersion() (string, error) {
	body, err := os.ReadFile(filepath.Join(s.Config.DataDir, "version.json"))
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

func (s *Server) version(c *gin.Context) {
	serveStaticFile(c, filepath.Join(s.Config.DataDir, "version.json"), "application/json; charset=utf-8")
}

func (s *Server) changelog(c *gin.Context) {
	serveStaticFile(c, filepath.Join(s.Config.DataDir, "changelog.json"), "application/json; charset=utf-8")
}

func (s *Server) apk(c *gin.Context) {
	name := c.Param("name")
	s.serveDownload(c, strings.TrimPrefix(name, "/"))
}

func (s *Server) bin(c *gin.Context) {
	name := c.Param("name")
	s.serveDownload(c, strings.TrimPrefix(name, "/"))
}

func (s *Server) serveDownload(c *gin.Context, name string) {
	if name == "" || strings.ContainsAny(name, "/\\") {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "invalid filename"})
		return
	}
	c.Header("Content-Type", "application/vnd.android.package-archive")
	c.Header("Cache-Control", "no-store")
	c.File(filepath.Join(s.Config.BinDir, name))
}

func serveStaticFile(c *gin.Context, path, ct string) {
	c.Header("Content-Type", ct)
	c.Header("Cache-Control", "no-store")
	c.File(path)
}

func (s *Server) register(c *gin.Context) {
	var req models.RegisterRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, map[string]string{"error": "bad_json", "message": err.Error()})
		return
	}
	if req.DeviceID == "" || req.FirebaseInstallID == "" {
		c.JSON(400, map[string]string{"error": "missing_field", "message": "deviceId and firebaseInstallId are required"})
		return
	}
	d, _, err := s.Store.Register(c.Request.Context(), req)
	if errors.Is(err, storage.ErrHijack) {
		c.JSON(409, map[string]string{"error": "device_hijack", "message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(500, map[string]string{"error": "register_failed", "message": err.Error()})
		return
	}
	c.JSON(200, models.RegisterResponse{DeviceID: d.ID, CommandSecret: d.CommandSecret, RegisteredAt: d.RegisteredAt.UnixMilli(), ServerTime: time.Now().UnixMilli()})
}

func (s *Server) status(c *gin.Context) {
	id := c.Param("id")
	d, ok, err := s.Store.Device(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, map[string]string{"error": "lookup_failed", "message": err.Error()})
		return
	}
	if !ok {
		c.JSON(404, map[string]string{"error": "unknown_device", "message": id})
		return
	}
	c.JSON(200, models.DeviceStatus{DeviceID: d.ID, Online: s.Hub.Online(id) || d.Online, LastSeen: d.LastSeen.UnixMilli(), AppVersion: d.AppVersion, DeviceClass: d.DeviceClass})
}

func (s *Server) fcmToken(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		FCMToken string `json:"fcmToken"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, map[string]string{"error": "bad_json", "message": err.Error()})
		return
	}
	if req.FCMToken == "" {
		c.JSON(400, map[string]string{"error": "missing_field", "message": "fcmToken is required"})
		return
	}
	if err := s.Store.UpdateFCM(c.Request.Context(), id, req.FCMToken); err != nil {
		c.JSON(500, map[string]string{"error": "update_failed", "message": err.Error()})
		return
	}
	c.JSON(200, map[string]any{"deviceId": id, "serverTime": time.Now().UnixMilli()})
}

func (s *Server) deleteDevice(c *gin.Context) {
	id := c.Param("id")
	if err := s.Store.DeleteDevice(c.Request.Context(), id); err != nil {
		c.JSON(500, map[string]string{"error": "delete_failed", "message": err.Error()})
		return
	}
	c.JSON(200, map[string]any{"deviceId": id, "deleted": true})
}

func (s *Server) command(c *gin.Context) {
	id := c.Param("id")
	var req models.CommandRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, map[string]string{"error": "bad_json", "message": err.Error()})
		return
	}
	if req.Command == "" {
		c.JSON(400, map[string]string{"error": "missing_field", "message": "command is required"})
		return
	}
	_, found, err := s.Store.Device(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, map[string]string{"error": "lookup_failed", "message": err.Error()})
		return
	}
	if !found {
		c.JSON(404, map[string]string{"error": "unknown_device", "message": id})
		return
	}
	frame := models.CommandFrame{Type: "command", DispatchID: storage.NewDispatchID(), Command: req.Command, Args: req.Args, Nonce: req.Nonce, Timestamp: req.Timestamp, Signature: req.Signature}
	delivery := "queued"
	if s.Hub.Send(id, frame) {
		delivery = "sent"
	}
	if err := s.Store.SaveCommand(c.Request.Context(), frame.DispatchID, id, req.Command, req.Args, delivery); err != nil {
		s.Log.Warn("save command failed", "dispatchId", frame.DispatchID, "err", err)
	}
	if delivery == "queued" {
		s.sendFCMWakeIfNeeded(id, req.Command, frame.DispatchID)
	}
	c.JSON(202, models.CommandResponse{DispatchID: frame.DispatchID, Delivery: delivery, ServerTime: time.Now().UnixMilli()})
}

// sendFCMWakeIfNeeded sends an FCM wake notification if notifier is configured.
func (s *Server) sendFCMWakeIfNeeded(deviceID, command, dispatchID string) {
	if s.Notifier == nil {
		return
	}
	d, found, err := s.Store.Device(context.Background(), deviceID)
	if err != nil || !found || d.FCMToken == "" {
		return
	}
	if err := s.Notifier.SendSilentWake(context.Background(), fcm.SilentWake{Token: d.FCMToken, Command: command, DispatchID: dispatchID, DeviceID: deviceID}); err != nil {
		s.Log.Warn("fcm wake failed", "deviceId", deviceID, "dispatchId", dispatchID, "err", err)
		if markErr := s.Store.MarkWake(context.Background(), dispatchID, err.Error()); markErr != nil {
			s.Log.Warn("mark wake failed", "dispatchId", dispatchID, "err", markErr)
		}
		return
	}
	if markErr := s.Store.MarkWake(context.Background(), dispatchID, ""); markErr != nil {
		s.Log.Warn("mark wake failed", "dispatchId", dispatchID, "err", markErr)
	}
}

func (s *Server) stream(c *gin.Context) {
	if s.Config.EnforceHMAC {
		if err := s.HMAC.Verify(c.Request.Method, c.Request.URL.RequestURI(), c.Param("id"), nil, c.Request.Header); err != nil {
			c.JSON(401, map[string]string{"error": "bad_hmac", "message": err.Error()})
			return
		}
	}
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.Log.Warn("websocket upgrade failed", "err", err)
		return
	}
	id := c.Param("id")
	wsClient := &hub.Client{DeviceID: id, Conn: conn, Send: make(chan models.CommandFrame, 32), Hub: s.Hub}
	s.Hub.Register(wsClient)
	go wsClient.WritePump()
	wsClient.ReadPump()
}

func (s *Server) dashboard(c *gin.Context) {
	path := c.Request.URL.Path

	// / → serve the native static landing page (pure HTML, no React)
	if path == "/" {
		c.File(filepath.Join(s.Config.PublicDir, "landing.html"))
		return
	}

	// All other non-API paths → serve the React SPA (index.html)
	// TanStack Router inside the SPA handles client-side routing
	clean := strings.TrimPrefix(filepath.Clean(path), "/")
	if clean == "." || clean == "" {
		clean = "index.html"
	}
	candidate := filepath.Join(s.Config.PublicDir, clean)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		c.File(candidate)
		return
	}
	c.File(filepath.Join(s.Config.PublicDir, "index.html"))
}

func (s *Server) dashboardDevices(c *gin.Context) {
	ds, err := s.Store.Devices(c.Request.Context())
	if err != nil {
		c.JSON(500, map[string]string{"error": "list_failed", "message": err.Error()})
		return
	}
	type row struct {
		DeviceID    string `json:"deviceId"`
		AppVersion  string `json:"appVersion"`
		DeviceClass string `json:"deviceClass"`
		LastSeen    int64  `json:"lastSeen"`
		Online      bool   `json:"online"`
	}
	out := make([]row, 0, len(ds))
	for _, d := range ds {
		out = append(out, row{
			DeviceID:    d.ID,
			Online:      s.Hub.Online(d.ID) || d.Online,
			LastSeen:    d.LastSeen.UnixMilli(),
			AppVersion:  d.AppVersion,
			DeviceClass: d.DeviceClass,
		})
	}
	c.JSON(200, map[string]any{"devices": out})
}

//nolint:unused
func (s *Server) _dashboardAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.Config.TokenSecret == "" && s.Config.Env != "production" {
			c.Next()
			return
		}
		auth := c.GetHeader("Authorization")
		token := c.GetHeader("X-Vyzorix-Token")
		if auth == "Bearer "+s.Config.TokenSecret || token == s.Config.TokenSecret {
			c.Next()
			return
		}
		c.JSON(401, map[string]string{"error": "unauthorized", "message": "missing or invalid dashboard token"})
		c.Abort()
	}
}

func (s *Server) requireHMAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := s.HMAC.ReadAndVerifyHTTP(c.Request)
		if err != nil {
			if !s.Config.EnforceHMAC && errors.Is(err, hmac.ErrMissing) {
				c.Next()
				return
			}
			c.JSON(401, map[string]string{"error": "bad_hmac", "message": err.Error()})
			c.Abort()
			return
		}
		c.Set("hmac_body", body)
		c.Next()
	}
}

// requireStrictHMAC checks the operator's strictHmac setting and enforces HMAC
// signature validation when enabled for that operator.
func (s *Server) requireStrictHMAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		op := getOperatorFromContext(c)
		if op == nil {
			// No operator in context means JWT auth didn't run or failed
			c.Next()
			return
		}
		// If operator has strictHmac disabled, skip HMAC validation
		if !op.Client.StrictHmac {
			c.Next()
			return
		}
		// Operator has strictHmac enabled — validate the signature
		_, err := s.HMAC.ReadAndVerifyHTTP(c.Request)
		if err != nil {
			c.JSON(401, map[string]string{"error": "bad_hmac", "message": "strictHmac is enabled: " + err.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}

// JWTAuth is a Gin middleware that validates the Bearer JWT and sets the operator in context.
// Use this to protect dashboard routes.
func JWTAuth(jwtManager *security.JWTManager, store *storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(401, map[string]string{"error": "unauthorized", "message": "missing or invalid Authorization header"})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtManager.Verify(token)
		if err != nil {
			msg := "invalid or expired token"
			if errors.Is(err, security.ErrExpiredToken) {
				msg = "token has expired"
			}
			c.JSON(401, map[string]string{"error": "unauthorized", "message": msg})
			c.Abort()
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		op, err := store.GetOperatorByEmail(ctx, claims.Email)
		if err != nil || op == nil {
			c.JSON(401, map[string]string{"error": "unauthorized", "message": "operator not found"})
			c.Abort()
			return
		}
		c.Set("operator", op)
		c.Next()
	}
}
