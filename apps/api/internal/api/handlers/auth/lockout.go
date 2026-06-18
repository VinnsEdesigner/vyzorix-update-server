package auth

import (
	"net/http"

	appauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// LockoutHandler handles account lockout endpoints.
type LockoutHandler struct {
	authService *appauth.AuthService
	lockout     *middleware.Lockout
}

// NewLockoutHandler creates a new LockoutHandler.
func NewLockoutHandler(authService *appauth.AuthService, lockout *middleware.Lockout) *LockoutHandler {
	return &LockoutHandler{
		authService: authService,
		lockout:     lockout,
	}
}

// GetLockoutStatus handles GET /v1/auth/lockout/status.
func (h *LockoutHandler) GetLockoutStatus(c *gin.Context) {
	sessionID, err := h.getSessionFromCookie(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get operator email to check lockout status
	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	locked, attemptsRemaining, retryAfter := h.lockout.GetLockoutInfo(op.Email)
	if locked {
		c.JSON(http.StatusOK, gin.H{
			"locked":             true,
			"reason":             "Too many failed attempts",
			"retry_after":        retryAfter.Seconds(),
			"attempts_remaining": attemptsRemaining,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"locked":             false,
		"attempts_remaining": attemptsRemaining,
	})
}

// UnlockAccount handles POST /v1/admin/lockout/unlock/:operator_id.
func (h *LockoutHandler) UnlockAccount(c *gin.Context) {
	sessionID, err := h.getSessionFromCookie(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Check if admin
	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if op.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	targetOperatorID := c.Param("operator_id")
	if targetOperatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operator_id required"})
		return
	}

	// Get target operator email to clear their lockout
	targetOp, err := h.authService.GetOperatorByID(c.Request.Context(), targetOperatorID)
	if err != nil || targetOp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "operator not found"})
		return
	}

	// Clear lockout using in-memory lockout (email-based)
	h.lockout.RecordSuccessfulAttempt(targetOp.Email)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Account unlocked successfully",
		"operator_id": targetOperatorID,
	})
}

// getSessionFromCookie extracts session ID from cookie.
func (h *LockoutHandler) getSessionFromCookie(c *gin.Context) (string, error) {
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		return "", err
	}

	return sessionID, nil
}
