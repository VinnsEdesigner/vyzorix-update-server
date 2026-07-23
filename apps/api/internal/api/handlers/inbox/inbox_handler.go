package inbox

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/inbox"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/appcheck"
	"github.com/gin-gonic/gin"
)

// Handler handles inbox-related HTTP requests.
type Handler struct {
	service              *inbox.Service
	deviceSecret        string
	attestationRequired bool 
	appCheckVerifier    *appcheck.Verifier
}

// NewHandler creates a new InboxHandler.
func NewHandler(service *inbox.Service, deviceSecret string) *Handler {
	return &Handler{
		service:              service,
		deviceSecret:        deviceSecret,
		attestationRequired: false,
	}
}

// NewHandlerWithAttestation creates a new InboxHandler with required attestation.
// Use this in production where device attestation is mandatory.
func NewHandlerWithAttestation(service *inbox.Service, deviceSecret string) *Handler {
	return &Handler{
		service:              service,
		deviceSecret:        deviceSecret,
		attestationRequired: true,
	}
}

// NewHandlerWithAppCheck creates a new InboxHandler with Firebase App Check verification.
// This is the recommended configuration for production.
func NewHandlerWithAppCheck(service *inbox.Service, deviceSecret string, appCheckVerifier *appcheck.Verifier) *Handler {
	return &Handler{
		service:              service,
		deviceSecret:        deviceSecret,
		attestationRequired: true,
		appCheckVerifier:    appCheckVerifier,
	}
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
// Requires attestation via Firebase App Check (preferred) or X-Device-Signature header (HMAC fallback).
func (h *Handler) CreateInboxRequest(c *gin.Context) {
	requiresAttestation := h.deviceSecret != "" || h.attestationRequired

	if requiresAttestation {
		// Priority 1: Firebase App Check (hardware-backed attestation).
		if h.appCheckVerifier != nil && h.appCheckVerifier.Enabled() {
			token := c.GetHeader("X-Firebase-AppCheck")
			if token == "" {
				c.JSON(http.StatusUnauthorized, inbox.ErrorResponse{
					Code:    "unauthorized",
					Message: "Missing X-Firebase-AppCheck header",
				})
				return
			}

			decoded, err := h.appCheckVerifier.VerifyToken(c.Request.Context(), token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, inbox.ErrorResponse{
					Code:    "unauthorized",
					Message: "Invalid Firebase App Check token",
				})
				return
			}

			// Log successful attestation for monitoring.
			c.Set("app_check_app_id", decoded.AppID)

		} else if h.deviceSecret != "" {
			// Priority 2: HMAC-SHA256 signature (legacy fallback).
			signature := c.GetHeader("X-Device-Signature")
			if signature == "" {
				c.JSON(http.StatusUnauthorized, inbox.ErrorResponse{
					Code:    "unauthorized",
					Message: "Missing X-Device-Signature header",
				})
				return
			}

			// Read body for signature verification.
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, inbox.ErrorResponse{
					Code:    "bad_request",
					Message: "Failed to read request body",
				})
				return
			}

			// Verify HMAC-SHA256 signature.
			mac := hmac.New(sha256.New, []byte(h.deviceSecret))
			mac.Write(body)
			expectedSig := hex.EncodeToString(mac.Sum(nil))
			if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSig)) != 1 {
				c.JSON(http.StatusUnauthorized, inbox.ErrorResponse{
					Code:    "unauthorized",
					Message: "Invalid X-Device-Signature header",
				})
				return
			}

			// Restore body for binding.
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
	}

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
