// Package handlers provides HTTP handlers for the Vyzorix API.
package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRF cookie configuration.
const (
	CSRFCookieName    = "vyz_csrf"
	CSRFCookieMaxAge  = 86400 // 24 hours in seconds
	CSRFCookiePath    = "/"
	CSRFTokenLength   = 32 // 32 bytes = 64 hex chars
)

// CSRFTokenResponse is the response for CSRF token endpoint.
type CSRFTokenResponse struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires_at"`
}

// GetCSRFToken returns a new CSRF token.
// The token is stored in a secure HttpOnly cookie and returned in the response.
func (s *Server) GetCSRFToken(c *gin.Context) {
	// Generate a random token.
	tokenBytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate CSRF token",
		})
		return
	}

	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(CSRFCookieMaxAge * time.Second).Unix()

	// Determine if secure cookie should be used (HTTPS only in production).
	isProduction := s.Config.Env == "production"
	secure := isProduction

	// Store token in secure HttpOnly cookie for validation.
	// Note: Secure=true is set in production, false in development (intentional for localhost testing).
	cookie := &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     CSRFCookiePath,
		MaxAge:   CSRFCookieMaxAge,
		HttpOnly: true,  // Prevent JavaScript access (XSS protection)
		Secure:   secure, // HTTPS only in production
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(c.Writer, cookie)

	// Also return token in JSON for clients that need it.
	c.JSON(http.StatusOK, CSRFTokenResponse{
		Token:   token,
		Expires: expiresAt,
	})
}

// ValidateCSRFToken validates the CSRF token from the request header against the cookie.
func (s *Server) ValidateCSRFToken(c *gin.Context) bool {
	// Get token from header (sent by client).
	headerToken := c.GetHeader("X-CSRF-Token")
	if headerToken == "" {
		return false
	}

	// Get stored token from cookie.
	storedToken, err := c.Cookie(CSRFCookieName)
	if err != nil {
		return false
	}

	// Constant-time comparison to prevent timing attacks.
	if len(headerToken) != len(storedToken) {
		return false
	}

	var result byte
	for i := 0; i < len(headerToken); i++ {
		result |= headerToken[i] ^ storedToken[i]
	}

	return result == 0
}
