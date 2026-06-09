// Package controllers provides HTTP handlers.
package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
	"github.com/gin-gonic/gin"
)

// UpdaterController handles OTA distribution endpoints.
// It serves version.json, changelog.json, and APK files with Range support.
type UpdaterController struct {
	notifier fcm.Notifier
	log      *slog.Logger
	store    *storage.Store
	config   config.Config
}

func NewUpdaterController(log *slog.Logger, cfg config.Config, st *storage.Store, notifier fcm.Notifier) *UpdaterController {
	return &UpdaterController{log: log, config: cfg, store: st, notifier: notifier}
}

// Version serves the version manifest for OTA updates.
// GET /api/v1/version.
func (s *UpdaterController) Version(c *gin.Context) {
	s.log.Info("ota version request", "path", c.Request.URL.Path)
	s.serveJSON(c, filepath.Join(s.config.DataDir, "version.json"))
}

// Changelog serves the release changelog.
// GET /api/v1/changelog.
func (s *UpdaterController) Changelog(c *gin.Context) {
	s.log.Info("ota changelog request", "path", c.Request.URL.Path)
	s.serveJSON(c, filepath.Join(s.config.DataDir, "changelog.json"))
}

// APK serves APK files with optional Range support for resume.
// GET /api/v1/apk/:filename.
func (s *UpdaterController) APK(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "filename required"})
		return
	}
	filename = strings.TrimPrefix(filename, "/")
	s.serveAPK(c, filename)
}

// Bin serves binary artifacts (same as APK but different path prefix).
// GET /bin/:filename.
func (s *UpdaterController) Bin(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "filename required"})
		return
	}
	filename = strings.TrimPrefix(filename, "/")
	s.serveAPK(c, filename)
}

// serveJSON serves a static JSON file.
func (s *UpdaterController) serveJSON(c *gin.Context, path string) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Cache-Control", "no-store")
	c.File(path)
}

// serveAPK serves an APK file with proper headers and Range support.
func (s *UpdaterController) serveAPK(c *gin.Context, filename string) {
	if filename == "" || strings.ContainsAny(filename, "/\\") {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "invalid filename"})
		return
	}

	fpath := filepath.Join(s.config.BinDir, filename)

	if c.Request.Method == http.MethodGet {
		c.Header("Content-Type", "application/vnd.android.package-archive")
		c.Header("Cache-Control", "no-store")
		c.File(fpath)
		return
	}

	if c.Request.Method == http.MethodHead {
		c.Header("Content-Type", "application/vnd.android.package-archive")
		c.File(fpath)
		return
	}

	c.JSON(400, map[string]string{"error": "bad_request", "message": "GET or HEAD required"})
}

// CheckUpdate checks if an update is available for a device.
// GET /api/v1/check-update?version_code=X.
func (s *UpdaterController) CheckUpdate(c *gin.Context) {
	versionCode := c.Query("version_code")
	s.log.Info("update check", "version_code", versionCode)

	var version models.VersionManifest
	data, err := os.ReadFile(filepath.Join(s.config.DataDir, "version.json"))
	if err != nil {
		c.JSON(500, map[string]string{"error": "server_error", "message": "cannot read version file"})
		return
	}
	if err := json.Unmarshal(data, &version); err != nil {
		c.JSON(500, map[string]string{"error": "server_error", "message": "invalid version file"})
		return
	}

	clientCode := 0
	if versionCode != "" {
		var err error
		clientCode, err = strconv.Atoi(versionCode)
		if err != nil {
			s.log.Warn("invalid client version code", "versionCode", versionCode, "err", err)
		}
	}

	updateAvailable := version.VersionCode > clientCode
	c.JSON(200, map[string]any{
		"update_available": updateAvailable,
		"version":          version,
	})
}

// DownloadProgress tracks download progress for analytics.
// POST /api/v1/download-progress.
func (s *UpdaterController) DownloadProgress(c *gin.Context) {
	var req struct {
		DeviceID    string `json:"deviceId"`
		Filename    string `json:"filename"`
		Progress    int    `json:"progress"`
		BytesLoaded int64  `json:"bytesLoaded"`
		TotalBytes  int64  `json:"totalBytes"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, map[string]string{"error": "bad_json", "message": err.Error()})
		return
	}
	s.log.Info("download progress", "deviceId", req.DeviceID, "filename", req.Filename, "progress", req.Progress)
	c.JSON(200, map[string]any{"recorded": true})
}

func (s *UpdaterController) Config() config.Config  { return s.config }
func (s *UpdaterController) Store() *storage.Store  { return s.store }
func (s *UpdaterController) Notifier() fcm.Notifier { return s.notifier }
