// Package diagnostics provides the diagnostics service implementation.
package diagnostics

import (
	"context"
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/diagnostics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
	ws "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
)

// InspectCacheEntry caches inspection data with TTL.
type InspectCacheEntry struct {
	Data      *HTTPInspectionResponse
	ExpiresAt time.Time
}

// Service provides diagnostics operations.
type Service struct {
	diagnosticsRepo diagnostics.Repository
	deviceRepo      device.Repository
	hub             *ws.Hub
	inspectCache   map[string]*InspectCacheEntry
	cacheMu        sync.RWMutex
	cacheTTL       time.Duration
	cfg            config.DiagnosticsConfig
}

// NewService creates a new diagnostics service.
func NewService(diagnosticsRepo diagnostics.Repository, deviceRepo device.Repository, hub *ws.Hub, cfg config.DiagnosticsConfig) *Service {
	s := &Service{
		diagnosticsRepo: diagnosticsRepo,
		deviceRepo:      deviceRepo,
		hub:             hub,
		inspectCache:   make(map[string]*InspectCacheEntry),
		cacheTTL:       time.Duration(cfg.InspectionCacheTTLSeconds) * time.Second,
		cfg:            cfg,
	}
	// Start cache cleanup goroutine
	go s.cleanupCache()
	return s
}

// cleanupCache periodically removes expired cache entries.
func (s *Service) cleanupCache() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.cacheMu.Lock()
		now := time.Now()
		for key, entry := range s.inspectCache {
			if entry.ExpiresAt.Before(now) {
				delete(s.inspectCache, key)
// getCachedInspection retrieves cached inspection if not expired.
func (s *Service) getCachedInspection(cacheKey string) *HTTPInspectionResponse {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	if entry, ok := s.inspectCache[cacheKey]; ok && entry.ExpiresAt.After(time.Now()) {
		return entry.Data
	}
	return nil
}

// cacheInspection stores inspection in cache.
func (s *Service) cacheInspection(cacheKey string, data *HTTPInspectionResponse) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.inspectCache[cacheKey] = &InspectCacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(s.cacheTTL),
	}
}

