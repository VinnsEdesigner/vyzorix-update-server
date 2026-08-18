package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/gin-gonic/gin"
)

// ConnectionStatusHandler handles WebSocket connection status requests.
type ConnectionStatusHandler struct {
	log        *slog.Logger
	hub        *hub.Hub
	deviceRepo *storage.DeviceRepository
}

// NewConnectionStatusHandler creates a new ConnectionStatusHandler.
func NewConnectionStatusHandler(log *slog.Logger, hub *hub.Hub, deviceRepo *storage.DeviceRepository) *ConnectionStatusHandler {
	return &ConnectionStatusHandler{
		log:        log,
		hub:        hub,
		deviceRepo: deviceRepo,
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

// GetStatus handles GET /v1/device/:id/connection-status.
// Returns the WebSocket connection status for a specific device.
func (h *ConnectionStatusHandler) GetStatus(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "device id is required"))

		return
	}

	// Get client from hub.
	client := h.hub.GetClient(deviceID)

	response := ConnectionStatusResponse{
		DeviceID: deviceID,
		Online:   client != nil,
	}

	if client != nil {
		metrics := client.GetMetrics()
		response.ConnectedAt = metrics.LastConnectedAt
		response.UptimeSeconds = client.Uptime()
		response.ClientMetrics = &metrics
	}

	// Get queue status for this device.
	{
		response.QueueMetrics = &hub.QueueMetrics{}

		if queueSize := h.hub.QueueSize(deviceID); queueSize >= 0 {
			if qm, ok := h.hub.QueueMetrics(); ok {
				response.QueueMetrics = &qm
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetAllStatus handles GET /v1/connections.
// Returns the status of all WebSocket connections within the organization.
func (h *ConnectionStatusHandler) GetAllStatus(c *gin.Context) {
	if h.hub == nil {
		c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "WebSocket hub not initialized"))

		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))

		return
	}

	// Get devices in this organization.
	var orgDeviceIDs = make(map[string]bool)
	if h.deviceRepo != nil {
		orgDevices, err := h.deviceRepo.ListByOrganization(c.Request.Context(), orgID)
		if err == nil {
			for _, d := range orgDevices {
				orgDeviceIDs[d.ID] = true
			}
		}
	}

	// Get all clients and filter by organization.
	allClients := h.hub.Clients()
	orgClients := make(map[string]*hub.Client)
	for deviceID, client := range allClients {
		if orgDeviceIDs[deviceID] {
			orgClients[deviceID] = client
		}
	}

	devices := make([]DeviceConnectionStatus, 0, len(orgClients))

	for deviceID, client := range orgClients {
		metrics := client.GetMetrics()

		status := DeviceConnectionStatus{
			DeviceID:         deviceID,
			Online:           true,
			ConnectedAt:      metrics.LastConnectedAt,
			UptimeSeconds:    client.Uptime(),
			MessagesSent:     metrics.MessagesSent,
			MessagesReceived: metrics.MessagesReceived,
			LastMessageAt:    metrics.LastMessageAt,
		}

		status.QueueSize = h.hub.QueueSize(deviceID)

		devices = append(devices, status)
	}

	var queueMetrics hub.QueueMetrics
	if qm, ok := h.hub.QueueMetrics(); ok {
		queueMetrics = qm
	}

	totalQueued := h.hub.TotalQueuedMessages()

	response := AllConnectionsResponse{
		TotalConnected: len(devices),
		TotalQueued:    totalQueued,
		Devices:        devices,
		QueueMetrics:   queueMetrics,
	}

	c.JSON(http.StatusOK, response)
}

// GetMetrics handles GET /v1/connections/metrics.
// Returns aggregate WebSocket metrics for the organization.
func (h *ConnectionStatusHandler) GetMetrics(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))

		return
	}

	// Get devices in this organization.
	var orgDeviceIDs = make(map[string]bool)
	if h.deviceRepo != nil {
		orgDevices, err := h.deviceRepo.ListByOrganization(c.Request.Context(), orgID)
		if err == nil {
			for _, d := range orgDevices {
				orgDeviceIDs[d.ID] = true
			}
		}
	}

	// Get all clients and filter by organization.
	allClients := h.hub.Clients()
	orgClients := make(map[string]*hub.Client)
	for deviceID, client := range allClients {
		if orgDeviceIDs[deviceID] {
			orgClients[deviceID] = client
		}
	}

	// Aggregate metrics.
	var totalMessagesSent, totalMessagesReceived int64
	var totalConnectAttempts, totalConnectSuccesses, totalConnectFailures int64

	for _, client := range orgClients {
		metrics := client.GetMetrics()
		totalMessagesSent += int64(metrics.MessagesSent)
		totalMessagesReceived += int64(metrics.MessagesReceived)
		totalConnectAttempts += int64(metrics.ConnectAttempts)
		totalConnectSuccesses += int64(metrics.ConnectSuccesses)
		totalConnectFailures += int64(metrics.ConnectFailures)
	}

	var queueMetrics hub.QueueMetrics
	if qm, ok := h.hub.QueueMetrics(); ok {
		queueMetrics = qm
	}

	c.JSON(http.StatusOK, gin.H{
		"timestamp":      time.Now().Unix(),
		"totalConnected": len(orgClients),
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

// DisconnectDevice handles POST /v1/device/:id/disconnect.
// Forcefully disconnects a device's WebSocket connection within the organization.
func (h *ConnectionStatusHandler) DisconnectDevice(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))

		return
	}

	deviceID := c.Param("id")
	if deviceID == "" {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "device id is required"))

		return
	}

	// Verify device belongs to organization.
	if h.deviceRepo != nil {
		_, err := h.deviceRepo.FindByIDAndOrganization(c.Request.Context(), deviceID, orgID)
		if err != nil {
			c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "device not found in organization"))

			return
		}
	}

	// Require operator auth.
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "authentication required"))

		return
	}

	// Get and disconnect client.
	client := h.hub.GetClient(deviceID)
	if client == nil {
		c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "device not connected"))

		return
	}

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
