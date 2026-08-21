package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// healthHandler handles health check requests.
func (s *Server) healthHandler(c *gin.Context) {
	dbOk := false

	var dbErr error

	if s.DB != nil {
		if err := s.DB.Ping(); err == nil {
			dbOk = true
		} else {
			dbErr = err
		}
	}

	if s.lifecycleManager != nil {
		workers := s.lifecycleManager.Health()
		allHealthy := true
		for _, healthy := range workers {
			if !healthy {
				allHealthy = false
			}
		}
		if !allHealthy {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "degraded",
				"workers":   workers,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"workers":   workers,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
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

// versionHandler serves the version.json file.
func (s *Server) versionHandler(c *gin.Context) {
	s.serveStaticFile(c, filepath.Join(s.config.DataDir, "version.json"), "application/json; charset=utf-8")
}

// changelogHandler serves the changelog.json file.
func (s *Server) changelogHandler(c *gin.Context) {
	s.serveStaticFile(c, filepath.Join(s.config.DataDir, "changelog.json"), "application/json; charset=utf-8")
}

// apkHandler serves APK download files.
func (s *Server) apkHandler(c *gin.Context) {
	name := c.Param("name")
	s.serveDownload(c, strings.TrimPrefix(name, "/"))
}

// binHandler serves binary download files.
func (s *Server) binHandler(c *gin.Context) {
	name := c.Param("name")
	s.serveDownload(c, strings.TrimPrefix(name, "/"))
}

// serveDownload serves a file download with APK headers.
func (s *Server) serveDownload(c *gin.Context, name string) {
	if name == "" || strings.ContainsAny(name, "/\\") {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid filename"))
		return
	}

	c.Header("Content-Type", "application/vnd.android.package-archive")
	c.Header("Cache-Control", "no-store")
	c.File(filepath.Join(s.config.BinDir, name))
}

// serveStaticFile serves a static file with specified content type.
func (s *Server) serveStaticFile(c *gin.Context, path, ct string) {
	c.Header("Content-Type", ct)
	c.Header("Cache-Control", "no-store")
	c.File(path)
}

// dashboardHandler serves the dashboard or landing page.
func (s *Server) dashboardHandler(c *gin.Context) {
	path := c.Request.URL.Path

	// All non-API paths → serve the landing page first (fallback).
	// SSR middleware handles "/" when enabled (proxies to SSR or serves fallback landing.html).
	clean := strings.TrimPrefix(filepath.Clean(path), "/")
	if clean == "." || clean == "" {
		clean = "landing.html"
	}

	candidate := filepath.Join(s.config.PublicDir, clean)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		c.File(candidate)
		return
	}

	c.File(filepath.Join(s.config.PublicDir, "landing.html"))
}

// requireHMAC is middleware that validates HMAC signatures for device API requests.

func (s *Server) requireHMAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := c.Param("imei")
		if deviceID == "" {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "device ID required for HMAC verification"))
			c.Abort()
			return
		}

		body, err := s.hmacVerifier.ReadAndVerify(c.Request, deviceID)
		if err != nil {
			if !s.config.EnforceHMAC {
				c.Next()
				return
			}
			_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Invalid request"))
			c.Abort()

			return
		}

		c.Set("hmac_body", body)
		c.Next()
	}
}

// requireStrictHMAC checks the operator's strictHmac setting and enforces HMAC.
func (s *Server) requireStrictHMAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		op := middleware.GetOperatorFromContext(c)
		if op == nil {
			c.Next()
			return
		}

		if !op.ClientSettings.StrictHmac {
			c.Next()
			return
		}

		_, err := s.hmacVerifier.ReadAndVerifyHTTP(c.Request)
		if err != nil {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "strictHmac is enabled: HMAC signature verification failed"))
			c.Abort()

			return
		}

		c.Next()
	}
}
