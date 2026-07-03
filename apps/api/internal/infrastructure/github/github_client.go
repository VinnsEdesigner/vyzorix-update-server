package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles GitHub API interactions.
type Client struct {
	httpClient *http.Client
	owner      string
	repo       string
	token      string
}

// NewClient creates a new GitHub client.
func NewClient(owner, repo, token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		owner: owner,
		repo:  repo,
		token: token,
	}
}

// GitHubRelease represents a GitHub release from the API.
type GitHubRelease struct {
	Name        string `json:"name"`
	TagName     string `json:"tag_name"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
	ID         int64 `json:"id"`
	Latest     bool  `json:"latest"`
	Draft      bool  `json:"draft"`
	Prerelease bool  `json:"prerelease"`
}

// FetchReleases fetches releases from a GitHub repository.
func (c *Client) FetchReleases(ctx context.Context) ([]GitHubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", c.owner, c.repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("User-Agent", "Vyzorix-Update-Server")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var releases []GitHubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to unmarshal releases: %w", err)
	}

	return releases, nil
}

// FetchVersionInfo fetches version information from GitHub.
func (c *Client) FetchVersionInfo(ctx context.Context, version string) (*VersionInfo, error) {
	tagURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", c.owner, c.repo, version)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tagURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("User-Agent", "Vyzorix-Update-Server")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("version %s not found", version)
		}
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to unmarshal release: %w", err)
	}

	assets := make([]struct {
		Name               string
		BrowserDownloadURL string
		Size               int64
	}, len(release.Assets))
	for i, a := range release.Assets {
		assets[i] = struct {
			Name               string
			BrowserDownloadURL string
			Size               int64
		}{
			Name:               a.Name,
			BrowserDownloadURL: a.BrowserDownloadURL,
			Size:               a.Size,
		}
	}

	return &VersionInfo{
		Version:   release.TagName,
		Notes:     release.Body,
		Published: release.PublishedAt,
		Assets:    assets,
	}, nil
}

// VersionInfo represents version information from GitHub.
type VersionInfo struct {
	Version   string
	Notes     string
	Published string
	Assets    []struct {
		Name               string
		BrowserDownloadURL string
		Size               int64
	}
}

// extractVersionFromTag extracts version from a tag name (removes 'v' prefix).
func extractVersionFromTag(tag string) string {
	if len(tag) > 0 && tag[0] == 'v' {
		return tag[1:]
	}
	return tag
}

// FetchAssetChecksum fetches the SHA256 checksum for an asset from a checksums file.
// It looks for common checksum file names like SHA256SUMS, checksums.txt, etc.
func (c *Client) FetchAssetChecksum(ctx context.Context, releaseTag, assetName string) (string, error) {
	checksumFiles := []string{"SHA256SUMS", "SHA256SUMS.txt", "checksums.txt", "checksums.txt.sha256"}

	for _, filename := range checksumFiles {
		url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", c.owner, c.repo, releaseTag, filename)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Vyzorix-Update-Server")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		// Parse checksum file - format is typically: "sha256  filename" or "sha256 *filename"
		content := string(body)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Handle both "sha256  filename" and "sha256 *filename" formats
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				checksum := parts[0]
				name := parts[len(parts)-1] // Last field is the filename
				// Remove any leading * for globs
				name = strings.TrimPrefix(name, "*")
				if name == assetName {
					return checksum, nil
				}
			}
		}
	}

	return "", fmt.Errorf("checksum not found for asset %s", assetName)
}
