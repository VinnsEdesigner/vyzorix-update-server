package device

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

var (
	ErrDeviceHijack = errors.New("device registration hijack detected")
)

// Service handles device operations.
type Service struct {
	deviceRepo  device.Repository
	operatorRepo operator.Repository
}

// NewService creates a new DeviceService.
func NewService(
	deviceRepo device.Repository,
	operatorRepo operator.Repository,
) *Service {
	return &Service{
		deviceRepo:  deviceRepo,
		operatorRepo: operatorRepo,
	}
}

// Register registers a new device.
func (s *Service) Register(ctx context.Context, req *dto.RegisterDeviceRequest) (*dto.RegisterDeviceResponse, error) {
	// Check if device already exists.
	existing, err := s.deviceRepo.FindByID(ctx, req.DeviceID)
	if err != nil && err != device.ErrNotFound {
		return nil, err
	}

	if existing != nil {
		// Device exists - check for hijacking.
		if existing.FirebaseInstallID != req.FirebaseInstallID {
			return nil, ErrDeviceHijack
		}

		// Update existing device.
		existing.FCMToken = req.FCMToken
		existing.AppVersion = req.AppVersion
		existing.DeviceClass = req.DeviceClass
		existing.Online = true
		existing.LastSeen = time.Now().UnixMilli()

		if err := s.deviceRepo.Update(ctx, existing); err != nil {
			return nil, err
		}

		// Return existing command secret (we don't regenerate it).
		return &dto.RegisterDeviceResponse{
			DeviceID:     existing.ID,
			CommandSecret: "", // Don't return secret on re-registration
			RegisteredAt:  existing.RegisteredAt,
		}, nil
	}

	// Generate command secret for new device.
	commandSecret, err := shared.GenerateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	d := &device.Device{
		ID:                req.DeviceID,
		FirebaseInstallID: req.FirebaseInstallID,
		FCMToken:         req.FCMToken,
		AppVersion:        req.AppVersion,
		DeviceClass:       req.DeviceClass,
		Online:           true,
		RegisteredAt:      now.UnixMilli(),
		LastSeen:         now.UnixMilli(),
		CreatedAt:        now,
		UpdatedAt:         now,
	}

	// Hash the command secret for storage.
	h := sha256.Sum256([]byte(commandSecret))
	d.CommandSecretHash = hex.EncodeToString(h[:])

	if err := s.deviceRepo.Create(ctx, d); err != nil {
		return nil, err
	}

	return &dto.RegisterDeviceResponse{
		DeviceID:     d.ID,
		CommandSecret: commandSecret,
		RegisteredAt:  d.RegisteredAt,
	}, nil
}

// GetStatus retrieves device status.
func (s *Service) GetStatus(ctx context.Context, deviceID string) (*dto.DeviceStatusResponse, error) {
	d, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		if err == device.ErrNotFound {
			return nil, application.ErrDeviceNotFound
		}
		return nil, err
	}

	return &dto.DeviceStatusResponse{
		DeviceID:    d.ID,
		Online:      d.Online,
		LastSeen:    d.LastSeen,
		AppVersion:  d.AppVersion,
		DeviceClass: d.DeviceClass,
	}, nil
}

// GetStatusByOperator retrieves device status with DOA verification.
// Only returns device if it belongs to the specified operator.
func (s *Service) GetStatusByOperator(ctx context.Context, deviceID, operatorID string) (*dto.DeviceStatusResponse, error) {
	d, err := s.deviceRepo.FindByIDAndOperator(ctx, deviceID, operatorID)
	if err != nil {
		if err == device.ErrNotFound {
			return nil, application.ErrDeviceNotFound
		}
		return nil, err
	}

	return &dto.DeviceStatusResponse{
		DeviceID:    d.ID,
		Online:      d.Online,
		LastSeen:    d.LastSeen,
		AppVersion:  d.AppVersion,
		DeviceClass: d.DeviceClass,
	}, nil
}

// GetDeviceByOperator retrieves a device with DOA verification.
func (s *Service) GetDeviceByOperator(ctx context.Context, deviceID, operatorID string) (*dto.DeviceResponse, error) {
	d, err := s.deviceRepo.FindByIDAndOperator(ctx, deviceID, operatorID)
	if err != nil {
		if err == device.ErrNotFound {
			return nil, application.ErrDeviceNotFound
		}
		return nil, err
	}

	return &dto.DeviceResponse{
		ID:                d.ID,
		FirebaseInstallID: d.FirebaseInstallID,
		AppVersion:        d.AppVersion,
		DeviceClass:       d.DeviceClass,
		Online:            d.Online,
		RegisteredAt:      d.RegisteredAt,
		LastSeen:          d.LastSeen,
	}, nil
}

