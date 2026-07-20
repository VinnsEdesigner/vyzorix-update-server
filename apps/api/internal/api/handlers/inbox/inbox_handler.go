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
// Returns paginated list of inbox entries for the authenticated operator within the organization.
func (h *Handler) GetInbox(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "organization context required",
		})
		return
	}

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

	operator := middleware.GetOperatorFromContext(c)
	operatorID := ""
	if operator != nil {
		operatorID = operator.ID
	}

	result, err := h.service.GetInbox(c.Request.Context(), operatorID, orgID, status, page, limit)
	if err != nil {
		se := inbox.ToServiceError(err)
		c.JSON(se.Status, se.ToErrorResponse())
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetInboxEntry handles GET /v1/device/inbox/:imei.
// Returns a single inbox entry by IMEI within the organization.
func (h *Handler) GetInboxEntry(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "organization context required",
		})
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "IMEI is required",
		})
		return
	}

	result, err := h.service.GetInboxEntry(c.Request.Context(), imei, orgID)
	if err != nil {
		se := inbox.ToServiceError(err)
		c.JSON(se.Status, se.ToErrorResponse())
		return
	}

	c.JSON(http.StatusOK, result)
}

// AckInbox handles POST /v1/device/inbox/:imei/ack.
// Acknowledges (approves or rejects) an inbox entry within the organization.
func (h *Handler) AckInbox(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "organization context required",
		})
		return
	}

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

	if req.Action != "acknowledge" && req.Action != "approve" && req.Action != "reject" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "Action must be 'acknowledge', 'approve', or 'reject'",
		})
		return
	}

	var operatorID string
	if req.Action != "acknowledge" {
		operator := middleware.GetOperatorFromContext(c)
		if operator != nil {
			operatorID = operator.ID
		}
	}

	result, err := h.service.AckInbox(c.Request.Context(), imei, req.Action, operatorID, orgID, req.Notes)
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
// Updates an inbox entry (e.g., add operator notes) within the organization.
func (h *Handler) UpdateInboxEntry(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "organization context required",
		})
		return
	}

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

	result, err := h.service.UpdateInboxEntry(c.Request.Context(), imei, operatorID, orgID, req.Notes)
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
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
			Code:    "bad_request",
			Message: "organization context required",
		})
		return
	}

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

	result, err := h.service.ResendApproval(c.Request.Context(), imei, operatorID, orgID)
	if err != nil {
		se := inbox.ToServiceError(err)
		c.JSON(se.Status, se.ToErrorResponse())
		return
	}

	c.JSON(http.StatusOK, result)
}
