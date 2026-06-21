package hub

import (
	"log/slog"
	"sync"
	"time"
)

// TelemetryFilterConfig holds configuration for telemetry filtering.
type TelemetryFilterConfig struct {
	// MaxSubscriptions is the maximum number of device subscriptions per client (default 50)
	MaxSubscriptions int
	// EnableServerSideFilter enables server-side filtering (default true)
	EnableServerSideFilter bool
}

// DefaultTelemetryFilterConfig returns the default telemetry filter configuration.
func DefaultTelemetryFilterConfig() *TelemetryFilterConfig {
	return &TelemetryFilterConfig{
		MaxSubscriptions:         50,
		EnableServerSideFilter:   true,
	}
}

// TelemetryFilterMetrics holds telemetry filter metrics.
type TelemetryFilterMetrics struct {
	TotalSubscriptions   int64 `json:"totalSubscriptions"`
	TotalUnsubscribes   int64 `json:"totalUnsubscribes"`
	TotalFiltered       int64 `json:"totalFiltered"`
	TotalForwarded      int64 `json:"totalForwarded"`
	ActiveSubscriptions  int   `json:"activeSubscriptions"`
}

// Subscription represents a client's subscription to a device's telemetry.
type Subscription struct {
	ClientID  string
	DeviceID  string
	SubscribedAt int64 // Unix timestamp
}

// TelemetryFilter manages client subscriptions to specific device telemetry.
// It supports server-side filtering to reduce bandwidth for dashboard clients.
type TelemetryFilter struct {
	log     *slog.Logger
	config  *TelemetryFilterConfig
	
	// subscriptions maps client ID -> set of device IDs they're subscribed to
	subscriptions map[string]map[string]*Subscription
	// deviceSubscribers maps device ID -> set of client IDs subscribed to it
	deviceSubscribers map[string]map[string]bool
	
	mu sync.RWMutex
	metrics TelemetryFilterMetrics
	metricsMu sync.RWMutex
}

// NewTelemetryFilter creates a new TelemetryFilter.
func NewTelemetryFilter(log *slog.Logger, cfg *TelemetryFilterConfig) *TelemetryFilter {
	if cfg == nil {
		cfg = DefaultTelemetryFilterConfig()
	}
	return &TelemetryFilter{
		log:               log,
		config:            cfg,
		subscriptions:     make(map[string]map[string]*Subscription),
		deviceSubscribers: make(map[string]map[string]bool),
	}
}

// Subscribe subscribes a client to a device's telemetry.
func (tf *TelemetryFilter) Subscribe(clientID, deviceID string) bool {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	// Check if client has reached max subscriptions
	if tf.clientSubscriptionCount(clientID) >= tf.config.MaxSubscriptions {
		tf.log.Warn("client max subscriptions reached",
			"clientId", clientID,
			"maxSubscriptions", tf.config.MaxSubscriptions,
		)
		return false
	}

	// Create subscription
	if tf.subscriptions[clientID] == nil {
		tf.subscriptions[clientID] = make(map[string]*Subscription)
	}

	tf.subscriptions[clientID][deviceID] = &Subscription{
		ClientID:  clientID,
		DeviceID:  deviceID,
		SubscribedAt: nowUnix(),
	}

	// Add to device subscriber index
	if tf.deviceSubscribers[deviceID] == nil {
		tf.deviceSubscribers[deviceID] = make(map[string]bool)
	}
	tf.deviceSubscribers[deviceID][clientID] = true

	tf.incrementSubscriptions()
	tf.log.Info("client subscribed to device telemetry",
		"clientId", clientID,
		"deviceId", deviceID,
	)

	return true
}

// Unsubscribe unsubscribes a client from a device's telemetry.
func (tf *TelemetryFilter) Unsubscribe(clientID, deviceID string) bool {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if tf.subscriptions[clientID] == nil {
		return false
	}

	if _, exists := tf.subscriptions[clientID][deviceID]; !exists {
		return false
	}

	delete(tf.subscriptions[clientID], deviceID)
	delete(tf.deviceSubscribers[deviceID], clientID)
	
	// Cleanup empty maps
	if len(tf.subscriptions[clientID]) == 0 {
		delete(tf.subscriptions, clientID)
	}
	if len(tf.deviceSubscribers[deviceID]) == 0 {
		delete(tf.deviceSubscribers, deviceID)
	}

	tf.incrementUnsubscribes()
	tf.log.Info("client unsubscribed from device telemetry",
		"clientId", clientID,
		"deviceId", deviceID,
	)

	return true
}

