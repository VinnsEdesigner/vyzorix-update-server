// Package middleware provides SSR proxy middleware
package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	security "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
)

// SSRProxy creates a reverse proxy to the Node.js SSR server with JWT validation.
// Includes health monitoring and automatic fallback to SPA when SSR is unavailable.
func SSRProxy(log *slog.Logger, ssrConfig config.SSRConfig, publicDir string, jwtSecret string) gin.HandlerFunc {
	if !ssrConfig.EnableSSR {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	ssrServerURL, err := url.Parse(ssrConfig.SSRServerURL)
	if err != nil {
		log.Error("invalid SSR server URL", "err", err, "url", ssrConfig.SSRServerURL)
		return func(c *gin.Context) {
			c.Next()
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(ssrServerURL)

	// Health monitoring state
	var (
		ssrHealthy   = true
		ssrHealthyMu sync.RWMutex
	)

	// Start background health checker
	go func() {
		ticker := time.NewTicker(time.Duration(ssrConfig.SSRHealthCheckInterval) * time.Second)
		defer ticker.Stop()

		client := &http.Client{Timeout: 5 * time.Second}

		for range ticker.C {
			resp, err := client.Get(ssrConfig.SSRServerURL + "/health")
			if err != nil || resp.StatusCode >= 500 {
				ssrHealthyMu.Lock()
				if ssrHealthy {
					log.Warn("SSR server health check failed, will use SPA fallback", "err", err)
					ssrHealthy = false
				}
				ssrHealthyMu.Unlock()
				if err == nil {
					_ = resp.Body.Close()
				}
				continue
			}
			_ = resp.Body.Close()

			ssrHealthyMu.Lock()
			wasHealthy := ssrHealthy
			ssrHealthy = true
			ssrHealthyMu.Unlock()

			if !wasHealthy {
				log.Info("SSR server recovered, resuming SSR mode")
			}
		}
	}()

	// Custom director to properly modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = ssrServerURL.Scheme
		req.URL.Host = ssrServerURL.Host

		// Forward important headers for SSR
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", ssrServerURL.Scheme)
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)

		// Keep original host for the SSR server to generate absolute URLs if needed
		req.Header.Set("X-Original-Host", req.Host)
		req.Header.Set("X-Original-URI", req.RequestURI)

		if originalDirector != nil {
			originalDirector(req)
		}
	}

	// Custom modify response to handle errors and logging
	proxy.ModifyResponse = func(res *http.Response) error {
		log.Debug("SSR proxy response", "status", res.StatusCode, "path", res.Request.URL.Path)

		if res.StatusCode >= 500 {
			body, err := io.ReadAll(res.Body)
			if err == nil {
				res.Body = io.NopCloser(bytes.NewBuffer(body))
				log.Error("SSR server error", "status", res.StatusCode, "path", res.Request.URL.Path, "body", string(body))
			}
		}
		return nil
	}

	// serveFallback serves the static SPA with resilience
	serveFallback := func(w http.ResponseWriter, req *http.Request) {
		// Try multiple fallback files
		fallbackFiles := []string{
			filepath.Join(publicDir, "index.html"),
			filepath.Join(publicDir, "landing.html"),
		}

		for _, fallbackPath := range fallbackFiles {
			if _, err := os.Stat(fallbackPath); err == nil {
				http.ServeFile(w, req, fallbackPath)
				return
			}
		}

		// Ultimate fallback - hardcoded minimal HTML
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html><head><title>Service Temporarily Unavailable</title></head>
<body><h1>Service Temporarily Unavailable</h1>
<p>The server is starting up. Please refresh in a moment.</p></body></html>`))
	}

	// Custom error handler with graceful fallback to static HTML
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		log.Error("SSR proxy error, falling back to SPA", "err", err, "path", req.URL.Path)
		serveFallback(w, req)
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip proxying for API, static assets, and health checks
		if strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/v1/") ||
			strings.HasPrefix(path, "/health") ||
			strings.HasPrefix(path, "/bin/") ||
			strings.Contains(path, ".") {
			c.Next()
			return
		}

		// Check SSR health before proxying
		ssrHealthyMu.RLock()
		healthy := ssrHealthy
		ssrHealthyMu.RUnlock()

		if !healthy {
			log.Debug("SSR unhealthy, using SPA fallback", "path", path)
			serveFallback(c.Writer, c.Request)
			c.Abort()
			return
		}

		// ============================================================
		// SECURITY: Validate JWT before proxying protected routes
		// ============================================================

		// Public routes that don't require authentication (must match React Router routes)
		publicRoutes := []string{
			"/login",
			"/create-account",
			"/forgot-password",
			"/set-password",
			"/waitVerify",
			"/auth/callback",
		}
		for _, public := range publicRoutes {
			if strings.HasPrefix(path, public) {
				log.Debug("Proxying to SSR server (public route)", "path", path)
				proxy.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// For all other routes, validate JWT cookie
		tokenCookie, err := c.Cookie("vyz.auth.token")
		if err != nil || tokenCookie == "" {
			log.Warn("SSR access denied - no JWT cookie", "path", path, "ip", c.ClientIP())
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			return
		}

		// Validate JWT
		jwtManager := security.NewJWTManager(jwtSecret, 0, "")
		claims, err := jwtManager.Verify(tokenCookie)
		if err != nil {
			log.Warn("SSR access denied - invalid JWT", "path", path, "ip", c.ClientIP(), "err", err)
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			return
		}

		log.Debug("SSR access granted", "path", path, "email", claims.Email)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
