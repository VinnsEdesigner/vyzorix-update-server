// Package annotation provides annotation CRUD and fleet timeline lookup.
package annotation

import (
	"context"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/annotation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/uuid"
)

var (
	ErrAnnotationNotFound = errors.New("annotation not found")
	ErrInvalidAnnotation  = errors.New("invalid annotation")
)

// AnnotationInput carries the mutable fields of a create/update request.
type AnnotationInput struct {
	StartTime time.Time
	EndTime   *time.Time
	OrgID     string
	Title     string
	Text      string
	Source    string
	Tags      []string
}

// Service provides annotation CRUD and filtered lookup.
type Service struct {
	repo annotation.Repository
}

// NewService creates a new annotation Service.
func NewService(repo annotation.Repository) *Service {
	return &Service{repo: repo}
}

// Create validates and persists an annotation.
func (s *Service) Create(ctx context.Context, in *AnnotationInput) (*annotation.Annotation, error) {
	now := time.Now()
	a := &annotation.Annotation{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		OrgID:     in.OrgID,
		Title:     in.Title,
		Text:      in.Text,
		Tags:      in.Tags,
		Source:    in.Source,
		StartTime: in.StartTime,
		EndTime:   in.EndTime,
	}
	if a.Tags == nil {
		a.Tags = []string{}
	}
	if err := a.Validate(); err != nil {
		return nil, errors.Join(ErrInvalidAnnotation, err)
	}
	if err := s.repo.Save(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Update replaces mutable fields of an existing annotation.
func (s *Service) Update(ctx context.Context, orgID, id string, in *AnnotationInput) (*annotation.Annotation, error) {
	a, err := s.getScoped(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	a.Title = in.Title
	a.Text = in.Text
	a.Tags = in.Tags
	if a.Tags == nil {
		a.Tags = []string{}
	}
	a.Source = in.Source
	a.StartTime = in.StartTime
	a.EndTime = in.EndTime
	a.UpdatedAt = time.Now()

	if err := a.Validate(); err != nil {
		return nil, errors.Join(ErrInvalidAnnotation, err)
	}
	if err := s.repo.Save(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Delete removes an annotation.
func (s *Service) Delete(ctx context.Context, orgID, id string) error {
	if _, err := s.getScoped(ctx, orgID, id); err != nil {
		return err
	}
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrAnnotationNotFound
	}
	return nil
}

// Get returns one annotation.
func (s *Service) Get(ctx context.Context, orgID, id string) (*annotation.Annotation, error) {
	return s.getScoped(ctx, orgID, id)
}

// List returns annotations matching the filter.
func (s *Service) List(ctx context.Context, f *annotation.Filter) ([]*annotation.Annotation, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) getScoped(ctx context.Context, orgID, id string) (*annotation.Annotation, error) {
	a, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, annotation.ErrNotFound) {
		return nil, ErrAnnotationNotFound
	}
	if err != nil {
		return nil, err
	}
	if a.OrgID != orgID {
		return nil, ErrAnnotationNotFound
	}
	return a, nil
}
