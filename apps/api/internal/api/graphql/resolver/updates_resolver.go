// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"strings"
	"time"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/graphql-go/graphql"
)

// ============================================================.
// Updates Query Resolvers.
// ============================================================.

// GetUpdatesStatus resolves the updatesStatus query.
func (r *Resolver) GetUpdatesStatus(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)
	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.MemberService != nil {
		if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID); err != nil {
			return nil, r.Presenter.ForbiddenError("not authorized to access this organization")
		}
	}

	if r.UpdatesSvc == nil {
		return nil, r.Presenter.InternalError("updates service not available")
	}

	status, err := r.UpdatesSvc.GetStatus(ctx)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get update status")
	}

	var currentVersion string
	var needsUpdate bool

	// If deviceId is provided, get device-specific status.
	if deviceID != "" && r.DeviceService != nil {
		dev, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
		if err == nil && dev != nil {
			currentVersion = dev.AppVersion
			// Compare versions to determine if update is needed.
			if status.Latest.Version != "" && currentVersion != "" {
				needsUpdate = currentVersion != status.Latest.Version
			}
		}
	} else if status.Device != nil {
		currentVersion = status.Device.CurrentVersion
		needsUpdate = status.Device.NeedsUpdate
	}

	result := map[string]interface{}{
		"sync": map[string]interface{}{
			"status":        status.Sync.Status,
			"lastSyncAt":    formatDateTimeInt64(status.Sync.LastSyncAt),
			"nextSyncAt":    formatDateTimeInt64(status.Sync.NextSyncAt),
			"versionsFound": nil,
			"error":         status.Sync.Error,
		},
		"device": map[string]interface{}{
			"currentVersion": currentVersion,
			"needsUpdate":   needsUpdate,
		},
		"version":     currentVersion,
		"apkFilename": nil,
		"sha256":      nil,
	}

	if status.Latest.Version != "" {
		result["latest"] = map[string]interface{}{
			"id":           "",
			"version":      status.Latest.Version,
			"releaseType":  strings.ToUpper(string(status.Latest.ReleaseType)),
			"releaseNotes": nil,
			"apkFilename":  status.Latest.APKFilename,
			"apkSize":      status.Latest.APKSize,
			"sha256":       status.Latest.SHA256,
			"releasedAt":   formatDateTimeInt64(status.Latest.ReleasedAt),
			"createdAt":    nil,
		}
	}

	return result, nil
}

// GetUpdatesVersions resolves the updatesVersions query.
func (r *Resolver) GetUpdatesVersions(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.UpdatesSvc == nil {
		return nil, r.Presenter.InternalError("updates service not available")
	}

	filterStatus, _ := p.Args["status"].(string)
	limit, _ := p.Args["limit"].(int)
	offset, _ := p.Args["offset"].(int)

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	resp, err := r.UpdatesSvc.GetVersions(ctx, filterStatus, limit, offset)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get versions")
	}

	result := make([]map[string]interface{}, 0, len(resp.Versions))
	for _, v := range resp.Versions {
		entry := map[string]interface{}{
			"id":           v.Version,
			"version":      v.Version,
			"releaseType":  strings.ToUpper(string(v.ReleaseType)),
			"releaseNotes": v.ReleaseNotes,
			"apkFilename":  v.APKFilename,
			"apkSize":      v.APKSize,
			"sha256":       v.SHA256,
			"releasedAt":   formatDateTimeInt64(v.ReleasedAt),
			"createdAt":    nil,
			"isLatest":     v.IsLatest,
		}
		result = append(result, entry)
	}

	return map[string]interface{}{
		"versions":   result,
		"pagination": r.paginationToMap(resp.Pagination),
	}, nil
}

// GetUpdatesChangelog resolves the updatesChangelog query.
func (r *Resolver) GetUpdatesChangelog(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.UpdatesSvc == nil {
		return nil, r.Presenter.InternalError("updates service not available")
	}

	version, _ := p.Args["version"].(string)

	resp, err := r.UpdatesSvc.GetChangelog(ctx, version)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get changelog")
	}

	result := make([]map[string]interface{}, 0, len(resp.Changelog))
	for _, e := range resp.Changelog {
		result = append(result, map[string]interface{}{
			"version": e.Version,
			"date":    e.Date,
			"type":    e.Type,
			"notes":   e.Notes,
		})
	}

	return result, nil
}

