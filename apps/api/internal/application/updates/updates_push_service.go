package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	ws "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
)

// UpdateCommandType is the command sent to devices to trigger update check.
const UpdateCommandType = "CHECK_UPDATE"

// PushService handles push-related operations.
type PushService struct {
	repo           updates.Repository
	deviceSvc      interface{ GetDevice(ctx context.Context, deviceID string) (*device.Device, error) }
	hub            *ws.Hub
	fcmNotifier    fcm.Notifier
	commandService interface {
		SendCommand(ctx context.Context, req *dto.SendCommandRequest) (*dto.SendCommandResponse, error)
		MarkDelivered(ctx context.Context, commandID string) error
	}
	logger *slog.Logger
}

// NewPushService creates a new push service.
func NewPushService(
	repo updates.Repository,
	deviceSvc interface{ GetDevice(ctx context.Context, deviceID string) (*device.Device, error) },
	hub *ws.Hub,
	fcmNotifier fcm.Notifier,
	commandService interface {
		SendCommand(ctx context.Context, req *dto.SendCommandRequest) (*dto.SendCommandResponse, error)
		MarkDelivered(ctx context.Context, commandID string) error
	},
	logger *slog.Logger,
) *PushService {
	if logger == nil {
		logger = slog.Default()
	}
	return &PushService{
		repo:           repo,
		deviceSvc:      deviceSvc,
		hub:            hub,
		fcmNotifier:    fcmNotifier,
		commandService: commandService,
		logger:         logger,
	}
}

