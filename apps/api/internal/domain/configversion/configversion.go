// Package configversion provides versioned snapshots of org-scoped config
// resources (settings, thresholds). Each write creates a versioned entry so
// changes can be listed, diffed, and restored.
package configversion

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrNotFound is returned when a config version is not found.
var ErrNotFound = errors.New("config version not found")

// ResourceType identifies the config surface being versioned.
type ResourceType string

const (
	ResourceTypeSettings   ResourceType = "org_settings"
	ResourceTypeThresholds ResourceType = "org_thresholds"
)

// Valid reports whether the resource type is known.
func (t ResourceType) Valid() bool {
	switch t {
	case ResourceTypeSettings, ResourceTypeThresholds:
		return true
	}
	return false
}

// ConfigVersion is one snapshot of a resource at write time.
type ConfigVersion struct {
	CreatedAt    time.Time
	ID           string
	OrgID        string
	ResourceType ResourceType
	Snapshot     string
	ChangedBy    string
	Version      int64
}

// Validate checks the config version is well-formed.
func (v *ConfigVersion) Validate() error {
	if strings.TrimSpace(v.OrgID) == "" {
		return errors.New("org_id is required")
	}
	if !v.ResourceType.Valid() {
		return errors.New("invalid resource type")
	}
	if strings.TrimSpace(v.Snapshot) == "" {
		return errors.New("snapshot is required")
	}
	return nil
}

// Repository persists config versions.
type Repository interface {
	// Append creates a new version entry and returns its assigned version.
	Append(ctx context.Context, v *ConfigVersion) (int64, error)
	// List returns versions of one resource, newest first.
	List(ctx context.Context, orgID string, resourceType ResourceType, limit int) ([]*ConfigVersion, error)
	// Get returns one version.
	Get(ctx context.Context, orgID string, resourceType ResourceType, version int64) (*ConfigVersion, error)
	// Latest returns the highest version number for a resource, or 0 when none.
	Latest(ctx context.Context, orgID string, resourceType ResourceType) (int64, error)
}
