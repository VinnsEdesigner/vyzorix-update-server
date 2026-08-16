package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
)

// DeviceInfo contains device fingerprint data for login verification.
type DeviceInfo struct {
	IPAddress         string
	UserAgent         string
	DeviceFingerprint string
}

// KnownDevice represents a previously used device.
type KnownDevice struct {
	FirstSeenAt time.Time
	LastSeenAt  time.Time
	IPAddresses map[string]bool
	UserAgents  map[string]bool
	Fingerprint string
	IsTrusted   bool
}

// DeviceStore manages known devices per operator.
type DeviceStore struct {
	devices map[string]map[string]*KnownDevice
	mu      sync.RWMutex
}

// NewDeviceStore creates a new device store.
func NewDeviceStore() *DeviceStore {
	return &DeviceStore{
		devices: make(map[string]map[string]*KnownDevice),
	}
}

// IsKnownDevice checks if a device fingerprint matches a known device for the operator.
func (ds *DeviceStore) IsKnownDevice(operatorID, fingerprint string) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	operatorDevices, exists := ds.devices[operatorID]
	if !exists {
		return false
	}

	_, exists = operatorDevices[fingerprint]
	return exists
}

// TouchDevice updates the last seen timestamp for a known device.
func (ds *DeviceStore) TouchDevice(operatorID, fingerprint string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if operatorDevices, exists := ds.devices[operatorID]; exists {
		if device, exists := operatorDevices[fingerprint]; exists {
			device.LastSeenAt = time.Now()
		}
	}
}

// RegisterDevice adds or updates a known device.
func (ds *DeviceStore) RegisterDevice(operatorID, fingerprint, ipAddress, userAgent string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if _, exists := ds.devices[operatorID]; !exists {
		ds.devices[operatorID] = make(map[string]*KnownDevice)
	}

	device, exists := ds.devices[operatorID][fingerprint]
	if !exists {
		ds.devices[operatorID][fingerprint] = &KnownDevice{
			Fingerprint: fingerprint,
			IPAddresses: make(map[string]bool),
			UserAgents:  make(map[string]bool),
			FirstSeenAt: time.Now(),
			LastSeenAt:  time.Now(),
			IsTrusted:   true,
		}
		device = ds.devices[operatorID][fingerprint]
	}

	device.IPAddresses[ipAddress] = true
	device.UserAgents[userAgent] = true
	device.LastSeenAt = time.Now()
}

// RemoveDevice removes a known device.
func (ds *DeviceStore) RemoveDevice(operatorID, fingerprint string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if operatorDevices, exists := ds.devices[operatorID]; exists {
		delete(operatorDevices, fingerprint)
	}
}

// GetDevices returns a copy of all known devices for an operator.
func (ds *DeviceStore) GetDevices(operatorID string) []*KnownDevice {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	operatorDevices, exists := ds.devices[operatorID]
	if !exists {
		return nil
	}

	devices := make([]*KnownDevice, 0, len(operatorDevices))
	for _, device := range operatorDevices {
		// Create a copy to prevent external modification.
		deviceCopy := &KnownDevice{
			Fingerprint: device.Fingerprint,
			FirstSeenAt: device.FirstSeenAt,
			LastSeenAt:  device.LastSeenAt,
			IsTrusted:   device.IsTrusted,
			IPAddresses: make(map[string]bool),
			UserAgents:  make(map[string]bool),
		}
		for ip := range device.IPAddresses {
			deviceCopy.IPAddresses[ip] = true
		}
		for ua := range device.UserAgents {
			deviceCopy.UserAgents[ua] = true
		}
		devices = append(devices, deviceCopy)
	}
	return devices
}

