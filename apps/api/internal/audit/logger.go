// Package audit provides audit logging functionality for security events.
package audit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/uuid"
)

// Logger handles logging of security events to the audit repository.
type Logger struct {
	repo *Repository
	log  *slog.Logger
}

// LoggerConfig holds configuration for the audit logger.
type LoggerConfig struct {
	Enabled       bool
	RetentionDays int
}

// DefaultLoggerConfig returns the default audit logger configuration.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Enabled:       true,
		RetentionDays: 90,
	}
}

// NewLogger creates a new audit logger.
func NewLogger(repo *Repository, log *slog.Logger, cfg LoggerConfig) *Logger {
	return &Logger{
		repo: repo,
		log:  log,
	}
}

// LogEvent logs a security event to the audit repository asynchronously.
func (l *Logger) LogEvent(ctx context.Context, entry *Entry) {
	if entry.ID == "" {
		entry.ID = uuid.New()
	}

	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	go func() {
		if err := l.repo.Log(context.Background(), entry); err != nil {
			l.log.Error("failed to write audit log",
				slog.String("action", string(entry.Action)),
				slog.String("error", err.Error()))
		}
	}()
}

// LoginSuccess logs a successful login event.
func (l *Logger) LoginSuccess(ctx context.Context, operatorID, ipAddress, userAgent string) {
	l.LogEvent(ctx, &Entry{
		OperatorID: operatorID,
		Action:     ActionLoginSuccess,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Result:     ResultSuccess,
	})
}

// LoginFailed logs a failed login attempt.
func (l *Logger) LoginFailed(ctx context.Context, operatorID, ipAddress, userAgent, reason string) {
	l.LogEvent(ctx, &Entry{
		OperatorID: operatorID,
		Action:     ActionLoginFailed,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Result:     ResultFailure,
		Metadata:   map[string]string{"reason": reason},
	})
}

// Logout logs a logout event.
func (l *Logger) Logout(ctx context.Context, operatorID, ipAddress, userAgent string) {
	l.LogEvent(ctx, &Entry{
		OperatorID: operatorID,
		Action:     ActionLogout,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Result:     ResultSuccess,
	})
}

// Register logs a registration event.
func (l *Logger) Register(ctx context.Context, operatorID, ipAddress, userAgent string) {
	l.LogEvent(ctx, &Entry{
		OperatorID: operatorID,
		Action:     ActionRegister,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Result:     ResultSuccess,
	})
}

// PasswordChange logs a password change event.
func (l *Logger) PasswordChange(ctx context.Context, operatorID, ipAddress, userAgent string) {
	l.LogEvent(ctx, &Entry{
		OperatorID: operatorID,
		Action:     ActionPasswordChange,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Result:     ResultSuccess,
	})
}

// AccountLocked logs an account lockout event.
func (l *Logger) AccountLocked(ctx context.Context, operatorID, ipAddress string, attempts int) {
	l.LogEvent(ctx, &Entry{
		OperatorID: operatorID,
		Action:     ActionAccountLocked,
		IPAddress:  ipAddress,
		Result:     ResultBlocked,
		Metadata:   map[string]string{"attempts": fmt.Sprintf("%d", attempts)},
	})
}

// SessionRevoked logs a session revocation event.
func (l *Logger) SessionRevoked(ctx context.Context, operatorID, ipAddress, userAgent, reason string) {
	l.LogEvent(ctx, &Entry{
		OperatorID: operatorID,
		Action:     ActionSessionRevoked,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Result:     ResultSuccess,
		Metadata:   map[string]string{"reason": reason},
	})
}

// CSRFFailure logs a CSRF validation failure.
func (l *Logger) CSRFFailure(ctx context.Context, ipAddress, userAgent string) {
	l.LogEvent(ctx, &Entry{
		Action:    ActionCSRFFailure,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Result:    ResultBlocked,
	})
}

