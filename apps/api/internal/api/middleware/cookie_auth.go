// Package middleware provides HTTP middleware.
package middleware

import (
	"context"
	"net/http"
	"time"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
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
			// No session cookie. If an API key is present, defer to the.
			// tenant API key middleware (registered after this one) instead.
			// of aborting, so tenant endpoints accept either credential.
			if apiKey := ctx.GetHeader("X-API-Key"); apiKey != "" {
				ctx.Next()
				return
			}
			responses.RespondStructured(ctx, http.StatusUnauthorized, "session cookie required")
			ctx.Abort()

			return
		}

		// Decrypt the session ID from the cookie.
		sessionID, err := c.sessionManager.DecryptSessionID(cookieValue)
		if err != nil {
			responses.RespondStructured(ctx, http.StatusUnauthorized, "invalid session")
			ctx.Abort()

			return
		}

		// Validate session (checks expiration and revocation).
		timeout, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
		defer cancel()

		sess, op, err := c.authService.ValidateSession(timeout, sessionID)
		if err != nil || op == nil {
			responses.RespondStructured(ctx, http.StatusUnauthorized, "session invalid or revoked")
			ctx.Abort()

			return
		}

		// Set operator and session in context for downstream handlers.
		ctx.Set(ContextKeyOperator, op)
		ctx.Set("operator_id", op.ID)
		ctx.Set("session", sess)
		// Also populate the GraphQL request context so REST middleware (e.g.
		// UpdatesAdminAuth.RequireAdmin) can read the operator via.
		// gqlcontext.GetOperator(c.Request.Context()).
		ctx.Request = ctx.Request.WithContext(gqlcontext.WithOperator(ctx.Request.Context(), op))
		ctx.Next()
	}
}

// GetOperatorFromContext retrieves the operator from the gin context.
func GetOperatorFromContext(c *gin.Context) *operator.Operator {
	val, exists := c.Get(ContextKeyOperator)
	if exists {
		if op, ok := val.(*operator.Operator); ok {
			return op
		}
	}

	// Fall back to operator_id (set by the tenant API key middleware for.
	// API-key-authenticated requests that never load the full operator).
	if opID, ok := c.Get("operator_id"); ok {
		if id, ok := opID.(string); ok && id != "" {
			return &operator.Operator{ID: id}
		}
	}

	return nil
}

// ContextKeySession is the key under which Middleware stores the validated.
// *session.Session (already decrypted from the vyz_session cookie).
const ContextKeySession = "session"

// GetSession retrieves the validated session the Middleware placed in context.
// Downstream handlers must use this (and GetCurrentSessionID) instead of.
// re-reading the raw vyz_session cookie: the cookie value is AES-GCM encrypted.
// and is NOT the plaintext session ID that the database stores under session.ID.
func GetSession(c *gin.Context) *session.Session {
	val, exists := c.Get(ContextKeySession)
	if !exists {
		return nil
	}

	sess, ok := val.(*session.Session)
	if !ok {
		return nil
	}

	return sess
}

// GetCurrentSessionID returns the decrypted ID of the session in context, or.
// "" when no authenticated session is present.
func GetCurrentSessionID(c *gin.Context) string {
	if sess := GetSession(c); sess != nil {
		return sess.ID
	}

	return ""
}
