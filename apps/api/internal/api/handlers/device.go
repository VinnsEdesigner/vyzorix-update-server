package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"

	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
	hmac "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/crypto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
	"github.com/gin-gonic/gin"
)

// DeviceController handles device registration, status, and management.
// It validates credentials, records active device statuses, and updates SQLite.
type DeviceController struct {
	log    *slog.Logger
	store  *storage.Store
	hub    *hub.Hub
	hmac   hmac.Verifier
	config config.Config
}

// NewDeviceController creates a new DeviceController with hub integration.
func NewDeviceController(log *slog.Logger, cfg config.Config, st *storage.Store, hmac hmac.Verifier, h *hub.Hub) *DeviceController {
	return &DeviceController{log: log, config: cfg, store: st, hmac: hmac, hub: h}
}

// Register handles device registration.
// POST /v1/device/register
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
		DeviceID:          d.ID,
		Online:            online,
		LastSeen:          d.LastSeen.UnixMilli(),
		AppVersion:        d.AppVersion,
		DeviceClass:       d.DeviceClass,
		FirebaseInstallID: d.FirebaseInstallID,
		FCMToken:          d.FCMToken,
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
		"deviceId":   id,
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

// List returns registered devices with cursor-based pagination and filtering.
// GET /v1/dashboard/devices?limit=50&cursor=<lastSeenTimestamp>&online=<true|false|all>
// Query params:
//   - limit: number of results (default 50, max 100)
//   - cursor: lastSeen timestamp in ms for pagination
//   - online: filter by online status ('true', 'false', or 'all' (default))
func (s *DeviceController) List(c *gin.Context) {
	// Parse pagination parameters
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100 // max limit
			}
		}
	}

	var cursor int64
	if cur := c.Query("cursor"); cur != "" {
		if parsed, err := strconv.ParseInt(cur, 10, 64); err == nil {
			cursor = parsed
		}
	}

	// Parse online filter
	onlineFilter := c.Query("online")
	var filterOnline *bool
	if onlineFilter == "true" {
		v := true
		filterOnline = &v
	} else if onlineFilter == "false" {
		v := false
		filterOnline = &v
	}

	// Fetch devices with pagination
	devices, err := s.store.DevicesPaginated(c.Request.Context(), limit+1, cursor)
	if err != nil {
		c.JSON(500, map[string]string{"error": "list_failed", "message": err.Error()})
		return
	}

	type deviceRow struct {
		DeviceID    string `json:"deviceId"`
		AppVersion  string `json:"appVersion"`
		DeviceClass string `json:"deviceClass"`
		LastSeen    int64  `json:"lastSeen"`
		Online      bool   `json:"online"`
	}

	out := make([]deviceRow, 0, len(devices))
	for _, d := range devices {
		isOnline := s.isDeviceOnline(d.ID) || d.Online

		// Apply online filter if specified
		if filterOnline != nil && isOnline != *filterOnline {
			continue
		}

		out = append(out, deviceRow{
			DeviceID:    d.ID,
			Online:      isOnline,
			LastSeen:    d.LastSeen.UnixMilli(),
			AppVersion:  d.AppVersion,
			DeviceClass: d.DeviceClass,
		})

		// Stop if we have enough results (before checking hasMore)
		if len(out) >= limit {
			break
		}
	}

	// Determine if there are more results
	response := map[string]any{"devices": out}
	if len(out) > 0 && len(out) == limit {
		response["nextCursor"] = out[len(out)-1].LastSeen
	}

	c.JSON(200, response)
}

// isDeviceOnline checks if a device has an active WebSocket connection via the hub.
func (s *DeviceController) isDeviceOnline(deviceID string) bool {
	if s.hub == nil {
		// Fallback to database state if hub not available
		return false
	}
	return s.hub.Online(deviceID)
}

// Config returns the controller configuration.
func (s *DeviceController) Config() config.Config { return s.config }

// Store returns the data store.
func (s *DeviceController) Store() *storage.Store { return s.store }

// Hub returns the WebSocket hub for device online status.
func (s *DeviceController) Hub() *hub.Hub { return s.hub }
