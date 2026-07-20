package organization

import (
	"context"
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// OrganizationSettingsService handles organization settings operations.
type OrganizationSettingsService struct {
	settingsRepo organization.OrganizationSettingsRepository
	orgRepo      organization.OrganizationRepository
	memberRepo   organization.MemberRepository
}

// NewOrganizationSettingsService creates a new OrganizationSettingsService.
func NewOrganizationSettingsService(
	settingsRepo organization.OrganizationSettingsRepository,
	orgRepo organization.OrganizationRepository,
	memberRepo organization.MemberRepository,
) *OrganizationSettingsService {
	return &OrganizationSettingsService{
		settingsRepo: settingsRepo,
		orgRepo:      orgRepo,
		memberRepo:   memberRepo,
	}
}

// SettingsRepo returns the settings repository for use by other services.
func (s *OrganizationSettingsService) SettingsRepo() organization.OrganizationSettingsRepository {
	return s.settingsRepo
}

// CreateSettings creates organization settings with defaults for a new organization.
func (s *OrganizationSettingsService) CreateSettings(ctx context.Context, orgID string) (*organization.OrganizationSettings, error) {
	// Verify organization exists
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrNotFound) {
			return nil, organization.ErrNotFound
		}
		return nil, err
	}

	if org.IsDeleted() {
		return nil, organization.ErrNotFound
	}

	// Check if settings already exist
	existing, err := s.settingsRepo.FindByOrganizationID(ctx, orgID)
	if err != nil && !errors.Is(err, organization.ErrSettingsNotFound) {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Already exists, return it
	}

	// Create new settings with defaults
	settings := organization.NewOrganizationSettings(orgID)

	if err := s.settingsRepo.Create(ctx, settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// GetSettings retrieves organization settings.
func (s *OrganizationSettingsService) GetSettings(ctx context.Context, orgID string) (*organization.OrganizationSettings, error) {
	settings, err := s.settingsRepo.FindByOrganizationID(ctx, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrSettingsNotFound) {
			return nil, organization.ErrSettingsNotFound
		}
		return nil, err
	}

	return settings, nil
}

// GetOrCreateSettings retrieves settings, creating them with defaults if they don't exist.
func (s *OrganizationSettingsService) GetOrCreateSettings(ctx context.Context, orgID string) (*organization.OrganizationSettings, error) {
	settings, err := s.settingsRepo.FindByOrganizationID(ctx, orgID)
	if err != nil && !errors.Is(err, organization.ErrSettingsNotFound) {
		return nil, err
	}

	if settings != nil {
		return settings, nil
	}

	// Create with defaults
	return s.CreateSettings(ctx, orgID)
}

// UpdateSettings updates organization settings.
func (s *OrganizationSettingsService) UpdateSettings(ctx context.Context, orgID string, req *organization.UpdateOrganizationSettingsRequest) (*organization.OrganizationSettings, error) {
	settings, err := s.settingsRepo.FindByOrganizationID(ctx, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrSettingsNotFound) {
			return nil, organization.ErrSettingsNotFound
		}
		return nil, err
	}

	// Apply updates
	if req.Timezone != nil {
		settings.Timezone = *req.Timezone
	}
	if req.DateFormat != nil {
		settings.DateFormat = *req.DateFormat
	}
	if req.AlertCooldownMinutes != nil {
		settings.AlertCooldownMinutes = *req.AlertCooldownMinutes
	}
	if req.DefaultThresholds != nil {
		if err := settings.UpdateThresholds(req.DefaultThresholds); err != nil {
			return nil, err
		}
	}

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateThresholds updates only the thresholds.
func (s *OrganizationSettingsService) UpdateThresholds(ctx context.Context, orgID string, req *organization.UpdateThresholdsRequest) (*organization.OrganizationSettings, error) {
	settings, err := s.settingsRepo.FindByOrganizationID(ctx, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrSettingsNotFound) {
			return nil, organization.ErrSettingsNotFound
		}
		return nil, err
	}

	// Merge with existing thresholds
	thresholds := settings.DefaultThresholds
	if thresholds == nil {
		thresholds = organization.DefaultThresholds()
	}

	// Apply updates from request
	if req.RiskWarn != nil {
		thresholds.RiskWarn = *req.RiskWarn
	}
	if req.RiskCrit != nil {
		thresholds.RiskCrit = *req.RiskCrit
	}
	if req.ThermalWarn != nil {
		thresholds.ThermalWarn = *req.ThermalWarn
	}
	if req.ThermalCrit != nil {
		thresholds.ThermalCrit = *req.ThermalCrit
	}
	if req.BufferWarn != nil {
		thresholds.BufferWarn = *req.BufferWarn
	}
	if req.BufferCrit != nil {
		thresholds.BufferCrit = *req.BufferCrit
	}

	// Validate the merged thresholds
	if err := thresholds.Validate(); err != nil {
		return nil, err
	}

	settings.DefaultThresholds = thresholds

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// DeleteSettings deletes organization settings.
func (s *OrganizationSettingsService) DeleteSettings(ctx context.Context, orgID string) error {
	err := s.settingsRepo.DeleteByOrganizationID(ctx, orgID)
	if err != nil && !errors.Is(err, organization.ErrSettingsNotFound) {
		return err
	}
	return nil
}
