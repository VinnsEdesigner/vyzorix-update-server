package auth

import (
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"

	"github.com/gin-gonic/gin"
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

// getOperatorFromSession extracts operator from session.
func (h *ClientCredentialsHandler) getOperatorFromSession(c *gin.Context) (string, error) {
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		return "", err
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		return "", err
	}

	return op.ID, nil
}

// Create handles POST /v1/auth/client-credentials.
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
		h.presenter.BadRequest(c, "Invalid request body")
		return
	}

	// Set defaults
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
		"clientSecret": secret, // Only returned once!
		"platform":     clientResp.Platform,
		"name":         clientResp.Name,
		"createdAt":    clientResp.CreatedAt,
	})
}

// List handles GET /v1/auth/client-credentials.
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

	h.presenter.OK(c, gin.H{"clients": clients})
}

// Get handles GET /v1/auth/client-credentials/:clientId.
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

	h.presenter.OK(c, gin.H{"client": clientResp})
}

// Delete handles DELETE /v1/auth/client-credentials/:clientId.
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

	// Verify ownership first
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
