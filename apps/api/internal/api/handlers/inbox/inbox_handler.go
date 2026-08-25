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
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/inbox"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/appcheck"
	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.InboxListResult
	_ openapi.InboxEntryResponse
	_ openapi.InboxAckResult
	_ openapi.InboxResendResult
	_ openapi.InboxRequest
	_ openapi.InboxAckRequest
	_ openapi.UpdateInboxEntryRequest
	_ openapi.ErrorResponse
)

// Handler handles inbox-related HTTP requests.
type Handler struct {
	service             *inbox.Service
	appCheckVerifier    *appcheck.Verifier
	deviceSecret        string
	attestationRequired bool
}

// NewHandler creates a new InboxHandler.
func NewHandler(service *inbox.Service, deviceSecret string) *Handler {
	return &Handler{
		service:             service,
		deviceSecret:        deviceSecret,
		attestationRequired: false,
	}
}

// NewHandlerWithAttestation creates a new InboxHandler with required attestation.
// Use this in production where device attestation is mandatory.
func NewHandlerWithAttestation(service *inbox.Service, deviceSecret string) *Handler {
	return &Handler{
		service:             service,
		deviceSecret:        deviceSecret,
		attestationRequired: true,
	}
}

// NewHandlerWithAppCheck creates a new InboxHandler with Firebase App Check verification.
// This is the recommended configuration for production.
func NewHandlerWithAppCheck(service *inbox.Service, deviceSecret string, appCheckVerifier *appcheck.Verifier) *Handler {
	return &Handler{
		service:             service,
		deviceSecret:        deviceSecret,
		attestationRequired: true,
		appCheckVerifier:    appCheckVerifier,
	}
}

// GetInbox handles GET /v1/inbox.
// @Summary      List inbox entries
// @Description  Returns paginated inbox entries for the authenticated operator within the organization
// @Tags         inbox
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        status  query string  false  "filter by status"
// @Param        page    query int    false  "page number (default 1)"
// @Param        limit   query int    false  "page size (default 20)"
// @Success      200  {object}  openapi.InboxListResult  "inbox entries"
// @Failure      400  {object}  openapi.ErrorResponse  "organization context required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/inbox [get]
func (h *Handler) GetInbox(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
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
		_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetInboxEntry handles GET /v1/inbox/:imei.
// @Summary      Get inbox entry
// @Description  Returns a single inbox entry by IMEI within the organization
// @Tags         inbox
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei     path  string  true  "device IMEI"
// @Success      200  {object}  openapi.InboxEntryResponse  "inbox entry"
// @Failure      400  {object}  openapi.ErrorResponse  "IMEI required"
// @Failure      404  {object}  openapi.ErrorResponse  "not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/inbox/{imei} [get]
func (h *Handler) GetInboxEntry(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "IMEI is required"))
		return
	}

	result, err := h.service.GetInboxEntry(c.Request.Context(), imei, orgID)
	if err != nil {
		se := inbox.ToServiceError(err)
		_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
		return
	}

	c.JSON(http.StatusOK, result)
}

// AckInbox handles POST /v1/device/inbox/:imei/ack.
// @Summary      Acknowledge inbox entry
// @Description  Acknowledges (approves or rejects) an inbox entry within the organization
// @Tags         inbox
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei     path  string  true  "device IMEI"
// @Param        body  body  openapi.InboxAckRequest  true  "ack action (acknowledge|approve|reject)"
// @Success      200  {object}  openapi.InboxAckResult  "ack result"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid action / IMEI required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/inbox/{imei}/ack [post]
func (h *Handler) AckInbox(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "IMEI is required"))
		return
	}

	var req inbox.AckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid request body"))
		return
	}

	if req.Action != "acknowledge" && req.Action != "approve" && req.Action != "reject" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Action must be 'acknowledge', 'approve', or 'reject'"))
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
		_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateInboxRequest handles POST /v1/inbox.
