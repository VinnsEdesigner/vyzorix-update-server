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
	"github.com/VinnsEdesigner/vyzorix-update-server/models"
	"github.com/VinnsEdesigner/vyzorix-update-server/security"
	"github.com/VinnsEdesigner/vyzorix-update-server/storage"
	"github.com/gorilla/websocket"
)

type Server struct {
	Log      *slog.Logger
	Config   config.Config
	Store    *storage.Store
	Hub      *hub.Hub
	HMAC     security.Verifier
	Upgrader websocket.Upgrader
}

func New(log *slog.Logger, cfg config.Config, st *storage.Store, h *hub.Hub) *Server {
	s := &Server{Log: log, Config: cfg, Store: st, Hub: h}
	s.HMAC = security.Verifier{Window: cfg.HMACWindow, Nonces: security.NewNonceCache(cfg.HMACWindow), Secret: func(id string) (string, bool) { return st.Secret(context.Background(), id) }}
	s.Upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return s.originAllowed(r.Header.Get("Origin")) }}
	return s
}
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/api/v1/version", s.version)
	mux.HandleFunc("/api/v1/changelog", s.changelog)
	mux.HandleFunc("/api/v1/apk/", s.apk)
	mux.HandleFunc("/bin/", s.bin)
	mux.HandleFunc("/v1/device/register", s.register)
	mux.HandleFunc("/v1/device/", s.deviceScoped)
	mux.HandleFunc("/v1/dashboard/devices", s.dashboardDevices)
	mux.HandleFunc("/", s.dashboard)
	return s.cors(s.logRequests(mux))
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		method(w, "GET, HEAD")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.Store.Ping(ctx); err != nil {
		writeJSON(w, 503, map[string]any{"ok": false, "database": "down"})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "database": "ok"})
}
func (s *Server) version(w http.ResponseWriter, r *http.Request) {
	serveStatic(w, r, filepath.Join(s.Config.DataDir, "version.json"), "application/json; charset=utf-8")
}
func (s *Server) changelog(w http.ResponseWriter, r *http.Request) {
	serveStatic(w, r, filepath.Join(s.Config.DataDir, "changelog.json"), "application/json; charset=utf-8")
}
func (s *Server) apk(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/apk/")
	s.serveDownload(w, r, name)
}
func (s *Server) bin(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/bin/")
	s.serveDownload(w, r, name)
}
func (s *Server) serveDownload(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		method(w, "GET, HEAD")
		return
	}
	if name == "" || strings.ContainsAny(name, "/\\") {
		http.Error(w, "invalid filename", 400)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.android.package-archive")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, filepath.Join(s.Config.BinDir, name))
}
func serveStatic(w http.ResponseWriter, r *http.Request, path, ct string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		method(w, "GET, HEAD")
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, path)
}
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		method(w, "POST")
		return
	}
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_json", err.Error())
		return
	}
	if req.DeviceID == "" || req.FirebaseInstallID == "" {
		writeError(w, 400, "missing_field", "deviceId and firebaseInstallId are required")
		return
	}
	d, _, err := s.Store.Register(r.Context(), req)
	if errors.Is(err, storage.ErrHijack) {
		writeError(w, 409, "device_hijack", err.Error())
		return
	}
	if err != nil {
		writeError(w, 500, "register_failed", err.Error())
		return
	}
	writeJSON(w, 200, models.RegisterResponse{DeviceID: d.ID, CommandSecret: d.CommandSecret, RegisteredAt: d.RegisteredAt.UnixMilli(), ServerTime: time.Now().UnixMilli()})
}
func (s *Server) deviceScoped(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/v1/device/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	tail := ""
	if len(parts) == 2 {
		tail = parts[1]
	}
	switch {
	case tail == "status" && r.Method == http.MethodGet:
		s.status(w, r, id)
	case tail == "fcm-token" && r.Method == http.MethodPatch:
		s.fcmToken(w, r, id)
	case tail == "command" && r.Method == http.MethodPost:
		s.command(w, r, id)
	case tail == "stream":
		s.stream(w, r, id)
	case tail == "" && r.Method == http.MethodDelete:
		s.deleteDevice(w, r, id)
	default:
		http.NotFound(w, r)
	}
}
func (s *Server) status(w http.ResponseWriter, r *http.Request, id string) {
	d, ok, err := s.Store.Device(r.Context(), id)
	if err != nil {
		writeError(w, 500, "lookup_failed", err.Error())
		return
	}
	if !ok {
		writeError(w, 404, "unknown_device", id)
		return
	}
	writeJSON(w, 200, models.DeviceStatus{DeviceID: d.ID, Online: s.Hub.Online(id) || d.Online, LastSeen: d.LastSeen.UnixMilli(), AppVersion: d.AppVersion, DeviceClass: d.DeviceClass})
}
func (s *Server) fcmToken(w http.ResponseWriter, r *http.Request, id string) {
	body, ok := s.requireHMAC(w, r, id)
	if !ok {
		return
	}
	var req struct {
		FCMToken string `json:"fcmToken"`
	}
	_ = json.Unmarshal(body, &req)
	if req.FCMToken == "" {
		writeError(w, 400, "missing_field", "fcmToken is required")
		return
	}
	if err := s.Store.UpdateFCM(r.Context(), id, req.FCMToken); err != nil {
		writeError(w, 500, "update_failed", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"deviceId": id, "serverTime": time.Now().UnixMilli()})
}
func (s *Server) deleteDevice(w http.ResponseWriter, r *http.Request, id string) {
	if _, ok := s.requireHMAC(w, r, id); !ok {
		return
	}
	if err := s.Store.DeleteDevice(r.Context(), id); err != nil {
		writeError(w, 500, "delete_failed", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"deviceId": id, "deleted": true})
}
func (s *Server) command(w http.ResponseWriter, r *http.Request, id string) {
	body, ok := s.authorizeDashboardOrHMAC(w, r, id)
	if !ok {
		return
	}
	var req models.CommandRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, 400, "bad_json", err.Error())
		return
	}
	if req.Command == "" {
		writeError(w, 400, "missing_field", "command is required")
		return
	}
	if _, found, err := s.Store.Device(r.Context(), id); err != nil {
		writeError(w, 500, "lookup_failed", err.Error())
		return
	} else if !found {
		writeError(w, 404, "unknown_device", id)
		return
	}
	frame := models.CommandFrame{Type: "command", DispatchID: storage.NewDispatchID(), Command: req.Command, Args: req.Args, Nonce: req.Nonce, Timestamp: req.Timestamp, Signature: req.Signature}
	delivery := "queued"
	if s.Hub.Send(id, frame) {
		delivery = "sent"
	}
	_ = s.Store.SaveCommand(r.Context(), frame.DispatchID, id, req.Command, req.Args, delivery)
	writeJSON(w, 202, models.CommandResponse{DispatchID: frame.DispatchID, Delivery: delivery, ServerTime: time.Now().UnixMilli()})
}
func (s *Server) stream(w http.ResponseWriter, r *http.Request, id string) {
	if s.Config.EnforceHMAC {
		if err := s.HMAC.Verify(r.Method, r.URL.RequestURI(), id, nil, r.Header); err != nil {
			writeError(w, 401, "bad_hmac", err.Error())
			return
		}
	}
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Log.Warn("websocket upgrade failed", "err", err)
		return
	}
	c := &hub.Client{DeviceID: id, Conn: conn, Send: make(chan models.CommandFrame, 32), Hub: s.Hub}
	s.Hub.Register(c)
	go c.WritePump()
	c.ReadPump()
}
func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		method(w, "GET, HEAD")
		return
	}
	clean := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
	if clean == "." || clean == "" {
		clean = "index.html"
	}
	candidate := filepath.Join(s.Config.PublicDir, clean)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		http.ServeFile(w, r, candidate)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.Config.PublicDir, "index.html"))
}
func (s *Server) dashboardDevices(w http.ResponseWriter, r *http.Request) {
	if !s.checkDashboard(r) {
		writeError(w, 401, "unauthorized", "missing or invalid dashboard token")
		return
	}
	ds, err := s.Store.Devices(r.Context())
	if err != nil {
		writeError(w, 500, "list_failed", err.Error())
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
	writeJSON(w, 200, map[string]any{"devices": out})
}
func (s *Server) requireHMAC(w http.ResponseWriter, r *http.Request, id string) ([]byte, bool) {
	body, err := s.HMAC.ReadAndVerify(r, id)
	if err != nil {
		if !s.Config.EnforceHMAC && errors.Is(err, security.ErrMissing) {
			return body, true
		}
		writeError(w, 401, "bad_hmac", err.Error())
		return body, false
	}
	return body, true
}
func (s *Server) authorizeDashboardOrHMAC(w http.ResponseWriter, r *http.Request, id string) ([]byte, bool) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, 400, "read_failed", err.Error())
		return nil, false
	}
	if s.checkDashboard(r) {
		return body, true
	}
	r.Body = bodyReader(body)
	if _, ok := s.requireHMAC(w, r, id); !ok {
		return body, false
	}
	return body, true
}
