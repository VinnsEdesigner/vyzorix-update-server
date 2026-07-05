// Package middleware provides SSR proxy middleware.
package middleware

import (
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

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
)

// SSRProxy creates a reverse proxy to the Node.js SSR server with JWT validation.
// Includes health monitoring and automatic fallback to SPA when SSR is unavailable.
// Returns a handler and a cleanup function that should be called when shutting down.
func SSRProxy(log *slog.Logger, ssrConfig config.SSRConfig, publicDir string, jwtSecret string) (gin.HandlerFunc, func()) {
	if !ssrConfig.EnableSSR {
		return func(c *gin.Context) { c.Next() }, func() {}
	}

	ssrServerURL, err := url.Parse(ssrConfig.SSRServerURL)
	if err != nil {
		log.Error("invalid SSR server URL", "err", err, "url", ssrConfig.SSRServerURL)
		return func(c *gin.Context) { c.Next() }, func() {}
	}

	proxy := httputil.NewSingleHostReverseProxy(ssrServerURL)
	stopCh := make(chan struct{})
	ssrHealthy, ssrHealthyMu := setupSSRHealthChecker(log, ssrConfig, stopCh)
	setupProxyDirector(proxy, ssrServerURL)
	setupProxyErrorHandler(proxy, log, publicDir)

	handler := handleSSRRequest(log, proxy, ssrHealthy, ssrHealthyMu, jwtSecret, publicDir)
	cleanup := func() { close(stopCh) }
	return handler, cleanup
}

func setupSSRHealthChecker(log *slog.Logger, ssrConfig config.SSRConfig, stopCh <-chan struct{}) (bool, *sync.RWMutex) {
	var ssrHealthy = true
	var ssrHealthyMu sync.RWMutex

	go func() {
		ticker := time.NewTicker(time.Duration(ssrConfig.SSRHealthCheckInterval) * time.Second)
		defer ticker.Stop()
		client := &http.Client{Timeout: 5 * time.Second}

		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				checkSSRHealth(log, ssrConfig, client, &ssrHealthy, &ssrHealthyMu)
			}
		}
	}()

	return ssrHealthy, &ssrHealthyMu
}

func checkSSRHealth(log *slog.Logger, ssrConfig config.SSRConfig, client *http.Client, ssrHealthy *bool, ssrHealthyMu *sync.RWMutex) {
	resp, err := client.Get(ssrConfig.SSRServerURL + "/health")
	if err != nil || resp.StatusCode >= 500 {
		ssrHealthyMu.Lock()
		if *ssrHealthy {
			log.Warn("SSR server health check failed, will use SPA fallback", "err", err)
			*ssrHealthy = false
		}
		ssrHealthyMu.Unlock()
		if err == nil {
			_ = resp.Body.Close()
		}
		return
	}
	_ = resp.Body.Close()
	ssrHealthyMu.Lock()
	wasHealthy := *ssrHealthy
	*ssrHealthy = true
	ssrHealthyMu.Unlock()
	if !wasHealthy {
		log.Info("SSR server recovered, resuming SSR mode")
	}
}

func setupProxyDirector(proxy *httputil.ReverseProxy, ssrServerURL *url.URL) {
	proxy.Rewrite = func(r *httputil.ProxyRequest) {
		r.Out.URL.Scheme = ssrServerURL.Scheme
		r.Out.URL.Host = ssrServerURL.Host
		r.Out.Header.Set("X-Forwarded-Host", ssrServerURL.Host)
		r.Out.Header.Set("X-Forwarded-Proto", ssrServerURL.Scheme)
		r.Out.Header.Set("X-Forwarded-For", r.In.RemoteAddr)
		r.Out.Header.Set("X-Original-Host", r.In.Host)
		r.Out.Header.Set("X-Original-URI", r.In.RequestURI)
	}
}

func setupProxyErrorHandler(proxy *httputil.ReverseProxy, log *slog.Logger, publicDir string) {
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		log.Error("SSR proxy error, falling back to SPA", "err", err, "path", req.URL.Path)
		serveFallback(w, req, publicDir)
	}
}

func handleSSRRequest(log *slog.Logger, proxy *httputil.ReverseProxy, ssrHealthy bool, ssrHealthyMu *sync.RWMutex, jwtSecret, publicDir string) gin.HandlerFunc {
	publicRoutes := []string{"/login", "/create-account", "/forgot-password", "/set-password", "/waitVerify", "/auth/callback"}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if isStaticPath(path) {
			c.Next()
			return
		}
		if !isSSRHealthy(ssrHealthy, ssrHealthyMu) {
			log.Debug("SSR unhealthy, using SPA fallback", "path", path)
			serveFallback(c.Writer, c.Request, publicDir)
			c.Abort()
			return
		}
		if isPublicRoute(path, publicRoutes) {
			log.Debug("Proxying to SSR server (public route)", "path", path)
			proxy.ServeHTTP(c.Writer, c.Request)
			return
		}
		handleAuthenticatedRequest(c, log, path, proxy, jwtSecret)
	}
}

func isStaticPath(path string) bool {
	return strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/v1/") ||
		strings.HasPrefix(path, "/health") ||
		strings.HasPrefix(path, "/bin/") ||
		strings.Contains(path, ".")
}

func isSSRHealthy(healthy bool, mu *sync.RWMutex) bool {
	mu.RLock()
	defer mu.RUnlock()
	return healthy
}

func isPublicRoute(path string, publicRoutes []string) bool {
	for _, public := range publicRoutes {
		if strings.HasPrefix(path, public) {
			return true
		}
	}
	return false
}

func handleAuthenticatedRequest(c *gin.Context, log *slog.Logger, path string, proxy *httputil.ReverseProxy, jwtSecret string) {
	tokenCookie, err := c.Cookie("vyz.infraauth.token")
	if err != nil || tokenCookie == "" {
		log.Warn("SSR access denied - no JWT cookie", "path", path, "ip", c.ClientIP())
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}
	jwtManager, err := infraauth.NewJWTManager(jwtSecret, 0, "")
	if err != nil {
		log.Error("SSR failed - invalid JWT secret configuration", "path", path, "ip", c.ClientIP(), "err", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}
	claims, err := jwtManager.Verify(tokenCookie)
	if err != nil {
		log.Warn("SSR access denied - invalid JWT", "path", path, "ip", c.ClientIP(), "err", err)
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}
	log.Debug("SSR access granted", "path", path, "email", claims.Email)
	proxy.ServeHTTP(c.Writer, c.Request)
}

func serveFallback(w http.ResponseWriter, req *http.Request, publicDir string) {
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html><head><title>Service Temporarily Unavailable</title></head>
<body><h1>Service Temporarily Unavailable</h1>
<p>The server is starting up. Please refresh in a moment.</p></body></html>`))
}
