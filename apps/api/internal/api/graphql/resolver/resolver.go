// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"context"

	gqladapters "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/adapters"
	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	gqlmiddleware "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/validator"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
)

// Resolver is the root GraphQL resolver.
type Resolver struct {
	// Services
	DeviceService  *device.Service
	CommandService *cmdapp.Service
	Hub            *hub.Hub
	TelemetryRepo  *storage.TelemetryRepository
	FCMNotifier    fcm.Notifier

	// Middleware
	AuthMiddleware *gqlmiddleware.AuthMiddleware

	// Presenter for audit logging and error handling
	Presenter *gqladapters.Presenter

	// Utilities
	Validator *validator.Validator
}

// NewResolver creates a new GraphQL resolver.
func NewResolver(
	deviceService *device.Service,
	commandService *cmdapp.Service,
	hub *hub.Hub,
	telemetryRepo *storage.TelemetryRepository,
	fcmNotifier fcm.Notifier,
	authMiddleware *gqlmiddleware.AuthMiddleware,
	presenter *gqladapters.Presenter,
) *Resolver {
	return &Resolver{
		DeviceService:  deviceService,
		CommandService: commandService,
		Hub:            hub,
		TelemetryRepo:  telemetryRepo,
		FCMNotifier:    fcmNotifier,
		AuthMiddleware: authMiddleware,
		Presenter:      presenter,
		Validator:      validator.New(),
	}
}

// RequireAuth ensures the operator is authenticated.
func (r *Resolver) RequireAuth(ctx context.Context) (*context.Context, error) {
	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	return &ctx, nil
}
