// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/validator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	cmdapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/graphql-go/graphql"
)

// Resolver is the root GraphQL resolver.
type Resolver struct {
	// Services
	DeviceService  *device.Service
	CommandService *cmdapp.Service
	Hub            *hub.Hub
	TelemetryRepo  *storage.TelemetryRepository
	FCMNotifier    fcm.Notifier

	// Middleware
	AuthMiddleware *middleware.AuthMiddleware

	// Utilities
	Log       *slog.Logger
	Validator *validator.Validator
}

// NewResolver creates a new GraphQL resolver.
func NewResolver(
	deviceService *device.Service,
	commandService *cmdapp.Service,
	hub *hub.Hub,
	telemetryRepo *storage.TelemetryRepository,
	fcmNotifier fcm.Notifier,
	authMiddleware *middleware.AuthMiddleware,
	log *slog.Logger,
) *Resolver {
	return &Resolver{
		DeviceService:  deviceService,
		CommandService: commandService,
		Hub:            hub,
		TelemetryRepo:  telemetryRepo,
		FCMNotifier:    fcmNotifier,
		AuthMiddleware: authMiddleware,
		Log:            log,
		Validator:      validator.New(),
	}
}

// RequireAuth ensures the operator is authenticated.
func (r *Resolver) RequireAuth(ctx context.Context) (*context.Context, error) {
	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}
	return &ctx, nil
}

// ============================================================
// Query Resolvers
// ============================================================

// GetDevice resolves the device query.
func (r *Resolver) GetDevice(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	id, ok := p.Args["id"].(string)
	if !ok || id == "" {
		return nil, errors.BadRequest("device ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Check device ownership
	dev, err := r.DeviceService.GetDeviceByOperator(ctx, id, op.ID)
	if err != nil {
		return nil, errors.NotFound("device not found")
	}

	return r.deviceToMap(ctx, dev), nil
}

// GetDevices resolves the devices list query.
func (r *Resolver) GetDevices(p graphql.GQLParams) (interface{}, error) {
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
		return nil, errors.ErrUnauthorized
	}

	devices, err := r.DeviceService.ListDevices(ctx, op.ID, limit, offset)
	if err != nil {
		r.Log.Error("failed to list devices", "err", err)
		return nil, errors.Internal("failed to list devices")
	}

	result := make([]map[string]interface{}, 0, len(devices))
	for _, dev := range devices {
		result = append(result, r.deviceToMap(ctx, dev))
	}

	return result, nil
}

// GetDeviceCount resolves the deviceCount query.
func (r *Resolver) GetDeviceCount(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	count, err := r.DeviceService.CountDevices(ctx, op.ID)
	if err != nil {
		r.Log.Error("failed to count devices", "err", err)
		return nil, errors.Internal("failed to count devices")
	}

	return count, nil
}

// GetCommand resolves the command query.
func (r *Resolver) GetCommand(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	dispatchID, ok := p.Args["dispatchId"].(string)
	if !ok || dispatchID == "" {
		return nil, errors.BadRequest("dispatch ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	cmd, err := r.CommandService.GetCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, errors.NotFound("command not found")
	}

	// Verify device ownership
	_, err = r.DeviceService.GetDeviceByOperator(ctx, cmd.DeviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("command not found")
	}

	return r.commandToMap(cmd), nil
}

// GetPendingCommands resolves the pendingCommands query.
func (r *Resolver) GetPendingCommands(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	deviceID, ok := p.Args["deviceId"].(string)
	if !ok || deviceID == "" {
		return nil, errors.BadRequest("device ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("device not found")
	}

	cmds, err := r.CommandService.GetPendingCommands(ctx, deviceID)
	if err != nil {
		r.Log.Error("failed to get pending commands", "err", err)
		return nil, errors.Internal("failed to get pending commands")
	}

	result := make([]map[string]interface{}, 0, len(cmds))
	for _, cmd := range cmds {
		result = append(result, r.commandToMap(cmd))
	}

	return result, nil
}

// GetTelemetryHistory resolves the telemetryHistory query.
func (r *Resolver) GetTelemetryHistory(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)
	startTime, _ := p.Args["startTime"].(int64)
	endTime, _ := p.Args["endTime"].(int64)
	limit, _ := p.Args["limit"].(int)

	if deviceID == "" {
		return nil, errors.BadRequest("device ID is required")
	}
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("device not found")
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
		r.Log.Error("failed to query telemetry", "err", err)
		return nil, errors.Internal("failed to query telemetry")
	}

	result := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		if entry.ReceivedAt.UnixMilli() <= endTime {
			result = append(result, map[string]interface{}{
				"id":           entry.ID,
				"deviceId":     entry.DeviceID,
				"receivedAt":   entry.ReceivedAt.Format(time.RFC3339),
				"riskScore":    entry.RiskScore,
				"bufferLevel":  entry.BufferLevel,
				"thermalTemp":  entry.ThermalTemp,
				"payload":      string(entry.Payload),
			})
		}
	}

	return result, nil
}

