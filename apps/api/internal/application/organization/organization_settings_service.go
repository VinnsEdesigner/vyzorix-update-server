package organization

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	configversionapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/configversion"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/configversion"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// OrganizationSettingsService handles organization settings operations.
type OrganizationSettingsService struct {
	settingsRepo organization.OrganizationSettingsRepository
	orgRepo      organization.OrganizationRepository
	memberRepo   organization.MemberRepository
	versionSvc   *configversionapp.Service
}

// SetVersionService wires the config versioning service for snapshot-on-write.
func (s *OrganizationSettingsService) SetVersionService(v *configversionapp.Service) {
	s.versionSvc = v
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
	// Verify organization exists.
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

	// Check if settings already exist.
	existing, err := s.settingsRepo.FindByOrganizationID(ctx, orgID)
	if err != nil && !errors.Is(err, organization.ErrSettingsNotFound) {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Already exists, return it.
	}

	// Create new settings with defaults.
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

	// Create with defaults.
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

	// Apply updates.
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
	s.snapshot(ctx, settings, "operator")
	return settings, nil
}

// snapshot captures the current settings as a new config version.
func (s *OrganizationSettingsService) snapshot(ctx context.Context, settings *organization.OrganizationSettings, changedBy string) {
	if s.versionSvc == nil {
		return
	}
	doc := map[string]interface{}{
		"OrganizationID":       settings.OrganizationID,
		"Timezone":             settings.Timezone,
		"DateFormat":           settings.DateFormat,
		"AlertCooldownMinutes": settings.AlertCooldownMinutes,
		"DefaultThresholds":    settings.DefaultThresholds,
		"UpdatedAt":            settings.UpdatedAt.UnixMilli(),
	}
	if _, err := s.versionSvc.Snapshot(ctx, settings.OrganizationID, configversion.ResourceTypeSettings, doc, changedBy); err != nil {
		slog.Error("config version snapshot failed", "org_id", settings.OrganizationID, "error", err)
	}
}

// RestoreSettings re-applies a snapshot from config versions to live settings.
// The snapshot is the JSON body of an OrganizationSettings.
func (s *OrganizationSettingsService) RestoreSettings(ctx context.Context, orgID, snapshot string) (*organization.OrganizationSettings, error) {
	settings, err := s.settingsRepo.FindByOrganizationID(ctx, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrSettingsNotFound) {
			return nil, organization.ErrSettingsNotFound
		}
		return nil, err
	}

	var doc struct {
		DefaultThresholds    map[string]interface{} `json:"DefaultThresholds"`
		Timezone             string                 `json:"Timezone"`
		DateFormat           string                 `json:"DateFormat"`
		AlertCooldownMinutes int                    `json:"AlertCooldownMinutes"`
	}
	if err := json.Unmarshal([]byte(snapshot), &doc); err != nil {
		return nil, err
	}

	settings.Timezone = doc.Timezone
	settings.DateFormat = doc.DateFormat
	settings.AlertCooldownMinutes = doc.AlertCooldownMinutes
	if doc.DefaultThresholds != nil {
		raw, err := json.Marshal(doc.DefaultThresholds)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &settings.DefaultThresholds)
	}
	settings.UpdatedAt = time.Now()

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}
	s.snapshot(ctx, settings, "operator")
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

	// Merge with existing thresholds.
	thresholds := settings.DefaultThresholds
	if thresholds == nil {
		thresholds = organization.DefaultThresholds()
	}

	// Apply updates from request.
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

	// Validate the merged thresholds.
	if err := thresholds.Validate(); err != nil {
		return nil, err
	}

	settings.DefaultThresholds = thresholds

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}
	s.snapshotThresholds(ctx, settings, "operator")
	return settings, nil
}

// snapshotThresholds captures threshold state as its own resource type.
func (s *OrganizationSettingsService) snapshotThresholds(ctx context.Context, settings *organization.OrganizationSettings, changedBy string) {
	if s.versionSvc == nil {
		return
	}
	if _, err := s.versionSvc.Snapshot(ctx, settings.OrganizationID, configversion.ResourceTypeThresholds, settings.DefaultThresholds, changedBy); err != nil {
		slog.Error("config version threshold snapshot failed", "org_id", settings.OrganizationID, "error", err)
	}
}

// DeleteSettings deletes organization settings.
func (s *OrganizationSettingsService) DeleteSettings(ctx context.Context, orgID string) error {
	err := s.settingsRepo.DeleteByOrganizationID(ctx, orgID)
	if err != nil && !errors.Is(err, organization.ErrSettingsNotFound) {
		return err
	}
	return nil
}
