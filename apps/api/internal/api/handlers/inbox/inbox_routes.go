package inbox

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all inbox routes.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// Inbox endpoints - authenticated operators
	rg.GET("/device/inbox", h.GetInbox)
	rg.GET("/device/inbox/:imei", h.GetInboxEntry)
	rg.POST("/device/inbox/:imei/ack", h.AckInbox)

	// Device registration inbox submission (used by devices)
	rg.POST("/device/inbox", h.CreateInboxRequest)
}
