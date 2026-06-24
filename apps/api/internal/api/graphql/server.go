// Package graphql provides GraphQL server integration.
package graphql

import (
	"log/slog"

	gqladapters "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/adapters"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/handler"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/subscription"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/validator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

// Config holds the GraphQL server configuration.
type Config struct {
	// Service dependencies
	AuthService    *auth.AuthService
	SessionManager *infraauth.SessionManager
	DeviceService  *device.Service
	CommandService *command.Service
	TelemetryRepo  *storage.TelemetryRepository
	Hub            *hub.Hub
	FCMNotifier    fcm.Notifier
	Log            *slog.Logger
	AuditLogger    *audit.Logger
}

// Server provides GraphQL HTTP handling.
type Server struct {
	handler *handler.Handler
}

// NewServer creates a new GraphQL server.
func NewServer(cfg *Config) (*Server, error) {
	// Create auth middleware
	authMw := middleware.NewAuthMiddleware(cfg.SessionManager, cfg.AuthService, cfg.Log)

	// Create GraphQL presenter for audit logging
	presenter := gqladapters.NewPresenter(cfg.AuditLogger)

	// Create resolver
	res := resolver.NewResolver(
		cfg.DeviceService,
		cfg.CommandService,
		cfg.Hub,
		cfg.TelemetryRepo,
		cfg.FCMNotifier,
		authMw,
		presenter,
	)

	// Create handler
	h, err := handler.NewHandler(&handler.Config{
		Resolver:       res,
		AuthMiddleware: authMw,
		PlaygroundPath: "/playground",
	})
	if err != nil {
		return nil, err
	}

	return &Server{
		handler: h,
	}, nil
}

// Routes registers GraphQL routes with Gin.
func (s *Server) Routes(r *gin.Engine) {
	s.handler.Routes(r)
}

// Handler returns the GraphQL handler for custom routing.
func (s *Server) Handler() *handler.Handler {
	return s.handler
}

// BuildSchema builds the GraphQL schema (exposed for testing).
func BuildSchema(res *resolver.Resolver) (graphql.Schema, error) {
	return schema.BuildSchema(res)
}

// DefaultValidator returns a new validator instance.
func DefaultValidator() *validator.Validator {
	return validator.New()
}

// NewSubscriptionHandler creates a WebSocket handler for GraphQL subscriptions.
func NewSubscriptionHandler(
	hub *hub.Hub,
	res *resolver.Resolver,
	authService *auth.AuthService,
	sessionManager *infraauth.SessionManager,
	log *slog.Logger,
	auditLogger *audit.Logger,
) *subscription.Handler {
	authMw := middleware.NewAuthMiddleware(sessionManager, authService, log)
	return subscription.NewHandler(&subscription.Config{
		Hub:          hub,
		Resolver:     res,
		AuthMw:      authMw,
		Logger:      log,
		AuditLogger: subscription.NewAuditLoggerAdapter(auditLogger),
	})
}
