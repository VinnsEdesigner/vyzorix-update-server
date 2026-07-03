// Package diagnostics provides the diagnostics service implementation.
package diagnostics

import (
	"context"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	ws "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
)

// Service provides diagnostics operations.
type Service struct {
	diagnosticsRepo diagnostics.Repository
	deviceRepo      device.Repository
	hub             *ws.Hub
}

// NewService creates a new diagnostics service.
func NewService(diagnosticsRepo diagnostics.Repository, deviceRepo device.Repository, hub *ws.Hub) *Service {
	return &Service{
		diagnosticsRepo: diagnosticsRepo,
		deviceRepo:      deviceRepo,
		hub:             hub,
	}
}

// GetDeviceInspection retrieves full device inspection data.
func (s *Service) GetDeviceInspection(ctx context.Context, imei string) (*diagnostics.DeviceInspection, error) {
	// Get device by IMEI
	dev, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil {
		return nil, err
	}
	if dev == nil {
		return nil, diagnostics.ErrDeviceNotFound
	}

	// Build identity info
	identity := diagnostics.IdentityInfo{
		IMEI:         dev.ID,
		DeviceName:   dev.DeviceName,
		Model:        dev.Model,
		Manufacturer: dev.Manufacturer,
	}

	// Build software info
	// Note: BuildID, SecurityPatch, OSVersion may come from telemetry metadata
	// For now, we use what's available in the device record
	software := diagnostics.SoftwareInfo{
		OSVersion:     dev.OSVersion,
		AppVersion:    dev.AppVersion,
		SecurityPatch: dev.SecurityPatch,
		BuildID:       "", // Will be populated from telemetry if available
	}

	// Build registration info
	var registeredAt *time.Time
	if dev.RegisteredAt > 0 {
		t := time.UnixMilli(dev.RegisteredAt)
		registeredAt = &t
	}

	registration := diagnostics.RegistrationInfo{
		Status:              s.determineDeviceStatus(dev),
		RegisteredAt:        registeredAt,
		FCMTokenValid:       dev.FCMToken != "" && dev.IsFCMTokenValid(),
		FCMTokenRefreshedAt: dev.FCMTokenRefreshedAtTime(),
		CommandSecretSet:    dev.CommandSecretHash != "",
	}

	// Build connection info
	var connectedAt *time.Time
	var clientIP string
	var wsConnected bool

	if s.hub != nil {
		wsConnected = s.hub.Online(dev.ID)
		if wsConnected {
			now := time.Now()
			connectedAt = &now
		}
		// Try to get client for IP - we'll skip this for now as RemoteAddr may not exist
	}

	var lastSeen *time.Time
	if dev.LastSeen > 0 {
		t := time.UnixMilli(dev.LastSeen)
		lastSeen = &t
	}

	connection := diagnostics.ConnectionInfo{
		WebSocketStatus: s.determineWebSocketStatus(dev, wsConnected),
		ConnectedAt:     connectedAt,
		FCMStatus:       s.determineFCMStatus(dev),
		LastSeen:        lastSeen,
		ClientIP:        clientIP,
		Protocol:        "WSS",
	}

	// Build telemetry info
	telemetry, err := s.diagnosticsRepo.GetTelemetryStats(ctx, dev.ID)
	if err != nil {
		// Log error but don't fail - telemetry might be empty for new devices
		telemetry = &diagnostics.TelemetryInfo{}
	}

	// Get last telemetry timestamp
	lastTelemetry, err := s.diagnosticsRepo.GetLastTelemetry(ctx, dev.ID)
	if err == nil && lastTelemetry != nil {
		telemetry.LastTimestamp = lastTelemetry.Timestamp
	}

	return &diagnostics.DeviceInspection{
		Identity:     identity,
		Software:     software,
		Registration: registration,
		Connection:   connection,
		Telemetry:    *telemetry,
	}, nil
}

// GetDeviceTimeline retrieves paginated timeline events for a device.
func (s *Service) GetDeviceTimeline(ctx context.Context, imei string, req *TimelineRequest) (*TimelineResponse, error) {
	// Verify device exists
	dev, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil {
		return nil, err
	}
	if dev == nil {
		return nil, diagnostics.ErrDeviceNotFound
	}

	// Build filter
	filter := &diagnostics.TimelineFilter{
		EventType: req.EventType,
		Limit:    req.Limit,
		Cursor:   req.Cursor,
	}

	// Apply time range
	if req.StartTime > 0 {
		filter.StartTime = time.UnixMilli(req.StartTime)
	} else {
		// Default to last 24 hours
		filter.StartTime = time.Now().Add(-24 * time.Hour)
	}

	if req.EndTime > 0 {
		filter.EndTime = time.UnixMilli(req.EndTime)
	}

	// Enforce max limit
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}

	// Get timeline events
	result, err := s.diagnosticsRepo.GetTimelineEvents(ctx, dev.ID, filter)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	events := make([]EventDTO, len(result.Events))
	for i, e := range result.Events {
		events[i] = EventDTO{
			ID:        e.ID,
			Type:      string(e.Type),
			Timestamp: e.Timestamp.UnixMilli(),
			Data:      e.Data,
		}
	}

	return &TimelineResponse{
		Events: events,
		Pagination: PaginationDTO{
			Limit:      filter.Limit,
			HasMore:    result.HasMore,
			NextCursor: result.NextCursor,
		},
	}, nil
}

// RecordDeviceEvent records a new device event.
func (s *Service) RecordDeviceEvent(ctx context.Context, deviceID string, eventType diagnostics.TimelineEventType, data map[string]any) error {
	event := &diagnostics.TimelineEvent{
		ID:        generateID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	return s.diagnosticsRepo.RecordEvent(ctx, event)
}

// determineDeviceStatus determines the device registration status.
func (s *Service) determineDeviceStatus(dev *device.Device) string {
	if dev.DeregisteredAt != nil {
		return "deregistered"
	}
	if dev.LastSeen > 0 && time.Since(time.UnixMilli(dev.LastSeen)) > 5*time.Minute {
		return "offline"
	}
	return "online"
}

// determineWebSocketStatus determines the WebSocket connection status.
func (s *Service) determineWebSocketStatus(dev *device.Device, wsConnected bool) string {
	if wsConnected {
		return "connected"
	}
	if dev.LastSeen > 0 && time.Since(time.UnixMilli(dev.LastSeen)) < 5*time.Minute {
		return "connected" // Might be using FCM
	}
	return "disconnected"
}

// determineFCMStatus determines the FCM token validity status.
func (s *Service) determineFCMStatus(dev *device.Device) string {
	if dev.FCMToken == "" {
		return "not_set"
	}
	if dev.FCMTokenRefreshedAt != nil && *dev.FCMTokenRefreshedAt > 0 {
		if time.Since(time.UnixMilli(*dev.FCMTokenRefreshedAt)) > 30*24*time.Hour {
			return "invalid"
		}
	}
	return "valid"
}

// generateID generates a unique ID for events.
func generateID() string {
	// Simple UUID-like generation
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> uint(i*8) & 0xff)
	}
	return fmt.Sprintf("evt_%x", b)
}
