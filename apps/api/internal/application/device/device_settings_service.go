package device

import (
	"context"
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// DeviceSettingsService handles device settings operations.
type DeviceSettingsService struct {
	settingsRepo  device.DeviceSettingsRepository
	deviceRepo    Repository
	orgSettingsRepo organization.OrganizationSettingsRepository
}

// Repository is the interface for device data access.
type Repository interface {
	FindByIMEI(ctx context.Context, imei string) (*device.Device, error)
}

// NewDeviceSettingsService creates a new DeviceSettingsService.
func NewDeviceSettingsService(
	settingsRepo device.DeviceSettingsRepository,
	deviceRepo Repository,
	orgSettingsRepo organization.OrganizationSettingsRepository,
) *DeviceSettingsService {
	return &DeviceSettingsService{
		settingsRepo:  settingsRepo,
		deviceRepo:    deviceRepo,
		orgSettingsRepo: orgSettingsRepo,
	}
}

// SettingsRepo returns the settings repository for use by other services.
func (s *DeviceSettingsService) SettingsRepo() device.DeviceSettingsRepository {
	return s.settingsRepo
}

// CreateSettings creates device settings with defaults for a new device.
func (s *DeviceSettingsService) CreateSettings(ctx context.Context, deviceIMEI string) (*device.DeviceSettings, error) {
	// Verify device exists
	d, err := s.deviceRepo.FindByIMEI(ctx, deviceIMEI)
	if err != nil {
		if errors.Is(err, device.ErrNotFound) {
			return nil, device.ErrNotFound
		}
		return nil, err
	}

	if d.OrganizationID == "" {
		return nil, errors.New("device must belong to an organization before creating settings")
	}

	// Check if settings already exist
	existing, err := s.settingsRepo.FindByDeviceIMEI(ctx, deviceIMEI)
	if err != nil && !errors.Is(err, device.ErrSettingsNotFound) {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Already exists, return it
	}

	// Create new settings
	settings := device.NewDeviceSettings(deviceIMEI)

	if err := s.settingsRepo.Create(ctx, settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// GetSettings retrieves device settings.
func (s *DeviceSettingsService) GetSettings(ctx context.Context, deviceIMEI string) (*device.DeviceSettings, error) {
	settings, err := s.settingsRepo.FindByDeviceIMEI(ctx, deviceIMEI)
	if err != nil {
		if errors.Is(err, device.ErrSettingsNotFound) {
			return nil, device.ErrSettingsNotFound
		}
		return nil, err
	}

	return settings, nil
}

// GetOrCreateSettings retrieves settings, creating them with defaults if they don't exist.
func (s *DeviceSettingsService) GetOrCreateSettings(ctx context.Context, deviceIMEI string) (*device.DeviceSettings, error) {
	settings, err := s.settingsRepo.FindByDeviceIMEI(ctx, deviceIMEI)
	if err != nil && !errors.Is(err, device.ErrSettingsNotFound) {
		return nil, err
	}

	if settings != nil {
		return settings, nil
	}

	// Create with defaults
	return s.CreateSettings(ctx, deviceIMEI)
}

// UpdateSettings updates device settings.
func (s *DeviceSettingsService) UpdateSettings(ctx context.Context, deviceIMEI string, req *device.UpdateDeviceSettingsRequest) (*device.DeviceSettings, error) {
	settings, err := s.settingsRepo.FindByDeviceIMEI(ctx, deviceIMEI)
	if err != nil {
		if errors.Is(err, device.ErrSettingsNotFound) {
			return nil, device.ErrSettingsNotFound
		}
		return nil, err
	}

	// Apply updates
	if req.CustomName != nil {
		settings.CustomName = *req.CustomName
	}
	if req.Location != nil {
		settings.Location = *req.Location
	}
	if req.Metadata != nil {
		settings.Metadata = req.Metadata
	}
	if req.Thresholds != nil {
		if err := settings.UpdateThresholds(req.Thresholds); err != nil {
			return nil, err
		}
	}

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateThresholds updates only the thresholds.
func (s *DeviceSettingsService) UpdateThresholds(ctx context.Context, deviceIMEI string, req *device.UpdateThresholdsRequest) (*device.DeviceSettings, error) {
	settings, err := s.settingsRepo.FindByDeviceIMEI(ctx, deviceIMEI)
	if err != nil {
		if errors.Is(err, device.ErrSettingsNotFound) {
			return nil, device.ErrSettingsNotFound
		}
		return nil, err
	}

	// Merge with existing thresholds or create new
	thresholds := settings.Thresholds
	if thresholds == nil {
		thresholds = &device.Thresholds{}
	}

	// Apply updates from request (only update non-nil values)
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

	settings.Thresholds = thresholds

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// DeleteSettings deletes device settings.
func (s *DeviceSettingsService) DeleteSettings(ctx context.Context, deviceIMEI string) error {
	err := s.settingsRepo.DeleteByDeviceIMEI(ctx, deviceIMEI)
	if err != nil && !errors.Is(err, device.ErrSettingsNotFound) {
		return err
	}
	return nil
}

// GetEffectiveThresholds returns the effective thresholds for a device using the hierarchy:
// device settings → organization settings → default thresholds
func (s *DeviceSettingsService) GetEffectiveThresholds(ctx context.Context, deviceIMEI string) (*device.Thresholds, error) {
	// Get device to find organization
	d, err := s.deviceRepo.FindByIMEI(ctx, deviceIMEI)
	if err != nil {
		return nil, err
	}

	// Get device settings
	var deviceSettings *device.DeviceSettings
	deviceSettings, err = s.settingsRepo.FindByDeviceIMEI(ctx, deviceIMEI)
	if err != nil && !errors.Is(err, device.ErrSettingsNotFound) {
		return nil, err
	}

	// Get organization settings
	var orgThresholds *device.Thresholds
	if d.OrganizationID != "" {
		orgSettings, err := s.orgSettingsRepo.FindByOrganizationID(ctx, d.OrganizationID)
		if err != nil && !errors.Is(err, organization.ErrSettingsNotFound) {
			return nil, err
		}
		if orgSettings != nil && orgSettings.DefaultThresholds != nil {
			orgThresholds = &device.Thresholds{
				RiskWarn:    orgSettings.DefaultThresholds.RiskWarn,
				RiskCrit:    orgSettings.DefaultThresholds.RiskCrit,
				ThermalWarn: orgSettings.DefaultThresholds.ThermalWarn,
				ThermalCrit: orgSettings.DefaultThresholds.ThermalCrit,
				BufferWarn:  orgSettings.DefaultThresholds.BufferWarn,
				BufferCrit:  orgSettings.DefaultThresholds.BufferCrit,
			}
		}
	}

	// Resolve using hierarchy
	return device.ResolveThresholds(deviceSettings, orgThresholds), nil
}