// SigningFailure logs a request signing validation failure.
func (l *Logger) SigningFailure(ctx context.Context, ipAddress, userAgent, reason string) {
	l.LogEvent(ctx, &Entry{
		Action:    ActionSigningFailure,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Result:    ResultBlocked,
		Metadata:  map[string]string{"reason": reason},
	})
}

// RateLimitExceeded logs a rate limit exceeded event.
func (l *Logger) RateLimitExceeded(ctx context.Context, ipAddress, userAgent, endpoint string) {
	l.LogEvent(ctx, &Entry{
		Action:       ActionRateLimitExceeded,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ResourceType: "endpoint",
		ResourceID:   endpoint,
		Result:       ResultBlocked,
	})
}

// APIClientCreated logs an API client creation event.
func (l *Logger) APIClientCreated(ctx context.Context, operatorID, clientID, ipAddress string) {
	l.LogEvent(ctx, &Entry{
		OperatorID:   operatorID,
		Action:       ActionAPIClientCreated,
		ResourceType: "api_client",
		ResourceID:   clientID,
		IPAddress:    ipAddress,
		Result:       ResultSuccess,
	})
}

// APIClientRevoked logs an API client revocation event.
func (l *Logger) APIClientRevoked(ctx context.Context, operatorID, clientID, ipAddress string) {
	l.LogEvent(ctx, &Entry{
		OperatorID:   operatorID,
		Action:       ActionAPIClientRevoked,
		ResourceType: "api_client",
		ResourceID:   clientID,
		IPAddress:    ipAddress,
		Result:       ResultSuccess,
	})
}

// SigningKeyRotated logs a signing key rotation event.
func (l *Logger) SigningKeyRotated(ctx context.Context, operatorID, clientID, ipAddress string) {
	l.LogEvent(ctx, &Entry{
		OperatorID:   operatorID,
		Action:       ActionSigningKeyRotated,
		ResourceType: "signing_key",
		ResourceID:   clientID,
		IPAddress:    ipAddress,
		Result:       ResultSuccess,
	})
}

// AdminAction logs an admin action event.
func (l *Logger) AdminAction(ctx context.Context, operatorID, action, resourceType, resourceID, ipAddress string, metadata map[string]string) {
	l.LogEvent(ctx, &Entry{
		OperatorID:   operatorID,
		Action:       ActionAdminAction,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    ipAddress,
		Result:       ResultSuccess,
		Metadata:     metadata,
	})
}

// UpdatePushed logs an update push event.
func (l *Logger) UpdatePushed(ctx context.Context, operatorID, pushID, version string, deviceCount int, ipAddress, userAgent string) {
	l.LogEvent(ctx, &Entry{
		OperatorID:   operatorID,
		Action:       ActionUpdatePushed,
		ResourceType: "update_push",
		ResourceID:   pushID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Result:       ResultSuccess,
		Metadata: map[string]string{
			"version":     version,
			"device_count": fmt.Sprintf("%d", deviceCount),
		},
	})
}

// UpdateCancelled logs an update cancellation event.
func (l *Logger) UpdateCancelled(ctx context.Context, operatorID, pushID, ipAddress, userAgent string) {
	l.LogEvent(ctx, &Entry{
		OperatorID:   operatorID,
		Action:       ActionUpdateCancelled,
		ResourceType: "update_push",
		ResourceID:   pushID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Result:       ResultSuccess,
	})
}

// UpdateSyncStarted logs an update sync start event.
func (l *Logger) UpdateSyncStarted(ctx context.Context, operatorID, ipAddress, userAgent string) {
	l.LogEvent(ctx, &Entry{
		OperatorID: operatorID,
		Action:    ActionUpdateSyncStarted,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Result:    ResultSuccess,
	})
}

// UpdateSyncFailed logs an update sync failure event.
func (l *Logger) UpdateSyncFailed(ctx context.Context, operatorID, ipAddress, userAgent, reason string) {
	l.LogEvent(ctx, &Entry{
		OperatorID: operatorID,
		Action:    ActionUpdateSyncFailed,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Result:    ResultFailure,
		Metadata:  map[string]string{"reason": reason},
	})
}
