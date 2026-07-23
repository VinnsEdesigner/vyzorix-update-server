// Package event provides event processing and broadcasting for real-time updates.
package event

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/notification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/event"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// EventBroadcaster broadcasts events to subscribed clients.
type EventBroadcaster interface {
	// BroadcastDeviceEvent sends an event to all subscribers of a device.
	BroadcastDeviceEvent(deviceID string, evt *event.Event) error
	// BroadcastOperatorEvent sends an event to all operators who own the device.
	BroadcastOperatorEvent(deviceID string, operatorID string, evt *event.Event) error
}

// Processor processes and emits real-time events.
type Processor struct {
	repo               event.Repository
	deviceRepo         device.Repository
	deviceSettingsRepo device.DeviceSettingsRepository
	orgSettingsRepo   organization.OrganizationSettingsRepository
	broadcaster        EventBroadcaster
	notificationSvc    *notification.Service
	log                *slog.Logger
	thresholds         *ThresholdConfig

	
	dedupMu          sync.RWMutex
	activeAlerts     map[string]time.Time // key: "deviceID:metric:type", value: last alert time
	dedupWindow      time.Duration        // cooldown period before new alert

	
	breachState    map[string]bool // key: "deviceID:metric", value: true if currently in breach
	hysteresisBand float64         // hysteresis band as percentage of threshold (0.1 = 10%)
}

// ThresholdConfig holds event emission thresholds.
type ThresholdConfig struct {
	RiskScoreWarning  int
	RiskScoreCritical int
	ThermalWarning    float64
	ThermalCritical   float64
	BufferWarning     int
	BufferCritical    int
}

// DefaultThresholdConfig returns default threshold configuration.
func DefaultThresholdConfig() *ThresholdConfig {
	return &ThresholdConfig{
		RiskScoreWarning:  50,
		RiskScoreCritical: 80,
		ThermalWarning:    45.0,
		ThermalCritical:   60.0,
		BufferWarning:     30,
		BufferCritical:    10,
	}
}

// NewProcessor creates a new event processor.
func NewProcessor(repo event.Repository, deviceRepo device.Repository, broadcaster EventBroadcaster, log *slog.Logger) *Processor {
	return &Processor{
		repo:          repo,
		deviceRepo:    deviceRepo,
		broadcaster:   broadcaster,
		log:           log,
		thresholds:    DefaultThresholdConfig(),
		activeAlerts:  make(map[string]time.Time),
		dedupWindow:   5 * time.Minute, // Default 5-minute dedup window
		breachState:   make(map[string]bool),
		hysteresisBand: 0.1, 
	}
}

// SetDeviceSettingsRepo sets the device settings repository for hierarchical threshold resolution.
func (p *Processor) SetDeviceSettingsRepo(repo device.DeviceSettingsRepository) {
	p.deviceSettingsRepo = repo
}

// SetOrgSettingsRepo sets the organization settings repository for hierarchical threshold resolution.
func (p *Processor) SetOrgSettingsRepo(repo organization.OrganizationSettingsRepository) {
	p.orgSettingsRepo = repo
}

// SetNotificationService sets the notification service for sending alerts.
func (p *Processor) SetNotificationService(svc *notification.Service) {
	p.notificationSvc = svc
}

// SetThresholds updates the threshold configuration.
func (p *Processor) SetThresholds(cfg *ThresholdConfig) {
	p.thresholds = cfg
}

