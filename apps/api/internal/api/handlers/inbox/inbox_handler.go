package inbox

import (
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/inbox"
	"github.com/gin-gonic/gin"
)

// Handler handles inbox-related HTTP requests.
type Handler struct {
	service *inbox.Service
}

// NewHandler creates a new InboxHandler.
func NewHandler(service *inbox.Service) *Handler {
	return &Handler{service: service}
}

// GetInbox handles GET /v1/device/inbox.
// Returns paginated list of inbox entries for the authenticated operator.
func (h *Handler) GetInbox(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
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

	// Get operator from context (set by auth middleware)
	operator := middleware.GetOperatorFromContext(c)
	operatorID := ""
	if operator != nil {
		operatorID = operator.ID
	}

	result, err := h.service.GetInbox(c.Request.Context(), operatorID, status, page, limit)
	if err != nil {
		se := inbox.ToServiceError(err)
		c.JSON(se.Status, se.ToErrorResponse())
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetInboxEntry handles GET /v1/device/inbox/:imei.
// Returns a single inbox entry by IMEI.
func (h *Handler) GetInboxEntry(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "IMEI is required",
		})
		return
	}

	result, err := h.service.GetInboxEntry(c.Request.Context(), imei)
	if err != nil {
		se := inbox.ToServiceError(err)
		c.JSON(se.Status, se.ToErrorResponse())
		return
	}

	c.JSON(http.StatusOK, result)
}

// AckInbox handles POST /v1/device/inbox/:imei/ack.
// Acknowledges (approves or rejects) an inbox entry.
func (h *Handler) AckInbox(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "IMEI is required",
		})
		return
	}

	var req inbox.AckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Validate action - support all 3 actions from 5-state model
	if req.Action != "acknowledge" && req.Action != "approve" && req.Action != "reject" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "Action must be 'acknowledge', 'approve', or 'reject'",
		})
		return
	}

	// For acknowledge action, no operator ID needed (device-side action)
	// For approve/reject actions, get operator ID from context
	var operatorID string
	if req.Action != "acknowledge" {
		operator := middleware.GetOperatorFromContext(c)
		if operator != nil {
			operatorID = operator.ID
		}
	}

	// Timeout is handled by middleware (Bug 49)
	result, err := h.service.AckInbox(c.Request.Context(), imei, req.Action, operatorID, req.Notes)
	if err != nil {
		se := inbox.ToServiceError(err)
		c.JSON(se.Status, se.ToErrorResponse())
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateInboxRequest handles POST /v1/device/inbox.
// Creates a new inbox entry (used by device registration flow).
func (h *Handler) CreateInboxRequest(c *gin.Context) {
	var req inbox.InboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	if req.IMEI == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "IMEI is required",
		})
		return
	}

	// Timeout and idempotency are handled by middleware (Bug 49, Bug 45)
	result, err := h.service.CreateInboxRequest(c.Request.Context(), &req)
	if err != nil {
		se := inbox.ToServiceError(err)
		c.JSON(se.Status, se.ToErrorResponse())
		return
	}

	c.JSON(http.StatusCreated, result)
}

// UpdateInboxEntryRequest represents the request for PATCH /v1/device/inbox/:imei.
type UpdateInboxEntryRequest struct {
	Notes string `json:"notes,omitempty"`
}

// UpdateInboxEntry handles PATCH /v1/device/inbox/:imei.
// Updates an inbox entry (e.g., add operator notes).
func (h *Handler) UpdateInboxEntry(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "IMEI is required",
		})
		return
	}

	var req UpdateInboxEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	operator := middleware.GetOperatorFromContext(c)
	operatorID := ""
	if operator != nil {
		operatorID = operator.ID
	}

	result, err := h.service.UpdateInboxEntry(c.Request.Context(), imei, operatorID, req.Notes)
	if err != nil {
		se := inbox.ToServiceError(err)
		c.JSON(se.Status, se.ToErrorResponse())
		return
	}

	c.JSON(http.StatusOK, result)
}

// ResendApproval handles POST /v1/device/inbox/:imei/resend.
// Resends the FCM notification to a device that was approved but may have missed the notification.
func (h *Handler) ResendApproval(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "IMEI is required",
		})
		return
	}

	operator := middleware.GetOperatorFromContext(c)
	operatorID := ""
	if operator != nil {
		operatorID = operator.ID
	}

	result, err := h.service.ResendApproval(c.Request.Context(), imei, operatorID)
	if err != nil {
		se := inbox.ToServiceError(err)
		c.JSON(se.Status, se.ToErrorResponse())
		return
	}

	c.JSON(http.StatusOK, result)
}
