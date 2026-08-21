// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"errors"
	"time"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	alertapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/alert"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	notifications "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/notifications"
	serviceaccount "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/serviceaccount"
	appannotation "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/annotation"
	annotationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/annotation"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	notificationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
	domainoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	orgdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	sadomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/serviceaccount"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/webhook"
	"github.com/graphql-go/graphql"
)

// validateWebhookURL delegates to the shared webhook package validator.
func validateWebhookURL(rawURL string) error {
	return webhook.ValidateURL(rawURL)
}

// ============================================================.
// Mutation Resolvers.
// ============================================================.

// UpdateFCMToken resolves the updateFCMToken mutation.
func (r *Resolver) UpdateFCMToken(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	deviceID, ok := p.Args["deviceId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("deviceId must be a string")
	}
	token, ok := p.Args["token"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("token must be a string")
	}
	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

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

	// Verify device exists in organization using org-scoped method.
	_, err = r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	// Update FCM token.
	err = r.DeviceService.UpdateFCMToken(ctx, deviceID, token)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to update FCM token")
	}

	// Log via presenter.
	r.Presenter.FCMTokenUpdate(ctx, op.ID, deviceID)

	// Fetch updated device to return fresh data.
	updatedDev, err := r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	return r.deviceDetailToMap(updatedDev), nil
}

// DeleteDevice resolves the deleteDevice mutation.
func (r *Resolver) DeleteDevice(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	deviceID, ok := p.Args["id"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("id must be a string")
	}
	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

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

	// Use organization-scoped deregistration method.
	_, err = r.DeviceService.DeregisterDeviceByOrganization(ctx, deviceID, orgID, true)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to delete device")
	}

	// Log via presenter.
	r.Presenter.DeviceDelete(ctx, op.ID, deviceID)

	return true, nil
}

// SendCommand resolves the sendCommand mutation.
func (r *Resolver) SendCommand(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	deviceID, ok := p.Args["deviceId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("deviceId must be a string")
	}
	cmdStr, ok := p.Args["command"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("command must be a string")
	}
	args, ok := p.Args["args"].(map[string]interface{})
	if !ok {
		args = nil // args is optional, default to nil.
	}
	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

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

	// Fetch the device entity (org-scoped) so authorization can check ownership.
	dev, err := r.DeviceService.DeviceRepo().FindByIMEIAndOrganization(ctx, deviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("device not found")
	}

	// Run the same authorization the REST command path enforces: role tier
	// (viewers cannot command), per-device ownership (non-admins only their own
	// devices), and the risk gate (MFA + confirmation token). Without this the
	// GraphQL path could dispatch factory_reset with no confirmation at all.
	confirmationToken, _ := p.Args["confirmationToken"].(string)
	if outcome := r.authorizeCommand(ctx, op, orgID, dev, cmdStr, confirmationToken); !outcome.Allowed {
		r.Presenter.CommandBlocked(ctx, op.ID, deviceID, cmdStr, string(outcome.Tier), outcome.Reason)
		if outcome.NeedsConfirmation {
			return nil, r.Presenter.ConflictError(outcome.Reason)
		}
		return nil, r.Presenter.ForbiddenError(outcome.Reason)
	}

	// Use command service for proper command creation and idempotency.
	cmdReq := &dto.SendCommandRequest{
		DeviceID: deviceID,
		Command:  cmdStr,
		Args:     args,
	}

	cmdResp, err := r.CommandService.SendCommand(ctx, cmdReq)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to send command")
	}

	// Command is now persisted in DB with status="pending".
	// The CommandOutbox background worker will:.
	// 1. Poll for pending commands.
	// 2. Attempt delivery via WebSocket (with confirmation) or FCM.
	// 3. Mark as delivered only on confirmed receipt.
	// 4. Retry with exponential backoff on failure.
	// 5. Mark as failed after MaxRetries exceeded.
	//
	// This implements the transactional outbox pattern:.
	// - Command write is atomic with DB transaction.
	// - Delivery is asynchronous, isolated from the HTTP request.
	// - No command loss on delivery failure (automatic retry).

	// Log via presenter.
	r.Presenter.CommandSend(ctx, op.ID, deviceID, cmdResp.CommandID)

	return map[string]interface{}{
		"dispatchId":   cmdResp.DispatchID,
		"commandId":    cmdResp.CommandID,
		"status":       "pending",
		"deviceOnline": r.Hub != nil && r.Hub.Online(deviceID),
	}, nil
}

