package updates

import (
	"context"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// HistoryService handles history-related operations.
type HistoryService struct {
	repo updates.Repository
}

// NewHistoryService creates a new history service.
func NewHistoryService(repo updates.Repository) *HistoryService {
	return &HistoryService{repo: repo}
}

// GetHistory returns paginated push history.
func (s *HistoryService) GetHistory(ctx context.Context, status string, page, limit int) (*ListHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	pushes, total, err := s.repo.ListPushes(ctx, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list pushes: %w", err)
	}

	entries := make([]PushHistoryEntry, 0, len(pushes))
	for _, p := range pushes {
		entry, err := s.pushToHistoryEntry(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("failed to get push history entry for %s: %w", p.ID, err)
		}
		entries = append(entries, entry)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return &ListHistoryResponse{
		Pushes: entries,
		Pagination: PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// getPushVersionString retrieves the version string for a push.
func (s *HistoryService) getPushVersionString(ctx context.Context, push *updates.UpdatePush) string {
	version, err := s.repo.GetVersionByVersion(ctx, push.VersionID)
	if err != nil {
		return push.VersionID
	}
	return version.Version
}

// GetPushDetail returns detailed information about a specific push.
func (s *HistoryService) GetPushDetail(ctx context.Context, pushID string) (*PushDetailResponse, error) {
	push, version, err := s.repo.GetPushByIDWithVersion(ctx, pushID)
	if err != nil {
		if err == updates.ErrPushNotFound {
			return nil, ErrPushNotFound
		}
		return nil, fmt.Errorf("failed to get push: %w", err)
	}

	devices, err := s.repo.GetPushDevices(ctx, pushID)
	if err != nil {
		return nil, fmt.Errorf("failed to get push devices: %w", err)
	}

	detailDevices := make([]PushDetailDevice, 0, len(devices))
	for _, d := range devices {
		detailDevices = append(detailDevices, PushDetailDevice{
			DeviceID:       d.DeviceID,
			Status:         string(d.Status),
			AcknowledgedAt: d.AcknowledgedAt,
			Error:          d.Error,
		})
	}

	versionStr := push.VersionID
	if version != nil {
		versionStr = version.Version
	}

	return &PushDetailResponse{
		ID:          push.ID,
		Version:     versionStr,
		InstallType: string(push.InstallType),
		Status:      string(push.Status),
		InitiatedBy: push.InitiatedBy,
		InitiatedAt: push.InitiatedAt,
		ScheduledAt: push.ScheduledAt,
		CompletedAt: push.CompletedAt,
		Devices:     detailDevices,
	}, nil
}

// CancelPush cancels a pending push.
func (s *HistoryService) CancelPush(ctx context.Context, pushID, cancelledBy string) (*CancelPushResponse, error) {
	push, err := s.repo.GetPushByID(ctx, pushID)
	if err != nil {
		if err == updates.ErrPushNotFound {
			return nil, ErrPushNotFound
		}
		return nil, fmt.Errorf("failed to get push: %w", err)
	}

	if !push.CanCancel() {
		return nil, ErrPushNotCancellable
	}

	now := time.Now()
	if err := s.repo.CancelPush(ctx, pushID, cancelledBy); err != nil {
		return nil, fmt.Errorf("failed to cancel push: %w", err)
	}

	return &CancelPushResponse{
		ID:          pushID,
		Status:      string(updates.UpdateStatusCancelled),
		CancelledAt: now.UnixMilli(),
		CancelledBy: cancelledBy,
	}, nil
}

// pushToHistoryEntry converts an UpdatePush to a PushHistoryEntry.
func (s *HistoryService) pushToHistoryEntry(ctx context.Context, p *updates.UpdatePush) (PushHistoryEntry, error) {
	// Get the version string for this push
	versionStr := s.getPushVersionString(ctx, p)

	entry := PushHistoryEntry{
		ID:          p.ID,
		Version:     versionStr,
		InstallType: string(p.InstallType),
		Status:      string(p.Status),
		InitiatedBy: p.InitiatedBy,
		InitiatedAt: p.InitiatedAt,
		CompletedAt: p.CompletedAt,
		CancelledAt: p.CancelledAt,
		ScheduledAt: p.ScheduledAt,
	}

	// Get device counts for the push - fail if any count fails
	pending, err := s.repo.CountPushDevicesByStatus(ctx, p.ID, updates.DevicePushStatusPending)
	if err != nil {
		return entry, fmt.Errorf("failed to count pending devices: %w", err)
	}
	sent, err := s.repo.CountPushDevicesByStatus(ctx, p.ID, updates.DevicePushStatusSent)
	if err != nil {
		return entry, fmt.Errorf("failed to count sent devices: %w", err)
	}
	acknowledged, err := s.repo.CountPushDevicesByStatus(ctx, p.ID, updates.DevicePushStatusAcknowledged)
	if err != nil {
		return entry, fmt.Errorf("failed to count acknowledged devices: %w", err)
	}
	failed, err := s.repo.CountPushDevicesByStatus(ctx, p.ID, updates.DevicePushStatusFailed)
	if err != nil {
		return entry, fmt.Errorf("failed to count failed devices: %w", err)
	}

	entry.DeviceCount = pending + sent + acknowledged + failed
	entry.Devices = HistoryDeviceCounts{
		Pending:      pending,
		Sent:         sent,
		Acknowledged: acknowledged,
		Failed:       failed,
	}

	return entry, nil
}
