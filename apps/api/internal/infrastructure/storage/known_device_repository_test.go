package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/devicestore"
)

func newKnownDeviceRepo(t *testing.T) *KnownDeviceRepository {
	t.Helper()

	s, err := Open(DefaultConfig(filepath.Join(t.TempDir(), "test.db")))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	return NewKnownDeviceRepository(s.DB())
}

// TestKnownDeviceRepo_NotifiesOnceThenSuppresses is the core behaviour against a
// real database: a new device registers and is reported known afterwards, and a
// repeat login from the same device/IP reports no change.
func TestKnownDeviceRepo_NotifiesOnceThenSuppresses(t *testing.T) {
	r := newKnownDeviceRepo(t)
	ctx := context.Background()

	const (
		op = "op-1"
		fp = "fp-1"
		ip = "203.0.113.10"
		ua = "UA-1"
	)

	known, err := r.IsKnownDevice(ctx, op, fp)
	if err != nil {
		t.Fatalf("IsKnownDevice: %v", err)
	}
	if known {
		t.Fatal("device should not be known before registration")
	}

	if err := r.RegisterDevice(ctx, op, fp, ip, ua); err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}

	known, err = r.IsKnownDevice(ctx, op, fp)
	if err != nil {
		t.Fatalf("IsKnownDevice: %v", err)
	}
	if !known {
		t.Fatal("device should be known after registration")
	}

	// The IP was recorded, so a repeat login is not a new-IP event.
	hasIP, err := r.HasSignal(ctx, op, fp, devicestore.KindIP, ip)
	if err != nil {
		t.Fatalf("HasSignal: %v", err)
	}
	if !hasIP {
		t.Fatal("the IP should have been recorded")
	}

	hasUA, err := r.HasSignal(ctx, op, fp, devicestore.KindUserAgent, ua)
	if err != nil {
		t.Fatalf("HasSignal: %v", err)
	}
	if !hasUA {
		t.Fatal("the User-Agent should have been recorded")
	}
}

// TestKnownDeviceRepo_NewIPIsNotASignal proves an unseen IP is reported absent, so
// the caller flags it.
func TestKnownDeviceRepo_NewIPIsNotASignal(t *testing.T) {
	r := newKnownDeviceRepo(t)
	ctx := context.Background()

	const (
		op = "op-1"
		fp = "fp-1"
	)

	if err := r.RegisterDevice(ctx, op, fp, "203.0.113.10", "UA-1"); err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}

	hasIP, err := r.HasSignal(ctx, op, fp, devicestore.KindIP, "198.51.100.99")
	if err != nil {
		t.Fatalf("HasSignal: %v", err)
	}
	if hasIP {
		t.Fatal("an unseen IP must not be reported as known")
	}
}

// TestKnownDeviceRepo_OperatorsAreIsolated proves one operator's device is not
// visible to another.
func TestKnownDeviceRepo_OperatorsAreIsolated(t *testing.T) {
	r := newKnownDeviceRepo(t)
	ctx := context.Background()

	if err := r.RegisterDevice(ctx, "op-1", "fp-shared", "203.0.113.10", "UA"); err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}

	known, err := r.IsKnownDevice(ctx, "op-2", "fp-shared")
	if err != nil {
		t.Fatalf("IsKnownDevice: %v", err)
	}
	if known {
		t.Fatal("a device registered for op-1 must not be known to op-2")
	}
}

// TestKnownDeviceRepo_DevicesBounded proves the per-operator device cap holds in SQL.
func TestKnownDeviceRepo_DevicesBounded(t *testing.T) {
	r := newKnownDeviceRepo(t)
	ctx := context.Background()

	for i := 0; i < maxKnownDevicesPerOperator*3; i++ {
		if err := r.RegisterDevice(ctx, "op-1", "fp-"+itoa(i), "203.0.113.10", "UA"); err != nil {
			t.Fatalf("RegisterDevice(%d): %v", i, err)
		}
	}

	var count int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM known_devices WHERE operator_id = ?`, "op-1").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count > maxKnownDevicesPerOperator {
		t.Fatalf("device count grew past the cap: %d > %d", count, maxKnownDevicesPerOperator)
	}
}

// TestKnownDeviceRepo_SignalsBounded proves the per-device signal cap holds in SQL.
func TestKnownDeviceRepo_SignalsBounded(t *testing.T) {
	r := newKnownDeviceRepo(t)
	ctx := context.Background()

	const (
		op = "op-1"
		fp = "fp-signals"
	)

	for i := 0; i < maxSignalsPerDevice*3; i++ {
		if err := r.RegisterDevice(ctx, op, fp, "10.0."+itoa(i/250)+"."+itoa(i%250), "UA"); err != nil {
			t.Fatalf("RegisterDevice(%d): %v", i, err)
		}
	}

	var count int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM known_device_signals
		  WHERE operator_id = ? AND fingerprint = ? AND kind = ?`,
		op, fp, devicestore.KindIP).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count > maxSignalsPerDevice {
		t.Fatalf("signal count grew past the cap: %d > %d", count, maxSignalsPerDevice)
	}
}

// TestKnownDeviceRepo_TouchUpdatesLastSeen proves TouchDevice moves the timestamp
// forward (used to keep a returning device from being evicted).
func TestKnownDeviceRepo_TouchUpdatesLastSeen(t *testing.T) {
	r := newKnownDeviceRepo(t)
	ctx := context.Background()

	const (
		op = "op-1"
		fp = "fp-1"
	)

	if err := r.RegisterDevice(ctx, op, fp, "203.0.113.10", "UA"); err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}

	var before int64
	if err := r.db.QueryRowContext(ctx,
		`SELECT last_seen_at FROM known_devices WHERE operator_id = ? AND fingerprint = ?`,
		op, fp).Scan(&before); err != nil {
		t.Fatalf("read last_seen: %v", err)
	}

	if err := r.TouchDevice(ctx, op, fp); err != nil {
		t.Fatalf("TouchDevice: %v", err)
	}

	var after int64
	if err := r.db.QueryRowContext(ctx,
		`SELECT last_seen_at FROM known_devices WHERE operator_id = ? AND fingerprint = ?`,
		op, fp).Scan(&after); err != nil {
		t.Fatalf("read last_seen: %v", err)
	}
	if after < before {
		t.Fatalf("last_seen went backwards: %d -> %d", before, after)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}

	return string(b)
}