// RetryCommand resolves the retryCommand mutation.
func (r *Resolver) RetryCommand(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	dispatchID, ok := p.Args["dispatchId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("dispatchId must be a string")
	}
	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

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

	// Get command to find device.
	cmd, err := r.CommandService.GetCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	// Verify device exists in organization.
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

	dispatchID, ok := p.Args["dispatchId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("dispatchId must be a string")
	}
	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

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

	// Get command to find device.
	cmd, err := r.CommandService.GetCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	// Verify device exists in organization.
	_, err = r.DeviceService.GetDeviceDetailByOrganization(ctx, cmd.DeviceID, orgID)
	if err != nil {
		return nil, r.Presenter.NotFoundError("command not found")
	}

	err = r.CommandService.CancelCommandByDispatchID(ctx, dispatchID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to cancel command")
	}

	// Log via presenter.
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

	deviceID, ok := p.Args["deviceId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("deviceId must be a string")
	}
	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

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

	// Verify device exists in organization.
	_, err = r.DeviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
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

	// Close the connection.
	_ = client.Conn.Close()

	// Log via presenter.
	r.Presenter.LogAction(ctx, op.ID, "graphql_device_disconnect", "device", deviceID)

	return true, nil
}

// ============================================================.
// Settings Mutation Resolvers.
// ============================================================.

// UpdateMyNotifications resolves the updateMyNotifications mutation.
func (r *Resolver) UpdateMyNotifications(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	input, ok := p.Args["input"].(map[string]interface{})
	if !ok {
		return nil, r.Presenter.BadRequestError("invalid input")
	}

	notifInput := r.parseNotificationInput(input)

	notifications, err := r.NotificationSvc.UpdateNotifications(ctx, op.ID, notifInput)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to update notifications")
	}

	return notifications, nil
}

// parseNotificationInput parses notification input from GraphQL args.
func (r *Resolver) parseNotificationInput(input map[string]interface{}) *domainoperator.NotificationInput {
	notifInput := &domainoperator.NotificationInput{}

	if v, ok := input["enabled"].(bool); ok {
		notifInput.Enabled = &v
	}

	if channels, ok := input["channels"].([]interface{}); ok {
		notifInput.Channels = r.parseStringArray(channels)
	}

	if email, ok := input["email"].(map[string]interface{}); ok {
		notifInput.Email = r.parseEmailInput(email)
	}

	if webhook, ok := input["webhook"].(map[string]interface{}); ok {
		notifInput.Webhook = r.parseWebhookInput(webhook)
	}

	if push, ok := input["push"].(map[string]interface{}); ok {
		notifInput.Push = r.parsePushInput(push)
	}

	return notifInput
}

// parseStringArray parses a slice of interface{} into a slice of string.
func (r *Resolver) parseStringArray(arr []interface{}) *[]string {
	if arr == nil {
		return nil
	}
	result := make([]string, len(arr))
	for i, c := range arr {
		if str, ok := c.(string); ok {
			result[i] = str
		}
	}
	return &result
}

// parseEmailInput parses email notification settings.
func (r *Resolver) parseEmailInput(email map[string]interface{}) *domainoperator.EmailNotificationInput {
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
	return emailInput
}

