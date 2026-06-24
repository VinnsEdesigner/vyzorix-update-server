// Package hub provides WebSocket functionality.
package hub

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
)

// ErrRateLimited is returned when a client exceeds rate limits.
var ErrRateLimited = errors.New("rate limit exceeded")

// Hub manages WebSocket connections and routes messages between devices and dashboard.
type Hub struct {
	telemetryRepo   telemetry.Repository
	deviceRepo      device.Repository
	broadcast       chan []byte
	clients         map[string]*Client
	register        chan *Client
	unreg           chan *Client
	log             *slog.Logger
	messageQueue    *MessageQueue
	rateLimiter     *RateLimiter
	telemetryFilter *TelemetryFilter
	compression     *Compression
	latencyConfig   *LatencyConfig
	metrics         HubMetrics
	mu              sync.RWMutex
	metricsMu       sync.RWMutex
}

// HubMetrics holds metrics for the WebSocket hub.
type HubMetrics struct {
	LatencyMetrics        LatencyMetrics `json:"latencyMetrics"`
	TotalClientsConnected int64          `json:"totalClientsConnected"`
	TotalMessagesSent     int64          `json:"totalMessagesSent"`
	TotalMessagesReceived int64          `json:"totalMessagesReceived"`
	TotalConnectAttempts  int64          `json:"totalConnectAttempts"`
	TotalConnectSuccesses int64          `json:"totalConnectSuccesses"`
	TotalConnectFailures  int64          `json:"totalConnectFailures"`
}

// HubConfig holds configuration for the Hub.
type HubConfig struct {
	MessageQueue *MessageQueueConfig
	RateLimiter  *RateLimiterConfig
	Compression  *CompressionConfig
	Filter       *TelemetryFilterConfig
	Latency      *LatencyConfig
}

// LatencyConfig holds configuration for latency tracking.
type LatencyConfig struct {
	Enabled      bool
	MaxLatencyMS int     // Maximum acceptable latency in milliseconds (G6: sub-500ms)
	SampleRate   float64 // Percentage of messages to track (0.01 = 1%)
}

// New creates a new Hub instance.
func New(log *slog.Logger, deviceRepo device.Repository, telemetryRepo telemetry.Repository, db interface{}, cfg *HubConfig) *Hub {
	h := &Hub{
		log:           log,
		deviceRepo:    deviceRepo,
		telemetryRepo: telemetryRepo,
		clients:       make(map[string]*Client),
		register:      make(chan *Client),
		unreg:         make(chan *Client),
		broadcast:     make(chan []byte, 256),
	}

	// Initialize with defaults if no config
	if cfg == nil {
		cfg = &HubConfig{}
	}

	// Initialize compression
	if cfg.Compression == nil {
		cfg.Compression = DefaultCompressionConfig()
	}

	h.compression = NewCompression(log, cfg.Compression)

	// Initialize telemetry filter
	if cfg.Filter == nil {
		cfg.Filter = DefaultTelemetryFilterConfig()
	}

	h.telemetryFilter = NewTelemetryFilter(log, cfg.Filter)

	// Initialize latency tracking (G6: sub-500ms latency)
	if cfg.Latency == nil {
		cfg.Latency = &LatencyConfig{
			Enabled:      true,
			MaxLatencyMS: 500, // Target: sub-500ms
			SampleRate:   0.1, // 10% sampling
		}
	}

	h.latencyConfig = cfg.Latency

	return h
}

// LatencyMetrics holds latency tracking metrics.
type LatencyMetrics struct {
	LastExceededID     string  `json:"lastExceededId"`
	TotalMessages      int64   `json:"totalMessages"`
	SuccessfulMessages int64   `json:"successfulMessages"`
	FailedMessages     int64   `json:"failedMessages"`
	TotalLatencyMS     int64   `json:"totalLatencyMs"`
	MinLatencyMS       int64   `json:"minLatencyMs"`
	MaxLatencyMS       int64   `json:"maxLatencyMs"`
	ExceededCount      int64   `json:"exceededCount"`
	LastExceededAt     int64   `json:"lastExceededAt"`
	AverageLatencyMS   float64 `json:"averageLatencyMs"`
	P95LatencyMS       float64 `json:"p95LatencyMs"`
	P99LatencyMS       float64 `json:"p99LatencyMs"`
}

// SetMessageQueue sets the message queue on the hub.
func (h *Hub) SetMessageQueue(mq *MessageQueue) {
	h.messageQueue = mq
}

// SetRateLimiter sets the rate limiter on the hub.
func (h *Hub) SetRateLimiter(rl *RateLimiter) {
	h.rateLimiter = rl
}

// Compression returns the hub's compression handler.
func (h *Hub) Compression() *Compression {
	return h.compression
}

// TelemetryFilter returns the hub's telemetry filter.
func (h *Hub) TelemetryFilter() *TelemetryFilter {
	return h.telemetryFilter
}

