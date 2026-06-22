package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	appauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// LockoutHandler handles account lockout endpoints.
type LockoutHandler struct {
	authService *appauth.AuthService
	lockout     *middleware.Lockout
	presenter  *response.Presenter
}

// NewLockoutHandler creates a new LockoutHandler.
func NewLockoutHandler(authService *appauth.AuthService, lockout *middleware.Lockout, presenter *response.Presenter) *LockoutHandler {
	return &LockoutHandler{
		authService: authService,
		lockout:     lockout,
		presenter:  presenter,
	}
}

// GetLockoutStatus handles GET /v1/auth/lockout/status.
func (h *LockoutHandler) GetLockoutStatus(c *gin.Context) {
	sessionID, err := h.getSessionFromCookie(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Get operator email to check lockout status
	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	if op == nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	locked, attemptsRemaining, retryAfter := h.lockout.GetLockoutInfo(op.Email)
	if locked {
		h.presenter.OK(c, gin.H{
			"locked":             true,
			"reason":             "Too many failed attempts",
			"retry_after":        retryAfter.Seconds(),
			"attempts_remaining": attemptsRemaining,
		})
		return
	}

	h.presenter.OK(c, gin.H{
		"locked":             false,
		"attempts_remaining": attemptsRemaining,
	})
}

// UnlockAccount handles POST /v1/admin/lockout/unlock/:operator_id.
func (h *LockoutHandler) UnlockAccount(c *gin.Context) {
	sessionID, err := h.getSessionFromCookie(c)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Check if admin
	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	if op.Role != "super_admin" {
		h.presenter.Forbidden(c, "")
		return
	}

	targetOperatorID := c.Param("operator_id")
	if targetOperatorID == "" {
		h.presenter.BadRequest(c, "")
		return
	}

	// Get target operator email to clear their lockout
	targetOp, err := h.authService.GetOperatorByID(c.Request.Context(), targetOperatorID)
	if err != nil || targetOp == nil {
		h.presenter.NotFound(c, "")
		return
	}

	// Clear lockout using in-memory lockout (email-based)
	h.lockout.RecordSuccessfulAttempt(targetOp.Email)

	h.presenter.AdminAction(c, op.ID, "unlock_account", "operator", targetOperatorID, nil)
	h.presenter.OK(c, gin.H{
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
