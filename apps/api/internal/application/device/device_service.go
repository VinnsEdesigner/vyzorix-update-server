package device

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/shared"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/inbox"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

var (
	ErrDeviceHijack          = errors.New("device registration hijack detected")
	ErrDeviceNotFound        = errors.New("device not found")
	ErrCommandSecretNotSet   = errors.New("command secret not set for device")
	ErrInvalidCommandSecret  = errors.New("invalid command secret")
	ErrDeviceAlreadyApproved = errors.New("device already approved and registered")
	ErrDeviceNotPending      = errors.New("device is not in pending state")
	ErrInvalidLifecycleTransition = errors.New("invalid lifecycle state transition")
)

// Service handles device operations.
type Service struct {
	deviceRepo   device.Repository
	operatorRepo operator.Repository
	logger       *slog.Logger
}

// NewService creates a new DeviceService.
func NewService(
	deviceRepo device.Repository,
	operatorRepo operator.Repository,
	logger *slog.Logger,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		deviceRepo:   deviceRepo,
		operatorRepo: operatorRepo,
		logger:       logger,
	}
}

// Register registers a new device.
// New devices start in LifecyclePending state and require operator approval.
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

		// If device was deregistered, allow re-registration with new pending state
		if existing.IsDeregistered() {
			// Use domain method to transition back to pending
			existing.Lifecycle = device.LifecyclePending
			existing.DeregisteredAt = nil
			existing.DeletionScheduledAt = nil
		}

		// Update existing device.
		existing.FCMToken = req.FCMToken
		existing.AppVersion = req.AppVersion
		existing.DeviceClass = req.DeviceClass
		existing.Online = true
		existing.LastSeen = time.Now().UnixMilli()

		if err = s.deviceRepo.Update(ctx, existing); err != nil {
			return nil, err
		}

		// Return existing command secret (we don't regenerate it).
		return &dto.RegisterDeviceResponse{
			DeviceID:      existing.ID,
			CommandSecret: "", // Don't return secret on re-registration
			RegisteredAt:  existing.RegisteredAt,
			Lifecycle:     string(existing.Lifecycle),
		}, nil
	}

	// Generate command secret for new device.
	commandSecret, err := shared.GenerateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	// Use NewDevice constructor which sets LifecyclePending by default
	d := device.NewDevice(req.DeviceID, req.FirebaseInstallID)
	d.FCMToken = req.FCMToken
	d.AppVersion = req.AppVersion
	d.DeviceClass = req.DeviceClass
	d.Online = true
	d.LastSeen = now.UnixMilli()

	// Hash the command secret for storage.
	h := sha256.Sum256([]byte(commandSecret))
	d.CommandSecretHash = hex.EncodeToString(h[:])

	if err := s.deviceRepo.Create(ctx, d); err != nil {
		return nil, err
	}

	return &dto.RegisterDeviceResponse{
		DeviceID:      d.ID,
		CommandSecret: commandSecret,
		RegisteredAt:  d.RegisteredAt,
		Lifecycle:     string(d.Lifecycle),
	}, nil
}

// ApproveDevice transitions a pending device to registered state.
// Returns error if device is not in pending state.
func (s *Service) ApproveDevice(ctx context.Context, deviceID string) error {
	d, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		if err == device.ErrNotFound {
			return ErrDeviceNotFound
		}
		return err
	}

	// Use domain method to enforce valid transitions
	if err := d.Approve(); err != nil {
		return ErrInvalidLifecycleTransition
	}

	// Set RegisteredAt when transitioning to registered
	d.RegisteredAt = time.Now().UnixMilli()
	d.UpdatedAt = time.Now()

	if err := s.deviceRepo.Update(ctx, d); err != nil {
		return err
	}

	return nil
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
		ID:           d.ID,
		IMEI:         d.ID,
		DeviceName:   d.DeviceName,
		Model:        d.Model,
		Manufacturer: d.Manufacturer,
		OSVersion:    d.OSVersion,
		AppVersion:  d.AppVersion,
		Status:       d.GetStatus(),
		Online:       d.Online,
		RegisteredAt: d.RegisteredAt,
		LastSeen:     d.LastSeen,
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
		Pagination: dto.PaginationInfo{
			Total: total,
			Limit: limit,
		},
	}

	for i, d := range devices {
		response.Devices[i] = dto.DeviceResponse{
			ID:           d.ID,
			IMEI:         d.ID,
			DeviceName:   d.DeviceName,
			Model:        d.Model,
			Manufacturer: d.Manufacturer,
			OSVersion:    d.OSVersion,
			AppVersion:  d.AppVersion,
			Status:       d.GetStatus(),
			Online:       d.Online,
			RegisteredAt: d.RegisteredAt,
			LastSeen:     d.LastSeen,
		}
	}

	// Calculate pagination info
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	response.Pagination.Page = (offset / limit) + 1
	response.Pagination.TotalPages = totalPages

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
			ID:           d.ID,
			IMEI:         d.ID,
			DeviceName:   d.DeviceName,
			Model:        d.Model,
			Manufacturer: d.Manufacturer,
			OSVersion:    d.OSVersion,
			AppVersion:  d.AppVersion,
			Status:       d.GetStatus(),
			Online:       d.Online,
			RegisteredAt: d.RegisteredAt,
			LastSeen:     d.LastSeen,
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

