package auth

import (
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.AdminOperatorListResult
	_ openapi.AdminOperator
	_ openapi.CreateOperatorRequest
	_ openapi.UpdateOperatorRequest
	_ openapi.SuccessResult
	_ openapi.ErrorResponse
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
// @Summary      List operators
// @Description  Lists all operators in the system (super_admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.AdminOperatorListResult  "operators"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/admin/operators [get]
func (h *AdminHandler) ListOperators(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	orgID := middleware.GetOrganizationID(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "not authenticated"))
		return
	}

	// Org-scoped check - operator must be super_admin in this organization.
	if !op.IsSuperAdminIn(orgID) {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthzInsufficientPermissions, "insufficient privileges"))
		return
	}

	operators, total, err := h.authService.ListAllOperators(c.Request.Context(), 20, 0)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to list operators"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"operators": operators, "total": total})
}

// CreateOperator handles POST /v1/auth/admin/operators.
// @Summary      Create operator
// @Description  Creates a new operator (super_admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        body  body  openapi.CreateOperatorRequest  true  "operator credentials"
// @Success      201  {object}  openapi.AdminOperator  "created operator"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      409  {object}  openapi.ErrorResponse  "email already in use"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/admin/operators [post]
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
	h.presenter.Created(c, schema.AdminOperator{
		ID:           newOp.ID,
		Email:        newOp.Email,
		Name:         newOp.Name,
		CreatedAt:    newOp.CreatedAt.UnixMilli(),
	})
}

// GetOperator handles GET /v1/auth/admin/operators/:id.
// @Summary      Get operator
// @Description  Returns a single operator by ID (super_admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "operator ID"
// @Success      200  {object}  openapi.AdminOperator  "operator"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      404  {object}  openapi.ErrorResponse  "not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/admin/operators/{id} [get]
func (h *AdminHandler) GetOperator(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	orgID := middleware.GetOrganizationID(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "not authenticated"))
		return
	}

	// Org-scoped check.
	if !op.IsSuperAdminIn(orgID) {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthzInsufficientPermissions, "insufficient privileges"))
		return
	}

	operatorID := c.Param("id")
	if operatorID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "operator id is required"))
		return
	}

	targetOp, err := h.authService.GetOperatorByID(c.Request.Context(), operatorID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "not_found"))
		return
	}

	c.JSON(http.StatusOK, schema.AdminOperator{
		ID:              targetOp.ID,
		Email:            targetOp.Email,
		Name:             targetOp.Name,
		MFAEnabled:       targetOp.MFASecret != "",
		EmailVerified:    targetOp.EmailVerified,
		CreatedAt:        targetOp.CreatedAt.UnixMilli(),
		UpdatedAt:        targetOp.UpdatedAt.UnixMilli(),
	})
}

// UpdateOperator handles PATCH /v1/auth/admin/operators/:id.
// @Summary      Update operator
// @Description  Updates mutable operator fields (super_admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id    path  string  true  "operator ID"
// @Param        body  body  openapi.UpdateOperatorRequest  true  "operator update"
// @Success      200  {object}  openapi.AdminOperator  "updated operator"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      404  {object}  openapi.ErrorResponse  "not found"
// @Failure      409  {object}  openapi.ErrorResponse  "email already in use"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/admin/operators/{id} [patch]
func (h *AdminHandler) UpdateOperator(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	orgID := middleware.GetOrganizationID(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "not authenticated"))
		return
	}

	// Org-scoped check.
	if !op.IsSuperAdminIn(orgID) {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthzInsufficientPermissions, "insufficient privileges"))
		return
	}

	operatorID := c.Param("id")
	if operatorID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "operator id is required"))
		return
	}

	var req auth.UpdateOperatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body"))
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

	c.JSON(http.StatusOK, schema.AdminOperator{
		ID:           updatedOp.ID,
		Email:        updatedOp.Email,
		Name:         updatedOp.Name,
		UpdatedAt:    updatedOp.UpdatedAt.UnixMilli(),
	})
}

// DeleteOperator handles DELETE /v1/auth/admin/operators/:id.
// @Summary      Delete operator
// @Description  Deletes an operator. Cannot delete yourself or the last super_admin
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "operator ID"
// @Success      200  {object}  openapi.SuccessResult  "operator deleted"
// @Failure      400  {object}  openapi.ErrorResponse  "operator id required / cannot delete self"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      404  {object}  openapi.ErrorResponse  "not found"
// @Failure      409  {object}  openapi.ErrorResponse  "cannot delete last super_admin"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/admin/operators/{id} [delete]
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
	h.presenter.OK(c, schema.SuccessResult{Success: true, Message: "operator deleted"})
}
