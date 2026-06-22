package auth

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"

	"github.com/gin-gonic/gin"
)

// AdminHandler handles admin endpoints.
type AdminHandler struct {
	authService *auth.AuthService
	presenter  *response.Presenter
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(authService *auth.AuthService, presenter *response.Presenter) *AdminHandler {
	return &AdminHandler{authService: authService, presenter: presenter}
}

// ListOperators handles GET /v1/auth/admin/operators.
func (h *AdminHandler) ListOperators(c *gin.Context) {
	// Verify session first
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
		return
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
		return
	}

	// Check role
	if op.Role != "super_admin" && op.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "forbidden"})
		return
	}

	operators, total, err := h.authService.ListAllOperators(c.Request.Context(), 20, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to list operators"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"operators": operators, "total": total})
}

// CreateOperator handles POST /v1/auth/admin/operators.
func (h *AdminHandler) CreateOperator(c *gin.Context) {
	// Verify session first
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	// Check role
	if op.Role != "super_admin" && op.Role != "admin" {
		h.presenter.Forbidden(c, "forbidden")
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
		h.presenter.BadRequest(c, "Invalid request")
		return
	}

	h.presenter.AdminAction(c, op.ID, "create_operator", "operator", newOp.ID, nil)
	h.presenter.Created(c, gin.H{
		"id":         newOp.ID,
		"email":      newOp.Email,
		"name":       newOp.Name,
		"role":       string(newOp.Role),
		"created_at": newOp.CreatedAt.UnixMilli(),
	})
}

// GetOperator handles GET /v1/auth/admin/operators/:id.
func (h *AdminHandler) GetOperator(c *gin.Context) {
	// Verify session first
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
		return
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
		return
	}

	// Check role
	if op.Role != "super_admin" && op.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "forbidden"})
		return
	}

	operatorID := c.Param("id")
	if operatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "operator id is required"})
		return
	}

	targetOp, err := h.authService.GetOperatorByID(c.Request.Context(), operatorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             targetOp.ID,
		"email":          targetOp.Email,
		"name":           targetOp.Name,
		"role":           string(targetOp.Role),
		"mfa_enabled":    targetOp.MFASecret != "",
		"email_verified": targetOp.EmailVerified,
		"created_at":     targetOp.CreatedAt.UnixMilli(),
		"updated_at":     targetOp.UpdatedAt.UnixMilli(),
	})
}

// UpdateOperator handles PATCH /v1/auth/admin/operators/:id.
func (h *AdminHandler) UpdateOperator(c *gin.Context) {
	// Verify session first
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
		return
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
		return
	}

	// Check role
	if op.Role != "super_admin" && op.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "forbidden"})
		return
	}

	operatorID := c.Param("id")
	if operatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "operator id is required"})
		return
	}

	var req auth.UpdateOperatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid request body"})
		return
	}

	updatedOp, err := h.authService.UpdateOperator(c.Request.Context(), operatorID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "internal_error", "message": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         updatedOp.ID,
		"email":      updatedOp.Email,
		"name":       updatedOp.Name,
		"role":       string(updatedOp.Role),
		"updated_at": updatedOp.UpdatedAt.UnixMilli(),
	})
}

// DeleteOperator handles DELETE /v1/auth/admin/operators/:id.
func (h *AdminHandler) DeleteOperator(c *gin.Context) {
	// Verify session first
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	// Check role
	if op.Role != "super_admin" && op.Role != "admin" {
		h.presenter.Forbidden(c, "forbidden")
		return
	}

	operatorID := c.Param("id")
	if operatorID == "" {
		h.presenter.BadRequest(c, "operator id is required")
		return
	}

	// Prevent deleting yourself
	if operatorID == op.ID {
		h.presenter.BadRequest(c, "cannot delete your own account")
		return
	}

	if err := h.authService.DeleteOperator(c.Request.Context(), operatorID); err != nil {
		h.presenter.InternalError(c, "Invalid request")
		return
	}

	h.presenter.AdminAction(c, op.ID, "delete_operator", "operator", operatorID, nil)
	h.presenter.OK(c, gin.H{"success": true, "message": "operator deleted"})
}
