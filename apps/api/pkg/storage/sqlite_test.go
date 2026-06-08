package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

func TestOpen_createsDatabase(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	if _, err := os.Stat(path); err != nil {
		t.Error("database file was not created")
	}
}

func TestOpen_createsParentDirectories(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "subdir", "nested", "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	if _, err := os.Stat(path); err != nil {
		t.Error("database file was not created in nested directory")
	}
}

func TestPing(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Ping(ctx); err != nil {
		t.Errorf("Ping() failed: %v", err)
	}
}

func TestRegister_newDevice(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	d, isNew, err := store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc123",
		FCMToken:          "fcm-token-xyz",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}
	if !isNew {
		t.Error("expected isNew=true for first registration")
	}
	if d.ID != "device-001" {
		t.Errorf("DeviceID = %s, want device-001", d.ID)
	}
	if d.CommandSecret == "" {
		t.Error("CommandSecret should not be empty")
	}
}

func TestRegister_idempotent(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	req := models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc123",
		FCMToken:          "fcm-token-xyz",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	}

	// First registration
	d1, isNew1, err := store.Register(ctx, req)
	if err != nil {
		t.Fatalf("First Register() failed: %v", err)
	}
	if !isNew1 {
		t.Error("expected isNew=true for first registration")
	}

	// Same device re-registration (idempotent)
	d2, isNew2, err := store.Register(ctx, req)
	if err != nil {
		t.Fatalf("Second Register() failed: %v", err)
	}
	if isNew2 {
		t.Error("expected isNew=false for re-registration")
	}
	if d2.ID != d1.ID {
		t.Errorf("DeviceID changed: %s != %s", d2.ID, d1.ID)
	}
	// CommandSecret must be preserved on re-registration
	if d2.CommandSecret != d1.CommandSecret {
		t.Errorf("CommandSecret changed on re-registration: %s != %s", d2.CommandSecret, d1.CommandSecret)
	}
}

func TestRegister_hijackDetection(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register device with firebase-1
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-1",
		FCMToken:          "token1",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("First Register() failed: %v", err)
	}

	// Try to re-register same device with different firebase ID (hijack attempt)
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-2", // Different!
		FCMToken:          "token2",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err == nil {
		t.Error("expected ErrHijack when firebase ID differs")
	}
	if err != ErrHijack {
		t.Errorf("expected ErrHijack, got: %v", err)
	}
}

func TestDevice(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register a device
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Lookup device
	d, found, err := store.Device(ctx, "device-001")
	if err != nil {
		t.Fatalf("Device() failed: %v", err)
	}
	if !found {
		t.Error("expected device to be found")
	}
	if d.ID != "device-001" {
		t.Errorf("DeviceID = %s, want device-001", d.ID)
	}
}

func TestDevice_notFound(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	_, found, err := store.Device(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Device() failed: %v", err)
	}
	if found {
		t.Error("expected device not to be found")
	}
}

func TestSetOnline(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register device
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Set online
	if err := store.SetOnline(ctx, "device-001", true); err != nil {
		t.Fatalf("SetOnline() failed: %v", err)
	}

	d, _, _ := store.Device(ctx, "device-001")
	if !d.Online {
		t.Error("expected device to be online")
	}
}

func TestTouch(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register device
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Touch device
	if err := store.Touch(ctx, "device-001"); err != nil {
		t.Fatalf("Touch() failed: %v", err)
	}

	d, _, _ := store.Device(ctx, "device-001")
	if d.LastSeen.IsZero() {
		t.Error("LastSeen should be updated")
	}
}

func TestSaveTelemetry(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register device
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Save telemetry
	frame := models.TelemetryFrame{
		DeviceID:    "device-001",
		RiskScore:   25,
		BufferLevel: 60,
		ThermalTemp: 35.5,
	}
	if err := store.SaveTelemetry(ctx, "device-001", []byte(`{"test":true}`), frame); err != nil {
		t.Fatalf("SaveTelemetry() failed: %v", err)
	}
}

func TestSaveCommand(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register device
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Save command
	err = store.SaveCommand(ctx, "dispatch-001", "device-001", "update", []byte(`{"version":"2.0.0"}`), "sent")
	if err != nil {
		t.Fatalf("SaveCommand() failed: %v", err)
	}
}

func TestMarkDelivered(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register device
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Save and mark delivered
	err = store.SaveCommand(ctx, "dispatch-001", "device-001", "update", nil, "queued")
	if err != nil {
		t.Fatalf("SaveCommand() failed: %v", err)
	}

	if err := store.MarkDelivered(ctx, "dispatch-001"); err != nil {
		t.Fatalf("MarkDelivered() failed: %v", err)
	}
}