// UpdateFCMToken updates the FCM token for a device.
func (s *Service) UpdateFCMToken(ctx context.Context, deviceID, fcmToken string) error {
	return s.deviceRepo.UpdateFCMToken(ctx, deviceID, fcmToken)
}

// UpdateFCMTokenAndReturn updates FCM token and returns the device.
func (s *Service) UpdateFCMTokenAndReturn(ctx context.Context, deviceID, fcmToken string) (*device.Device, error) {
	if err := s.deviceRepo.UpdateFCMToken(ctx, deviceID, fcmToken); err != nil {
		return nil, err
	}
	return s.deviceRepo.FindByID(ctx, deviceID)
}

// SetOnline sets the online status of a device.
func (s *Service) SetOnline(ctx context.Context, deviceID string, online bool) error {
	return s.deviceRepo.SetOnline(ctx, deviceID, online)
}

// List returns a paginated list of devices.
func (s *Service) List(ctx context.Context, limit, offset int) (*dto.DeviceListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	devices, total, err := s.deviceRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	response := &dto.DeviceListResponse{
		Devices: make([]dto.DeviceResponse, len(devices)),
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}

	for i, d := range devices {
		response.Devices[i] = dto.DeviceResponse{
			ID:                d.ID,
			FirebaseInstallID: d.FirebaseInstallID,
			AppVersion:        d.AppVersion,
			DeviceClass:       d.DeviceClass,
			Online:            d.Online,
			RegisteredAt:        d.RegisteredAt,
			LastSeen:          d.LastSeen,
		}
	}

	return response, nil
}

// ListByOperator returns all devices for an operator.
func (s *Service) ListByOperator(ctx context.Context, operatorID string) ([]dto.DeviceResponse, error) {
	devices, err := s.deviceRepo.ListByOperator(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.DeviceResponse, len(devices))
	for i, d := range devices {
		response[i] = dto.DeviceResponse{
			ID:                d.ID,
			FirebaseInstallID: d.FirebaseInstallID,
			AppVersion:        d.AppVersion,
			DeviceClass:       d.DeviceClass,
			Online:            d.Online,
			RegisteredAt:        d.RegisteredAt,
			LastSeen:          d.LastSeen,
		}
	}

	return response, nil
}

// ListByOperatorEntity returns all devices for an operator as domain entities.
func (s *Service) ListByOperatorEntity(ctx context.Context, operatorID string) ([]*device.Device, error) {
	return s.deviceRepo.ListByOperator(ctx, operatorID)
}

// Delete deletes a device.
func (s *Service) Delete(ctx context.Context, deviceID string) error {
	return s.deviceRepo.Delete(ctx, deviceID)
}

// DeleteDevice deletes a device (alias for Delete).
func (s *Service) DeleteDevice(ctx context.Context, deviceID string) error {
	return s.deviceRepo.Delete(ctx, deviceID)
}

// GetDevice retrieves a device by ID.
func (s *Service) GetDevice(ctx context.Context, deviceID string) (*device.Device, error) {
	return s.deviceRepo.FindByID(ctx, deviceID)
}

// Count returns the total number of devices.
func (s *Service) Count(ctx context.Context) (int, error) {
	return s.deviceRepo.Count(ctx)
}

// CountByOperator returns the total number of devices for an operator.
func (s *Service) CountByOperator(ctx context.Context, operatorID string) (int, error) {
	devices, err := s.deviceRepo.ListByOperator(ctx, operatorID)
	if err != nil {
		return 0, err
	}
	return len(devices), nil
}

// ListByOperatorPaginated returns paginated devices for an operator.
func (s *Service) ListByOperatorPaginated(ctx context.Context, operatorID string, limit, offset int) ([]*device.Device, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	allDevices, err := s.deviceRepo.ListByOperator(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	if offset >= len(allDevices) {
		return []*device.Device{}, nil
	}

	end := offset + limit
	if end > len(allDevices) {
		end = len(allDevices)
	}

	return allDevices[offset:end], nil
}