// parseWebhookInput parses webhook notification settings.
func (r *Resolver) parseWebhookInput(webhook map[string]interface{}) *domainoperator.WebhookNotificationInput {
	webhookInput := &domainoperator.WebhookNotificationInput{}
	if v, ok := webhook["enabled"].(bool); ok {
		webhookInput.Enabled = &v
	}
	if v, ok := webhook["url"].(string); ok {
		webhookInput.URL = &v
	}
	if v, ok := webhook["types"].([]interface{}); ok {
		webhookInput.Types = r.parseStringSlice(v)
	}
	return webhookInput
}

// parsePushInput parses push notification settings.
func (r *Resolver) parsePushInput(push map[string]interface{}) *domainoperator.PushNotificationInput {
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
	return pushInput
}

// parseStringSlice parses a slice of interface{} into a slice of string.
func (r *Resolver) parseStringSlice(v []interface{}) []string {
	types := make([]string, 0, len(v))
	for _, t := range v {
		if str, ok := t.(string); ok {
			types = append(types, str)
		}
	}
	return types
}

// UpdateDeviceSettings resolves the updateDeviceSettings mutation.
func (r *Resolver) UpdateDeviceSettings(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

	deviceImei, ok := p.Args["deviceImei"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("device IMEI is required")
	}

	// Check if operator is a member of the organization.
	err = r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID)
	if err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	input, ok := p.Args["input"].(map[string]interface{})
	if !ok {
		return nil, r.Presenter.BadRequestError("invalid input")
	}

	// Parse device settings input.
	settingsReq := &devicedomain.UpdateDeviceSettingsRequest{}

	if v, ok := input["customName"].(string); ok {
		settingsReq.CustomName = &v
	}

	if v, ok := input["location"].(string); ok {
		settingsReq.Location = &v
	}

	if metadata, ok := input["metadata"].([]interface{}); ok {
		settingsReq.Metadata = make(map[string]string)
		for _, m := range metadata {
			if kv, ok := m.(map[string]interface{}); ok {
				if k, ok := kv["key"].(string); ok {
					if v, ok := kv["value"].(string); ok {
						settingsReq.Metadata[k] = v
					}
				}
			}
		}
	}

	if thresholds, ok := input["thresholds"].(map[string]interface{}); ok {
		settingsReq.Thresholds = parseDeviceThresholds(thresholds)
	}

	// First, ensure device settings exist (create with defaults if not).
	if _, err = r.DeviceSettingsService.GetOrCreateSettings(ctx, deviceImei); err != nil {
		return nil, r.Presenter.InternalError("failed to create device settings")
	}

	// Now update with the requested changes.
	settings, err := r.DeviceSettingsService.UpdateSettings(ctx, deviceImei, settingsReq)
	if err != nil {
		if errors.Is(err, devicedomain.ErrInvalidThreshold) {
			return nil, r.Presenter.BadRequestError("invalid threshold values: warning must be less than critical")
		}
		return nil, r.Presenter.InternalError("failed to update device settings")
	}

	// Get organization settings for effective thresholds.
	orgSettings, err := r.OrgSettingsService.GetSettings(ctx, orgID)
	if err != nil {
		if errors.Is(err, orgdomain.ErrSettingsNotFound) {
			return nil, r.Presenter.NotFoundError("organization settings not found")
		}
		return nil, r.Presenter.InternalError("failed to get organization settings")
	}

	// Convert org thresholds to device thresholds for resolution.
	orgThresholds := devicedomain.FromOrgThresholds(orgSettings.DefaultThresholds)
	effectiveThresholds := devicedomain.ResolveThresholds(settings, orgThresholds)

	return map[string]interface{}{
		"id":                  settings.ID,
		"deviceImei":          settings.DeviceIMEI,
		"customName":          settings.CustomName,
		"location":            settings.Location,
		"metadata":            convertMetadataToList(settings.Metadata),
		"thresholds":          settings.Thresholds,
		"effectiveThresholds": effectiveThresholds,
		"createdAt":           settings.CreatedAt,
		"updatedAt":           settings.UpdatedAt,
	}, nil
}

