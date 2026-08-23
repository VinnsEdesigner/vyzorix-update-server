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

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.UpdaterVersionManifestResult
	_ openapi.UpdaterCheckResult
	_ openapi.DownloadProgressResult
	_ openapi.ErrorResponse
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
	log      *slog.Logger
	manifest *VersionManifest
	dataDir  string
	binDir   string
	config   config.Config
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

// loadManifest loads the version manifest for APK integrity verification.

func (h *Handler) loadManifest() error {
	data, err := os.ReadFile(filepath.Join(h.dataDir, "version.json"))
	if err != nil {
		return err
	}

	var manifest VersionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	h.manifest = &manifest
	return nil
}

// ensureManifestLoaded loads the manifest if not already loaded.
func (h *Handler) ensureManifestLoaded() {
	if h.manifest == nil {
		if err := h.loadManifest(); err != nil {
			h.log.Warn("failed to load manifest for APK verification", "error", err)
		}
	}
}

// Version handles GET /api/v1/version.
// @Summary      OTA version manifest
// @Description  Returns the current OTA version manifest for Android clients
// @Tags         updater
// @Accept       json
// @Produce      json
// @Success      200  {object}  openapi.UpdaterVersionManifestResult  "version manifest"
// @Failure      404  {object}  openapi.ErrorResponse  "manifest not found"
// @Router       /version [get]
func (h *Handler) Version(c *gin.Context) {
	h.log.Info("ota version request", "path", c.Request.URL.Path)
	h.serveJSON(c, filepath.Join(h.dataDir, "version.json"))
}

// Changelog handles GET /api/v1/changelog.
// @Summary      OTA changelog
// @Description  Returns the release changelog served to Android clients
// @Tags         updater
// @Accept       json
// @Produce      json
// @Success      200  {object}  openapi.UpdateChangelogResult  "changelog"
// @Failure      404  {object}  openapi.ErrorResponse  "changelog not found"
// @Router       /changelog [get]
func (h *Handler) Changelog(c *gin.Context) {
	h.log.Info("ota changelog request", "path", c.Request.URL.Path)
	h.serveJSON(c, filepath.Join(h.dataDir, "changelog.json"))
}

// APK handles GET /api/v1/apk/:filename.
// @Summary      Download APK
// @Description  Serves APK files with optional Range support for resume
// @Tags         updater
// @Produce      octet-stream
// @Param        name  path  string  true  "APK filename"
// @Success      200  {file}  binary  "APK binary"
// @Failure      400  {object}  openapi.ErrorResponse  "filename required"
// @Failure      404  {object}  openapi.ErrorResponse  "file not found"
// @Router       /apk/{name} [get]
func (h *Handler) APK(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "filename required"))
		return
	}
	// Remove leading slash if present.
	filename = strings.TrimPrefix(filename, "/")
	h.serveAPK(c, filename)
}

// Bin handles GET /bin/:filename.
// @Summary      Download binary artifact
// @Description  Serves binary artifacts (same as APK but different path prefix)
// @Tags         updater
// @Produce      octet-stream
// @Param        name  path  string  true  "binary filename"
// @Success      200  {file}  binary  "binary artifact"
// @Failure      400  {object}  openapi.ErrorResponse  "filename required"
// @Failure      404  {object}  openapi.ErrorResponse  "file not found"
// @Router       /bin/{name} [get]
func (h *Handler) Bin(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "filename required"))
		return
	}
	// Remove leading slash if present.
	filename = strings.TrimPrefix(filename, "/")
	h.serveAPK(c, filename)
}

