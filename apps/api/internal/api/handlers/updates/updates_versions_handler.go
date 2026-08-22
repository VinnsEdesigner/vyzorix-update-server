package updates

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.UpdateVersionManifest
	_ openapi.UpdateChangelogEntry
	_ openapi.UpdateCheckRequest
	_ openapi.UpdateCheckResult
	_ openapi.DownloadProgressRequest
	_ openapi.DownloadProgressResult
	_ openapi.ErrorResponse
)

// UpdatesVersionsHandler handles version-related HTTP requests.
type UpdatesVersionsHandler struct {
	service *updates.Service
}

// NewUpdatesVersionsHandler creates a new UpdatesVersionsHandler.
func NewUpdatesVersionsHandler(service *updates.Service) *UpdatesVersionsHandler {
	return &UpdatesVersionsHandler{service: service}
}

// GetStatus handles GET /v1/updates/status.
// @Summary      Get update sync status
// @Description  Returns the current GitHub-release sync status.
// @Tags         updates
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  updates.SyncState  "sync status"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /updates/status [get]
func (h *UpdatesVersionsHandler) GetStatus(c *gin.Context) {
	status, err := h.service.GetStatus(c.Request.Context())
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to get status"))
		return
	}
	c.JSON(http.StatusOK, status)
}

// GetVersions handles GET /v1/updates/versions.
// @Summary      List update versions
// @Description  Returns a paginated list of synced update versions.
// @Tags         updates
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        status  query string  false  "filter by version status"
// @Param        page    query int     false  "page number (default 1)"
// @Param        limit   query int     false  "page size (default 20, max 50)"
// @Success      200  {object}  updates.VersionListResult  "versions"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /updates/versions [get]
func (h *UpdatesVersionsHandler) GetVersions(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	result, err := h.service.GetVersions(c.Request.Context(), status, page, limit)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to get versions"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetChangelog handles GET /v1/updates/changelog.
// @Summary      Get update changelog
// @Description  Returns the changelog for a version (or all versions when omitted).
// @Tags         updates
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        version  query string  false  "version filter (default 'all')"
// @Success      200  {object}  updates.ChangelogResult  "changelog"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /updates/versions/{id}/changelog [get]
func (h *UpdatesVersionsHandler) GetChangelog(c *gin.Context) {
	version := c.Query("version")
	if version == "" {
		version = "all"
	}

	changelog, err := h.service.GetChangelog(c.Request.Context(), version)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to get changelog"))
		return
	}
	c.JSON(http.StatusOK, changelog)
}

// GetUpdateStatus is an alias for GetStatus to match expected handler names.
// @Summary      Get update status (alias)
// @Description  Alias for GET /updates/status.
// @Tags         updates
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "version ID"
// @Success      200  {object}  updates.SyncState  "sync status"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /updates/{id}/status [get]
func (h *UpdatesVersionsHandler) GetUpdateStatus(c *gin.Context) {
	h.GetStatus(c)
}

// ExportVersions is an alias for Export to match expected handler names.
func (h *UpdatesVersionsHandler) ExportVersions(c *gin.Context) {
	h.Export(c)
}

// Export handles GET /v1/updates/export.
// @Summary      Export versions
// @Description  Exports versions as JSON or CSV.
// @Tags         updates
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        format             query string  false  "json or csv (default json)"
// @Param        version            query string  false  "version filter"
// @Param        includeChangelog   query bool    false  "include changelog (default true)"
// @Param        includeApkInfo     query bool    false  "include APK metadata (default true)"
// @Success      200  {object}  updates.ExportResponse  "exported versions"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid format"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /updates/export [get]
func (h *UpdatesVersionsHandler) Export(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	version := c.Query("version")
	includeChangelog := c.DefaultQuery("includeChangelog", "true") == "true"
	includeApkInfo := c.DefaultQuery("includeApkInfo", "true") == "true"

	if format != "json" && format != "csv" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid format. Supported formats: json, csv"))
		return
	}

	result, err := h.service.ExportVersions(c.Request.Context(), format, version, includeChangelog, includeApkInfo)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to export versions"))
		return
	}

	if format == "csv" {
		// Generate CSV content.
		csv := h.generateCSV(result, includeApkInfo, includeChangelog)
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=versions.csv")
		c.String(http.StatusOK, csv)
		return
	}

	c.JSON(http.StatusOK, result)
}

// generateCSV generates CSV content from export response.
func (h *UpdatesVersionsHandler) generateCSV(result *updates.ExportResponse, includeApkInfo, includeChangelog bool) string {
	var csvBuilder strings.Builder

	// Write CSV header.
	if includeApkInfo {
		csvBuilder.WriteString("Version,Filename,Size (bytes),SHA256,Released,Release Type,Release Notes\n")
	} else {
		csvBuilder.WriteString("Version,Released,Release Type,Release Notes\n")
	}

	// Write version rows.
	for _, v := range result.Versions {
		if includeApkInfo {
			fmt.Fprintf(&csvBuilder, "%s,%s,%d,%s,%d,%s,%s\n",
				escapeCSV(v.Version),
				escapeCSV(v.APKFilename),
				v.APKSize,
				escapeCSV(v.SHA256),
				v.ReleasedAt,
				escapeCSV(v.Status),
				escapeCSV(v.ReleaseNotes))
		} else {
			fmt.Fprintf(&csvBuilder, "%s,%d,%s,%s\n",
				escapeCSV(v.Version),
				v.ReleasedAt,
				escapeCSV(v.Status),
				escapeCSV(v.ReleaseNotes))
		}
	}

	// Write changelog if included.
	if includeChangelog && len(result.Changelog) > 0 {
		csvBuilder.WriteString("\n--- Changelog ---\n")
		csvBuilder.WriteString("Version,Date,Type,Notes\n")
		for _, entry := range result.Changelog {
			fmt.Fprintf(&csvBuilder, "%s,%s,%s,%s\n",
				escapeCSV(entry.Version),
				escapeCSV(entry.Date),
				escapeCSV(entry.Type),
				escapeCSV(entry.Notes))
		}
	}

	return csvBuilder.String()
}

// escapeCSV escapes a string for CSV output.
func escapeCSV(s string) string {
	if s == "" {
		return ""
	}
	// If the string contains comma, quote, or newline, wrap in quotes and escape internal quotes.
	if strings.ContainsAny(s, ",\"\n\r") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}
