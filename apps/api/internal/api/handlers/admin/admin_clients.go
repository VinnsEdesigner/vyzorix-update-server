package admin

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"

	"github.com/gin-gonic/gin"
)

// ClientsHandler handles admin client management endpoints.
type ClientsHandler struct {
	clientService *client.Service
}

// NewClientsHandler creates a new ClientsHandler.
func NewClientsHandler(clientService *client.Service) *ClientsHandler {
	return &ClientsHandler{clientService: clientService}
}

// requireAdmin is middleware that ensures the request is from an admin user in the org context.
func requireAdmin(c *gin.Context) bool {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}

	// Require org context for admin access.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Organization ID required"})
		return false
	}

	// Check if operator is super_admin in this specific organization.
	if !op.IsSuperAdminIn(orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "admin access required"})
		return false
	}

	return true
}

// List handles GET /v1/admin/clients.
func (h *ClientsHandler) List(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	clients, total, err := h.clientService.ListAll(c.Request.Context(), 20, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to list clients"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clients": clients, "total": total})
}

// Get handles GET /v1/admin/clients/:clientId.
func (h *ClientsHandler) Get(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	clientResp, err := h.clientService.Get(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Client not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"client": clientResp})
}

// Update handles PATCH /v1/admin/clients/:clientId.
func (h *ClientsHandler) Update(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	var req dto.UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request body"})
		return
	}

	clientResp, err := h.clientService.Update(c.Request.Context(), clientID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to update client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"client": clientResp})
}

// Delete handles DELETE /v1/admin/clients/:clientId.
func (h *ClientsHandler) Delete(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	if err := h.clientService.Delete(c.Request.Context(), clientID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to delete client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "clientId": clientID})
}

// RotateKey handles POST /v1/admin/clients/:clientId/rotate-key.
// Rotates the signing key with a 24-hour grace period.
func (h *ClientsHandler) RotateKey(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	keyVersion, err := h.clientService.RotateKey(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to rotate key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Key rotated. Client must fetch new credentials.",
		"clientId":   clientID,
		"keyVersion": keyVersion,
	})
}
