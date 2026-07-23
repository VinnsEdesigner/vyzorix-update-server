// Package fcm provides Firebase Cloud Messaging integration.
package fcm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"firebase.google.com/go/v4/messaging"
)

// ErrUnavailable indicates FCM service is temporarily unavailable.
var ErrUnavailable = errors.New("fcm: temporarily unavailable")

// FCMConfig holds configuration for FCM behavior.
type FCMConfig struct {
	// MaxRetries is the maximum number of retry attempts (default 3)
	MaxRetries int
	// BaseRetryDelay is the base delay for exponential backoff (default 1 second)
	BaseRetryDelay time.Duration
	// TokenValidationEnabled enables FCM token validation before sending
	TokenValidationEnabled bool
}

// DefaultFCMConfig returns the default FCM configuration.
func DefaultFCMConfig() *FCMConfig {
	return &FCMConfig{
		MaxRetries:             3,
		BaseRetryDelay:         1 * time.Second,
		TokenValidationEnabled: true,
	}
}

// FCMMetrics holds FCM delivery metrics.
type FCMMetrics struct {
	LastSuccessID   string `json:"lastSuccessId"`
	LastFailureID   string `json:"lastFailureId"`
	SuccessCount    int64  `json:"successCount"`
	FailureCount    int64  `json:"failureCount"`
	RetryCount      int64  `json:"retryCount"`
	TokenErrorCount int64  `json:"tokenErrorCount"`
	LastSuccessAt   int64  `json:"lastSuccessAt"`
	LastFailureAt   int64  `json:"lastFailureAt"`
}

type SilentWake struct {
	Token         string
	Command       string
	DispatchID    string
	DeviceID      string
	Priority      string
	CommandSecret string
	APKFilename   string
	SHA256        string
	DownloadURL   string
	APKSize       int64
}

type Notifier interface {
	SendSilentWake(ctx context.Context, wake SilentWake) error
	GetMetrics() FCMMetrics
}

// SafeNotifier wraps a Notifier with graceful degradation and circuit breaker.
// If FCM fails, it logs the error but doesn't propagate it,
// allowing the service to continue operating.
// Includes circuit breaker to prevent cascading failures.
type SafeNotifier struct {
	Notifier       Notifier
	circuitBreaker *CircuitBreaker
}

// NewSafeNotifier creates a SafeNotifier with optional circuit breaker.
func NewSafeNotifier(notifier Notifier) *SafeNotifier {
	return &SafeNotifier{
		Notifier:       notifier,
		circuitBreaker: NewCircuitBreaker(DefaultCircuitBreakerConfig()),
	}
}

// SendSilentWake attempts to send via FCM with circuit breaker protection.
// Returns nil if FCM is disabled or fails, allowing the service to continue.
func (s *SafeNotifier) SendSilentWake(ctx context.Context, wake SilentWake) error {
	if s.Notifier == nil {
		return nil // Graceful degradation: no notifier configured
	}

	// Check circuit breaker before attempting FCM call
	if s.circuitBreaker != nil && !s.circuitBreaker.Allow() {
		s.Notifier.GetMetrics() // Trigger any side effects
		return nil              // Fail fast but gracefully
	}

	err := s.Notifier.SendSilentWake(ctx, wake)
	if err != nil {
		// Record failure in circuit breaker
		if s.circuitBreaker != nil {
			s.circuitBreaker.RecordFailure()
		}

		if errors.Is(err, ErrDisabled) {
			// Not an error - FCM is intentionally disabled.
			return nil
		}
		
		slog.Warn("fcm: failed to send silent wake",
			"deviceToken", wake.Token,
			"deviceID", wake.DeviceID,
			"command", wake.Command,
			"dispatchID", wake.DispatchID,
			"error", err,
		)
		return nil
	}

	// Record success in circuit breaker
	if s.circuitBreaker != nil {
		s.circuitBreaker.RecordSuccess()
	}

	return nil
}

// CircuitBreakerState returns the current state of the circuit breaker.
func (s *SafeNotifier) CircuitBreakerState() string {
	if s.circuitBreaker == nil {
		return "disabled"
	}
	return s.circuitBreaker.State().String()
}

// EnhancedNotifier wraps Client with retry logic, metrics, and token validation.

type EnhancedNotifier struct {
	*Client
	config    *FCMConfig
	metrics   FCMMetrics
	metricsMu sync.RWMutex
	db        *sql.DB
}

// NewEnhancedNotifier creates a new enhanced FCM notifier with retry and metrics.
func NewEnhancedNotifier(client *Client, cfg *FCMConfig) *EnhancedNotifier {
	if cfg == nil {
		cfg = DefaultFCMConfig()
	}

	return &EnhancedNotifier{
		Client: client,
		config: cfg,
	}
}

// NewEnhancedNotifierWithDB creates a new enhanced FCM notifier with persistent retry support.
func NewEnhancedNotifierWithDB(client *Client, cfg *FCMConfig, db *sql.DB) *EnhancedNotifier {
	notifier := NewEnhancedNotifier(client, cfg)
	notifier.db = db
	return notifier
}

