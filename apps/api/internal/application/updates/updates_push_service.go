package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/cache"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	ws "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/google/uuid"
)

// UpdateCommandType is the command sent to devices to trigger update check.
const UpdateCommandType = "CHECK_UPDATE"

// PushService handles push-related operations.
type PushService struct {
	repo      updates.Repository
	deviceSvc interface {
		GetDevice(ctx context.Context, deviceID string) (*device.Device, error)
	}
	hub            *ws.Hub
	fcmNotifier    fcm.Notifier
	commandSigner  *cryptohmac.CommandSigner
	commandService interface {
		SendCommand(ctx context.Context, req *dto.SendCommandRequest) (*dto.SendCommandResponse, error)
		MarkDelivered(ctx context.Context, commandID string) error
	}
	annotator RolloutAnnotator
	dashboardCache *cache.Section
	logger *slog.Logger
}

// SetDashboardCache wires the dashboard stats cache so rollout start
// invalidates the org's dashboard stats entry.
func (s *PushService) SetDashboardCache(c *cache.Section) {
	s.dashboardCache = c
}

// RolloutAnnotator marks update rollout milestones on the fleet timeline.
type RolloutAnnotator interface {
	AnnotateRollout(ctx context.Context, orgID, version, deviceCount string, initiatedBy string) error
}

// SetAnnotator wires the fleet timeline annotator for rollout milestones.
func (s *PushService) SetAnnotator(a RolloutAnnotator) {
	s.annotator = a
}

// NewPushService creates a new push service.
func NewPushService(
	repo updates.Repository,
	deviceSvc interface {
		GetDevice(ctx context.Context, deviceID string) (*device.Device, error)
	},
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
		commandSigner:  cryptohmac.NewCommandSigner(),
		commandService: commandService,
		logger:         logger,
	}
}

