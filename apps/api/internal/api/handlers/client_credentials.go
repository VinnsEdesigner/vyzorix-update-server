// Package controllers provides HTTP handlers.
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
)

// ClientCredentialsResponse is the response for client credentials request.
type ClientCredentialsResponse struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"` // Only shown once!
	Platform     string `json:"platform"`
	Name         string `json:"name"`
	CreatedAt    int64  `json:"createdAt"`
}

// GetClientCredentials returns the client credentials for the authenticated operator.
func (ac *AuthController) GetClientCredentials(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authentication required"})
		return
	}

	var req struct {
		Name     string   `json:"name" binding:"required"`
		Platform string   `json:"platform" binding:"required,oneof=web ios android"`
		Origins  []string `json:"allowedOrigins"`
		Paths    []string `json:"allowedPaths"`
		RateLimit int     `json:"rateLimit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request body"})
		return
	}

	// Set defaults.
	if req.RateLimit == 0 {
		req.RateLimit = 100
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	client, secret, err := ac.store.CreateAPIClient(ctx, storage.CreateAPIClientRequest{
		OperatorID:     operator.ID,
		Name:           req.Name,
		Platform:       req.Platform,
		AllowedOrigins:  req.Origins,
		AllowedPaths:   req.Paths,
		RateLimit:      req.RateLimit,
	})
	if err != nil {
		ac.log.Error("failed to create client credentials", "error", err, "operatorId", operator.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create client credentials"})
		return
	}

	c.JSON(http.StatusOK, ClientCredentialsResponse{
		ClientID:     client.ID,
		ClientSecret: secret, // Only returned once!
		Platform:     client.Platform,
		Name:         client.Name,
		CreatedAt:    client.CreatedAt,
	})
}

// ListClientCredentials returns all client credentials for the authenticated operator.
func (ac *AuthController) ListClientCredentials(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authentication required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	clients, err := ac.store.GetAPIClientByOperator(ctx, operator.ID)
	if err != nil {
		ac.log.Error("failed to list clients", "error", err, "operatorId", operator.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to list clients"})
		return
	}

	// Don't expose secret hashes.
	type ClientInfo struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		Platform       string   `json:"platform"`
		AllowedOrigins []string `json:"allowedOrigins"`
		AllowedPaths   []string `json:"allowedPaths"`
		RateLimit      int      `json:"rateLimit"`
		IsActive       bool     `json:"isActive"`
		RequestCount   int64    `json:"requestCount"`
		LastRequestAt  *int64   `json:"lastRequestAt"`
		CreatedAt      int64    `json:"createdAt"`
	}

	result := make([]ClientInfo, 0, len(clients))
	for _, client := range clients {
		result = append(result, ClientInfo{
			ID:             client.ID,
			Name:           client.Name,
			Platform:       client.Platform,
			AllowedOrigins: client.AllowedOrigins,
			AllowedPaths:   client.AllowedPaths,
			RateLimit:      client.RateLimit,
			IsActive:       client.IsActive,
			RequestCount:   client.RequestCount,
			LastRequestAt:  client.LastRequestAt,
			CreatedAt:      client.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"clients": result})
}

// RevokeClientCredentials deactivates a client credential.
func (ac *AuthController) RevokeClientCredentials(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authentication required"})
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Verify the client belongs to this operator.
	client, err := ac.store.GetAPIClient(ctx, clientID)
	if err != nil {
		if err == storage.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Client not found"})
			return
		}
		ac.log.Error("failed to get client", "error", err, "clientId", clientID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to revoke client"})
		return
	}

	if client.OperatorID != operator.ID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Client not found"})
		return
	}

	// Deactivate the client.
	if err := ac.store.UpdateAPIClientActive(ctx, clientID, false); err != nil {
		ac.log.Error("failed to deactivate client", "error", err, "clientId", clientID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to revoke client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "clientId": clientID})
}

// GetClientCredentialsDetail returns details for a specific client.
func (ac *AuthController) GetClientCredentialsDetail(c *gin.Context) {
	operator := GetOperatorFromContext(c)
	if operator == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authentication required"})
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	client, err := ac.store.GetAPIClient(ctx, clientID)
	if err != nil {
		if err == storage.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Client not found"})
			return
		}
		ac.log.Error("failed to get client", "error", err, "clientId", clientID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to get client"})
		return
	}

	if client.OperatorID != operator.ID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Client not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client": gin.H{
			"id":              client.ID,
			"name":            client.Name,
			"platform":        client.Platform,
			"allowedOrigins":   client.AllowedOrigins,
			"allowedPaths":    client.AllowedPaths,
			"rateLimit":       client.RateLimit,
			"isActive":        client.IsActive,
			"requestCount":    client.RequestCount,
			"lastRequestAt":   client.LastRequestAt,
			"createdAt":       client.CreatedAt,
			"updatedAt":       client.UpdatedAt,
		},
	})
}