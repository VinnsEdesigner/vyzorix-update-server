// Package annotation provides org-scoped time-range markers on the fleet
// timeline. Annotations pair with alert rules and updates so operators can
// correlate "rollout v2.3 started" with a subsequent failure spike.
package annotation

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrNotFound is returned when an annotation is not found.
var ErrNotFound = errors.New("annotation not found")

// Annotation is a time-range marker with freeform text and filterable tags.
type Annotation struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	StartTime time.Time
	EndTime   *time.Time
	ID        string
	OrgID     string
	Title     string
	Text      string
	Source    string
	Tags      []string
}

// Validate checks the annotation is well-formed.
func (a *Annotation) Validate() error {
	if strings.TrimSpace(a.OrgID) == "" {
		return errors.New("org_id is required")
	}
	if strings.TrimSpace(a.Title) == "" {
		return errors.New("title is required")
	}
	if a.StartTime.IsZero() {
		return errors.New("start_time is required")
	}
	if a.EndTime != nil && a.EndTime.Before(a.StartTime) {
		return errors.New("end_time must not be before start_time")
	}
	return nil
}

// Filter carries query parameters for listing annotations.
type Filter struct {
	StartTime time.Time
	EndTime   time.Time
	OrgID     string
	Tag       string
	Limit     int
}

// Repository persists annotations.
type Repository interface {
	Save(ctx context.Context, a *Annotation) error
	GetByID(ctx context.Context, id string) (*Annotation, error)
	List(ctx context.Context, f *Filter) ([]*Annotation, error)
	Delete(ctx context.Context, id string) (bool, error)
}
