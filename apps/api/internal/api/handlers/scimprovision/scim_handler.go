package scimprovision

import (
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	scimapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/scim"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/scim"
	"github.com/gin-gonic/gin"
)

type SCIMHandler struct {
	scimSvc *scimapp.SCIMService
}

func NewSCIMHandler(scimSvc *scimapp.SCIMService) *SCIMHandler {
	return &SCIMHandler{scimSvc: scimSvc}
}

func (h *SCIMHandler) RegisterRoutes(rg *gin.RouterGroup) {
	scimGroup := rg.Group("/scim/v2")
	scimGroup.Use(middleware.RequirePermission("operator:write"))
	{
		scimGroup.GET("/Users", h.ListUsers)
		scimGroup.POST("/Users", h.CreateUser)
		scimGroup.GET("/Users/:id", h.GetUser)
		scimGroup.PUT("/Users/:id", h.UpdateUser)
		scimGroup.DELETE("/Users/:id", h.DeleteUser)
	}
}

func (h *SCIMHandler) ListUsers(c *gin.Context) {
	startIndex, _ := strconv.Atoi(c.DefaultQuery("startIndex", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "100"))

	users, err := h.scimSvc.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"schemas":     []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		"totalResults": len(users),
		"startIndex":  startIndex,
		"itemsPerPage": count,
		"Resources":   users,
	})
}

func (h *SCIMHandler) CreateUser(c *gin.Context) {
	var user scim.SCIMUser
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	op, err := h.scimSvc.ProvisionUser(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": err.Error()})
		return
	}

	_ = op

	c.JSON(http.StatusCreated, gin.H{
		"schemas": []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"id":      user.ID,
		"status":  "created",
	})
}

func (h *SCIMHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.scimSvc.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *SCIMHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var user scim.SCIMUser
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	user.ID = id
	_, err := h.scimSvc.ProvisionUser(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"schemas": []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"id":      id,
		"status":  "updated",
	})
}

func (h *SCIMHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.scimSvc.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
