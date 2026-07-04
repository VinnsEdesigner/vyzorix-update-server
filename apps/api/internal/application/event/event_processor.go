// Package event provides event processing and broadcasting for real-time updates.
package event

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/notification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/event"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
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
	operatorRepo       operator.Repository
	broadcaster        EventBroadcaster
	notificationSvc    *notification.Service
	log                *slog.Logger
	thresholds         *ThresholdConfig
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
		repo:        repo,
		deviceRepo:  deviceRepo,
		broadcaster: broadcaster,
		log:         log,
		thresholds:  DefaultThresholdConfig(),
	}
}

// SetOperatorRepo sets the operator repository for fetching per-operator thresholds.
func (p *Processor) SetOperatorRepo(repo operator.Repository) {
	p.operatorRepo = repo
}

// SetNotificationService sets the notification service for sending alerts.
func (p *Processor) SetNotificationService(svc *notification.Service) {
	p.notificationSvc = svc
}

// SetThresholds updates the threshold configuration.
func (p *Processor) SetThresholds(cfg *ThresholdConfig) {
	p.thresholds = cfg
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
// It fetches operator-specific thresholds to ensure each operator's custom settings are used.
func (p *Processor) ProcessTelemetry(ctx context.Context, deviceID string, telemetryData map[string]any) error {
	device, err := p.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		p.log.Warn("failed to get device for telemetry", "deviceId", deviceID, "err", err)
	}

	operatorID := ""
	if device != nil {
		operatorID = device.OperatorID
	}

	// Get operator-specific thresholds
	thresholds := p.thresholds // Start with defaults
	if operatorID != "" && p.operatorRepo != nil {
		opThresholds, err := p.operatorRepo.GetThresholds(ctx, operatorID)
		if err != nil {
			p.log.Warn("failed to get operator thresholds, using defaults", "operatorId", operatorID, "err", err)
		} else {
			// Convert operator thresholds to ThresholdConfig
			// Note: operator.Thresholds uses int for thermal values, ThresholdConfig uses float64
			thresholds = &ThresholdConfig{
				RiskScoreWarning:  opThresholds.RiskWarn,
				RiskScoreCritical: opThresholds.RiskCrit,
				ThermalWarning:    float64(opThresholds.ThermalWarn),
				ThermalCritical:   float64(opThresholds.ThermalCrit),
				BufferWarning:     opThresholds.BufferWarn,
				BufferCritical:    opThresholds.BufferCrit,
			}
		}
	}

	// Check for threshold breaches using operator-specific thresholds
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

	return nil
}

// checkThresholdsWithConfig checks telemetry against provided thresholds and returns breach events.
// This ensures operator-specific thresholds are used for accurate alerting.
func (p *Processor) checkThresholdsWithConfig(deviceID, operatorID string, data map[string]any, thresholds *ThresholdConfig) []*event.Event {
	var events []*event.Event

	// Check risk score
	if riskScore, ok := data["riskScore"].(float64); ok {
		if riskScore >= float64(thresholds.RiskScoreCritical) {
			events = append(events, &event.Event{
				ID:         generateEventID(),
				DeviceID:   deviceID,
				OperatorID: operatorID,
				Type:       event.EventTypeRiskScoreAlert,
				Severity:   event.SeverityCritical,
				Timestamp:  time.Now(),
				Data:       map[string]any{"riskScore": riskScore, "threshold": thresholds.RiskScoreCritical},
				Source:     "server",
			})
		} else if riskScore >= float64(thresholds.RiskScoreWarning) {
			events = append(events, &event.Event{
				ID:         generateEventID(),
				DeviceID:   deviceID,
				OperatorID: operatorID,
				Type:       event.EventTypeThresholdBreach,
				Severity:   event.SeverityWarning,
				Timestamp:  time.Now(),
				Data:       map[string]any{"riskScore": riskScore, "threshold": thresholds.RiskScoreWarning},
				Source:     "server",
			})
		}
	}

	// Check thermal temp
	if thermalTemp, ok := data["thermalTemp"].(float64); ok {
		if thermalTemp >= thresholds.ThermalCritical {
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
		} else if thermalTemp >= thresholds.ThermalWarning {
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

	// Check buffer level
	if bufferLevel, ok := data["bufferLevel"].(float64); ok {
		if bufferLevel <= float64(thresholds.BufferCritical) {
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
		} else if bufferLevel <= float64(thresholds.BufferWarning) {
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

	return events
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
func isThresholdBreachEvent(evtType event.EventType) bool {
	switch evtType {
	case event.EventTypeThresholdBreach, event.EventTypeRiskScoreAlert,
		event.EventTypeThermalAlert, event.EventTypeBufferLevelAlert:
		return true
	}
	return false
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