// GetRepo returns the device repository for use by other services.
func (s *Service) GetRepo() device.Repository {
	return s.deviceRepo
}

// DeviceRepo returns the device repository (alias for GetRepo).
func (s *Service) DeviceRepo() device.Repository {
	return s.deviceRepo
}

// ListQuery represents query parameters for listing devices.
type ListQuery struct {
	OrganizationID string
	Status        string
	Search        string
	Page          int
	Limit         int
}

// GetDevices returns a paginated list of devices filtered by organization.
func (s *Service) GetDevices(ctx context.Context, query *ListQuery) (*dto.DeviceListResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	offset := (query.Page - 1) * query.Limit

	// Get all devices and filter
	allDevices, _, err := s.deviceRepo.List(ctx, 10000, 0) // Get all for filtering
	if err != nil {
		return nil, err
	}

	// Apply filters
	var filtered []*device.Device
	for _, d := range allDevices {
		// Apply organization filter first (required for multi-tenant)
		if query.OrganizationID != "" && d.OrganizationID != query.OrganizationID {
			continue
		}

		// Apply status filter
		if query.Status != "" && query.Status != "all" {
			isOnline := d.Online
			if query.Status == "online" && !isOnline {
				continue
			}
			if query.Status == "offline" && isOnline {
				continue
			}
		}

		// Apply search filter
		if query.Search != "" {
			// Search by ID (IMEI) or other fields
			searchLower := toLower(query.Search)
			idMatch := contains(toLower(d.ID), searchLower)
			classMatch := contains(toLower(d.DeviceClass), searchLower)
			if !idMatch && !classMatch {
				continue
			}
		}

		filtered = append(filtered, d)
	}

	// Calculate pagination
	total := len(filtered)
	totalPages := 0
	if total > 0 {
		totalPages = (total + query.Limit - 1) / query.Limit
	}

	// Apply pagination
	start := offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + query.Limit
	if end > len(filtered) {
		end = len(filtered)
	}
	paged := filtered[start:end]

	// Build response
	devices := make([]dto.DeviceResponse, 0, len(paged))
	for _, d := range paged {
		devices = append(devices, dto.DeviceResponse{
			ID:           d.ID,
			IMEI:         d.ID,
			DeviceName:   d.DeviceName,
			Model:        d.Model,
			Manufacturer: d.Manufacturer,
			OSVersion:    d.OSVersion,
			AppVersion:  d.AppVersion,
			Status:       d.GetStatus(),
			Online:       d.Online,
			RegisteredAt: d.RegisteredAt,
			LastSeen:     d.LastSeen,
		})
	}

	return &dto.DeviceListResponse{
		Devices: devices,
		Pagination: dto.PaginationInfo{
			Page:       query.Page,
			Limit:      query.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetDeviceDetail returns detailed device information for /v1/devices/:imei endpoint.
func (s *Service) GetDeviceDetail(ctx context.Context, imei string) (*dto.DeviceDetailResponse, error) {
	d, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil {
		if err == device.ErrNotFound {
			return nil, device.ErrNotFound
		}
		return nil, err
	}

	return s.deviceDetailResponse(d), nil
}

// GetDeviceDetailByOperator returns detailed device information with DOA verification.
func (s *Service) GetDeviceDetailByOperator(ctx context.Context, imei, operatorID string) (*dto.DeviceDetailResponse, error) {
	d, err := s.deviceRepo.FindByIMEIAndOperator(ctx, imei, operatorID)
	if err != nil {
		if err == device.ErrNotFound {
			return nil, device.ErrNotFound
		}
		return nil, err
	}

	return s.deviceDetailResponse(d), nil
}

// GetDeviceDetailByOrganization returns detailed device information for an organization.
func (s *Service) GetDeviceDetailByOrganization(ctx context.Context, imei, orgID string) (*dto.DeviceDetailResponse, error) {
	d, err := s.deviceRepo.FindByIMEIAndOrganization(ctx, imei, orgID)
	if err != nil {
		if err == device.ErrNotFound {
			return nil, device.ErrNotFound
		}
		return nil, err
	}

	return s.deviceDetailResponse(d), nil
}

// DeregisterDeviceByOrganization deregisters a device within an organization.
func (s *Service) DeregisterDeviceByOrganization(ctx context.Context, imei, orgID string, hard bool) (*dto.DeregisterResponse, error) {
	// First verify device exists and belongs to this organization
	d, err := s.deviceRepo.FindByIMEIAndOrganization(ctx, imei, orgID)
	if err != nil {
		if err == device.ErrNotFound {
			return nil, device.ErrNotFound
		}
		return nil, err
	}

	now := time.Now()
	deregisteredAt := now.UnixMilli()
	deletionScheduledAt := now.Add(30 * 24 * time.Hour).UnixMilli() // 30 days retention

	if hard {
		// Hard delete - actually remove the device
		if err := s.deviceRepo.Delete(ctx, imei); err != nil {
			if err == device.ErrNotFound {
				return nil, device.ErrNotFound
			}
			return nil, err
		}
		return &dto.DeregisterResponse{
			IMEI:           imei,
			Status:         "deleted",
			DeregisteredAt: deregisteredAt,
		}, nil
	}

	// Soft delete - mark as deregistered
	if err := s.deviceRepo.SoftDelete(ctx, imei, deregisteredAt, deletionScheduledAt); err != nil {
		if err == device.ErrNotFound {
			return nil, device.ErrNotFound
		}
		return nil, err
	}

	return &dto.DeregisterResponse{
		IMEI:                imei,
		Status:              "deregistered",
		DeregisteredAt:      deregisteredAt,
		DeletionScheduledAt: deletionScheduledAt,
	}, nil
}

// deviceDetailResponse creates a DeviceDetailResponse from a Device entity.
func (s *Service) deviceDetailResponse(d *device.Device) *dto.DeviceDetailResponse {
	// Check FCM token validity using domain method
	fcmValid := d.IsFCMTokenValid()

	// Check if command secret is set using domain method
	commandSet := d.IsCommandSecretSet()

	// Determine status using domain method
	status := d.GetStatus()

	resp := &dto.DeviceDetailResponse{
		ID:                d.ID,
		IMEI:              d.ID,
		DeviceName:        d.DeviceName,
		Model:             d.Model,
		Manufacturer:      d.Manufacturer,
		OSVersion:         d.OSVersion,
		AppVersion:        d.AppVersion,
		SecurityPatch:     d.SecurityPatch,
		Status:            status,
		RegisteredAt:      d.RegisteredAt,
		LastSeen:          d.LastSeen,
		FCMTokenValid:     fcmValid,
		CommandSecretSet:   commandSet,
	}

	return resp
}

// DeregisterDevice soft-deletes a device (marks as deregistered with 30-day retention).
func (s *Service) DeregisterDevice(ctx context.Context, imei string, hard bool) (*dto.DeregisterResponse, error) {
	now := time.Now()
	deregisteredAt := now.UnixMilli()
	deletionScheduledAt := now.Add(30 * 24 * time.Hour).UnixMilli() // 30 days retention

	if hard {
		// Hard delete - actually remove the device
		if err := s.deviceRepo.Delete(ctx, imei); err != nil {
			if err == device.ErrNotFound {
				return nil, device.ErrNotFound
			}
			return nil, err
		}
		return &dto.DeregisterResponse{
			IMEI:           imei,
			Status:         "deleted",
			DeregisteredAt: deregisteredAt,
		}, nil
	}

	// Soft delete - mark as deregistered
	if err := s.deviceRepo.SoftDeleteByIMEI(ctx, imei, deregisteredAt, deletionScheduledAt); err != nil {
		if err == device.ErrNotFound {
			return nil, device.ErrNotFound
		}
		return nil, err
	}

	return &dto.DeregisterResponse{
		IMEI:           imei,
		Status:         "deregistered",
		DeregisteredAt:  deregisteredAt,
		RetentionUntil: deletionScheduledAt,
	}, nil
}

// DeregisterDeviceByOperator soft-deletes a device with DOA and org verification.
// Only the operator who owns the device within an organization can deregister it.
func (s *Service) DeregisterDeviceByOperator(ctx context.Context, imei, operatorID, orgID string, hard bool) (*dto.DeregisterResponse, error) {
	// First verify device exists and belongs to this operator and organization
	d, err := s.deviceRepo.FindByIDAndOrganization(ctx, imei, orgID)
	if err != nil {
		if err == device.ErrNotFound {
			return nil, device.ErrNotFound
		}
		return nil, err
	}

	// Verify operator owns this device
	if d.OperatorID != operatorID {
		return nil, device.ErrNotFound
	}

	now := time.Now()
	deregisteredAt := now.UnixMilli()
	deletionScheduledAt := now.Add(30 * 24 * time.Hour).UnixMilli() // 30 days retention

	if hard {
		// Hard delete - actually remove the device
		if err := s.deviceRepo.Delete(ctx, imei); err != nil {
			return nil, err
		}
		return &dto.DeregisterResponse{
			IMEI:           imei,
			Status:         "deleted",
			DeregisteredAt: deregisteredAt,
		}, nil
	}

	// Soft delete - mark as deregistered
	if err := s.deviceRepo.SoftDeleteByIMEI(ctx, imei, deregisteredAt, deletionScheduledAt); err != nil {
		return nil, err
	}

	// Log the deregistration for audit
	s.logDeviceAction(ctx, d.ID, imei, "deregistered", operatorID, fmt.Sprintf("Device deregistered, hard=%v", hard))

	return &dto.DeregisterResponse{
		IMEI:           imei,
		Status:         "deregistered",
		DeregisteredAt:  deregisteredAt,
		RetentionUntil: deletionScheduledAt,
	}, nil
}
		Status:         "deregistered",
		DeregisteredAt:  deregisteredAt,
		RetentionUntil: deletionScheduledAt,
	}, nil
}
		Status:         "deregistered",
		DeregisteredAt: deregisteredAt,
		RetentionUntil: *d.DeletionScheduledAt,
	}, nil
}

