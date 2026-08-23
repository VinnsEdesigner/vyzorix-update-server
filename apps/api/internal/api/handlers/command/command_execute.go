package command

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	cmdSvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"

	domainerrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	cryptohmac "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.CommandRequest
	_ openapi.CommandDispatchResult
	_ openapi.CommandStatus
	_ openapi.CommandRetryResult
	_ openapi.CommandCancelResult
	_ openapi.CommandPendingResult
	_ openapi.CommandResponse
	_ openapi.ErrorResponse
)

// AuditLogger is the audit interface the handler depends on. It mirrors the.
// subset of *audit.Logger the handler uses, allowing a no-op stand-in for.
// environments without an audit store (and easy mocking in tests). *audit.Logger.
// and *audit.NoOpLogger both satisfy it.
type AuditLogger interface {
	CommandExecuted(ctx context.Context, e audit.CommandExecutedEvent)
}

// ExecuteHandler handles device command execution.
type ExecuteHandler struct {
	commandService *cmdSvc.Service
	deviceService  *device.Service
	hub            *hub.Hub
	fcmNotifier    fcm.Notifier
	commandSigner  *cryptohmac.CommandSigner
	authorizer     *cmdSvc.Authorizer
	audit          AuditLogger
	log            *slog.Logger
}

// NewExecuteHandler creates a new ExecuteHandler. authorizer applies the shared
// command risk gate (MFA + confirmation); a nil confirmation backing means
// confirmation-gated commands are always blocked with 425.
func NewExecuteHandler(commandService *cmdSvc.Service, deviceService *device.Service, hub *hub.Hub, fcmNotifier fcm.Notifier, authorizer *cmdSvc.Authorizer, auditLogger AuditLogger) *ExecuteHandler {
	return &ExecuteHandler{
		commandService: commandService,
		deviceService:  deviceService,
		hub:            hub,
		fcmNotifier:    fcmNotifier,
		commandSigner:  cryptohmac.NewCommandSigner(),
		authorizer:     authorizer,
		audit:          auditLogger,
		log:            slog.Default(),
	}
}

// verifyDeviceInOrganization verifies the device belongs to the organization.
func (h *ExecuteHandler) verifyDeviceInOrganization(ctx context.Context, deviceID, orgID string) error {
	_, err := h.deviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
	return err
}

// commandRequest is the JSON payload for POST /v1/device/:imei/command.
//
//nolint:govet // fieldalignment: reordered for best packing
type commandRequest struct {
	Command           string                 `json:"command"`
	ConfirmationToken string                 `json:"confirmation_token,omitempty"`
	DispatchID        string                 `json:"dispatch_id,omitempty"`
	Nonce             string                 `json:"nonce"`
	Signature         string                 `json:"signature,omitempty"`
	Timestamp         int64                  `json:"timestamp"`
	Args              map[string]interface{} `json:"args,omitempty"`
}

// Handle handles POST /v1/device/:imei/command.
// @Summary      Execute command
// @Description  Dispatches a command to a device. Risk-gated commands may require a confirmation token
// @Tags         commands
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei  path  string  true  "device IMEI"
// @Param        body  body  openapi.CommandRequest  true  "command request"
// @Success      202  {object}  openapi.CommandDispatchResult  "dispatch result"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      404  {object}  openapi.ErrorResponse  "device not found"
// @Failure      425  {object}  openapi.ErrorResponse  "confirmation required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/{imei}/command [post]
func (h *ExecuteHandler) Handle(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "device imei required"))
		return
	}

	var req commandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "invalid request body"))
		return
	}

	if verr := h.validateCommandRequest(imei, req.Command, req.Nonce); verr != nil {
		// Record the structured validation error and let the error middleware.
		// render a 400 with field-level details + trace id + docs link.
		_ = c.Error(verr)
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "organization context required"))
		return
	}

	if err := h.verifyDeviceInOrganization(c.Request.Context(), imei, orgID); err != nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeResourceNotFound, "device not found"))
		return
	}

	// Risk gate: classify the command and authorize against the actor context.
	// This runs after validation/org checks so a bad request never reaches the.
	// risk evaluator, and before dispatch so dangerous commands can be blocked.
	if !h.authorizeCommand(c, req, imei) {
		return
	}

	cmdResp, frame, err := h.sendCommandAndBuildFrame(c, imei, req)
	if err != nil {
		return
	}

	delivery := h.deliverCommand(c, imei, frame, cmdResp)

	// OldValue/NewValue record the command state transition for change-tracking.
	// compliance: a newly-created command starts at "pending" and may transition.
	// to "delivered" if the device is online. The actor type/email are sourced.
	// from the authenticated operator.
	op := middleware.GetOperatorFromContext(c)
	actorType, actorEmail := "operator", ""
	if op != nil {
		actorEmail = op.Email
	}
	oldState := ""
	if delivery == "sent" {
		oldState = "pending"
	}

	h.audit.CommandExecuted(c.Request.Context(), audit.CommandExecutedEvent{
		OperatorID: operatorIDFromContext(c),
		DeviceID:   imei,
		Command:    req.Command,
		DispatchID: cmdResp.DispatchID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		TraceID:    middleware.GetTraceID(c),
		RiskTier:   string(command.LookupRiskProfile(req.Command).Tier),
		Result:     audit.ResultSuccess,
		ActorType:  actorType,
		ActorEmail: actorEmail,
		OldValue:   oldState,
		NewValue:   delivery,
	})

	c.JSON(http.StatusAccepted, gin.H{
		"status":        delivery,
		"device_online": delivery == "sent",
		"dispatchId":    cmdResp.DispatchID,
		"command_id":    cmdResp.CommandID,
		"serverTime":    time.Now().Unix(),
	})
}

