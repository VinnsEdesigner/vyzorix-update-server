package command

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	cmdSvc "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
)

// ExecuteHandler handles device command execution.
type ExecuteHandler struct {
	commandService *cmdSvc.Service
	deviceService  *device.Service
	hub            *hub.Hub
	fcmNotifier    fcm.Notifier
	log            *slog.Logger
}

// NewExecuteHandler creates a new ExecuteHandler.
func NewExecuteHandler(commandService *cmdSvc.Service, deviceService *device.Service, hub *hub.Hub, fcmNotifier fcm.Notifier) *ExecuteHandler {
	return &ExecuteHandler{
		commandService: commandService,
		deviceService:  deviceService,
		hub:            hub,
		fcmNotifier:    fcmNotifier,
		log:            slog.Default(),
	}
}

// verifyDeviceInOrganization verifies the device belongs to the organization.
func (h *ExecuteHandler) verifyDeviceInOrganization(ctx context.Context, deviceID, orgID string) error {
	_, err := h.deviceService.GetDeviceDetailByOrganization(ctx, deviceID, orgID)
	return err
}

// Handle handles POST /v1/device/:imei/command.
func (h *ExecuteHandler) Handle(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "device imei required"})
		return
	}

	var req struct {
		Args       map[string]interface{} `json:"args,omitempty"`
		Command    string                 `json:"command"`
		Nonce      string                 `json:"nonce"`
		Signature  string                 `json:"signature,omitempty"`
		DispatchID string                 `json:"dispatch_id,omitempty"`
		Timestamp  int64                  `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid request body"})
		return
	}

	if err := h.validateCommandRequest(imei, req.Command, req.Nonce); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	if err := h.verifyDeviceInOrganization(c.Request.Context(), imei, orgID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "device not found"})
		return
	}

	cmdResp, frame, err := h.sendCommandAndBuildFrame(c, imei, req)
	if err != nil {
		return
	}

	delivery := h.deliverCommand(c, imei, frame, cmdResp)

	c.JSON(http.StatusAccepted, gin.H{
		"status":        delivery,
		"device_online": delivery == "sent",
		"dispatchId":    cmdResp.DispatchID,
		"command_id":    cmdResp.CommandID,
		"serverTime":    time.Now().Unix(),
	})
}

// validateCommandRequest validates the command request parameters.
func (h *ExecuteHandler) validateCommandRequest(imei, command, nonce string) error {
	if err := dto.ValidateCommand(command); err != nil {
		return err
	}
	if err := dto.ValidateDeviceID(imei); err != nil {
		return err
	}
	if nonce != "" {
		if err := dto.ValidateNonce(nonce); err != nil {
			return err
		}
	}
	if command == "" {
		return errors.New("command is required")
	}
	return nil
}

// sendCommandAndBuildFrame sends the command via service and builds the command frame.
func (h *ExecuteHandler) sendCommandAndBuildFrame(c *gin.Context, imei string, req struct {
	Args       map[string]interface{} `json:"args,omitempty"`
	Command    string                 `json:"command"`
	Nonce      string                 `json:"nonce"`
	Signature  string                 `json:"signature,omitempty"`
	DispatchID string                 `json:"dispatch_id,omitempty"`
	Timestamp  int64                  `json:"timestamp"`
}) (*dto.SendCommandResponse, command.CommandFrame, error) {
	argsJSON, err := json.Marshal(req.Args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to marshal args"})
		return nil, command.CommandFrame{}, err
	}

	frame := command.CommandFrame{
		Type:       req.Command,
		Command:    req.Command,
		DispatchID: req.DispatchID,
		Args:       argsJSON,
		Timestamp:  req.Timestamp,
		Nonce:      req.Nonce,
		Signature:  req.Signature,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to send command"})
		return nil, command.CommandFrame{}, err
	}

	frame.DispatchID = cmdResp.DispatchID
	return cmdResp, frame, nil
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
func (h *ExecuteHandler) GetStatus(c *gin.Context) {
	dispatchID := c.Param("dispatchId")
	if dispatchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "dispatch id required"})
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	cmdStatus, err := h.commandService.GetCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		if errors.Is(err, application.ErrCommandNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "command not found"})
			return
		}
		h.log.Error("failed to get command status", "error", err, "dispatchId", dispatchID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to get command status"})

		return
	}

	// Verify the device belongs to this organization.
	if err := h.verifyDeviceInOrganization(c.Request.Context(), cmdStatus.DeviceID, orgID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "command not found"})
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
func (h *ExecuteHandler) Retry(c *gin.Context) {
	dispatchID := c.Param("dispatchId")
	if dispatchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "dispatch id required"})
		return
	}

	// Get operator from context.
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication required"})
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	// Get command by dispatchId to find the device.
	cmd, err := h.commandService.GetCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "command not found"})
		return
	}

	// Verify the device belongs to this organization.
	if err = h.verifyDeviceInOrganization(c.Request.Context(), cmd.DeviceID, orgID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "command not found"})
		return
	}

	newCmd, err := h.commandService.RetryCommand(c.Request.Context(), dispatchID)
	if err != nil {
		h.log.Error("failed to retry command", "error", err, "dispatchId", dispatchID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to retry command"})

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
func (h *ExecuteHandler) GetPending(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "device imei required"})
		return
	}

	// Get operator from context.
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication required"})
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	// Verify the device belongs to this organization.
	if err := h.verifyDeviceInOrganization(c.Request.Context(), imei, orgID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "device not found"})
		return
	}

	pendingCmds, err := h.commandService.GetPendingCommands(c.Request.Context(), imei)
	if err != nil {
		h.log.Error("failed to get pending commands", "error", err, "deviceId", imei)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to get pending commands"})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"commands": pendingCmds,
	})
}

// Cancel handles DELETE /v1/command/:dispatchId.
func (h *ExecuteHandler) Cancel(c *gin.Context) {
	dispatchID := c.Param("dispatchId")
	if dispatchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "dispatch id required"})
		return
	}

	// Get operator from context.
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication required"})
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	// Get command by dispatchId to find the device.
	cmd, err := h.commandService.GetCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "command not found"})
		return
	}

	// Verify the device belongs to this organization.
	if err = h.verifyDeviceInOrganization(c.Request.Context(), cmd.DeviceID, orgID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "command not found"})
		return
	}

	err = h.commandService.CancelCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		h.log.Error("failed to cancel command", "error", err, "dispatchId", dispatchID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to cancel command"})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dispatchId": dispatchID,
		"cancelled":  true,
		"serverTime": time.Now().Unix(),
	})
}
