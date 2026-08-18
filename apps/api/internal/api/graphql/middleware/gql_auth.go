// Package middleware provides GraphQL middleware functions.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	gqlerrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/errors"
	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides authentication for GraphQL resolvers.
type AuthMiddleware struct {
	SessionManager *infraauth.SessionManager
	AuthService    *appsvc.AuthService
	Log            *slog.Logger
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(sessionManager *infraauth.SessionManager, authService *appsvc.AuthService, log *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		SessionManager: sessionManager,
		AuthService:    authService,
		Log:            log,
	}
}

// GetOperatorFromGinContext extracts the operator from the gin context, where.
// it was set by the system middleware chain (cookieAuth or tenantAPIKeyAuth).
// This allows the subscription handler to reuse the authenticated operator.
// without re-authenticating.
func (m *AuthMiddleware) GetOperatorFromGinContext(c *gin.Context) *operator.Operator {
	val, exists := c.Get("operator")
	if !exists {
		return nil
	}
	op, ok := val.(*operator.Operator)
	if !ok || op == nil {
		return nil
	}
	return op
}

// Authenticate extracts and validates the operator from request headers/cookies.
// Returns an error if authentication fails.
func (m *AuthMiddleware) Authenticate(ctx context.Context, headers map[string]string) (*operator.Operator, error) {
	// Try session cookie first.
	if cookieHeader, ok := headers["Cookie"]; ok {
		op, err := m.authenticateSession(ctx, cookieHeader)
		if err == nil && op != nil {
			return op, nil
		}
	}

	// Try Authorization header (Bearer token).
	if authHeader, ok := headers["Authorization"]; ok {
		op, err := m.authenticateBearer(ctx, authHeader)
		if err == nil && op != nil {
			return op, nil
		}
	}

	return nil, gqlerrors.ErrUnauthorized
}

// authenticateSession validates the session cookie.
func (m *AuthMiddleware) authenticateSession(ctx context.Context, cookieHeader string) (*operator.Operator, error) {
	// Parse cookies using net/http.
	cookieJar := http.Header{}
	cookieJar["Cookie"] = []string{cookieHeader}
	cookieReq := &http.Request{Header: cookieJar}

	sessionCookie, err := cookieReq.Cookie("vyz_session")
	if err != nil || sessionCookie.Value == "" {
		return nil, gqlerrors.ErrUnauthorized
	}

	// Decrypt session ID from cookie.
	decryptedSessionID, err := m.SessionManager.DecryptSessionID(sessionCookie.Value)
	if err != nil {
		m.Log.Debug("session decryption failed", "err", err)
		return nil, gqlerrors.ErrUnauthorized
	}

	// Validate the session.
	_, op, err := m.AuthService.ValidateSession(ctx, decryptedSessionID)
	if err != nil || op == nil {
		m.Log.Debug("session validation failed", "sessionID", decryptedSessionID, "err", err)
		return nil, gqlerrors.ErrUnauthorized
	}

	return op, nil
}

// authenticateBearer validates a Bearer token using JWT.
func (m *AuthMiddleware) authenticateBearer(ctx context.Context, authHeader string) (*operator.Operator, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, gqlerrors.ErrUnauthorized
	}

	token := parts[1]

	claims, err := m.AuthService.VerifyJWT(token)
	if err != nil {
		m.Log.Debug("bearer token validation failed", "err", err)
		return nil, gqlerrors.ErrUnauthorized
	}

	op, err := m.AuthService.GetOperatorByID(ctx, claims.OperatorID)
	if err != nil || op == nil {
		m.Log.Debug("operator not found for token", "operatorID", claims.OperatorID)
		return nil, gqlerrors.ErrUnauthorized
	}

	return op, nil
}

// RequireAuth is a middleware function that ensures authentication.
// Use this as a wrapper for resolvers that require authentication.
func RequireAuth(resolverFunc func(ctx context.Context) (interface{}, error)) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		op, ok := gqlcontext.GetOperator(ctx)
		if !ok || op == nil {
			return nil, gqlerrors.ErrUnauthorized
		}

		return resolverFunc(ctx)
	}
}

// RequireSuperAdmin requires the operator to be a SuperAdmin.
func RequireSuperAdmin(resolverFunc func(ctx context.Context) (interface{}, error)) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		op, ok := gqlcontext.GetOperator(ctx)
		if !ok || op == nil {
			return nil, gqlerrors.ErrUnauthorized
		}

		if !op.IsSuperAdmin() {
			return nil, gqlerrors.Forbidden("superadmin required")
		}

		return resolverFunc(ctx)
	}
}

// RequireAdmin requires the operator to be an Admin in at least one organization.
func RequireAdmin(resolverFunc func(ctx context.Context) (interface{}, error)) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		op, ok := gqlcontext.GetOperator(ctx)
		if !ok || op == nil {
			return nil, gqlerrors.ErrUnauthorized
		}

		if !op.IsAdmin() {
			return nil, gqlerrors.Forbidden("admin required")
		}

		return resolverFunc(ctx)
	}
}

// RequireOperator requires the operator to be an Operator in at least one organization.
func RequireOperator(resolverFunc func(ctx context.Context) (interface{}, error)) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		op, ok := gqlcontext.GetOperator(ctx)
		if !ok || op == nil {
			return nil, gqlerrors.ErrUnauthorized
		}

		if !op.IsOperator() {
			return nil, gqlerrors.Forbidden("operator required")
		}

		return resolverFunc(ctx)
	}
}

// RequireViewer requires the operator to be a Viewer in at least one organization.
func RequireViewer(resolverFunc func(ctx context.Context) (interface{}, error)) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		op, ok := gqlcontext.GetOperator(ctx)
		if !ok || op == nil {
			return nil, gqlerrors.ErrUnauthorized
		}

		if !op.IsViewer() {
			return nil, gqlerrors.Forbidden("viewer required")
		}

		return resolverFunc(ctx)
	}
}
