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

	s.logger.Info("Starting GitHub sync")

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
	for _, release := range releases {
		if !s.shouldProcessRelease(release) {
			continue
		}

		asset, err := s.findAPKAsset(release)
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
	}

	// Update latest flag
	if versionsSynced > 0 {
		latest, err := s.repo.GetLatestVersion(ctx)
		if err == nil && latest != nil {
			if err := s.repo.UpdateLatestFlag(ctx, latest.ID); err != nil {
				s.logger.Warn("Failed to update latest flag", "error", err)
			}
		}
	}

	result.Status = "synced"
	result.VersionsFound = versionsSynced
	result.Message = fmt.Sprintf("Successfully synced %d versions", versionsSynced)

	s.logger.Info("GitHub sync completed", "versions", versionsSynced)

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

// findAPKAsset finds the APK asset in a release.
func (s *SyncService) findAPKAsset(release GitHubRelease) (*VersionAsset, error) {
	for _, asset := range release.Assets {
		if strings.HasSuffix(strings.ToLower(asset.Name), ".apk") {
			return &VersionAsset{
				Name:               asset.Name,
				BrowserDownloadURL: asset.BrowserDownloadURL,
				Size:               asset.Size,
				SHA256:             "", // SHA256 would need to be fetched separately
			}, nil
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

	return &updates.UpdateVersion{
		Version:      extractVersionFromTag(release.TagName),
		APKFilename:  asset.Name,
		APKSize:      asset.Size,
		SHA256:       asset.SHA256,
		ReleaseNotes: notes,
		ReleaseType:  releaseType,
		ReleaseDate:  publishedTime.UnixMilli(),
		IsLatest:     release.Latest,
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
