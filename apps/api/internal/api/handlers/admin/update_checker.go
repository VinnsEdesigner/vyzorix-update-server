package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateCheckerHandler struct {
	currentVersion string
	repo           string
}

func NewUpdateCheckerHandler(currentVersion, repo string) *UpdateCheckerHandler {
	return &UpdateCheckerHandler{currentVersion: currentVersion, repo: repo}
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

func (h *UpdateCheckerHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", h.repo)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "failed to check for updates", "current_version": h.currentVersion})
		return
	}
	defer func() { _ = resp.Body.Close() }()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "failed to parse release info", "current_version": h.currentVersion})
		return
	}

	updateAvailable := release.TagName != h.currentVersion && release.TagName != ""
	c.JSON(http.StatusOK, gin.H{
		"current_version":   h.currentVersion,
		"latest_version":    release.TagName,
		"update_available":  updateAvailable,
		"release_name":      release.Name,
		"release_url":       release.HTMLURL,
	})
}
