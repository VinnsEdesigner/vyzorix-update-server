// Package confirmation provides the HTTP endpoints that issue single-use.
// confirmation tokens for risky device commands.
package confirmation

import (
	"errors"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/confirmation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"

	"github.com/gin-gonic/gin"
)

// Handler issues confirmation tokens for risky device commands.
type Handler struct {
	confirmationService *confirmation.Service
	deviceService       *device.Service
	riskEvaluator       *command.RiskEvaluator
}

// NewHandler creates a confirmation Handler. The device service is used to.
// verify the target device belongs to the caller's organization before a.
// confirmation is issued, so tokens never authorize cross-tenant actions.
func NewHandler(confirmationService *confirmation.Service, deviceService *device.Service, riskEvaluator *command.RiskEvaluator) *Handler {
	return &Handler{
		confirmationService: confirmationService,
		deviceService:       deviceService,
		riskEvaluator:       riskEvaluator,
	}
}

// requestConfirmationRequest is the JSON body for POST /v1/device/:imei/command/confirm.
type requestConfirmationRequest struct {
	Command string `json:"command"`
}

// RequestConfirmation issues a single-use confirmation token for a risky.
// command on a device. If the command does not require confirmation, the.
// endpoint responds with confirmation_required:false so the client can issue.
// the command directly.
func (h *Handler) RequestConfirmation(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		responses.RespondStructured(c, http.StatusBadRequest, "device imei required")
		return
	}

	var req requestConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.RespondStructured(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Command == "" {
		responses.RespondStructured(c, http.StatusBadRequest, "command is required")
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		responses.RespondStructured(c, http.StatusBadRequest, "organization context required")
		return
	}

	// Verify the device belongs to this organization before issuing a token.
	if h.deviceService != nil {
		if _, err := h.deviceService.GetDeviceDetailByOrganization(c.Request.Context(), imei, orgID); err != nil {
			responses.RespondStructured(c, http.StatusNotFound, "device not found")
			return
		}
	}

	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		responses.RespondStructured(c, http.StatusUnauthorized, "authentication required")
		return
	}

	// Only issue tokens for commands that actually require confirmation.
	profile := command.LookupRiskProfile(req.Command)
	if !profile.RequiresConfirmation && profile.Tier != command.RiskTierCritical {
		c.JSON(http.StatusOK, gin.H{
			"confirmation_required": false,
			"risk_tier":             string(profile.Tier),
			"trace_id":              middleware.GetTraceID(c),
		})
		return
	}

	pending, err := h.confirmationService.RequestConfirmation(
		c.Request.Context(), op.ID, orgID, req.Command, imei,
	)
	if err != nil {
		responses.RespondStructured(c, http.StatusInternalServerError, "failed to issue confirmation")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"confirmation_token":    pending.Token,
		"confirmation_required": true,
		"risk_tier":             pending.RiskTier,
		"expires_at":            pending.ExpiresAt.Unix(),
		"ttl_seconds":           int(time.Until(pending.ExpiresAt).Seconds()),
		"trace_id":              middleware.GetTraceID(c),
	})
}

// ConsumeForCommand is a thin wrapper exposing the service's consume logic to.
// the command handler without coupling it to the confirmation service's.
// internals. It is the single consume path; the handler calls it when a.
// confirmation token is presented on command execution.
func (h *Handler) ConsumeForCommand(c *gin.Context, token, operatorID, commandName, deviceID string) (*command.CommandRiskProfile, error) {
	pending, err := h.confirmationService.ConsumeForCommand(c.Request.Context(), token, operatorID, commandName, deviceID)
	if err != nil {
		// Translate confirmation errors into a sentinel that the caller maps.
		// to an HTTP status; the profile is still useful for the audit body.
		profile := command.LookupRiskProfile(commandName)
		return &profile, err
	}
	profile := command.LookupRiskProfile(pending.Command)
	return &profile, nil
}

// ErrRequiresConfirmation is returned by the handler-less consume helper when.
// a token is missing; exported so callers can distinguish "no token" from.
// "bad token" without importing confirmation domain errors.
var ErrRequiresConfirmation = errors.New("confirmation token required")
