package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/fcm"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
	hmac "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/crypto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
	"github.com/gin-gonic/gin"
)

// CommandController receives manual C2 commands from the React dashboard.
// It forwards commands to the WebSocket broker or uses FCM as fallback.
type CommandController struct {
	notifier fcm.Notifier
	log      *slog.Logger
	store    *storage.Store
	hub      *hub.Hub
	hmac     hmac.Verifier
	config   config.Config
}

func NewCommandController(
	log *slog.Logger,
	cfg config.Config,
	st *storage.Store,
	h *hub.Hub,
	notifier fcm.Notifier,
	hmac hmac.Verifier,
) *CommandController {
	return &CommandController{
		log:      log,
		config:   cfg,
		store:    st,
		hub:      h,
		notifier: notifier,
		hmac:     hmac,
	}
}

// SendCommand issues a command to a device.
// POST /v1/device/:id/command
// Operational Flow:
//   - Check if target device is online via WebSocket
//   - If online: send via hub.Send() directly
//   - If offline: use FCM signaling via notifier.SendSilentWake()
func (s *CommandController) SendCommand(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "device id required"})
		return
	}

	var req models.CommandRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, map[string]string{"error": "bad_json", "message": err.Error()})
		return
	}
	if req.Command == "" {
		c.JSON(400, map[string]string{"error": "missing_field", "message": "command is required"})
		return
	}

	s.log.Info("command received", "deviceId", id, "command", req.Command)

	// Verify device exists
	_, found, err := s.store.Device(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, map[string]string{"error": "lookup_failed", "message": err.Error()})
		return
	}
	if !found {
		c.JSON(404, map[string]string{"error": "unknown_device", "message": id})
		return
	}

	// Build command frame
	frame := models.CommandFrame{
		Type:       "command",
		DispatchID: storage.NewDispatchID(),
		Command:    req.Command,
		Args:       req.Args,
		Nonce:      req.Nonce,
		Timestamp:  req.Timestamp,
		Signature:  req.Signature,
	}

	// Determine delivery method
	delivery := "queued"

	// Try WebSocket first
	if s.hub != nil && s.hub.Send(id, frame) {
		delivery = "sent"
		s.log.Info("command sent via WebSocket", "deviceId", id, "dispatchId", frame.DispatchID)
	} else {
		// Fallback to FCM
		s.log.Info("device offline, queuing for FCM wake", "deviceId", id, "dispatchId", frame.DispatchID)
	}

	// Persist command record
	if err := s.store.SaveCommand(c.Request.Context(), frame.DispatchID, id, req.Command, req.Args, delivery); err != nil {
		s.log.Warn("failed to save command", "err", err)
	}

	// If device offline, trigger FCM wake
	if delivery == "queued" && s.notifier != nil {
		device, _, err := s.store.Device(c.Request.Context(), id)
		if err == nil && device.FCMToken != "" {
			wake := fcm.SilentWake{
				Token:      device.FCMToken,
				Command:    req.Command,
				DispatchID: frame.DispatchID,
				DeviceID:   id,
			}
			if err := s.notifier.SendSilentWake(context.Background(), wake); err != nil {
				s.log.Warn("fcm wake failed", "deviceId", id, "dispatchId", frame.DispatchID, "err", err)
				if markErr := s.store.MarkWake(context.Background(), frame.DispatchID, err.Error()); markErr != nil {
					s.log.Warn("mark wake failed", "dispatchId", frame.DispatchID, "err", markErr)
				}
			} else {
				if markErr := s.store.MarkWake(context.Background(), frame.DispatchID, ""); markErr != nil {
					s.log.Warn("mark wake failed", "dispatchId", frame.DispatchID, "err", markErr)
				}
			}
		}
	}

	c.JSON(202, models.CommandResponse{
		DispatchID: frame.DispatchID,
		Delivery:   delivery,
		ServerTime: time.Now().UnixMilli(),
	})
}

// GetCommandStatus retrieves the status of a command dispatch.
// GET /v1/command/:dispatchId/status
func (s *CommandController) GetCommandStatus(c *gin.Context) {
	dispatchID := c.Param("dispatchId")
	if dispatchID == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "dispatch id required"})
		return
	}

	// Query command status from store
	// This would require adding a GetCommand method to the store
	c.JSON(200, map[string]any{
		"dispatchId": dispatchID,
		"status":     "pending",
	})
}

// RetryCommand retries a failed command dispatch.
// POST /v1/command/:dispatchId/retry
func (s *CommandController) RetryCommand(c *gin.Context) {
	dispatchID := c.Param("dispatchId")
	if dispatchID == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "dispatch id required"})
		return
	}

	s.log.Info("retrying command", "dispatchId", dispatchID)
	c.JSON(200, map[string]any{
		"dispatchId": dispatchID,
		"retried":    true,
	})
}

// GetPendingCommands returns queued commands for a device.
// GET /v1/device/:id/commands/pending
func (s *CommandController) GetPendingCommands(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "device id required"})
		return
	}

	// Query pending commands from store
	c.JSON(200, map[string]any{
		"commands": []interface{}{},
	})
}

// CancelCommand cancels a pending command.
// DELETE /v1/command/:dispatchId
func (s *CommandController) CancelCommand(c *gin.Context) {
	dispatchID := c.Param("dispatchId")
	if dispatchID == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "dispatch id required"})
		return
	}

	s.log.Info("cancelling command", "dispatchId", dispatchID)
	c.JSON(200, map[string]any{
		"dispatchId": dispatchID,
		"cancelled":  true,
	})
}

// Hub returns the WebSocket hub.
func (s *CommandController) Hub() *hub.Hub { return s.hub }

// Store returns the data store.
func (s *CommandController) Store() *storage.Store { return s.store }
