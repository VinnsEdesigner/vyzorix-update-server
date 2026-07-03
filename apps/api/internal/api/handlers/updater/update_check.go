package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

// VersionManifest represents the OTA version manifest served to Android clients.
type VersionManifest struct {
	Version      string `json:"version"`
	APKFilename  string `json:"apk_filename"`
	APKSHA256    string `json:"apk_sha256"`
	ReleaseNotes string `json:"release_notes"`
	VersionCode  int    `json:"version_code"`
	APKSizeBytes int64  `json:"apk_size_bytes"`
}

// Handler handles OTA update distribution endpoints.
type Handler struct {
	log     *slog.Logger
	dataDir string
	binDir  string
	config  config.Config
}

// NewHandler creates a new UpdaterHandler.
func NewHandler(log *slog.Logger, cfg config.Config) *Handler {
	return &Handler{
		log:     log,
		config:  cfg,
		dataDir: cfg.DataDir,
		binDir:  cfg.BinDir,
	}
}

// Version handles GET /api/v1/version.
// Returns the version manifest for OTA updates.
func (h *Handler) Version(c *gin.Context) {
	h.log.Info("ota version request", "path", c.Request.URL.Path)
	h.serveJSON(c, filepath.Join(h.dataDir, "version.json"))
}

// Changelog handles GET /api/v1/changelog.
// Returns the release changelog.
func (h *Handler) Changelog(c *gin.Context) {
	h.log.Info("ota changelog request", "path", c.Request.URL.Path)
	h.serveJSON(c, filepath.Join(h.dataDir, "changelog.json"))
}

// APK handles GET /api/v1/apk/:filename.
// Serves APK files with optional Range support for resume.
func (h *Handler) APK(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "filename required"})
		return
	}
	// Remove leading slash if present
	filename = strings.TrimPrefix(filename, "/")
	h.serveAPK(c, filename)
}

// Bin handles GET /bin/:filename.
// Serves binary artifacts (same as APK but different path prefix).
func (h *Handler) Bin(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "filename required"})
		return
	}
	// Remove leading slash if present
	filename = strings.TrimPrefix(filename, "/")
	h.serveAPK(c, filename)
}

// CheckUpdate handles GET /api/v1/check-update.
// Checks if an update is available for a device.
func (h *Handler) CheckUpdate(c *gin.Context) {
	versionCode := c.Query("version_code")
	h.log.Info("update check", "version_code", versionCode)

	var version VersionManifest

	data, err := os.ReadFile(filepath.Join(h.dataDir, "version.json"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "cannot read version file"})
		return
	}

	if err := json.Unmarshal(data, &version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "invalid version file"})
		return
	}

	clientCode := 0

	if versionCode != "" {
		var parseErr error

		clientCode, parseErr = strconv.Atoi(versionCode)
		if parseErr != nil {
			h.log.Warn("invalid client version code", "versionCode", versionCode, "err", parseErr)
		}
	}

	updateAvailable := version.VersionCode > clientCode
	c.JSON(http.StatusOK, gin.H{
		"update_available": updateAvailable,
		"version":          version,
	})
}

// DownloadProgress handles POST /api/v1/download-progress.
// Tracks download progress for analytics.
func (h *Handler) DownloadProgress(c *gin.Context) {
	var req struct {
		DeviceID    string `json:"deviceId"`
		Filename    string `json:"filename"`
		Progress    int    `json:"progress"`
		BytesLoaded int64  `json:"bytesLoaded"`
		TotalBytes  int64  `json:"totalBytes"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request"})
		return
	}

	h.log.Info("download progress", "deviceId", req.DeviceID, "filename", req.Filename, "progress", req.Progress)
	c.JSON(http.StatusOK, gin.H{"recorded": true})
}

// serveJSON serves a static JSON file with proper headers.
func (h *Handler) serveJSON(c *gin.Context, path string) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Cache-Control", "no-store")
	c.File(path)
}

// serveAPK serves an APK file with proper headers and Range support.
// Security: Verifies APK hash if X-APK-SHA256 header is provided by client.
func (h *Handler) serveAPK(c *gin.Context, filename string) {
	if filename == "" || strings.ContainsAny(filename, "/\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid filename"})
		return
	}

	fpath := filepath.Join(h.binDir, filename)

	// Verify APK hash if client provides expected SHA256
	if clientHash := c.GetHeader("X-APK-SHA256"); clientHash != "" {
		actualHash, err := h.computeFileHash(fpath)
		if err != nil {
			h.log.Error("failed to compute APK hash", "error", err, "file", fpath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to verify APK integrity"})
			return
		}

		// Constant-time comparison to prevent timing attacks
		if !secureCompare(clientHash, actualHash) {
			h.log.Warn("APK hash mismatch", "expected", clientHash, "actual", actualHash)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "hash_mismatch",
				"message": "APK integrity check failed",
				"expected": clientHash,
			})
			return
		}

		h.log.Info("APK hash verified", "file", filename, "hash", actualHash[:16]+"...")
	}

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

	c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "GET or HEAD required"})
}

// computeFileHash computes the SHA256 hash of a file.
func (h *Handler) computeFileHash(fpath string) (string, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// secureCompare performs a constant-time comparison of two strings.
// This prevents timing attacks when comparing sensitive values like hashes.
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
