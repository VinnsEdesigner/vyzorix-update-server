// Package middleware provides SSR proxy middleware
package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
)

// SSRProxy creates a reverse proxy to the Node.js SSR server.
func SSRProxy(log *slog.Logger, ssrConfig config.SSRConfig) gin.HandlerFunc {
	if !ssrConfig.EnableSSR {
		// If SSR is disabled, return a no-op middleware.
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Parse the SSR server URL.
	ssrServerURL, err := url.Parse(ssrConfig.SSRServerURL)
	if err != nil {
		log.Error("invalid SSR server URL", "err", err, "url", ssrConfig.SSRServerURL)
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(ssrServerURL)

	// Custom director to modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// Set the target host
		req.URL.Scheme = ssrServerURL.Scheme
		req.URL.Host = ssrServerURL.Host

		// Copy original director behavior
		if originalDirector != nil {
			originalDirector(req)
		}
	}

	// Custom modify response to handle errors
	proxy.ModifyResponse = func(res *http.Response) error {
		// Log SSR response status
		log.Debug("SSR proxy response", "status", res.StatusCode, "path", res.Request.URL.Path)

		// If SSR server returns an error, we could fall back to client-side rendering
		// But for now, we'll just log it
		if res.StatusCode >= 500 {
			body, err := io.ReadAll(res.Body)
			if err == nil {
				res.Body = io.NopCloser(bytes.NewBuffer(body))
				log.Error("SSR server error", "status", res.StatusCode, "path", res.Request.URL.Path, "body", string(body))
			}
		}

		return nil
	}

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		log.Error("SSR proxy error", "err", err, "path", req.URL.Path)

		// Fallback: serve the client-side HTML and let React handle routing
		// This is a graceful degradation strategy
		http.Error(w, "SSR unavailable - falling back to client-side rendering", http.StatusBadGateway)
	}

	return func(c *gin.Context) {
		// Check if this is an HTML request (not API, not static assets)
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/v1/") ||
			strings.HasPrefix(path, "/health") ||
			strings.HasPrefix(path, "/bin/") ||
			strings.Contains(path, ".") { // static files
			// Let the next handler (Go server) handle API and static files
			c.Next()
			return
		}

		// Log the SSR request
		log.Info("Proxying to SSR server", "path", path, "method", c.Request.Method)

		// Proxy the request to SSR server
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
