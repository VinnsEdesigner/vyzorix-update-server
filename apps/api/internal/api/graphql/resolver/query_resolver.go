// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"time"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	appmetrics "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	diagnosticsapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	orgdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/graphql-go/graphql"
)

// ============================================================
// Settings Query Resolvers
// ============================================================

// GetMySettings resolves the mySettings query.
func (r *Resolver) GetMySettings(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	settings, err := r.OperatorRepo.GetOperatorSettings(ctx, op.ID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get settings")
	}

	return map[string]interface{}{
		"client": map[string]interface{}{
			"serverUrl":           settings.Client.ServerURL,
			"deviceId":            settings.Client.DeviceID,
			"requestTimeoutMs":    settings.Client.RequestTimeoutMs,
			"autoReconnect":      settings.Client.AutoReconnect,
			"strictHmac":         settings.Client.StrictHmac,
			"logBufferLimit":     settings.Client.LogBufferLimit,
			"signalHistoryLimit":  settings.Client.SignalHistoryLimit,
		},
		"notifications": settings.Notifications,
	}, nil
}

// GetMyNotifications resolves the myNotifications query.
func (r *Resolver) GetMyNotifications(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	notifications, err := r.NotificationSvc.GetNotifications(ctx, op.ID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get notifications")
	}

	return notifications, nil
}

// GetDeviceSettings resolves the deviceSettings query.
func (r *Resolver) GetDeviceSettings(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("organization ID is required")
	}

	deviceImei, ok := p.Args["deviceImei"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("device IMEI is required")
	}

	// Get device settings with effective thresholds (device → org → default)
	settings, err := r.DeviceSettingsService.GetSettings(ctx, deviceImei)
	if err != nil {
		if err == devicedomain.ErrSettingsNotFound {
			return nil, r.Presenter.NotFoundError("device settings not found")
		}
		return nil, r.Presenter.InternalError("failed to get device settings")
	}

	// Get organization settings for threshold resolution
	orgSettings, err := r.OrgSettingsService.GetSettings(ctx, orgID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get organization settings")
	}

	// Convert org thresholds to device thresholds for resolution
	orgThresholds := devicedomain.FromOrgThresholds(orgSettings.DefaultThresholds)
	effectiveThresholds := devicedomain.ResolveThresholds(settings, orgThresholds)

	return map[string]interface{}{
		"id":                   settings.ID,
		"deviceImei":           settings.DeviceIMEI,
		"customName":           settings.CustomName,
		"location":             settings.Location,
		"metadata":             convertMetadataToList(settings.Metadata),
		"thresholds":            settings.Thresholds,
		"effectiveThresholds":  effectiveThresholds,
		"createdAt":            settings.CreatedAt,
		"updatedAt":            settings.UpdatedAt,
	}, nil
}

// convertMetadataToList converts a metadata map to a list of {key, value} entries.
func convertMetadataToList(metadata map[string]string) []map[string]string {
	if metadata == nil {
		return nil
	}
	result := make([]map[string]string, 0, len(metadata))
	for k, v := range metadata {
		result = append(result, map[string]string{"key": k, "value": v})
	}
	return result
}

// GetOrganizationSettings resolves the organizationSettings query.
func (r *Resolver) GetOrganizationSettings(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("organization ID is required")
	}

	settings, err := r.OrgSettingsService.GetSettings(ctx, orgID)
	if err != nil {
		if err == orgdomain.ErrSettingsNotFound {
			return nil, r.Presenter.NotFoundError("organization settings not found")
		}
		return nil, r.Presenter.InternalError("failed to get organization settings")
	}

	return map[string]interface{}{
		"id":                    settings.ID,
		"organizationId":        settings.OrganizationID,
		"timezone":              settings.Timezone,
		"dateFormat":            settings.DateFormat,
		"alertCooldownMinutes":  settings.AlertCooldownMinutes,
		"defaultThresholds":     settings.DefaultThresholds,
		"createdAt":             settings.CreatedAt,
		"updatedAt":             settings.UpdatedAt,
	}, nil
}

// ============================================================
// Query Resolvers
// ============================================================

// GetDevice resolves the device query.
func (r *Resolver) GetDevice(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	id, ok := p.Args["id"].(string)
	if !ok || id == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Use organization-scoped method for multi-tenant isolation
	dev, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, id, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	// Log via presenter
	r.Presenter.DeviceView(ctx, op.ID, id)

	return r.deviceDetailToMap(dev), nil
}