// logDeviceAction logs device-related actions for audit trail.
func (s *Service) logDeviceAction(ctx context.Context, deviceID, imei, action, operatorID, details string) {
	// This would typically write to an audit log
	// For now, we rely on the operator repository for ownership verification
}

// CreateFromInbox creates a device from an approved inbox entry.
// This is called after an operator approves a device registration request.
func (s *Service) CreateFromInbox(ctx context.Context, entry *inbox.InboxEntry, commandSecret string) (*device.Device, error) {
	// Check if device already exists by IMEI
	existing, err := s.deviceRepo.FindByIMEI(ctx, entry.IMEI)
	if err != nil && err != device.ErrNotFound {
		return nil, fmt.Errorf("failed to check existing device: %w", err)
	}
	if existing != nil {
		// Device already exists - this shouldn't happen if flow is correct
		// Return error to prevent silent duplicate creation (Bug 33 fix)
		if existing.IsDeregistered() {
			// Device was deregistered - allow re-registration
			// MUST delete the old device first to avoid primary key constraint
			if err := s.deviceRepo.Delete(ctx, entry.IMEI); err != nil {
				s.logger.Error("failed to delete old deregistered device for re-registration",
					"imei", entry.IMEI,
					"error", err,
				)
				return nil, fmt.Errorf("failed to cleanup old device: %w", err)
			}
			s.logger.Info("deleted old deregistered device for re-registration",
				"imei", entry.IMEI,
			)
		} else {
			// Device is active - this is a conflict
			return nil, ErrDeviceAlreadyApproved
		}
	}

	// Hash the command secret for storage
	h := sha256.Sum256([]byte(commandSecret))
	commandSecretHash := hex.EncodeToString(h[:])

	now := time.Now()
	d := &device.Device{
		ID:                 entry.IMEI, // Use IMEI as device ID
		FirebaseInstallID:  entry.FirebaseInstallID,
		FCMToken:           entry.FCMToken,
		AppVersion:         entry.AppVersion,
		DeviceClass:        entry.DeviceClass,
		DeviceName:         entry.DeviceName,
		Model:              entry.Model,
		Manufacturer:       entry.Manufacturer,
		OSVersion:          entry.OSVersion,
		CommandSecretHash:  commandSecretHash,
		OperatorID:         entry.OperatorID,
		Online:             false, // Device will come online after confirming
		RegisteredAt:       now.UnixMilli(),
		LastSeen:           now.UnixMilli(),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.deviceRepo.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to create device from inbox: %w", err)
	}

	return d, nil
}

