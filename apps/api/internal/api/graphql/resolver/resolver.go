// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"context"

	gqladapters "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/adapters"
	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	gqlmiddleware "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/validator"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dashboard"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
)

// Resolver is the root GraphQL resolver.
type Resolver struct {
	// Services
	DeviceService   *device.Service
	CommandService  *cmdapp.Service
	HistoryService  *cmdapp.HistoryService
	DashboardSvc    *dashboard.Service
	LogsSvc         *logs.Service
	MetricsSvc       *appmetrics.Service
	Hub             *hub.Hub
	TelemetryRepo   *storage.TelemetryRepository
	LogsRepo        *storage.LogsRepository
	MetricsRepo     *storage.MetricsRepository
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
	historyService *cmdapp.HistoryService,
	dashboardSvc *dashboard.Service,
	logsSvc *logs.Service,
	metricsSvc *appmetrics.Service,
	hub *hub.Hub,
	telemetryRepo *storage.TelemetryRepository,
	logsRepo *storage.LogsRepository,
	metricsRepo *storage.MetricsRepository,
	fcmNotifier fcm.Notifier,
	authMiddleware *gqlmiddleware.AuthMiddleware,
	presenter *gqladapters.Presenter,
) *Resolver {
	return &Resolver{
		DeviceService:  deviceService,
		CommandService: commandService,
		HistoryService: historyService,
		DashboardSvc:   dashboardSvc,
		LogsSvc:        logsSvc,
		MetricsSvc:     metricsSvc,
		Hub:             hub,
		TelemetryRepo:   telemetryRepo,
		LogsRepo:        logsRepo,
		MetricsRepo:     metricsRepo,
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
