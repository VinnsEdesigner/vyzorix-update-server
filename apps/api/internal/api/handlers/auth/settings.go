package auth

import (
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"

	"github.com/gin-gonic/gin"
)

// SettingsHandler handles /me settings endpoints.
type SettingsHandler struct {
	authService *auth.AuthService
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(authService *auth.AuthService) *SettingsHandler {
	return &SettingsHandler{authService: authService}
}

// getOperatorFromSession extracts operator ID from session.
func (h *SettingsHandler) getOperatorFromSession(c *gin.Context) (string, *operator.Operator, error) {
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		return "", nil, err
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		return "", nil, err
	}

	return op.ID, op, nil
}

// UpdateName handles PATCH /v1/auth/me.
func (h *SettingsHandler) UpdateName(c *gin.Context) {
	operatorID, _, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "an error occurred"})
		return
	}

	var req struct {
		Name *string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid JSON body"})
		return
	}

	if req.Name == nil || *req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "name is required"})
		return
	}

	op, err := h.authService.UpdateOperatorName(c.Request.Context(), operatorID, *req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              op.ID,
		"email":           op.Email,
		"name":            op.Name,
		"role":            op.Role,
		"mfa_enabled":     op.MFAEnabled,
		"email_verified":  op.EmailVerified,
		"thresholds":      op.Thresholds,
		"client":         op.ClientSettings,
	})
}

// UpdateSettings handles PATCH /v1/auth/me/settings.
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	operatorID, _, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "an error occurred"})
		return
	}

	var req auth.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid JSON body"})
		return
	}

	var op *operator.Operator

	if req.Reset {
		op, err = h.authService.ResetSettings(c.Request.Context(), operatorID)
	} else {
		op, err = h.authService.UpdateSettings(c.Request.Context(), operatorID, &req)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              op.ID,
		"email":           op.Email,
		"name":            op.Name,
		"role":            op.Role,
		"mfa_enabled":     op.MFAEnabled,
		"email_verified":  op.EmailVerified,
		"thresholds":      op.Thresholds,
		"client":         op.ClientSettings,
	})
}