// resolveThresholds resolves thresholds using hierarchical resolution:
// device settings → organization settings → defaults
func (p *Processor) resolveThresholds(ctx context.Context, deviceID, orgID string) *ThresholdConfig {
	// Start with defaults
	result := DefaultThresholdConfig()

	// Get organization thresholds
	if orgID != "" && p.orgSettingsRepo != nil {
		orgSettings, err := p.orgSettingsRepo.FindByOrganizationID(ctx, orgID)
		if err == nil && orgSettings != nil && orgSettings.DefaultThresholds != nil {
			result.RiskScoreWarning = orgSettings.DefaultThresholds.RiskWarn
			result.RiskScoreCritical = orgSettings.DefaultThresholds.RiskCrit
			result.ThermalWarning = float64(orgSettings.DefaultThresholds.ThermalWarn)
			result.ThermalCritical = float64(orgSettings.DefaultThresholds.ThermalCrit)
			result.BufferWarning = orgSettings.DefaultThresholds.BufferWarn
			result.BufferCritical = orgSettings.DefaultThresholds.BufferCrit
		}
	}

	// Override with device-specific thresholds
	if deviceID != "" && p.deviceSettingsRepo != nil {
		deviceSettings, err := p.deviceSettingsRepo.FindByDeviceIMEI(ctx, deviceID)
		if err == nil && deviceSettings != nil && deviceSettings.HasThresholds() {
			if deviceSettings.Thresholds.RiskWarn != 0 {
				result.RiskScoreWarning = deviceSettings.Thresholds.RiskWarn
			}
			if deviceSettings.Thresholds.RiskCrit != 0 {
				result.RiskScoreCritical = deviceSettings.Thresholds.RiskCrit
			}
			if deviceSettings.Thresholds.ThermalWarn != 0 {
				result.ThermalWarning = float64(deviceSettings.Thresholds.ThermalWarn)
			}
			if deviceSettings.Thresholds.ThermalCrit != 0 {
				result.ThermalCritical = float64(deviceSettings.Thresholds.ThermalCrit)
			}
			if deviceSettings.Thresholds.BufferWarn != 0 {
				result.BufferWarning = deviceSettings.Thresholds.BufferWarn
			}
			if deviceSettings.Thresholds.BufferCrit != 0 {
				result.BufferCritical = deviceSettings.Thresholds.BufferCrit
			}
		}
	}

	return result
}

// ProcessDeviceConnected handles device connection events.
func (p *Processor) ProcessDeviceConnected(ctx context.Context, deviceID string, metadata map[string]any) error {
	// Get device info for operator ID
	device, err := p.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		p.log.Warn("failed to get device for connected event", "deviceId", deviceID, "err", err)
	}

	operatorID := ""
	if device != nil {
		operatorID = device.OperatorID
	}

	evt := &event.Event{
		ID:         generateEventID(),
		DeviceID:   deviceID,
		OperatorID: operatorID,
		Type:       event.EventTypeDeviceConnected,
		Severity:   event.SeverityInfo,
		Timestamp:  time.Now(),
		Data:       metadata,
		Source:     "server",
	}

	if err := p.repo.Store(ctx, evt); err != nil {
		p.log.Error("failed to store device connected event", "deviceId", deviceID, "err", err)
	}

	// Broadcast to subscribers
	if p.broadcaster != nil {
		if err := p.broadcaster.BroadcastDeviceEvent(deviceID, evt); err != nil {
			p.log.Warn("failed to broadcast device connected event", "deviceId", deviceID, "err", err)
		}
		if operatorID != "" {
			if err := p.broadcaster.BroadcastOperatorEvent(deviceID, operatorID, evt); err != nil {
				p.log.Warn("failed to broadcast operator event", "deviceId", deviceID, "err", err)
			}
		}
	}

	// Send notification for device online event
	if p.notificationSvc != nil && operatorID != "" {
		notifData := notification.EventData{
			EventType:  notification.EventTypeDeviceOnline,
			DeviceID:   deviceID,
			DeviceName: getDeviceName(device),
			OperatorID: operatorID,
			Timestamp:  evt.Timestamp,
		}
		if err := p.notificationSvc.SendNotification(ctx, notifData); err != nil {
			p.log.Warn("failed to send device online notification", "deviceId", deviceID, "err", err)
		}
	}

	return nil
}