// GetDevices resolves the devices list query.
func (r *Resolver) GetDevices(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	limit, _ := p.Args["limit"].(int)
	offset, _ := p.Args["offset"].(int)

	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Use organization-scoped method with pagination
	page := (offset / limit) + 1
	result, err := r.DeviceService.GetDevices(ctx, &device.ListQuery{
		OrganizationID: orgID,
		Page:          page,
		Limit:         limit,
	})
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list devices")
	}

	// Log via presenter
	r.Presenter.DeviceList(ctx, op.ID)

	devices := make([]map[string]interface{}, 0, len(result.Devices))
	for _, dev := range result.Devices {
		devices = append(devices, r.deviceDTOToMap(&dev))
	}

	return devices, nil
}

// GetDeviceCount resolves the deviceCount query.
func (r *Resolver) GetDeviceCount(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Use organization-scoped count method
	count, err := r.DeviceService.CountByOrganization(ctx, orgID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to count devices")
	}

	// Log via presenter
	r.Presenter.DeviceCount(ctx, op.ID)

	return count, nil
}

// GetCommand resolves the command query.
func (r *Resolver) GetCommand(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	dispatchID, ok := p.Args["dispatchId"].(string)
	if !ok || dispatchID == "" {
		return nil, r.Presenter.BadRequestError("dispatch ID is required")
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	cmd, err := r.CommandService.GetCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	// Verify device exists in organization
	_, err = r.DeviceService.GetDeviceDetailByOrganization(ctx, cmd.DeviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	// Log via presenter
	r.Presenter.CommandView(ctx, op.ID, cmd.CommandID)

	return r.commandStatusToMap(cmd), nil
}

// GetPendingCommands resolves the pendingCommands query.
func (r *Resolver) GetPendingCommands(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	deviceID, ok := p.Args["deviceId"].(string)
	if !ok || deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device exists in organization
	_, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	cmds, err := r.CommandService.GetPendingCommands(ctx, deviceID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get pending commands")
	}

	result := make([]map[string]interface{}, 0, len(cmds))
	for _, cmd := range cmds {
		result = append(result, r.commandToMap(cmd))
	}

	return result, nil
}

// GetTelemetryHistory resolves the telemetryHistory query.
func (r *Resolver) GetTelemetryHistory(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)
	orgID, _ := p.Args["organizationId"].(string)
	startTime, _ := p.Args["startTime"].(int64)
	endTime, _ := p.Args["endTime"].(int64)
	limit, _ := p.Args["limit"].(int)

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	if orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device exists in organization
	_, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	// Default time range to last hour
	now := time.Now()
	if endTime <= 0 {
		endTime = now.UnixMilli()
	}

	if startTime <= 0 {
		startTime = now.Add(-1 * time.Hour).UnixMilli()
	}

	entries, err := r.TelemetryRepo.ListSince(ctx, deviceID, startTime, limit)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to query telemetry")
	}

	// Log via presenter
	r.Presenter.TelemetryQuery(ctx, op.ID, deviceID)

	result := make([]map[string]interface{}, 0, len(entries))

	for _, entry := range entries {
		ts := entry.ReceivedAt.UnixMilli()
		if ts >= startTime && ts <= endTime {
			result = append(result, map[string]interface{}{
				"id":          entry.ID,
				"deviceId":    entry.DeviceID,
				"receivedAt":  entry.ReceivedAt.Format(time.RFC3339),
				"riskScore":   entry.RiskScore,
				"bufferLevel": entry.BufferLevel,
				"thermalTemp": entry.ThermalTemp,
				"payload":     string(entry.Payload),
			})
		}
	}

	return result, nil
}

// GetLatestTelemetry resolves the latestTelemetry query.
func (r *Resolver) GetLatestTelemetry(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)
	orgID, _ := p.Args["organizationId"].(string)

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	if orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device exists in organization
	_, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	entries, err := r.TelemetryRepo.List(ctx, deviceID, 1)
	if err != nil || len(entries) == 0 {
		return nil, r.Presenter.NotFoundError("no telemetry found")
	}

	// Log via presenter
	r.Presenter.TelemetryQuery(ctx, op.ID, deviceID)

	entry := entries[0]

	return map[string]interface{}{
		"id":          entry.ID,
		"deviceId":    entry.DeviceID,
		"receivedAt":  entry.ReceivedAt.Format(time.RFC3339),
		"riskScore":   entry.RiskScore,
		"bufferLevel": entry.BufferLevel,
		"thermalTemp": entry.ThermalTemp,
		"payload":     string(entry.Payload),
	}, nil
}