// UpdateOrganizationSettings resolves the updateOrganizationSettings mutation.
func (r *Resolver) UpdateOrganizationSettings(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

	// Check if operator is a member of the organization.
	err = r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID)
	if err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	input, ok := p.Args["input"].(map[string]interface{})
	if !ok {
		return nil, r.Presenter.BadRequestError("invalid input")
	}

	// Parse organization settings input.
	settingsReq := &orgdomain.UpdateOrganizationSettingsRequest{}

	if v, ok := input["timezone"].(string); ok {
		settingsReq.Timezone = &v
	}

	if v, ok := input["dateFormat"].(string); ok {
		settingsReq.DateFormat = &v
	}

	if v, ok := input["alertCooldownMinutes"].(int); ok {
		settingsReq.AlertCooldownMinutes = &v
	}

	if thresholds, ok := input["defaultThresholds"].(map[string]interface{}); ok {
		settingsReq.DefaultThresholds = parseOrgThresholds(thresholds)
	}

	settings, err := r.OrgSettingsService.UpdateSettings(ctx, orgID, settingsReq)
	if err != nil {
		if errors.Is(err, orgdomain.ErrSettingsNotFound) {
			return nil, r.Presenter.NotFoundError("organization settings not found")
		}
		if errors.Is(err, orgdomain.ErrInvalidThreshold) {
			return nil, r.Presenter.BadRequestError("invalid threshold values: warning must be less than critical")
		}
		return nil, r.Presenter.InternalError("failed to update organization settings")
	}

	return map[string]interface{}{
		"id":                   settings.ID,
		"organizationId":       settings.OrganizationID,
		"timezone":             settings.Timezone,
		"dateFormat":           settings.DateFormat,
		"alertCooldownMinutes": settings.AlertCooldownMinutes,
		"defaultThresholds":    settings.DefaultThresholds,
		"createdAt":            settings.CreatedAt,
		"updatedAt":            settings.UpdatedAt,
	}, nil
}

// parseDeviceThresholds parses device threshold input.
func parseDeviceThresholds(input map[string]interface{}) *devicedomain.Thresholds {
	thresholds := &devicedomain.Thresholds{}

	if v, ok := input["riskWarn"].(int); ok {
		thresholds.RiskWarn = v
	}
	if v, ok := input["riskCrit"].(int); ok {
		thresholds.RiskCrit = v
	}
	if v, ok := input["thermalWarn"].(int); ok {
		thresholds.ThermalWarn = v
	}
	if v, ok := input["thermalCrit"].(int); ok {
		thresholds.ThermalCrit = v
	}
	if v, ok := input["bufferWarn"].(int); ok {
		thresholds.BufferWarn = v
	}
	if v, ok := input["bufferCrit"].(int); ok {
		thresholds.BufferCrit = v
	}

	return thresholds
}

// parseOrgThresholds parses organization threshold input.
func parseOrgThresholds(input map[string]interface{}) *orgdomain.Thresholds {
	thresholds := &orgdomain.Thresholds{}

	if v, ok := input["riskWarn"].(int); ok {
		thresholds.RiskWarn = v
	}
	if v, ok := input["riskCrit"].(int); ok {
		thresholds.RiskCrit = v
	}
	if v, ok := input["thermalWarn"].(int); ok {
		thresholds.ThermalWarn = v
	}
	if v, ok := input["thermalCrit"].(int); ok {
		thresholds.ThermalCrit = v
	}
	if v, ok := input["bufferWarn"].(int); ok {
		thresholds.BufferWarn = v
	}
	if v, ok := input["bufferCrit"].(int); ok {
		thresholds.BufferCrit = v
	}

	return thresholds
}

