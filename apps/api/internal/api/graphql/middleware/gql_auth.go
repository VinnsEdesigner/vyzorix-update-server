// Package middleware provides GraphQL middleware functions.
package middleware

import (
	"context"
	"log/slog"
	"strings"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	gqlerrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/errors"
	appsvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
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

// Authenticate extracts and validates the operator from request headers/cookies.
// Returns an error if authentication fails.
func (m *AuthMiddleware) Authenticate(ctx context.Context, headers map[string]string) (*operator.Operator, error) {
	// Try session cookie first
	if cookie, ok := headers["Cookie"]; ok {
		op, err := m.authenticateSession(ctx, cookie)
		if err == nil && op != nil {
			return op, nil
		}
	}

	// Try Authorization header (Bearer token)
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
	var sessionID string

	cookies := strings.Split(cookieHeader, ";")
	for _, cookie := range cookies {
		parts := strings.SplitN(strings.TrimSpace(cookie), "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "vyz_session" {
			sessionID = parts[1]
			break
		}
	}

	if sessionID == "" {
		return nil, gqlerrors.ErrUnauthorized
	}

	// Decrypt session ID from cookie
	decryptedSessionID, err := m.SessionManager.DecryptSessionID(sessionID)
	if err != nil {
		m.Log.Debug("session decryption failed", "err", err)
		return nil, gqlerrors.ErrUnauthorized
	}

	// Validate the session
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
// RequireRole returns a middleware that requires a specific role.
// IMPORTANT: In the organization-scoped model, roles are per-organization.
// This middleware only checks for ACTIVE MEMBERSHIP (any role). 
// Org-scoped authorization must be performed by individual resolvers 
// using GetMembership() or CheckCanManage*() methods.
func RequireRole(roles ...operator.OperatorRole) func(resolverFunc func(ctx context.Context) (interface{}, error)) func(ctx context.Context) (interface{}, error) {
	return func(resolverFunc func(ctx context.Context) (interface{}, error)) func(ctx context.Context) (interface{}, error) {
		return func(ctx context.Context) (interface{}, error) {
			op, ok := gqlcontext.GetOperator(ctx)
			if !ok || op == nil {
				return nil, gqlerrors.ErrUnauthorized
			}

			// Check if operator has any active membership
			// NOTE: Super admin and admin checks have been removed from here.
			// Org-scoped role checks are performed by individual resolvers.
			hasRole := false
			for _, role := range roles {
				switch role {
				case operator.RoleSuperAdmin, operator.RoleAdmin:
					// These are org-scoped - check in resolver
					// For now, check if operator has any active membership
					for _, m := range op.Memberships {
						if m.IsActive() && m.Role.Level() >= 1 {
							hasRole = true
							break
						}
					}
				case operator.RoleOperator:
					for _, m := range op.Memberships {
						if m.IsActive() && (m.Role.Level() >= 2) {
							hasRole = true
							break
						}
					}
				case operator.RoleViewer:
					if len(op.Memberships) > 0 {
						for _, m := range op.Memberships {
							if m.IsActive() {
								hasRole = true
								break
							}
						}
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				return nil, gqlerrors.Forbidden("active membership required")
			}

			return resolverFunc(ctx)
		}
	}
}
			}

			for _, role := range roles {
				if op.Role == role {
					return resolverFunc(ctx)
				}
			}

			return nil, gqlerrors.Forbidden("role %v required", roles)
		}
	}
}
