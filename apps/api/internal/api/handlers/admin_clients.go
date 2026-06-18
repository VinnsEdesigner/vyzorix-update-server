// Package controllers provides HTTP handlers.
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
)

// AdminClientResponse represents an API client for admin listing.
type AdminClientResponse struct {
	ID             string   `json:"id"`
	OperatorID     string   `json:"operatorId"`
	Name           string   `json:"name"`
	Platform       string   `json:"platform"`
	AllowedOrigins []string `json:"allowedOrigins"`
	AllowedPaths   []string `json:"allowedPaths"`
	RateLimit      int      `json:"rateLimit"`
	IsActive       bool     `json:"isActive"`
	RequestCount   int64    `json:"requestCount"`
	LastRequestAt *int64   `json:"lastRequestAt"`
	CreatedAt      int64    `json:"createdAt"`
	UpdatedAt      int64    `json:"updatedAt"`
}

// ListAllClients returns all API clients (admin only).
func (ac *AuthController) ListAllClients(c *gin.Context) {
	// Check admin role.
	operator := GetOperatorFromContext(c)
	if operator == nil || operator.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Admin access required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get all clients from storage.
	// Note: This requires a new method in storage to list all clients.
	clients, err := ac.store.ListAllAPIClients(ctx)
	if err != nil {
		ac.log.Error("failed to list clients", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to list clients"})
		return
	}

	result := make([]AdminClientResponse, 0, len(clients))
	for _, client := range clients {
		result = append(result, AdminClientResponse{
			ID:             client.ID,
			OperatorID:     client.OperatorID,
			Name:           client.Name,
			Platform:       client.Platform,
			AllowedOrigins: client.AllowedOrigins,
			AllowedPaths:   client.AllowedPaths,
			RateLimit:      client.RateLimit,
			IsActive:       client.IsActive,
			RequestCount:   client.RequestCount,
			LastRequestAt: client.LastRequestAt,
			CreatedAt:     client.CreatedAt,
			UpdatedAt:     client.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"clients": result})
}

// GetClient returns details for a specific client (admin only).
func (ac *AuthController) GetClient(c *gin.Context) {
	// Check admin role.
	operator := GetOperatorFromContext(c)
	if operator == nil || operator.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Admin access required"})
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

	c.JSON(http.StatusOK, gin.H{"client": AdminClientResponse{
		ID:             client.ID,
		OperatorID:     client.OperatorID,
		Name:           client.Name,
		Platform:       client.Platform,
		AllowedOrigins: client.AllowedOrigins,
		AllowedPaths:   client.AllowedPaths,
		RateLimit:      client.RateLimit,
		IsActive:       client.IsActive,
		RequestCount:   client.RequestCount,
		LastRequestAt: client.LastRequestAt,
		CreatedAt:     client.CreatedAt,
		UpdatedAt:     client.UpdatedAt,
	}})
}

// UpdateClient updates a client's settings (admin only).
func (ac *AuthController) UpdateClient(c *gin.Context) {
	// Check admin role.
	operator := GetOperatorFromContext(c)
	if operator == nil || operator.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Admin access required"})
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	var req struct {
		Name           *string  `json:"name"`
		AllowedOrigins []string `json:"allowedOrigins"`
		AllowedPaths   []string `json:"allowedPaths"`
		RateLimit     *int     `json:"rateLimit"`
		IsActive      *bool    `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request body"})
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

	// Update fields.
	if req.Name != nil {
		client.Name = *req.Name
	}
	if req.AllowedOrigins != nil {
		client.AllowedOrigins = req.AllowedOrigins
	}
	if req.AllowedPaths != nil {
		client.AllowedPaths = req.AllowedPaths
	}
	if req.RateLimit != nil {
		client.RateLimit = *req.RateLimit
	}
	if req.IsActive != nil {
		client.IsActive = *req.IsActive
	}

	// Save updates.
	if err := ac.store.UpdateAPIClient(ctx, client); err != nil {
		ac.log.Error("failed to update client", "error", err, "clientId", clientID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to update client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "client": client})
}

// DeleteClient deletes a client (admin only).
func (ac *AuthController) DeleteClient(c *gin.Context) {
	// Check admin role.
	operator := GetOperatorFromContext(c)
	if operator == nil || operator.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Admin access required"})
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := ac.store.DeleteAPIClient(ctx, clientID); err != nil {
		ac.log.Error("failed to delete client", "error", err, "clientId", clientID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to delete client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "clientId": clientID})
}

// RotateClientKey rotates the signing key for a client (admin only).
func (ac *AuthController) RotateClientKey(c *gin.Context) {
	// Check admin role.
	operator := GetOperatorFromContext(c)
	if operator == nil || operator.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Admin access required"})
		return
	}

	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "clientId is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Verify client exists.
	_, err := ac.store.GetAPIClient(ctx, clientID)
	if err != nil {
		if err == storage.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Client not found"})
			return
		}
		ac.log.Error("failed to get client", "error", err, "clientId", clientID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to get client"})
		return
	}

	// Rotate the key with 24-hour grace period.
	key, _, err := ac.store.RotateSigningKey(ctx, clientID, 24*time.Hour)
	if err != nil {
		ac.log.Error("failed to rotate key", "error", err, "clientId", clientID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to rotate key"})
		return
	}

	_ = key // Key is rotated, client should fetch new credentials

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Key rotated. Client must fetch new credentials.",
		"clientId":   clientID,
		"keyVersion": key.Version,
	})
}