// TestWebhook resolves the testWebhook mutation.
func (r *Resolver) TestWebhook(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	url, ok := p.Args["url"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("url must be a string")
	}
	if url == "" {
		return nil, r.Presenter.BadRequestError("url is required")
	}

	if err := validateWebhookURL(url); err != nil {
		return nil, r.Presenter.BadRequestError("invalid webhook URL: " + err.Error())
	}

	result, err := r.WebhookClient.Test(ctx, url)
	if err != nil {
		return nil, r.Presenter.InternalError("webhook test failed - please verify the URL is correct and accessible")
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

// CreateAlertRule resolves the createAlertRule mutation.
func (r *Resolver) CreateAlertRule(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	if r.AlertSvc == nil {
		return nil, r.Presenter.InternalError("alert service not available")
	}
	in := alertapp.RuleInputFromArgs(p.Args, orgID)
	rule, err := r.AlertSvc.CreateRule(p.Context, in)
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return r.alertRuleView(p, orgID, rule.ID)
}

// UpdateAlertRule resolves the updateAlertRule mutation.
func (r *Resolver) UpdateAlertRule(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	ruleID, argOK := p.Args["ruleId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: ruleId")
	}
	if r.AlertSvc == nil {
		return nil, r.Presenter.InternalError("alert service not available")
	}
	in := alertapp.RuleInputFromArgs(p.Args, orgID)
	rule, err := r.AlertSvc.UpdateRule(p.Context, orgID, ruleID, in)
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return r.alertRuleView(p, orgID, rule.ID)
}

// DeleteAlertRule resolves the deleteAlertRule mutation.
func (r *Resolver) DeleteAlertRule(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	ruleID, argOK := p.Args["ruleId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: ruleId")
	}
	if r.AlertSvc == nil {
		return nil, r.Presenter.InternalError("alert service not available")
	}
	if err := r.AlertSvc.DeleteRule(p.Context, orgID, ruleID); err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return true, nil
}

// EvaluateAlertRule resolves the evaluateAlertRule mutation.
func (r *Resolver) EvaluateAlertRule(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	ruleID, argOK := p.Args["ruleId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: ruleId")
	}
	if r.AlertEvaluator == nil {
		return nil, r.Presenter.InternalError("alert evaluator not available")
	}
	transitioned, err := r.AlertEvaluator.EvaluateRule(p.Context, ruleID, time.Now())
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return transitioned, nil
}

func (r *Resolver) alertRuleView(p graphql.ResolveParams, orgID, ruleID string) (interface{}, error) {
	view, err := r.AlertSvc.GetRule(p.Context, orgID, ruleID)
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return map[string]interface{}{
		"id":                    view.Rule.ID,
		"orgId":                 view.Rule.OrgID,
		"name":                  view.Rule.Name,
		"metric":                string(view.Rule.Metric),
		"condition":             string(view.Rule.Condition),
		"threshold":             view.Rule.Threshold,
		"forSeconds":            view.Rule.ForSeconds,
		"notifyIntervalSeconds": view.Rule.NotifyIntervalSeconds,
		"webhookUrl":            view.Rule.WebhookURL,
		"enabled":               view.Rule.Enabled,
		"state":                 string(view.State),
		"value":                 view.Value,
		"evaluatedAt":           view.EvaluatedAt,
		"createdAt":             view.Rule.CreatedAt,
		"updatedAt":             view.Rule.UpdatedAt,
	}, nil
}

// CreateContactPoint resolves the createContactPoint mutation.
func (r *Resolver) CreateContactPoint(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	if r.ContactPointSvc == nil {
		return nil, r.Presenter.InternalError("contact point service not available")
	}
	in := contactPointInputFromArgs(p.Args, orgID)
	cp, err := r.ContactPointSvc.Create(p.Context, in)
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return contactPointView(cp), nil
}

// UpdateContactPoint resolves the updateContactPoint mutation.
func (r *Resolver) UpdateContactPoint(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	contactPointID, argOK := p.Args["contactPointId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: contactPointId")
	}
	if r.ContactPointSvc == nil {
		return nil, r.Presenter.InternalError("contact point service not available")
	}
	in := contactPointInputFromArgs(p.Args, orgID)
	cp, err := r.ContactPointSvc.Update(p.Context, orgID, contactPointID, in)
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return contactPointView(cp), nil
}

