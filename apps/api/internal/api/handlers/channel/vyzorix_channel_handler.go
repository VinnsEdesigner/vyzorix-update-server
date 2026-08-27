// Package channel handles client-facing subscription management for
// org-scoped live channels (subscribe/unsubscribe/status).
package channel

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	wschannel "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws/channel"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/schema"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.ChannelSubscribeRequest
	_ openapi.ChannelStatusResult
	_ openapi.ChannelSubscribeResult
	_ openapi.ChannelUnsubscribeResult
	_ openapi.ErrorResponse
)

// Handler serves channel subscription requests.
type Handler struct {
	bridge    *wschannel.HubBridge
	presenter *response.Presenter
}

// NewHandler creates a channel subscription Handler.
func NewHandler(bridge *wschannel.HubBridge, presenter *response.Presenter) *Handler {
	return &Handler{bridge: bridge, presenter: presenter}
}

// Status returns active channel streams for the org.
// @Summary      Channel status
// @Description  Returns active channel streams for the org
// @Tags         channels
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.ChannelStatusResult  "channel status"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /channels/status [get]
func (h *Handler) Status(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	streams := h.bridge.Manager().StreamCount()
	c.JSON(http.StatusOK, schema.ChannelStatusResult{Org: orgID, ActiveStreams: streams})
}

// Subscribe registers a logical subscription to a channel scope for the
// operator. Events arrive on the operator's existing websocket.
// @Summary      Subscribe to channel
// @Description  Registers a logical subscription to a channel scope for the operator
// @Tags         channels
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        body  body  openapi.ChannelSubscribeRequest  true  "subscription request"
// @Success      200  {object}  openapi.ChannelSubscribeResult  "subscription result"
// @Failure      400  {object}  openapi.ErrorResponse  "scope required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Router       /channels/subscribe [post]
func (h *Handler) Subscribe(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}
	orgID := middleware.GetOrganizationID(c)

	var req struct {
		Scope string `json:"scope" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "scope required")
		return
	}

	addr := "stream/" + orgID + "/" + req.Scope
	ch, err := wschannel.Parse(addr)
	if err != nil {
		h.presenter.BadRequest(c, err.Error())
		return
	}
	if !validScope(ch.Scope) {
		h.presenter.BadRequest(c, "unknown scope: "+ch.Scope)
		return
	}

	if _, err := h.bridge.Manager().Subscribe(c.Request.Context(), wschannel.SubscribeEvent{SubjectID: op.ID, Channel: ch}); err != nil {
		h.presenter.Forbidden(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, schema.ChannelSubscribeResult{Subscribed: addr})
}

// Unsubscribe removes a logical subscription.
// @Summary      Unsubscribe from channel
// @Description  Removes a logical subscription to a channel scope
// @Tags         channels
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        scope  query string  true  "channel scope"
// @Success      200  {object}  openapi.ChannelUnsubscribeResult  "unsubscription result"
// @Failure      400  {object}  openapi.ErrorResponse  "scope required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Router       /channels/unsubscribe [post]
func (h *Handler) Unsubscribe(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}
	orgID := middleware.GetOrganizationID(c)
	scope := c.Query("scope")
	if scope == "" {
		h.presenter.BadRequest(c, "scope required")
		return
	}
	h.bridge.Unsubscribe("stream/"+orgID+"/"+scope, op.ID)
	c.JSON(http.StatusOK, schema.ChannelUnsubscribeResult{Unsubscribed: orgID + "/" + scope})
}

func validScope(scope string) bool {
	switch scope {
	case wschannel.Alert, wschannel.Commands, wschannel.Members:
		return true
	}
	return false
}
