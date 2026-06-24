// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"encoding/json"
	"time"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	"github.com/graphql-go/graphql"
)

// ============================================================
// Mutation Resolvers
// ============================================================

// UpdateFCMToken resolves the updateFCMToken mutation.
func (r *Resolver) UpdateFCMToken(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)
	token, _ := p.Args["token"].(string)

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	if token == "" {
		return nil, r.Presenter.BadRequestError("FCM token is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device ownership - returns *dto.DeviceResponse
	dev, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	// Update FCM token
	err = r.DeviceService.UpdateFCMToken(ctx, deviceID, token)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to update FCM token")
	}

	// Log via presenter
	r.Presenter.FCMTokenUpdate(ctx, op.ID, deviceID)

	// Return the updated device as a map
	return r.deviceDTOToMap(dev), nil
}

// DeleteDevice resolves the deleteDevice mutation.
func (r *Resolver) DeleteDevice(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["id"].(string)

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

	err = r.DeviceService.DeleteDevice(ctx, deviceID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to delete device")
	}

	// Log via presenter
	r.Presenter.DeviceDelete(ctx, op.ID, deviceID)

	return true, nil
}

// SendCommand resolves the sendCommand mutation.
func (r *Resolver) SendCommand(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["deviceId"].(string)
	cmdStr, _ := p.Args["command"].(string)
	args, _ := p.Args["args"].(map[string]interface{})

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	if cmdStr == "" {
		return nil, r.Presenter.BadRequestError("command is required")
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

	// Use command service for proper command creation and idempotency
	cmdReq := &dto.SendCommandRequest{
		DeviceID: deviceID,
		Command:  cmdStr,
		Args:     args,
	}

	cmdResp, err := r.CommandService.SendCommand(ctx, cmdReq)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to send command")
	}

	// Send via WebSocket if device is online
	delivery := "queued"

	if r.Hub != nil && r.Hub.Online(deviceID) {
		// Build command frame
		argsJSON, err := json.Marshal(args)
		if err != nil {
			argsJSON = []byte("{}")
		}

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
				// Log but don't fail - command remains queued
				r.Presenter.LogAction(ctx, op.ID, "fcm_send_failed", "device", deviceID)
			} else {
				delivery = "queued_fcm"
			}
		}
	}

	// Log via presenter
	r.Presenter.CommandSend(ctx, op.ID, deviceID, cmdResp.CommandID)

	return map[string]interface{}{
		"dispatchId":   cmdResp.DispatchID,
		"commandId":    cmdResp.CommandID,
		"status":       delivery,
		"deviceOnline": delivery == "sent",
	}, nil
}

// RetryCommand resolves the retryCommand mutation.
func (r *Resolver) RetryCommand(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	dispatchID, _ := p.Args["dispatchId"].(string)

	if dispatchID == "" {
		return nil, r.Presenter.BadRequestError("dispatch ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Get command to find device
	cmd, err := r.CommandService.GetCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	// Verify device ownership
	_, err = r.DeviceService.GetDeviceByOperator(ctx, cmd.DeviceID, op.ID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	newCmd, err := r.CommandService.RetryCommand(ctx, dispatchID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to retry command")
	}

	return map[string]interface{}{
		"dispatchId": newCmd.DispatchID,
		"commandId":  newCmd.CommandID,
		"status":     newCmd.Status,
	}, nil
}

// CancelCommand resolves the cancelCommand mutation.
func (r *Resolver) CancelCommand(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	dispatchID, _ := p.Args["dispatchId"].(string)

	if dispatchID == "" {
		return nil, r.Presenter.BadRequestError("dispatch ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Get command to find device
	cmd, err := r.CommandService.GetCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	// Verify device ownership
	_, err = r.DeviceService.GetDeviceByOperator(ctx, cmd.DeviceID, op.ID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	err = r.CommandService.CancelCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to cancel command")
	}

	// Log via presenter
	r.Presenter.CommandCancel(ctx, op.ID, cmd.CommandID)

	return true, nil
}

// DisconnectDevice resolves the disconnectDevice mutation.
func (r *Resolver) DisconnectDevice(p graphql.ResolveParams) (interface{}, error) {
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
		return false, nil
	}

	client := r.Hub.GetClient(deviceID)
	if client == nil {
		return false, nil
	}

	// Close the connection
	_ = client.Conn.Close()

	// Log via presenter
	r.Presenter.LogAction(ctx, op.ID, "graphql_device_disconnect", "device", deviceID)

	return true, nil
}
