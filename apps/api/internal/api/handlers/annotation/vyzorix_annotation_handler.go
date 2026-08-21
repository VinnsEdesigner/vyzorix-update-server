// Package annotation provides HTTP handlers for org-scoped annotations.
package annotation

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	appannotation "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/annotation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/annotation"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
)

// Handler processes annotation CRUD requests.
type Handler struct {
	service *appannotation.Service
}

// NewHandler creates a new annotation Handler.
func NewHandler(service *appannotation.Service) *Handler {
	return &Handler{service: service}
}

// Service returns the underlying application service (GraphQL wiring).
func (h *Handler) Service() *appannotation.Service { return h.service }

type annotationRequest struct {
	Title     string   `json:"title"`
	Text      string   `json:"text"`
	Source    string   `json:"source"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	Tags      []string `json:"tags"`
}

func (r *annotationRequest) toInput(orgID string) *appannotation.AnnotationInput {
	in := &appannotation.AnnotationInput{
		OrgID:  orgID,
		Title:  r.Title,
		Text:   r.Text,
		Tags:   r.Tags,
		Source: r.Source,
	}
	if t, err := time.Parse(time.RFC3339, r.StartTime); err == nil {
		in.StartTime = t
	}
	if r.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, r.EndTime); err == nil {
			in.EndTime = &t
		}
	}
	return in
}

func annotationJSON(a *annotation.Annotation) gin.H {
	return gin.H{
		"id":         a.ID,
		"org_id":     a.OrgID,
		"title":      a.Title,
		"text":       a.Text,
		"tags":       a.Tags,
		"source":     a.Source,
		"start_time": a.StartTime,
		"end_time":   a.EndTime,
		"created_at": a.CreatedAt,
		"updated_at": a.UpdatedAt,
	}
}

// List handles GET /v1/annotations.
func (h *Handler) List(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	f := &annotation.Filter{OrgID: orgID, Limit: 200}
	if tag := c.Query("tag"); tag != "" {
		f.Tag = tag
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			f.Limit = n
		}
	}
	if v := c.Query("start_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.StartTime = t
		}
	}
	if v := c.Query("end_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.EndTime = t
		}
	}

	items, err := h.service.List(c.Request.Context(), f)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to list annotations"))
		return
	}
	result := make([]gin.H, 0, len(items))
	for _, a := range items {
		result = append(result, annotationJSON(a))
	}
	c.JSON(http.StatusOK, gin.H{"annotations": result})
}

// Create handles POST /v1/annotations.
func (h *Handler) Create(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	var req annotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body: "+err.Error()))
		return
	}
	a, err := h.service.Create(c.Request.Context(), req.toInput(orgID))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, annotationJSON(a))
}

// Get handles GET /v1/annotations/:id.
func (h *Handler) Get(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	a, err := h.service.Get(c.Request.Context(), orgID, c.Param("id"))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, annotationJSON(a))
}

// Update handles PATCH /v1/annotations/:id.
func (h *Handler) Update(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	var req annotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid request body: "+err.Error()))
		return
	}
	a, err := h.service.Update(c.Request.Context(), orgID, c.Param("id"), req.toInput(orgID))
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, annotationJSON(a))
}

// Delete handles DELETE /v1/annotations/:id.
func (h *Handler) Delete(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if err := h.service.Delete(c.Request.Context(), orgID, c.Param("id")); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, appannotation.ErrAnnotationNotFound):
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "annotation not found"))
	case errors.Is(err, appannotation.ErrInvalidAnnotation):
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, err.Error()))
	default:
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "annotation operation failed"))
	}
}
