package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	appauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/schema"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.LockoutStatusResult
	_ openapi.SuccessResult
	_ openapi.ErrorResponse
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
// @Summary      Get lockout status
// @Description  Returns the current account lockout status for the authenticated operator
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  openapi.LockoutStatusResult  "lockout status"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Router       /auth/lockout/status [get]
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
// @Summary      Unlock account
// @Description  Clears an operator's account lockout. Requires org-scoped super_admin
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        operator_id  path  string  true  "operator ID to unlock"
// @Success      200  {object}  openapi.SuccessResult  "account unlocked"
// @Failure      400  {object}  openapi.ErrorResponse  "operator_id required"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      404  {object}  openapi.ErrorResponse  "operator not found"
// @Router       /auth/admin/lockout/unlock/{operator_id} [post]
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
	h.presenter.OK(c, schema.SuccessResult{Success: true, Message: "Account unlocked successfully"})
}