// CheckUpdate handles GET /api/v1/check-update.
// @Summary      Check for update
// @Description  Checks if an update is available for a device by version code
// @Tags         updater
// @Accept       json
// @Produce      json
// @Param        version_code  query int  false  "client version code"
// @Success      200  {object}  openapi.UpdaterCheckResult  "update check result"
// @Failure      500  {object}  openapi.ErrorResponse  "manifest not loaded"
// @Router       /check-update [get]
func (h *Handler) CheckUpdate(c *gin.Context) {
	versionCode := c.Query("version_code")
	h.log.Info("update check", "version_code", versionCode)

	data, err := os.ReadFile(filepath.Join(h.dataDir, "version.json"))
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "cannot read version file"))
		return
	}

	var manifest VersionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "invalid version file"))
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

	// Flatten response: update_available at top level, version fields directly accessible.
	c.JSON(http.StatusOK, gin.H{
		"update_available": manifest.VersionCode > clientCode,
		"version":          manifest.Version,
		"version_code":     manifest.VersionCode,
		"apk_filename":     manifest.APKFilename,
		"apk_sha256":       manifest.APKSHA256,
		"release_notes":    manifest.ReleaseNotes,
		"apk_size_bytes":   manifest.APKSizeBytes,
	})
}

// DownloadProgress handles POST /api/v1/download-progress.
// @Summary      Report download progress
// @Description  Tracks download progress for analytics
// @Tags         updater
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.DownloadProgressRequest  true  "download progress"
// @Success      200  {object}  openapi.DownloadProgressResult  "recorded"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid request"
// @Router       /download-progress [post]
func (h *Handler) DownloadProgress(c *gin.Context) {
	var req struct {
		DeviceID    string `json:"deviceId"`
		Filename    string `json:"filename"`
		Progress    int    `json:"progress"`
		BytesLoaded int64  `json:"bytesLoaded"`
		TotalBytes  int64  `json:"totalBytes"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid request"))
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
// Security: Verifies APK hash against manifest before serving.

func (h *Handler) serveAPK(c *gin.Context, filename string) {
	// Use filepath.Base to normalize path and prevent traversal attacks.
	// This handles .., %2F, %5C and other encoded sequences.
	baseFilename := filepath.Base(filename)

	// Ensure the result is a valid filename (not empty, no path separators).
	if baseFilename == "" || baseFilename == "." || baseFilename == ".." ||
		strings.ContainsAny(baseFilename, "/\\") {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid filename"))
		return
	}

	fpath := filepath.Join(h.binDir, baseFilename)

	// This ensures the APK hasn't been tampered with since the manifest was created.
	h.ensureManifestLoaded()
	if h.manifest != nil && h.manifest.APKFilename == baseFilename && h.manifest.APKSHA256 != "" {
		actualHash, err := h.computeFileHash(fpath)
		if err != nil {
			h.log.Error("failed to compute APK hash", "error", err, "file", fpath)
			_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to verify APK integrity"))
			return
		}

		// Constant-time comparison to prevent timing attacks.
		if !secureCompare(h.manifest.APKSHA256, actualHash) {
			h.log.Error("APK integrity check failed - hash mismatch with manifest",
				"file", baseFilename,
				"manifest_hash", h.manifest.APKSHA256,
				"actual_hash", actualHash,
			)
			_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "APK integrity verification failed - file may have been tampered with"))

			return
		}

		h.log.Info("APK integrity verified against manifest", "file", baseFilename)
	}

	// Also verify if client provides expected SHA256 (client-side verification).
	if clientHash := c.GetHeader("X-APK-SHA256"); clientHash != "" {
		actualHash, err := h.computeFileHash(fpath)
		if err != nil {
			h.log.Error("failed to compute APK hash", "error", err, "file", fpath)
			_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to verify APK integrity"))
			return
		}

		// Constant-time comparison to prevent timing attacks.
		if !secureCompare(clientHash, actualHash) {
			h.log.Warn("APK hash mismatch", "expected", clientHash, "actual", actualHash)
			_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthzInsufficientPermissions, "APK integrity check failed"))

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
	_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "GET or HEAD required"))
}

// computeFileHash computes the SHA256 hash of a file.
func (h *Handler) computeFileHash(fpath string) (string, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

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