// authorizeCommand evaluates the command's risk profile against the actor and.
// blocks dispatch when a confirmation is required but not satisfied. A.
// confirmation is satisfied by presenting a valid, unconsumed confirmation.
// token (issued by the confirm endpoint). It writes the response and an audit.
// "blocked" entry on denial, returning false if the handler should abort.
func (h *ExecuteHandler) authorizeCommand(c *gin.Context, req commandRequest, imei string) bool {
	actor := command.ActorContext{
		OperatorID: operatorIDFromContext(c),
		OrgID:      middleware.GetOrganizationID(c),
	}
	if op := middleware.GetOperatorFromContext(c); op != nil {
		actor.IsSuperAdmin = op.IsSuperAdmin()
	}
	// MFA-verified is derived from the authenticated session. Critical-tier.
	// commands require it; without an MFA-verified session they are gated even.
	// when a confirmation token is presented.
	if sess := middleware.GetSession(c); sess != nil && sess.MFAVerifiedAt != nil {
		actor.MFAVerified = true
	}

	outcome := h.authorizer.Authorize(c.Request.Context(), actor, req.Command, imei, req.ConfirmationToken)
	if outcome.Allowed {
		return true
	}

	h.audit.CommandExecuted(c.Request.Context(), audit.CommandExecutedEvent{
		OperatorID: actor.OperatorID,
		DeviceID:   imei,
		Command:    req.Command,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		TraceID:    middleware.GetTraceID(c),
		RiskTier:   string(outcome.Tier),
		Result:     audit.ResultBlocked,
		Reason:     outcome.Reason,
	})
	if outcome.NeedsConfirmation {
		responses.RespondStructured(c, http.StatusTooEarly, outcome.Message)
	} else {
		responses.RespondStructured(c, http.StatusForbidden, outcome.Message)
	}
	return false
}

// operatorIDFromContext returns the authenticated operator's ID, or "" for.
// system-originated requests.
func operatorIDFromContext(c *gin.Context) string {
	if op := middleware.GetOperatorFromContext(c); op != nil {
		return op.ID
	}
	return ""
}

// validateCommandRequest validates the command request parameters and returns.
// a structured *domainerrors.ValidationError with field-level details, or nil.
// when the request is valid. The returned error is consumed by the error.
// middleware, which renders a structured 400 with the details.
func (h *ExecuteHandler) validateCommandRequest(imei, command, nonce string) *domainerrors.ValidationError {
	var details []domainerrors.ValidationDetail
	add := func(field, msg string) {
		details = append(details, domainerrors.NewValidationDetail(field, msg))
	}

	if err := dto.ValidateDeviceID(imei); err != nil {
		add("deviceId", err.Error())
	}
	if err := dto.ValidateCommand(command); err != nil {
		add("command", err.Error())
	}
	if nonce != "" {
		if err := dto.ValidateNonce(nonce); err != nil {
			add("nonce", err.Error())
		}
	}

	if len(details) == 0 {
		return nil
	}
	return domainerrors.NewValidationError(details)
}

