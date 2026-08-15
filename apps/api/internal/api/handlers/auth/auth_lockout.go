package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	appauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"

	"github.com/gin-gonic/gin"
)

// LockoutHandler handles account lockout endpoints.
type LockoutHandler struct {
	authService *appauth.AuthService
	lockout     *middleware.Lockout
	presenter   *response.Presenter
}

// NewLockoutHandler creates a new LockoutHandler.
func NewLockoutHandler(authService *appauth.AuthService, lockout *middleware.Lockout, presenter *response.Presenter) *LockoutHandler {
	return &LockoutHandler{
		authService: authService,
		lockout:     lockout,
		presenter:   presenter,
	}
}

// Middleware returns a Gin middleware that checks lockout status.
// 3: Exposed so other handlers can use it.
func (h *LockoutHandler) Middleware() gin.HandlerFunc {
	return middleware.LockoutMiddleware(h.lockout)
}

// GetLockoutStatus handles GET /v1/auth/lockout/status.
func (h *LockoutHandler) GetLockoutStatus(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
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
}

// UnlockAccount handles POST /v1/admin/lockout/unlock/:operator_id.
// Requires org-scoped super_admin access.
func (h *LockoutHandler) UnlockAccount(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "")
		return
	}

	// Org-scoped check.
	orgID := middleware.GetOrganizationID(c)
	if !op.IsSuperAdminIn(orgID) {
		h.presenter.Forbidden(c, "")
		return
	}

	targetOperatorID := c.Param("operator_id")
	if targetOperatorID == "" {
		h.presenter.BadRequest(c, "")
		return
	}

	// Get target operator email to clear their lockout.
	targetOp, err := h.authService.GetOperatorByID(c.Request.Context(), targetOperatorID)
	if err != nil || targetOp == nil {
		h.presenter.NotFound(c, "")
		return
	}

	// Clear lockout using in-memory lockout (email-based).
	h.lockout.RecordSuccessfulAttempt(targetOp.Email)

	h.presenter.AdminAction(c, op.ID, "unlock_account", "operator", targetOperatorID, nil)
	h.presenter.OK(c, gin.H{
		"success":     true,
		"message":     "Account unlocked successfully",
		"operator_id": targetOperatorID,
	})
}
