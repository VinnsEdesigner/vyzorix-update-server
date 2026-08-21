// Package configversion provides HTTP handlers for config version history
// and restore.
package configversion

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	appconfig "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/configversion"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/configversion"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
)

// Handler processes config version list/restore requests.
type Handler struct {
	versionSvc *appconfig.Service
	settingsSvc *organization.OrganizationSettingsService
}

// NewHandler creates a new config version Handler.
func NewHandler(versionSvc *appconfig.Service, settingsSvc *organization.OrganizationSettingsService) *Handler {
	return &Handler{versionSvc: versionSvc, settingsSvc: settingsSvc}
}

// Service returns the underlying versioning service (GraphQL wiring).
func (h *Handler) Service() *appconfig.Service { return h.versionSvc }

func versionJSON(v *configversion.ConfigVersion) gin.H {
	return gin.H{
		"id":            v.ID,
		"org_id":        v.OrgID,
		"resource_type": string(v.ResourceType),
		"version":       v.Version,
		"snapshot":      v.Snapshot,
		"changed_by":    v.ChangedBy,
		"created_at":    v.CreatedAt,
	}
}

// List handles GET /v1/config-versions/:resource.
func (h *Handler) List(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	resourceType := configversion.ResourceType(c.Param("resource"))
	if !resourceType.Valid() {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid resource type"))
		return
	}
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	versions, err := h.versionSvc.List(c.Request.Context(), orgID, resourceType, limit)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to list config versions"))
		return
	}
	items := make([]gin.H, 0, len(versions))
	for _, v := range versions {
		items = append(items, versionJSON(v))
	}
	c.JSON(http.StatusOK, gin.H{"versions": items})
}

// Get handles GET /v1/config-versions/:resource/:version.
func (h *Handler) Get(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	resourceType := configversion.ResourceType(c.Param("resource"))
	if !resourceType.Valid() {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid resource type"))
		return
	}
	version, err := strconv.ParseInt(c.Param("version"), 10, 64)
	if err != nil || version <= 0 {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "version must be positive integer"))
		return
	}
	v, err := h.versionSvc.Get(c.Request.Context(), orgID, resourceType, version)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, versionJSON(v))
}

// Restore handles POST /v1/config-versions/:resource/:version/restore.
// Re-applies the version's snapshot as the live settings.
func (h *Handler) Restore(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	resourceType := configversion.ResourceType(c.Param("resource"))
	if !resourceType.Valid() {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid resource type"))
		return
	}
	version, err := strconv.ParseInt(c.Param("version"), 10, 64)
	if err != nil || version <= 0 {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "version must be positive integer"))
		return
	}
	v, err := h.versionSvc.Get(c.Request.Context(), orgID, resourceType, version)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	restored, err := h.settingsSvc.RestoreSettings(c.Request.Context(), orgID, v.Snapshot)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "restore failed: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"restored_to_version": version, "settings": restored})
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, appconfig.ErrVersionNotFound):
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "config version not found"))
	case errors.Is(err, appconfig.ErrInvalidVersion):
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, err.Error()))
	default:
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "config version operation failed"))
	}
}