// GetUpdatesHistory resolves the updatesHistory query.
func (r *Resolver) GetUpdatesHistory(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	if r.MemberService != nil {
		if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID); err != nil {
			return nil, r.Presenter.ForbiddenError("not authorized to access this organization")
		}
	}

	if r.UpdatesSvc == nil {
		return nil, r.Presenter.InternalError("updates service not available")
	}

	filterStatus, _ := p.Args["status"].(string)
	page, _ := p.Args["page"].(int)
	limit, _ := p.Args["limit"].(int)

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 20
	}

	resp, err := r.UpdatesSvc.GetHistory(ctx, filterStatus, page, limit, orgID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get push history")
	}

	pushes := make([]map[string]interface{}, 0, len(resp.Pushes))
	for _, push := range resp.Pushes {
		entry := map[string]interface{}{
			"id":           push.ID,
			"version":      push.Version,
			"installType":  push.InstallType,
			"scheduledAt":  formatDateTimeInt64Ptr(push.ScheduledAt),
			"status":       push.Status,
			"initiatedBy":  push.InitiatedBy,
			"initiatedAt":  push.InitiatedAt,
			"completedAt":  formatDateTimeInt64Ptr(push.CompletedAt),
			"cancelledAt":  formatDateTimeInt64Ptr(push.CancelledAt),
			"deviceCount":  push.DeviceCount,
			"pending":      push.Devices.Pending,
			"sent":         push.Devices.Sent,
			"acknowledged": push.Devices.Acknowledged,
			"failed":       push.Devices.Failed,
		}
		pushes = append(pushes, entry)
	}

	return map[string]interface{}{
		"pushes": pushes,
		"pagination": map[string]interface{}{
			"total":   resp.Pagination.Total,
			"limit":   resp.Pagination.Limit,
			"offset":  (resp.Pagination.Page - 1) * resp.Pagination.Limit,
			"hasMore": resp.Pagination.Page < resp.Pagination.TotalPages,
		},
	}, nil
}

// GetUpdatesHistoryDetail resolves the updatesHistoryDetail query.
func (r *Resolver) GetUpdatesHistoryDetail(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	id, ok := p.Args["id"].(string)
	if !ok || id == "" {
		return nil, r.Presenter.BadRequestError("push ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	if r.MemberService != nil {
		if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID); err != nil {
			return nil, r.Presenter.ForbiddenError("not authorized to access this organization")
		}
	}

	if r.UpdatesSvc == nil {
		return nil, r.Presenter.InternalError("updates service not available")
	}

	push, err := r.UpdatesSvc.GetPushDetail(ctx, id, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("push not found")
	}

	devices := make([]map[string]interface{}, 0, len(push.Devices))
	for _, d := range push.Devices {
		devices = append(devices, map[string]interface{}{
			"id":             d.ID,
			"deviceId":       d.DeviceID,
			"deviceName":     d.DeviceName,
			"status":         d.Status,
			"sentAt":         formatDateTimeInt64Ptr(d.SentAt),
			"acknowledgedAt": formatDateTimeInt64Ptr(d.AcknowledgedAt),
			"error":          d.Error,
		})
	}

	return map[string]interface{}{
		"id":           push.ID,
		"version":      push.Version,
		"installType":  push.InstallType,
		"scheduledAt":  formatDateTimeInt64Ptr(push.ScheduledAt),
		"status":       push.Status,
		"initiatedBy":  push.InitiatedBy,
		"initiatedAt":  formatDateTimeInt64(push.InitiatedAt),
		"completedAt":  formatDateTimeInt64Ptr(push.CompletedAt),
		"cancelledAt":  formatDateTimeInt64Ptr(push.CancelledAt),
		"deviceCount":  len(push.Devices),
		"devices":      devices,
	}, nil
}

// GetUpdatesSyncStatus resolves the updatesSyncStatus query.
func (r *Resolver) GetUpdatesSyncStatus(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.UpdatesSvc == nil {
		return nil, r.Presenter.InternalError("updates service not available")
	}

	status, err := r.UpdatesSvc.GetStatus(ctx)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get sync status")
	}

	return map[string]interface{}{
		"status":        status.Sync.Status,
		"lastSyncAt":    formatDateTimeInt64(status.Sync.LastSyncAt),
		"nextSyncAt":    formatDateTimeInt64(status.Sync.NextSyncAt),
		"versionsFound": nil,
		"error":         status.Sync.Error,
	}, nil
}

