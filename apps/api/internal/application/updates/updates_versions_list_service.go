package updates

import (
	"context"
	"fmt"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// VersionsListService handles version listing operations.
type VersionsListService struct {
	repo updates.Repository
}

// NewVersionsListService creates a new VersionsListService.
func NewVersionsListService(repo updates.Repository) *VersionsListService {
	return &VersionsListService{repo: repo}
}

// GetVersions returns paginated versions.
func (s *VersionsListService) GetVersions(ctx context.Context, status string, page, limit int) (*ListVersionsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	offset := (page - 1) * limit

	versions, total, err := s.repo.ListVersions(ctx, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	versionResponses := make([]VersionResponse, 0, len(versions))
	for _, v := range versions {
		versionStatus := "previous"
		if v.IsLatest {
			versionStatus = "latest"
		}

		versionResponses = append(versionResponses, VersionResponse{
			Version:      v.Version,
			APKFilename:  v.APKFilename,
			APKSize:      v.APKSize,
			SHA256:       v.SHA256,
			ReleasedAt:   v.ReleaseDate,
			ReleaseNotes: v.ReleaseNotes,
			Status:       versionStatus,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return &ListVersionsResponse{
		Versions: versionResponses,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
