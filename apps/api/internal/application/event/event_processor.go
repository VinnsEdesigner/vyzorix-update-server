// Package event provides event processing and broadcasting for real-time updates.
package event

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/event"
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
	repo        event.Repository
	deviceRepo  device.Repository
	broadcaster EventBroadcaster
	log         *slog.Logger
	thresholds  *ThresholdConfig
}

// ThresholdConfig holds event emission thresholds.
type ThresholdConfig struct {
	RiskScoreWarning  int
	RiskScoreCritical int
	ThermalWarning    float64
	ThermalCritical  float64
	BufferWarning    int
	BufferCritical   int
}

// DefaultThresholdConfig returns default threshold configuration.
func DefaultThresholdConfig() *ThresholdConfig {
	return &ThresholdConfig{
		RiskScoreWarning:  50,
		RiskScoreCritical: 80,
		ThermalWarning:    45.0,
		ThermalCritical:   60.0,
		BufferWarning:     30,
		BufferCritical:   10,
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

	return nil
}

// ProcessTelemetry processes telemetry data and emits events for threshold breaches.
func (p *Processor) ProcessTelemetry(ctx context.Context, deviceID string, telemetryData map[string]any) error {
	device, err := p.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		p.log.Warn("failed to get device for telemetry", "deviceId", deviceID, "err", err)
	}

	operatorID := ""
	if device != nil {
		operatorID = device.OperatorID
	}

	// Check for threshold breaches
	events := p.checkThresholds(deviceID, operatorID, telemetryData)

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

// checkThresholds checks telemetry against thresholds and returns breach events.
func (p *Processor) checkThresholds(deviceID, operatorID string, data map[string]any) []*event.Event {
	var events []*event.Event

	// Check risk score
	if riskScore, ok := data["riskScore"].(float64); ok {
		if riskScore >= float64(p.thresholds.RiskScoreCritical) {
			events = append(events, &event.Event{
				ID:         generateEventID(),
				DeviceID:   deviceID,
				OperatorID: operatorID,
				Type:       event.EventTypeRiskScoreAlert,
				Severity:   event.SeverityCritical,
				Timestamp:  time.Now(),
				Data:       map[string]any{"riskScore": riskScore, "threshold": p.thresholds.RiskScoreCritical},
				Source:     "server",
			})
		} else if riskScore >= float64(p.thresholds.RiskScoreWarning) {
			events = append(events, &event.Event{
				ID:         generateEventID(),
				DeviceID:   deviceID,
				OperatorID: operatorID,
				Type:       event.EventTypeThresholdBreach,
				Severity:   event.SeverityWarning,
				Timestamp:  time.Now(),
				Data:       map[string]any{"riskScore": riskScore, "threshold": p.thresholds.RiskScoreWarning},
				Source:     "server",
			})
		}
	}

	// Check thermal temp
	if thermalTemp, ok := data["thermalTemp"].(float64); ok {
		if thermalTemp >= p.thresholds.ThermalCritical {
			events = append(events, &event.Event{
				ID:         generateEventID(),
				DeviceID:   deviceID,
				OperatorID: operatorID,
				Type:       event.EventTypeThermalAlert,
				Severity:   event.SeverityCritical,
				Timestamp:  time.Now(),
				Data:       map[string]any{"thermalTemp": thermalTemp, "threshold": p.thresholds.ThermalCritical},
				Source:     "server",
			})
		} else if thermalTemp >= p.thresholds.ThermalWarning {
			events = append(events, &event.Event{
				ID:         generateEventID(),
				DeviceID:   deviceID,
				OperatorID: operatorID,
				Type:       event.EventTypeThresholdBreach,
				Severity:   event.SeverityWarning,
				Timestamp:  time.Now(),
				Data:       map[string]any{"thermalTemp": thermalTemp, "threshold": p.thresholds.ThermalWarning},
				Source:     "server",
			})
		}
	}

	// Check buffer level
	if bufferLevel, ok := data["bufferLevel"].(float64); ok {
		if bufferLevel <= float64(p.thresholds.BufferCritical) {
			events = append(events, &event.Event{
				ID:         generateEventID(),
				DeviceID:   deviceID,
				OperatorID: operatorID,
				Type:       event.EventTypeBufferLevelAlert,
				Severity:   event.SeverityCritical,
				Timestamp:  time.Now(),
				Data:       map[string]any{"bufferLevel": bufferLevel, "threshold": p.thresholds.BufferCritical},
				Source:     "server",
			})
		} else if bufferLevel <= float64(p.thresholds.BufferWarning) {
			events = append(events, &event.Event{
				ID:         generateEventID(),
				DeviceID:   deviceID,
				OperatorID: operatorID,
				Type:       event.EventTypeThresholdBreach,
				Severity:   event.SeverityWarning,
				Timestamp:  time.Now(),
				Data:       map[string]any{"bufferLevel": bufferLevel, "threshold": p.thresholds.BufferWarning},
				Source:     "server",
			})
		}
	}

	return events
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
