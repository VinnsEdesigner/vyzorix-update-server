// Package adapters provides presenter/adapter implementations for GraphQL.
package adapters

import (
	"context"

	gqlerrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
)

// Presenter handles GraphQL response formatting and infrastructure concerns.
type Presenter struct {
	auditLogger *audit.Logger
}

// NewPresenter creates a new GraphQL presenter.
func NewPresenter(auditLogger *audit.Logger) *Presenter {
	return &Presenter{
		auditLogger: auditLogger,
	}
}

// LogAction logs a GraphQL operation using AdminAction.
func (p *Presenter) LogAction(ctx context.Context, operatorID, action, resourceType, resourceID string) {
	if p.auditLogger != nil && operatorID != "" {
		p.auditLogger.AdminAction(ctx, operatorID, action, resourceType, resourceID, "", nil)
	}
}

// DeviceView logs a device view.
func (p *Presenter) DeviceView(ctx context.Context, operatorID, deviceID string) {
	p.LogAction(ctx, operatorID, "graphql_device_view", "device", deviceID)
}

// DeviceList logs a device list query.
func (p *Presenter) DeviceList(ctx context.Context, operatorID string) {
	p.LogAction(ctx, operatorID, "graphql_device_list", "devices", "")
}

// DeviceCount logs a device count query.
func (p *Presenter) DeviceCount(ctx context.Context, operatorID string) {
	p.LogAction(ctx, operatorID, "graphql_device_count", "devices", "")
}

// CommandSend logs a command send.
func (p *Presenter) CommandSend(ctx context.Context, operatorID, deviceID, commandID string) {
	p.LogAction(ctx, operatorID, "graphql_command_send", "device", deviceID)
}

// CommandView logs a command view.
func (p *Presenter) CommandView(ctx context.Context, operatorID, commandID string) {
	p.LogAction(ctx, operatorID, "graphql_command_view", "command", commandID)
}

// CommandCancel logs a command cancel.
func (p *Presenter) CommandCancel(ctx context.Context, operatorID, commandID string) {
	p.LogAction(ctx, operatorID, "graphql_command_cancel", "command", commandID)
}

// DeviceDelete logs a device deletion.
func (p *Presenter) DeviceDelete(ctx context.Context, operatorID, deviceID string) {
	p.LogAction(ctx, operatorID, "graphql_device_delete", "device", deviceID)
}

// TelemetryQuery logs a telemetry query.
func (p *Presenter) TelemetryQuery(ctx context.Context, operatorID, deviceID string) {
	p.LogAction(ctx, operatorID, "graphql_telemetry_query", "device", deviceID)
}

// FCMTokenUpdate logs an FCM token update.
func (p *Presenter) FCMTokenUpdate(ctx context.Context, operatorID, deviceID string) {
	p.LogAction(ctx, operatorID, "graphql_fcm_token_update", "device", deviceID)
}

// BadRequestError returns a GraphQL bad request error.
func (p *Presenter) BadRequestError(message string) error {
	return gqlerrors.BadRequest("%s", message)
}

// NotFoundError returns a GraphQL not found error.
func (p *Presenter) NotFoundError(message string) error {
	return gqlerrors.NotFound("%s", message)
}

// InternalError returns a GraphQL internal error.
func (p *Presenter) InternalError(message string) error {
	return gqlerrors.Internal("%s", message)
}

// UnauthorizedError returns a GraphQL unauthorized error.
func (p *Presenter) UnauthorizedError() error {
	return gqlerrors.ErrUnauthorized
}

// ForbiddenError returns a GraphQL forbidden error.
func (p *Presenter) ForbiddenError(message string) error {
	return gqlerrors.Forbidden("%s", message)
}

// ConflictError returns a GraphQL conflict error.
func (p *Presenter) ConflictError(message string) error {
	return gqlerrors.Conflict("%s", message)
}
