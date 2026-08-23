// Package notification provides HTTP handlers for org-scoped contact points.
package notification

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	appnotification "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/notifications"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	notificationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.ContactPoint
	_ openapi.ContactPointRequest
	_ openapi.ContactPointListResult
	_ openapi.ContactPointTestResult
	_ openapi.DeletedResult
	_ openapi.ErrorResponse
)

// Handler processes contact point CRUD and test deliveries.
type Handler struct {
	service    *appnotification.Service
	dispatcher *appnotification.Dispatcher
}

// NewHandler creates a new notification Handler.
func NewHandler(service *appnotification.Service, dispatcher *appnotification.Dispatcher) *Handler {
	return &Handler{service: service, dispatcher: dispatcher}
}

// Service returns the underlying application service (GraphQL wiring).
func (h *Handler) Service() *appnotification.Service { return h.service }

// Dispatcher returns the underlying dispatcher (GraphQL wiring).
func (h *Handler) Dispatcher() *appnotification.Dispatcher { return h.dispatcher }

type contactPointRequest struct {
	Name       string            `json:"name"`
	Channel    string            `json:"channel"`
	Secret     string            `json:"secret"`
	Config     map[string]string `json:"config"`
	TemplateID string            `json:"template_id"`
	Enabled    bool              `json:"enabled"`
}

func (r *contactPointRequest) toInput(orgID string) *appnotification.ContactPointInput {
	return &appnotification.ContactPointInput{
		OrgID:      orgID,
		Name:       r.Name,
		Channel:    notificationdomain.ChannelType(r.Channel),
		Secret:     r.Secret,
		Config:     r.Config,
		TemplateID: r.TemplateID,
		Enabled:    r.Enabled,
	}
}

func contactPointJSON(cp *notificationdomain.ContactPoint) gin.H {
	return gin.H{
		"id":          cp.ID,
		"org_id":      cp.OrgID,
		"name":        cp.Name,
		"channel":     string(cp.Channel),
		"secret":      cp.Secret != "",
		"config":      cp.Config,
		"template_id": cp.TemplateID,
		"enabled":     cp.Enabled,
		"created_at":  cp.CreatedAt,
		"updated_at":  cp.UpdatedAt,
	}
}

// List handles GET /v1/notifications/contact-points.
// @Summary      List contact points
// @Description  Returns all org-scoped contact points
// @Tags         contact-points
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.ContactPointListResult  "contact points"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /notifications/contact-points [get]
func (h *Handler) List(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	points, err := h.service.List(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to list contact points"))
		return
	}
	items := make([]gin.H, 0, len(points))
	for _, cp := range points {
		items = append(items, contactPointJSON(cp))
	}
	c.JSON(http.StatusOK, gin.H{"contact_points": items})
}

// Create handles POST /v1/notifications/contact-points.
// @Summary      Create contact point
// @Description  Creates a new org-scoped contact point
// @Tags         contact-points
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        body  body  openapi.ContactPointRequest  true  "contact point definition"
// @Success      201  {object}  openapi.ContactPoint  "created contact point"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /notifications/contact-points [post]
func (h *Handler) Create(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	var req contactPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body: "+err.Error()))
		return
	}
	cp, err := h.service.Create(c.Request.Context(), req.toInput(orgID))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, contactPointJSON(cp))
}

// Get handles GET /v1/notifications/contact-points/:id.
// @Summary      Get contact point
// @Description  Returns one contact point by ID
// @Tags         contact-points
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "contact point ID"
// @Success      200  {object}  openapi.ContactPoint  "contact point"
// @Failure      400  {object}  openapi.ErrorResponse  "not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /notifications/contact-points/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	cp, err := h.service.Get(c.Request.Context(), orgID, c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, contactPointJSON(cp))
}

// Update handles PATCH /v1/notifications/contact-points/:id.
// @Summary      Update contact point
// @Description  Replaces a contact point's mutable fields
// @Tags         contact-points
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "contact point ID"
// @Param        body  body  openapi.ContactPointRequest  true  "contact point definition"
// @Success      200  {object}  openapi.ContactPoint  "updated contact point"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input / not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /notifications/contact-points/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	var req contactPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body: "+err.Error()))
		return
	}
	cp, err := h.service.Update(c.Request.Context(), orgID, c.Param("id"), req.toInput(orgID))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, contactPointJSON(cp))
}

// Delete handles DELETE /v1/notifications/contact-points/:id.
// @Summary      Delete contact point
// @Description  Removes a contact point
// @Tags         contact-points
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "contact point ID"
// @Success      200  {object}  openapi.DeletedResult  "deleted confirmation"
// @Failure      400  {object}  openapi.ErrorResponse  "not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /notifications/contact-points/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if err := h.service.Delete(c.Request.Context(), orgID, c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// Test handles POST /v1/notifications/contact-points/:id/test.
// Sends a one-off test notification through the contact point.
// @Summary      Test contact point
// @Description  Sends a one-off test notification through the contact point
// @Tags         contact-points
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "contact point ID"
// @Success      200  {object}  openapi.ContactPointTestResult  "test result"
// @Failure      400  {object}  openapi.ErrorResponse  "not found"
// @Failure      500  {object}  openapi.ErrorResponse  "delivery failed"
// @Router       /notifications/contact-points/{id}/test [post]
func (h *Handler) Test(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	cp, err := h.service.Get(c.Request.Context(), orgID, c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	msg := &notificationdomain.Message{
		Subject: "Vyzorix contact point test",
		Body:    "This is a test notification from Vyzorix.",
		Event:   "test",
		Data:    map[string]string{"contactPointId": cp.ID, "channel": string(cp.Channel)},
	}
	if err := h.dispatcher.SendToPoint(c.Request.Context(), cp, msg); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "test delivery failed: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"sent": true, "tested_at": time.Now()})
}

// Deliveries handles GET /v1/notifications/contact-points/:id/deliveries.
func (h *Handler) Deliveries(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if _, err := h.service.Get(c.Request.Context(), orgID, c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	// DeliveriesRepository is not exposed through Service yet; return empty list
	// for now until the service gains a ListDeliveries method. The endpoint is
	// reserved for the delivery log UI.
	c.JSON(http.StatusOK, gin.H{"deliveries": []any{}})
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, appnotification.ErrContactPointNotFound):
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "contact point not found"))
	case errors.Is(err, appnotification.ErrInvalidContactPoint):
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, err.Error()))
	default:
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "contact point operation failed"))
	}
}