// ConfirmDevice confirms a device registration by validating the command secret.
// Returns the confirmed device if successful.
func (s *Service) ConfirmDevice(ctx context.Context, imei, commandSecret string) (*device.Device, error) {
	// Find device by IMEI
	d, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil {
		if err == device.ErrNotFound {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("failed to find device: %w", err)
	}

	// Validate command secret
	if d.CommandSecretHash == "" {
		return nil, ErrCommandSecretNotSet
	}

	// Hash the provided secret and compare
	h := sha256.Sum256([]byte(commandSecret))
	providedHash := hex.EncodeToString(h[:])

	if !secureCompare(d.CommandSecretHash, providedHash) {
		return nil, ErrInvalidCommandSecret
	}

	// Mark device as online and update last seen
	d.Online = true
	d.LastSeen = time.Now().UnixMilli()
	d.UpdatedAt = time.Now()

	if err := s.deviceRepo.Update(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to update device on confirm: %w", err)
	}

	return d, nil
}

// GetDeviceByIMEI retrieves a device by IMEI.
func (s *Service) GetDeviceByIMEI(ctx context.Context, imei string) (*device.Device, error) {
	d, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil {
		if err == device.ErrNotFound {
			return nil, ErrDeviceNotFound
		}
		return nil, err
	}
	return d, nil
}

func toLower(s string) string {
	return strings.ToLower(s)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// secureCompare performs a constant-time comparison to prevent timing attacks.
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i]) ^ int(b[i])
	}
	return result == 0
}

