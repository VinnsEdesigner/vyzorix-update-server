package updates

import (
	"context"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/github"
)

// SyncService handles GitHub sync operations.
type SyncService struct {
	repo     updates.Repository
	githubSvc *github.SyncService
}

// NewSyncService creates a new sync service.
func NewSyncService(repo updates.Repository, githubSvc *github.SyncService) *SyncService {
	return &SyncService{
		repo:     repo,
		githubSvc: githubSvc,
	}
}

// SyncFromGitHub triggers a manual sync from GitHub.
func (s *SyncService) SyncFromGitHub(ctx context.Context) (*SyncResponse, error) {
	// Atomically try to acquire sync lock to prevent race conditions.
	acquired, currentState, err := s.repo.TryAcquireSyncLock(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire sync lock: %w", err)
	}

	if !acquired {
		return &SyncResponse{
			Status:    "syncing",
			StartedAt: time.Now().UnixMilli(),
			Message:   "Sync already in progress",
		}, ErrSyncAlreadyInProgress
	}

	// Perform sync.
	result, err := s.githubSvc.SyncFromGitHub(ctx)
	if err != nil {
		errorState := &updates.SyncState{
			Status: updates.SyncStatusError,
			Error:  err.Error(),
		}
		if currentState != nil && currentState.LastSyncAt != nil {
			errorState.LastSyncAt = currentState.LastSyncAt
		}
		_ = s.repo.UpdateSyncState(ctx, errorState)
		return nil, fmt.Errorf("sync failed: %w", err)
	}

	// Update status to synced.
	now := time.Now()
	nextSync := now.Add(24 * time.Hour)
	syncedState := &updates.SyncState{
		Status:        updates.SyncStatusSynced,
		LastSyncAt:    PtrToInt64(now.UnixMilli()),
		NextSyncAt:    PtrToInt64(nextSync.UnixMilli()),
		VersionsFound: result.VersionsFound,
	}
	if err := s.repo.UpdateSyncState(ctx, syncedState); err != nil {
		return nil, fmt.Errorf("failed to update sync state: %w", err)
	}

	return &SyncResponse{
		Status:        result.Status,
		StartedAt:     result.StartedAt,
		Message:       result.Message,
		VersionsFound: result.VersionsFound,
	}, nil
}

// GetSyncStatus returns the current sync status.
func (s *SyncService) GetSyncStatus(ctx context.Context) (*GetSyncStatusResponse, error) {
	syncState, err := s.repo.GetSyncState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync state: %w", err)
	}

	response := &GetSyncStatusResponse{
		Status: "idle",
	}

	if syncState != nil {
		response.Status = string(syncState.Status)
		if syncState.LastSyncAt != nil {
			response.LastSyncAt = *syncState.LastSyncAt
		}
		if syncState.NextSyncAt != nil {
			response.NextSyncAt = *syncState.NextSyncAt
		}
		if syncState.VersionsFound > 0 {
			response.VersionsFound = syncState.VersionsFound
		}
		if syncState.Error != "" {
			response.Error = syncState.Error
		}
	}

	return response, nil
}
