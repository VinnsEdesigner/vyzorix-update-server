package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

func TestStorage_UpdateOperatorClientSettings(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	op := &models.Operator{
		ID:           "op-client-settings",
		Email:        "client-settings-test@example.com",
		Name:         "ClientSettings Test",
		PasswordHash: "hash",
		Role:         models.RoleOperator,
	}
	if err := store.CreateOperator(ctx, op); err != nil {
		t.Fatalf("CreateOperator() failed: %v", err)
	}

	client := models.ClientSettings{
		StrictHmac:           true,
		AutoReconnect:        false,
		NotificationsEnabled: false,
	}

	err = store.UpdateOperatorClientSettings(ctx, op.ID, client)
	if err != nil {
		t.Fatalf("UpdateOperatorClientSettings() failed: %v", err)
	}

	// Fetch the operator and verify
	fetched, err := store.GetOperatorByID(ctx, op.ID)
	if err != nil {
		t.Fatalf("GetOperatorByID() failed: %v", err)
	}
	if fetched == nil {
		t.Fatal("operator not found after update")
	}

	if !fetched.Client.StrictHmac {
		t.Error("StrictHmac = false, want true")
	}
	if fetched.Client.AutoReconnect {
		t.Error("AutoReconnect = true, want false")
	}
	if fetched.Client.NotificationsEnabled {
		t.Error("NotificationsEnabled = true, want false")
	}

	// Update again with different values
	client2 := models.ClientSettings{
		StrictHmac:           false,
		AutoReconnect:       true,
		NotificationsEnabled: true,
	}
	err = store.UpdateOperatorClientSettings(ctx, op.ID, client2)
	if err != nil {
		t.Fatalf("UpdateOperatorClientSettings() second call failed: %v", err)
	}

	fetched2, err := store.GetOperatorByID(ctx, op.ID)
	if err != nil {
		t.Fatalf("GetOperatorByID() second call failed: %v", err)
	}
	if fetched2.Client.StrictHmac {
		t.Error("StrictHmac = true, want false after second update")
	}
	if !fetched2.Client.AutoReconnect {
		t.Error("AutoReconnect = false, want true after second update")
	}
	if !fetched2.Client.NotificationsEnabled {
		t.Error("NotificationsEnabled = false, want true after second update")
	}
}

func TestStorage_UpdateOperatorThresholds(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	op := &models.Operator{
		ID:           "op-thresholds",
		Email:        "thresholds-test@example.com",
		Name:         "Thresholds Test",
		PasswordHash: "hash",
		Role:         models.RoleOperator,
	}
	if err := store.CreateOperator(ctx, op); err != nil {
		t.Fatalf("CreateOperator() failed: %v", err)
	}

	thresholds := models.Thresholds{
		RiskWarn:    30,
		RiskCrit:    60,
		ThermalWarn: 40,
		ThermalCrit: 50,
		BufferWarn:  40,
		BufferCrit:  70,
	}

	err = store.UpdateOperatorThresholds(ctx, op.ID, thresholds)
	if err != nil {
		t.Fatalf("UpdateOperatorThresholds() failed: %v", err)
	}

	// Fetch and verify
	fetched, err := store.GetOperatorByID(ctx, op.ID)
	if err != nil {
		t.Fatalf("GetOperatorByID() failed: %v", err)
	}
	if fetched == nil {
		t.Fatal("operator not found after threshold update")
	}

	if fetched.Thresholds.RiskWarn != 30 {
		t.Errorf("RiskWarn = %d, want 30", fetched.Thresholds.RiskWarn)
	}
	if fetched.Thresholds.RiskCrit != 60 {
		t.Errorf("RiskCrit = %d, want 60", fetched.Thresholds.RiskCrit)
	}
	if fetched.Thresholds.BufferCrit != 70 {
		t.Errorf("BufferCrit = %d, want 70", fetched.Thresholds.BufferCrit)
	}
}

