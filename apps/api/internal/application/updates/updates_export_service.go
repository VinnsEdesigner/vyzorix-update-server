package updates

import (
	"context"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// ExportService handles version export operations.
type ExportService struct {
	repo updates.Repository
}

// NewExportService creates a new ExportService.
func NewExportService(repo updates.Repository) *ExportService {
	return &ExportService{repo: repo}
}

// ExportVersions exports version data.
func (s *ExportService) ExportVersions(ctx context.Context, format, version string, includeChangelog, includeApkInfo bool) (*ExportResponse, error) {
	exportedAt := time.Now().UnixMilli()

	var versions []*updates.UpdateVersion
	var err error

	if version == "" || version == "all" {
		versions, _, err = s.repo.ListVersions(ctx, "all", 100, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list versions: %w", err)
		}
	} else {
		v, err := s.repo.GetVersionByVersion(ctx, version)
		if err != nil {
			if err == updates.ErrVersionNotFound {
				return nil, ErrVersionNotFound
			}
			return nil, fmt.Errorf("failed to get version: %w", err)
		}
		versions = []*updates.UpdateVersion{v}
	}

	versionResponses := make([]VersionResponse, 0, len(versions))
	for _, v := range versions {
		statusStr := "previous"
		if v.IsLatest {
			statusStr = "latest"
		}

		vr := VersionResponse{
			Version:      v.Version,
			Status:       statusStr,
			ReleasedAt:   v.ReleaseDate,
			ReleaseNotes: v.ReleaseNotes,
		}
		if includeApkInfo {
			vr.APKFilename = v.APKFilename
			vr.APKSize = v.APKSize
			vr.SHA256 = v.SHA256
		}
		versionResponses = append(versionResponses, vr)
	}

	response := &ExportResponse{
		Format:     format,
		ExportedAt: exportedAt,
		Versions:   versionResponses,
	}

	if includeChangelog {
		changelogEntries := make([]ChangelogEntry, 0, len(versions))
		for _, v := range versions {
			date := time.UnixMilli(v.ReleaseDate).Format("2006-01-02")
			changelogEntries = append(changelogEntries, ChangelogEntry{
				Version: v.Version,
				Date:    date,
				Type:    string(v.ReleaseType),
				Notes:   v.ReleaseNotes,
			})
		}
		response.Changelog = changelogEntries
	}

	return response, nil
}
