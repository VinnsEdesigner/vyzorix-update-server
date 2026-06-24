// Package middleware provides HTTP middleware.
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

const (
	// CookieName is the name of the session cookie.
	CookieName = "vyz_session"
	// ContextKeyOperator is the key for the operator in gin context.
	ContextKeyOperator = "operator"
)

// CookieAuth is middleware that validates the HttpOnly session cookie and sets the operator in context.
type CookieAuth struct {
	sessionManager *infraauth.SessionManager
	authService    *auth.AuthService
}

// NewCookieAuth creates a new CookieAuth middleware.
func NewCookieAuth(sessionManager *infraauth.SessionManager, authService *auth.AuthService) *CookieAuth {
	return &CookieAuth{
		sessionManager: sessionManager,
		authService:    authService,
	}
}

// Middleware returns the gin handler function.
func (c *CookieAuth) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Read the session cookie.
		cookieValue, err := ctx.Cookie(CookieName)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "session cookie required"})
			ctx.Abort()

			return
		}

		// Decrypt the session ID from the cookie.
		sessionID, err := c.sessionManager.DecryptOperatorID(cookieValue)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "invalid session"})
			ctx.Abort()

			return
		}

		// Validate session (checks expiration and revocation).
		timeout, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
		defer cancel()

		sess, op, err := c.authService.ValidateSession(timeout, sessionID)
		if err != nil || op == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "session invalid or revoked"})
			ctx.Abort()

			return
		}

		// Set operator and session in context for downstream handlers.
		ctx.Set(ContextKeyOperator, op)
		ctx.Set("session", sess)
		ctx.Next()
	}
}

// GetOperatorFromContext retrieves the operator from the gin context.
func GetOperatorFromContext(c *gin.Context) *operator.Operator {
	val, exists := c.Get(ContextKeyOperator)
	if !exists {
		return nil
	}

	op, ok := val.(*operator.Operator)
	if !ok {
		return nil
	}

	return op
}