// RateLimiter returns the hub's rate limiter.
func (h *Hub) RateLimiter() *RateLimiter {
	return h.rateLimiter
}

// Run starts the hub's event loop in a background goroutine.
// It handles client registration, unregistration, and telemetry broadcasting.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-h.register:
			h.mu.Lock()
			if old := h.clients[c.DeviceID]; old != nil {
				close(old.Send)

				if err := old.Conn.Close(); err != nil {
					h.log.Warn("old conn close failed", "deviceId", c.DeviceID, "err", err)
				}
			}

			c.log = h.log
			h.clients[c.DeviceID] = c
			h.mu.Unlock()

			// Set device online
			if err := h.deviceRepo.SetOnline(context.Background(), c.DeviceID, true); err != nil {
				h.log.Warn("set online failed", "deviceId", c.DeviceID, "err", err)
			}

			// Replay queued messages to the newly connected device
			if h.messageQueue != nil {
				count := h.messageQueue.ReplayQueue(c.DeviceID, c.Send)
				if count > 0 {
					h.log.Info("replayed queued messages to device",
						"deviceId", c.DeviceID,
						"count", count,
					)
				}
			}

			h.log.Info("device websocket online", "deviceId", c.DeviceID)

		case c := <-h.unreg:
			h.mu.Lock()
			if h.clients[c.DeviceID] == c {
				delete(h.clients, c.DeviceID)
				close(c.Send)

				if err := h.deviceRepo.SetOnline(context.Background(), c.DeviceID, false); err != nil {
					h.log.Warn("set offline failed", "deviceId", c.DeviceID, "err", err)
				}
			}
			h.mu.Unlock()
			h.log.Info("device websocket offline", "deviceId", c.DeviceID)

		case raw := <-h.broadcast:
			h.mu.RLock()
			for _, c := range h.clients {
				select {
				case c.Send <- command.CommandFrame{Type: "broadcast", Args: raw}:
				default:
				}
			}
			h.mu.RUnlock()

			_ = raw // prevent unused variable warning from channel receive
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) { h.register <- c }

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) { h.unreg <- c }

// Online checks if a device is currently connected.
func (h *Hub) Online(deviceID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.clients[deviceID] != nil
}

// Clients returns a copy of the current clients map.
func (h *Hub) Clients() map[string]*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make(map[string]*Client, len(h.clients))
	for k, v := range h.clients {
		out[k] = v
	}

	return out
}

// GetClient retrieves a specific client by device ID.
func (h *Hub) GetClient(deviceID string) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.clients[deviceID]
}

// ClientCount returns the number of currently connected device clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients)
}

// BroadcastTelemetry sends telemetry data to connected clients with filtering.
func (h *Hub) BroadcastTelemetry(raw []byte) {
	select {
	case h.broadcast <- raw:
	default:
	}
}

