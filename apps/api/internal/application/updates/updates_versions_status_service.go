package updates

import (
	"context"
	"fmt"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// VersionsStatusService handles version status operations.
type VersionsStatusService struct {
	repo updates.Repository
}

// NewVersionsStatusService creates a new VersionsStatusService.
func NewVersionsStatusService(repo updates.Repository) *VersionsStatusService {
	return &VersionsStatusService{repo: repo}
}

// GetStatus returns the current update system status.
func (s *VersionsStatusService) GetStatus(ctx context.Context) (*GetStatusResponse, error) {
	syncState, err := s.repo.GetSyncState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync state: %w", err)
	}

	var syncStatus SyncStatusInfo
	if syncState != nil {
		syncStatus = SyncStatusInfo{
			Status: string(syncState.Status),
		}
		if syncState.LastSyncAt != nil {
			syncStatus.LastSyncAt = *syncState.LastSyncAt
		}
		if syncState.NextSyncAt != nil {
			syncStatus.NextSyncAt = *syncState.NextSyncAt
		}
		if syncState.Error != "" {
			syncStatus.Error = syncState.Error
		}
	}

	latest, err := s.repo.GetLatestVersion(ctx)
	if err != nil && err != updates.ErrVersionNotFound {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	var latestInfo LatestVersionInfo
	if latest != nil {
		latestInfo = LatestVersionInfo{
			Version:     latest.Version,
			APKFilename: latest.APKFilename,
			APKSize:     latest.APKSize,
			SHA256:      latest.SHA256,
			ReleasedAt:  latest.ReleaseDate,
		}
	}

	// NOTE: Device status requires a deviceId which is not part of this endpoint.
	// Device update status should be retrieved via device-specific endpoints or
	// by passing a deviceId query parameter to this endpoint in the future.
	return &GetStatusResponse{
		Sync:   syncStatus,
		Latest: latestInfo,
	}, nil
}
