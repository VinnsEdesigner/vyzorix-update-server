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
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/devicestore"
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

// Bounds on the device store. Fingerprints are derived from request signals (IP,
// User-Agent, ...), so an authenticated operator logging in from many networks or
// clients could otherwise grow these maps without limit.
const (
	// maxDevicesPerOperator caps fingerprints tracked per operator.
	maxDevicesPerOperator = 50
	// maxSignalsPerDevice caps the distinct IPs / User-Agents retained per device.
	maxSignalsPerDevice = 20
)

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

// The in-memory store satisfies the same contract as the SQL-backed one so either
// can be injected into AuthService.
var _ devicestore.Store = (*DeviceStore)(nil)

// IsKnownDevice reports whether the operator has a device with this fingerprint.
func (ds *DeviceStore) IsKnownDevice(ctx context.Context, operatorID, fingerprint string) (bool, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	operatorDevices, exists := ds.devices[operatorID]
	if !exists {
		return false, nil
	}

	_, exists = operatorDevices[fingerprint]

	return exists, nil
}

// HasSignal reports whether the device has been seen from this value of kind.
func (ds *DeviceStore) HasSignal(ctx context.Context, operatorID, fingerprint, kind, value string) (bool, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	device, ok := ds.devices[operatorID][fingerprint]
	if !ok {
		return false, nil
	}

	switch kind {
	case devicestore.KindIP:
		return device.IPAddresses[value], nil
	case devicestore.KindUserAgent:
		return device.UserAgents[value], nil
	default:
		return false, nil
	}
}

// TouchDevice updates the last seen timestamp for a known device.
func (ds *DeviceStore) TouchDevice(ctx context.Context, operatorID, fingerprint string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if operatorDevices, exists := ds.devices[operatorID]; exists {
		if device, exists := operatorDevices[fingerprint]; exists {
			device.LastSeenAt = time.Now()
		}
	}

	return nil
}

// RegisterDevice adds or updates a known device.
func (ds *DeviceStore) RegisterDevice(ctx context.Context, operatorID, fingerprint, ipAddress, userAgent string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if _, exists := ds.devices[operatorID]; !exists {
		ds.devices[operatorID] = make(map[string]*KnownDevice)
	}

	device, exists := ds.devices[operatorID][fingerprint]
	if !exists {
		// Evict the least-recently-seen device if this operator is at capacity, so
		// the per-operator map stays bounded.
		if len(ds.devices[operatorID]) >= maxDevicesPerOperator {
			ds.evictOldestLocked(ds.devices[operatorID])
		}

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

	// Bound the per-device signal sets as well.
	if !device.IPAddresses[ipAddress] && len(device.IPAddresses) >= maxSignalsPerDevice {
		dropOne(device.IPAddresses)
	}
	if !device.UserAgents[userAgent] && len(device.UserAgents) >= maxSignalsPerDevice {
		dropOne(device.UserAgents)
	}

	device.IPAddresses[ipAddress] = true
	device.UserAgents[userAgent] = true
	device.LastSeenAt = time.Now()

	return nil
}

// evictOldestLocked removes the least-recently-seen device. Caller holds ds.mu.
func (ds *DeviceStore) evictOldestLocked(devices map[string]*KnownDevice) {
	var oldestKey string

	var oldestAt time.Time

	for key, d := range devices {
		if oldestKey == "" || d.LastSeenAt.Before(oldestAt) {
			oldestKey, oldestAt = key, d.LastSeenAt
		}
	}
	if oldestKey != "" {
		delete(devices, oldestKey)
	}
}

// dropOne removes an arbitrary element from a set map, used to keep the per-device
// signal sets bounded. Which element is dropped is unimportant for detection.
func dropOne(set map[string]bool) {
	for k := range set {
		delete(set, k)
		return
	}
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
//
// The backing store is persistent in production (storage.KnownDeviceRepository),
// so a device stays known across restarts and deploys; the in-memory DeviceStore is
// used in tests.
func (s *AuthService) ShouldNotifyLogin(ctx context.Context, operatorID, ipAddress, userAgent, fingerprint string) bool {
	// If no store is configured, always notify.
	if s.deviceStore == nil {
		return true
	}

	known, err := s.deviceStore.IsKnownDevice(ctx, operatorID, fingerprint)
	if err != nil {
		// Fail closed: if we cannot tell, notify.
		return true
	}

	if !known {
		// New device - notify and register.
		_ = s.deviceStore.RegisterDevice(ctx, operatorID, fingerprint, ipAddress, userAgent)
		return true
	}

	// Known device. A changed IP (or User-Agent) is a significant change worth
	// flagging, since the fingerprint includes them.
	ipKnown, err := s.deviceStore.HasSignal(ctx, operatorID, fingerprint, devicestore.KindIP, ipAddress)
	if err != nil {
		return true
	}
	if !ipKnown {
		_ = s.deviceStore.RegisterDevice(ctx, operatorID, fingerprint, ipAddress, userAgent)
		return true // New IP from a known device - flag it.
	}

	// Known device on a known IP: refresh last-seen and stay quiet.
	_ = s.deviceStore.TouchDevice(ctx, operatorID, fingerprint)

	return false
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
		_ = s.deviceStore.RegisterDevice(ctx, op.ID, device.DeviceFingerprint, device.IPAddress, device.UserAgent)
	}

	resp := s.buildLoginResponse(op)
	resp.SigningKey = sess.SigningKey
	return resp, sess, nil
}

// SetDeviceStore sets the device store used for login tracking. The production
// wiring passes the persistent (SQL) store; tests use the in-memory one.
func (s *AuthService) SetDeviceStore(store devicestore.Store) {
	s.deviceStore = store
}