// SendSilentWake sends a silent wake notification with retry logic and metrics tracking.

// table for the background worker to retry later.
func (e *EnhancedNotifier) SendSilentWake(ctx context.Context, wake SilentWake) error {
	if e == nil || !e.Enabled() {
		return ErrDisabled
	}

	// Validate token before attempting send
	if e.config.TokenValidationEnabled {
		if !validateFCMToken(wake.Token) {
			e.incrementTokenError()
			e.log.Warn("fcm invalid token",
				"deviceId", wake.DeviceID,
				"dispatchId", wake.DispatchID,
			)

			return fmt.Errorf("invalid fcm token for device %s", wake.DeviceID)
		}
	}

	client := e.Messaging()
	if client == nil {
		return ErrUnavailable
	}

	// Determine message priority
	priority := "high"
	if wake.Priority == "normal" {
		priority = "normal"
	}

	var lastErr error

	for attempt := 1; attempt <= e.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Build FCM message
		msg := &messaging.Message{
			Token: wake.Token,
			Android: &messaging.AndroidConfig{
				Priority: priority,
				TTL:      ptr24Hours(),
				Data:     buildFCMData(wake),
			},
			Data: buildFCMData(wake),
		}

		result, err := client.Send(ctx, msg)
		if err == nil {
			e.incrementSuccess()
			e.setLastSuccess(wake.DispatchID)
			e.log.Info("fcm silent wake sent",
				"deviceId", wake.DeviceID,
				"dispatchId", wake.DispatchID,
				"messageId", result,
				"attempt", attempt,
			)

			return nil
		}

		lastErr = err

		e.incrementRetry()

		// Log failure
		e.log.Warn("fcm send failed",
			"deviceId", wake.DeviceID,
			"dispatchId", wake.DispatchID,
			"attempt", attempt,
			"maxRetries", e.config.MaxRetries,
			"err", err,
		)

		// Don't retry on last attempt
		if attempt == e.config.MaxRetries {
			break
		}

		// Exponential backoff: 1s, 4s, 9s (attempt^2 seconds)
		delay := e.config.BaseRetryDelay * time.Duration(attempt*attempt)
		e.log.Debug("fcm retrying after backoff",
			"deviceId", wake.DeviceID,
			"attempt", attempt,
			"delay", delay,
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	// All retries exhausted
	e.incrementFailure()
	e.setLastFailure(wake.DispatchID)
	e.log.Error("fcm send failed after all retries",
		"deviceId", wake.DeviceID,
		"dispatchId", wake.DispatchID,
		"totalAttempts", e.config.MaxRetries,
		"lastErr", lastErr,
	)

	
	if e.db != nil {
		if persistErr := e.persistForRetry(ctx, wake, lastErr); persistErr != nil {
			e.log.Warn("fcm: failed to persist for retry",
				"deviceId", wake.DeviceID,
				"dispatchId", wake.DispatchID,
				"error", persistErr,
			)
		} else {
			e.log.Info("fcm: persisted for background retry",
				"deviceId", wake.DeviceID,
				"dispatchId", wake.DispatchID,
			)
			// Return nil since we're persisting for retry - caller doesn't need to know
			return nil
		}
	}

	return fmt.Errorf("fcm send failed after %d attempts: %w", e.config.MaxRetries, lastErr)
}

// persistForRetry saves a failed notification to the pending_fcm table for background retry.
func (e *EnhancedNotifier) persistForRetry(ctx context.Context, wake SilentWake, lastErr error) error {
	query := `
		INSERT INTO pending_fcm (dispatch_id, device_id, token, command, priority, retry_count, next_retry_at, last_error, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?, ?)
	`
	now := time.Now().UnixMilli()
	// First retry in 1 minute
	nextRetry := now + int64(time.Minute/time.Millisecond)

	_, err := e.db.ExecContext(ctx, query,
		wake.DispatchID,
		wake.DeviceID,
		wake.Token,
		wake.Command,
		wake.Priority,
		nextRetry,
		lastErr.Error(),
		now,
		now,
	)
	return err
}

// TopicMessage represents a message to be sent to an FCM topic (FR-8: Topic Messaging Support).
type TopicMessage struct {
	Topic      string
	Command    string
	DispatchID string
	Priority   string // "high" or "normal"
}

// SendToTopic sends a message to an FCM topic.
// Topic messaging allows sending to multiple devices subscribed to the same topic.
func (e *EnhancedNotifier) SendToTopic(ctx context.Context, msg TopicMessage) error {
	if e == nil || !e.Enabled() {
		return ErrDisabled
	}

	if msg.Topic == "" {
		return fmt.Errorf("topic is required for topic messaging")
	}

	client := e.Messaging()
	if client == nil {
		return ErrUnavailable
	}

	// Determine priority (FR-8.5: Message Prioritization)
	priority := "high"
	if msg.Priority == "normal" {
		priority = "normal"
	}

	fcmMsg := &messaging.Message{
		Topic: msg.Topic,
		Android: &messaging.AndroidConfig{
			Priority: priority,
			TTL:      ptr24Hours(),
			Data: map[string]string{
				"action":      "WAKE_DAEMON",
				"command":     msg.Command,
				"dispatch_id": msg.DispatchID,
			},
		},
		Data: map[string]string{
			"action":      "WAKE_DAEMON",
			"command":     msg.Command,
			"dispatch_id": msg.DispatchID,
		},
	}

	result, err := client.Send(ctx, fcmMsg)
	if err != nil {
		e.incrementFailure()
		e.log.Warn("fcm topic send failed",
			"topic", msg.Topic,
			"dispatchId", msg.DispatchID,
			"err", err,
		)

		return fmt.Errorf("fcm topic send: %w", err)
	}

	e.incrementSuccess()
	e.log.Info("fcm topic message sent",
		"topic", msg.Topic,
		"dispatchId", msg.DispatchID,
		"messageId", result,
	)

	return nil
}

// GetMetrics returns a copy of the current FCM metrics.
func (e *EnhancedNotifier) GetMetrics() FCMMetrics {
	e.metricsMu.RLock()
	defer e.metricsMu.RUnlock()

	return e.metrics
}

func (e *EnhancedNotifier) incrementSuccess() {
	atomic.AddInt64(&e.metrics.SuccessCount, 1)
}

func (e *EnhancedNotifier) incrementFailure() {
	atomic.AddInt64(&e.metrics.FailureCount, 1)
}

func (e *EnhancedNotifier) incrementRetry() {
	atomic.AddInt64(&e.metrics.RetryCount, 1)
}

func (e *EnhancedNotifier) incrementTokenError() {
	atomic.AddInt64(&e.metrics.TokenErrorCount, 1)
}

func (e *EnhancedNotifier) setLastSuccess(dispatchID string) {
	e.metricsMu.Lock()
	e.metrics.LastSuccessAt = time.Now().Unix()
	e.metrics.LastSuccessID = dispatchID
	e.metricsMu.Unlock()
}

func (e *EnhancedNotifier) setLastFailure(dispatchID string) {
	e.metricsMu.Lock()
	e.metrics.LastFailureAt = time.Now().Unix()
	e.metrics.LastFailureID = dispatchID
	e.metricsMu.Unlock()
}

// validateFCMToken validates the format of an FCM token.
func validateFCMToken(token string) bool {
	if token == "" {
		return false
	}

	// FCM tokens are typically 150-200 characters
	if len(token) < 50 || len(token) > 500 {
		return false
	}

	// FCM tokens are base64-like (alphanumeric with -_=)
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_="
	for _, c := range token {
		if !strings.Contains(validChars, string(c)) {
			return false
		}
	}

	return true
}

// SendSilentWake sends a silent wake notification without retry.
func (c *Client) SendSilentWake(ctx context.Context, wake SilentWake) error {
	if c == nil {
		return ErrDisabled
	}

	if !c.enabled {
		return ErrDisabled
	}

	if wake.Token == "" {
		return fmt.Errorf("missing fcm token for device %s", wake.DeviceID)
	}

	client := c.Messaging()
	if client == nil {
		return ErrUnavailable
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	msg := &messaging.Message{
		Token: wake.Token,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			TTL:      ptr24Hours(),
			Data:     buildFCMData(wake),
		},
		Data: buildFCMData(wake),
	}

	result, err := client.Send(ctx, msg)
	if err != nil {
		c.log.Warn("fcm send failed",
			"deviceId", wake.DeviceID,
			"dispatchId", wake.DispatchID,
			"err", err)

		return fmt.Errorf("fcm send: %w", err)
	}

	c.log.Info("fcm silent wake sent", "deviceId", wake.DeviceID, "dispatchId", wake.DispatchID, "messageId", result)

	return nil
}

func ptr24Hours() *time.Duration {
	d := 24 * time.Hour
	return &d
}

// buildFCMData constructs the data payload for FCM messages.
// Includes APK info for update commands and CommandSecret for registration.
func buildFCMData(wake SilentWake) map[string]string {
	data := map[string]string{
		"action":      "WAKE_DAEMON",
		"command":     wake.Command,
		"dispatch_id": wake.DispatchID,
		"device_id":   wake.DeviceID,
	}

	// Include CommandSecret for registration approval
	if wake.CommandSecret != "" {
		data["command_secret"] = wake.CommandSecret
	}

	// Include APK info for update commands
	if wake.APKFilename != "" {
		data["apkFilename"] = wake.APKFilename
		data["sha256"] = wake.SHA256
		data["apkSize"] = strconv.FormatInt(wake.APKSize, 10)
		// Include download URL if provided, otherwise device can construct from filename
		if wake.DownloadURL != "" {
			data["downloadUrl"] = wake.DownloadURL
		}
	}

	return data
}