// sendCommandAndBuildFrame sends the command via service and builds a signed.
// command frame. The frame is HMAC-signed with the device's command secret.
// (Domain B: server→device command signing) so the Android device can verify.
// authenticity. Client-provided nonce/signature/timestamp are intentionally.
// discarded — the server is the signing authority, never the web client.
func (h *ExecuteHandler) sendCommandAndBuildFrame(c *gin.Context, imei string, req commandRequest) (*dto.SendCommandResponse, command.CommandFrame, error) {
	argsJSON, err := json.Marshal(req.Args)
	if err != nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeInternalServerError, "failed to marshal args"))
		return nil, command.CommandFrame{}, err
	}

	cmdReq := &dto.SendCommandRequest{
		DeviceID:   imei,
		Command:    req.Command,
		Args:       req.Args,
		DispatchID: req.DispatchID,
	}

	cmdResp, err := h.commandService.SendCommand(c.Request.Context(), cmdReq)
	if err != nil {
		h.log.Error("failed to send command", "error", err, "deviceId", imei)
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeInternalServerError, "failed to send command"))
		return nil, command.CommandFrame{}, err
	}

	// Build the frame with the server-generated dispatch ID.
	frame := command.CommandFrame{
		Type:       req.Command,
		Command:    req.Command,
		DispatchID: cmdResp.DispatchID,
		Args:       argsJSON,
		Timestamp:  h.commandSigner.GenerateTimestampMs(),
	}

	// Sign the frame with the device's command secret so the device can.
	// verify the command originated from the server (Domain B).
	if err := h.signCommandFrame(c.Request.Context(), imei, &frame); err != nil {
		h.log.Warn("failed to sign command frame; aborting dispatch", "error", err, "deviceId", imei)
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeInternalServerError, "failed to sign command"))
		return nil, command.CommandFrame{}, err
	}

	return cmdResp, frame, nil
}

// signCommandFrame retrieves the device's command secret and signs the frame.
// in place (sets Nonce + Signature). The device's CommandSecretHash is a.
// deterministic derivation of the plaintext secret (SHA-256), so both the.
// server and the device can compute the same HMAC key without the server.
// storing the plaintext.
func (h *ExecuteHandler) signCommandFrame(ctx context.Context, imei string, frame *command.CommandFrame) error {
	dev, err := h.deviceService.GetDevice(ctx, imei)
	if err != nil {
		return err
	}
	if dev.CommandSecretHash == "" {
		return errors.New("device has no command secret — re-registration required")
	}

	nonce, sig, err := h.commandSigner.SignCommand(frame, imei, dev.CommandSecretHash)
	if err != nil {
		return err
	}
	frame.Nonce = nonce
	frame.Signature = sig
	return nil
}

// deliverCommand delivers the command via WebSocket or FCM.
func (h *ExecuteHandler) deliverCommand(c *gin.Context, imei string, frame command.CommandFrame, cmdResp *dto.SendCommandResponse) string {
	delivery := "queued"

	if h.hub != nil && h.hub.Online(imei) {
		if sent := h.hub.Send(imei, frame); sent {
			delivery = "sent"
			if err := h.commandService.MarkDelivered(c.Request.Context(), cmdResp.CommandID); err != nil {
				h.log.Warn("failed to mark command delivered", "error", err)
			}
		}
	}

	if delivery == "queued" && h.fcmNotifier != nil {
		h.tryFCMWake(c, imei, cmdResp, frame.Command, &delivery)
	}

	return delivery
}

// tryFCMWake attempts to wake the device via FCM.
func (h *ExecuteHandler) tryFCMWake(c *gin.Context, imei string, cmdResp *dto.SendCommandResponse, command string, delivery *string) {
	device, err := h.deviceService.GetDevice(c.Request.Context(), imei)
	if err == nil && device.FCMToken != "" {
		wake := fcm.SilentWake{
			Token:      device.FCMToken,
			Command:    command,
			DispatchID: cmdResp.DispatchID,
			DeviceID:   imei,
		}
		if err := h.fcmNotifier.SendSilentWake(c.Request.Context(), wake); err != nil {
			h.log.Warn("fcm wake failed", "deviceId", imei, "err", err)
		} else {
			*delivery = "queued_fcm"
		}
	}
}

// GetStatus handles GET /v1/command/:dispatchId/status.
// @Summary      Get command status
// @Description  Returns the dispatch and command status for a command
// @Tags         commands
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        dispatchId  path  string  true  "dispatch ID"
// @Success      200  {object}  openapi.CommandStatus  "command status"
// @Failure      400  {object}  openapi.ErrorResponse  "dispatch id required"
// @Failure      404  {object}  openapi.ErrorResponse  "command not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /command/{dispatchId}/status [get]
func (h *ExecuteHandler) GetStatus(c *gin.Context) {
	dispatchID := c.Param("dispatchId")
	if dispatchID == "" {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "dispatch id required"))
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "organization context required"))
		return
	}

	cmdStatus, err := h.commandService.GetCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		if errors.Is(err, application.ErrCommandNotFound) {
			_ = c.Error(domainerrors.NewServerError(domainerrors.CodeResourceNotFound, "command not found"))
			return
		}
		h.log.Error("failed to get command status", "error", err, "dispatchId", dispatchID)
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeInternalServerError, "failed to get command status"))

		return
	}

	// Verify the device belongs to this organization.
	if err := h.verifyDeviceInOrganization(c.Request.Context(), cmdStatus.DeviceID, orgID); err != nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeResourceNotFound, "command not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dispatchId": dispatchID,
		"command_id": cmdStatus.CommandID,
		"device_id":  cmdStatus.DeviceID,
		"command":    cmdStatus.Command,
		"status":     cmdStatus.Status,
		"serverTime": time.Now().Unix(),
	})
}