// DeleteContactPoint resolves the deleteContactPoint mutation.
func (r *Resolver) DeleteContactPoint(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	contactPointID, argOK := p.Args["contactPointId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: contactPointId")
	}
	if r.ContactPointSvc == nil {
		return nil, r.Presenter.InternalError("contact point service not available")
	}
	if err := r.ContactPointSvc.Delete(p.Context, orgID, contactPointID); err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return true, nil
}

// TestContactPoint resolves the testContactPoint mutation.
func (r *Resolver) TestContactPoint(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	contactPointID, argOK := p.Args["contactPointId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: contactPointId")
	}
	if r.ContactPointSvc == nil || r.Dispatcher == nil {
		return nil, r.Presenter.InternalError("contact point service not available")
	}
	cp, err := r.ContactPointSvc.Get(p.Context, orgID, contactPointID)
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	msg := &notificationdomain.Message{
		Subject: "Vyzorix contact point test",
		Body:    "Test notification from Vyzorix",
		Event:   "test",
	}
	if err := r.Dispatcher.SendToPoint(p.Context, cp, msg); err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return true, nil
}

func contactPointInputFromArgs(args map[string]interface{}, orgID string) *notifications.ContactPointInput {
	in := &notifications.ContactPointInput{OrgID: orgID}
	if v, ok := args["name"].(string); ok {
		in.Name = v
	}
	if v, ok := args["channel"].(string); ok {
		in.Channel = notificationdomain.ChannelType(v)
	}
	if v, ok := args["secret"].(string); ok {
		in.Secret = v
	}
	if v, ok := args["config"].(map[string]interface{}); ok {
		cfg := make(map[string]string)
		for k, val := range v {
			if str, ok := val.(string); ok {
				cfg[k] = str
			}
		}
		in.Config = cfg
	}
	if v, ok := args["enabled"].(bool); ok {
		in.Enabled = v
	}
	return in
}

func contactPointView(cp *notificationdomain.ContactPoint) map[string]interface{} {
	return map[string]interface{}{
		"id":        cp.ID,
		"orgId":     cp.OrgID,
		"name":      cp.Name,
		"channel":   string(cp.Channel),
		"secret":    cp.Secret != "",
		"config":    cp.Config,
		"enabled":   cp.Enabled,
		"createdAt": cp.CreatedAt,
		"updatedAt": cp.UpdatedAt,
	}
}

// CreateServiceAccount resolves the createServiceAccount mutation.
func (r *Resolver) CreateServiceAccount(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	name, argOK := p.Args["name"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: name")
	}
	if r.ServiceAccountSvc == nil {
		return nil, r.Presenter.InternalError("service account service not available")
	}
	sa, err := r.ServiceAccountSvc.Create(p.Context, orgID, name)
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return map[string]interface{}{
		"id":        sa.ID,
		"orgId":     sa.OrgID,
		"name":      sa.Name,
		"enabled":   sa.Enabled,
		"createdAt": sa.CreatedAt,
	}, nil
}

// DeleteServiceAccount resolves the deleteServiceAccount mutation.
func (r *Resolver) DeleteServiceAccount(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	serviceAccountID, argOK := p.Args["serviceAccountId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: serviceAccountId")
	}
	if r.ServiceAccountSvc == nil {
		return nil, r.Presenter.InternalError("service account service not available")
	}
	if err := r.ServiceAccountSvc.Delete(p.Context, orgID, serviceAccountID); err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return true, nil
}

