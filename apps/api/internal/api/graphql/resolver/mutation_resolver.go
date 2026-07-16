// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"encoding/json"
	"time"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	domainoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
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
	orgID, _ := p.Args["organizationId"].(string)

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	if token == "" {
		return nil, r.Presenter.BadRequestError("FCM token is required")
	}

	if orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device exists in organization using org-scoped method
	_, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
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

	// Fetch updated device to return fresh data
	updatedDev, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	return r.deviceDTOToMap(&updatedDev.Device), nil
}

// DeleteDevice resolves the deleteDevice mutation.
func (r *Resolver) DeleteDevice(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	deviceID, _ := p.Args["id"].(string)
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

	// Use organization-scoped deregistration method
	_, err := r.DeviceService.DeregisterDeviceByOrganization(ctx, deviceID, orgID, true)
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
	orgID, _ := p.Args["organizationId"].(string)

	if deviceID == "" {
		return nil, r.Presenter.BadRequestError("device ID is required")
	}

	if cmdStr == "" {
		return nil, r.Presenter.BadRequestError("command is required")
	}

	if orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Verify device exists in organization using org-scoped method
	_, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
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
	orgID, _ := p.Args["organizationId"].(string)

	if dispatchID == "" {
		return nil, r.Presenter.BadRequestError("dispatch ID is required")
	}

	if orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
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

	// Verify device exists in organization
	_, err = r.DeviceService.GetDeviceDetailByOrganization(ctx, cmd.DeviceID, orgID)
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
	orgID, _ := p.Args["organizationId"].(string)

	if dispatchID == "" {
		return nil, r.Presenter.BadRequestError("dispatch ID is required")
	}

	if orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
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

	// Verify device exists in organization
	_, err = r.DeviceService.GetDeviceDetailByOrganization(ctx, cmd.DeviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	err = r.CommandService.CancelCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to cancel command")
	}

	// Log via presenter
	r.Presenter.CommandCancel(ctx, op.ID, cmd.CommandID)

	return map[string]interface{}{
		"dispatchId":  dispatchID,
		"cancelledAt": time.Now().UnixMilli(),
		"status":      "cancelled",
	}, nil
}

// DisconnectDevice resolves the disconnectDevice mutation.
func (r *Resolver) DisconnectDevice(p graphql.ResolveParams) (interface{}, error) {
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

// ============================================================
// Settings Mutation Resolvers
// ============================================================

// UpdateMyThresholds resolves the updateMyThresholds mutation.
func (r *Resolver) UpdateMyThresholds(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Parse thresholds input
	input, ok := p.Args["input"].(map[string]interface{})
	if !ok {
		return nil, r.Presenter.BadRequestError("invalid input")
	}

	thresholdsInput := &domainoperator.ThresholdsInput{}
	if v, ok := input["riskWarn"].(int); ok {
		thresholdsInput.RiskWarn = &v
	}
	if v, ok := input["riskCrit"].(int); ok {
		thresholdsInput.RiskCrit = &v
	}
	if v, ok := input["thermalWarn"].(int); ok {
		thresholdsInput.ThermalWarn = &v
	}
	if v, ok := input["thermalCrit"].(int); ok {
		thresholdsInput.ThermalCrit = &v
	}
	if v, ok := input["bufferWarn"].(int); ok {
		thresholdsInput.BufferWarn = &v
	}
	if v, ok := input["bufferCrit"].(int); ok {
		thresholdsInput.BufferCrit = &v
	}

	thresholds, err := r.ThresholdService.UpdateThresholds(ctx, op.ID, thresholdsInput)
	if err != nil {
		if err == domainoperator.ErrValidation {
			return nil, r.Presenter.BadRequestError(err.Error())
		}
		return nil, r.Presenter.InternalError("failed to update thresholds")
	}

	return thresholds, nil
}

// UpdateMyNotifications resolves the updateMyNotifications mutation.
func (r *Resolver) UpdateMyNotifications(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Parse notification input
	input, ok := p.Args["input"].(map[string]interface{})
	if !ok {
		return nil, r.Presenter.BadRequestError("invalid input")
	}

	notifInput := &domainoperator.NotificationInput{}

	if v, ok := input["enabled"].(bool); ok {
		notifInput.Enabled = &v
	}

	if channels, ok := input["channels"].([]interface{}); ok {
		ch := make([]string, len(channels))
		for i, c := range channels {
			if str, ok := c.(string); ok {
				ch[i] = str
			}
		}
		notifInput.Channels = &ch
	}

	if email, ok := input["email"].(map[string]interface{}); ok {
		emailInput := &domainoperator.EmailNotificationInput{}
		if v, ok := email["thresholdBreach"].(bool); ok {
			emailInput.ThresholdBreach = &v
		}
		if v, ok := email["deviceOffline"].(bool); ok {
			emailInput.DeviceOffline = &v
		}
		if v, ok := email["deviceOnline"].(bool); ok {
			emailInput.DeviceOnline = &v
		}
		notifInput.Email = emailInput
	}

	if webhook, ok := input["webhook"].(map[string]interface{}); ok {
		webhookInput := &domainoperator.WebhookNotificationInput{}
		if v, ok := webhook["enabled"].(bool); ok {
			webhookInput.Enabled = &v
		}
		if v, ok := webhook["url"].(string); ok {
			webhookInput.URL = &v
		}
		if v, ok := webhook["types"].([]interface{}); ok {
			types := make([]string, 0, len(v))
			for _, t := range v {
				if str, ok := t.(string); ok {
					types = append(types, str)
				}
			}
			webhookInput.Types = types
		}
		notifInput.Webhook = webhookInput
	}

	if push, ok := input["push"].(map[string]interface{}); ok {
		pushInput := &domainoperator.PushNotificationInput{}
		if v, ok := push["thresholdBreach"].(bool); ok {
			pushInput.ThresholdBreach = &v
		}
		if v, ok := push["deviceOffline"].(bool); ok {
			pushInput.DeviceOffline = &v
		}
		if v, ok := push["deviceOnline"].(bool); ok {
			pushInput.DeviceOnline = &v
		}
		notifInput.Push = pushInput
	}

	notifications, err := r.NotificationSvc.UpdateNotifications(ctx, op.ID, notifInput)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to update notifications")
	}

	return notifications, nil
}

// TestWebhook resolves the testWebhook mutation.
func (r *Resolver) TestWebhook(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	url, _ := p.Args["url"].(string)
	if url == "" {
		return nil, r.Presenter.BadRequestError("url is required")
	}

	result, err := r.WebhookClient.Test(ctx, url)
	if err != nil {
		return nil, r.Presenter.InternalError("webhook test failed: " + err.Error())
	}

	return result, nil
}

// RotateWebhookSecret resolves the rotateWebhookSecret mutation.
func (r *Resolver) RotateWebhookSecret(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	secret, err := r.NotificationSvc.RotateWebhookSecret(ctx, op.ID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to rotate webhook secret")
	}

	return secret, nil
}
