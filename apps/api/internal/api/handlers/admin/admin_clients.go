package admin

import (
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.AdminClientListResult
	_ openapi.AdminClientResult
	_ openapi.UpdateAdminClientRequest
	_ openapi.SuccessResult
	_ openapi.ErrorResponse
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
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "unauthorized"))
		return false
	}

	// Require org context for admin access.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Organization ID required"))
		return false
	}

	// Check if operator is super_admin in this specific organization.
	if !op.IsSuperAdminIn(orgID) {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthzInsufficientPermissions, "admin access required"))
		return false
	}

	return true
}

// List handles GET /v1/admin/clients.
// @Summary      List admin clients
// @Description  Returns all OAuth clients (super_admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.AdminClientListResult  "clients"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /admin/clients [get]
func (h *ClientsHandler) List(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	clients, total, err := h.clientService.ListAll(c.Request.Context(), 20, 0)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to list clients"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"clients": clients, "total": total})
}

// Get handles GET /v1/admin/clients/:clientId.
// @Summary      Get admin client
// @Description  Returns a single OAuth client by ID (super_admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        clientId  path  string  true  "client ID"
// @Success      200  {object}  openapi.AdminClientResult  "client"
// @Failure      400  {object}  openapi.ErrorResponse  "clientId required"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      404  {object}  openapi.ErrorResponse  "client not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /admin/clients/{clientId} [get]
func (h *ClientsHandler) Get(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "clientId is required"))
		return
	}

	clientResp, err := h.clientService.Get(c.Request.Context(), clientID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Client not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"client": clientResp})
}

// Update handles PATCH /v1/admin/clients/:clientId.
// @Summary      Update admin client
// @Description  Updates mutable client fields (super_admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        clientId  path  string  true  "client ID"
// @Param        body  body  openapi.UpdateAdminClientRequest  true  "client update"
// @Success      200  {object}  openapi.AdminClientResult  "updated client"
// @Failure      400  {object}  openapi.ErrorResponse  "clientId required / invalid body"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /admin/clients/{clientId} [patch]
func (h *ClientsHandler) Update(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "clientId is required"))
		return
	}

	var req dto.UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid request body"))
		return
	}

	clientResp, err := h.clientService.Update(c.Request.Context(), clientID, &req)
	if err != nil {
		if errors.Is(err, application.ErrClientNotFound) {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Client not found"))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to update client"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"client": clientResp})
}

// Delete handles DELETE /v1/admin/clients/:clientId.
// @Summary      Delete admin client
// @Description  Revokes a client (super_admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        clientId  path  string  true  "client ID"
// @Success      200  {object}  openapi.SuccessResult  "client deleted"
// @Failure      400  {object}  openapi.ErrorResponse  "clientId required"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /admin/clients/{clientId} [delete]
func (h *ClientsHandler) Delete(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "clientId is required"))
		return
	}

	// Verify existence so deleting a nonexistent client returns 404, not 200.
	if _, err := h.clientService.Get(c.Request.Context(), clientID); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Client not found"))
		return
	}

	if err := h.clientService.Delete(c.Request.Context(), clientID); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to delete client"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "clientId": clientID})
}

// RotateKey handles POST /v1/admin/clients/:clientId/rotate-key.
// @Summary      Rotate admin client key
// @Description  Rotates the signing key with a 24-hour grace period (super_admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        clientId  path  string  true  "client ID"
// @Success      200  {object}  openapi.AdminClientResult  "client with rotated key"
// @Failure      400  {object}  openapi.ErrorResponse  "clientId required"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      404  {object}  openapi.ErrorResponse  "client not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /admin/clients/{clientId}/rotate-key [post]
func (h *ClientsHandler) RotateKey(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "clientId is required"))
		return
	}

	// Verify existence so rotating a nonexistent client returns 404, not 500.
	if _, err := h.clientService.Get(c.Request.Context(), clientID); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Client not found"))
		return
	}

	keyVersion, err := h.clientService.RotateKey(c.Request.Context(), clientID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to rotate key"))
		return
	}
	_ = keyVersion
	_ = c.Error(apperrors.NewServerErrorFromStatus(http.StatusOK, "Key rotated. Client must fetch new credentials."))
}
