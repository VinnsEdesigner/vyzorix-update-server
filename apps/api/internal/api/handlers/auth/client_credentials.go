package auth

import (
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/client"

	"github.com/gin-gonic/gin"
)

// ClientCredentialsHandler handles client credentials endpoints.
type ClientCredentialsHandler struct {
	authService    *auth.AuthService
	clientService  *client.Service
}

// NewClientCredentialsHandler creates a new ClientCredentialsHandler.
func NewClientCredentialsHandler(authService *auth.AuthService, clientService *client.Service) *ClientCredentialsHandler {
	return &ClientCredentialsHandler{
		authService:   authService,
		clientService: clientService,
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "an error occurred"})
		return
	}

	var req struct {
		Name           string   `json:"name" binding:"required"`
		Platform       string   `json:"platform" binding:"required,oneof=web ios android"`
		AllowedOrigins []string `json:"allowedOrigins"`
		AllowedPaths   []string `json:"allowedPaths"`
		RateLimit     int      `json:"rateLimit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request body"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create client credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clientId":     clientResp.ID,
		"clientSecret": secret, // Only returned once!
		"platform":     clientResp.Platform,
		"name":         clientResp.Name,
		"createdAt":   clientResp.CreatedAt,
	})
}

// List handles GET /v1/auth/client-credentials.
func (h *ClientCredentialsHandler) List(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "an error occurred"})
		return
	}

	clients, err := h.clientService.ListByOperatorID(c.Request.Context(), operatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to list clients"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clients": clients})
}

// Get handles GET /v1/auth/client-credentials/:clientId.
func (h *ClientCredentialsHandler) Get(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "an error occurred"})
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	clientResp, err := h.clientService.GetByOperatorID(c.Request.Context(), clientID, operatorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Client not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"client": clientResp})
}

// Delete handles DELETE /v1/auth/client-credentials/:clientId.
func (h *ClientCredentialsHandler) Delete(c *gin.Context) {
	operatorID, err := h.getOperatorFromSession(c)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "not authenticated"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "an error occurred"})
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	// Verify ownership first
	_, err = h.clientService.GetByOperatorID(c.Request.Context(), clientID, operatorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Client not found"})
		return
	}

	if err := h.clientService.Deactivate(c.Request.Context(), clientID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to revoke client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "clientId": clientID})
}
