package scimprovisioning

import (
	"encoding/json"
	"net/http"
	"strconv"

	scimapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/scim"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/scim"
	"github.com/gin-gonic/gin"
)

const (
	authHeader   = "Authorization"
	bearerPrefix = "Bearer "
)

// Handler handles SCIM provisioning requests.
type Handler struct {
	svc *scimapp.SCIMService
}

// NewHandler creates a new SCIM handler.
func NewHandler(svc *scimapp.SCIMService) *Handler {
	return &Handler{svc: svc}
}

// ListUsers handles GET /v2/Users
func (h *Handler) ListUsers(c *gin.Context) {
	if !h.authenticate(c) {
		return
	}

	startIndex, _ := strconv.Atoi(c.DefaultQuery("startIndex", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "100"))

	resp, err := h.svc.ListUsers(c.Request.Context(), startIndex, count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetUser handles GET /v2/Users/:id
func (h *Handler) GetUser(c *gin.Context) {
	if !h.authenticate(c) {
		return
	}

	id := c.Param("id")
	user, err := h.svc.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser handles POST /v2/Users
func (h *Handler) CreateUser(c *gin.Context) {
	if !h.authenticate(c) {
		return
	}

	var req scim.SCIMUser
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body", "message": err.Error()})
		return
	}

	user, err := h.svc.CreateUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "create_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser handles PUT /v2/Users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	if !h.authenticate(c) {
		return
	}

	id := c.Param("id")

	var req scim.SCIMUser
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body", "message": err.Error()})
		return
	}

	user, err := h.svc.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "update_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE /v2/Users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	if !h.authenticate(c) {
		return
	}

	id := c.Param("id")
	if err := h.svc.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) authenticate(c *gin.Context) bool {
	auth := c.GetHeader(authHeader)
	if auth == "" || len(auth) <= len(bearerPrefix) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Missing or invalid authorization header"})
		return false
	}

	token := auth[len(bearerPrefix):]
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Missing bearer token"})
		return false
	}

	scimToken := c.GetHeader("X-SCIM-Token")
	if scimToken == "" {
		scimToken = token
	}

	if !h.validateToken(scimToken) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid SCIM token"})
		return false
	}

	return true
}

func (h *Handler) validateToken(token string) bool {
	return token != "" && len(token) >= 8
}
