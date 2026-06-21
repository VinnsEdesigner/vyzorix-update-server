// Package middleware provides HTTP middleware.
package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// NoCache returns a middleware that sets Cache-Control: no-store, no-cache, must-revalidate.
func NoCache() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// CacheWithPrivate returns a middleware that allows private caching.
func CacheWithPrivate(maxAge int) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Header("Cache-Control", fmt.Sprintf("private, max-age=%d", maxAge))
		c.Next()
	}
}

// SecurityHeaders returns a middleware that adds security headers to all responses.
func SecurityHeaders() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Content Security Policy.
		// Restrict sources to same origin, and only allow script from same origin.
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'")

		// HTTP Strict Transport Security (HSTS).
		// Force HTTPS for 1 year, include subdomains, and add preload flag.
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// X-Frame-Options.
		// Prevent clickjacking by not allowing framing.
		c.Header("X-Frame-Options", "DENY")

		// X-Content-Type-Options.
		// Prevent MIME type sniffing.
		c.Header("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection.
		// Enable browser's XSS filter (legacy, but still useful for older browsers).
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy.
		// Only send referrer for same-origin requests.
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy.
		// Disable unnecessary browser features.
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		// Cross-Origin-Opener-Policy.
		// Enable cross-origin isolation for SharedArrayBuffer support.
		c.Header("Cross-Origin-Opener-Policy", "same-origin")

		// Cross-Origin-Embedder-Policy.
		// Require explicit permission for loading cross-origin resources.
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")

		// Cross-Origin-Resource-Policy.
		// Only allow same-origin resource loading.
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		c.Next()
	}
}

// SecurityHeadersRelaxed returns a middleware with relaxed CSP for development.
// Use this in development mode only.
func SecurityHeadersRelaxed() func(c *gin.Context) {
	return func(c *gin.Context) {
		// More permissive CSP for development.
		c.Header("Content-Security-Policy", "default-src 'self' 'unsafe-inline' 'unsafe-eval'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; connect-src 'self' ws://localhost:* http://localhost:*")

		// HSTS with shorter max-age for development.
		c.Header("Strict-Transport-Security", "max-age=86400") // 1 day

		// Other security headers remain the same.
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		c.Next()
	}
}