// GetLatestTelemetry resolves the latestTelemetry query.
func (r *Resolver) GetLatestTelemetry(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)

	if deviceID == "" {
		return nil, errors.BadRequest("device ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("device not found")
	}

	entries, err := r.TelemetryRepo.List(ctx, deviceID, 1)
	if err != nil || len(entries) == 0 {
		return nil, errors.NotFound("no telemetry found")
	}

	entry := entries[0]
	return map[string]interface{}{
		"id":           entry.ID,
		"deviceId":     entry.DeviceID,
		"receivedAt":   entry.ReceivedAt.Format(time.RFC3339),
		"riskScore":    entry.RiskScore,
		"bufferLevel":  entry.BufferLevel,
		"thermalTemp":  entry.ThermalTemp,
		"payload":      string(entry.Payload),
	}, nil
}

// GetTelemetryStats resolves the telemetryStats query.
func (r *Resolver) GetTelemetryStats(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)

	if deviceID == "" {
		return nil, errors.BadRequest("device ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("device not found")
	}

	entries, err := r.TelemetryRepo.List(ctx, deviceID, 100)
	if err != nil || len(entries) == 0 {
		return nil, errors.NotFound("no telemetry found")
	}

	// Calculate stats
	var totalRisk, totalBuffer, totalTemp int
	var minRisk, maxRisk int = 999, -1
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
func (r *Resolver) GetConnectionStatus(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)

	if deviceID == "" {
		return nil, errors.BadRequest("device ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("device not found")
	}

	if r.Hub == nil {
		return nil, errors.Internal("WebSocket hub not available")
	}

	client := r.Hub.GetClient(deviceID)
	if client == nil {
		return map[string]interface{}{
			"deviceId":       deviceID,
			"connected":      false,
			"connectedAt":    nil,
			"lastMessageAt":  nil,
			"uptimeSeconds":  0,
		}, nil
	}

	metrics := client.GetMetrics()
	return map[string]interface{}{
		"deviceId":       deviceID,
		"connected":      client.IsConnected(),
		"connectedAt":    time.Unix(metrics.LastConnectedAt, 0).Format(time.RFC3339),
		"lastMessageAt":  nil,
		"uptimeSeconds":  client.Uptime(),
	}, nil
}

// GetAllConnections resolves the allConnections query.
func (r *Resolver) GetAllConnections(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	if r.Hub == nil {
		return nil, errors.Internal("WebSocket hub not available")
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
			"deviceId":       deviceID,
			"connected":      client.IsConnected(),
			"connectedAt":    time.Unix(metrics.LastConnectedAt, 0).Format(time.RFC3339),
			"lastMessageAt":  nil,
			"uptimeSeconds":  client.Uptime(),
		})
	}

	return result, nil
}

// ============================================================
// Mutation Resolvers
// ============================================================

// UpdateFCMToken resolves the updateFCMToken mutation.
func (r *Resolver) UpdateFCMToken(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)
	token, _ := p.Args["token"].(string)

	if deviceID == "" {
		return nil, errors.BadRequest("device ID is required")
	}
	if token == "" {
		return nil, errors.BadRequest("FCM token is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verify device ownership
	dev, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("device not found")
	}

	// Update FCM token
	updated, err := r.DeviceService.UpdateFCMToken(ctx, deviceID, token)
	if err != nil {
		r.Log.Error("failed to update FCM token", "err", err)
		return nil, errors.Internal("failed to update FCM token")
	}

	return r.deviceToMap(ctx, updated), nil
}