// GetTelemetryStats resolves the telemetryStats query.
func (r *Resolver) GetTelemetryStats(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)
	orgID, _ := p.Args["organizationId"].(string)

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	if orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device exists in organization
	_, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	entries, err := r.TelemetryRepo.List(ctx, deviceID, 100)
	if err != nil || len(entries) == 0 {
		return nil, r.Presenter.NotFoundError("no telemetry found")
	}

	// Calculate stats
	var totalRisk, totalBuffer, totalTemp int

	var minRisk, maxRisk = 999, -1

	var minTemp, maxTemp float64 = 999, -999

	for _, e := range entries {
		totalRisk += e.RiskScore
		totalBuffer += e.BufferLevel
		totalTemp += int(e.ThermalTemp * 100)

		if e.RiskScore < minRisk {
			minRisk = e.RiskScore
		}

		if e.RiskScore > maxRisk {
			maxRisk = e.RiskScore
		}

		if e.ThermalTemp < minTemp {
			minTemp = e.ThermalTemp
		}

		if e.ThermalTemp > maxTemp {
			maxTemp = e.ThermalTemp
		}
	}

	count := len(entries)

	return map[string]interface{}{
		"deviceId":    deviceID,
		"sampleCount": count,
		"riskScore": map[string]interface{}{
			"avg": float64(totalRisk) / float64(count),
			"min": minRisk,
			"max": maxRisk,
		},
		"bufferLevel": map[string]interface{}{
			"avg": float64(totalBuffer) / float64(count),
		},
		"thermalTemp": map[string]interface{}{
			"avg": float64(totalTemp) / float64(count) / 100,
			"min": minTemp,
			"max": maxTemp,
		},
	}, nil
}

// GetConnectionStatus resolves the connectionStatus query.
func (r *Resolver) GetConnectionStatus(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)
	orgID, _ := p.Args["organizationId"].(string)

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	if orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device exists in organization
	_, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	if r.Hub == nil {
		return nil, r.Presenter.InternalError("WebSocket hub not available")
	}

	client := r.Hub.GetClient(deviceID)
	if client == nil {
		return map[string]interface{}{
			"deviceId":      deviceID,
			"connected":     false,
			"connectedAt":   nil,
			"lastMessageAt": nil,
			"uptimeSeconds": 0,
		}, nil
	}

	metrics := client.GetMetrics()

	var lastMsgAt interface{}
	if metrics.LastMessageAt > 0 {
		lastMsgAt = time.Unix(metrics.LastMessageAt, 0).Format(time.RFC3339)
	}

	return map[string]interface{}{
		"deviceId":      deviceID,
		"connected":     client.IsConnected(),
		"connectedAt":   time.Unix(metrics.LastConnectedAt, 0).Format(time.RFC3339),
		"lastMessageAt": lastMsgAt,
		"uptimeSeconds": client.Uptime(),
	}, nil
}

// GetAllConnections resolves the allConnections query.
func (r *Resolver) GetAllConnections(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.Hub == nil {
		return nil, r.Presenter.InternalError("WebSocket hub not available")
	}

	// Get all devices for this organization
	devices, err := r.DeviceService.GetDevices(ctx, &device.ListQuery{
		OrganizationID: orgID,
		Page:          1,
		Limit:         1000,
	})
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get organization devices")
	}

	// Create a map of device IDs for the organization
	orgDevices := make(map[string]bool)
	for _, d := range devices.Devices {
		orgDevices[d.ID] = true
	}

	clients := r.Hub.Clients()
	result := make([]map[string]interface{}, 0, len(clients))

	for deviceID := range clients {
		// Only include devices in this organization
		if !orgDevices[deviceID] {
			continue
		}

		client := clients[deviceID]
		metrics := client.GetMetrics()

		var lastMsgAt interface{}
		if metrics.LastMessageAt > 0 {
			lastMsgAt = time.Unix(metrics.LastMessageAt, 0).Format(time.RFC3339)
		}

		result = append(result, map[string]interface{}{
			"deviceId":      deviceID,
			"connected":     client.IsConnected(),
			"connectedAt":   time.Unix(metrics.LastConnectedAt, 0).Format(time.RFC3339),
			"lastMessageAt": lastMsgAt,
			"uptimeSeconds": client.Uptime(),
		})
	}

	return result, nil
}