// @Summary      Create inbox request
// @Description  Submits a device registration request to the operator inbox. Requires device attestation when configured
// @Tags         inbox
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID    header  string  false  "Organization ID"
// @Param        X-Device-Signature   header  string  false  "HMAC-SHA256 body signature (legacy attestation)"
// @Param        X-Firebase-AppCheck  header  string  false  "Firebase App Check token (recommended attestation)"
// @Param        body  body  openapi.InboxRequest  true  "device registration request"
// @Success      201  {object}  openapi.InboxEntryResponse  "created inbox entry"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid request body"
// @Failure      401  {object}  openapi.ErrorResponse  "attestation failed"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/inbox [post]
func (h *Handler) CreateInboxRequest(c *gin.Context) {
	requiresAttestation := h.deviceSecret != "" || h.attestationRequired

	if requiresAttestation {
		// Priority 1: Firebase App Check (hardware-backed attestation).
		if h.appCheckVerifier != nil && h.appCheckVerifier.Enabled() {
			token := c.GetHeader("X-Firebase-AppCheck")
			if token == "" {
				_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Missing X-Firebase-AppCheck header"))
				return
			}

			decoded, err := h.appCheckVerifier.VerifyToken(c.Request.Context(), token)
			if err != nil {
				_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Invalid Firebase App Check token"))
				return
			}

			// Log successful attestation for monitoring.
			c.Set("app_check_app_id", decoded.AppID)

		} else if h.deviceSecret != "" {
			// Priority 2: HMAC-SHA256 signature (legacy fallback).
			signature := c.GetHeader("X-Device-Signature")
			if signature == "" {
				_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Missing X-Device-Signature header"))
				return
			}

			// Read body for signature verification.
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Failed to read request body"))
				return
			}

			// Verify HMAC-SHA256 signature.
			mac := hmac.New(sha256.New, []byte(h.deviceSecret))
			mac.Write(body)
			expectedSig := hex.EncodeToString(mac.Sum(nil))
			if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSig)) != 1 {
				_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Invalid X-Device-Signature header"))
				return
			}

			// Restore body for binding.
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
	}

	var req inbox.InboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid request body"))
		return
	}

	result, err := h.service.CreateInboxRequest(c.Request.Context(), &req)
	if err != nil {
		se := inbox.ToServiceError(err)
		_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
		return
	}

	c.JSON(http.StatusCreated, result)
}

// UpdateInboxEntryRequest represents the request for PATCH /v1/device/inbox/:imei.
type UpdateInboxEntryRequest struct {
	Notes string `json:"notes,omitempty"`
}

// UpdateInboxEntry handles PATCH /v1/device/inbox/:imei.
// @Summary      Update inbox entry
// @Description  Updates an inbox entry (e.g., add operator notes) within the organization
// @Tags         inbox
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei     path  string  true  "device IMEI"
// @Param        body  body  openapi.UpdateInboxEntryRequest  true  "inbox update (notes)"
// @Success      200  {object}  openapi.InboxEntryResponse  "updated inbox entry"
// @Failure      400  {object}  openapi.ErrorResponse  "IMEI required / invalid body"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/inbox/{imei} [patch]
func (h *Handler) UpdateInboxEntry(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "IMEI is required"))
		return
	}

	var req schema.UpdateInboxEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid request body"))
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
		_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
		return
	}

	c.JSON(http.StatusOK, result)
}

// ResendApproval handles POST /v1/inbox/:imei/resend.
// @Summary      Resend approval notification
// @Description  Resends the FCM notification to a device that was approved but may have missed the notification
// @Tags         inbox
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei     path  string  true  "device IMEI"
// @Success      200  {object}  openapi.InboxResendResult  "resend result"
// @Failure      400  {object}  openapi.ErrorResponse  "IMEI required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/inbox/{imei}/resend [post]
func (h *Handler) ResendApproval(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "IMEI is required"))
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
		_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
		return
	}

	c.JSON(http.StatusOK, result)
}