// PushUpdate pushes an update to devices.
// It validates the version exists, creates a push record, creates per-device push records,
// and dispatches update commands to each device via WebSocket (if online) or FCM (if offline).
func (s *PushService) PushUpdate(ctx context.Context, req *PushUpdateRequest, initiatedBy string) (*PushUpdateResponse, error) {
	if len(req.DeviceIDs) == 0 {
		return nil, ErrBadRequest
	}

	// Resolve version to get APK metadata for the update command payload.
	version, err := s.repo.GetVersionByVersion(ctx, req.Version)
	if err != nil {
		if err == updates.ErrVersionNotFound {
			return nil, ErrVersionNotFound
		}
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	now := time.Now()
	push := &updates.UpdatePush{
		VersionID:    version.ID,
		InstallType: updates.InstallType(req.InstallType),
		ScheduledAt:  req.ScheduledAt,
		Status:       updates.UpdateStatusPending,
		InitiatedBy:  initiatedBy,
		InitiatedAt:  now.UnixMilli(),
	}

	if err := s.repo.CreatePush(ctx, push); err != nil {
		return nil, fmt.Errorf("failed to create push: %w", err)
	}

	// Build the update payload sent to each device via command args.
	updatePayload := map[string]interface{}{
		"version":     version.Version,
		"apkFilename": version.APKFilename,
		"sha256":      version.SHA256,
		"apkSize":     version.APKSize,
		"installType": req.InstallType,
	}
	payloadBytes, payloadErr := json.Marshal(updatePayload)
	if payloadErr != nil {
		return nil, fmt.Errorf("failed to marshal update payload: %w", payloadErr)
	}

	var (
		deviceIDs      = make([]string, 0, len(req.DeviceIDs))
		failedDevices  = make([]FailedDevice, 0)
		pendingCount   = 0
		sentCount      = 0
		acknowledged   = 0
		failedCount    = 0
	)

	for _, deviceID := range req.DeviceIDs {
		// Create per-device push record.
		devicePush := &updates.UpdatePushDevice{
			PushID:     push.ID,
			DeviceID:   deviceID,
			Status:     updates.DevicePushStatusPending,
			RetryCount: 0,
			CreatedAt:  now.UnixMilli(),
			UpdatedAt:  now.UnixMilli(),
		}
		if err := s.repo.CreatePushDevice(ctx, devicePush); err != nil {
			failedDevices = append(failedDevices, FailedDevice{
				DeviceID: deviceID,
				Reason:   "failed to register device for update",
			})
			failedCount++
			continue
		}

		// Dispatch the update command to the device.
		notifyErr := s.dispatchUpdateCommand(ctx, deviceID, push.ID, version.Version, payloadBytes)
		if notifyErr != nil {
			failedDevices = append(failedDevices, FailedDevice{
				DeviceID: deviceID,
				Reason:   notifyErr.Error(),
			})
			failedCount++
			// Mark device push as failed but keep it in the push so operators can see it.
			_ = s.repo.UpdatePushDeviceStatus(ctx, devicePush.ID, updates.DevicePushStatusFailed, notifyErr.Error())
			continue
		}

		// Command dispatched successfully.
		deviceIDs = append(deviceIDs, deviceID)
		sentCount++
		_ = s.repo.UpdatePushDeviceStatus(ctx, devicePush.ID, updates.DevicePushStatusSent, "")
	}

	// If ALL devices failed to register, return an error.
	if len(deviceIDs) == 0 && len(failedDevices) > 0 {
		return nil, &ServiceError{
			Code:    "all_devices_failed",
			Message: fmt.Sprintf("all %d devices failed to be registered for update", len(failedDevices)),
			Status:  http.StatusBadRequest,
			Details: failedDevices,
		}
	}

	return &PushUpdateResponse{
		PushID:        push.ID,
		Version:       req.Version,
		InstallType:   req.InstallType,
		InitiatedBy:   initiatedBy,
		Status:        string(push.Status),
		DeviceIDs:     deviceIDs,
		FailedDevices: failedDevices,
		Devices: PushDeviceCounts{
			Total:        len(deviceIDs) + len(failedDevices),
			Pending:      pendingCount,
			Sent:         sentCount,
			Acknowledged: acknowledged,
			Failed:       failedCount,
		},
		InitiatedAt: push.InitiatedAt,
	}, nil
}

// dispatchUpdateCommand sends the update command to a device.
// It first tries WebSocket (if device is online), then falls back to FCM.
// The command is also always persisted via CommandService for reliability.
func (s *PushService) dispatchUpdateCommand(ctx context.Context, deviceID, pushID, version string, payload []byte) error {
	// Build the command frame.
	frame := command.CommandFrame{
		Type:       UpdateCommandType,
		Command:    UpdateCommandType,
		DispatchID: pushID,
		Args:       payload,
		Timestamp:  time.Now().UnixMilli(),
	}

	// Always persist the command first for reliability (idempotent via dispatchID = pushID).
	cmdReq := &dto.SendCommandRequest{
		DeviceID:   deviceID,
		Command:    UpdateCommandType,
		DispatchID: pushID,
		Args:       json.RawMessage(payload),
	}
	cmdResp, cmdErr := s.commandService.SendCommand(ctx, cmdReq)
	if cmdErr != nil {
		return fmt.Errorf("failed to persist update command: %w", cmdErr)
	}

	// Use the actual command dispatch ID from the persisted command.
	frame.DispatchID = cmdResp.DispatchID

	// Try WebSocket first if hub is available and device is online.
	if s.hub != nil && s.hub.Online(deviceID) {
		if sent := s.hub.Send(deviceID, frame); sent {
			// Device received via WSS — mark command as delivered.
			_ = s.commandService.MarkDelivered(ctx, cmdResp.CommandID)
			s.logger.Info("update command sent via WSS",
				"deviceId", deviceID, "pushId", pushID, "version", version)
			return nil
		}
	}

	// Fall back to FCM for offline devices.
	if s.fcmNotifier != nil {
		dev, devErr := s.deviceSvc.GetDevice(ctx, deviceID)
		if devErr != nil {
			return fmt.Errorf("device not found: %w", devErr)
		}
		if dev.FCMToken != "" {
			wake := fcm.SilentWake{
				Token:       dev.FCMToken,
				Command:     UpdateCommandType,
				DispatchID:  cmdResp.DispatchID,
				DeviceID:    deviceID,
				Priority:    "high",
			}
			if fcmErr := s.fcmNotifier.SendSilentWake(ctx, wake); fcmErr != nil {
				s.logger.Warn("FCM wake failed for update",
					"deviceId", deviceID, "pushId", pushID, "err", fcmErr)
				// FCM failure is not fatal — command is queued in DB.
				// Device will pick it up on next poll or reconnect.
				return nil
			}
			s.logger.Info("update command sent via FCM",
				"deviceId", deviceID, "pushId", pushID, "version", version)
			return nil
		}
	}

	// Neither WSS nor FCM available — command is in DB, device will poll.
	s.logger.Info("update command queued (no live channel)",
		"deviceId", deviceID, "pushId", pushID, "version", version)
	return nil
}

// UpdatePushDeviceStatus is an alias for the repository method exposed for testing.
func (s *PushService) UpdatePushDeviceStatus(ctx context.Context, id string, status updates.DevicePushStatus, errorMsg string) error {
	return s.repo.UpdatePushDeviceStatus(ctx, id, status, errorMsg)
}

// Hub returns the hub for direct use by tests.
func (s *PushService) Hub() *ws.Hub {
	return s.hub
}
