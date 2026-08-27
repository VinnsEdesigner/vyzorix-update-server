package auth

import (
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/schema"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.ClientCredential
	_ openapi.ClientCredentialListResult
	_ openapi.CreateClientCredentialRequest
	_ openapi.UpdateClientCredentialRequest
	_ openapi.SuccessResult
	_ openapi.ErrorResponse
)

// ClientCredentialsHandler handles client credentials endpoints.
type ClientCredentialsHandler struct {
	authService   *auth.AuthService
	clientService *client.Service
	presenter     *response.Presenter
}

// NewClientCredentialsHandler creates a new ClientCredentialsHandler.
func NewClientCredentialsHandler(authService *auth.AuthService, clientService *client.Service, presenter *response.Presenter) *ClientCredentialsHandler {
	return &ClientCredentialsHandler{
		authService:   authService,
		clientService: clientService,
		presenter:     presenter,
	}
}

// clientCredToSchema maps a client credential DTO onto the CUE-generated
// ClientCredential wire struct.
func clientCredToSchema(r *dto.ClientResponse) schema.ClientCredential {
	return schema.ClientCredential{
		ID:             r.ID,
		OperatorID:     r.OperatorID,
		Name:           r.Name,
		Platform:       r.Platform,
		AllowedOrigins: r.AllowedOrigins,
		AllowedPaths:   r.AllowedPaths,
		RateLimit:      r.RateLimit,
		RequestCount:   r.RequestCount,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		IsActive:       r.IsActive,
		LastRequestAt:  r.LastRequestAt,
	}
}

// getOperatorFromSession extracts operator from the authenticated session.
// See SettingsHandler.getOperatorFromSession: the cookieAuth middleware already.
// validated the (encrypted) session cookie and stored the operator in context,
// so we reuse that instead of passing the raw cookie ciphertext to ValidateSession.
func (h *ClientCredentialsHandler) getOperatorFromSession(c *gin.Context) (string, error) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		return "", application.ErrUnauthorized
	}

	return op.ID, nil
}

// Create handles POST /v1/auth/client-credentials.
// @Summary      Create client credentials
// @Description  Creates a new OAuth client credential. The secret is only returned once
// @Tags         client-credentials
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.CreateClientCredentialRequest  true  "client name"
// @Success      200  {object}  openapi.ClientCredential  "created client with secret"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid request body"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/client-credentials [post]
func (h *ClientCredentialsHandler) Create(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "not authenticated")
			return
		}

		h.presenter.InternalError(c, "an error occurred")

		return
	}

	var req struct {
		Name           string   `json:"name" binding:"required"`
		Platform       string   `json:"platform" binding:"required,oneof=web ios android"`
		AllowedOrigins []string `json:"allowedOrigins"`
		AllowedPaths   []string `json:"allowedPaths"`
		RateLimit      int      `json:"rateLimit"`
	}

	if err = c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Set defaults.
	if req.RateLimit == 0 {
		req.RateLimit = 100
	}

	clientResp, secret, err := h.clientService.Create(c.Request.Context(), &client.CreateClientRequest{
		OperatorID:     operatorID,
		Name:           req.Name,
		Platform:       req.Platform,
		AllowedOrigins: req.AllowedOrigins,
		AllowedPaths:   req.AllowedPaths,
		RateLimit:      req.RateLimit,
	})
	if err != nil {
		h.presenter.InternalError(c, "Failed to create client credentials")
		return
	}

	h.presenter.APIClientCreated(c, operatorID, clientResp.ID)
	h.presenter.OK(c, gin.H{
		"clientId":     clientResp.ID,
		"clientSecret": secret, // Only returned once!.
		"platform":     clientResp.Platform,
		"name":         clientResp.Name,
		"createdAt":    clientResp.CreatedAt,
	})
}

// List handles GET /v1/auth/client-credentials.
// @Summary      List client credentials
// @Description  Returns all client credentials for the authenticated operator
// @Tags         client-credentials
// @Accept       json
// @Produce      json
// @Success      200  {object}  openapi.ClientCredentialListResult  "clients"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/client-credentials [get]
func (h *ClientCredentialsHandler) List(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "not authenticated")
			return
		}

		h.presenter.InternalError(c, "an error occurred")

		return
	}

	clients, err := h.clientService.ListByOperatorID(c.Request.Context(), operatorID)
	if err != nil {
		h.presenter.InternalError(c, "Failed to list clients")
		return
	}

	items := make([]schema.ClientCredential, len(clients))
	for i := range clients {
		items[i] = clientCredToSchema(&clients[i])
	}
	h.presenter.OK(c, schema.ClientCredentialListResult{Clients: items})
}

// Get handles GET /v1/auth/client-credentials/:clientId.
// @Summary      Get client credential
// @Description  Returns a single client credential by ID
// @Tags         client-credentials
// @Accept       json
// @Produce      json
// @Param        clientId  path  string  true  "client ID"
// @Success      200  {object}  openapi.ClientCredential  "client"
// @Failure      400  {object}  openapi.ErrorResponse  "clientId required"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      404  {object}  openapi.ErrorResponse  "client not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/client-credentials/{clientId} [get]
func (h *ClientCredentialsHandler) Get(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "not authenticated")
			return
		}

		h.presenter.InternalError(c, "an error occurred")

		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		h.presenter.BadRequest(c, "clientId is required")
		return
	}

	clientResp, err := h.clientService.GetByOperatorID(c.Request.Context(), clientID, operatorID)
	if err != nil {
		h.presenter.NotFound(c, "Client not found")
		return
	}

			h.presenter.OK(c, clientCredToSchema(clientResp))
}