// ProcessDeviceDisconnected handles device disconnection events.
func (p *Processor) ProcessDeviceDisconnected(ctx context.Context, deviceID string, reason string, metadata map[string]any) error {
	device, err := p.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		p.log.Warn("failed to get device for disconnected event", "deviceId", deviceID, "err", err)
	}

	operatorID := ""
	if device != nil {
		operatorID = device.OperatorID
	}

	evt := &event.Event{
		ID:         generateEventID(),
		DeviceID:   deviceID,
		OperatorID: operatorID,
		Type:       event.EventTypeDeviceDisconnected,
		Severity:   event.SeverityWarning,
		Timestamp:  time.Now(),
		Data:       metadata,
		Source:     "server",
	}
	if metadata == nil {
		evt.Data = make(map[string]any)
	}
	evt.Data["reason"] = reason

	if err := p.repo.Store(ctx, evt); err != nil {
		p.log.Error("failed to store device disconnected event", "deviceId", deviceID, "err", err)
	}

	if p.broadcaster != nil {
		if err := p.broadcaster.BroadcastDeviceEvent(deviceID, evt); err != nil {
			p.log.Warn("failed to broadcast device disconnected event", "deviceId", deviceID, "err", err)
		}
		if operatorID != "" {
			if err := p.broadcaster.BroadcastOperatorEvent(deviceID, operatorID, evt); err != nil {
				p.log.Warn("failed to broadcast operator event", "deviceId", deviceID, "err", err)
			}
		}
	}

	// Send notification for device offline event
	if p.notificationSvc != nil && operatorID != "" {
		notifData := notification.EventData{
			EventType:  notification.EventTypeDeviceOffline,
			DeviceID:   deviceID,
			DeviceName: getDeviceName(device),
			OperatorID: operatorID,
			Timestamp:  evt.Timestamp,
		}
		if err := p.notificationSvc.SendNotification(ctx, notifData); err != nil {
			p.log.Warn("failed to send device offline notification", "deviceId", deviceID, "err", err)
		}
	}

	return nil
}

// ProcessTelemetry processes telemetry data and emits events for threshold breaches.
// It uses hierarchical threshold resolution: device settings → organization settings → defaults.
func (p *Processor) ProcessTelemetry(ctx context.Context, deviceID string, telemetryData map[string]any) error {
	device, err := p.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		p.log.Warn("failed to get device for telemetry", "deviceId", deviceID, "err", err)
	}

	operatorID := ""
	orgID := ""
	if device != nil {
		operatorID = device.OperatorID
		orgID = device.OrganizationID
	}

	// Get thresholds using hierarchical resolution: device → org → default
	thresholds := p.resolveThresholds(ctx, deviceID, orgID)

	// Check for threshold breaches using resolved thresholds
	events := p.checkThresholdsWithConfig(deviceID, operatorID, telemetryData, thresholds)

	// Store and broadcast events
	for _, evt := range events {
		if err := p.repo.Store(ctx, evt); err != nil {
			p.log.Error("failed to store threshold breach event", "deviceId", deviceID, "err", err)
		}

		if p.broadcaster != nil {
			if err := p.broadcaster.BroadcastDeviceEvent(deviceID, evt); err != nil {
				p.log.Warn("failed to broadcast threshold event", "deviceId", deviceID, "err", err)
			}
			if operatorID != "" {
				if err := p.broadcaster.BroadcastOperatorEvent(deviceID, operatorID, evt); err != nil {
					p.log.Warn("failed to broadcast operator threshold event", "deviceId", deviceID, "err", err)
				}
			}
		}

		// Send notification for threshold breach (all types: breach, risk, thermal, buffer)
		if p.notificationSvc != nil && operatorID != "" && isThresholdBreachEvent(evt.Type) {
			// Extract alert details from event data
			alertType := ""
			currentValue := ""
			thresholdValue := ""

			if v, ok := evt.Data["riskScore"]; ok {
				alertType = "Risk Score"
				currentValue = fmt.Sprintf("%.0f", v)
			} else if v, ok := evt.Data["thermalTemp"]; ok {
				alertType = "Thermal Temperature"
				currentValue = fmt.Sprintf("%.1f°C", v)
			} else if v, ok := evt.Data["bufferLevel"]; ok {
				alertType = "Buffer Level"
				currentValue = fmt.Sprintf("%.0f%%", v)
			}

			if v, ok := evt.Data["threshold"]; ok {
				thresholdValue = fmt.Sprintf("%v", v)
			}

			notifData := notification.EventData{
				EventType:    notification.EventTypeThresholdBreach,
				DeviceID:     deviceID,
				DeviceName:   getDeviceName(device),
				OperatorID:   operatorID,
				AlertType:    alertType,
				CurrentValue: currentValue,
				Threshold:    thresholdValue,
				Timestamp:    evt.Timestamp,
			}
			if err := p.notificationSvc.SendNotification(ctx, notifData); err != nil {
				p.log.Warn("failed to send threshold breach notification", "deviceId", deviceID, "err", err)
			}
		}
	}

	return nil
}

