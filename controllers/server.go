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

	"github.com/VinnsEdesigner/vyzorix-update-server/config"
	"github.com/VinnsEdesigner/vyzorix-update-server/hub"
	"github.com/VinnsEdesigner/vyzorix-update-server/middleware"
	"github.com/VinnsEdesigner/vyzorix-update-server/models"
	"github.com/VinnsEdesigner/vyzorix-update-server/security"
	"github.com/VinnsEdesigner/vyzorix-update-server/services/fcm"
	"github.com/VinnsEdesigner/vyzorix-update-server/storage"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
	Log       *slog.Logger
	Config   config.Config
	Store    *storage.Store
	Hub      *hub.Hub
	Notifier fcm.Notifier
	HMAC     security.Verifier
	Limiter  *middleware.RateLimiter
	jwtCtrl  *AuthController
}

func New(log *slog.Logger, cfg config.Config, st *storage.Store, h *hub.Hub, notifier fcm.Notifier) *Server {
	s := &Server{Log: log, Config: cfg, Store: st, Hub: h, Notifier: notifier}
	s.HMAC = security.Verifier{
		Window: cfg.HMACWindow,
		Nonces: security.NewNonceCache(cfg.HMACWindow),
		Secret: func(id string) (string, bool) { return st.Secret(context.Background(), id) },
	}
	s.Limiter = middleware.NewRateLimiter(100, time.Minute)
	s.jwtCtrl = NewAuthController(log, cfg, st)
	return s
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) Engine() *gin.Engine {
	if s.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(s.Log))
	r.Use(middleware.CORSHandler(s.Config.AllowedOrigins))

	// Auth routes (no JWT required for login/register; JWT required for /me and logout)
	auth := r.Group("/v1/auth")
	auth.GET("/google", s.authCtrl.GoogleLoginRedirect) // triggers OAuth redirect
	auth.GET("/google/callback", s.authCtrl.GoogleCallback) // OAuth callback from Google
	auth.POST("/login", s.authCtrl.Login)
	auth.POST("/register", s.authCtrl.Register)
	// /me and /logout require JWT — middleware applied inline
	auth.GET("/me", JWTAuth(s.jwtCtrl, s.Store), s.authCtrl.Me)
	auth.PATCH("/me", JWTAuth(s.jwtCtrl, s.Store), s.authCtrl.UpdateName)
	auth.POST("/logout", JWTAuth(s.jwtCtrl, s.Store), s.authCtrl.Logout)

	r.GET("/health", s.health)
	r.GET("/healthz", s.health)
	r.GET("/api/v1/version", s.version)
	r.GET("/api/v1/changelog", s.changelog)
	r.GET("/api/v1/apk/*name", s.apk)
	r.GET("/bin/*name", s.bin)

	// Root path → native HTML landing page (no React needed)
	// Explicit route so Gin doesn't need to resolve /*path wildcard for /
	r.GET("/", s.dashboard)

	// Device routes
	r.POST("/v1/device/register", s.register)
	r.GET("/v1/device/:id/status", s.status)
	r.PATCH("/v1/device/:id/fcm-token", s.requireHMAC(), s.fcmToken)
	r.POST("/v1/device/:id/command", s.authorizeDashboardOrHMAC(), s.command)
	r.DELETE("/v1/device/:id", s.requireHMAC(), s.deleteDevice)

	// Dashboard routes — protected by JWT
	r.GET("/v1/dashboard/devices", JWTAuth(s.jwtCtrl, s.Store), s.dashboardDevices)

	// WebSocket — HMAC or JWT
	r.GET("/v1/device/:id/stream", s.stream)

	// SPA fallback — must be last
	r.GET("/*path", s.dashboard)

	return r
}

func (s *Server) Routes() http.Handler { return s.Engine() }

func (s *Server) health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	dbOk := s.Store.Ping(ctx) == nil

	connectedDevices := 0
	if s.Hub != nil {
		connectedDevices = s.Hub.ClientCount()
	}

	version := ""
	if v, err := s.readVersion(); err == nil {
		version = v
	}

	c.JSON(200, map[string]any{
		"ok":                true,
		"database":          map[bool]string{true: "ok", false: "down"}[dbOk],
		"dbOk":              dbOk,
		"serverTime":        time.Now().UnixMilli(),
		"connectedDevices":  connectedDevices,
		"version":           version,
	})
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

func (s *Server) version(c *gin.Context)  { serveStaticFile(c, filepath.Join(s.Config.DataDir, "version.json"), "application/json; charset=utf-8") }
func (s *Server) changelog(c *gin.Context) { serveStaticFile(c, filepath.Join(s.Config.DataDir, "changelog.json"), "application/json; charset=utf-8") }

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
	var req struct{ FCMToken string `json:"fcmToken"` }
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
	_ = s.Store.SaveCommand(c.Request.Context(), frame.DispatchID, id, req.Command, req.Args, delivery)
	if delivery == "queued" && s.Notifier != nil {
		if d, found, err := s.Store.Device(c.Request.Context(), id); err == nil && found {
			if err := s.Notifier.SendSilentWake(c.Request.Context(), fcm.SilentWake{Token: d.FCMToken, Command: req.Command, DispatchID: frame.DispatchID, DeviceID: id}); err != nil {
				s.Log.Warn("fcm wake failed", "deviceId", id, "dispatchId", frame.DispatchID, "err", err)
				_ = s.Store.MarkWake(c.Request.Context(), frame.DispatchID, err.Error())
			} else {
				_ = s.Store.MarkWake(c.Request.Context(), frame.DispatchID, "")
			}
		}
	}
	c.JSON(202, models.CommandResponse{DispatchID: frame.DispatchID, Delivery: delivery, ServerTime: time.Now().UnixMilli()})
}

func (s *Server) stream(c *gin.Context) {
	if s.Config.EnforceHMAC {
		if err := s.HMAC.Verify(c.Request.Method, c.Request.URL.RequestURI(), c.Param("id"), nil, c.Request.Header); err != nil {
			c.JSON(401, map[string]string{"error": "bad_hmac", "message": err.Error()})
			return
		}
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
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
		Online      bool   `json:"online"`
		LastSeen    int64  `json:"lastSeen"`
		AppVersion  string `json:"appVersion"`
		DeviceClass string `json:"deviceClass"`
	}
	out := make([]row, 0, len(ds))
	for _, d := range ds {
		out = append(out, row{d.ID, s.Hub.Online(d.ID) || d.Online, d.LastSeen.UnixMilli(), d.AppVersion, d.DeviceClass})
	}
	c.JSON(200, map[string]any{"devices": out})
}

func (s *Server) dashboardAuth() gin.HandlerFunc {
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
			if !s.Config.EnforceHMAC && errors.Is(err, security.ErrMissing) {
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

func (s *Server) authorizeDashboardOrHMAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.Config.TokenSecret != "" {
			auth := c.GetHeader("Authorization")
			token := c.GetHeader("X-Vyzorix-Token")
			if auth == "Bearer "+s.Config.TokenSecret || token == s.Config.TokenSecret {
				c.Next()
				return
			}
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
