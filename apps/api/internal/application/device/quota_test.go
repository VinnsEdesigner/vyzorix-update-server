package device

import (
	"testing"
)

func TestCheckDeviceQuota_NoLimit(t *testing.T) {
	s := &Service{}
	err := s.CheckDeviceQuota(nil, "org-1", 0)
	if err != nil {
		t.Errorf("maxDevices=0 should allow, got: %v", err)
	}
}

func TestCheckDeviceQuota_NegativeLimit(t *testing.T) {
	s := &Service{}
	err := s.CheckDeviceQuota(nil, "org-1", -1)
	if err != nil {
		t.Errorf("negative maxDevices should allow, got: %v", err)
	}
}

func TestErrDeviceQuotaExceeded_Message(t *testing.T) {
	if ErrDeviceQuotaExceeded.Error() != "device quota exceeded for organization" {
		t.Errorf("unexpected error message: %s", ErrDeviceQuotaExceeded.Error())
	}
}