func TestSecret(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register device
	d, _, err := store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Get secret
	secret, found := store.Secret(ctx, "device-001")
	if !found {
		t.Error("expected secret to be found")
	}
	if secret != d.CommandSecret {
		t.Errorf("Secret = %s, want %s", secret, d.CommandSecret)
	}
}

func TestSecret_notFound(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	_, found := store.Secret(ctx, "nonexistent")
	if found {
		t.Error("expected secret not to be found")
	}
}

func TestUpdateFCM(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register device
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "old-token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Update FCM token
	if err := store.UpdateFCM(ctx, "device-001", "new-token"); err != nil {
		t.Fatalf("UpdateFCM() failed: %v", err)
	}

	d, _, _ := store.Device(ctx, "device-001")
	if d.FCMToken != "new-token" {
		t.Errorf("FCMToken = %s, want new-token", d.FCMToken)
	}
}

func TestDeleteDevice(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register device
	_, _, err = store.Register(ctx, models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	})
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Delete device
	if err := store.DeleteDevice(ctx, "device-001"); err != nil {
		t.Fatalf("DeleteDevice() failed: %v", err)
	}

	// Verify deleted
	_, found, _ := store.Device(ctx, "device-001")
	if found {
		t.Error("expected device to be deleted")
	}
}

// Operator tests
func TestCreateOperator(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	op := &models.Operator{
		ID:           "op-001",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hash",
		Role:         models.RoleOperator,
	}

	if err := store.CreateOperator(ctx, op); err != nil {
		t.Fatalf("CreateOperator() failed: %v", err)
	}

	// Verify created
	got, err := store.GetOperatorByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetOperatorByEmail() failed: %v", err)
	}
	if got.Name != "Test User" {
		t.Errorf("Name = %s, want Test User", got.Name)
	}
}

func TestGetOperatorByEmail_notFound(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	op, err := store.GetOperatorByEmail(ctx, "nonexistent@example.com")
	if err != nil {
		t.Fatalf("GetOperatorByEmail() unexpected error: %v", err)
	}
	if op != nil {
		t.Error("expected nil operator for nonexistent email")
	}
}

func TestGetOperatorByGoogleID(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	op := &models.Operator{
		ID:           "op-001",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hash",
		Role:         models.RoleOperator,
		GoogleID:     "google-123",
	}

	store.CreateOperator(ctx, op) //nolint:errcheck

	got, err := store.GetOperatorByGoogleID(ctx, "google-123")
	if err != nil {
		t.Fatalf("GetOperatorByGoogleID() failed: %v", err)
	}
	if got.Email != "test@example.com" {
		t.Errorf("Email = %s, want test@example.com", got.Email)
	}
}

// Session tests
func TestCreateSession(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create operator first
	op := &models.Operator{
		ID:           "op-001",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hash",
		Role:         models.RoleOperator,
	}
	store.CreateOperator(ctx, op) //nolint:errcheck

	// Create session
	sess := &models.Session{
		ID:         "sess-001",
		OperatorID: "op-001",
		TokenHash:  "hash123",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		CreatedAt:  time.Now(),
	}

	if err := store.CreateSession(ctx, sess); err != nil {
		t.Fatalf("CreateSession() failed: %v", err)
	}

	// Verify session exists
	got, err := store.GetSessionByTokenHash(ctx, "hash123")
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() failed: %v", err)
	}
	if got.OperatorID != "op-001" {
		t.Errorf("OperatorID = %s, want op-001", got.OperatorID)
	}
}

func TestDeleteSession(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create operator and session
	op := &models.Operator{
		ID:           "op-001",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hash",
		Role:         models.RoleOperator,
	}
	store.CreateOperator(ctx, op) //nolint:errcheck

	sess := &models.Session{
		ID:         "sess-001",
		OperatorID: "op-001",
		TokenHash:  "hash123",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		CreatedAt:  time.Now(),
	}
	store.CreateSession(ctx, sess) //nolint:errcheck

	// Delete session
	if err := store.DeleteSession(ctx, "hash123"); err != nil {
		t.Fatalf("DeleteSession() failed: %v", err)
	}

	// Verify deleted
	got, err := store.GetSessionByTokenHash(ctx, "hash123")
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil session for deleted token")
	}
}

func TestGetSessionByTokenHash_notFound(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	sess, err := store.GetSessionByTokenHash(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetSessionByTokenHash() unexpected error: %v", err)
	}
	if sess != nil {
		t.Error("expected nil session for nonexistent token")
	}
}
