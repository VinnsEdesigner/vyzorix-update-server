package auth

import (
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"

	"github.com/gin-gonic/gin"
)

// AdminHandler handles admin endpoints with organization-scoped access control.
// All endpoints require the operator to be a super_admin in the current organization.
type AdminHandler struct {
	authService *auth.AuthService
	presenter   *response.Presenter
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(authService *auth.AuthService, presenter *response.Presenter) *AdminHandler {
	return &AdminHandler{authService: authService, presenter: presenter}
}

// ListOperators handles GET /v1/auth/admin/operators.
// Lists all operators in the system (for super_admin reference).
func (h *AdminHandler) ListOperators(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	orgID := middleware.GetOrganizationID(c)
	if op == nil {
		c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "not authenticated"))
		return
	}

	// Org-scoped check - operator must be super_admin in this organization.
	if !op.IsSuperAdminIn(orgID) {
		c.Error(apperrors.NewServerError(apperrors.CodeAuthzInsufficientPermissions, "insufficient privileges"))
		return
	}

	operators, total, err := h.authService.ListAllOperators(c.Request.Context(), 20, 0)
	if err != nil {
		c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to list operators"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"operators": operators, "total": total})
}

// CreateOperator handles POST /v1/auth/admin/operators.
func (h *AdminHandler) CreateOperator(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	orgID := middleware.GetOrganizationID(c)
	if op == nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	// Org-scoped check.
	if !op.IsSuperAdminIn(orgID) {
		h.presenter.Forbidden(c, "insufficient privileges")
		return
	}

	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid request body")
		return
	}

	if req.Email == "" {
		h.presenter.BadRequest(c, "email is required")
		return
	}

	if req.Password == "" {
		h.presenter.BadRequest(c, "password is required")
		return
	}

	newOp, err := h.authService.CreateOperator(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, application.ErrEmailExists) {
			h.presenter.Conflict(c, "email already in use")
			return
		}
		if errors.Is(err, application.ErrInvalidInput) {
			h.presenter.BadRequest(c, "invalid input")
			return
		}
		h.presenter.InternalError(c, "Invalid request")
		return
	}

	h.presenter.AdminAction(c, op.ID, "create_operator", "operator", newOp.ID, nil)
	h.presenter.Created(c, gin.H{
		"id":         newOp.ID,
		"email":      newOp.Email,
		"name":       newOp.Name,
		"created_at": newOp.CreatedAt.UnixMilli(),
	})
}

// GetOperator handles GET /v1/auth/admin/operators/:id.
func (h *AdminHandler) GetOperator(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	orgID := middleware.GetOrganizationID(c)
	if op == nil {
		c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "not authenticated"))
		return
	}

	// Org-scoped check.
	if !op.IsSuperAdminIn(orgID) {
		c.Error(apperrors.NewServerError(apperrors.CodeAuthzInsufficientPermissions, "insufficient privileges"))
		return
	}

	operatorID := c.Param("id")
	if operatorID == "" {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "operator id is required"))
		return
	}

	targetOp, err := h.authService.GetOperatorByID(c.Request.Context(), operatorID)
	if err != nil {
		c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "not_found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             targetOp.ID,
		"email":          targetOp.Email,
		"name":           targetOp.Name,
		"mfa_enabled":    targetOp.MFASecret != "",
		"email_verified": targetOp.EmailVerified,
		"created_at":     targetOp.CreatedAt.UnixMilli(),
		"updated_at":     targetOp.UpdatedAt.UnixMilli(),
	})
}

// UpdateOperator handles PATCH /v1/auth/admin/operators/:id.
func (h *AdminHandler) UpdateOperator(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	orgID := middleware.GetOrganizationID(c)
	if op == nil {
		c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "not authenticated"))
		return
	}

	// Org-scoped check.
	if !op.IsSuperAdminIn(orgID) {
		c.Error(apperrors.NewServerError(apperrors.CodeAuthzInsufficientPermissions, "insufficient privileges"))
		return
	}

	operatorID := c.Param("id")
	if operatorID == "" {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "operator id is required"))
		return
	}

	var req auth.UpdateOperatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}

	updatedOp, err := h.authService.UpdateOperator(c.Request.Context(), operatorID, &req)
	if err != nil {
		if errors.Is(err, application.ErrOperatorNotFound) {
			h.presenter.NotFound(c, "operator not found")
			return
		}
		if errors.Is(err, application.ErrEmailExists) {
			h.presenter.Conflict(c, "email already in use")
			return
		}
		if errors.Is(err, application.ErrInvalidInput) {
			h.presenter.BadRequest(c, "invalid request body")
			return
		}
		h.presenter.InternalError(c, "Invalid request")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         updatedOp.ID,
		"email":      updatedOp.Email,
		"name":       updatedOp.Name,
		"updated_at": updatedOp.UpdatedAt.UnixMilli(),
	})
}

// DeleteOperator handles DELETE /v1/auth/admin/operators/:id.
func (h *AdminHandler) DeleteOperator(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	orgID := middleware.GetOrganizationID(c)
	if op == nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	// Org-scoped check.
	if !op.IsSuperAdminIn(orgID) {
		h.presenter.Forbidden(c, "insufficient privileges")
		return
	}

	operatorID := c.Param("id")
	if operatorID == "" {
		h.presenter.BadRequest(c, "operator id is required")
		return
	}

	if operatorID == op.ID {
		h.presenter.BadRequest(c, "cannot delete your own account")
		return
	}

	if err := h.authService.DeleteOperator(c.Request.Context(), operatorID); err != nil {
		if errors.Is(err, application.ErrOperatorNotFound) {
			h.presenter.NotFound(c, "operator not found")
			return
		}
		if errors.Is(err, application.ErrCannotDeleteLastSuperAdmin) {
			h.presenter.Conflict(c, "cannot delete the last super admin of an organization")
			return
		}
		h.presenter.InternalError(c, "Invalid request")
		return
	}

	h.presenter.AdminAction(c, op.ID, "delete_operator", "operator", operatorID, nil)
	h.presenter.OK(c, gin.H{"success": true, "message": "operator deleted"})
}
