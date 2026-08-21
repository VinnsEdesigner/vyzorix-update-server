// Package configversion provides snapshot-on-write versioning for org-scoped
// config resources (settings, thresholds). Restore re-applies a snapshot.
package configversion

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/configversion"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/uuid"
)

var (
	ErrVersionNotFound = errors.New("config version not found")
	ErrInvalidVersion  = errors.New("invalid config version")
)

// Service provides versioned snapshots of config resources.
type Service struct {
	repo configversion.Repository
}

// NewService creates a new versioning Service.
func NewService(repo configversion.Repository) *Service {
	return &Service{repo: repo}
}

// Snapshot captures a resource state as a new version.
func (s *Service) Snapshot(ctx context.Context, orgID string, resourceType configversion.ResourceType, snapshot interface{}, changedBy string) (int64, error) {
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return 0, err
	}
	v := &configversion.ConfigVersion{
		ID:           uuid.New(),
		OrgID:        orgID,
		ResourceType: resourceType,
		Snapshot:     string(raw),
		ChangedBy:    changedBy,
		CreatedAt:    time.Now(),
	}
	if err := v.Validate(); err != nil {
		return 0, errors.Join(ErrInvalidVersion, err)
	}
	return s.repo.Append(ctx, v)
}

// List returns versions of one resource, newest first.
func (s *Service) List(ctx context.Context, orgID string, resourceType configversion.ResourceType, limit int) ([]*configversion.ConfigVersion, error) {
	return s.repo.List(ctx, orgID, resourceType, limit)
}

// Get returns one version's snapshot.
func (s *Service) Get(ctx context.Context, orgID string, resourceType configversion.ResourceType, version int64) (*configversion.ConfigVersion, error) {
	v, err := s.repo.Get(ctx, orgID, resourceType, version)
	if errors.Is(err, configversion.ErrNotFound) {
		return nil, ErrVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Latest returns the current highest version number for a resource.
func (s *Service) Latest(ctx context.Context, orgID string, resourceType configversion.ResourceType) (int64, error) {
	return s.repo.Latest(ctx, orgID, resourceType)
}
