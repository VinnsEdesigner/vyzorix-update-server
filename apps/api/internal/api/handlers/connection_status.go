package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/gin-gonic/gin"
)

// ConnectionStatusHandler handles WebSocket connection status requests.
type ConnectionStatusHandler struct {
	log *slog.Logger
	hub *hub.Hub
}

// NewConnectionStatusHandler creates a new ConnectionStatusHandler.
func NewConnectionStatusHandler(log *slog.Logger, hub *hub.Hub) *ConnectionStatusHandler {
	return &ConnectionStatusHandler{
		log: log,
		hub: hub,
	}
}

// DeviceConnectionStatus represents the connection status of a single device.
type DeviceConnectionStatus struct {
	DeviceID         string `json:"deviceId"`
	Online           bool   `json:"online"`
	ConnectedAt      int64  `json:"connectedAt,omitempty"`
	UptimeSeconds    int64  `json:"uptimeSeconds,omitempty"`
	MessagesSent     int32  `json:"messagesSent,omitempty"`
	MessagesReceived int32  `json:"messagesReceived,omitempty"`
	QueueSize        int    `json:"queueSize,omitempty"`
	LastMessageAt    int64  `json:"lastMessageAt,omitempty"`
}

// ConnectionStatusResponse represents the response for connection status queries.
type ConnectionStatusResponse struct {
	ClientMetrics      *hub.ClientMetrics      `json:"clientMetrics,omitempty"`
	QueueMetrics       *hub.QueueMetrics       `json:"queueMetrics,omitempty"`
	RateLimiterMetrics *hub.RateLimiterMetrics `json:"rateLimiterMetrics,omitempty"`
	DeviceID           string                  `json:"deviceId"`
	ConnectedAt        int64                   `json:"connectedAt,omitempty"`
	UptimeSeconds      int64                   `json:"uptimeSeconds,omitempty"`
	Online             bool                    `json:"online"`
}

// AllConnectionsResponse represents the status of all connected devices.
type AllConnectionsResponse struct {
	Devices        []DeviceConnectionStatus `json:"devices"`
	QueueMetrics   hub.QueueMetrics         `json:"queueMetrics"`
	TotalConnected int                      `json:"totalConnected"`
	TotalQueued    int                      `json:"totalQueued"`
}

// GetStatus handles GET /v1/device/:id/connection-status
// Returns the WebSocket connection status for a specific device.
func (h *ConnectionStatusHandler) GetStatus(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "device id is required",
		})

		return
	}

	// Get client from hub
	client := h.hub.GetClient(deviceID)

	response := ConnectionStatusResponse{
		DeviceID: deviceID,
		Online:   client != nil,
	}

	if client != nil {
		// Device is connected
		metrics := client.GetMetrics()
		response.ConnectedAt = metrics.LastConnectedAt
		response.UptimeSeconds = client.Uptime()
		response.ClientMetrics = &metrics
	}

	// Get queue status for this device
	{
		response.QueueMetrics = &hub.QueueMetrics{}

		if queueSize := h.hub.QueueSize(deviceID); queueSize >= 0 {
			// Get full metrics if available
			if qm, ok := h.hub.QueueMetrics(); ok {
				response.QueueMetrics = &qm
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetAllStatus handles GET /v1/connections
// Returns the status of all WebSocket connections.
func (h *ConnectionStatusHandler) GetAllStatus(c *gin.Context) {
	// Get all clients
	clients := h.hub.Clients()

	devices := make([]DeviceConnectionStatus, 0, len(clients))

	for deviceID, client := range clients {
		metrics := client.GetMetrics()

		status := DeviceConnectionStatus{
			DeviceID:         deviceID,
			Online:           true,
			ConnectedAt:      metrics.LastConnectedAt,
			UptimeSeconds:    client.Uptime(),
			MessagesSent:     metrics.MessagesSent,
			MessagesReceived: metrics.MessagesReceived,
		}

		// Get queue size for this device
		status.QueueSize = h.hub.QueueSize(deviceID)

		devices = append(devices, status)
	}

	// Get overall queue metrics
	var queueMetrics hub.QueueMetrics
	if qm, ok := h.hub.QueueMetrics(); ok {
		queueMetrics = qm
	}

	// Calculate total queued messages
	totalQueued := h.hub.TotalQueuedMessages()

	response := AllConnectionsResponse{
		TotalConnected: len(devices),
		TotalQueued:    totalQueued,
		Devices:        devices,
		QueueMetrics:   queueMetrics,
	}

	c.JSON(http.StatusOK, response)
}

// GetMetrics handles GET /v1/connections/metrics
// Returns aggregate WebSocket metrics.
func (h *ConnectionStatusHandler) GetMetrics(c *gin.Context) {
	clients := h.hub.Clients()

	// Aggregate metrics
	var totalMessagesSent, totalMessagesReceived int64

	var totalConnectAttempts, totalConnectSuccesses, totalConnectFailures int64

	for _, client := range clients {
		metrics := client.GetMetrics()
		totalMessagesSent += int64(metrics.MessagesSent)
		totalMessagesReceived += int64(metrics.MessagesReceived)
		totalConnectAttempts += int64(metrics.ConnectAttempts)
		totalConnectSuccesses += int64(metrics.ConnectSuccesses)
		totalConnectFailures += int64(metrics.ConnectFailures)
	}

	// Get queue metrics
	var queueMetrics hub.QueueMetrics
	if qm, ok := h.hub.QueueMetrics(); ok {
		queueMetrics = qm
	}

	c.JSON(http.StatusOK, gin.H{
		"timestamp":      time.Now().Unix(),
		"totalConnected": len(clients),
		"totalQueued":    h.hub.TotalQueuedMessages(),
		"aggregateMetrics": gin.H{
			"totalMessagesSent":     totalMessagesSent,
			"totalMessagesReceived": totalMessagesReceived,
			"totalConnectAttempts":  totalConnectAttempts,
			"totalConnectSuccesses": totalConnectSuccesses,
			"totalConnectFailures":  totalConnectFailures,
		},
		"queueMetrics": queueMetrics,
	})
}

// DisconnectDevice handles POST /v1/device/:id/disconnect
// Forcefully disconnects a device's WebSocket connection.
func (h *ConnectionStatusHandler) DisconnectDevice(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "device id is required",
		})

		return
	}

	// Require operator auth
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "authentication required",
		})

		return
	}

	// Require admin or operator role
	if op.Role != "admin" && op.Role != "operator" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "admin or operator role required",
		})

		return
	}

	// Get and disconnect client
	client := h.hub.GetClient(deviceID)
	if client == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "device not connected",
		})

		return
	}

	// Remove from hub (this will close the connection via ReadPump defer)
	h.hub.Unregister(client)

	h.log.Info("device disconnected by operator",
		"deviceId", deviceID,
		"operatorId", op.ID,
	)

	c.JSON(http.StatusOK, gin.H{
		"deviceId":     deviceID,
		"disconnected": true,
		"operatorId":   op.ID,
	})
}
