package updates

import (
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.UpdatePushHistoryListResult
	_ openapi.UpdatePushDetailResult
	_ openapi.UpdateCancelPushResult
	_ openapi.ErrorResponse
)

// UpdatesHistoryHandler handles history-related HTTP requests.
type UpdatesHistoryHandler struct {
	service     *updates.Service
	auditLogger *audit.Logger
}

// NewUpdatesHistoryHandler creates a new UpdatesHistoryHandler.
func NewUpdatesHistoryHandler(service *updates.Service, auditLogger *audit.Logger) *UpdatesHistoryHandler {
	return &UpdatesHistoryHandler{
		service:     service,
		auditLogger: auditLogger,
	}
}

// GetHistory handles GET /v1/updates/history.
// @Summary      List update push history
// @Description  Returns paginated push update history
// @Tags         updates
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        page   query int    false  "page number (default 1)"
// @Param        limit  query int    false  "page size (default 20)"
// @Success      200  {object}  openapi.UpdatePushHistoryListResult  "push history"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /updates/history [get]
func (h *UpdatesHistoryHandler) GetHistory(c *gin.Context) {
	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	result, err := h.service.GetHistory(c.Request.Context(), status, page, limit, orgID)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to get history"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetPushDetail handles GET /v1/updates/history/:pushId.
// @Summary      Get push detail
// @Description  Returns details for a specific push update including per-device status
// @Tags         updates
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        pushId  path  string  true  "push ID"
// @Success      200  {object}  openapi.UpdatePushDetailResult  "push detail"
// @Failure      400  {object}  openapi.ErrorResponse  "pushId required"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      404  {object}  openapi.ErrorResponse  "push not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /updates/history/{pushId} [get]
func (h *UpdatesHistoryHandler) GetPushDetail(c *gin.Context) {
	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	pushID := c.Param("pushId")
	if pushID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Push ID is required"))
		return
	}

	detail, err := h.service.GetPushDetail(c.Request.Context(), pushID, orgID)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to get push detail"))
		return
	}
	c.JSON(http.StatusOK, detail)
}

// CancelPush handles POST /v1/updates/history/:pushId/cancel.
// @Summary      Cancel push
// @Description  Cancels a pending or in-progress push update
// @Tags         updates
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        pushId  path  string  true  "push ID"
// @Success      200  {object}  openapi.UpdateCancelPushResult  "push cancelled"
// @Failure      400  {object}  openapi.ErrorResponse  "pushId required"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      404  {object}  openapi.ErrorResponse  "push not found"
// @Failure      409  {object}  openapi.ErrorResponse  "push cannot be cancelled"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /updates/history/{pushId}/cancel [post]
func (h *UpdatesHistoryHandler) CancelPush(c *gin.Context) {
	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	pushID := c.Param("pushId")
	if pushID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Push ID is required"))
		return
	}

	operator := middleware.GetOperatorFromContext(c)
	if operator == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Operator not found"))
		return
	}

	result, err := h.service.CancelPush(c.Request.Context(), pushID, operator.ID, orgID)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to cancel push"))
		return
	}

	// Audit log the cancellation.
	if h.auditLogger != nil {
		h.auditLogger.UpdateCancelled(
			c.Request.Context(),
			operator.ID,
			pushID,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)
	}

	c.JSON(http.StatusOK, result)
}