// ProcessCommandEvent processes command lifecycle events.
func (p *Processor) ProcessCommandEvent(ctx context.Context, deviceID string, commandType event.EventType, metadata map[string]any) error {
	device, err := p.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		p.log.Warn("failed to get device for command event", "deviceId", deviceID, "err", err)
	}

	operatorID := ""
	if device != nil {
		operatorID = device.OperatorID
	}

	evt := &event.Event{
		ID:         generateEventID(),
		DeviceID:   deviceID,
		OperatorID: operatorID,
		Type:       commandType,
		Severity:   event.GetSeverity(commandType),
		Timestamp:  time.Now(),
		Data:       metadata,
		Source:     "server",
	}

	if err := p.repo.Store(ctx, evt); err != nil {
		p.log.Error("failed to store command event", "deviceId", deviceID, "err", err)
	}

	if p.broadcaster != nil {
		if err := p.broadcaster.BroadcastDeviceEvent(deviceID, evt); err != nil {
			p.log.Warn("failed to broadcast command event", "deviceId", deviceID, "err", err)
		}
		if operatorID != "" {
			if err := p.broadcaster.BroadcastOperatorEvent(deviceID, operatorID, evt); err != nil {
				p.log.Warn("failed to broadcast operator command event", "deviceId", deviceID, "err", err)
			}
		}
	}

	// Send notification for command failures
	if p.notificationSvc != nil && operatorID != "" && commandType == event.EventTypeCommandFailed {
		commandName := ""
		failureReason := ""
		if v, ok := metadata["commandName"].(string); ok {
			commandName = v
		}
		if v, ok := metadata["reason"].(string); ok {
			failureReason = v
		} else if v, ok := metadata["error"].(string); ok {
			failureReason = v
		}

		notifData := notification.EventData{
			EventType:      notification.EventTypeCommandFailed,
			DeviceID:       deviceID,
			DeviceName:     getDeviceName(device),
			OperatorID:     operatorID,
			CommandName:    commandName,
			FailureReason: failureReason,
			Timestamp:      evt.Timestamp,
		}
		if err := p.notificationSvc.SendNotification(ctx, notifData); err != nil {
			p.log.Warn("failed to send command failed notification", "deviceId", deviceID, "err", err)
		}
	}

	return nil
}

// ProcessError processes error events.
func (p *Processor) ProcessError(ctx context.Context, deviceID string, errMsg string, metadata map[string]any) error {
	device, err := p.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		p.log.Warn("failed to get device for error event", "deviceId", deviceID, "err", err)
	}

	operatorID := ""
	if device != nil {
		operatorID = device.OperatorID
	}

	if metadata == nil {
		metadata = make(map[string]any)
	}
	metadata["error"] = errMsg

	evt := &event.Event{
		ID:         generateEventID(),
		DeviceID:   deviceID,
		OperatorID: operatorID,
		Type:       event.EventTypeError,
		Severity:   event.SeverityCritical,
		Timestamp:  time.Now(),
		Data:       metadata,
		Source:     "server",
	}

	if err := p.repo.Store(ctx, evt); err != nil {
		p.log.Error("failed to store error event", "deviceId", deviceID, "err", err)
	}

	if p.broadcaster != nil {
		if err := p.broadcaster.BroadcastDeviceEvent(deviceID, evt); err != nil {
			p.log.Warn("failed to broadcast error event", "deviceId", deviceID, "err", err)
		}
		if operatorID != "" {
			if err := p.broadcaster.BroadcastOperatorEvent(deviceID, operatorID, evt); err != nil {
				p.log.Warn("failed to broadcast operator error event", "deviceId", deviceID, "err", err)
			}
		}
	}

	// Send notification for error events
	if p.notificationSvc != nil && operatorID != "" {
		notifData := notification.EventData{
			EventType:    notification.EventTypeError,
			DeviceID:     deviceID,
			DeviceName:   getDeviceName(device),
			OperatorID:   operatorID,
			ErrorMessage: errMsg,
			Timestamp:    evt.Timestamp,
		}
		if err := p.notificationSvc.SendNotification(ctx, notifData); err != nil {
			p.log.Warn("failed to send error notification", "deviceId", deviceID, "err", err)
		}
	}

	return nil
}

