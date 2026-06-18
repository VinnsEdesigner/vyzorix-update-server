// Package handlers provides HTTP handlers for account lockout operations.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
)

// LockoutHandler handles account lockout-related HTTP requests.
type LockoutHandler struct {
	Store   *storage.Store
	Lockout *auth.AccountLockout
}

// NewLockoutHandler creates a new lockout handler.
func NewLockoutHandler(store *storage.Store, lockout *auth.AccountLockout) *LockoutHandler {
	return &LockoutHandler{
		Store:   store,
		Lockout: lockout,
	}
}

// GetLockoutStatus returns the lockout status for the current operator.
// GET /v1/auth/lockout/status.
func (h *LockoutHandler) GetLockoutStatus(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	locked, reason, err := h.Lockout.IsLocked(c.Request.Context(), operator.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check lockout status"})
		return
	}

	if locked {
		c.JSON(http.StatusOK, gin.H{
			"locked":  true,
			"reason":  reason.Reason,
			"until":   reason.Until,
			"attempts": reason.Attempts,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"locked": false,
	})
}

// UnlockAccount unlocks an operator's account (admin only).
// POST /v1/admin/lockout/unlock/:operator_id.
func (h *LockoutHandler) UnlockAccount(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Check if admin.
	if operator.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	targetOperatorID := c.Param("operator_id")
	if targetOperatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operator_id required"})
		return
	}

	err := h.Lockout.ClearLockout(c.Request.Context(), targetOperatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlock account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Account unlocked successfully",
		"operator_id": targetOperatorID,
	})
}

// LockoutStatusResponse represents the response for lockout status.
type LockoutStatusResponse struct {
	Locked   bool   `json:"locked"`
	Reason   string `json:"reason,omitempty"`
	Until    int64  `json:"until,omitempty"`
	Attempts int    `json:"attempts,omitempty"`
}

// LockoutInfoResponse represents the response when account is locked.
type LockoutInfoResponse struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	OperatorID       string `json:"operator_id,omitempty"`
	Locked           bool   `json:"locked"`
	RetryAfterSeconds int64 `json:"retry_after_seconds,omitempty"`
}
