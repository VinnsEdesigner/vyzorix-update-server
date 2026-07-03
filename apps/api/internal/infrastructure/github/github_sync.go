package github

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// SyncService handles syncing versions from GitHub.
type SyncService struct {
	client *Client
	repo   updates.Repository
	logger *slog.Logger
	skipSHA256 bool // For testing or when checksums not available
}

// VersionAsset represents APK asset information.
type VersionAsset struct {
	Name               string
	BrowserDownloadURL string
	SHA256             string
	Size               int64
}

// NewSyncService creates a new GitHub sync service.
func NewSyncService(client *Client, repo updates.Repository, logger *slog.Logger) *SyncService {
	return &SyncService{
		client: client,
		repo:   repo,
		logger: logger,
	}
}

// SyncResult represents the result of a sync operation.
type SyncResult struct {
	Error         error
	Status        string
	Message       string
	StartedAt     int64
	VersionsFound int
}

// SyncFromGitHub syncs all versions from GitHub releases.
func (s *SyncService) SyncFromGitHub(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		StartedAt: time.Now().UnixMilli(),
		Status:    "syncing",
	}

	s.logger.InfoContext(ctx, "Starting GitHub sync")

	releases, err := s.client.FetchReleases(ctx)
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("Failed to fetch releases: %v", err)
		result.Error = err
		s.logger.Error("Failed to fetch GitHub releases", "error", err)
		return result, err
	}

	s.logger.Info("Fetched GitHub releases", "count", len(releases))

	versionsSynced := 0
	var latestVersion *updates.UpdateVersion
	var latestReleaseDate int64

	for _, release := range releases {
		if !s.shouldProcessRelease(release) {
			continue
		}

		asset, err := s.findAPKAsset(ctx, release)
		if err != nil {
			s.logger.Warn("No APK asset found for release", "release", release.TagName)
			continue
		}

		existing, err := s.repo.GetVersionByVersion(ctx, extractVersionFromTag(release.TagName))
		if err != nil && err != updates.ErrVersionNotFound {
			s.logger.Warn("Failed to check existing version", "version", release.TagName, "error", err)
			continue
		}

		version := s.createUpdateVersion(release, asset)
		if existing != nil {
			version.ID = existing.ID
			s.logger.Info("Updating existing version", "version", version.Version)
		} else {
			s.logger.Info("Creating new version", "version", version.Version)
		}

		if existing != nil {
			if err := s.updateVersion(ctx, version); err != nil {
				s.logger.Warn("Failed to update version", "version", version.Version, "error", err)
				continue
			}
		} else {
			if err := s.repo.CreateVersion(ctx, version); err != nil {
				s.logger.Warn("Failed to create version", "version", version.Version, "error", err)
				continue
			}
		}
		versionsSynced++

		// Track the version with the most recent release date
		if latestVersion == nil || version.ReleaseDate > latestReleaseDate {
			latestVersion = version
			latestReleaseDate = version.ReleaseDate
		}
	}

	// Update latest flag based on release date, not GitHub's "Latest" flag
	if latestVersion != nil {
		if err := s.repo.UpdateLatestFlag(ctx, latestVersion.ID); err != nil {
			s.logger.Warn("Failed to update latest flag", "error", err)
		} else {
			s.logger.Info("Updated latest version", "version", latestVersion.Version, "released", latestReleaseDate)
		}
	}

	result.Status = "synced"
	result.VersionsFound = versionsSynced
	result.Message = fmt.Sprintf("Successfully synced %d versions", versionsSynced)

	s.logger.InfoContext(ctx, "GitHub sync completed", "versions", versionsSynced)

	return result, nil
}

// shouldProcessRelease determines if a release should be processed.
func (s *SyncService) shouldProcessRelease(release GitHubRelease) bool {
	if release.Draft {
		return false
	}
	if release.Prerelease {
		return false
	}
	if release.TagName == "" {
		return false
	}
	if !strings.HasPrefix(release.TagName, "v") && !strings.HasPrefix(release.TagName, "V") {
		return false
	}
	return true
}

// findAPKAsset finds the APK asset in a release and fetches its SHA256 checksum.
func (s *SyncService) findAPKAsset(ctx context.Context, release GitHubRelease) (*VersionAsset, error) {
	for _, asset := range release.Assets {
		if strings.HasSuffix(strings.ToLower(asset.Name), ".apk") {
			versionAsset := &VersionAsset{
				Name:               asset.Name,
				BrowserDownloadURL: asset.BrowserDownloadURL,
				Size:               asset.Size,
			}

			// Try to fetch SHA256 checksum if not in skip mode
			if !s.skipSHA256 {
				sha256, err := s.client.FetchAssetChecksum(ctx, release.TagName, asset.Name)
				if err != nil {
					// Log warning but don't fail - checksum is optional
					s.logger.Warn("Failed to fetch SHA256 for asset",
						"asset", asset.Name,
						"release", release.TagName,
						"error", err)
				} else {
					versionAsset.SHA256 = sha256
				}
			}

			return versionAsset, nil
		}
	}
	return nil, fmt.Errorf("no APK asset found")
}

// createUpdateVersion creates an UpdateVersion from a GitHub release.
func (s *SyncService) createUpdateVersion(release GitHubRelease, asset *VersionAsset) *updates.UpdateVersion {
	publishedTime := time.Now()
	if release.PublishedAt != "" {
		if t, err := time.Parse(time.RFC3339, release.PublishedAt); err == nil {
			publishedTime = t
		}
	}

	releaseType := s.determineReleaseType(release.Body)
	notes := release.Body
	if notes == "" {
		notes = release.Name
	}

	// NOTE: IsLatest is NOT set here. The latest version is determined by release date
	// during the sync process to avoid relying on GitHub's manually-set "Latest" flag.
	return &updates.UpdateVersion{
		Version:      extractVersionFromTag(release.TagName),
		APKFilename:  asset.Name,
		APKSize:      asset.Size,
		SHA256:       asset.SHA256,
		ReleaseNotes: notes,
		ReleaseType:  releaseType,
		ReleaseDate:  publishedTime.UnixMilli(),
		IsLatest:     false, // Will be set correctly during sync based on release date
		CreatedAt:    time.Now().UnixMilli(),
		UpdatedAt:    time.Now().UnixMilli(),
	}
}

// determineReleaseType determines the release type from the release notes.
func (s *SyncService) determineReleaseType(notes string) updates.ReleaseType {
	lower := strings.ToLower(notes)
	if strings.Contains(lower, "breaking") || strings.Contains(lower, "major change") {
		return updates.ReleaseTypeMajor
	}
	if strings.Contains(lower, "new feature") || strings.Contains(lower, "added") || strings.Contains(lower, "enhancement") {
		return updates.ReleaseTypeMinor
	}
	return updates.ReleaseTypePatch
}

// updateVersion updates an existing version in the repository.
func (s *SyncService) updateVersion(ctx context.Context, version *updates.UpdateVersion) error {
	version.UpdatedAt = time.Now().UnixMilli()
	return s.repo.UpdateVersion(ctx, version)
}
