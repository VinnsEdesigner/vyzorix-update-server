package notifications

import (
	"context"
	"errors"
	"time"

	notificationdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/notification"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/uuid"
)

var (
	ErrContactPointNotFound = errors.New("contact point not found")
	ErrInvalidContactPoint  = errors.New("invalid contact point")
)

// ContactPointInput carries the mutable fields of a create/update request.
type ContactPointInput struct {
	OrgID      string
	Name       string
	Channel    notificationdomain.ChannelType
	Secret     string
	Config     map[string]string
	TemplateID string
	Enabled    bool
}

// Service provides contact point CRUD.
type Service struct {
	repo notificationdomain.Repository
}

// NewService creates a new contact point Service.
func NewService(repo notificationdomain.Repository) *Service {
	return &Service{repo: repo}
}

// Create validates and persists a contact point.
func (s *Service) Create(ctx context.Context, in *ContactPointInput) (*notificationdomain.ContactPoint, error) {
	now := time.Now()
	cp := &notificationdomain.ContactPoint{
		ID:         uuid.New(),
		CreatedAt:  now,
		UpdatedAt:  now,
		OrgID:      in.OrgID,
		Name:       in.Name,
		Channel:    in.Channel,
		Secret:     in.Secret,
		Config:     in.Config,
		TemplateID: in.TemplateID,
		Enabled:    in.Enabled,
	}
	if cp.Config == nil {
		cp.Config = make(map[string]string)
	}
	if err := cp.Validate(); err != nil {
		return nil, errors.Join(ErrInvalidContactPoint, err)
	}
	if err := s.repo.Save(ctx, cp); err != nil {
		return nil, err
	}
	return cp, nil
}

// Update replaces mutable fields of an existing contact point.
func (s *Service) Update(ctx context.Context, orgID, id string, in *ContactPointInput) (*notificationdomain.ContactPoint, error) {
	cp, err := s.getScoped(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	cp.Name = in.Name
	cp.Channel = in.Channel
	cp.Secret = in.Secret
	cp.Config = in.Config
	if cp.Config == nil {
		cp.Config = make(map[string]string)
	}
	cp.TemplateID = in.TemplateID
	cp.Enabled = in.Enabled
	cp.UpdatedAt = time.Now()

	if err := cp.Validate(); err != nil {
		return nil, errors.Join(ErrInvalidContactPoint, err)
	}
	if err := s.repo.Save(ctx, cp); err != nil {
		return nil, err
	}
	return cp, nil
}

// Delete removes a contact point.
func (s *Service) Delete(ctx context.Context, orgID, id string) error {
	if _, err := s.getScoped(ctx, orgID, id); err != nil {
		return err
	}
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrContactPointNotFound
	}
	return nil
}

// Get returns one contact point.
func (s *Service) Get(ctx context.Context, orgID, id string) (*notificationdomain.ContactPoint, error) {
	return s.getScoped(ctx, orgID, id)
}

// List returns all contact points of an org.
func (s *Service) List(ctx context.Context, orgID string) ([]*notificationdomain.ContactPoint, error) {
	return s.repo.ListByOrg(ctx, orgID)
}

func (s *Service) getScoped(ctx context.Context, orgID, id string) (*notificationdomain.ContactPoint, error) {
	cp, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, notificationdomain.ErrNotFound) {
		return nil, ErrContactPointNotFound
	}
	if err != nil {
		return nil, err
	}
	if cp.OrgID != orgID {
		return nil, ErrContactPointNotFound
	}
	return cp, nil
}
