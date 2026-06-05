package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix-update-server/config"
	"github.com/VinnsEdesigner/vyzorix-update-server/models"
	"github.com/VinnsEdesigner/vyzorix-update-server/security"
	"github.com/VinnsEdesigner/vyzorix-update-server/storage"
	"github.com/gin-gonic/gin"
)

// DeviceController handles device registration, status, and management.
// It validates credentials, records active device statuses, and updates SQLite.
type DeviceController struct {
	log    *slog.Logger
	config config.Config
	store  *storage.Store
	hmac   security.Verifier
}

func NewDeviceController(log *slog.Logger, cfg config.Config, st *storage.Store, hmac security.Verifier) *DeviceController {
	return &DeviceController{log: log, config: cfg, store: st, hmac: hmac}
}

// Register handles device registration.
// POST /v1/device/register
// POST /api/v1/device/register (alternate path for compatibility)
func (s *DeviceController) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, map[string]string{"error": "bad_json", "message": err.Error()})
		return
	}
	if req.DeviceID == "" || req.FirebaseInstallID == "" {
		c.JSON(400, map[string]string{"error": "missing_field", "message": "deviceId and firebaseInstallId are required"})
		return
	}

	s.log.Info("device registration", "deviceId", req.DeviceID, "firebaseInstallId", req.FirebaseInstallID)

	d, isNew, err := s.store.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, storage.ErrHijack) {
			c.JSON(409, map[string]string{"error": "device_hijack", "message": err.Error()})
			return
		}
		c.JSON(500, map[string]string{"error": "register_failed", "message": err.Error()})
		return
	}

	s.log.Info("device registered", "deviceId", d.ID, "isNew", isNew, "commandSecret", d.CommandSecret[:8]+"...")
	c.JSON(200, models.RegisterResponse{
		DeviceID:      d.ID,
		CommandSecret: d.CommandSecret,
		RegisteredAt:  d.RegisteredAt.UnixMilli(),
		ServerTime:    time.Now().UnixMilli(),
	})
}

// Status returns the current status of a device.
// GET /v1/device/:id/status
func (s *DeviceController) Status(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "device id required"})
		return
	}

	d, ok, err := s.store.Device(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, map[string]string{"error": "lookup_failed", "message": err.Error()})
		return
	}
	if !ok {
		c.JSON(404, map[string]string{"error": "unknown_device", "message": id})
		return
	}

	online := s.isDeviceOnline(id) || d.Online
	c.JSON(200, models.DeviceStatus{
		DeviceID:    d.ID,
		Online:      online,
		LastSeen:    d.LastSeen.UnixMilli(),
		AppVersion:  d.AppVersion,
		DeviceClass: d.DeviceClass,
	})
}

// UpdateFCMToken updates the FCM token for a device.
// PATCH /v1/device/:id/fcm-token
func (s *DeviceController) UpdateFCMToken(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "device id required"})
		return
	}

	var req struct {
		FCMToken string `json:"fcmToken"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, map[string]string{"error": "bad_json", "message": err.Error()})
		return
	}
	if req.FCMToken == "" {
		c.JSON(400, map[string]string{"error": "missing_field", "message": "fcmToken is required"})
		return
	}

	s.log.Info("updating fcm token", "deviceId", id)
	if err := s.store.UpdateFCM(c.Request.Context(), id, req.FCMToken); err != nil {
		c.JSON(500, map[string]string{"error": "update_failed", "message": err.Error()})
		return
	}

	c.JSON(200, map[string]any{
		"deviceId":    id,
		"serverTime": time.Now().UnixMilli(),
	})
}

// Delete removes a device from the registry.
// DELETE /v1/device/:id
func (s *DeviceController) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, map[string]string{"error": "bad_request", "message": "device id required"})
		return
	}

	s.log.Info("deleting device", "deviceId", id)
	if err := s.store.DeleteDevice(c.Request.Context(), id); err != nil {
		c.JSON(500, map[string]string{"error": "delete_failed", "message": err.Error()})
		return
	}

	c.JSON(200, map[string]any{"deviceId": id, "deleted": true})
}

// List returns all registered devices.
// GET /v1/dashboard/devices
func (s *DeviceController) List(c *gin.Context) {
	devices, err := s.store.Devices(c.Request.Context())
	if err != nil {
		c.JSON(500, map[string]string{"error": "list_failed", "message": err.Error()})
		return
	}

	type deviceRow struct {
		DeviceID    string `json:"deviceId"`
		Online      bool   `json:"online"`
		LastSeen    int64  `json:"lastSeen"`
		AppVersion  string `json:"appVersion"`
		DeviceClass string `json:"deviceClass"`
	}

	out := make([]deviceRow, 0, len(devices))
	for _, d := range devices {
		out = append(out, deviceRow{
			DeviceID:    d.ID,
			Online:      s.isDeviceOnline(d.ID) || d.Online,
			LastSeen:    d.LastSeen.UnixMilli(),
			AppVersion:  d.AppVersion,
			DeviceClass: d.DeviceClass,
		})
	}

	c.JSON(200, map[string]any{"devices": out})
}

// isDeviceOnline checks if a device has an active WebSocket connection.
func (s *DeviceController) isDeviceOnline(deviceID string) bool {
	// This would be implemented by checking the hub's active connections
	// In a full implementation, this would integrate with the hub
	return false
}

// Config returns the controller configuration.
func (s *DeviceController) Config() config.Config { return s.config }

// Store returns the data store.
func (s *DeviceController) Store() *storage.Store { return s.store }