// ============================================================
// Dashboard Commands & Logs Query Resolvers
// ============================================================

// GetDeviceMetrics resolves the deviceMetrics query.
func (r *Resolver) GetDeviceMetrics(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	imei, ok := p.Args["imei"].(string)
	if !ok || imei == "" {
		return nil, r.Presenter.BadRequestError("device IMEI is required")
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device exists in organization
	device, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, imei, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	// Get metrics from service
	if r.MetricsSvc == nil {
		return nil, r.Presenter.InternalError("metrics service not available")
	}

	rangeVal, _ := p.Args["range"].(string)
	if rangeVal == "" {
		rangeVal = "6h"
	}

	req := &appmetrics.GetMetricsRequest{
		DeviceID:   imei,
		Range:      rangeVal,
	}

	if startTime, ok := p.Args["startTime"].(int64); ok && startTime > 0 {
		req.StartTime = startTime
	}

	if endTime, ok := p.Args["endTime"].(int64); ok && endTime > 0 {
		req.EndTime = endTime
	}

	if resolution, ok := p.Args["resolution"].(string); ok {
		req.Resolution = resolution
	}

	resp, err := r.MetricsSvc.GetDeviceMetrics(ctx, req)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get device metrics")
	}

	// Convert to GraphQL response format
	return map[string]interface{}{
		"device": map[string]interface{}{
			"imei":       resp.Device.IMEI,
			"deviceName": device.DeviceName,
		},
		"timeRange": map[string]interface{}{
			"start":      resp.TimeRange.Start,
			"end":        resp.TimeRange.End,
			"range":      resp.TimeRange.Range,
			"resolution": resp.TimeRange.Resolution,
		},
		"metrics": map[string]interface{}{
			"riskScore": map[string]interface{}{
				"current":   resp.Metrics.RiskScore.Current,
				"avg":       resp.Metrics.RiskScore.Avg,
				"min":       resp.Metrics.RiskScore.Min,
				"max":       resp.Metrics.RiskScore.Max,
				"unit":      resp.Metrics.RiskScore.Unit,
				"chart":     r.convertChartPoints(resp.Metrics.RiskScore.Chart),
				"threshold": resp.Metrics.RiskScore.Threshold,
			},
			"thermalTemp": map[string]interface{}{
				"current":   resp.Metrics.ThermalTemp.Current,
				"avg":       resp.Metrics.ThermalTemp.Avg,
				"min":       resp.Metrics.ThermalTemp.Min,
				"max":       resp.Metrics.ThermalTemp.Max,
				"unit":      resp.Metrics.ThermalTemp.Unit,
				"chart":     r.convertChartPoints(resp.Metrics.ThermalTemp.Chart),
				"threshold": resp.Metrics.ThermalTemp.Threshold,
			},
			"bufferLevel": map[string]interface{}{
				"current":   resp.Metrics.BufferLevel.Current,
				"avg":       resp.Metrics.BufferLevel.Avg,
				"min":       resp.Metrics.BufferLevel.Min,
				"max":       resp.Metrics.BufferLevel.Max,
				"unit":      resp.Metrics.BufferLevel.Unit,
				"chart":     r.convertChartPoints(resp.Metrics.BufferLevel.Chart),
				"threshold": resp.Metrics.BufferLevel.Threshold,
			},
			"uptime": map[string]interface{}{
				"current":   resp.Metrics.Uptime.Current,
				"avg":       0,
				"min":       0,
				"max":       0,
				"unit":      "s",
				"chart":     []interface{}{},
				"threshold": map[string]interface{}{"warning": 0, "critical": 0},
			},
		},
		"events": r.convertThresholdEvents(resp.Events),
	}, nil
}