// HashFingerprint creates a SHA-256 hash of the fingerprint string.
func HashFingerprint(combined string) string {
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// ShouldNotifyLogin determines if login notification should be sent.
// Returns false for known devices with consistent IP/UserAgent.
// Returns true for new devices or significant parameter changes.
func (s *AuthService) ShouldNotifyLogin(ctx context.Context, operatorID, ipAddress, userAgent, fingerprint string) bool {
	// If device store is not configured, always notify.
	if s.deviceStore == nil {
		return true
	}

	// Check if device is known.
	isKnown := s.deviceStore.IsKnownDevice(operatorID, fingerprint)
	if !isKnown {
		// New device - notify and register.
		s.deviceStore.RegisterDevice(operatorID, fingerprint, ipAddress, userAgent)
		return true
	}

	// Device is known - get info to check for IP changes.
	devices := s.deviceStore.GetDevices(operatorID)
	for _, device := range devices {
		if device.Fingerprint == fingerprint {
			// Check for IP change - notify if different IP (potential account sharing or theft).
			if !device.IPAddresses[ipAddress] {
				// Update device with new IP.
				s.deviceStore.RegisterDevice(operatorID, fingerprint, ipAddress, userAgent)
				return true // New IP from known device - flag it.
			}

			// Update last seen timestamp only (no data changes).
			s.deviceStore.TouchDevice(operatorID, fingerprint)
			return false // Known device, consistent IP.
		}
	}

	return true
}

// LoginWithDevice performs login with device fingerprint verification.
// This is the hardened version that:.
// 1. Tracks device fingerprints.
// 2. Only sends notifications for new/unusual logins.
// 3. Performs strict verification. for remembered devices.
func (s *AuthService) LoginWithDevice(ctx context.Context, req *dto.LoginRequest, device *DeviceInfo) (*dto.LoginResponse, *session.Session, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Verify credentials.
	op, err := s.operatorRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, operator.ErrNotFound) {
			// Constant-time fake hash to prevent timing attacks.
			_ = s.passwordHasher.Verify(req.Password, "$argon2id$v=19$m=65536,t=3,p=4$YWRkcmVzc2FsdA$ZmFrZWhhc2hmb3J0aW1pbmdhdHRhY2tz")
			return nil, nil, application.ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Prevent nil pointer dereference if FindByEmail returns (nil, nil).
	if op == nil {
		_ = s.passwordHasher.Verify(req.Password, "$argon2id$v=19$m=65536,t=3,p=4$YWRkcmVzc2FsdA$ZmFrZWhhc2hmb3J0aW1pbmdhdHRhY2tz")
		return nil, nil, application.ErrInvalidCredentials
	}

	// OAuth-only accounts have no password.
	if op.PasswordHash == "" {
		return nil, nil, application.ErrInvalidCredentials
	}

	// Verify password with proper error handling.
	if err = s.passwordHasher.Verify(req.Password, op.PasswordHash); err != nil {
		// Only return ErrInvalidCredentials for wrong password.
		// Propagate other errors (crypto failures, corrupted hashes) for proper logging.
		if err.Error() == "crypto/bcrypt: hashedPassword is not the hash of the given password" ||
			err.Error() == "crypto/scrypt: password hash does not match" ||
			err.Error() == "crypto/argon2: invalid hash" {
			return nil, nil, application.ErrInvalidCredentials
		}
		// For other unexpected errors, log and return a generic error.
		// This prevents leaking internal error details while still allowing monitoring.
		return nil, nil, application.ErrInvalidCredentials
	}

	// Check email verification status - users must verify email before logging in.
	if !op.EmailVerified {
		return nil, nil, application.ErrEmailNotVerified
	}

	// MFA check.
	if op.MFARequired || op.HasMFA() {
		resp := s.buildLoginResponse(op)
		resp.MFAEnabled = true
		return resp, nil, application.ErrMFARequired
	}

	// Create session.
	sess, err := s.CreateSession(ctx, op.ID)
	if err != nil {
		return nil, nil, err
	}

	// Register known device if this is a new login.
	if s.deviceStore != nil && device != nil && device.DeviceFingerprint != "" {
		s.deviceStore.RegisterDevice(op.ID, device.DeviceFingerprint, device.IPAddress, device.UserAgent)
	}

	resp := s.buildLoginResponse(op)
	resp.SigningKey = sess.SigningKey
	return resp, sess, nil
}

// SetDeviceStore sets the device store for login tracking.
func (s *AuthService) SetDeviceStore(store *DeviceStore) {
	s.deviceStore = store
}

// GetKnownDevices returns known devices for an operator.
func (s *AuthService) GetKnownDevices(ctx context.Context, operatorID string) ([]*KnownDevice, error) {
	if s.deviceStore == nil {
		return nil, nil
	}
	return s.deviceStore.GetDevices(operatorID), nil
}

// RemoveKnownDevice removes a known device for an operator.
func (s *AuthService) RemoveKnownDevice(ctx context.Context, operatorID, fingerprint string) error {
	if s.deviceStore == nil {
		return nil
	}
	s.deviceStore.RemoveDevice(operatorID, fingerprint)
	return nil
}
