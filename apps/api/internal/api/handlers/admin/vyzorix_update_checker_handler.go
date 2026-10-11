package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/gin-gonic/gin"

	usagestats "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/usagestats"
)

func usageSnapshotToOpenAPI(s *usagestats.Snapshot) *openapi.UsageStatsSnapshot {
	if s == nil {
		return nil
	}
	return &openapi.UsageStatsSnapshot{
		Counts: openapi.UsageStatsCounts{
			Devices:       s.Counts.Devices,
			Operators:     s.Counts.Operators,
			Organizations: s.Counts.Organizations,
			AlertRules:    s.Counts.AlertRules,
			Annotations:   s.Counts.Annotations,
			ContactPoints: s.Counts.ContactPoints,
		},
		Toggles:     s.Toggles,
		CollectedAt: s.CollectedAt.UnixMilli(),
	}
}

// Compile-time references for swaggo-annotated openapi DTO types.
var _ openapi.UpdateCheckerResult

type UpdateCheckerHandler struct {
	statsSvc       *usagestats.Service
	currentVersion string
	repo           string
}

func NewUpdateCheckerHandler(currentVersion, repo string) *UpdateCheckerHandler {
	return &UpdateCheckerHandler{currentVersion: currentVersion, repo: repo}
}

// SetUsageStats wires the usage stats snapshot collector for the checker response.
func (h *UpdateCheckerHandler) SetUsageStats(svc *usagestats.Service) {
	h.statsSvc = svc
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// Check handles GET /v1/admin/updates/check.
// @Summary      Check for server updates
// @Description  Checks GitHub for the latest server release and includes usage stats
// @Tags         usage-stats
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.UpdateCheckerResult  "update check result"
// @Router       /admin/updates/check [get]
func (h *UpdateCheckerHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", h.repo)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, openapi.UpdateCheckerResult{Error: "failed to check for updates", CurrentVersion: h.currentVersion, UpdateAvailable: false})
		return
	}
	defer func() { _ = resp.Body.Close() }()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		c.JSON(http.StatusOK, openapi.UpdateCheckerResult{Error: "failed to parse release info", CurrentVersion: h.currentVersion, UpdateAvailable: false})
		return
	}

	updateAvailable := release.TagName != h.currentVersion && release.TagName != ""
	response := openapi.UpdateCheckerResult{
		CurrentVersion:  h.currentVersion,
		LatestVersion:   release.TagName,
		UpdateAvailable: updateAvailable,
		ReleaseName:     release.Name,
		ReleaseURL:      release.HTMLURL,
	}
	if h.statsSvc != nil {
		if snap := h.statsSvc.Snapshot(); snap != nil {
			response.UsageStats = usageSnapshotToOpenAPI(snap)
		}
	}
	c.JSON(http.StatusOK, response)
}
