// Package hub provides WebSocket functionality.
package hub

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
)

// ErrRateLimited is returned when a client exceeds rate limits.
var ErrRateLimited = errors.New("rate limit exceeded")

// EventProcessor handles real-time event processing.
type EventProcessor interface {
	ProcessDeviceConnected(ctx context.Context, deviceID string, metadata map[string]any) error
	ProcessDeviceDisconnected(ctx context.Context, deviceID string, reason string, metadata map[string]any) error
	ProcessTelemetry(ctx context.Context, deviceID string, telemetryData map[string]any) error
}

// DashboardBroadcaster sends events to dashboard clients.
type DashboardBroadcaster interface {
	BroadcastToDashboard(deviceID string, operatorID string, eventType string, data []byte) error
}

// Hub manages WebSocket connections and routes messages between devices and dashboard.
type Hub struct {
	telemetryRepo        telemetry.Repository
	deviceRepo           device.Repository
	eventProcessor       EventProcessor
	dashboardBroadcaster DashboardBroadcaster
	broadcast            chan []byte
	clients              map[string]*Client
	register             chan *Client
	unreg                chan *Client
	deviceStatus         chan DeviceStatusUpdate
	log                  *slog.Logger
	messageQueue         *MessageQueue
	rateLimiter          *RateLimiter
	telemetryFilter      *TelemetryFilter
	compression          *Compression
	latencyConfig        *LatencyConfig
	metrics              HubMetrics
	deviceLatency        map[string]*LatencyMetrics
	mu                   sync.RWMutex
	metricsMu            sync.RWMutex
}

// DeviceStatusUpdate represents a device online/offline status change.

type DeviceStatusUpdate struct {
	DeviceID string
	Source   string
	Online   bool
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
	MaxLatencyMS int     // Maximum acceptable latency in milliseconds (G6: sub-500ms).
	SampleRate   float64 // Percentage of messages to track (0.01 = 1%).
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
		deviceStatus:  make(chan DeviceStatusUpdate, 256),
		broadcast:     make(chan []byte, 256),
		deviceLatency: make(map[string]*LatencyMetrics),
	}

	// Initialize with defaults if no config.
	if cfg == nil {
		cfg = &HubConfig{}
	}

	// Initialize compression.
	if cfg.Compression == nil {
		cfg.Compression = DefaultCompressionConfig()
	}

	h.compression = NewCompression(log, cfg.Compression)

	// Initialize telemetry filter.
	if cfg.Filter == nil {
		cfg.Filter = DefaultTelemetryFilterConfig()
	}

	h.telemetryFilter = NewTelemetryFilter(log, cfg.Filter)

	// Initialize latency tracking.
	if cfg.Latency == nil {
		cfg.Latency = &LatencyConfig{
			Enabled:      true,
			MaxLatencyMS: 500, // Informational target (not enforced).
			SampleRate:   0.1, // 10% sampling.
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

// SetEventProcessor sets the event processor on the hub.
func (h *Hub) SetEventProcessor(ep EventProcessor) {
	h.eventProcessor = ep
}

// SetDashboardBroadcaster sets the dashboard broadcaster on the hub.
func (h *Hub) SetDashboardBroadcaster(db DashboardBroadcaster) {
	h.dashboardBroadcaster = db
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
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("hub panic recovered, restarting run loop", "panic", r)
			go h.Run(ctx)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case c := <-h.register:
			h.handleClientRegistration(ctx, c)
		case c := <-h.unreg:
			h.handleClientUnregistration(ctx, c)
		case raw := <-h.broadcast:
			h.handleBroadcast(raw)
		case status := <-h.deviceStatus:
			h.handleDeviceStatus(ctx, status)
		}
	}
}