// checkThresholdsWithConfig checks telemetry against provided thresholds and returns breach events.
// This ensures operator-specific thresholds are used for accurate alerting.
func (p *Processor) checkThresholdsWithConfig(deviceID, operatorID string, data map[string]any, thresholds *ThresholdConfig) []*event.Event {
	var events []*event.Event

	// Calculate server-side risk score from corroborating telemetry fields
	serverRiskScore := calculateServerSideRiskScore(data, thresholds)

	// Check risk score - use server-calculated score for threshold checks
	// Clamp device-reported riskScore to [0, 100] and emit security event if divergence is significant
	if deviceRiskScore, ok := data["riskScore"].(float64); ok {
		// Clamp to valid range
		if deviceRiskScore < 0 {
			deviceRiskScore = 0
		} else if deviceRiskScore > 100 {
			deviceRiskScore = 100
		}

		// Check for significant divergence (> 20 points) between device-reported and server-calculated
		divergence := math.Abs(deviceRiskScore - serverRiskScore)
		if divergence > 20 {
			events = append(events, &event.Event{
				ID:         generateEventID(),
				DeviceID:   deviceID,
				OperatorID: operatorID,
				Type:       event.EventTypeError,
				Severity:   event.SeverityWarning,
				Timestamp:  time.Now(),
				Data:       map[string]any{"riskScoreDivergence": divergence, "deviceReported": deviceRiskScore, "serverCalculated": serverRiskScore, "issue": "risk_score_manipulation_detected"},
				Source:     "server",
			})
		}

		// Use the server-calculated risk score for threshold checks
		if serverRiskScore >= float64(thresholds.RiskScoreCritical) {
			
			if p.shouldSendAlert(deviceID, "riskScore", event.EventTypeRiskScoreAlert) {
				events = append(events, &event.Event{
					ID:         generateEventID(),
					DeviceID:   deviceID,
					OperatorID: operatorID,
					Type:       event.EventTypeRiskScoreAlert,
					Severity:   event.SeverityCritical,
					Timestamp:  time.Now(),
					Data:       map[string]any{"riskScore": serverRiskScore, "threshold": thresholds.RiskScoreCritical, "deviceReported": deviceRiskScore},
					Source:     "server",
				})
			}
		} else if serverRiskScore >= float64(thresholds.RiskScoreWarning) {
			
			if p.shouldSendAlert(deviceID, "riskScore", event.EventTypeThresholdBreach) {
				events = append(events, &event.Event{
					ID:         generateEventID(),
					DeviceID:   deviceID,
					OperatorID: operatorID,
					Type:       event.EventTypeThresholdBreach,
					Severity:   event.SeverityWarning,
					Timestamp:  time.Now(),
					Data:       map[string]any{"riskScore": serverRiskScore, "threshold": thresholds.RiskScoreWarning, "deviceReported": deviceRiskScore},
					Source:     "server",
				})
			}
		}
	}

	// Check thermal temp
	if thermalTemp, ok := data["thermalTemp"].(float64); ok {
		if thermalTemp >= thresholds.ThermalCritical {
			
			if p.shouldSendAlert(deviceID, "thermal", event.EventTypeThermalAlert) {
				events = append(events, &event.Event{
					ID:         generateEventID(),
					DeviceID:   deviceID,
					OperatorID: operatorID,
					Type:       event.EventTypeThermalAlert,
					Severity:   event.SeverityCritical,
					Timestamp:  time.Now(),
					Data:       map[string]any{"thermalTemp": thermalTemp, "threshold": thresholds.ThermalCritical},
					Source:     "server",
				})
			}
		} else if thermalTemp >= thresholds.ThermalWarning {
			
			if p.shouldSendAlert(deviceID, "thermal", event.EventTypeThresholdBreach) {
				events = append(events, &event.Event{
					ID:         generateEventID(),
					DeviceID:   deviceID,
					OperatorID: operatorID,
					Type:       event.EventTypeThresholdBreach,
					Severity:   event.SeverityWarning,
					Timestamp:  time.Now(),
					Data:       map[string]any{"thermalTemp": thermalTemp, "threshold": thresholds.ThermalWarning},
					Source:     "server",
				})
			}
		}
	}

	// Check buffer level
	if bufferLevel, ok := data["bufferLevel"].(float64); ok {
		if bufferLevel <= float64(thresholds.BufferCritical) {
			
			if p.shouldSendAlert(deviceID, "buffer", event.EventTypeBufferLevelAlert) {
				events = append(events, &event.Event{
					ID:         generateEventID(),
					DeviceID:   deviceID,
					OperatorID: operatorID,
					Type:       event.EventTypeBufferLevelAlert,
					Severity:   event.SeverityCritical,
					Timestamp:  time.Now(),
					Data:       map[string]any{"bufferLevel": bufferLevel, "threshold": thresholds.BufferCritical},
					Source:     "server",
				})
			}
		} else if bufferLevel <= float64(thresholds.BufferWarning) {
			
			if p.shouldSendAlert(deviceID, "buffer", event.EventTypeThresholdBreach) {
				events = append(events, &event.Event{
					ID:         generateEventID(),
					DeviceID:   deviceID,
					OperatorID: operatorID,
					Type:       event.EventTypeThresholdBreach,
					Severity:   event.SeverityWarning,
					Timestamp:  time.Now(),
					Data:       map[string]any{"bufferLevel": bufferLevel, "threshold": thresholds.BufferWarning},
					Source:     "server",
				})
			}
		}
	}

	return events
}

