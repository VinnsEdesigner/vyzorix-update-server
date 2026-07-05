package inbox

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all inbox routes (authenticated).
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// Inbox endpoints - authenticated operators
	rg.GET("/inbox", h.GetInbox)
	rg.GET("/inbox/:imei", h.GetInboxEntry)
	rg.PATCH("/inbox/:imei", h.UpdateInboxEntry) // FR-37: notes/annotations update (implemented)
	rg.POST("/inbox/:imei/ack", h.AckInbox)
	rg.POST("/inbox/:imei/resend", h.ResendApproval) // Resend FCM notification for approved devices
}

// RegisterPublicRoutes registers public inbox routes (no auth required).
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	// Device registration inbox submission (used by devices)
	rg.POST("/inbox", h.CreateInboxRequest)
}
