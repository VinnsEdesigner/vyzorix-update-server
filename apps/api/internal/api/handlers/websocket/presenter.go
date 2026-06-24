// Package websocket provides WebSocket handler implementations.
package websocket

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
)

// Presenter handles WebSocket audit logging and infrastructure concerns.
type Presenter struct {
	auditLogger *audit.Logger
	log         Logger
}

// Logger interface for WebSocket logging.
type Logger interface {
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// NewPresenter creates a new WebSocket presenter.
func NewPresenter(auditLogger *audit.Logger, log Logger) *Presenter {
	return &Presenter{
		auditLogger: auditLogger,
		log:         log,
	}
}

// LogConnect logs a device WebSocket connection.
func (p *Presenter) LogConnect(ctx context.Context, deviceID string) {
	if p.log != nil {
		p.log.Info("device websocket connected", "deviceId", deviceID)
	}
}

// LogDisconnect logs a device WebSocket disconnection.
func (p *Presenter) LogDisconnect(ctx context.Context, deviceID string) {
	if p.log != nil {
		p.log.Info("device websocket disconnected", "deviceId", deviceID)
	}
}

// LogUpgradeFailed logs a WebSocket upgrade failure.
func (p *Presenter) LogUpgradeFailed(ctx context.Context, deviceID string, reason string) {
	if p.log != nil {
		p.log.Warn("websocket upgrade failed", "deviceId", deviceID, "reason", reason)
	}
}

// LogHMACFailed logs an HMAC verification failure.
func (p *Presenter) LogHMACFailed(ctx context.Context, deviceID string) {
	if p.log != nil {
		p.log.Warn("websocket HMAC verification failed", "deviceId", deviceID)
	}
}

// LogTelemetryReceived logs telemetry received from a device.
func (p *Presenter) LogTelemetryReceived(ctx context.Context, deviceID string, riskScore int) {
	if p.log != nil {
		p.log.Debug("telemetry received", "deviceId", deviceID, "riskScore", riskScore)
	}
}

// LogTelemetrySaveFailed logs a telemetry save failure.
func (p *Presenter) LogTelemetrySaveFailed(ctx context.Context, deviceID string, err error) {
	if p.log != nil {
		p.log.Warn("telemetry save failed", "deviceId", deviceID, "err", err)
	}
}

// LogBadFrame logs a malformed WebSocket frame.
func (p *Presenter) LogBadFrame(ctx context.Context, deviceID string, err error) {
	if p.log != nil {
		p.log.Warn("bad websocket frame", "deviceId", deviceID, "err", err)
	}
}

// LogRateLimited logs a rate limiting event.
func (p *Presenter) LogRateLimited(ctx context.Context, deviceID string, dispatchID string) {
	if p.log != nil {
		p.log.Debug("client rate limited, dropping message", "deviceId", deviceID, "dispatchId", dispatchID)
	}
}

// LogReadError logs a WebSocket read error.
func (p *Presenter) LogReadError(ctx context.Context, deviceID string, err error) {
	if p.log != nil {
		p.log.Debug("ws read error", "deviceId", deviceID, "err", err)
	}
}

// LogWriteError logs a WebSocket write error.
func (p *Presenter) LogWriteError(ctx context.Context, deviceID string, err error) {
	if p.log != nil {
		p.log.Warn("ws write error", "deviceId", deviceID, "err", err)
	}
}

// LogClientCloseFailed logs a client close failure.
func (p *Presenter) LogClientCloseFailed(ctx context.Context, deviceID string, err error) {
	if p.log != nil {
		p.log.Warn("client close failed", "deviceId", deviceID, "err", err)
	}
}

// AuditDeviceConnect logs a device connection audit event.
func (p *Presenter) AuditDeviceConnect(ctx context.Context, deviceID string) {
	if p.auditLogger != nil {
		p.auditLogger.AdminAction(ctx, "", "websocket_connect", "device", deviceID, "", nil)
	}
}

// AuditDeviceDisconnect logs a device disconnection audit event.
func (p *Presenter) AuditDeviceDisconnect(ctx context.Context, deviceID string) {
	if p.auditLogger != nil {
		p.auditLogger.AdminAction(ctx, "", "websocket_disconnect", "device", deviceID, "", nil)
	}
}

// AuditTelemetryReceived logs telemetry received audit event.
func (p *Presenter) AuditTelemetryReceived(ctx context.Context, deviceID string) {
	if p.auditLogger != nil {
		p.auditLogger.AdminAction(ctx, "", "websocket_telemetry", "device", deviceID, "", nil)
	}
}
