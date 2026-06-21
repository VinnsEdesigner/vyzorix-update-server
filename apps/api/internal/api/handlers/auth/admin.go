package auth

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"

	"github.com/gin-gonic/gin"
)

// AdminHandler handles admin endpoints.
type AdminHandler struct {
	authService *auth.AuthService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(authService *auth.AuthService) *AdminHandler {
	return &AdminHandler{authService: authService}
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

	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid request body"})
		return
	}

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "email is required"})
		return
	}
	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "password is required"})
		return
	}

	newOp, err := h.authService.CreateOperator(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "internal_error", "message": "Invalid request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
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

	// Prevent deleting yourself
	if operatorID == op.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "cannot delete your own account"})
		return
	}

	if err := h.authService.DeleteOperator(c.Request.Context(), operatorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "operator deleted"})
}