// IsSubscribed checks if a client is subscribed to a device.
func (tf *TelemetryFilter) IsSubscribed(clientID, deviceID string) bool {
	tf.mu.RLock()
	defer tf.mu.RUnlock()

	if tf.subscriptions[clientID] == nil {
		return false
	}
	_, exists := tf.subscriptions[clientID][deviceID]
	return exists
}

// GetSubscriptions returns all device IDs a client is subscribed to.
func (tf *TelemetryFilter) GetSubscriptions(clientID string) []string {
	tf.mu.RLock()
	defer tf.mu.RUnlock()

	if tf.subscriptions[clientID] == nil {
		return nil
	}

	devices := make([]string, 0, len(tf.subscriptions[clientID]))
	for deviceID := range tf.subscriptions[clientID] {
		devices = append(devices, deviceID)
	}
	return devices
}

// GetSubscribers returns all client IDs subscribed to a device.
func (tf *TelemetryFilter) GetSubscribers(deviceID string) []string {
	tf.mu.RLock()
	defer tf.mu.RUnlock()

	if tf.deviceSubscribers[deviceID] == nil {
		return nil
	}

	clients := make([]string, 0, len(tf.deviceSubscribers[deviceID]))
	for clientID := range tf.deviceSubscribers[deviceID] {
		clients = append(clients, clientID)
	}
	return clients
}

// ShouldForward checks if telemetry from a device should be forwarded to a specific client.
func (tf *TelemetryFilter) ShouldForward(clientID, deviceID string) bool {
	// If filtering is disabled, forward all
	if !tf.config.EnableServerSideFilter {
		return true
	}

	// If client has no subscriptions, forward all (dashboard mode)
	tf.mu.RLock()
	subs, ok := tf.subscriptions[clientID]
	hasSubscriptions := ok && len(subs) > 0
	tf.mu.RUnlock()

	if !hasSubscriptions {
		return true
	}

	// Check subscription
	isSubscribed := tf.IsSubscribed(clientID, deviceID)
	
	if isSubscribed {
		tf.incrementForwarded()
	} else {
		tf.incrementFiltered()
	}
	
	return isSubscribed
}

// clientSubscriptionCount returns the number of subscriptions for a client.
func (tf *TelemetryFilter) clientSubscriptionCount(clientID string) int {
	if tf.subscriptions[clientID] == nil {
		return 0
	}
	return len(tf.subscriptions[clientID])
}

// GetMetrics returns a copy of the current metrics.
func (tf *TelemetryFilter) GetMetrics() TelemetryFilterMetrics {
	tf.metricsMu.RLock()
	defer tf.metricsMu.RUnlock()

	// Update active count
	metrics := tf.metrics
	metrics.ActiveSubscriptions = tf.totalSubscriptions()

	return metrics
}

func (tf *TelemetryFilter) totalSubscriptions() int {
	tf.mu.RLock()
	defer tf.mu.RUnlock()

	total := 0
	for _, subs := range tf.subscriptions {
		total += len(subs)
	}
	return total
}

func (tf *TelemetryFilter) incrementSubscriptions() {
	tf.metricsMu.Lock()
	tf.metrics.TotalSubscriptions++
	tf.metricsMu.Unlock()
}

func (tf *TelemetryFilter) incrementUnsubscribes() {
	tf.metricsMu.Lock()
	tf.metrics.TotalUnsubscribes++
	tf.metricsMu.Unlock()
}

func (tf *TelemetryFilter) incrementFiltered() {
	tf.metricsMu.Lock()
	tf.metrics.TotalFiltered++
	tf.metricsMu.Unlock()
}

func (tf *TelemetryFilter) incrementForwarded() {
	tf.metricsMu.Lock()
	tf.metrics.TotalForwarded++
	tf.metricsMu.Unlock()
}

// UnsubscribeAll unsubscribes a client from all devices.
func (tf *TelemetryFilter) UnsubscribeAll(clientID string) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	devices, ok := tf.subscriptions[clientID]
	if !ok {
		return
	}

	for deviceID := range devices {
		delete(tf.deviceSubscribers[deviceID], clientID)
		if len(tf.deviceSubscribers[deviceID]) == 0 {
			delete(tf.deviceSubscribers, deviceID)
		}
	}
	delete(tf.subscriptions, clientID)

	tf.metricsMu.Lock()
	tf.metrics.TotalUnsubscribes += int64(len(devices))
	tf.metricsMu.Unlock()
}

// nowUnix returns current Unix timestamp.
func nowUnix() int64 {
	return time.Now().Unix()
}
