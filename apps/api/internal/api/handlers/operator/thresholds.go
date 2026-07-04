package operator

import (
	"net/http"

	appoperator "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"

	"github.com/gin-gonic/gin"
)

// ThresholdHandler handles threshold endpoints.
type ThresholdHandler struct {
	service *appoperator.ThresholdService
}

// NewThresholdHandler creates a new ThresholdHandler.
func NewThresholdHandler(svc *appoperator.ThresholdService) *ThresholdHandler {
	return &ThresholdHandler{service: svc}
}

// GetThresholds handles GET /v1/auth/me/thresholds.
func (h *ThresholdHandler) GetThresholds(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	thresholds, err := h.service.GetThresholds(c.Request.Context(), op.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get thresholds"})
		return
	}

	c.JSON(http.StatusOK, thresholds)
}

// PatchThresholds handles PATCH /v1/auth/me/thresholds.
func (h *ThresholdHandler) PatchThresholds(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req operator.ThresholdsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	thresholds, err := h.service.UpdateThresholds(c.Request.Context(), op.ID, &req)
	if err != nil {
		if err == operator.ErrValidation {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update thresholds"})
		return
	}

	c.JSON(http.StatusOK, thresholds)
}
