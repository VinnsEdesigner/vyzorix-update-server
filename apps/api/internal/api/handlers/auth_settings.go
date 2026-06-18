package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

// UpdateName updates the display name for the authenticated operator.
// Email and role are server-controlled and cannot be changed via this endpoint.
func (ac *AuthController) UpdateName(c *gin.Context) {
	op := GetOperatorFromContext(c)
	if op == nil {
		c.JSON(401, models.ErrorResponse{Error: "unauthorized", Message: "authentication required"})
		return
	}

	var req models.UpdateNameRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}
	if req.Name == nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "name is required"})
		return
	}
	name := strings.TrimSpace(*req.Name)
	if name == "" {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "name cannot be empty"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := ac.store.UpdateOperatorName(ctx, op.ID, name); err != nil {
		ac.log.Warn("updateName: failed", "operatorID", op.ID, "err", err)
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "update failed"})
		return
	}

	// Re-fetch to return the updated operator.
	updated, err := ac.store.GetOperatorByEmail(ctx, op.Email)
	if err != nil || updated == nil {
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "update succeeded but fetch failed"})
		return
	}

	ac.log.Info("updateName: success", "operatorID", op.ID, "name", name)
	c.JSON(200, updated.ToResponse())
}

// UpdateSettings updates operator settings (name, thresholds, and client preferences).
// PATCH /v1/auth/me/settings.
func (ac *AuthController) UpdateSettings(c *gin.Context) {
	op := GetOperatorFromContext(c)
	if op == nil {
		c.JSON(401, models.ErrorResponse{Error: "unauthorized", Message: "authentication required"})
		return
	}

	var req models.UpdateSettingsRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Reset all settings to defaults.
	if req.Reset {
		if err := ac.store.ResetOperatorSettings(ctx, op.ID); err != nil {
			ac.log.Warn("updateSettings: reset failed", "operatorID", op.ID, "err", err)
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "reset failed"})
			return
		}
		ac.log.Info("updateSettings: reset to defaults", "operatorID", op.ID)
		updated, err := ac.store.GetOperatorByID(ctx, op.ID)
		if err != nil || updated == nil {
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "reset succeeded but fetch failed"})
			return
		}
		c.JSON(200, updated.ToResponse())
		return
	}

	// Update name if provided.
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "name cannot be empty"})
			return
		}
		if err := ac.store.UpdateOperatorName(ctx, op.ID, name); err != nil {
			ac.log.Warn("updateSettings: name update failed", "operatorID", op.ID, "err", err)
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "update failed"})
			return
		}
	}

	// Update thresholds if provided.
	if req.Thresholds != nil {
		if err := ac.store.UpdateOperatorThresholds(ctx, op.ID, *req.Thresholds); err != nil {
			ac.log.Warn("updateSettings: thresholds update failed", "operatorID", op.ID, "err", err)
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "update failed"})
			return
		}
	}

	// Update client settings if provided.
	if req.Client != nil {
		if err := ac.store.UpdateOperatorClientSettings(ctx, op.ID, *req.Client); err != nil {
			ac.log.Warn("updateSettings: client settings update failed", "operatorID", op.ID, "err", err)
			c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "update failed"})
			return
		}
	}

	// Re-fetch to return the updated operator.
	updated, err := ac.store.GetOperatorByID(ctx, op.ID)
	if err != nil || updated == nil {
		c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "update succeeded but fetch failed"})
		return
	}

	ac.log.Info("updateSettings: success", "operatorID", op.ID)
	c.JSON(200, updated.ToResponse())
}