// handleClientRegistration handles new client registration.
func (h *Hub) handleClientRegistration(ctx context.Context, c *Client) {
	h.mu.Lock()
	if old := h.clients[c.DeviceID]; old != nil {
		close(old.Send)
		if old.Conn != nil {
			if err := old.Conn.Close(); err != nil {
				h.log.Warn("old conn close failed", "deviceId", c.DeviceID, "err", err)
			}
		}
	}

	c.log = h.log
	h.clients[c.DeviceID] = c
	h.mu.Unlock()

	h.deviceStatus <- DeviceStatusUpdate{DeviceID: c.DeviceID, Online: true, Source: "websocket"}
	h.replayQueuedMessages(c)
	h.emitDeviceConnected(ctx, c)
	h.log.Info("device websocket online", "deviceId", c.DeviceID)
}

// replayQueuedMessages replays queued messages to a newly connected device.
func (h *Hub) replayQueuedMessages(c *Client) {
	if h.messageQueue == nil {
		return
	}
	result := h.messageQueue.ReplayQueue(c.DeviceID, c.Send)
	if result.Count > 0 {
		h.log.Info("replayed queued messages to device", "deviceId", c.DeviceID, "count", result.Count)
	}
	if result.HasMore {
		h.schedulePartialReplay(c.DeviceID, result.Remaining)
	}
}

// emitDeviceConnected emits the device connected event.
func (h *Hub) emitDeviceConnected(ctx context.Context, c *Client) {
	if h.eventProcessor == nil {
		return
	}
	metadata := map[string]any{
		"clientIP":  c.Conn.RemoteAddr().String(),
		"timestamp": time.Now().UnixMilli(),
	}
	if err := h.eventProcessor.ProcessDeviceConnected(ctx, c.DeviceID, metadata); err != nil {
		h.log.Warn("failed to emit device connected event", "deviceId", c.DeviceID, "err", err)
	}
}

// handleClientUnregistration handles client disconnection.
func (h *Hub) handleClientUnregistration(ctx context.Context, c *Client) {
	h.mu.Lock()
	if h.clients[c.DeviceID] == c {
		delete(h.clients, c.DeviceID)
		close(c.Send)
		h.emitDeviceDisconnected(ctx, c)
	}
	h.mu.Unlock()

	h.deviceStatus <- DeviceStatusUpdate{DeviceID: c.DeviceID, Online: false, Source: "websocket"}
	h.log.Info("device websocket offline", "deviceId", c.DeviceID)
}

// emitDeviceDisconnected emits the device disconnected event.
func (h *Hub) emitDeviceDisconnected(ctx context.Context, c *Client) {
	if h.eventProcessor == nil {
		return
	}
	metadata := map[string]any{
		"timestamp": time.Now().UnixMilli(),
	}
	if err := h.eventProcessor.ProcessDeviceDisconnected(ctx, c.DeviceID, "client_disconnect", metadata); err != nil {
		h.log.Warn("failed to emit device disconnected event", "deviceId", c.DeviceID, "err", err)
	}
}

// handleBroadcast sends a broadcast message to all connected clients.
func (h *Hub) handleBroadcast(raw []byte) {
	h.mu.RLock()
	for _, c := range h.clients {
		select {
		case c.Send <- command.CommandFrame{Type: "broadcast", Args: raw}:
		default:
		}
	}
	h.mu.RUnlock()
	_ = raw // prevent unused variable warning.
}

