package updates

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/gin-gonic/gin"
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
func (h *UpdatesVersionsHandler) GetStatus(c *gin.Context) {
	status, err := h.service.GetStatus(c.Request.Context())
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			c.JSON(se.Status, se.ToErrorResponse())
			return
		}
		c.JSON(http.StatusInternalServerError, updates.ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to get status",
		})
		return
	}
	c.JSON(http.StatusOK, status)
}

// GetVersions handles GET /v1/updates/versions.
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
			c.JSON(se.Status, se.ToErrorResponse())
			return
		}
		c.JSON(http.StatusInternalServerError, updates.ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to get versions",
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetChangelog handles GET /v1/updates/changelog.
func (h *UpdatesVersionsHandler) GetChangelog(c *gin.Context) {
	version := c.Query("version")
	if version == "" {
		version = "all"
	}

	changelog, err := h.service.GetChangelog(c.Request.Context(), version)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			c.JSON(se.Status, se.ToErrorResponse())
			return
		}
		c.JSON(http.StatusInternalServerError, updates.ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to get changelog",
		})
		return
	}
	c.JSON(http.StatusOK, changelog)
}

// GetUpdateStatus is an alias for GetStatus to match expected handler names.
func (h *UpdatesVersionsHandler) GetUpdateStatus(c *gin.Context) {
	h.GetStatus(c)
}

// ExportVersions is an alias for Export to match expected handler names.
func (h *UpdatesVersionsHandler) ExportVersions(c *gin.Context) {
	h.Export(c)
}

// Export handles GET /v1/updates/export.
func (h *UpdatesVersionsHandler) Export(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	version := c.Query("version")
	includeChangelog := c.DefaultQuery("includeChangelog", "true") == "true"
	includeApkInfo := c.DefaultQuery("includeApkInfo", "true") == "true"

	if format != "json" && format != "csv" {
		c.JSON(http.StatusBadRequest, updates.ErrorResponse{
			Code:    "bad_request",
			Message: "Invalid format. Supported formats: json, csv",
		})
		return
	}

	result, err := h.service.ExportVersions(c.Request.Context(), format, version, includeChangelog, includeApkInfo)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			c.JSON(se.Status, se.ToErrorResponse())
			return
		}
		c.JSON(http.StatusInternalServerError, updates.ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to export versions",
		})
		return
	}

	if format == "csv" {
		// Generate CSV content
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

	// Write CSV header
	if includeApkInfo {
		csvBuilder.WriteString("Version,Filename,Size (bytes),SHA256,Released,Release Type,Release Notes\n")
	} else {
		csvBuilder.WriteString("Version,Released,Release Type,Release Notes\n")
	}

	// Write version rows
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

	// Write changelog if included
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
	// If the string contains comma, quote, or newline, wrap in quotes and escape internal quotes
	if strings.ContainsAny(s, ",\"\n\r") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}