// calculateServerSideRiskScore derives a risk score from corroborating telemetry fields.
// This prevents devices from manipulating their reported riskScore to suppress or flood alerts.
func calculateServerSideRiskScore(data map[string]any, thresholds *ThresholdConfig) float64 {
	var score float64

	hasThermal := false
	hasBuffer := false

	// Thermal contribution: 0-40 points (normalized from 0 to ThermalCritical*1.5)
	if thermalTemp, ok := data["thermalTemp"].(float64); ok {
		hasThermal = true
		thermalScore := (thermalTemp / (thresholds.ThermalCritical * 1.5)) * 40
		if thermalScore > 40 {
			thermalScore = 40
		}
		score += thermalScore
	}

	// Buffer contribution: 0-40 points (inverse - low buffer is high risk)
	if bufferLevel, ok := data["bufferLevel"].(float64); ok {
		hasBuffer = true
		// Buffer 100% = 0 risk, buffer 0% = 40 risk
		bufferScore := (100 - bufferLevel) / 100 * 40
		score += bufferScore
	}

	// Additional factor: if both thermal and buffer are present, add 20 points for combined risk
	if hasThermal && hasBuffer {
		score += 20
	}

	// Clamp to [0, 100]
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// getDeviceName extracts the device name from a device entity.
func getDeviceName(d *device.Device) string {
	if d == nil {
		return ""
	}
	// Try to get name from metadata or use ID as fallback
	if d.DeviceName != "" {
		return d.DeviceName
	}
	return d.ID
}

// isThresholdBreachEvent returns true if the event type is a threshold breach or alert.
//
func isThresholdBreachEvent(evtType event.EventType) bool {
	switch evtType {
	case event.EventTypeThresholdBreach, event.EventTypeRiskScoreAlert,
		event.EventTypeThermalAlert, event.EventTypeBufferLevelAlert:
		return true
	default:
		return false
	}
}
// shouldSendAlert checks if an alert should be sent based on deduplication.
// Returns false if an alert was recently sent for the same device/metric/type combination.

func (p *Processor) shouldSendAlert(deviceID, metric string, evtType event.EventType) bool {
	key := fmt.Sprintf("%s:%s:%s", deviceID, metric, evtType)

	p.dedupMu.Lock()
	defer p.dedupMu.Unlock()

	now := time.Now()
	if lastTime, exists := p.activeAlerts[key]; exists {
		// Check if we're still within the dedup window
		if now.Sub(lastTime) < p.dedupWindow {
			return false // Duplicate alert, suppress
		}
	}

	// Mark this alert as sent
	p.activeAlerts[key] = now

	// Periodic cleanup of old entries to prevent memory leaks
	if len(p.activeAlerts) > 10000 {
		p.cleanupActiveAlertsLocked(now)
	}

	return true
}

// cleanupActiveAlertsLocked removes expired entries from activeAlerts map.
// Caller must hold p.dedupMu.
func (p *Processor) cleanupActiveAlertsLocked(now time.Time) {
	cutoff := now.Add(-p.dedupWindow * 2)
	for key, lastTime := range p.activeAlerts {
		if lastTime.Before(cutoff) {
			delete(p.activeAlerts, key)
		}
	}
}

// generateEventID generates a unique event ID.
func generateEventID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "evt_" + hex.EncodeToString([]byte(time.Now().String()))
	}
	return "evt_" + hex.EncodeToString(b)
}

// EventToJSON converts an event to JSON bytes.
func EventToJSON(evt *event.Event) ([]byte, error) {
	return json.Marshal(evt)
}
