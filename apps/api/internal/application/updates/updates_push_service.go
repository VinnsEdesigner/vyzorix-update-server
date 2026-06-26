package updates

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// PushService handles push-related operations.
type PushService struct {
	repo updates.Repository
}

// NewPushService creates a new push service.
func NewPushService(repo updates.Repository) *PushService {
	return &PushService{repo: repo}
}

// PushUpdate pushes an update to devices.
func (s *PushService) PushUpdate(ctx context.Context, req *PushUpdateRequest, initiatedBy string) (*PushUpdateResponse, error) {
	if len(req.DeviceIDs) == 0 {
		return nil, ErrBadRequest
	}

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
		Status:      updates.UpdateStatusPending,
		InitiatedBy: initiatedBy,
		InitiatedAt: now.UnixMilli(),
	}

	if err := s.repo.CreatePush(ctx, push); err != nil {
		return nil, fmt.Errorf("failed to create push: %w", err)
	}

	deviceIDs := make([]string, 0, len(req.DeviceIDs))
	failedDevices := make([]FailedDevice, 0)

	for _, deviceID := range req.DeviceIDs {
		devicePush := &updates.UpdatePushDevice{
			PushID:     push.ID,
			DeviceID:   deviceID,
			Status:     updates.DevicePushStatusPending,
			RetryCount: 0,
		}
		if err := s.repo.CreatePushDevice(ctx, devicePush); err != nil {
			// Record the failure instead of silently ignoring
			failedDevices = append(failedDevices, FailedDevice{
				DeviceID: deviceID,
				Reason:   "failed to register device for update",
			})
			continue
		}
		deviceIDs = append(deviceIDs, deviceID)
	}

	// If ALL devices failed, return an error
	if len(deviceIDs) == 0 && len(failedDevices) > 0 {
		return nil, &ServiceError{
			Code:    "all_devices_failed",
			Message: fmt.Sprintf("all %d devices failed to be registered for update", len(failedDevices)),
			Status:  http.StatusBadRequest,
			Details: failedDevices,
		}
	}

	return &PushUpdateResponse{
		InstallType:  req.InstallType,
		Status:       string(push.Status),
		InitiatedBy:  initiatedBy,
		PushID:       push.ID,
		Version:      req.Version,
		InitiatedAt:  push.InitiatedAt,
		DeviceIDs:    deviceIDs,
		FailedDevices: failedDevices,
		Devices: PushDeviceCounts{
			Total: len(deviceIDs),
		},
	}, nil
}
