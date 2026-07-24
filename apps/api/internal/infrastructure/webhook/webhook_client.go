package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"
)

// ErrPrivateOrInternalIP indicates the URL resolves to a private/internal IP.
var ErrPrivateOrInternalIP = errors.New("webhook URL cannot resolve to a private or internal IP address")

const (
	maxRetries     = 3
	baseRetryDelay = 1 * time.Second
)

// ValidateURL validates a webhook URL to prevent SSRF attacks.
// Returns nil if valid, or ErrPrivateOrInternalIP if the URL resolves to a private/internal IP.
// Requires HTTPS scheme.
func ValidateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return errors.New("invalid URL format")
	}

	if parsedURL.Scheme != "https" {
		return errors.New("webhook URL must use HTTPS")
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return errors.New("webhook URL must have a valid host")
	}

	// Check if hostname itself is an IP.
	if ip := net.ParseIP(hostname); ip != nil {
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			return ErrPrivateOrInternalIP
		}
	}

	// Resolve hostname and check all IPs.
	ips, err := net.LookupIP(hostname)
	if err != nil {
		// DNS failure will fail at connection time anyway, so we skip the IP check.
		//nolint:nilerr // Intentional: connection will fail anyway if DNS is broken.
		return nil
	}

	for _, ip := range ips {
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			return ErrPrivateOrInternalIP
		}
	}

	return nil
}

// EventType represents the type of webhook event.
type EventType string

const (
	EventTypeThresholdBreach     EventType = "threshold_breach"
	EventTypeDeviceOffline       EventType = "device_offline"
	EventTypeDeviceOnline        EventType = "device_online"
	EventTypeUpdateAvailable     EventType = "update_available"
	EventTypeCommandFailed       EventType = "command_failed"
	EventTypeRegistrationRequest EventType = "registration_request"
	EventTypeError               EventType = "error"
)

// Payload represents a webhook payload.
type Payload struct {
	Timestamp  time.Time              `json:"timestamp"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Type       EventType              `json:"type"`
	DeviceID   string                 `json:"deviceId,omitempty"`
	OperatorID string                 `json:"operatorId,omitempty"`
}

// TestResult represents the result of a webhook test.
type TestResult struct {
	Error        string `json:"error,omitempty"`
	Message      string `json:"message,omitempty"`
	StatusCode   int    `json:"statusCode,omitempty"`
	ResponseTime int64  `json:"responseTime"`
	Success      bool   `json:"success"`
}

// Client is a webhook client for sending notifications.
type Client struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewClient creates a new webhook client.
func NewClient(timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		timeout:    timeout,
	}
}

// Test sends a test ping to a webhook URL.
func (c *Client) Test(ctx context.Context, url string) (*TestResult, error) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	payload := Payload{
		Type:      EventTypeThresholdBreach,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"test":    true,
			"message": "This is a test webhook from Vyzorix",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &TestResult{
			Success:      false,
			Error:        "marshal_error",
			Message:      err.Error(),
			ResponseTime: time.Since(start).Milliseconds(),
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return &TestResult{
			Success:      false,
			Error:        "request_error",
			Message:      err.Error(),
			ResponseTime: time.Since(start).Milliseconds(),
		}, nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Vyzorix-Webhook/1.0")

	resp, err := c.httpClient.Do(req)
	responseTime := time.Since(start).Milliseconds()

	if err != nil {
		return &TestResult{
			Success:      false,
			Error:        "webhook_timeout",
			Message:      fmt.Sprintf("Webhook did not respond within %s", c.timeout),
			ResponseTime: responseTime,
		}, nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &TestResult{
			Success:      true,
			StatusCode:   resp.StatusCode,
			ResponseTime: responseTime,
		}, nil
	}

	return &TestResult{
		Success:      false,
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
		Error:        "webhook_failed",
		Message:      fmt.Sprintf("Webhook returned status %d", resp.StatusCode),
	}, nil
}

// Send sends a webhook notification with HMAC signature.

func (c *Client) Send(ctx context.Context, url, secret string, payload *Payload) error {
	// Reject private/internal IPs to prevent SSRF.
	if err := ValidateURL(url); err != nil {
		return fmt.Errorf("webhook URL rejected: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s.
			delay := baseRetryDelay * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := c.doSend(ctx, url, secret, body, payload)
		if err == nil {
			return nil
		}

		// Check if error is retryable (5xx status code).
		if !isRetryableError(err) {
			return err
		}
		lastErr = err

		// Log retry attempt.
		if slog.Default().Enabled(ctx, slog.LevelWarn) {
			slog.Default().Warn("webhook delivery failed, retrying",
				"attempt", attempt+1,
				"maxRetries", maxRetries,
				"error", err)
		}
	}

	return fmt.Errorf("webhook delivery failed after %d attempts: %w", maxRetries, lastErr)
}

// doSend performs a single webhook delivery attempt.
func (c *Client) doSend(ctx context.Context, url, secret string, body []byte, payload *Payload) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Vyzorix-Webhook/1.0")
	req.Header.Set("X-Vyzorix-Event", string(payload.Type))

	// Add HMAC signature if secret is provided.
	if secret != "" {
		signature := computeHMAC(body, secret)
		req.Header.Set("X-Vyzorix-Signature", signature)
		req.Header.Set("X-Vyzorix-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	// Check resp before dereferencing.
	if resp == nil {
		return fmt.Errorf("webhook returned nil response")
	}

	// Close body and discard contents.
	if resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &retryableError{statusCode: resp.StatusCode, message: fmt.Sprintf("webhook returned status %d", resp.StatusCode)}
	}

	return nil
}

// retryableError represents an error that can be retried.
type retryableError struct {
	message    string
	statusCode int
}

func (e *retryableError) Error() string {
	return e.message
}

// isRetryableError returns true if the error is a transient error that should be retried.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// Check for retryable error type.
	if re, ok := err.(*retryableError); ok {
		// Only retry 5xx errors (server errors), not 4xx (client errors).
		return re.statusCode >= 500 && re.statusCode < 600
	}
	// Retry on context deadline exceeded (might succeed on retry).
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// Retry on temporary network errors.
	return false
}

// computeHMAC computes HMAC-SHA256 signature of the payload.
func computeHMAC(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