// CreateServiceAccountToken resolves the createServiceAccountToken mutation.
func (r *Resolver) CreateServiceAccountToken(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	serviceAccountID, argOK := p.Args["serviceAccountId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: serviceAccountId")
	}
	if r.ServiceAccountSvc == nil {
		return nil, r.Presenter.InternalError("service account service not available")
	}
	in := &serviceaccount.TokenInput{
		ServiceID: serviceAccountID,
	}
	if v, ok := p.Args["name"].(string); ok {
		in.Name = v
	}
	if v, ok := p.Args["scopes"].([]interface{}); ok {
		scopes := make([]string, 0, len(v))
		for _, s := range v {
			if str, ok := s.(string); ok {
				scopes = append(scopes, str)
			}
		}
		in.Scopes = scopes
	}
	if v, ok := p.Args["expiresAt"].(string); ok && v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			in.ExpiresAt = &t
		}
	}
	token, fullKey, err := r.ServiceAccountSvc.CreateToken(p.Context, in)
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	result := tokenView(token)
	result["fullKey"] = fullKey
	return result, nil
}

// RotateServiceAccountToken resolves the rotateServiceAccountToken mutation.
func (r *Resolver) RotateServiceAccountToken(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	serviceAccountID, argOK := p.Args["serviceAccountId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: serviceAccountId")
	}
	tokenID, argOK := p.Args["tokenId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: tokenId")
	}
	if r.ServiceAccountSvc == nil {
		return nil, r.Presenter.InternalError("service account service not available")
	}
	token, fullKey, err := r.ServiceAccountSvc.RotateToken(p.Context, tokenID, &serviceaccount.TokenInput{
		ServiceID: serviceAccountID,
	})
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	result := tokenView(token)
	result["fullKey"] = fullKey
	return result, nil
}

func tokenView(token *sadomain.Token) map[string]interface{} {
	return map[string]interface{}{
		"id":        token.ID,
		"serviceId": token.ServiceID,
		"name":      token.Name,
		"keyPrefix": token.KeyPrefix,
		"scopes":    token.Scopes,
		"valid":     token.Valid,
		"expiresAt": token.ExpiresAt,
		"createdAt": token.CreatedAt,
	}
}

// CreateAnnotation resolves the createAnnotation mutation.
func (r *Resolver) CreateAnnotation(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	if r.AnnotationSvc == nil {
		return nil, r.Presenter.InternalError("annotation service not available")
	}
	in := &appannotation.AnnotationInput{OrgID: orgID}
	if v, ok := p.Args["title"].(string); ok {
		in.Title = v
	}
	if v, ok := p.Args["text"].(string); ok {
		in.Text = v
	}
	if v, ok := p.Args["tags"].([]interface{}); ok {
		tags := make([]string, 0, len(v))
		for _, t := range v {
			if str, ok := t.(string); ok {
				tags = append(tags, str)
			}
		}
		in.Tags = tags
	}
	if v, ok := p.Args["source"].(string); ok {
		in.Source = v
	}
	if v, ok := p.Args["startTime"].(string); ok && v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			in.StartTime = t
		}
	}
	if v, ok := p.Args["endTime"].(string); ok && v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			in.EndTime = &t
		}
	}
	a, err := r.AnnotationSvc.Create(p.Context, in)
	if err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return annotationView(a), nil
}

// DeleteAnnotation resolves the deleteAnnotation mutation.
func (r *Resolver) DeleteAnnotation(p graphql.ResolveParams) (interface{}, error) {
	op, ok := gqlcontext.GetOperator(p.Context)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}
	orgID, argOK := p.Args["organizationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: organizationId")
	}
	annotationID, argOK := p.Args["annotationId"].(string)
	if !argOK {
		return nil, errors.New("invalid argument: annotationId")
	}
	if r.AnnotationSvc == nil {
		return nil, r.Presenter.InternalError("annotation service not available")
	}
	if err := r.AnnotationSvc.Delete(p.Context, orgID, annotationID); err != nil {
		return nil, r.Presenter.InternalError(err.Error())
	}
	return true, nil
}

func annotationView(a *annotationdomain.Annotation) map[string]interface{} {
	return map[string]interface{}{
		"id":        a.ID,
		"orgId":     a.OrgID,
		"title":     a.Title,
		"text":      a.Text,
		"tags":      a.Tags,
		"source":    a.Source,
		"startTime": a.StartTime,
		"endTime":   a.EndTime,
		"createdAt": a.CreatedAt,
	}
}
