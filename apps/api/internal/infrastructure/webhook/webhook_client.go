package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EventType represents the type of webhook event.
type EventType string

const (
	EventTypeThresholdBreach     EventType = "threshold_breach"
	EventTypeDeviceOffline      EventType = "device_offline"
	EventTypeDeviceOnline       EventType = "device_online"
	EventTypeUpdateAvailable    EventType = "update_available"
	EventTypeCommandFailed      EventType = "command_failed"
	EventTypeRegistrationRequest EventType = "registration_request"
	EventTypeError              EventType = "error"
)

// Payload represents a webhook payload.
type Payload struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	DeviceID  string                `json:"deviceId,omitempty"`
	OperatorID string                `json:"operatorId,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// TestResult represents the result of a webhook test.
type TestResult struct {
	Success     bool   `json:"success"`
	StatusCode  int    `json:"statusCode,omitempty"`
	ResponseTime int64  `json:"responseTime"`
	Error       string `json:"error,omitempty"`
	Message     string `json:"message,omitempty"`
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
			"test": true,
			"message": "This is a test webhook from Vyzorix",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &TestResult{
			Success:     false,
			Error:       "marshal_error",
			Message:     err.Error(),
			ResponseTime: time.Since(start).Milliseconds(),
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return &TestResult{
			Success:     false,
			Error:       "request_error",
			Message:     err.Error(),
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
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Vyzorix-Webhook/1.0")
	req.Header.Set("X-Vyzorix-Event", string(payload.Type))

	// Add HMAC signature if secret is provided
	if secret != "" {
		signature := computeHMAC(body, secret)
		req.Header.Set("X-Vyzorix-Signature", signature)
		req.Header.Set("X-Vyzorix-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	// Check resp before dereferencing
	if resp == nil {
		return fmt.Errorf("webhook returned nil response")
	}

	// Close body and discard contents
	if resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// computeHMAC computes HMAC-SHA256 signature of the payload.
func computeHMAC(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
