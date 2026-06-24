// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"time"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/graphql-go/graphql"
)

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

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check device ownership - returns *dto.DeviceResponse
	dev, err := r.DeviceService.GetDeviceByOperator(ctx, id, op.ID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	// Log via presenter
	r.Presenter.DeviceView(ctx, op.ID, id)

	return r.deviceDTOToMap(dev), nil
}

// GetDevices resolves the devices list query.
func (r *Resolver) GetDevices(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
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

	// ListByOperatorPaginated returns []*domain/device.Device
	devices, err := r.DeviceService.ListByOperatorPaginated(ctx, op.ID, limit, offset)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list devices")
	}

	// Log via presenter
	r.Presenter.DeviceList(ctx, op.ID)

	result := make([]map[string]interface{}, 0, len(devices))
	for _, dev := range devices {
		result = append(result, r.deviceToMap(ctx, dev))
	}

	return result, nil
}

// GetDeviceCount resolves the deviceCount query.
func (r *Resolver) GetDeviceCount(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	count, err := r.DeviceService.CountByOperator(ctx, op.ID)
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

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	cmd, err := r.CommandService.GetCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	// Verify device ownership
	_, err = r.DeviceService.GetDeviceByOperator(ctx, cmd.DeviceID, op.ID)
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

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
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
	startTime, _ := p.Args["startTime"].(int64)
	endTime, _ := p.Args["endTime"].(int64)
	limit, _ := p.Args["limit"].(int)

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
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
		if entry.ReceivedAt.UnixMilli() <= endTime {
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

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
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

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
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

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
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

	return map[string]interface{}{
		"deviceId":      deviceID,
		"connected":     client.IsConnected(),
		"connectedAt":   time.Unix(metrics.LastConnectedAt, 0).Format(time.RFC3339),
		"lastMessageAt": nil,
		"uptimeSeconds": client.Uptime(),
	}, nil
}

// GetAllConnections resolves the allConnections query.
func (r *Resolver) GetAllConnections(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	if r.Hub == nil {
		return nil, r.Presenter.InternalError("WebSocket hub not available")
	}

	clients := r.Hub.Clients()
	result := make([]map[string]interface{}, 0, len(clients))

	for deviceID := range clients {
		// Only include devices owned by this operator
		_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
		if err != nil {
			continue
		}

		client := clients[deviceID]
		metrics := client.GetMetrics()

		result = append(result, map[string]interface{}{
			"deviceId":      deviceID,
			"connected":     client.IsConnected(),
			"connectedAt":   time.Unix(metrics.LastConnectedAt, 0).Format(time.RFC3339),
			"lastMessageAt": nil,
			"uptimeSeconds": client.Uptime(),
		})
	}

	return result, nil
}
