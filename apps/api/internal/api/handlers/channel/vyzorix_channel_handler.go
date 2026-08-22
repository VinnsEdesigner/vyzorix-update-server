// Package channel handles client-facing subscription management for
// org-scoped live channels (subscribe/unsubscribe/status).
package channel

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	wschannel "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws/channel"
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
func (h *Handler) Status(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	streams := h.bridge.Manager().StreamCount()
	c.JSON(http.StatusOK, gin.H{"org": orgID, "active_streams": streams})
}

// Subscribe registers a logical subscription to a channel scope for the
// operator. Events arrive on the operator's existing websocket.
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
	c.JSON(http.StatusOK, gin.H{"subscribed": addr})
}

// Unsubscribe removes a logical subscription.
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
	c.JSON(http.StatusOK, gin.H{"unsubscribed": orgID + "/" + scope})
}

func validScope(scope string) bool {
	switch scope {
	case wschannel.Alert, wschannel.Commands, wschannel.Members:
		return true
	}
	return false
}
