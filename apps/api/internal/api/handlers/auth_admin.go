package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

// ListOperators returns all operators in the system (super_admin only).
// GET /v1/auth/admin/operators.
func (ac *AuthController) ListOperators(c *gin.Context) {
	op := GetOperatorFromContext(c)
	if op == nil || op.Role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "only super_admin can list operators"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	operators, err := ac.store.ListOperators(ctx)
	if err != nil {
		ac.log.Warn("listOperators: failed", "err", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "failed to list operators"})
		return
	}

	response := make([]models.OperatorResponse, 0, len(operators))
	for _, o := range operators {
		response = append(response, o.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{"operators": response})
}