// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"context"
	"encoding/json"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
)

// ============================================================
// Helper Methods
// ============================================================

// deviceToMap converts a domain Device entity to a GraphQL map.
func (r *Resolver) deviceToMap(_ context.Context, dev *device.Device) map[string]interface{} {
	return map[string]interface{}{
		"id":       dev.ID,
		"name":     "",
		"online":   r.Hub != nil && r.Hub.Online(dev.ID),
		"lastSeen": time.UnixMilli(dev.LastSeen).Format(time.RFC3339),
		"version":  dev.AppVersion,
	}
}

// deviceDTOToMap converts a DeviceResponse DTO to a GraphQL map.
func (r *Resolver) deviceDTOToMap(dev *dto.DeviceResponse) map[string]interface{} {
	return map[string]interface{}{
		"id":       dev.ID,
		"name":     "",
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
