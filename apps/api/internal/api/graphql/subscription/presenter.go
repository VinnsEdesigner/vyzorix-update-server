// Package subscription provides GraphQL subscription support via WebSocket.
package subscription

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

// auditLoggerAdapter wraps audit.Logger to implement AuditLogger.
type auditLoggerAdapter struct {
	logger *audit.Logger
}

// NewAuditLoggerAdapter creates an adapter that wraps audit.Logger.
func NewAuditLoggerAdapter(logger *audit.Logger) AuditLogger {
	return &auditLoggerAdapter{logger: logger}
}

// LogAction implements AuditLogger.
func (a *auditLoggerAdapter) LogAction(ctx context.Context, operatorID, action, resourceType, resourceID string) {
	a.logger.AdminAction(ctx, operatorID, action, resourceType, resourceID, "", nil)
}

// Presenter handles subscription logging and audit.
type Presenter struct {
	auditLogger AuditLogger
	log         Logger
}

// NewPresenter creates a new subscription presenter.
func NewPresenter(log Logger, auditLogger AuditLogger) *Presenter {
	return &Presenter{
		log:         log,
		auditLogger: auditLogger,
	}
}

// LogConnect logs a subscription connection.
func (p *Presenter) LogConnect(ctx context.Context, op *operator.Operator) {
	if p.log != nil {
		p.log.Info("subscription client connected", "operatorID", op.ID)
	}
}

// LogDisconnect logs a subscription disconnection.
func (p *Presenter) LogDisconnect(ctx context.Context, op *operator.Operator) {
	if p.log != nil {
		p.log.Info("subscription client disconnected", "operatorID", op.ID)
	}
}

// LogAuthFail logs an authentication failure.
func (p *Presenter) LogAuthFail(ctx context.Context, err error) {
	if p.log != nil {
		p.log.Debug("subscription auth failed", "err", err)
	}
}

// LogWebSocketError logs a WebSocket error.
func (p *Presenter) LogWebSocketError(ctx context.Context, operation string, err error) {
	if p.log != nil {
		p.log.Error("websocket "+operation+" failed", "err", err)
	}
}

// LogMessageError logs a message processing error.
func (p *Presenter) LogMessageError(ctx context.Context, err error) {
	if p.log != nil {
		p.log.Debug("websocket message error", "err", err)
	}
}

// AuditSubscribe logs a subscription audit event.
func (p *Presenter) AuditSubscribe(ctx context.Context, op *operator.Operator, subscriptionType, resourceID string) {
	if p.auditLogger != nil {
		p.auditLogger.LogAction(ctx, op.ID, "graphql_subscription_"+subscriptionType, "subscription", resourceID)
	}
}

// AuditUnsubscribe logs an unsubscription audit event.
func (p *Presenter) AuditUnsubscribe(ctx context.Context, op *operator.Operator, subscriptionType string) {
	if p.auditLogger != nil {
		p.auditLogger.LogAction(ctx, op.ID, "graphql_unsubscription_"+subscriptionType, "subscription", "")
	}
}

// Logger interface for subscription logging.
type Logger interface {
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}