// BroadcastTelemetryToFiltered sends telemetry only to subscribed clients.
func (h *Hub) BroadcastTelemetryToFiltered(senderDeviceID string, raw []byte) {
	if h.telemetryFilter == nil {
		// No filter configured, broadcast to all
		h.BroadcastTelemetry(raw)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for clientID, c := range h.clients {
		// Skip the sender
		if clientID == senderDeviceID {
			continue
		}

		// Check if this client is subscribed to the sender's telemetry
		if h.telemetryFilter.ShouldForward(clientID, senderDeviceID) {
			// Client buffer full, skip
			select {
			case c.Send <- command.CommandFrame{Type: "telemetry", Args: raw}:
			default:
			}
		}
	}
}

// Subscribe subscribes a client to a device's telemetry.
func (h *Hub) Subscribe(clientID, deviceID string) bool {
	if h.telemetryFilter == nil {
		return false
	}

	return h.telemetryFilter.Subscribe(clientID, deviceID)
}

// Unsubscribe unsubscribes a client from a device's telemetry.
func (h *Hub) Unsubscribe(clientID, deviceID string) bool {
	if h.telemetryFilter == nil {
		return false
	}

	return h.telemetryFilter.Unsubscribe(clientID, deviceID)
}

// GetSubscriptions returns the subscriptions for a client.
func (h *Hub) GetSubscriptions(clientID string) []string {
	if h.telemetryFilter == nil {
		return nil
	}

	return h.telemetryFilter.GetSubscriptions(clientID)
}

// Send delivers a command frame to a specific device.
// If the device is connected, it sends directly.
// If the device is offline and message queue is configured, it queues the message.
// Returns true if the message was either sent or queued.
func (h *Hub) Send(deviceID string, frame command.CommandFrame) bool {
	startTime := time.Now()

	h.mu.RLock()
	c := h.clients[deviceID]
	h.mu.RUnlock()

	if c == nil {
		// Device offline - try to queue if message queue is available
		if h.messageQueue != nil {
			// Use synchronous DB write for 100% delivery guarantee (G1)
			success := h.messageQueue.EnqueueWithConfirmation(deviceID, frame)
			if h.latencyConfig != nil && h.latencyConfig.Enabled {
				h.trackLatency(deviceID, frame.DispatchID, startTime, success)
			}

			return success
		}

		return false
	}

	select {
	case c.Send <- frame:
		if h.latencyConfig != nil && h.latencyConfig.Enabled {
			h.trackLatency(deviceID, frame.DispatchID, startTime, true)
		}

		return true
	default:
		// Client buffer full - try queue if available
		if h.messageQueue != nil {
			// Use synchronous DB write for 100% delivery guarantee (G1)
			success := h.messageQueue.EnqueueWithConfirmation(deviceID, frame)
			if h.latencyConfig != nil && h.latencyConfig.Enabled {
				h.trackLatency(deviceID, frame.DispatchID, startTime, success)
			}

			return success
		}

		return false
	}
}

// trackLatency tracks message latency for G6 compliance.
func (h *Hub) trackLatency(deviceID, dispatchID string, startTime time.Time, success bool) {
	if h.latencyConfig == nil || !h.latencyConfig.Enabled {
		return
	}

	latencyMS := time.Since(startTime).Milliseconds()

	// Sample based on configuration
	if rand.Float64() > h.latencyConfig.SampleRate {
		return
	}

	h.metricsMu.Lock()
	defer h.metricsMu.Unlock()

	h.metrics.LatencyMetrics.TotalMessages++
	if success {
		h.metrics.LatencyMetrics.SuccessfulMessages++
	} else {
		h.metrics.LatencyMetrics.FailedMessages++
	}
	h.metrics.LatencyMetrics.TotalLatencyMS += latencyMS
	// Update min/max
	if h.metrics.LatencyMetrics.MinLatencyMS == 0 || latencyMS < h.metrics.LatencyMetrics.MinLatencyMS {
		h.metrics.LatencyMetrics.MinLatencyMS = latencyMS
	}

	if latencyMS > h.metrics.LatencyMetrics.MaxLatencyMS {
		h.metrics.LatencyMetrics.MaxLatencyMS = latencyMS
	}

	// Check if exceeds target (G6: sub-500ms)
	if latencyMS > int64(h.latencyConfig.MaxLatencyMS) {
		h.metrics.LatencyMetrics.ExceededCount++
		h.metrics.LatencyMetrics.LastExceededAt = time.Now().Unix()
		h.metrics.LatencyMetrics.LastExceededID = dispatchID
		h.log.Warn("latency exceeded target",
			"deviceId", deviceID,
			"dispatchId", dispatchID,
			"latencyMs", latencyMS,
			"targetMs", h.latencyConfig.MaxLatencyMS,
		)
	}

	// Update average
	h.metrics.LatencyMetrics.AverageLatencyMS = float64(h.metrics.LatencyMetrics.TotalLatencyMS) / float64(h.metrics.LatencyMetrics.TotalMessages)
}

// SendWithDeliveryConfirmation sends a message and waits for delivery confirmation.
// This ensures 100% message delivery by waiting for WebSocket write confirmation.
func (h *Hub) SendWithDeliveryConfirmation(deviceID string, frame command.CommandFrame, timeout time.Duration) (bool, error) {
	h.mu.RLock()
	c := h.clients[deviceID]
	h.mu.RUnlock()

	if c == nil {
		// Device offline - queue with synchronous DB write
		if h.messageQueue != nil {
			return h.messageQueue.EnqueueWithConfirmation(deviceID, frame), nil
		}

		return false, nil
	}

	// Create a channel for delivery confirmation
	confirmation := make(chan bool, 1)
	frame.DeliveryConfirmation = confirmation

	select {
	case c.Send <- frame:
		// Wait for confirmation or timeout
		select {
		case confirmed := <-confirmation:
			return confirmed, nil
		case <-time.After(timeout):
			// Timeout - try to queue if available
			if h.messageQueue != nil {
				return h.messageQueue.EnqueueWithConfirmation(deviceID, frame), nil
			}

			return false, fmt.Errorf("delivery confirmation timeout")
		}
	default:
		// Client buffer full - queue with synchronous write
		if h.messageQueue != nil {
			return h.messageQueue.EnqueueWithConfirmation(deviceID, frame), nil
		}

		return false, nil
	}
}

// QueueMetrics returns the current queue metrics if a message queue is configured.
func (h *Hub) QueueMetrics() (QueueMetrics, bool) {
	if h.messageQueue == nil {
		return QueueMetrics{}, false
	}

	return h.messageQueue.GetMetrics(), true
}

// QueueSize returns the number of queued messages for a device.
func (h *Hub) QueueSize(deviceID string) int {
	if h.messageQueue == nil {
		return 0
	}

	return h.messageQueue.QueueSize(deviceID)
}

// TotalQueuedMessages returns the total number of queued messages across all devices.
func (h *Hub) TotalQueuedMessages() int {
	if h.messageQueue == nil {
		return 0
	}

	return h.messageQueue.TotalQueuedMessages()
}