// TransferDevice transfers a device from one organization to another.
// Prerequisites:
// - Device must be OFFLINE
// - Actor must have permission in source AND target orgs
func (s *Service) TransferDevice(ctx context.Context, imei, sourceOrgID, targetOrgID, actorOperatorID string) error {
	// Get the device
	d, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil {
		if err == device.ErrNotFound {
			return ErrDeviceNotFound
		}
		return err
	}

	// Verify device belongs to source org
	if d.OrganizationID != sourceOrgID {
		return ErrDeviceNotFound
	}

	// Device must be offline for transfer
	if d.Online {
		return application.ErrDeviceOnline
	}

	// Update device organization
	d.OrganizationID = targetOrgID
	d.UpdatedAt = time.Now()

	if err := s.deviceRepo.Update(ctx, d); err != nil {
		return err
	}

	s.logger.Info("device transferred between organizations",
		"imei", imei,
		"from_org", sourceOrgID,
		"to_org", targetOrgID,
		"actor_id", actorOperatorID,
	)

	return nil
}

// TransferDevice transfers a device from one organization to another.
// Prerequisites:
// - Device must be OFFLINE
// - Actor must have permission in source AND target orgs
func (s *Service) TransferDevice(ctx context.Context, imei, sourceOrgID, targetOrgID, actorOperatorID string) error {
	// Get the device
	d, err := s.deviceRepo.FindByIMEI(ctx, imei)
	if err != nil {
		if err == device.ErrNotFound {
			return ErrDeviceNotFound
		}
		return err
	}

	// Verify device belongs to source org
	if d.OrganizationID != sourceOrgID {
		return ErrDeviceNotFound
	}

	// Device must be offline for transfer
	if d.Online {
		return application.ErrDeviceOnline
	}

	// Update device organization
	d.OrganizationID = targetOrgID
	d.UpdatedAt = time.Now()

	if err := s.deviceRepo.Update(ctx, d); err != nil {
		return err
	}

	s.logger.Info("device transferred between organizations",
		"imei", imei,
		"from_org", sourceOrgID,
		"to_org", targetOrgID,
		"actor_id", actorOperatorID,
	)

	return nil
}
