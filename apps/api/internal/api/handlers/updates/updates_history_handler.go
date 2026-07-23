package updates

import (
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/gin-gonic/gin"
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
func (h *UpdatesHistoryHandler) GetHistory(c *gin.Context) {
	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, updates.ErrorResponse{
			Code:    "bad_request",
			Message: "organization context required",
		})
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
			c.JSON(se.Status, se.ToErrorResponse())
			return
		}
		c.JSON(http.StatusInternalServerError, updates.ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to get history",
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetPushDetail handles GET /v1/updates/history/:pushId.
func (h *UpdatesHistoryHandler) GetPushDetail(c *gin.Context) {
	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, updates.ErrorResponse{
			Code:    "bad_request",
			Message: "organization context required",
		})
		return
	}

	pushID := c.Param("pushId")
	if pushID == "" {
		c.JSON(http.StatusBadRequest, updates.ErrorResponse{
			Code:    "bad_request",
			Message: "Push ID is required",
		})
		return
	}

	detail, err := h.service.GetPushDetail(c.Request.Context(), pushID, orgID)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			c.JSON(se.Status, se.ToErrorResponse())
			return
		}
		c.JSON(http.StatusInternalServerError, updates.ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to get push detail",
		})
		return
	}
	c.JSON(http.StatusOK, detail)
}

// CancelPush handles POST /v1/updates/history/:pushId/cancel.
func (h *UpdatesHistoryHandler) CancelPush(c *gin.Context) {
	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, updates.ErrorResponse{
			Code:    "bad_request",
			Message: "organization context required",
		})
		return
	}

	pushID := c.Param("pushId")
	if pushID == "" {
		c.JSON(http.StatusBadRequest, updates.ErrorResponse{
			Code:    "bad_request",
			Message: "Push ID is required",
		})
		return
	}

	operator := middleware.GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, updates.ErrorResponse{
			Code:    "unauthorized",
			Message: "Operator not found",
		})
		return
	}

	result, err := h.service.CancelPush(c.Request.Context(), pushID, operator.ID, orgID)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			c.JSON(se.Status, se.ToErrorResponse())
			return
		}
		c.JSON(http.StatusInternalServerError, updates.ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to cancel push",
		})
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
