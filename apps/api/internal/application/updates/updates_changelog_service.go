package updates

import (
	"context"
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// ChangelogService handles changelog operations.
type ChangelogService struct {
	repo updates.Repository
}

// NewChangelogService creates a new ChangelogService.
func NewChangelogService(repo updates.Repository) *ChangelogService {
	return &ChangelogService{repo: repo}
}

// GetChangelog returns the changelog.
func (s *ChangelogService) GetChangelog(ctx context.Context, version string) (*GetChangelogResponse, error) {
	var versions []*updates.UpdateVersion
	var err error

	if version == "" || version == "all" {
		versions, _, err = s.repo.ListVersions(ctx, "all", 100, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list versions for changelog: %w", err)
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

	changelog := make([]ChangelogEntry, 0, len(versions))
	for _, v := range versions {
		date := time.UnixMilli(v.ReleaseDate).Format("2006-01-02")
		changelog = append(changelog, ChangelogEntry{
			Version: v.Version,
			Date:    date,
			Type:    string(v.ReleaseType),
			Notes:   v.ReleaseNotes,
		})
	}

	return &GetChangelogResponse{
		Changelog: changelog,
	}, nil
}