// GetDeviceInspection retrieves full device inspection data.
// Requires orgID for multi-tenant isolation.
func (s *Service) GetDeviceInspection(ctx context.Context, imei, orgID string) (*diagnostics.DeviceInspection, error) {
	dev, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil {
		return nil, err
	}
	if dev == nil {
		return nil, diagnostics.ErrDeviceNotFound
	}

	// Verify device belongs to organization
	if orgID != "" && dev.OrganizationID != orgID {
		return nil, diagnostics.ErrDeviceNotFound
	}
	}

	// Verify device belongs to organization
	if orgID != "" && dev.OrganizationID != orgID {
		return nil, diagnostics.ErrDeviceNotFound
	}
	dev, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil {
		return nil, err
	}
	if dev == nil {
		return nil, diagnostics.ErrDeviceNotFound
	}

	// Verify device belongs to organization
	if orgID != "" && dev.OrganizationID != orgID {
		return nil, diagnostics.ErrDeviceNotFound
	}

	identity := diagnostics.IdentityInfo{
		IMEI:         dev.ID,
		DeviceName:   dev.DeviceName,
		Model:        dev.Model,
		Manufacturer: dev.Manufacturer,
	}

	software := diagnostics.SoftwareInfo{
		OSVersion:     dev.OSVersion,
		AppVersion:    dev.AppVersion,
		SecurityPatch: dev.SecurityPatch,
		BuildID:       "",
	}

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

	var connectedAt *time.Time
	var clientIP string
	wsConnected := false

	if s.hub != nil {
		connInfo := s.hub.GetConnectionInfo(dev.ID)
		if connInfo != nil {
			wsConnected = connInfo.Connected
			connectedAt = &connInfo.ConnectedAt
			clientIP = connInfo.ClientIP
		}
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

	telemetry, err := s.diagnosticsRepo.GetTelemetryStats(ctx, dev.ID)
	if err != nil {
		telemetry = &diagnostics.TelemetryInfo{}
	}

	if s.hub != nil && wsConnected {
		telemetry.AvgLatencyMs = s.hub.GetAverageLatency(dev.ID)
	}

	lastTelemetry, err := s.diagnosticsRepo.GetLastTelemetry(ctx, dev.ID)
	// ErrNoTelemetryData is expected - device may have no telemetry yet
	// Other errors are also treated as "no telemetry" since it's supplementary data
// GetDeviceInspectionHTTP returns HTTP-specific response with int64 timestamps per spec.
// Results are cached for 10 seconds per spec.
// Requires orgID for multi-tenant isolation.
func (s *Service) GetDeviceInspectionHTTP(ctx context.Context, imei, orgID string) (*HTTPInspectionResponse, error) {
	// Check cache first (include orgID in cache key for multi-tenant isolation)
	cacheKey := imei + ":" + orgID
	if cached := s.getCachedInspection(cacheKey); cached != nil {
		return cached, nil
	}

	inspection, err := s.GetDeviceInspection(ctx, imei, orgID)
	if err != nil {
		return nil, err
	}

	resp := &HTTPInspectionResponse{
		Identity: HTTPIdentityInfo{
			IMEI:         inspection.Identity.IMEI,
			DeviceName:   inspection.Identity.DeviceName,
			Model:        inspection.Identity.Model,
			Manufacturer: inspection.Identity.Manufacturer,
		},
		Software: HTTPSoftwareInfo{
			OSVersion:     inspection.Software.OSVersion,
			AppVersion:   inspection.Software.AppVersion,
			SecurityPatch: inspection.Software.SecurityPatch,
			BuildID:       inspection.Software.BuildID,
		},
		Registration: HTTPRegistrationInfo{
			Status:           inspection.Registration.Status,
			FCMTokenValid:    inspection.Registration.FCMTokenValid,
			CommandSecretSet: inspection.Registration.CommandSecretSet,
		},
		Connection: HTTPConnectionInfo{
			WebSocketStatus: inspection.Connection.WebSocketStatus,
			FCMStatus:       inspection.Connection.FCMStatus,
			ClientIP:        inspection.Connection.ClientIP,
			Protocol:        inspection.Connection.Protocol,
		},
		Telemetry: HTTPTelemetryInfo{
			FramesToday:     inspection.Telemetry.FramesToday,
			AvgLatencyMs:    inspection.Telemetry.AvgLatencyMs,
			TotalBytesToday: inspection.Telemetry.TotalBytesToday,
			SessionsToday:   inspection.Telemetry.SessionsToday,
		},
	}

	// Convert timestamps to int64 ms
	if inspection.Registration.RegisteredAt != nil {
		resp.Registration.RegisteredAt = inspection.Registration.RegisteredAt.UnixMilli()
	}
	if inspection.Registration.FCMTokenRefreshedAt != nil {
		resp.Registration.FCMTokenRefreshedAt = inspection.Registration.FCMTokenRefreshedAt.UnixMilli()
	}
	if inspection.Connection.ConnectedAt != nil {
		resp.Connection.ConnectedAt = inspection.Connection.ConnectedAt.UnixMilli()
	}
	if inspection.Connection.LastSeen != nil {
		resp.Connection.LastSeen = inspection.Connection.LastSeen.UnixMilli()
	}
	if !inspection.Telemetry.LastTimestamp.IsZero() {
		resp.Telemetry.LastTimestamp = inspection.Telemetry.LastTimestamp.UnixMilli()
	}

	// Cache the response
	s.cacheInspection(cacheKey, resp)

	return resp, nil
}
	return resp, nil
}
	}
	if inspection.Connection.ConnectedAt != nil {
		resp.Connection.ConnectedAt = inspection.Connection.ConnectedAt.UnixMilli()
	}
	if inspection.Connection.LastSeen != nil {
		resp.Connection.LastSeen = inspection.Connection.LastSeen.UnixMilli()
	}
	if !inspection.Telemetry.LastTimestamp.IsZero() {
		resp.Telemetry.LastTimestamp = inspection.Telemetry.LastTimestamp.UnixMilli()
	}

	// Cache the response
	s.cacheInspection(cacheKey, resp)

	return resp, nil
}
// GetDeviceTimeline retrieves paginated timeline events for a device.
// Requires orgID for multi-tenant isolation.
func (s *Service) GetDeviceTimeline(ctx context.Context, imei string, req *TimelineRequest, orgID string) (*TimelineResponse, error) {
	// Verify device exists and belongs to organization
	dev, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil {
		return nil, err
	}
	if dev == nil {
		return nil, diagnostics.ErrDeviceNotFound
	}

	// Verify device belongs to organization
	if orgID != "" && dev.OrganizationID != orgID {
		return nil, diagnostics.ErrDeviceNotFound
	}
	}

	// Verify device belongs to organization
	if orgID != "" && dev.OrganizationID != orgID {
		return nil, diagnostics.ErrDeviceNotFound
	}
		return nil, diagnostics.ErrDeviceNotFound
	}

	// Verify device belongs to organization
	if orgID != "" && dev.OrganizationID != orgID {
		return nil, diagnostics.ErrDeviceNotFound
	}

	// Map frontend event type categories to actual event types per spec
	eventTypes := s.mapEventTypeCategory(req.EventType)

	// Build filter
	filter := &diagnostics.TimelineFilter{
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

	// Get timeline events for each mapped event type
	var allEvents []diagnostics.TimelineEvent

	// Build a set for efficient lookup when filtering
	eventTypeSet := make(map[string]bool)
	for _, et := range eventTypes {
		eventTypeSet[et] = true
	}

	// When eventTypes is nil/empty, fetch all events (no type filter)
	if len(eventTypes) == 0 {
		filter.EventType = "" // Clear any previous filter
		result, err := s.diagnosticsRepo.GetTimelineEvents(ctx, dev.ID, filter)
		if err != nil {
			return nil, err
		}
		allEvents = result.Events
	} else {
		// Fetch all events without type filter to get correct chronological order
		// then filter by type in memory
		filter.EventType = ""
		result, err := s.diagnosticsRepo.GetTimelineEvents(ctx, dev.ID, filter)
		if err != nil {
			return nil, err
		}

		// Filter by event types in memory to maintain correct chronological order
		for _, event := range result.Events {
			if eventTypeSet[string(event.Type)] {
				allEvents = append(allEvents, event)
			}
		}
	}

	// Check limit
	hasMore := len(allEvents) > filter.Limit
	if hasMore {
		allEvents = allEvents[:filter.Limit]
	}

	// Generate next cursor
	var nextCursor string
	if hasMore && len(allEvents) > 0 {
		last := allEvents[len(allEvents)-1]
		nextCursor = encodeCursor(last.Timestamp, last.ID)
	}

	// Convert to response DTOs
	events := make([]EventDTO, len(allEvents))
	for i, e := range allEvents {
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
			HasMore:    hasMore,
			NextCursor: nextCursor,
		},
	}, nil
}