// Retry handles POST /v1/command/:dispatchId/retry.
// @Summary      Retry command
// @Description  Retries a failed command by dispatch ID, issuing a new dispatch
// @Tags         commands
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        dispatchId  path  string  true  "dispatch ID"
// @Success      200  {object}  openapi.CommandRetryResult  "retry result"
// @Failure      400  {object}  openapi.ErrorResponse  "dispatch id required"
// @Failure      404  {object}  openapi.ErrorResponse  "command not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /command/{dispatchId}/retry [post]
func (h *ExecuteHandler) Retry(c *gin.Context) {
	dispatchID := c.Param("dispatchId")
	if dispatchID == "" {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "dispatch id required"))
		return
	}

	// Get operator from context.
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeAuthTokenInvalid, "authentication required"))
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "organization context required"))
		return
	}

	// Get command by dispatchId to find the device.
	cmd, err := h.commandService.GetCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeResourceNotFound, "command not found"))
		return
	}

	// Verify the device belongs to this organization.
	if err = h.verifyDeviceInOrganization(c.Request.Context(), cmd.DeviceID, orgID); err != nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeResourceNotFound, "command not found"))
		return
	}

	newCmd, err := h.commandService.RetryCommand(c.Request.Context(), dispatchID)
	if err != nil {
		h.log.Error("failed to retry command", "error", err, "dispatchId", dispatchID)
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeInternalServerError, "failed to retry command"))

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dispatchId": newCmd.DispatchID,
		"command_id": newCmd.CommandID,
		"retried":    true,
		"serverTime": time.Now().Unix(),
	})
}

// GetPending handles GET /v1/device/:imei/commands/pending.
// @Summary      List pending commands
// @Description  Returns commands pending delivery for a device
// @Tags         commands
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei  path  string  true  "device IMEI"
// @Success      200  {object}  openapi.CommandPendingResult  "pending commands"
// @Failure      400  {object}  openapi.ErrorResponse  "device imei required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      404  {object}  openapi.ErrorResponse  "device not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/{imei}/commands/pending [get]
func (h *ExecuteHandler) GetPending(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "device imei required"))
		return
	}

	// Get operator from context.
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeAuthTokenInvalid, "authentication required"))
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "organization context required"))
		return
	}

	// Verify the device belongs to this organization.
	if err := h.verifyDeviceInOrganization(c.Request.Context(), imei, orgID); err != nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeResourceNotFound, "device not found"))
		return
	}

	pendingCmds, err := h.commandService.GetPendingCommands(c.Request.Context(), imei)
	if err != nil {
		h.log.Error("failed to get pending commands", "error", err, "deviceId", imei)
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeInternalServerError, "failed to get pending commands"))

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"commands": pendingCmds,
	})
}

// Cancel handles DELETE /v1/command/:dispatchId.
// @Summary      Cancel command
// @Description  Cancels a pending or in-flight command by dispatch ID
// @Tags         commands
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        dispatchId  path  string  true  "dispatch ID"
// @Success      200  {object}  openapi.CommandCancelResult  "cancel result"
// @Failure      400  {object}  openapi.ErrorResponse  "dispatch id required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      404  {object}  openapi.ErrorResponse  "command not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /command/{dispatchId} [delete]
func (h *ExecuteHandler) Cancel(c *gin.Context) {
	dispatchID := c.Param("dispatchId")
	if dispatchID == "" {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "dispatch id required"))
		return
	}

	// Get operator from context.
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeAuthTokenInvalid, "authentication required"))
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeValidationFailed, "organization context required"))
		return
	}

	// Get command by dispatchId to find the device.
	cmd, err := h.commandService.GetCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeResourceNotFound, "command not found"))
		return
	}

	// Verify the device belongs to this organization.
	if err = h.verifyDeviceInOrganization(c.Request.Context(), cmd.DeviceID, orgID); err != nil {
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeResourceNotFound, "command not found"))
		return
	}

	err = h.commandService.CancelCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		h.log.Error("failed to cancel command", "error", err, "dispatchId", dispatchID)
		_ = c.Error(domainerrors.NewServerError(domainerrors.CodeInternalServerError, "failed to cancel command"))

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dispatchId": dispatchID,
		"cancelled":  true,
		"serverTime": time.Now().Unix(),
	})
}
