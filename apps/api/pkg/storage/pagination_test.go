package storage

import (
	"context"
	"os"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

func TestDevicesPaginated_Basic(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vyzorix-pagination-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register multiple devices
	for i := 0; i < 10; i++ {
		_, _, err = store.Register(ctx, models.RegisterRequest{
			DeviceID:          "device-" + string(rune('0'+i)),
			FirebaseInstallID: "firebase-" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Fetch first page
	devices, err := store.DevicesPaginated(ctx, 5, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 5 {
		t.Errorf("First page got %d devices, want 5", len(devices))
	}
}

func TestDevicesPaginated_WithCursor(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vyzorix-pagination-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register multiple devices
	for i := 0; i < 10; i++ {
		_, _, err = store.Register(ctx, models.RegisterRequest{
			DeviceID:          "device-" + string(rune('0'+i)),
			FirebaseInstallID: "firebase-" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Fetch first page
	page1, err := store.DevicesPaginated(ctx, 5, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1) != 5 {
		t.Fatalf("First page got %d devices, want 5", len(page1))
	}

	// Use last device's timestamp as cursor
	cursor := page1[len(page1)-1].LastSeen.UnixMilli()

	// Fetch second page
	page2, err := store.DevicesPaginated(ctx, 5, cursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 5 {
		t.Errorf("Second page got %d devices, want 5", len(page2))
	}

	// Pages should not overlap
	for _, d1 := range page1 {
		for _, d2 := range page2 {
			if d1.ID == d2.ID {
				t.Error("Pages should not have overlapping devices")
			}
		}
	}
}

func TestDevicesPaginated_Limit(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vyzorix-pagination-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	// Register 3 devices
	for i := 0; i < 3; i++ {
		_, _, err = store.Register(ctx, models.RegisterRequest{
			DeviceID:          "device-" + string(rune('0'+i)),
			FirebaseInstallID: "firebase-" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Request more than available
	devices, err := store.DevicesPaginated(ctx, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 3 {
		t.Errorf("Got %d devices, want 3 (all available)", len(devices))
	}
}