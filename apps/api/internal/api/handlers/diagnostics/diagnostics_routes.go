// Package diagnostics provides HTTP handlers for device diagnostics.
package diagnostics

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers the diagnostics routes.
func RegisterRoutes(router *gin.RouterGroup, inspectHandler *InspectHandler, timelineHandler *TimelineHandler) {
	if inspectHandler != nil {
		router.GET("/:imei/inspect", inspectHandler.RateLimit(), inspectHandler.GetDeviceInspection)
	}
	if timelineHandler != nil {
		router.GET("/:imei/timeline", timelineHandler.RateLimit(), timelineHandler.GetDeviceTimeline)
	}
}
