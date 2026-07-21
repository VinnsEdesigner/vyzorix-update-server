// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"encoding/json"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
)

// ============================================================
// Helper Methods
// ============================================================

// deviceDTOToMap converts a DeviceResponse DTO to a GraphQL map.
func (r *Resolver) deviceDTOToMap(dev *dto.DeviceResponse) map[string]interface{} {
	return map[string]interface{}{
		"id":       dev.ID,
		"name":     dev.DeviceName,
		"online":   r.Hub != nil && r.Hub.Online(dev.ID),
		"lastSeen": time.UnixMilli(dev.LastSeen).Format(time.RFC3339),
		"version":  dev.AppVersion,
	}
}

// deviceDetailToMap converts a DeviceDetailResponse DTO to a GraphQL map.
func (r *Resolver) deviceDetailToMap(dev *dto.DeviceDetailResponse) map[string]interface{} {
	return map[string]interface{}{
		"id":       dev.ID,
		"name":     dev.DeviceName,
		"online":   r.Hub != nil && r.Hub.Online(dev.ID),
		"lastSeen": time.UnixMilli(dev.LastSeen).Format(time.RFC3339),
		"version":  dev.AppVersion,
	}
}

// commandStatusToMap converts a CommandStatusResponse to a GraphQL map.
func (r *Resolver) commandStatusToMap(cmd *dto.CommandStatusResponse) map[string]interface{} {
	var deliveredAt interface{}
	if cmd.DeliveredAt != nil {
		deliveredAt = cmd.DeliveredAt.Format(time.RFC3339)
	}

	return map[string]interface{}{
		"dispatchId":  cmd.DispatchID,
		"commandId":   cmd.CommandID,
		"deviceId":    cmd.DeviceID,
		"command":     cmd.Command,
		"status":      cmd.Status,
		"deliveredAt": deliveredAt,
	}
}

// commandToMap converts a CommandResponse to a GraphQL map.
func (r *Resolver) commandToMap(cmd dto.CommandResponse) map[string]interface{} {
	var args map[string]interface{}
	if len(cmd.Args) > 0 {
		_ = json.Unmarshal(cmd.Args, &args)
	}

	return map[string]interface{}{
		"dispatchId": cmd.DispatchID,
		"commandId":  cmd.ID,
		"deviceId":   cmd.DeviceID,
		"command":    cmd.Command,
		"args":       args,
		"status":     cmd.Status,
		"createdAt":  cmd.CreatedAt.Format(time.RFC3339),
	}
}

// paginationToMap converts updates.Pagination to a GraphQL map.
func (r *Resolver) paginationToMap(p updates.Pagination) map[string]interface{} {
	return map[string]interface{}{
		"total":    p.Total,
		"limit":    p.Limit,
		"offset":   (p.Page - 1) * p.Limit,
		"hasMore":  p.Page < p.TotalPages,
	}
}