// GetDeviceCommandHistory resolves the deviceCommandHistory query.
func (r *Resolver) GetDeviceCommandHistory(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	imei, ok := p.Args["imei"].(string)
	if !ok || imei == "" {
		return nil, r.Presenter.BadRequestError("device IMEI is required")
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device exists in organization
	_, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, imei, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	if r.HistoryService == nil {
		return nil, r.Presenter.InternalError("history service not available")
	}

	status, _ := p.Args["status"].(string)
	page, _ := p.Args["page"].(int)
	limit, _ := p.Args["limit"].(int)

	req := &cmdapp.GetHistoryRequest{
		DeviceID: imei,
		Status:  status,
		Page:    page,
		Limit:   limit,
	}

	if startTime, ok := p.Args["startTime"].(int64); ok && startTime > 0 {
		req.StartTime = startTime
	}

	if endTime, ok := p.Args["endTime"].(int64); ok && endTime > 0 {
		req.EndTime = endTime
	}

	resp, err := r.HistoryService.GetHistory(ctx, req)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get command history")
	}

	return map[string]interface{}{
		"commands": r.convertCommandHistory(resp.Commands),
		"pagination": map[string]interface{}{
			"page":       resp.Pagination.Page,
			"limit":      resp.Pagination.Limit,
			"total":      resp.Pagination.Total,
			"totalPages": resp.Pagination.TotalPages,
			"hasMore":    resp.Pagination.HasMore,
		},
	}, nil
}

// GetDashboardStats resolves the dashboardStats query.
func (r *Resolver) GetDashboardStats(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.DashboardSvc == nil {
		return nil, r.Presenter.InternalError("dashboard service not available")
	}

	stats, err := r.DashboardSvc.GetDashboardStats(ctx, op.ID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get dashboard stats")
	}

	return map[string]interface{}{
		"devices": map[string]interface{}{
			"total":  stats.Devices.Total,
			"online": stats.Devices.Online,
			"offline": stats.Devices.Offline,
		},
		"commands": map[string]interface{}{
			"totalToday": stats.Commands.TotalToday,
			"pending":   stats.Commands.Pending,
			"failed":    stats.Commands.Failed,
		},
		"activity": map[string]interface{}{
			"last24h": map[string]interface{}{
				"commands":        stats.Activity.Last24h.Commands,
				"registrations":   stats.Activity.Last24h.Registrations,
				"deregistrations": stats.Activity.Last24h.Deregistrations,
			},
		},
	}, nil
}

// Helper functions for converting response types

func (r *Resolver) convertChartPoints(points []appmetrics.MetricPointDTO) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(points))
	for _, p := range points {
		result = append(result, map[string]interface{}{
			"timestamp": p.Timestamp,
			"value":     p.Value,
		})
	}
	return result
}

func (r *Resolver) convertThresholdEvents(events []appmetrics.ThresholdEventDTO) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(events))
	for _, e := range events {
		result = append(result, map[string]interface{}{
			"timestamp": e.Timestamp,
			"type":      e.Type,
			"metric":    e.Metric,
			"value":     e.Value,
			"threshold": e.Threshold,
		})
	}
	return result
}

func (r *Resolver) convertLogEvents(events []logs.LogEvent) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(events))
	for _, e := range events {
		result = append(result, map[string]interface{}{
			"id":        e.ID,
			"type":      e.Type,
			"timestamp": e.Timestamp,
			"data":      e.Data,
		})
	}
	return result
}

func (r *Resolver) convertCommandHistory(commands []cmdapp.CommandEntry) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(commands))
	for _, c := range commands {
		entry := map[string]interface{}{
			"dispatchId": c.DispatchID,
			"commandId":  c.ID,
			"deviceId":   c.DeviceID,
			"command":    c.Command,
			"status":     c.Status,
			"sentAt":     time.UnixMilli(c.SentAt).Format(time.RFC3339),
		}
		if c.DeliveredAt > 0 {
			entry["deliveredAt"] = time.UnixMilli(c.DeliveredAt).Format(time.RFC3339)
		}
		if c.CompletedAt > 0 {
			entry["completedAt"] = time.UnixMilli(c.CompletedAt).Format(time.RFC3339)
		}
		if c.LatencyMs > 0 {
			entry["latencyMs"] = c.LatencyMs
		}
		result = append(result, entry)
	}
	return result
}