func TestStorage_ResetOperatorSettings(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	op := &models.Operator{
		ID:           "op-reset",
		Email:        "reset-test@example.com",
		Name:         "Reset Test",
		PasswordHash: "hash",
		Role:         models.RoleOperator,
	}
	if err := store.CreateOperator(ctx, op); err != nil {
		t.Fatalf("CreateOperator() failed: %v", err)
	}

	// Set non-default values
	err = store.UpdateOperatorClientSettings(ctx, op.ID, models.ClientSettings{
		StrictHmac:           true,
		AutoReconnect:        false,
		NotificationsEnabled: false,
	})
	if err != nil {
		t.Fatalf("UpdateOperatorClientSettings() setup failed: %v", err)
	}

	err = store.UpdateOperatorThresholds(ctx, op.ID, models.Thresholds{
		RiskWarn:    10,
		RiskCrit:    20,
		ThermalWarn: 35,
		ThermalCrit: 45,
		BufferWarn:  20,
		BufferCrit:  40,
	})
	if err != nil {
		t.Fatalf("UpdateOperatorThresholds() setup failed: %v", err)
	}

	// Reset
	err = store.ResetOperatorSettings(ctx, op.ID)
	if err != nil {
		t.Fatalf("ResetOperatorSettings() failed: %v", err)
	}

	// Verify defaults restored
	fetched, err := store.GetOperatorByID(ctx, op.ID)
	if err != nil {
		t.Fatalf("GetOperatorByID() failed: %v", err)
	}
	if fetched == nil {
		t.Fatal("operator not found after reset")
	}

	// Check client settings defaults
	if fetched.Client.StrictHmac {
		t.Error("StrictHmac = true, want false (default)")
	}
	if !fetched.Client.AutoReconnect {
		t.Error("AutoReconnect = false, want true (default)")
	}
	if !fetched.Client.NotificationsEnabled {
		t.Error("NotificationsEnabled = false, want true (default)")
	}

	// Check threshold defaults
	if fetched.Thresholds.RiskWarn != 50 {
		t.Errorf("RiskWarn = %d, want 50 (default)", fetched.Thresholds.RiskWarn)
	}
	if fetched.Thresholds.RiskCrit != 75 {
		t.Errorf("RiskCrit = %d, want 75 (default)", fetched.Thresholds.RiskCrit)
	}
}

func TestStorage_ClientSettings_Defaults(t *testing.T) {
	tmp := t.TempDir()
	store, err := Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	op := &models.Operator{
		ID:           "op-defaults",
		Email:        "defaults-test@example.com",
		Name:         "Defaults Test",
		PasswordHash: "hash",
		Role:         models.RoleOperator,
	}
	if err := store.CreateOperator(ctx, op); err != nil {
		t.Fatalf("CreateOperator() failed: %v", err)
	}

	// Fetch the freshly created operator — GetOperatorByID applies COALESCE defaults
	fetched, err := store.GetOperatorByID(ctx, op.ID)
	if err != nil {
		t.Fatalf("GetOperatorByID() failed: %v", err)
	}
	if fetched == nil {
		t.Fatal("operator not found after creation")
	}

	// Verify default client settings from COALESCE
	if !fetched.Client.AutoReconnect {
		t.Error("new operator AutoReconnect = false, want true (default from COALESCE)")
	}
	if fetched.Client.StrictHmac {
		t.Error("new operator StrictHmac = true, want false (default from COALESCE)")
	}
	if !fetched.Client.NotificationsEnabled {
		t.Error("new operator NotificationsEnabled = false, want true (default from COALESCE)")
	}

	// Verify default thresholds from COALESCE
	if fetched.Thresholds.RiskWarn != 50 {
		t.Errorf("new operator RiskWarn = %d, want 50 (default from COALESCE)", fetched.Thresholds.RiskWarn)
	}
	if fetched.Thresholds.RiskCrit != 75 {
		t.Errorf("new operator RiskCrit = %d, want 75 (default from COALESCE)", fetched.Thresholds.RiskCrit)
	}
}