// handleDeviceStatus updates device online status in the database.
func (h *Hub) handleDeviceStatus(ctx context.Context, status DeviceStatusUpdate) {
	dbTimeout := 5 * time.Second
	opCtx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	if err := h.deviceRepo.SetOnline(opCtx, status.DeviceID, status.Online); err != nil {
		h.log.Warn("device status update failed",
			"deviceId", status.DeviceID, "online", status.Online, "source", status.Source, "err", err)
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

// SetDeviceOnline updates device online status via hub's event loop.

// channel, ensuring all status changes (from WS and REST) are serialized.
func (h *Hub) SetDeviceOnline(deviceID string, online bool) {
	h.deviceStatus <- DeviceStatusUpdate{DeviceID: deviceID, Online: online, Source: "rest"}
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
		// No filter configured, broadcast to all.
		h.BroadcastTelemetry(raw)
		return
	}

	// Snapshot clients under RLock, then release lock before iterating.
	h.mu.RLock()
	clients := make(map[string]*Client, len(h.clients))
	for id, c := range h.clients {
		clients[id] = c
	}
	h.mu.RUnlock()

	// Iterate snapshot without holding lock.
	for clientID, c := range clients {
		// Skip the sender.
		if clientID == senderDeviceID {
			continue
		}

		// Check if this client is subscribed to the sender's telemetry.
		if h.telemetryFilter.ShouldForward(clientID, senderDeviceID) {
			// Client buffer full, skip.
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
		// Device offline - try to queue if message queue is available.
		if h.messageQueue != nil {
			// Use synchronous DB write for 100% delivery guarantee (G1).
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
		// Client buffer full - try queue if available.
		if h.messageQueue != nil {
			// Use synchronous DB write for 100% delivery guarantee (G1).
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

	// Sample based on configuration.
	if rand.Float64() > h.latencyConfig.SampleRate {
		return
	}

	h.metricsMu.Lock()
	defer h.metricsMu.Unlock()

	// Update global metrics.
	h.updateLatencyMetrics(&h.metrics.LatencyMetrics, deviceID, dispatchID, latencyMS, success)

	// Update per-device metrics.
	dm, ok := h.deviceLatency[deviceID]
	if !ok {
		dm = &LatencyMetrics{}
		h.deviceLatency[deviceID] = dm
	}
	h.updateLatencyMetrics(dm, deviceID, dispatchID, latencyMS, success)
}

// updateLatencyMetrics updates a LatencyMetrics struct with a new sample.
func (h *Hub) updateLatencyMetrics(m *LatencyMetrics, deviceID, dispatchID string, latencyMS int64, success bool) {
	m.TotalMessages++
	if success {
		m.SuccessfulMessages++
	} else {
		m.FailedMessages++
	}
	m.TotalLatencyMS += latencyMS
	// Update min/max.
	if m.MinLatencyMS == 0 || latencyMS < m.MinLatencyMS {
		m.MinLatencyMS = latencyMS
	}

	if latencyMS > m.MaxLatencyMS {
		m.MaxLatencyMS = latencyMS
	}

	// Check if exceeds target (threshold is informational only).
	if latencyMS > int64(h.latencyConfig.MaxLatencyMS) {
		m.ExceededCount++
		m.LastExceededAt = time.Now().Unix()
		m.LastExceededID = dispatchID
		h.log.Warn("latency exceeded target",
			"deviceId", deviceID,
			"dispatchId", dispatchID,
			"latencyMs", latencyMS,
			"targetMs", h.latencyConfig.MaxLatencyMS,
		)
	}

	// Update average.
	m.AverageLatencyMS = float64(m.TotalLatencyMS) / float64(m.TotalMessages)
}

// SendWithDeliveryConfirmation sends a message and waits for delivery confirmation.
// This ensures 100% message delivery by waiting for WebSocket write confirmation.
func (h *Hub) SendWithDeliveryConfirmation(deviceID string, frame command.CommandFrame, timeout time.Duration) (bool, error) {
	h.mu.RLock()
	c := h.clients[deviceID]
	h.mu.RUnlock()

	if c == nil {
		// Device offline - queue with synchronous DB write.
		if h.messageQueue != nil {
			return h.messageQueue.EnqueueWithConfirmation(deviceID, frame), nil
		}

		return false, nil
	}

	// Create a channel for delivery confirmation.
	confirmation := make(chan bool, 1)
	frame.DeliveryConfirmation = confirmation

	select {
	case c.Send <- frame:
		// Wait for confirmation or timeout using timer that can be stopped.
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		select {
		case confirmed := <-confirmation:
			return confirmed, nil
		case <-timer.C:
			// Timeout - try to queue if available.
			if h.messageQueue != nil {
				return h.messageQueue.EnqueueWithConfirmation(deviceID, frame), nil
			}

			return false, fmt.Errorf("delivery confirmation timeout")
		}
	default:
		// Client buffer full - queue with synchronous write.
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

// schedulePartialReplay schedules a retry for partially replayed messages.
// Uses exponential backoff with a maximum of 3 retries.
func (h *Hub) schedulePartialReplay(deviceID string, remaining int) {
	h.mu.RLock()
	client, exists := h.clients[deviceID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	// Schedule retry with exponential backoff (100ms, 200ms, 400ms).
	go func() {
		for attempt := 0; attempt < 3; attempt++ {
			backoff := time.Duration(100<<attempt) * time.Millisecond
			time.Sleep(backoff)

			// Check if client still exists.
			h.mu.RLock()
			client, exists = h.clients[deviceID]
			h.mu.RUnlock()

			if !exists || client == nil {
				return
			}

			// Try to replay remaining messages.
			result := h.messageQueue.ReplayQueue(deviceID, client.Send)
			if result.Count > 0 {
				h.log.Info("retry replayed queued messages to device",
					"deviceId", deviceID,
					"count", result.Count,
					"attempt", attempt+1,
				)
			}

			if !result.HasMore {
				return
			}
		}

		h.log.Warn("partial replay exhausted retries",
			"deviceId", deviceID,
			"remaining", remaining,
		)
	}()
}

// BroadcastEvent emits an event to all subscribed dashboard clients.
func (h *Hub) BroadcastEvent(evtType string, data []byte) error {
	if h.dashboardBroadcaster == nil {
		return nil
	}

	// Broadcast to all connected dashboards.
	return h.dashboardBroadcaster.BroadcastToDashboard("", "", evtType, data)
}

// EmitDeviceConnected emits a device connected event.
func (h *Hub) EmitDeviceConnected(deviceID string) error {
	if h.eventProcessor == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return h.eventProcessor.ProcessDeviceConnected(ctx, deviceID, nil)
}

// EmitDeviceDisconnected emits a device disconnected event.
func (h *Hub) EmitDeviceDisconnected(deviceID string) error {
	if h.eventProcessor == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return h.eventProcessor.ProcessDeviceDisconnected(ctx, deviceID, "ws_disconnect", nil)
}

// EmitThresholdBreach emits a threshold breach event.
func (h *Hub) EmitThresholdBreach(deviceID string, metric string, value float64) error {
	if h.eventProcessor == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metadata := map[string]any{
		"metric": metric,
		"value":  value,
	}

	return h.eventProcessor.ProcessTelemetry(ctx, deviceID, metadata)
}

// ConnectionInfo holds WebSocket connection information for a device.
type ConnectionInfo struct {
	ConnectedAt time.Time
	ClientIP    string
	Connected   bool
}

// GetConnectionInfo retrieves WebSocket connection information for a device.
func (h *Hub) GetConnectionInfo(deviceID string) *ConnectionInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	c, ok := h.clients[deviceID]
	if !ok || c == nil {
		return nil
	}

	// Check actual connection state, not just presence in map.
	isConnected := c.IsConnected()
	if !isConnected {
		return nil
	}

	info := &ConnectionInfo{
		Connected:   true,
		ConnectedAt: time.Unix(c.connectedAt, 0),
	}

	if c.Conn != nil {
		info.ClientIP = c.Conn.RemoteAddr().String()
		if idx := strings.LastIndex(info.ClientIP, ":"); idx > 0 {
			info.ClientIP = info.ClientIP[:idx]
		}
	}

	return info
}

// GetAverageLatency returns the average latency in milliseconds for a device.
// Falls back to the global average only when no per-device samples exist.
func (h *Hub) GetAverageLatency(deviceID string) int {
	h.metricsMu.RLock()
	defer h.metricsMu.RUnlock()

	if dm, ok := h.deviceLatency[deviceID]; ok && dm.TotalMessages > 0 {
		return int(dm.AverageLatencyMS)
	}

	if h.metrics.LatencyMetrics.TotalMessages == 0 {
		return 0
	}

	return int(h.metrics.LatencyMetrics.AverageLatencyMS)
}