// PushUpdate pushes an update to devices.
// It validates the version exists, creates a push record, creates per-device push records,.
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
		ID:             uuid.NewString(),
		VersionID:      version.ID,
		OrganizationID: req.OrganizationID,
		InstallType:    updates.InstallType(req.InstallType),
		ScheduledAt:    req.ScheduledAt,
		Status:         updates.UpdateStatusPending,
		InitiatedBy:    initiatedBy,
		InitiatedAt:    now.UnixMilli(),
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
		// DeviceIDs contains ALL requested device IDs (per spec).
		deviceIDs     = make([]string, 0, len(req.DeviceIDs))
		failedDevices = make([]FailedDevice, 0)
		pendingCount  = 0
		sentCount     = 0
		acknowledged  = 0 // Acknowledged count is updated asynchronously by device responses.
		failedCount   = 0
	)

	for _, deviceID := range req.DeviceIDs {
		// Add to deviceIDs list immediately (per spec - all requested devices).
		deviceIDs = append(deviceIDs, deviceID)

		// Create per-device push record.
		devicePush := &updates.UpdatePushDevice{
			ID:         uuid.NewString(),
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

		// Increment pending count since device was successfully registered.
		pendingCount++

		// Dispatch the update command to the device.
		notifyErr := s.dispatchUpdateCommand(ctx, deviceID, push.ID, version.Version, version.APKFilename, version.SHA256, version.APKSize, payloadBytes)
		if notifyErr != nil {
			failedDevices = append(failedDevices, FailedDevice{
				DeviceID: deviceID,
				Reason:   notifyErr.Error(),
			})
			failedCount++  // Count as failed.
			pendingCount-- // No longer pending.
			// Mark device push as failed but keep it in the push so operators can see it.
			_ = s.repo.UpdatePushDeviceStatus(ctx, devicePush.ID, updates.DevicePushStatusFailed, notifyErr.Error())
			continue
		}

		// Command dispatched successfully - move from pending to sent.
		sentCount++
		pendingCount-- // No longer pending.
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

	if s.annotator != nil {
		if err := s.annotator.AnnotateRollout(ctx, req.OrganizationID, req.Version, strconv.Itoa(len(deviceIDs)), initiatedBy); err != nil {
			s.logger.Error("failed to annotate rollout", "push_id", push.ID, "error", err)
		}
	}
	if s.dashboardCache != nil {
		s.dashboardCache.Delete(req.OrganizationID)
	}

	return &PushUpdateResponse{
		PushID:        push.ID,
		Version:       req.Version,
		InstallType:   req.InstallType,
		InitiatedBy:   initiatedBy,
		Status:        string(push.Status),
		DeviceIDs:     deviceIDs, // All requested device IDs (per spec).
		FailedDevices: failedDevices,
		Devices: PushDeviceCounts{
			Total:        len(deviceIDs),
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
func (s *PushService) dispatchUpdateCommand(ctx context.Context, deviceID, pushID, version string, apkFilename, sha256 string, apkSize int64, payload []byte) error {
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

	// Sign the frame with the device's command secret (Domain B: server→device).
	dev, devErr := s.deviceSvc.GetDevice(ctx, deviceID)
	if devErr != nil {
		return fmt.Errorf("device not found: %w", devErr)
	}
	if dev.CommandSecretHash == "" {
		return fmt.Errorf("device %s has no command secret — re-registration required", deviceID)
	}
	if err := cryptohmac.SignCommandFrame(s.commandSigner, &frame, deviceID, dev.CommandSecretHash); err != nil {
		return fmt.Errorf("failed to sign update command frame: %w", err)
	}

	// Try WebSocket first if hub is available and device is online.
	if s.hub != nil && s.hub.Online(deviceID) {
		if sent := s.hub.Send(deviceID, frame); sent {
			// Device received via WSS — mark command as delivered.

			if err := s.commandService.MarkDelivered(ctx, cmdResp.CommandID); err != nil {
				s.logger.Error("failed to mark command as delivered",
					"commandId", cmdResp.CommandID,
					"deviceId", deviceID,
					"error", err)
			}
			s.logger.Info("update command sent via WSS",
				"deviceId", deviceID, "pushId", pushID, "version", version)
			return nil
		}
	}

	// Fall back to FCM for offline devices.
	if s.fcmNotifier != nil {
		if dev.FCMToken != "" {
			wake := fcm.SilentWake{
				Token:       dev.FCMToken,
				Command:     UpdateCommandType,
				DispatchID:  cmdResp.DispatchID,
				DeviceID:    deviceID,
				Priority:    "high",
				APKFilename: apkFilename,
				SHA256:      sha256,
				APKSize:     apkSize,
				// Device should prepend its server base URL to construct full download URL.
				DownloadURL: "/api/v1/apk/" + apkFilename,
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

// UpdateDeviceStatusByDispatch updates a push device status by dispatch_id (push_id) and device_id.
// This is called by device callbacks with the dispatch_id and device_id.
func (s *PushService) UpdateDeviceStatusByDispatch(ctx context.Context, dispatchID, deviceID string, status updates.DevicePushStatus, errorMsg string) error {
	// Find the push device record using push_id and device_id.
	devicePush, err := s.repo.GetPushDeviceByPushAndDevice(ctx, dispatchID, deviceID)
	if err != nil {
		if err == updates.ErrPushNotFound {
			return updates.ErrDeviceNotFound
		}
		return err
	}

	// Update the device status.
	if err := s.repo.UpdatePushDeviceStatus(ctx, devicePush.ID, status, errorMsg); err != nil {
		return err
	}

	// Check if all devices are done and update push status.
	// Don't fail the device callback if push completion check fails.
	if err := s.checkPushCompletion(ctx, dispatchID); err != nil {
		s.logger.Warn("failed to check push completion",
			"dispatchId", dispatchID, "err", err)
		// Don't return error - device status was already updated successfully.
	}

	return nil
}

// Hub returns the hub for direct use by tests.
func (s *PushService) Hub() *ws.Hub {
	return s.hub
}

// checkPushCompletion checks if all devices in a push have completed.
// and updates the push status accordingly.
func (s *PushService) checkPushCompletion(ctx context.Context, pushID string) error {
	devices, err := s.repo.GetPushDevicesByPushID(ctx, pushID)
	if err != nil {
		return err
	}

	var pendingCount, inProgressCount, completedCount, failedCount int
	for _, d := range devices {
		switch d.Status {
		case updates.DevicePushStatusPending, updates.DevicePushStatusSent:
			pendingCount++
		case updates.DevicePushStatusInProgress:
			inProgressCount++
		case updates.DevicePushStatusAcknowledged: // Legacy status.
			inProgressCount++
		case updates.DevicePushStatusCompleted:
			completedCount++
		case updates.DevicePushStatusFailed:
			failedCount++
		}
	}

	total := len(devices)
	allTerminal := pendingCount == 0 && inProgressCount == 0

	if allTerminal {
		// Determine push status based on device outcomes.
		var pushStatus updates.UpdateStatus
		if failedCount == 0 {
			pushStatus = updates.UpdateStatusCompleted
		} else if completedCount == 0 {
			pushStatus = updates.UpdateStatusFailed
		} else {
			// Mixed results - some succeeded, some failed.
			pushStatus = updates.UpdateStatusCompleted // Partial success.
		}

		if err := s.repo.UpdatePushStatus(ctx, pushID, pushStatus); err != nil {
			s.logger.Warn("failed to update push status",
				"pushId", pushID, "status", pushStatus, "err", err)
			return err
		}

		s.logger.Info("push completed",
			"pushId", pushID,
			"total", total,
			"completed", completedCount,
			"failed", failedCount,
			"status", pushStatus)
	}

	return nil
}
