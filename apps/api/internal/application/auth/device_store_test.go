package auth

import (
	"context"
	"strconv"
	"testing"
)

// TestDeviceStore_NotifiesOnceThenSuppresses is the core behaviour of the login
// notification feature: the first login from a device notifies, and a repeat login
// from the same device (same fingerprint, same IP) does not.
func TestDeviceStore_NotifiesOnceThenSuppresses(t *testing.T) {
	s := &AuthService{deviceStore: NewDeviceStore()}
	ctx := context.Background()

	const (
		op = "op-1"
		fp = "fingerprint-a"
		ip = "203.0.113.10"
		ua = "Mozilla/5.0"
	)

	if !s.ShouldNotifyLogin(ctx, op, ip, ua, fp) {
		t.Fatal("first login from a new device should notify")
	}
	if s.ShouldNotifyLogin(ctx, op, ip, ua, fp) {
		t.Fatal("repeat login from the same device/IP should NOT notify")
	}
}

// TestDeviceStore_NewFingerprintNotifies proves a different device (different
// fingerprint) for the same operator notifies.
func TestDeviceStore_NewFingerprintNotifies(t *testing.T) {
	s := &AuthService{deviceStore: NewDeviceStore()}
	ctx := context.Background()

	const op = "op-1"

	if !s.ShouldNotifyLogin(ctx, op, "203.0.113.10", "UA-1", "fp-1") {
		t.Fatal("first device should notify")
	}
	if s.ShouldNotifyLogin(ctx, op, "203.0.113.10", "UA-1", "fp-1") {
		t.Fatal("same device should not notify")
	}
	if !s.ShouldNotifyLogin(ctx, op, "198.51.100.7", "UA-2", "fp-2") {
		t.Fatal("a second, different device should notify")
	}
}

// TestDeviceStore_NewIPOnKnownFingerprintNotifies proves an IP change on an
// otherwise-known device is flagged.
func TestDeviceStore_NewIPOnKnownFingerprintNotifies(t *testing.T) {
	s := &AuthService{deviceStore: NewDeviceStore()}
	ctx := context.Background()

	const (
		op = "op-1"
		fp = "fp-1"
		ua = "UA-1"
	)

	if !s.ShouldNotifyLogin(ctx, op, "203.0.113.10", ua, fp) {
		t.Fatal("first login should notify")
	}
	// Same fingerprint but a new IP.
	if !s.ShouldNotifyLogin(ctx, op, "198.51.100.99", ua, fp) {
		t.Fatal("a new IP on a known device should notify")
	}
	// Now that IP is known too, no further notification.
	if s.ShouldNotifyLogin(ctx, op, "198.51.100.99", ua, fp) {
		t.Fatal("known device + known IP should not notify")
	}
}

// TestDeviceStore_OperatorsAreIsolated proves one operator's known device does not
// suppress another operator's notification.
func TestDeviceStore_OperatorsAreIsolated(t *testing.T) {
	s := &AuthService{deviceStore: NewDeviceStore()}
	ctx := context.Background()

	const (
		fp = "fp-shared"
		ip = "203.0.113.10"
		ua = "UA-1"
	)

	if !s.ShouldNotifyLogin(ctx, "op-1", ip, ua, fp) {
		t.Fatal("op-1 first login should notify")
	}
	if !s.ShouldNotifyLogin(ctx, "op-2", ip, ua, fp) {
		t.Fatal("op-2 has never logged in from this device and should notify")
	}
}

// TestDeviceStore_StaysBounded proves the per-operator device map and the
// per-device signal sets cannot grow without bound.
func TestDeviceStore_StaysBounded(t *testing.T) {
	ds := NewDeviceStore()

	// Many distinct fingerprints for one operator.
	for i := 0; i < maxDevicesPerOperator*3; i++ {
		ds.RegisterDevice(context.Background(), "op-1", fingerprintN(i), "203.0.113.10", "UA")
	}

	ds.mu.RLock()
	deviceCount := len(ds.devices["op-1"])
	ds.mu.RUnlock()
	if deviceCount > maxDevicesPerOperator {
		t.Fatalf("device map grew past its cap: %d > %d", deviceCount, maxDevicesPerOperator)
	}

	// Many distinct IPs for a single device.
	const fp = "fp-signals"
	for i := 0; i < maxSignalsPerDevice*3; i++ {
		ds.RegisterDevice(context.Background(), "op-2", fp, ipN(i), "UA")
	}

	ds.mu.RLock()
	signals := len(ds.devices["op-2"][fp].IPAddresses)
	ds.mu.RUnlock()
	if signals > maxSignalsPerDevice {
		t.Fatalf("signal set grew past its cap: %d > %d", signals, maxSignalsPerDevice)
	}
}

func fingerprintN(i int) string {
	return "fp-" + strconv.Itoa(i)
}

func ipN(i int) string {
	return "10.0." + strconv.Itoa(i/250) + "." + strconv.Itoa(i%250)
}
