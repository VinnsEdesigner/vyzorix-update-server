package command

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
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

// verifyDeviceOwnership verifies the device belongs to the operator (DOA check).
// Returns unauthorized if the device doesn't exist OR doesn't belong to the operator.
// This prevents enumeration attacks by returning the same error for both cases.
func (h *ExecuteHandler) verifyDeviceOwnership(ctx context.Context, deviceID, operatorID string) error {
	_, err := h.deviceService.GetDeviceByOperator(ctx, deviceID, operatorID)
	return err
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

	if req.Command == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "command is required"})
		return
	}

	// Get operator from context (set by cookie auth middleware)
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication required"})
		return
	}

	// Get organization ID from context
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	// Verify the device belongs to this organization
	if err := h.verifyDeviceInOrganization(c.Request.Context(), imei, orgID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "device not found"})
		return
	}

	// Marshal args
	argsJSON, err := json.Marshal(req.Args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to marshal args"})
		return
	}

	// Build command frame for WebSocket
	frame := command.CommandFrame{
		Type:       req.Command,
		Command:    req.Command,
		DispatchID: req.DispatchID,
		Args:       argsJSON,
		Timestamp:  req.Timestamp,
		Nonce:      req.Nonce,
		Signature:  req.Signature,
	}

	// Use command service for proper command creation and idempotency
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

		return
	}

	// Update frame dispatch ID with the one from service (for idempotency)
	frame.DispatchID = cmdResp.DispatchID

	// Check if device is online via WebSocket and send
	delivery := "queued"

	if h.hub != nil && h.hub.Online(imei) {
		if sent := h.hub.Send(imei, frame); sent {
			delivery = "sent"
			// Mark as delivered
			if err := h.commandService.MarkDelivered(c.Request.Context(), cmdResp.CommandID); err != nil {
				h.log.Warn("failed to mark command delivered", "error", err)
			}
		}
	}

	// If not sent via WebSocket, try FCM wake for offline devices
	if delivery == "queued" && h.fcmNotifier != nil {
		device, err := h.deviceService.GetDevice(c.Request.Context(), imei)
		if err == nil && device.FCMToken != "" {
			wake := fcm.SilentWake{
				Token:      device.FCMToken,
				Command:    req.Command,
				DispatchID: cmdResp.DispatchID,
				DeviceID:   imei,
			}
			if err := h.fcmNotifier.SendSilentWake(c.Request.Context(), wake); err != nil {
				h.log.Warn("fcm wake failed", "deviceId", imei, "err", err)
			} else {
				delivery = "queued_fcm"
			}
		}
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":        delivery,
		"device_online": delivery == "sent",
		"dispatchId":    cmdResp.DispatchID,
		"command_id":    cmdResp.CommandID,
		"serverTime":    time.Now().Unix(),
	})
}

// GetStatus handles GET /v1/command/:dispatchId/status.
func (h *ExecuteHandler) GetStatus(c *gin.Context) {
	dispatchID := c.Param("dispatchId")
	if dispatchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "dispatch id required"})
		return
	}

	// Get organization ID from context
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	cmdStatus, err := h.commandService.GetCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		h.log.Error("failed to get command status", "error", err, "dispatchId", dispatchID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to get command status"})

		return
	}

	// Verify the device belongs to this organization
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

	// Get operator from context
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication required"})
		return
	}

	// Get organization ID from context
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	// Get command by dispatchId to find the device
	cmd, err := h.commandService.GetCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "command not found"})
		return
	}

	// Verify the device belongs to this organization
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

	// Get operator from context
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication required"})
		return
	}

	// Get organization ID from context
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	// Verify the device belongs to this organization
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

	// Get operator from context
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "authentication required"})
		return
	}

	// Get organization ID from context
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	// Get command by dispatchId to find the device
	cmd, err := h.commandService.GetCommandByDispatchID(c.Request.Context(), dispatchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "command not found"})
		return
	}

	// Verify the device belongs to this organization
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