// ============================================================.
// Updates Mutation Resolvers.
// ============================================================.

// PushUpdate resolves the pushUpdate mutation.
func (r *Resolver) PushUpdate(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	version, _ := p.Args["version"].(string)
	deviceIdsRaw, _ := p.Args["deviceIds"].([]interface{})
	installType, _ := p.Args["installType"].(string)
	scheduledAt, _ := p.Args["scheduledAt"].(int64)

	if version == "" {
		return nil, r.Presenter.BadRequestError("version is required")
	}

	if len(deviceIdsRaw) == 0 {
		return nil, r.Presenter.BadRequestError("at least one device ID is required")
	}

	deviceIds := make([]string, 0, len(deviceIdsRaw))
	for _, id := range deviceIdsRaw {
		if idStr, isString := id.(string); isString {
			deviceIds = append(deviceIds, idStr)
		}
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.MemberService != nil {
		if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID); err != nil {
			return nil, r.Presenter.ForbiddenError("not authorized to access this organization")
		}
	}

	if r.UpdatesSvc == nil {
		return nil, r.Presenter.InternalError("updates service not available")
	}

	var schedPtr *int64
	if scheduledAt > 0 {
		schedPtr = &scheduledAt
	}

	pushReq := &updates.PushUpdateRequest{
		Version:        version,
		DeviceIDs:      deviceIds,
		InstallType:    installType,
		ScheduledAt:    schedPtr,
		OrganizationID: orgID,
	}

	resp, err := r.UpdatesSvc.PushUpdate(ctx, pushReq, op.ID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to push update: " + err.Error())
	}

	var schedAt interface{}
	if resp.ScheduledAt != nil && *resp.ScheduledAt > 0 {
		schedAt = *resp.ScheduledAt
	}

	return map[string]interface{}{
		"pushId":       resp.PushID,
		"version":      resp.Version,
		"installType":  resp.InstallType,
		"scheduledAt":  schedAt,
		"status":       resp.Status,
		"initiatedBy":  resp.InitiatedBy,
		"initiatedAt":  resp.InitiatedAt,
		"deviceCount":  len(resp.DeviceIDs),
	}, nil
}

// CancelUpdate resolves the cancelUpdate mutation.
func (r *Resolver) CancelUpdate(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	id, ok := p.Args["id"].(string)
	if !ok || id == "" {
		return nil, r.Presenter.BadRequestError("push ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	if r.MemberService != nil {
		if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID); err != nil {
			return nil, r.Presenter.ForbiddenError("not authorized to access this organization")
		}
	}

	if r.UpdatesSvc == nil {
		return nil, r.Presenter.InternalError("updates service not available")
	}

	resp, err := r.UpdatesSvc.CancelPush(ctx, id, op.ID, orgID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to cancel push")
	}

	return map[string]interface{}{
		"id":          resp.ID,
		"status":      resp.Status,
		"cancelledAt": resp.CancelledAt,
		"cancelledBy": resp.CancelledBy,
	}, nil
}

// SyncFromGitHub resolves the syncFromGitHub mutation.
func (r *Resolver) SyncFromGitHub(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.UpdatesSvc == nil {
		return nil, r.Presenter.InternalError("updates service not available")
	}

	resp, err := r.UpdatesSvc.SyncFromGitHub(ctx)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to sync from GitHub")
	}

	return map[string]interface{}{
		"status":        resp.Status,
		"startedAt":     resp.StartedAt,
		"message":       resp.Message,
		"versionsFound": resp.VersionsFound,
	}, nil
}

// ============================================================.
// Helper Functions.
// ============================================================.

func formatDateTimeInt64(ts int64) *string {
	if ts == 0 {
		return nil
	}
	t := time.UnixMilli(ts)
	s := t.Format(time.RFC3339)
	return &s
}

func formatDateTimeInt64Ptr(ts *int64) *string {
	if ts == nil || *ts == 0 {
		return nil
	}
	t := time.UnixMilli(*ts)
	s := t.Format(time.RFC3339)
	return &s
}
