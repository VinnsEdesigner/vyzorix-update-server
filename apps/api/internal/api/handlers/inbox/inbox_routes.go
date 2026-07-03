package inbox

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all inbox routes.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// Inbox endpoints - authenticated operators
	rg.GET("/inbox", h.GetInbox)
	rg.GET("/inbox/:imei", h.GetInboxEntry)
	rg.POST("/inbox/:imei/ack", h.AckInbox)

	// Device registration inbox submission (used by devices)
	rg.POST("/inbox", h.CreateInboxRequest)
}
