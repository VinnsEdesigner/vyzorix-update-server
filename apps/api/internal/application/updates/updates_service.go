package updates

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// Service is the main updates service that coordinates all update operations.
type Service struct {
	repo              updates.Repository
	versionsStatusSvc *VersionsStatusService
	versionsListSvc   *VersionsListService
	changelogSvc     *ChangelogService
	exportSvc        *ExportService
	pushSvc          *PushService
	historySvc       *HistoryService
	syncSvc          *SyncService
}

// NewService creates a new updates service.
func NewService(
	repo updates.Repository,
	versionsStatusSvc *VersionsStatusService,
	versionsListSvc *VersionsListService,
	changelogSvc *ChangelogService,
	exportSvc *ExportService,
	pushSvc *PushService,
	historySvc *HistoryService,
	syncSvc *SyncService,
) *Service {
	return &Service{
		repo:              repo,
		versionsStatusSvc: versionsStatusSvc,
		versionsListSvc:   versionsListSvc,
		changelogSvc:     changelogSvc,
		exportSvc:        exportSvc,
		pushSvc:          pushSvc,
		historySvc:       historySvc,
		syncSvc:          syncSvc,
	}
}

// GetRepo returns the updates repository.
func (s *Service) GetRepo() updates.Repository {
	return s.repo
}


// GetPushService returns the push service for handler initialization.
func (s *Service) GetPushService() *PushService {
        return s.pushSvc
}
// GetStatus returns the current update system status.
func (s *Service) GetStatus(ctx context.Context) (*GetStatusResponse, error) {
	return s.versionsStatusSvc.GetStatus(ctx)
}

// GetUpdateStatus is an alias for GetStatus to match expected method names.
func (s *Service) GetUpdateStatus(ctx context.Context) (*GetStatusResponse, error) {
	return s.GetStatus(ctx)
}

// GetVersions returns paginated versions.
func (s *Service) GetVersions(ctx context.Context, status string, page, limit int) (*ListVersionsResponse, error) {
	return s.versionsListSvc.GetVersions(ctx, status, page, limit)
}

// GetChangelog returns the changelog.
func (s *Service) GetChangelog(ctx context.Context, version string) (*GetChangelogResponse, error) {
	return s.changelogSvc.GetChangelog(ctx, version)
}

// PushUpdate pushes an update to devices.
func (s *Service) PushUpdate(ctx context.Context, req *PushUpdateRequest, initiatedBy string) (*PushUpdateResponse, error) {
	return s.pushSvc.PushUpdate(ctx, req, initiatedBy)
}

// GetHistory returns paginated push history.
func (s *Service) GetHistory(ctx context.Context, status string, page, limit int, orgID string) (*ListHistoryResponse, error) {
	return s.historySvc.GetHistory(ctx, status, page, limit, orgID)
}

// GetPushDetail returns detailed information about a specific push.
func (s *Service) GetPushDetail(ctx context.Context, pushID string, orgID string) (*PushDetailResponse, error) {
	return s.historySvc.GetPushDetail(ctx, pushID, orgID)
}

// CancelPush cancels a pending push.
func (s *Service) CancelPush(ctx context.Context, pushID, cancelledBy string, orgID string) (*CancelPushResponse, error) {
	return s.historySvc.CancelPush(ctx, pushID, cancelledBy, orgID)
}

// SyncFromGitHub triggers a manual sync from GitHub.
func (s *Service) SyncFromGitHub(ctx context.Context) (*SyncResponse, error) {
	return s.syncSvc.SyncFromGitHub(ctx)
}

// GetSyncStatus returns the current sync status.
func (s *Service) GetSyncStatus(ctx context.Context) (*GetSyncStatusResponse, error) {
	return s.syncSvc.GetSyncStatus(ctx)
}

// ExportVersions exports version data.
func (s *Service) ExportVersions(ctx context.Context, format, version string, includeChangelog, includeApkInfo bool) (*ExportResponse, error) {
	return s.exportSvc.ExportVersions(ctx, format, version, includeChangelog, includeApkInfo)
}