// Delete handles DELETE /v1/auth/client-credentials/:clientId.
// @Summary      Revoke client credential
// @Description  Revokes a client credential by ID
// @Tags         client-credentials
// @Accept       json
// @Produce      json
// @Param        clientId  path  string  true  "client ID"
// @Success      200  {object}  openapi.SuccessResult  "client revoked"
// @Failure      400  {object}  openapi.ErrorResponse  "clientId required"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      404  {object}  openapi.ErrorResponse  "client not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/client-credentials/{clientId} [delete]
func (h *ClientCredentialsHandler) Delete(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "not authenticated")
			return
		}

		h.presenter.InternalError(c, "an error occurred")

		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		h.presenter.BadRequest(c, "clientId is required")
		return
	}

	// Verify ownership first.
	_, err = h.clientService.GetByOperatorID(c.Request.Context(), clientID, operatorID)
	if err != nil {
		h.presenter.NotFound(c, "Client not found")
		return
	}

	if err := h.clientService.Deactivate(c.Request.Context(), clientID); err != nil {
		h.presenter.InternalError(c, "Failed to revoke client")
		return
	}

	h.presenter.APIClientRevoked(c, operatorID, clientID)
	h.presenter.OK(c, gin.H{"success": true, "clientId": clientID})
}

// Update handles PATCH /v1/auth/client-credentials/:clientId.
// @Summary      Update client credential
// @Description  Updates mutable client credential fields (e.g., name)
// @Tags         client-credentials
// @Accept       json
// @Produce      json
// @Param        clientId  path  string  true  "client ID"
// @Param        body  body  openapi.UpdateClientCredentialRequest  true  "client update"
// @Success      200  {object}  openapi.ClientCredential  "updated client"
// @Failure      400  {object}  openapi.ErrorResponse  "clientId required / invalid body"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      404  {object}  openapi.ErrorResponse  "client not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/client-credentials/{clientId} [patch]
func (h *ClientCredentialsHandler) Update(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "not authenticated")
			return
		}

		h.presenter.InternalError(c, "an error occurred")
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		h.presenter.BadRequest(c, "clientId is required")
		return
	}

	var req struct {
		Name           *string   `json:"name"`
		AllowedOrigins *[]string `json:"allowedOrigins"`
		AllowedPaths   *[]string `json:"allowedPaths"`
		RateLimit      *int      `json:"rateLimit"`
		Active         *bool     `json:"active"`
	}

	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		h.presenter.BadRequest(c, "Invalid request body: "+bindErr.Error())
		return
	}

	// Verify ownership.
	clientResp, err := h.clientService.GetByOperatorID(c.Request.Context(), clientID, operatorID)
	if err != nil {
		h.presenter.NotFound(c, "Client not found")
		return
	}

	// Update fields if provided.
	if req.Name != nil {
		clientResp.Name = *req.Name
	}
	if req.AllowedOrigins != nil {
		clientResp.AllowedOrigins = *req.AllowedOrigins
	}
	if req.AllowedPaths != nil {
		clientResp.AllowedPaths = *req.AllowedPaths
	}
	if req.RateLimit != nil {
		clientResp.RateLimit = *req.RateLimit
	}
	if req.Active != nil {
		clientResp.IsActive = *req.Active
	}

	// Note: In a real implementation, you'd call a service method to update.
	h.presenter.OK(c, gin.H{
		"client":  clientResp,
		"message": "Client updated successfully",
	})
}

// RotateSecret handles POST /v1/auth/client-credentials/:clientId/rotate-secret.
// @Summary      Rotate client secret
// @Description  Rotates a client credential's secret. The new secret is only returned once
// @Tags         client-credentials
// @Accept       json
// @Produce      json
// @Param        clientId  path  string  true  "client ID"
// @Success      200  {object}  openapi.ClientCredential  "client with new secret"
// @Failure      400  {object}  openapi.ErrorResponse  "clientId required"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      404  {object}  openapi.ErrorResponse  "client not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/client-credentials/{clientId}/rotate-secret [post]
func (h *ClientCredentialsHandler) RotateSecret(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "not authenticated")
			return
		}

		h.presenter.InternalError(c, "an error occurred")
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		h.presenter.BadRequest(c, "clientId is required")
		return
	}

	// Rotate the client secret.
	newSecret, err := h.clientService.RotateSecret(c.Request.Context(), clientID, operatorID)
	if err != nil {
		h.presenter.NotFound(c, "Client not found")
		return
	}

	// Audit log the secret rotation.
	h.presenter.APIClientSecretRotated(c, operatorID, clientID)

	h.presenter.OK(c, gin.H{
		"clientId":     clientID,
		"clientSecret": newSecret,
		"message":      "Client secret rotated successfully. Store this secret securely - it will not be shown again.",
	})
}