// GetDeviceInspection resolves the deviceInspection query.
func (r *Resolver) GetDeviceInspection(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	imei, ok := p.Args["imei"].(string)
	if !ok || imei == "" {
		return nil, r.Presenter.BadRequestError("IMEI is required")
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	_, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, imei, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	if r.DiagnosticsSvc == nil {
		return nil, r.Presenter.InternalError("diagnostics service not available")
	}

	inspection, err := r.DiagnosticsSvc.GetDeviceInspection(ctx, imei, orgID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get device inspection")
	}
	if inspection.Registration.RegisteredAt != nil {
		registeredAt = inspection.Registration.RegisteredAt.Format(time.RFC3339)
	}

	var lastSeen interface{} = nil
	if inspection.Connection.LastSeen != nil {
		lastSeen = inspection.Connection.LastSeen.Format(time.RFC3339)
	}

	var connectedAt interface{} = nil
	if inspection.Connection.ConnectedAt != nil {
		connectedAt = inspection.Connection.ConnectedAt.Format(time.RFC3339)
	}

	var fcmTokenRefreshedAt interface{} = nil
	if inspection.Registration.FCMTokenRefreshedAt != nil {
		fcmTokenRefreshedAt = inspection.Registration.FCMTokenRefreshedAt.Format(time.RFC3339)
	}

	var lastTimestamp interface{} = nil
	if !inspection.Telemetry.LastTimestamp.IsZero() {
		lastTimestamp = inspection.Telemetry.LastTimestamp.Format(time.RFC3339)
	}

	return map[string]interface{}{
		"identity": map[string]interface{}{
			"imei":         inspection.Identity.IMEI,
			"deviceName":   inspection.Identity.DeviceName,
			"model":        inspection.Identity.Model,
			"manufacturer": inspection.Identity.Manufacturer,
		},
		"software": map[string]interface{}{
			"osVersion":    inspection.Software.OSVersion,
			"appVersion":   inspection.Software.AppVersion,
			"securityPatch": inspection.Software.SecurityPatch,
			"buildId":       inspection.Software.BuildID,
		},
		"registration": map[string]interface{}{
			"status":               inspection.Registration.Status,
			"registeredAt":        registeredAt,
			"fcmTokenValid":       inspection.Registration.FCMTokenValid,
			"fcmTokenRefreshedAt": fcmTokenRefreshedAt,
			"commandSecretSet":     inspection.Registration.CommandSecretSet,
		},
		"connection": map[string]interface{}{
			"webSocketStatus": inspection.Connection.WebSocketStatus,
			"connectedAt":     connectedAt,
			"fcmStatus":       inspection.Connection.FCMStatus,
			"lastSeen":        lastSeen,
			"clientIp":        inspection.Connection.ClientIP,
			"protocol":        inspection.Connection.Protocol,
		},
		"telemetry": map[string]interface{}{
			"lastTimestamp":   lastTimestamp,
			"framesToday":     inspection.Telemetry.FramesToday,
			"avgLatencyMs":   inspection.Telemetry.AvgLatencyMs,
			"totalBytesToday": inspection.Telemetry.TotalBytesToday,
		},
	}, nil
	}
}

// GetDeviceTimeline resolves the deviceTimeline query.
func (r *Resolver) GetDeviceTimeline(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	imei, ok := p.Args["imei"].(string)
	if !ok || imei == "" {
		return nil, r.Presenter.BadRequestError("IMEI is required")
	}

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	_, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, imei, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	if r.DiagnosticsSvc == nil {
		return nil, r.Presenter.InternalError("diagnostics service not available")
	}

	eventType, _ := p.Args["eventType"].(string)
	startTime, _ := p.Args["startTime"].(int64)
	endTime, _ := p.Args["endTime"].(int64)
	limit, _ := p.Args["limit"].(int)
	cursor, _ := p.Args["cursor"].(string)

	req := &diagnosticsapp.TimelineRequest{
		IMEI:      imei,
		EventType: eventType,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     limit,
		Cursor:    cursor,
	}

	resp, err := r.DiagnosticsSvc.GetDeviceTimeline(ctx, imei, req, orgID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get device timeline")
	}

	events := make([]map[string]interface{}, 0, len(resp.Events))
	for _, e := range resp.Events {
		events = append(events, map[string]interface{}{
			"id":        e.ID,
			"type":      e.Type,
			"timestamp": time.UnixMilli(e.Timestamp).Format(time.RFC3339),
			"data":      e.Data,
		})
	}

	return map[string]interface{}{
		"events":     events,
		"hasMore":    resp.Pagination.HasMore,
		"nextCursor": resp.Pagination.NextCursor,
	}, nil
}