// mapEventTypeCategory maps frontend category names to backend event types per spec.
func (s *Service) mapEventTypeCategory(category string) []string {
	switch category {
	case "telemetry":
		return []string{string(diagnostics.EventTypeTelemetry)}
	case "command":
		return []string{
			string(diagnostics.EventTypeCommandSent),
			string(diagnostics.EventTypeCommandAck),
			string(diagnostics.EventTypeCommandFailed),
		}
	case "connection":
		return []string{
			string(diagnostics.EventTypeConnectionOpen),
			string(diagnostics.EventTypeConnectionLost),
			string(diagnostics.EventTypeFCMFallback),
			string(diagnostics.EventTypeReconnected),
		}
	case "error":
		return []string{
			string(diagnostics.EventTypeError),
			string(diagnostics.EventTypeThresholdBreach),
		}
	case "all", "":
		// Return empty to indicate no filter (fetch all)
		return nil
	default:
		// Assume it's an exact event type match
		return []string{category}
	}
}

// encodeCursor encodes a pagination cursor.
func encodeCursor(t time.Time, id string) string {
	cursor := struct {
		T string `json:"t"`
		I string `json:"i"`
	}{t.Format(time.RFC3339Nano), id}
	data, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

// RecordDeviceEvent records a new device event.
func (s *Service) RecordDeviceEvent(ctx context.Context, deviceID string, eventType diagnostics.TimelineEventType, data map[string]any) error {
	event := &diagnostics.TimelineEvent{
		ID:        generateID(),
		DeviceID:  deviceID,
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	return s.diagnosticsRepo.RecordEvent(ctx, event)
}

// determineDeviceStatus determines the device registration status.
func (s *Service) determineDeviceStatus(dev *device.Device) string {
	// Use domain lifecycle method
	if dev.IsDeregistered() {
		return "deregistered"
	}
	if dev.RegisteredAt > 0 {
		// Device is registered
		if dev.LastSeen > 0 && time.Since(time.UnixMilli(dev.LastSeen)) > time.Duration(s.cfg.OfflineThresholdMinutes)*time.Minute {
			return "offline"
		}
		return "registered"
	}
	return "offline"
}

// determineWebSocketStatus determines the WebSocket connection status.
func (s *Service) determineWebSocketStatus(dev *device.Device, wsConnected bool) string {
	if wsConnected {
		return "connected"
	}
	if dev.LastSeen > 0 && time.Since(time.UnixMilli(dev.LastSeen)) < time.Duration(s.cfg.OfflineThresholdMinutes)*time.Minute {
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
		if time.Since(time.UnixMilli(*dev.FCMTokenRefreshedAt)) > time.Duration(s.cfg.FCMTokenExpiryDays)*24*time.Hour {
			return "invalid"
		}
	}
	return "valid"
}

// generateID generates a unique ID for events using crypto/rand.
func generateID() string {
	b := make([]byte, 16)
	if _, err := cryptoRand.Read(b); err != nil {
		// Fallback to timestamp-based if crypto/rand fails
		return fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), time.Now().UnixMicro())
// AuthorizationResponse represents device authorization result.
type AuthorizationResponse struct {
	Authorized bool
	Forbidden  bool
}

// VerifyDeviceOwnership checks if the operator owns the device and device belongs to org (org-scoped DOA check).
func (s *Service) VerifyDeviceOwnership(ctx context.Context, imei, operatorID, orgID string) *AuthorizationResponse {
	if operatorID == "" {
		// No operator context - unauthorized
		return &AuthorizationResponse{Authorized: false, Forbidden: false}
	}

	dev, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil || dev == nil {
		// Device not found - treat as forbidden (not unauthorized - they found it but don't own it)
		return &AuthorizationResponse{Authorized: false, Forbidden: true}
	}

	// Check org membership - device must belong to the organization
	if orgID != "" && dev.OrganizationID != orgID {
		return &AuthorizationResponse{Authorized: false, Forbidden: true}
	}

	if dev.OperatorID != operatorID {
		// Device exists but belongs to different operator
		return &AuthorizationResponse{Authorized: false, Forbidden: true}
	}

	return &AuthorizationResponse{Authorized: true, Forbidden: false}
}
		return &AuthorizationResponse{Authorized: false, Forbidden: true}
	}

	return &AuthorizationResponse{Authorized: true, Forbidden: false}
}
	}

	if dev.OperatorID != operatorID {
		// Device exists but belongs to different operator
		return &AuthorizationResponse{Authorized: false, Forbidden: true}
	}

	return &AuthorizationResponse{Authorized: true, Forbidden: false}
}