// DeleteDevice resolves the deleteDevice mutation.
func (r *Resolver) DeleteDevice(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["id"].(string)

	if deviceID == "" {
		return nil, errors.BadRequest("device ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("device not found")
	}

	err = r.DeviceService.DeleteDevice(ctx, deviceID)
	if err != nil {
		r.Log.Error("failed to delete device", "err", err)
		return nil, errors.Internal("failed to delete device")
	}

	return true, nil
}

// SendCommand resolves the sendCommand mutation.
func (r *Resolver) SendCommand(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)
	cmdStr, _ := p.Args["command"].(string)
	args, _ := p.Args["args"].(map[string]interface{})

	if deviceID == "" {
		return nil, errors.BadRequest("device ID is required")
	}
	if cmdStr == "" {
		return nil, errors.BadRequest("command is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("device not found")
	}

	// Use command service for proper command creation and idempotency
	cmdReq := &dto.SendCommandRequest{
		DeviceID: deviceID,
		Command:  cmdStr,
		Args:     args,
	}

	cmdResp, err := r.CommandService.SendCommand(ctx, cmdReq)
	if err != nil {
		r.Log.Error("failed to send command", "err", err)
		return nil, errors.Internal("failed to send command")
	}

	// Send via WebSocket if device is online
	delivery := "queued"
	if r.Hub != nil && r.Hub.Online(deviceID) {
		// Build command frame
		argsJSON, _ := json.Marshal(args)
		frame := command.CommandFrame{
			Type:       cmdStr,
			Command:    cmdStr,
			DispatchID: cmdResp.DispatchID,
			Args:       argsJSON,
			Timestamp:  time.Now().UnixMilli(),
		}
		if r.Hub.Send(deviceID, frame) {
			delivery = "sent"
			// Mark as delivered
			_ = r.CommandService.MarkDelivered(ctx, cmdResp.CommandID)
		}
	}

	// If not sent via WebSocket, try FCM
	if delivery == "queued" && r.FCMNotifier != nil {
		dev, _ := r.DeviceService.GetDevice(ctx, deviceID)
		if dev != nil && dev.FCMToken != "" {
			wake := fcm.SilentWake{
				Token:      dev.FCMToken,
				Command:    cmdStr,
				DispatchID: cmdResp.DispatchID,
				DeviceID:   deviceID,
			}
			if err := r.FCMNotifier.SendSilentWake(ctx, wake); err != nil {
				r.Log.Warn("fcm wake failed", "err", err)
			} else {
				delivery = "queued_fcm"
			}
		}
	}

	return map[string]interface{}{
		"dispatchId":   cmdResp.DispatchID,
		"commandId":    cmdResp.CommandID,
		"status":       delivery,
		"deviceOnline": delivery == "sent",
	}, nil
}

// RetryCommand resolves the retryCommand mutation.
func (r *Resolver) RetryCommand(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	dispatchID, _ := p.Args["dispatchId"].(string)

	if dispatchID == "" {
		return nil, errors.BadRequest("dispatch ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Get command to find device
	cmd, err := r.CommandService.GetCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, errors.NotFound("command not found")
	}

	// Verify device ownership
	_, err = r.DeviceService.GetDeviceByOperator(ctx, cmd.DeviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("command not found")
	}

	newCmd, err := r.CommandService.RetryCommand(ctx, dispatchID)
	if err != nil {
		r.Log.Error("failed to retry command", "err", err)
		return nil, errors.Internal("failed to retry command")
	}

	return r.commandToMap(newCmd), nil
}

// CancelCommand resolves the cancelCommand mutation.
func (r *Resolver) CancelCommand(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	dispatchID, _ := p.Args["dispatchId"].(string)

	if dispatchID == "" {
		return nil, errors.BadRequest("dispatch ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Get command to find device
	cmd, err := r.CommandService.GetCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, errors.NotFound("command not found")
	}

	// Verify device ownership
	_, err = r.DeviceService.GetDeviceByOperator(ctx, cmd.DeviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("command not found")
	}

	err = r.CommandService.CancelCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		r.Log.Error("failed to cancel command", "err", err)
		return nil, errors.Internal("failed to cancel command")
	}

	return true, nil
}

// DisconnectDevice resolves the disconnectDevice mutation.
func (r *Resolver) DisconnectDevice(p graphql.GQLParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)

	if deviceID == "" {
		return nil, errors.BadRequest("device ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, errors.ErrUnauthorized
	}

	// Verify device ownership
	_, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
	if err != nil {
		return nil, errors.NotFound("device not found")
	}

	if r.Hub == nil {
		return false, nil
	}

	client := r.Hub.GetClient(deviceID)
	if client == nil {
		return false, nil
	}

	// Close the connection
	client.Conn.Close()

	return true, nil
}

// ============================================================
// Helper Methods
// ============================================================

func (r *Resolver) deviceToMap(ctx context.Context, dev *device.Device) map[string]interface{} {
	return map[string]interface{}{
		"id":        dev.ID,
		"name":      "", // Device entity doesn't have a name field
		"online":    r.Hub != nil && r.Hub.Online(dev.ID),
		"lastSeen":  time.UnixMilli(dev.LastSeen).Format(time.RFC3339),
		"fcmToken":  dev.FCMToken,
		"version":   dev.AppVersion,
		"createdAt": dev.CreatedAt.Format(time.RFC3339),
	}
}

func (r *Resolver) commandToMap(cmd *command.Command) map[string]interface{} {
	var args map[string]interface{}
	if len(cmd.Args) > 0 {
		_ = json.Unmarshal(cmd.Args, &args)
	}

	var deliveredAt interface{}
	if cmd.DeliveredAt != nil {
		deliveredAt = time.UnixMilli(*cmd.DeliveredAt).Format(time.RFC3339)
	}

	return map[string]interface{}{
		"dispatchId":  cmd.DispatchID,
		"commandId":   cmd.ID,
		"deviceId":    cmd.DeviceID,
		"command":     cmd.Command,
		"args":        args,
		"status":      string(cmd.Status),
		"createdAt":   cmd.CreatedAt.Format(time.RFC3339),
		"deliveredAt": deliveredAt,
	}
}